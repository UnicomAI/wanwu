import requests
import json
import os
import time
# from sklearn.feature_extraction.text import TfidfVectorizer
import re
# import jieba.posseg as pseg
import logging
import utils.es_utils as es_utils

# import model_manager
from model_manager.model_config import RerankModelConfig
from model_manager.model_config import get_model_configure
from langchain.prompts import PromptTemplate
from utils.prompts import PROMPT_TEMPLATE, CITATION_INSTRUCTION
from settings import TRUNCATE_PROMPT, CONTEXT_LENGTH, REPLACE_MINIO_DOWNLOAD_URL, MINIO_ADDRESS

logger = logging.getLogger(__name__)

def qa_weighted_rerank(query, weights, top_k, search_list_infos):
    response_info = {'code': 0, "message": "成功", "data": {"sorted_scores":[], "sorted_search_list": []}}

    if not search_list_infos:
        return response_info

    return es_utils.qa_rescore(query, weights, top_k, search_list_infos)


def get_weighted_rerank(query, weights, search_list, top_k):
    search_list_infos = {}
    for item in search_list:
        if "content_type" in item and item["content_type"] == "image":
            continue
        base_name = item["kb_name"]
        base_id = item.get("kb_id")
        user_id = item["user_id"]

        if user_id not in search_list_infos:
            search_list_infos[user_id] = {
                "base_infos": [],
                "search_list": []
            }

        search_list_infos[user_id]["base_infos"].append({"kb_name": base_name, "kb_id": base_id})
        search_list_infos[user_id]["search_list"].append(item)

    return es_utils.kb_rescore(query, weights, search_list_infos, top_k)

def get_model_rerank(query: str|dict,
                     top_n: int,
                     documents: list,
                     raw_search_list: list,
                     model_name: str,
                     endpoint_url: str,
                     api_key: str):
    headers = {"Content-Type": "application/json", "Authorization": f"Bearer {api_key}"}

    logger.info(f"rerank docList: {documents}")
    rerank_data = {
        "model": model_name,
        "query": query,
        "documents": documents,
        "top_n": top_n,
        "return_documents": True
    }
    
    # 将请求体转换为JSON字符串用于日志记录
    request_body = json.dumps(rerank_data, ensure_ascii=False, indent=2)
    logger.info(f"Rerank request details:\nURL: {endpoint_url}\nHeaders: {headers}\nBody: {request_body}")

    last_error = None
    for attempt in range(3):  # 重试3次
        try:
            # 发送请求（注意使用request_body变量会二次编码，应使用原始数据）
            response = requests.post(
                endpoint_url,
                headers=headers,
                data=json.dumps(rerank_data, ensure_ascii=False).encode('utf-8')
            )
            
            # 记录原始响应以便调试
            logger.info(f"Rerank raw response: {response.text}")
            
            if response.status_code != 200:
                raise Exception(f"HTTP {response.status_code}: {response.text}")
                
            results = json.loads(response.text)
            
            # 关键检查：确保结果字段存在
            if "results" not in results:
                raise KeyError(f"Missing 'results' field in response: {results.keys()}")
                
            # 处理正常结果
            result_data = results["results"]
            sorted_scores = []
            sorted_search_list = []
            for item in result_data[:top_n]:
                sorted_scores.append(item['relevance_score'])
                sorted_search_list.append(raw_search_list[item["index"]])
                
            return {
                "code": 0,
                "message": "",
                "data": {
                    "sorted_scores": sorted_scores,
                    "sorted_search_list": sorted_search_list
                }
            }
            
        except Exception as e:
            last_error = str(e)
            logger.error(f"Attempt {attempt+1} failed: {str(e)}")
            time.sleep(0.5 * (attempt + 1))  # 指数退避

    error_msg = f"{model_name} connection error after retries. Last error: {last_error}"
    return {"code": 1, "message": error_msg}


