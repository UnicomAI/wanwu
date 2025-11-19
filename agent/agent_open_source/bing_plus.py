import requests
import json
import html
import datetime,time
from typing import Any, Dict, List
import re
import chardet
import math
from utils.bing_search import BingSearchAPIWrapper
from utils.auth import AccessTokenManager
from utils.timing import advanced_timing_decorator
#from utils.llm_tools import req_llm_structed_output
from utils.model_tools import *
from utils.uni_id import generate_unique_id
from utils.output_parser import extract_json
from utils.rerank import hybrid_rerank_results
from utils.bocha_search import bocha_cleaned_search
from utils.lc_web_parser import async_crawl_and_parse_webpage
from utils.build_prompt import build_bing_prompt_from_search_list
from bs4 import BeautifulSoup
import asyncio
import aiohttp
from aiohttp import ClientTimeout
import nest_asyncio
import pytz
import concurrent.futures
import traceback

# 允许在现有事件循环中嵌套使用 asyncio [EN] Allows nested use of asyncio within existing event loops
nest_asyncio.apply()

import logging

logger = logging.getLogger(__name__)

import configparser

config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

BING_SUBSCRIPTION_KEY = config["BING"]["BING_SUBSCRIPTION_KEY"]
BING_SEARCH_URL = config["BING"]["BING_SEARCH_URL"]

BING_RESULT_LEN =  int(config["BING"]["BING_RESULT_LEN"])
BING_TOP_K =  int(config["BING"]["BING_TOP_K"])
BING_THRESHOLD =  float(config["BING"]["BING_THRESHOLD"])
BING_SENTENCE_SIZE =  int(config["BING"]["BING_SENTENCE_SIZE"])
BING_TIME_OUT =  float(config["BING"]["BING_TIME_OUT"])
TARGET_SUCCESS = int(config["BING"]["TARGET_SUCCESS"])
USE_CHROME = bool(config["BING"]["USE_CHROME"])

BING_WEIGHT =  float(config["BING"]["BING_WEIGHT"])
BOCHA_WEIGHT = float(config["BING"]["BOCHA_WEIGHT"])


QUERY_SHORT=10

token_manager = AccessTokenManager()

