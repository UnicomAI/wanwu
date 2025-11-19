from re import sub
import requests
import json
import logging
import datetime,time
import logging
import pytz
import math
import concurrent.futures
import traceback
from utils.model_tools import *
from utils.bm25 import rerank_by_bm25
from utils.tokenizers import CustomTokenizer
from utils.llm_tools import format_prompt_template
from utils.auth import AccessTokenManager
# from utils.bing_plus import *



logger = logging.getLogger(__name__)

import configparser


token_manager = AccessTokenManager()


BING_TOP_K =  5
BING_THRESHOLD =  0.4










def rerank_by_emb(query,search_rerank_id,raw_search_list, top_k=5, threshold = 0.4,model = "bge"):

    if not raw_search_list:
        return {}
    url = f"http://bff-service:6668/callback/v1/model/{search_rerank_id}"
    logger.info(f"输入url是: {url}")
    response = requests.get(url)
    model_name = ''
    # 检查响应状态码 [EN] Check response status code
    if response.status_code == 200:
        data = response.json()
        # 获取 model 字段 [EN] Get model fields
        model_name = data.get("data", {}).get("model")
        logger.info(f"Model:{model_name}")
    else:
        print("Request failed with status code:", response.status_code)

    url  = f"http://bff-service:6668/callback/v1/model/{search_rerank_id}/rerank"

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json"
    }
    logger.info(f"rerank前的list是:{raw_search_list}")
    sorted_search_list = []
    snippets = [item['snippet'] for item in raw_search_list]
    data = {
        "model": model_name,
        "query": query,
        "documents": snippets
    }
    # data = json.dumps(data)
    logger.info(f"data is:{data}")
    response = requests.post(url, headers=headers, data=json.dumps(data))

    # 处理返回结果 [EN] Process the returned results
    if response.status_code == 200:
        json_data = response.json()
        logger.info(f"json_data is:{json_data}")
        results = json_data.get("results", [])

        # 按得分降序排序 [EN] Sort by score in descending order
        sorted_search_list = [raw_search_list[r["index"]] for r in sorted(results, key=lambda x: -x["relevance_score"])]
    else:
        print("Request failed:", response.status_code)
    return sorted_search_list





# 定义并发排序函数 [EN] Define concurrent sorting function
@advanced_timing_decorator(task_name="execute_rerank")
def execute_rerank(query,search_rerank_id, search_list, top_k=BING_TOP_K, threshold=BING_THRESHOLD, rerank_type = "bge"):
    if rerank_type == "bge":
        sub_result_by_bge= rerank_by_emb(query,search_rerank_id, search_list, top_k=top_k, threshold=threshold,model =rerank_type)
        logger.info(f"sub_result_by_bge::{sub_result_by_bge}")
        return sub_result_by_bge



# 新增函数：合并并重排序搜索结果 [EN] New function: merge and reorder search results
@advanced_timing_decorator(task_name="hybrid_rerank_results")
def hybrid_rerank_results(query,search_rerank_id, search_list, top_k=BING_TOP_K, threshold=BING_THRESHOLD):
    """
    对搜索结果进行并行重排序并合并结果

    参数:
    query (str): 搜索查询
    sub_results (list): 搜索结果列表
    top_k (int): 返回的结果数量
    threshold (float): 相关性阈值

    返回:
    list: 重排序后的结果列表
    """

    with concurrent.futures.ThreadPoolExecutor() as executor:
        try:

            rerank_task_bge = executor.submit(execute_rerank, query,search_rerank_id, search_list, top_k, threshold, rerank_type="bge")
            result = rerank_task_bge.result()
        except Exception as e:
            logger.info(f"An exception occurred during concurrent execution: {e}")
            traceback.print_exc()
            raise

    return result[:top_k] if result else []




if __name__ == "__main__":




    search_list = [
        {
            "id": "62bc33e8057f47f4a4f1c9bb32e3abce",
            "title": "奖牌统计 - 2024年巴黎奥运会奖牌榜 - Olympics.com",
            "snippet": "查看2024年巴黎奥运会的奖牌统计，包括金、银、铜奖牌得主的国家和数量。美利坚合众国、中国、日本等国家领跑奖牌榜，法国、英国、德国等国家也有不错的表现。",
            "link": "https://olympics.com/zh/paris-2024/medals",
            "datePublished": "",
            "dateLastCrawled": "2024-08-28"
        },
        {
            "id": "6d49cfcef74e4c76947fb0fed918b06d",
            "title": "奖牌榜_2024巴黎奥运会_体育_央视网(cctv.com)",
            "snippet": "查看2024年巴黎奥运会的金牌、银牌和铜牌排名，以及各国家和地区的奖牌数量和总数。中国队目前获得了两个金牌，排名第二，法国和日本分别领跑第一和第三。",
            "link": "https://sports.cctv.com/Paris2024/medal_list/index.shtml",
            "datePublished": "",
            "dateLastCrawled": "2024-08-25"
        },
        {
            "id": "e8674b8334a24bdcbc2b0a06be6461c0",
            "title": "2024巴黎奥运会_奖牌榜_网易体育",
            "snippet": "中国奖牌榜 项目奖牌榜 排名 国家/地区 金牌 银牌 铜牌 总数 1 美国 40 44 42 126 2 中国 40 27 24 91 3 日本 20 12 13 45 4 澳大利亚 18 19 16 53 5 法国 16 26 22 64 6 ...",
            "link": "https://sports.163.com/paris2024/medal",
            "datePublished": "",
            "dateLastCrawled": "2024-08-27"
        }
    ]
    query = "巴黎 奥运会 金牌 数量 排名"

    # gen_prompt = build_prompt_from_search_list(query,search_list)
    # print(gen_prompt)
#     top_k = 10
#     threshold = 0  # 根据实际情况调整阈值 [EN] threshold = 0 #Adjust the threshold according to the actual situation


    results = hybrid_rerank_results(query, search_list)
    print(results)