def model_rerank(query: dict|str,
                 top_k: int,
                 documents: list,
                 raw_search_list: list,
                 rerank_model_id: str,
                 model_config:RerankModelConfig=None):
    model_name = model_config.model_name
    model_url_suffix = "/rerank"
    is_support_multimodal = model_config.is_multimodal
    logger.info(f"model_rerank, query: {query}, is_support_multimodal: {is_support_multimodal}")
    if is_support_multimodal:
        model_url_suffix = "/multimodal-rerank"
    if isinstance(model_config, RerankModelConfig):
        model_url = model_config.endpoint_url + model_url_suffix
        api_key = model_config.api_key
    else:
        raise Exception(f"model type of model {rerank_model_id} is not rerank")

    # replace url by internal minio url
    for doc in documents:
        if isinstance(doc, dict):
            for key, value in doc.items():
                if key == "image" and not value.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in value:
                    suffix = value.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
                    doc[key] = f"http://{MINIO_ADDRESS}/{suffix}"

    return get_model_rerank(query, top_k, documents, raw_search_list, model_name, model_url, api_key)

def assemble_search_result(question, sorted_scores, search_list, threshold, return_meta, prompt_template, default_answer,
                  auto_citation):
    response_info = {'code': 0, "message": "成功", "data": {"prompt": "", "searchList": [], "score": []}}

    try:
        if not return_meta:
            for x in search_list:
                if 'meta_data' in x: x['meta_data'] = {}
                if "child_content_list" in x:
                    for item in x["child_content_list"]:
                        if "meta_data" in item:
                            item["meta_data"] = {}
        res_score = []
        res_search_list = []
        for score, doc_item in zip(sorted_scores, search_list):
            if score >= threshold:
                res_score.append(score)
                res_search_list.append(doc_item)
        response_info['data']['searchList'] = res_search_list
        response_info['data']['score'] = res_score
        if auto_citation:
            context = "\n".join([f"\n【{i + 1}^】\n{x['snippet']}" for i, x in enumerate(res_search_list)])
        else:
            context = "\n".join([x['snippet'] for x in res_search_list])
        # 判断是否临时截断 context
        if TRUNCATE_PROMPT:
            context = context[:CONTEXT_LENGTH]

        if len(prompt_template) > 0 and "{question}" in prompt_template and "{context}" in prompt_template:
            # prompt = prompt_template.replace("{question}", question).replace("{context}", context)
            formatted_prompt = PromptTemplate(
                template=prompt_template,
                input_variables=["question", "context"]
            )
            prompt = formatted_prompt.format(
                question=question,
                context=context

            )
        else:
            citation = CITATION_INSTRUCTION if auto_citation else ""
            # default_answer = DEFAULT_ANSWER_INSTRUCTION if auto_citation and default_answer else ""
            if auto_citation and default_answer:
                default_answer = "请仅基于提供的参考信息中上下文提供答案。如果提供的参考信息中的所有上下文对回答问题均无帮助，请直接输出:%s" % default_answer
            else:
                default_answer = ""

            formatted_prompt = PromptTemplate(
                template=PROMPT_TEMPLATE,
                input_variables=["citation", "default_answer", "question", "context"]
            )
            prompt = formatted_prompt.format(
                citation=citation,
                context=context,
                default_answer=default_answer,
                question=question
            )

        response_info['data']['prompt'] = prompt
        logger.info(f'context len: {len(context)}')
        logger.info(f'prompt len: {len(prompt)}')
        logger.info('content rerank请求成功')
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error('content rerank请求异常：' + repr(e))
        return response_info


def gen_rag_list(searchList, es_list):
    tmp_content = []
    search_list = []
    logger.info("searchList = %s", repr(searchList))
    logger.info("es_List = %s", repr(es_list))
    for i in searchList:
        if i["content"] in tmp_content: continue
        tmp_content.append(i["content"])
        search_list.append({"text": i["content"]})
    for i in es_list:
        if i["snippet"] in tmp_content: continue
        tmp_content.append(i["snippet"])
        search_list.append({"text": i["snippet"]})
    return search_list