def clean_text(text):
    """清除文本中的特殊字符和多余的空白，以及HTML标签。"""
    patterns = [
        r'\xa0+', r'\u3000', r'\t+', r'\r+', r'\n+',   # 清除特殊空白字符和多行换行符 [EN] Clear special whitespace characters and multiline newlines
        r'<[/]?b>|&gt;|&lt;'                        # 清除HTML标签 [EN] Clear HTML tags
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

def clean_list(data_list, black_domains=None):
    """对列表中的每个元素进行清理，并跳过包含指定黑名单网站的链接。"""
    if black_domains is None:
        # 如果没有传入黑名单网站列表，则默认使用 bendibao.com （天气预报不准确） [EN] If no blacklist website list is passed in, bendibao.com is used by default (the weather forecast is inaccurate)
        
        black_domains = ["bendibao.com"]

    return [
        {
            "type": "SE",
            "id": item.get("id", ""),
            "title": clean_text(item.get("title", "")),
            "snippet": clean_text(item.get("snippet", "")),
            "link": item.get("link", ""),
            "datePublished": item.get("datePublished", ""),
            "dateLastCrawled": item.get("dateLastCrawled", "")
        }
        for item in data_list
        # 跳过 link 中包含任意黑名单域名的元素 [EN] Skip elements in links containing any blacklisted domain name
        if not any(domain in item.get("link", "") for domain in black_domains)
    ]





# 主流程：step1：Bing 搜索 [EN] Main process: step1: Bing search
@advanced_timing_decorator(task_name="bing_cleaned_search")
def bing_cleaned_search(query, result_len=BING_RESULT_LEN, days_limit=-1,setLang ="zh-hans"):
    """
    通过 Bing 搜索获取结果，并可选择根据发布日期筛选最近的条目。

    参数:
    query (str): 搜索查询字符串。
    result_len (int): 返回的搜索结果数量，默认值为 BING_RESULT_LEN。
    days_limit (int or None): 筛选结果的天数限制。如果为 None，则不进行日期筛选；否则只返回最近指定天数内的结果。

    返回:
    list: 根据查询条件和筛选天数返回的搜索结果列表。
    """

    # 检查是否设置了 Bing 搜索相关的环境变量 [EN] Check whether Bing search related environment variables are set
    if not (BING_SEARCH_URL and BING_SUBSCRIPTION_KEY):
        # 如果没有设置相关环境变量，则返回一个提示信息 [EN] If the relevant environment variables are not set, a prompt message will be returned.
        return [{"type":"SE",
                 "snippet": "please set BING_SUBSCRIPTION_KEY and BING_SEARCH_URL in os ENV",
                 "title": "env info is not found",
                 "link": "https://python.langchain.com/en/latest/modules/agents/tools/examples/bing_search.html"}]
    
    # 初始化 Bing 搜索器对象 [EN] Initialize the Bing searcher object
    bing_searcher = BingSearchAPIWrapper(bing_subscription_key=BING_SUBSCRIPTION_KEY,
                                         bing_search_url=BING_SEARCH_URL)

    # 获取 Bing 搜索结果 [EN] Get Bing search results 
    bing_result = bing_searcher.results(query, result_len)
    
#     print(len(bing_result))    
#     print(json.dumps(bing_result,ensure_ascii=False))

    # 如果 days_limit =-1，表示不进行日期筛选，直接返回清理后的搜索结果 [EN] If days_limit =-1, it means no date filtering is performed and the cleaned search results are returned directly.
    if days_limit == -1:
        return clean_list(bing_result)[:result_len]

    # 获取东八区（中国标准时间）时区 [EN] Get the time zone of East Eighth District (China Standard Time)
    tz = pytz.timezone('Asia/Shanghai')

    # 获取当前的东八区时间 [EN] Get the current Dongba District time
    today = datetime.datetime.now(tz)

    # 计算指定天数之前的日期，作为筛选的截止日期 [EN] Calculate the date before the specified number of days as the cut-off date for filtering
    cutoff_date = today - datetime.timedelta(days=days_limit)

    # 筛选符合条件的搜索结果（包括发布日期为空或者在指定天数内的结果） [EN] Filter search results that match the criteria (including results with a blank publication date or within a specified number of days)
    recent_articles = []
    
    for item in bing_result:
        # 获取发布日期字段，如果不存在则返回空字符串 [EN] Gets the publication date field, returning an empty string if it does not exist
        date_published_str = item.get('datePublished', '')  
        
        if not date_published_str:
            # 如果 datePublished 为空或缺失，则将此条目加入筛选结果 [EN] If datePublished is empty or missing, add this entry to the filter results
            recent_articles.append(item)
        else:
            try:
                # 将发布日期字符串解析为 datetime 对象，并转换为东八区时间 [EN] Parse the release date string into a datetime object and convert it to East Eighth District time
                published_date = datetime.datetime.strptime(date_published_str, "%Y-%m-%d").replace(tzinfo=pytz.UTC).astimezone(tz)
                
                # 如果发布日期在指定的天数范围内，则将该条目加入筛选结果 [EN] If the release date is within the specified number of days, add the entry to the filter results
                if published_date > cutoff_date:
                    recent_articles.append(item)
            except ValueError:
                # 如果日期格式不正确，打印警告并将该条目包含在筛选结果中 [EN] If the date format is incorrect, print a warning and include the entry in the filter results
                print(f"Warning: Invalid date format for {date_published_str}, including it in results.")
                recent_articles.append(item)
   
    # print("############")
    # print(len(recent_articles))
    # print(json.dumps(recent_articles,ensure_ascii=False))

    # 清理并返回筛选后的搜索结果，限制结果数量为 result_len [EN] Clean and return filtered search results, limiting the number of results to result_len
    return clean_list(recent_articles)[:result_len]

@advanced_timing_decorator(task_name="build_bing_query")
def build_bing_query(query):
    return [query]


    

async def req_url_parser(bing_single_item, query, sentence_size, time_out, is_google):
    parser_url = config["BING"]["DEFAULT_URL_PARSER"]
    overlap_size = float(config["BING"]["BING_OVERLAP_SIZE"])
    headers = {
        "Content_type": "application/json;charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}",
    }
    payload = {
        "url": bing_single_item["link"],
        "is_google": is_google,
        "sentence_size": sentence_size,
        "overlap_size": overlap_size,
    }
    timeout = aiohttp.ClientTimeout(total=time_out)

    async with aiohttp.ClientSession(timeout=timeout) as session:
        try:
            start_time = datetime.datetime.now()
            async with session.post(parser_url, json=payload, headers=headers) as response:
                try:
                    result_dict = json.loads(await response.text())
                except aiohttp.ContentTypeError:
                    elapsed_time = (datetime.datetime.now() - start_time).total_seconds()
                    logger.error(f"Failed to parse JSON response. {bing_single_item['link']} ---> 耗时: {elapsed_time}s")
                    logger.error(f"Response text: {await response.text()}")
                    return []
                except Exception as e:
                    pass
                    # logger.error(f"Unexpected error during JSON parsing: {e}")
                    return []

                
                docs = result_dict.get("docs", [])
                elapsed_time = (datetime.datetime.now() - start_time).total_seconds()
                if not docs:
                    # logger.info(f"解析为空:{query[:QUERY_SHORT]}  ---> 耗时: {elapsed_time}s ---> {bing_single_item['link']}") [EN] logger.info(f"Resolved as empty: {query[:QUERY_SHORT]} ---> Time consuming: {elapsed_time}s ---> {bing_single_item['link']}")
                    return []

                logger.info(f"解析【正常】:{query[:QUERY_SHORT]}  ---> 耗时: {elapsed_time}s ---> 个数：{len(docs)}---> {bing_single_item['link']}")
                # logger.info(f"解析【正常】:{docs}") [EN] logger.info(f"Parse [normal]:{docs}")

                updated_data_list = [
                    {
                        "type": "SE",
                        "id": generate_unique_id(),
                        "title": bing_single_item["title"],
                        "snippet": item["text"],
                        "link": bing_single_item["link"],
                        "datePublished": bing_single_item["datePublished"],
                        "dateLastCrawled": bing_single_item["dateLastCrawled"],
                    }
                    for item in docs
                ]
                return updated_data_list
        except aiohttp.ClientError as e:
            logger.error(f"Client error during request: {e}")
        except asyncio.TimeoutError:
            elapsed_time = (datetime.datetime.now() - start_time).total_seconds()
            # logger.info(f"解析超时:{query[:QUERY_SHORT]}  ---> 耗时: {elapsed_time}s ---> {bing_single_item['link']}") [EN] logger.info(f"Parse timeout: {query[:QUERY_SHORT]} ---> Time consuming: {elapsed_time}s ---> {bing_single_item['link']}")
        except Exception as e:
            pass
            # logger.error(f"Unexpected error during request: {e}")
        return []


async def process_url_with_fallback(bing_single_item, query, sentence_size, time_out):
    async def fetch_parser(is_google):
        try:
            return await req_url_parser(bing_single_item, query, sentence_size, time_out, is_google)
        except asyncio.CancelledError:
            return []  # 任务被取消，返回空列表 [EN] The task is canceled and an empty list is returned.
        except Exception as e:
            logger.error(f"解析器任务出错: {e}")
            return []

    # 创建多个并行任务 [EN] Create multiple parallel tasks
    tasks = [
        #asyncio.create_task(fetch_parser(0)),
        #asyncio.create_task(fetch_parser(1)),
        asyncio.create_task(
            async_crawl_and_parse_webpage(
                bing_single_item=bing_single_item, 
                query=query, 
                sentence_size=sentence_size,
                time_out=time_out
            )
        )
    ]


    

    try:
        while tasks:
            # 等待最先完成的任务 [EN] Wait for the first task to complete
            done, pending = await asyncio.wait(tasks, return_when=asyncio.FIRST_COMPLETED)

            success_flag = False
            for task in done:
                result = await task
                if result:  # 如果该任务返回了有效结果 [EN] If the task returns valid results
                    # 取消所有其他未完成的任务 [EN] Cancel all other outstanding tasks
                    for p in pending:
                        p.cancel()
                    
                    await asyncio.gather(*pending, return_exceptions=True)  # 确保任务完全终止 [EN] Make sure the task is completely terminated
                    
                    return result  # 只返回第一个成功的结果 [EN] Only return the first successful result

            # 如果 `done` 任务都返回了空数据，则继续等待剩余任务 [EN] If all `done` tasks return empty data, continue to wait for the remaining tasks.
            tasks = list(pending)

    finally:
        # 确保所有任务都被取消，避免资源泄露 [EN] Ensure all tasks are canceled to avoid resource leakage
        for task in tasks:
            task.cancel()
        await asyncio.gather(*tasks, return_exceptions=True)

    return bing_single_item  # 如果所有任务都失败，返回 bing原始结果 [EN] If all tasks fail, return the original result of bing
    # return []




async def process_urls(hybrid_search_list, query, sentence_size, time_out, target_success=TARGET_SUCCESS):
    start_time = datetime.datetime.now()
    aggregated_results = []
    success_count = 0

    tasks = [asyncio.create_task(process_url_with_fallback(item, query, sentence_size, time_out)) for item in hybrid_search_list]

    try:
        for completed_task in asyncio.as_completed(tasks):
            try:
                result = await completed_task
                if isinstance(result, list) and result:
                    aggregated_results.extend(result)
                    success_count += 1
                    if success_count >= target_success:
                        logger.info(f"目标已达成，成功获取 {success_count} 个结果，开始取消剩余任务。")

                        # 取消所有未完成的任务 [EN] Cancel all unfinished tasks
                        for task in tasks:
                            if not task.done():
                                task.cancel()
                        
                        break  # 立即退出循环 [EN] Exit the loop now
            except asyncio.CancelledError:
                logger.info("process_urls 任务被取消")
                return aggregated_results  # 确保返回已有结果 [EN] Make sure to return existing results
            except Exception as e:
                # logger.error(f"任务执行出错: {e}") [EN] logger.error(f"Task execution error: {e}")
                pass

    finally:
        for task in tasks:
            task.cancel()
        await asyncio.sleep(0)  # 让出控制权，确保取消生效 [EN] Give up control and ensure cancellation takes effect
        await asyncio.gather(*tasks, return_exceptions=True)

    return aggregated_results




# 定义并发查询函数 [EN] Define concurrent query functions
@advanced_timing_decorator(task_name="handle_sub_query")
def handle_sub_query(
    search_url,
    search_key,
    search_rerank_id,
    sub_query,
    bing_target_success,
    top_k=BING_TOP_K,
    threshold=BING_THRESHOLD, 
    sentence_size=BING_SENTENCE_SIZE, 
    time_out= BING_TIME_OUT,
    days_limit=-1,
    result_len = BING_RESULT_LEN,
    isConcurrent=False,
    setLang ="zh-hans",
    is_rerank = True,
):  
    
    logger.info(f"search_url打印出来: {search_url}")
    logger.info(f"search_key打印出来: {search_key}")

    #---- 调用bing original api查询 [EN] ---- Call bing original api query
    hybrid_search_list = hybrid_clean_search(sub_query, search_url,search_key,result_len=result_len)  
    # bing_search_list = bing_cleaned_search(sub_query,result_len = result_len )
    
    #---- url内容抓取解析 [EN] ---- URL content capture and analysis
    start_time = datetime.datetime.now()
    sub_results = asyncio.run(process_urls(hybrid_search_list, sub_query, sentence_size, time_out, bing_target_success))
    # sub_results = asyncio.run(process_urls_batch(hybrid_search_list, sub_query, sentence_size, time_out,batch_size))
    
    finish_time2 = datetime.datetime.now()
    time_difference = finish_time2 - start_time

    logger.info(f"{sub_query} URL解析个数:{len(sub_results)}")  

    if sub_results:
        if isConcurrent:
            # 并发调用时，将sub_query后面增加箭头，目的是优化给LLM的prompt [EN] When calling concurrently, add an arrow after sub_query to optimize the prompt to LLM. 
            for sub_result in sub_results:
                sub_result["sub_query"] = sub_query + "-> "
    else:   
        # 如果URL解析接口为空，使用搜索引擎原始结果 [EN] If the URL parsing interface is empty, the original search engine results will be used.
        sub_results.extend(hybrid_search_list)
    
    if is_rerank:        
        # 使用新的重排序函数 [EN] Use new reorder function
        sub_aggregated_results = hybrid_rerank_results(sub_query,search_rerank_id, sub_results, top_k, threshold)
        return sub_aggregated_results,hybrid_search_list

    else:
        return sub_results, hybrid_search_list


@advanced_timing_decorator(task_name="req_bing_multiquery")
def req_bing_multiquery(
    sub_query_list,
    top_k,
    search_url,
    search_key,
    threshold=BING_THRESHOLD,
    sentence_size=BING_SENTENCE_SIZE,
    time_out=BING_TIME_OUT,
    days_limit=-1,
    result_len=BING_RESULT_LEN,
    target_success = TARGET_SUCCESS,
    setLang="zh-hans",
    is_rerank = True,
    # batch_size=PARSE_BATCH_SIZE
):
    aggregated_results = []
    aggregated_origin_results = []
    
    with concurrent.futures.ThreadPoolExecutor() as executor:
        futures = [
            executor.submit(
                handle_sub_query, 
                search_url,
                search_key,
                sub_query,
                target_success,
                top_k,
                threshold,
                sentence_size,
                time_out,
                days_limit,
                result_len,
                isConcurrent=True,
                setLang = setLang,
                is_rerank = is_rerank,
            )
            for sub_query in sub_query_list
        ]
        
        for future in concurrent.futures.as_completed(futures):
            try:
                result,origin_result = future.result()
                aggregated_results.extend(result)
                aggregated_origin_results.extend(origin_result)
            except Exception as e:
                # 打印异常的详细信息 [EN] Print exception details
                logger.info(f"An exception occurred during concurrent execution: {e}")
                traceback.print_exc()
                # 抛出异常以确保调用者知道发生了错误 [EN] Throw an exception to ensure the caller knows an error has occurred
                raise
    
    return aggregated_results,aggregated_origin_results

def combine_search_list(search_list):
    combined_dict = {}

    for item in search_list:
        link = item['link']
        if link in combined_dict:
            # 合并snippet，使用换行符连接 [EN] Merge snippets, use newlines to connect
            combined_dict[link]['snippet'] += '\n' + item['snippet']
            
        else:
            # 初始化字典项，复制当前元素以避免修改原数据 [EN] Initialize dictionary items, copying current elements to avoid modifying the original data
            combined_dict[link] = item.copy()

    # 将合并后的字典转换回列表形式 [EN] Convert the merged dictionary back to list form
    combined_list = list(combined_dict.values())
    return combined_list

@advanced_timing_decorator()
def hybrid_clean_search(
    sub_query,
    search_url,
    search_key,
    result_len,
    bing_weight=BING_WEIGHT,
    bocha_weight=BOCHA_WEIGHT,
    setLang ="zh-hans"
    ):
    # 创建一个线程池，对联网查询结果进行emb和bm25两路并发排序 [EN] Create a thread pool to perform emb and bm25 concurrent sorting on network query results.
    bing_search_list=[]
    bocha_search_list=[]

    
    with concurrent.futures.ThreadPoolExecutor() as executor:
        # 准备两个任务的参数 [EN] Prepare parameters for two tasks

        try:
            if bing_weight==1.0:
                bing_search_task = executor.submit(bing_cleaned_search, sub_query,result_len = result_len,setLang=setLang) 
                bing_search_list = bing_search_task.result() 
            elif bocha_weight==1.0:
                bocha_search_task = executor.submit(bocha_cleaned_search, sub_query,search_url,search_key,result_len = result_len) 
                bocha_search_list = bocha_search_task.result()
            else:
                bing_search_task = executor.submit(bing_cleaned_search, sub_query,result_len = result_len,setLang=setLang) 
                bocha_search_task = executor.submit(bocha_cleaned_search, sub_query,search_url,search_key,result_len = result_len) 
                bing_search_list = bing_search_task.result()    
                bocha_search_list = bocha_search_task.result()

            bing_num = math.ceil(result_len*bing_weight)
            bocha_num = result_len - bing_num
            hybrid_search_list = bing_search_list[:bing_num]+bocha_search_list[:bocha_num]
            
            # logger.info(f"{sub_query} ") 
            logger.info(f"{sub_query} __Hybrid original total个数:{len(hybrid_search_list)} -->bing:{len(bing_search_list[:bing_num])}  bocha:{len(bocha_search_list[:bocha_num])}")  
            logger.info(f"hybrid_search_list输出是: {hybrid_search_list}")
            
            return hybrid_search_list
                
        except Exception as e:
            # 打印异常的详细信息 [EN] Print exception details
            logger.info(f"An exception occurred during concurrent execution: {e}")
            traceback.print_exc()
            # 抛出异常以确保调用者知道发生了错误 [EN] Throw an exception to ensure the caller knows an error has occurred
            raise
        

@advanced_timing_decorator(task_name="get_bing_multi_lang_query")
def get_bing_multi_lang_query(sub_query_list_cn,sub_query_list_en,days_limit=-1,auto_citation=False):
    prompt =''
    aggregated_results = []
    
    #----bing query改写 [EN] ----Bing query rewritten
    start_time = datetime.datetime.now()    
   
    
    time_out = BING_TIME_OUT
    result_len = BING_RESULT_LEN
    
        
    aggregated_results_cn,orgin_results_cn   = req_bing_multiquery(
        sub_query_list_cn,
        time_out=time_out,
        top_k= BING_TOP_K,
        days_limit=days_limit,
        result_len = result_len,
        setLang="zh-hans",
        is_rerank = False,
        )


    aggregated_results_en,orgin_results_en =  req_bing_multiquery(
        sub_query_list_en,
        time_out=time_out,
        top_k= BING_TOP_K,
        days_limit=days_limit,
        result_len = result_len,
        setLang="en",
        is_rerank = False,
        )
    aggregated_results = aggregated_results_cn+aggregated_results_en
    orgin_results = orgin_results_cn+orgin_results_en
    finish_time2 = datetime.datetime.now()
    time_difference2 = finish_time2 - start_time
    logger.info(f"req_bing_multiquery time: {time_difference2}")
    


    # logger.info(f"bing_multi_search_result: {json.dumps(aggregated_results,ensure_ascii=False,indent=4)}")
    if auto_citation:        
        combined_list = combine_search_list(aggregated_results)
        combine_origin_list = combine_search_list(orgin_results)
        logger.info(f"合并对照: {len(aggregated_results)} -->{len(combined_list)}")
    else:
        combined_list =  deduplicate_by_title(aggregated_results)
        combine_origin_list = deduplicate_by_title(orgin_results)

    for item in combined_list:
        item.pop("datePublished","")
        item.pop("dateLastCrawled","")
        item.pop("sub_query","")
    

    return combined_list,combine_origin_list



def deduplicate_by_title(search_list):
    combined_dict = {}
    original_length = len(search_list)

    for item in search_list:
        title = item['title']
        if title in combined_dict:
            # 如果标题已存在，比较snippet长度，保留较长的那个 [EN] If the title already exists, compare the snippet lengths and keep the longer one.
            if len(item['snippet']) > len(combined_dict[title]['snippet']):
                combined_dict[title] = item.copy()
        else:
            # 初始化字典项，复制当前元素以避免修改原数据 [EN] Initialize dictionary items, copying current elements to avoid modifying the original data
            combined_dict[title] = item.copy()

    # 将合并后的字典转换回列表形式 [EN] Convert the merged dictionary back to list form
    combined_list = list(combined_dict.values())
    final_length = len(combined_list)
    
    logger.info(f"搜索结果去重: {original_length} --> {final_length}")
    
    return combined_list

@advanced_timing_decorator(task_name="get_bing_multi_search_result")
def get_bing_multi_search_result(query,bing_top_k,bing_time_out,bing_target_success,bing_result_len,model,search_url,search_key,search_rerank_id,days_limit=-1,auto_citation=False):
    prompt =''
    aggregated_results = []
    
    #----bing query改写 [EN] ----Bing query rewritten
    start_time = datetime.datetime.now()    
    sub_query_list = build_bing_query(query)

    logger.info(f"BING sub query: {json.dumps(sub_query_list,ensure_ascii=False)}")

    print('aaaaaaaaaaa')    
    time_out = bing_time_out
    result_len = bing_result_len
    BING_TOP_K = bing_top_k

    sub_query_num =len(sub_query_list)

    # 控制多个子查询时total_top_k:  =sub_query_num * top_k_ration * BING_TOP_K [EN] When controlling multiple subqueries total_top_k: =sub_query_num * top_k_ration * BING_TOP_K
    if sub_query_num <= 3:
        top_k_ration = 1
    elif sub_query_num <= 6:
        top_k_ration = 0.5
    elif sub_query_num <= 10:
        top_k_ration = 0.3
    else:
        top_k_ration = 0.1



    if len(sub_query_list)>1:
        aggregated_results,_ = req_bing_multiquery(sub_query_list,time_out=time_out,top_k= math.ceil(BING_TOP_K*top_k_ration),days_limit=days_limit,result_len = result_len,target_success = bing_target_success,search_url=search_url,search_key=search_key)
        finish_time2 = datetime.datetime.now()
    
    else:
        logger.info(f"子query只有一个: {json.dumps(sub_query_list,ensure_ascii=False)}")
        logger.info(f"search_url打印出来: {search_url}")
        logger.info(f"search_key打印出来: {search_key}")
        
        aggregated_results,_ = handle_sub_query(search_url,search_key,search_rerank_id,sub_query_list[0],bing_target_success,bing_top_k,BING_THRESHOLD,BING_SENTENCE_SIZE,bing_time_out,days_limit,bing_result_len)
        
    finish_time2 = datetime.datetime.now()
    time_difference2 = finish_time2 - start_time        
    logger.info(f"req_bing_multiquery time: {time_difference2}")
    
    logger.info(f"aggregated_results 是什么: {aggregated_results}")
    
    #----准备提示词 [EN] ----Prepare prompt words
    prompt = build_bing_prompt_from_search_list(query,aggregated_results,auto_citation,model)    

    logger.info(f"prompt是:{prompt}")
    return prompt,aggregated_results
async def async_search(query, bing_top_k,bing_time_out,bing_target_success,bing_result_len,model,search_url,search_key,search_rerank_id,days_limit=-1, auto_citation=False):
    # 假设 get_bing_multi_search_result 是同步的，需要运行在线程池中 [EN] Assume get_bing_multi_search_result is synchronous and needs to run in the thread pool
    loop = asyncio.get_running_loop()
    result = await loop.run_in_executor(
        None,  # 默认线程池 [EN] Default thread pool
        get_bing_multi_search_result,  # 同步函数 [EN] Synchronous function
        query,  # 参数1 [EN] Parameter 1
        bing_top_k,
        bing_time_out,
        bing_target_success,
        bing_result_len,
        model,  # 参数2 [EN] Parameter 2
        search_url,
        search_key,
        search_rerank_id,
        days_limit,  # 参数2 [EN] Parameter 2
        auto_citation# 参数3 [EN] Parameter 3
    )
    return result
    

# 启动任务并返回任务对象 [EN] Start the task and return the task object
def start_async_search(loop, query,bing_top_k,bing_time_out,bing_target_success,bing_result_len,model,days_limit, auto_citation,search_url,search_key,search_rerank_id):
    task = loop.create_task(async_search(query, bing_top_k,bing_time_out,bing_target_success,bing_result_len,model,search_url,search_key,search_rerank_id,days_limit, auto_citation))
    # print("异步任务已启动，不会阻塞后续代码") [EN] print("The asynchronous task has been started and will not block subsequent code")
    return task

# 用于运行异步函数的示例 [EN] Example for running an asynchronous function
if __name__ == "__main__":
    query = "明天北京天气怎么样啊"

    
    query = "今天北京天气"
    
    # query = "今年的举重世锦赛将在哪里举办？" [EN] query = "Where will this year's World Weightlifting Championships be held?"
    # query = "邓紫棋举办过多少场演唱会？" [EN] query = "How many concerts has Deng Ziqi held?"
    # query = "腾讯美股股价" [EN] query = "Tencent US stock price"
    # query = "阿里美股股价" [EN] query = "Alibaba US stock price"
    # query = "百度美股股价" [EN] query = "Baidu US stock price"
    # query = "2024中非论坛为什么不在非常举办，而是在北京" [EN] query = "Why is the 2024 China-Africa Forum not held in Beijing, but in Beijing?"

    result = build_bing_query(query)
    print(result)