def gen_raw_list(searchList, es_list):
    raw_search_list = []
    tmp_content = []

    for i in searchList:
        if i["content"] in tmp_content: continue
        content = {
            "title": i["file_name"],
            "snippet": i["content"],
            "kb_name": i["kb_name"],
            "content_id": i["content_id"],
            "meta_data": i["meta_data"]
        }
        if "is_parent" in i:
            content["is_parent"] = i["is_parent"]
        raw_search_list.append(content)
        tmp_content.append(i["content"])

    for i in es_list:
        if i["snippet"] in tmp_content: continue
        raw_search_list.append(i)
        tmp_content.append(i["snippet"])

    return raw_search_list

def gen_rerank_search_list(milvus_list, es_list, search_field="content"):
    """ 根据 search_field 生成 rerank search_list"""
    milvus_search_list = []
    es_search_list = []
    milvus_dup_content = []  # 去重
    es_dup_content = []  # 去重
    for i in milvus_list:
        if i["content"] in milvus_dup_content: continue
        milvus_search_list.append({"title": i["file_name"], "snippet": i[search_field], "kb_name": i["kb_name"],
                                   "meta_data": i["meta_data"], "content": i["content"]})
        milvus_dup_content.append(i["content"])
    for i in es_list:
        if i["snippet"] in es_dup_content: continue
        es_search_list.append({"title": i["title"], "snippet": i["snippet"], "kb_name": i["kb_name"],
                               "meta_data": i["meta_data"], "content": i["snippet"]})
        es_dup_content.append(i["snippet"])
    return milvus_search_list, es_search_list

def extract_keyword_entities(query):
    """
    提取查询中的关键实体和关键词。

    参数:
        query (str): 输入的查询文本。

    返回:
        dict: 包含两个键值对：
            - "sequence_entities": 提取的序列实体（如数字、特殊字符组合等）。
            - "keyword_entities": 提取的关键词（名词、地名、专有名词等）。
    """
    # 定义正则表达式模式，用于匹配非纯字母和非纯数字的序列
    # 匹配由 a-zA-Z0-9_- 组成的连续串（长度 >= 2），并保证前后不是这些字符（保证整片段）
    sequence_pattern = r'(?<![A-Za-z0-9_-])[A-Za-z0-9_-]{2,}(?![A-Za-z0-9_-])'
    raw_matches = re.findall(sequence_pattern, query)
    # 过滤掉纯字母或纯数字
    sequence_entities = [m for m in raw_matches if not re.fullmatch(r'[A-Za-z]+', m) and not re.fullmatch(r'\d+', m)]
    # 使用jieba.posseg进行分词和词性标注
    words = jieba.lcut(query)
    word_pos_list = [(word, pseg.lcut(word)[0].flag) for word in words]

    # 保留特定词性的词，例如名词（n）、地名（ns）、专有名词（nz）、机构名（nt）、数字（m）、数量词（mq）
    keyword_entities = [word for word, pos in word_pos_list if pos in ['eng', 'n', 'ns', 'nz', 'nt', 'm', 'mq']]

    # 返回提取结果
    return {
        "sequence_entities": sequence_entities,
        "keyword_entities": keyword_entities
    }


def get_keyword_tfidf_scores(keyword_entities, search_list):
    """
    计算关键词TF-IDF得分，带词频加权

    参数:
        keyword_entities (list): 提取的关键词列表。
        search_list (list): 搜索结果列表，每个元素是一个字典，包含"snippet"键，表示文本片段。

    返回:
        list: 每个搜索结果中关键词的TF-IDF得分列表。
    """

    # 自定义分词函数
    def jieba_tokenize(text):
        return jieba.lcut(text)  # 使用精确模式分词

    # 提取搜索结果中的文本片段
    context_list = [x["snippet"] for x in search_list]

    # 初始化TfidfVectorizer，并传入自定义分词函数
    vectorizer = TfidfVectorizer(tokenizer=jieba_tokenize, lowercase=False)

    # 计算TF-IDF矩阵
    tfidf_matrix = vectorizer.fit_transform(context_list)

    # 获取词汇表
    words = vectorizer.get_feature_names_out()
    word_index = {w: idx for idx, w in enumerate(words)}
    keyword_scores = []
    for i, context in enumerate(context_list):
        score = 0
        for keyword in keyword_entities:
            if keyword in word_index:
                idx = word_index[keyword]
                freq_weight = context.count(keyword)
                score += tfidf_matrix[i, idx] * (1 + 0.1 * freq_weight)
        keyword_scores.append(score)

    # 返回结果
    return keyword_scores


def get_sequence_entities_scores(sequence_entities, search_list):
    """
    计算序列实体得分，按长度和位置加权

    参数:
        sequence_entities (list): 提取的序列实体列表。
        search_list (list): 搜索结果列表，每个元素是一个字典，包含"snippet"键，表示文本片段。

    返回:
        list: 每个搜索结果中序列实体的得分列表。
    """
    sn_scores = []
    context_list = [x["snippet"] for x in search_list]
    for i, context in enumerate(context_list):
        score = 0
        for s_n in sequence_entities:
            if s_n in context:
                # length_weight = min(0.2, 0.02 * len(s_n))
                length_weight = 0.12
                position_bonus = 0.2 if context.startswith(s_n) else 0
                score += length_weight + position_bonus
        sn_scores.append(score)
    return sn_scores


def hybrid_term_weight_rerank(query, search_list, scores=[], top_k=5, term_weight_coefficient=1, threshold=0):
    """
    根据混合关键实体和关键词的权重对搜索结果进行重排序。

    参数:
        query (str): 输入的查询文本。
        search_list (list): 搜索结果列表，每个元素是一个字典，包含"snippet"键，表示文本片段。
        scores (list, optional): 初始得分列表，默认为空列表。
        top_k (int, optional): 返回的顶部结果数量，默认为5。
        term_weight_coefficient (float, optional): 关键词权重系数，默认为1。
        threshold (float, optional): 得分阈值，低于该值的结果将被过滤，默认为0.4。

    返回:
        tuple: 包含两个元素：
            - res_search_list: 重排序后的搜索结果列表。
            - res_score: 对应的得分列表。
    """
    if not scores:  # 若无scores，初始化一个
        scores = [0 for _ in search_list]

    # 提取关键实体和关键词
    result = extract_keyword_entities(query)
    sequence_entities = result["sequence_entities"]
    keyword_entities = result["keyword_entities"]

    # 计算序列实体和关键词的得分
    sn_scores = get_sequence_entities_scores(sequence_entities, search_list)
    kw_scores = get_keyword_tfidf_scores(keyword_entities, search_list)

    term_weight_scores = [a + b for a, b in zip(sn_scores, kw_scores)]
    # 使用列表推导式和zip函数计算综合得分
    rerank_weight_coefficient = 1 + (1 - term_weight_coefficient) * 0.3
    hybrid_scores = [(a * rerank_weight_coefficient + b * term_weight_coefficient) for a, b in
                     zip(scores, term_weight_scores)]

    # 重新按得分大小排序
    sorted_pairs = sorted(zip(hybrid_scores, search_list), key=lambda x: x[0], reverse=True)

    # 分别提取排序后的search_list和score
    res_score = [min(1, pair[0]) for pair in sorted_pairs if pair[0]>=threshold]  # 得分限制在0-1之间
    res_search_list = [pair[1] for pair in sorted_pairs][:len(res_score)]

    # 返回重排序结果
    return res_search_list[:top_k], res_score[:top_k]
