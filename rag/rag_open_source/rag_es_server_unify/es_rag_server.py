# -*- coding: utf-8 -*-
import json
import time
import copy
import traceback
from copy import deepcopy

from log.logger import logger
from datetime import datetime
from flask import Flask, request, Response

from settings import EMBEDDING_BATCH_SIZE
from settings import INDEX_NAME_PREFIX, SNIPPET_INDEX_NAME_PREFIX, KBNAME_MAPPING_INDEX
import utils.es_util as es_ops
import utils.meta_util as meta_ops
import utils.kb_info as kb_info_ops
import utils.qa_util as qa_ops
import utils.mapping_util as es_mapping
from utils import emb_util
from utils.util import get_qa_index_name
from utils.http_util import validate_request
from utils.otel import init_tracer
from utils.trace import register_tracing
from model.model_manager import is_multimodal_model

# 初始化 OpenTelemetry（必须在 Flask app 创建之前）
init_tracer("rag-wanwu-es-proxy")

# 自动 instrument requests 库（出站 HTTP 传播 traceparent）
from opentelemetry.instrumentation.requests import RequestsInstrumentor

RequestsInstrumentor().instrument()

app = Flask(__name__)

# 自动 instrument Flask（入站 HTTP 提取 traceparent，创建 Span）
from opentelemetry.instrumentation.flask import FlaskInstrumentor

FlaskInstrumentor().instrument_app(app)

# 添加日志记录（含 trace_id/span_id 关联）
register_tracing(app)

def batch_list(lst: list, batch_size=32):
    """ 切分生成器 """
    for i in range(0, len(lst), batch_size):
        yield lst[i:i + batch_size]


def log_exception_with_trace(e, msg=""):
    stack_trace = traceback.format_exc()

    log_content = (
        f"【异常捕获】\n"
        f"Context Info: {msg}\n"
        f"Error Message: {str(e)}\n"
        f"Stack Trace:\n{stack_trace}"
    )

    logger.warning(log_content)


@app.route('/rag/kn/init_kb', methods=['POST'])
@validate_request
def init_kb(request_json=None):
    """ ES 模拟RAG主控 初始化 init_kb 接口"""
    logger.info("--------------------------启动向量库初始化---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    content_index_name = 'content_control_' + index_name
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    kb_name = kb_info.get("kb_name")
    kb_id = kb_info.get("kb_id")  # 必须字段
    embedding_model_id = request_json.get("embedding_model_id")  # 必须字段
    enable_knowledge_graph = request_json.get("enable_knowledge_graph", False)
    is_multimodal = request_json.get("is_multimodal", False)
    userId_kb_names = []
    dense_dim = 1024
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        judge_time1 = time.time()

        es_ops.create_index_if_not_exists(content_index_name, mappings=es_mapping.cc_mappings)  # 确保 主控表 已创建
        es_ops.create_index_if_not_exists(KBNAME_MAPPING_INDEX, mappings=es_mapping.uk_mappings)  # 确保 KBNAME_MAPPING_INDEX 已创建
        es_ops.create_index_if_not_exists(index_name, mappings=es_mapping.mappings)

        kb_names = kb_info_ops.get_uk_kb_name_list(KBNAME_MAPPING_INDEX, userId)  # 从映射表中获取
        logger.info(f"当前用户:{userId},共有知识库：{len(kb_names)}个，分别为{kb_names}")
        judge_time2 = time.time()
        judge_time = judge_time2 - judge_time1
        logger.info(f"--------------------------查询kb_map时间:{judge_time}---------------------------\n")
        if kb_name in kb_names:
            raise RuntimeError(f"已存在同名知识库{kb_name}")

        utc_now = datetime.utcnow()
        formatted_time = utc_now.strftime('%Y-%m-%d %H:%M:%S')
        uk_data = [
            {"index_name": index_name, "userId": userId, "kb_name": kb_name,
             "creat_time": formatted_time, "kb_id": kb_id, "embedding_model_id": embedding_model_id,
             "enable_graph": enable_knowledge_graph, "is_multimodal": is_multimodal}
        ]
        kb_info_ops.bulk_add_uk_index_data(KBNAME_MAPPING_INDEX, uk_data)
        # ====== 新建完成，需要校验 kb_id 是否写入成功 ======
        kb_id_list = kb_info_ops.get_uk_kb_id_list(KBNAME_MAPPING_INDEX, userId)
        if kb_id not in kb_id_list:
            raise RuntimeError("ini知识库失败，ES写入失败")

        # 新建成功，返回
        logger.info(f"当前用户:{userId},知识库:{kb_name},save_kb_id:{kb_id}")
        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_name},ini知识库的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="init_kb")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_name},ini知识库的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/add', methods=['POST'])
@validate_request
def add_vector_data(request_json=None):
    """ 往 ES 中建向量索引数据，当前方法要校验索引名 kb_name 是否已存在"""
    logger.info("--------------------------启动数据添加---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    content_index_name = 'content_control_' + index_name
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    kb_name = kb_info.get("kb_name")
    kb_id = kb_info.get("kb_id")
    embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(userId, kb_name)
    doc_list = request_json.get("data")
    userId_kb_names = []
    cc_doc_list = []  # content主控表的数据
    cc_duplicate_list = []
    if not kb_id:
        result = {
            "code": 1,
            "message": "kb_info.kb_id is required"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        return jsonarr
    # # **************** 校验 kb_name 是否已经初始化过 ****************
    # userId_kb_ids = kb_info.get_uk_kb_id_list(KBNAME_MAPPING_INDEX, userId)  # 从映射表中获取
    # if kb_id not in userId_kb_ids:
    #     result = {
    #         "code": 1,
    #         "message": f"{kb_id}知识库不存在"
    #     }
    #     jsonarr = json.dumps(result, ensure_ascii=False)
    #     logger.info(f"{userId},/rag/kn/add的接口返回结果为：{jsonarr},userId_kb_names:{userId_kb_ids}")
    #     return jsonarr
    # # **************** 校验 kb_name 是否已经初始化过 ****************
    # ========= 将 content 主控表数据过滤好 =============
    for doc in copy.deepcopy(doc_list):
        cc_str = str(doc["content"]) + doc["file_name"] + str(doc["meta_data"]["chunk_current_num"])
        if cc_str not in cc_duplicate_list:
            # 提取的多模态数据content_type 是 image，就不写 content 主控表，只写向量表
            if "content_type" in doc and doc["content_type"] == "image":
                continue
            doc.pop("embedding_content")  # 去掉不需要的字段
            doc["content_type"] = "text"  # 主控表都是 text
            doc["status"] = True  # 初始化启停状态
            if "is_parent" in doc:
                doc["is_parent"] = True
                doc["child_chunk_total_num"] = doc["meta_data"]["child_chunk_total_num"]
                doc["meta_data"].pop("child_chunk_current_num")
                doc["meta_data"].pop("child_chunk_total_num")
            cc_doc_list.append(doc)
            cc_duplicate_list.append(cc_str)
    # ========= 将 content 主控表数据过滤好 =============
    for doc in doc_list:
        doc.pop("labels", None)  # 去掉不需要的字段, labels 只写content 主控表

    try:
        # ========= 将 embedding_content 编码好向量 =============
        content_vector_exist = False
        mapping_properties = {}
        field_name = ""
        for batch_doc in batch_list(doc_list, batch_size=EMBEDDING_BATCH_SIZE):
            if is_multimodal_model(embedding_model_id): # 多模态模型则按多模态去编码
                inputs = []
                for x in batch_doc:
                    if x.get("content_type", "text") == "image":
                        inputs.append({"image": x["embedding_content"]})
                    else:
                        inputs.append({"text": x["embedding_content"]})
                res = emb_util.get_multimodal_embs(inputs, embedding_model_id=embedding_model_id)
            else:  # 非多模态知识库则按之前的文本去编码
                res = emb_util.get_embs([x["embedding_content"] for x in batch_doc], embedding_model_id=embedding_model_id)
            dense_vector_dim = len(res["result"][0]["dense_vec"]) if res["result"] else 1024
            field_name = f"q_{dense_vector_dim}_content_vector"
            if dense_vector_dim == 1024:
                # 兼容老索引，避免创建两个1024 dim的向量字段
                if not mapping_properties:
                    content_vector_exist, mapping_properties = es_ops.is_field_exist(index_name, "content_vector")
                if content_vector_exist:
                    logger.info(f"es 索引 {index_name} 字段 {field_name} 存在，回退到默认字段 content_vector")
                    field_name = "content_vector"

            for i, x in enumerate(batch_doc):
                if len(batch_doc) != len(res["result"]):
                    raise RuntimeError(f"Error getting embeddings:{batch_doc}")
                x[field_name] = res["result"][i]["dense_vec"]

        # 过滤掉向量字段值为 None 的数据
        doc_list = [
            doc for doc in doc_list
            if field_name in doc and doc[field_name] is not None
        ]
        # 检查过滤后是否还有数据需要写入
        if not doc_list:
            logger.warning(f"kb_id: {kb_id} 过滤后没有有效的向量数据可供存储")
        else:
            # ========= 将 embedding_content 编码好向量 =============
            es_result = es_ops.bulk_add_index_data(index_name, kb_id, doc_list)  # 注意 存储的时候传入 kb_id
            logger.info(f"{es_result}")
            es_cc_result = es_ops.bulk_add_cc_index_data(content_index_name, kb_id, cc_doc_list)  # 注意 存储的时候传入 kb_id
            if es_result["success"] and es_cc_result["success"]:  # bulk_add_index_data 成功了则返回
                result = {
                    "code": 0,
                    "message": "success"
                }
                jsonarr = json.dumps(result, ensure_ascii=False)
                logger.info(f"当前用户:{userId},知识库:{kb_name},add的接口返回结果为：{jsonarr}")
                return jsonarr
            else:  # bulk_add_index_data 报错了则返回错误信息
                result = {
                    "code": 1,
                    "message": es_result.get("error", "") + es_cc_result.get("error", "")
                }
                jsonarr = json.dumps(result, ensure_ascii=False)
                logger.info(f"当前用户:{userId},知识库:{kb_name},add的接口返回结果为：{jsonarr}")
                return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="add_vector_data")
        result = {
            "code": 2,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_name},add的接口返回结果为：{jsonarr}")
        return jsonarr
    finally:
        logger.info(f"user_id: {userId}, kb_name: {kb_name}, bulk_add end")

@app.route('/rag/kn/get_kb_info', methods=['POST'])
@validate_request
def get_kb_info(request_json=None):
    """ 查询知识库详情"""
    logger.info("-----------------------启动知识库info查询-------------------\n")
    userId = request_json.get("userId")
    kb_name = request_json.get("kb_name")
    try:
        # ******** 先检查 是否有新建 index ***********
        es_ops.create_index_if_not_exists(KBNAME_MAPPING_INDEX, mappings=es_mapping.uk_mappings)  # 确保 KBNAME_MAPPING_INDEX 已创建
        kb_info = kb_info_ops.get_uk_kb_info(userId, kb_name)
        logger.info(f"当前用户:{userId},知识库:{kb_name}, kb_info: {kb_info}")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "kb_info": kb_info
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库info查询接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, "get_kb_info")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库详情查询的接口返回结果为：{jsonarr}")
        return jsonarr

@app.route('/rag/kn/list_kb_names', methods=['POST'])
@validate_request
def list_kb_names(request_json=None):
    logger.info("--------------------------启动知识库查询---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    try:
        # ******** 先检查 是否有新建 index ***********
        es_ops.create_index_if_not_exists(KBNAME_MAPPING_INDEX, mappings=es_mapping.uk_mappings)  # 确保 KBNAME_MAPPING_INDEX 已创建
        is_exists = es_ops.create_index_if_not_exists(index_name, mappings=es_mapping.mappings)
        # ******** 先检查 是否有新建 index ***********
        # userId_kb_names = es_ops.get_kb_name_list(index_name) # 不使用此方式
        userId_kb_names = kb_info_ops.get_uk_kb_name_list(KBNAME_MAPPING_INDEX, userId)  # 从映射表中获取
        logger.info(f"/rag/kn/list_kb_names:用户{index_name}共有{len(userId_kb_names)}个知识库")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "kb_names": userId_kb_names
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库查询的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, "list_kb_names")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库查询的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/list_file_names', methods=['POST'])
@validate_request
def list_file_names(request_json=None):
    logger.info("--------------------------启动文件查询---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        # **************** 校验 kb_name 是否已经初始化过 ****************
        userId_kb_ids = kb_info_ops.get_uk_kb_id_list(KBNAME_MAPPING_INDEX, userId)  # 从映射表中获取
        if kb_id not in userId_kb_ids:
            result = {
                "code": 1,
                "message": f"{kb_id}知识库不存在"
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(f"{userId},/rag/kn/list_file_names的接口返回结果为：{jsonarr},userId_kb_names:{userId_kb_ids}")
            return jsonarr
        # **************** 校验 kb_name 是否已经初始化过 ****************
        file_names = es_ops.get_file_name_list(index_name, kb_id)
        logger.info(f"用户{index_name}的知识库{kb_id}共有{len(file_names)}个文件")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "file_names": file_names
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件查询的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="查询文件名称时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件查询的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/list_file_names_after_filtering', methods=['POST'])
@validate_request
def list_file_names_after_filtering(request_json=None):
    logger.info("--------------------------启动文件过滤查询---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    filtering_conditions = request_json.get("filtering_conditions")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id}, filtering_conditions: {filtering_conditions}")

        final_conditions = []
        for condition in filtering_conditions:
            if condition["filtering_kb_name"] == display_kb_name:
                condition["filtering_kb_name"] = kb_id
                final_conditions.append(deepcopy(condition))
        file_names = []
        if final_conditions:
            file_names = meta_ops.search_with_doc_meta_filter(index_name, final_conditions)
        logger.info(f"用户{index_name}的知识库{display_kb_name}过滤后共有{len(file_names)}个文件")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "file_names": file_names
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件过滤查询的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="过滤查询文件名称时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件过滤查询的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/list_file_download_links', methods=['POST'])
@validate_request
def list_file_download_links(request_json=None):
    logger.info("--------------------------启动获取知识库里所有文档的下载链接查询---------------------------")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        file_names = es_ops.get_file_download_link_list(index_name, kb_id)
        logger.info(f"用户{index_name}的知识库{kb_id}共有{len(file_names)}个文件的下载链接")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "file_download_links": file_names
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件下载链接查询的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="查询文件下载链接时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},文件下载链接查询的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/search', methods=['POST'])
@validate_request
def es_knn_search(request_json=None):
    """ 多知识库 KNN检索 """
    logger.info("--------------------------启动向量库检索---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    content_index_name = 'content_control_' + index_name
    userId = request_json.get("userId")
    kb_infos = request_json.get("kb_infos")  # list of {kb_name, kb_id}
    top_k = request_json.get("topk", 10)
    query = request_json.get("question")
    min_score = request_json.get("threshold", 0)
    filter_file_name_list = request_json.get("filter_file_name_list", [])
    metadata_filtering_conditions = request_json.get("metadata_filtering_conditions", [])
    enable_vision = request_json.get("enable_vision", [])
    attachment_files = request_json.get("attachment_files", [])
    kb_id_2_kb_name = {}
    emb_id2kb_names = {}
    logger.info(f"用户:{index_name},请求查询的kb_infos为:{kb_infos}")
    logger.info(f"用户请求的query为:{query}")
    try:
        if not kb_infos:
            raise ValueError("kb_infos is required")
        filtering_conditions = {}
        for condition in metadata_filtering_conditions:
            kb_name = condition["filtering_kb_name"]
            filtering_conditions[kb_name] = condition

        final_conditions = []
        for kb_info_item in kb_infos:
            kb_name = kb_info_item["kb_name"]
            kb_id = kb_info_item["kb_id"]
            if not kb_id:
                raise ValueError(f"kb_id is required for kb_name: {kb_name}")
            kb_id_2_kb_name[kb_id] = kb_name
            if kb_name in filtering_conditions:
                condition = filtering_conditions[kb_name]
                condition["filtering_kb_name"] = kb_id
                final_conditions.append(deepcopy(condition))
            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(userId, kb_name)
            if embedding_model_id not in emb_id2kb_names:
                emb_id2kb_names[embedding_model_id] = []
            emb_id2kb_names[embedding_model_id].append(kb_id)
        meta_filter_file_name_list = []
        if final_conditions:
            meta_filter_file_name_list = meta_ops.search_with_doc_meta_filter(content_index_name, final_conditions)
            logger.info(f"用户请求的query为:{query}, filter_file_name_list: {filter_file_name_list}, meta_filter_file_name_list: {meta_filter_file_name_list}")
            if len(meta_filter_file_name_list) == 0:
                result = {
                    "code": 0,
                    "message": "success",
                    "data": {
                        "search_list": [],
                        "scores": []
                    }
                }
                jsonarr = json.dumps(result, ensure_ascii=False)
                logger.info(f"当前用户:{userId},知识库:{kb_infos},query:{query},向量库检索的接口返回结果为：{jsonarr}")
                return jsonarr

        if meta_filter_file_name_list:
            filter_file_name_list = filter_file_name_list + meta_filter_file_name_list
        # ============= 先检查 kb_names 是不是都存在 =============
        # ============= 开始检索召回 ===============
        search_list = []
        scores = []
        for embedding_model_id, kb_names in emb_id2kb_names.items():
            logger.info(f"用户:{index_name},请求查询的kb_names为:{kb_names},embedding_model_id:{embedding_model_id}")
            result_dict = es_ops.search_data_knn_recall(index_name, kb_names, query, top_k, min_score,
                                                        filter_file_name_list=filter_file_name_list,
                                                        embedding_model_id=embedding_model_id,
                                                        enable_vision=enable_vision,
                                                        attachment_files=attachment_files)
            search_list.extend(result_dict["search_list"])
            scores.extend(result_dict["scores"])

        if len(search_list) > top_k:
            # 合并search_list和scores，按score降序排序
            combined_results = list(zip(search_list, scores))
            combined_results.sort(key=lambda x: x[1], reverse=True)

            # 取前top_k个结果
            top_results = combined_results[:top_k]
            search_list = [item[0] for item in top_results]
            scores = [item[1] for item in top_results]

        for item in search_list:  # 将 kb_id 转换为 kb_name
            item["kb_id"] = item["kb_name"]
            item["kb_name"] = kb_id_2_kb_name[item["kb_name"]]

        result = {
            "code": 0,
            "message": "success",
            "data": {
                "search_list": search_list,
                "scores": scores
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_infos},query:{query},向量库检索的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="查询知识库时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_infos},query:{query},向量库检索的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/del_kb', methods=['POST'])
@validate_request
def del_kb(request_json=None):
    logger.info("--------------------------启动知识库删除---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    content_index_name = 'content_control_' + index_name
    file_index_name = 'file_control_' + index_name
    community_report_name = 'community_report_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        es_result = es_ops.delete_data_by_kbname(index_name, kb_id)
        es_cc_result = es_ops.delete_data_by_kbname(content_index_name, kb_id)  # 主控表也需要删除
        es_file_result = es_ops.delete_data_by_kbname(file_index_name, kb_id)
        es_uk_result = es_ops.delete_uk_data_by_kbname(userId, display_kb_name)  # uid索引映射表需要删除,传display_kb_name
        es_cr_result = es_ops.delete_data_by_kbname(community_report_name, kb_id)
        if es_result["success"] and es_cc_result["success"] and es_uk_result["success"] and es_file_result["success"] and es_cr_result["success"]:  # delete_data_by_kbname 成功了则返回
            logger.info(f"用户{index_name},对应的{kb_id}记录删除成功")
            result = {
                "code": 0,
                "message": "success"
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},知识库删除的接口返回结果为：{jsonarr},{es_result},{es_cc_result},{es_uk_result},{es_cr_result}")
            return jsonarr
        else:
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},知识库删除时发生错误：{es_result},{es_cc_result},{es_uk_result},{es_file_result},{es_cr_result}")
            result = {
                "code": 1,
                "message": es_result.get("error", "") + es_cc_result.get("error", "") + es_file_result.get("error", "") + es_cr_result.get("error", "")
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(f"当前用户:{userId},知识库:{kb_id},知识库删除的接口返回结果为：{jsonarr}")
            return jsonarr

    except Exception as e:
        logger.info(f"用户{index_name},对应的{kb_id}知识库删除时发生错误：{e}")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},知识库删除的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/del_files', methods=['POST'])
@validate_request
def del_files(request_json=None):
    logger.info("--------------------------启动文件删除---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_names = request_json.get("file_names")
    content_index_name = 'content_control_' + index_name
    file_index_name = 'file_control_' + index_name

    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")

        # ********* 单独删除，获取每一个文件状态
        success = []
        failed = []
        for file in file_names:
            es_result = es_ops.delete_data_by_kbname_file_name(index_name, kb_id, file)
            es_cc_result = es_ops.delete_data_by_kbname_file_name(content_index_name, kb_id, file)
            es_file_result = es_ops.delete_data_by_kbname_file_name(file_index_name, kb_id, file)
            if es_result["success"] and es_cc_result["success"] and es_file_result["success"]:  # delete_data_by_kbname_file_names 成功了则返回
                logger.info(f"当前用户{index_name}的知识库{kb_id}删除的文档为：{file}")
                success.append(file)
            else:
                logger.info(
                    f"当前用户:{userId},知识库:{kb_id},file_names:{file_names},文件删除时发生错误：{es_result}")
                result = {
                    "code": 1,
                    "message": es_result.get("error", "") + es_cc_result.get("error", "") + es_file_result.get("error", "")
                }
                jsonarr = json.dumps(result, ensure_ascii=False)
                logger.info(
                    f"当前用户:{userId},知识库:{kb_id},file_names:{file_names},知识库删除的接口返回结果为：{jsonarr}")
                return jsonarr

        # ======== 没有报错，则返回成功 ========
        failed = [file for file in file_names if file not in success]
        logger.info(f"----------当前用户:{userId},知识库{kb_id}完成{file_names}的delete--------------")
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "success": success,
                "failed": failed
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},file_names:{file_names},文件删除的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e)
        logger.info(f"知识库{kb_id},{file_names},在文件删除时发生错误：{e}")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},file_names:{file_names},文件删除的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/get_content_list', methods=['POST'])
@validate_request
def get_content_list(request_json=None):
    """ 获取 主控表中 知识片段的分页展示 """
    logger.info("--------------------------获取主控表中知识片段的分页展示---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    page_size = request_json.get("page_size")
    search_after = request_json.get("search_after")
    content_type = request_json.get("content_type", "text")
    query_text = request_json.get("query_text")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},请求的kb_id为:{kb_id},file_name:{file_name},page_size:{page_size},search_after:{search_after}")
        searched_index_name = ""
        if content_type == "text":
            searched_index_name = 'content_control_' + index_name
        elif content_type == "community_report":
            searched_index_name = 'community_report_' + index_name
        content_result = es_ops.get_file_content_list(searched_index_name, kb_id, file_name, page_size, search_after,
                                                       query_text=query_text)
        content_list = content_result["content_list"]
        for content in content_list:
            content["kb_id"] = kb_id
            content["kb_name"] = display_kb_name
            if "is_parent" in content and content["is_parent"]:
                child_result = es_ops.get_child_contents(index_name, kb_id, content["content_id"])
                content["child_chunk_total_num"] = child_result["child_chunk_total_num"]
        result = {
            "code": 0,
            "message": "success",
            "data": content_result
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},page_size:{page_size},search_after:{search_after},知识片段分页查询的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="获取主控表中知识片段的分页展示时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},获取主控表中知识片段的分页展示的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/get_child_content_list', methods=['POST'])
@validate_request
def get_child_content_list(request_json=None):
    """ 获取子片段"""
    logger.info("--------------------------获取子片段---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    chunk_id = request_json.get("chunk_id")
    child_chunk_current_num = request_json.get("child_chunk_current_num", None)
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},请求的kb_id为:{kb_id},file_name:{file_name},chunk_id:{chunk_id}")
        content_result = es_ops.get_child_contents(index_name, kb_id, chunk_id, child_chunk_current_num)
        for child_content in content_result["child_content_list"]:
            child_content["kb_id"] = kb_id
            child_content["kb_name"] = display_kb_name
        result = {
            "code": 0,
            "message": "success",
            "data": content_result
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name}, chunk_id:{chunk_id},子分段查询的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="获取子分段时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},获取子分段的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/update_child_chunk', methods=['POST'])
@validate_request
def update_child_chunk(request_json=None):
    logger.info("--------------------------更新知识库子段数据---------------------------\n")
    userId = request_json.get("userId")
    index_name = INDEX_NAME_PREFIX + userId
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(userId, display_kb_name)
    snippet_index_name = SNIPPET_INDEX_NAME_PREFIX + userId.replace('-', '_')
    chunk_id = request_json.get("chunk_id")
    child_chunk = request_json.get("child_chunk")
    chunk_current_num = request_json.get("chunk_current_num")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        child_content = child_chunk["child_content"]
        child_chunk_current_num = child_chunk["child_chunk_current_num"]
        index_update_data = {
            "embedding_content": child_content,
        }
        logger.info(f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id}, chunk_id: {chunk_id}, "
                    f"chunk_current_num: {chunk_current_num}, child_chunk: {child_chunk}")

        # 更新子分段前，先删除旧的图片向量
        es_image_delete_result = es_ops.delete_image_chunks(index_name, kb_id, chunk_current_num, [child_chunk_current_num])
        logger.info(f"用户:{userId},知识库:{kb_id}, 更新子分段前删除旧图片向量结果: {es_image_delete_result}")

        if is_multimodal_model(embedding_model_id):  # 多模态模型则按多模态去编码
            inputs = [{"text": child_content}]
            res = emb_util.get_multimodal_embs(inputs, embedding_model_id=embedding_model_id)
        else:  # 非多模态知识库则按之前的文本去编码
            res = emb_util.get_embs([child_content], embedding_model_id=embedding_model_id)

        dense_vector_dim = len(res["result"][0]["dense_vec"]) if res["result"] else 1024
        field_name = f"q_{dense_vector_dim}_content_vector"
        if dense_vector_dim == 1024:
            # 兼容老索引，避免创建两个1024 dim的向量字段
            content_vector_exist, mapping_properties = es_ops.is_field_exist(index_name, "content_vector")
            if content_vector_exist:
                logger.info(f"es 索引 {index_name} 字段 {field_name} 存在，回退到默认字段 content_vector")
                field_name = "content_vector"

        index_update_data[field_name] = res["result"][0]["dense_vec"]
        snippet_index_update_data = {
            "snippet": child_content,
        }

        # cc index的content id == chunk id
        index_update_actions = es_ops.get_index_update_content_actions(index_name, kb_id, chunk_id, chunk_current_num,
                                                                       child_chunk_current_num, index_update_data)

        snippet_index_update_actions = es_ops.get_index_update_content_actions(snippet_index_name, kb_id, chunk_id,
                                                                               chunk_current_num, child_chunk_current_num, snippet_index_update_data)

        update_actions = {
            index_name: index_update_actions,
            snippet_index_name: snippet_index_update_actions
        }
        result = es_ops.update_index_data(update_actions)
        json_arr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},更新知识库元数据的接口返回结果为：{json_arr}")
        return json_arr
    except Exception as e:
        log_exception_with_trace(e, "更新知识库元数据时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        json_arr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{display_kb_name},更新知识库元数据的接口返回结果为：{json_arr}")
        return json_arr


@app.route('/rag/kn/update_file_metas', methods=['POST'])
@validate_request
def update_file_metas(request_json=None):
    logger.info("--------------------------更新知识库元数据---------------------------\n")
    userId = request_json.get("userId")
    index_name = INDEX_NAME_PREFIX + userId
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    update_datas = request_json.get("update_datas")
    file_index_name = 'file_control_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id}, update_datas: {update_datas}")

        # 兼容老版本，没有file index的需要创建
        es_ops.create_index_if_not_exists(file_index_name, mappings=es_mapping.mappings)
        result = meta_ops.update_file_metas(userId, kb_id, update_datas)
        json_arr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},更新知识库元数据的接口返回结果为：{json_arr}")
        return json_arr
    except Exception as e:
        log_exception_with_trace(e, msg="更新知识库元数据时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        json_arr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{display_kb_name},更新知识库元数据的接口返回结果为：{json_arr}")
        return json_arr

@app.route('/rag/kn/batch_delete_chunks', methods=['POST'])
@validate_request
def batch_delete_chunks(request_json=None):
    logger.info("--------------------------根据fileName和chunk_ids删除分段---------------------------\n")
    userId = request_json.get("userId")
    index_name = INDEX_NAME_PREFIX + userId
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    chunk_ids = request_json.get("chunk_ids")
    content_index_name = 'content_control_' + index_name
    snippet_index_name = SNIPPET_INDEX_NAME_PREFIX + userId.replace('-', '_')
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, chunk_ids: {chunk_ids}")

        # 删除分段前，先获取要删除分段的chunk_current_num列表，用于删除关联的图片向量
        chunk_current_nums = []
        contents = es_ops.get_contents_by_ids(content_index_name, kb_id, chunk_ids)
        for content in contents:
            if "meta_data" in content and "chunk_current_num" in content["meta_data"]:
                chunk_current_nums.append(content["meta_data"]["chunk_current_num"])

        es_result = es_ops.delete_chunks_by_content_ids(index_name, kb_id, chunk_ids)
        es_cc_result = es_ops.delete_chunks_by_content_ids(content_index_name, kb_id, chunk_ids)  # 主控表也需要删除
        es_snippet_result = es_ops.delete_chunks_by_content_ids(snippet_index_name, kb_id, chunk_ids)

        # 删除关联的图片向量
        if chunk_current_nums:
            es_image_result = es_ops.delete_image_chunks(index_name, kb_id, chunk_current_nums)
            logger.info(f"用户:{userId},知识库:{kb_id}, 删除图片向量结果: {es_image_result}")

        if es_result["success"] and es_cc_result["success"] and es_snippet_result["success"]:
            logger.info(f"用户{index_name},对应的知识库{kb_id}, chunks: {chunk_ids}记录分段删除成功")
            result = {
                "code": 0,
                "message": "success",
                "data": {
                    "success_count": es_cc_result["deleted"]
                }
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},chunks:{chunk_ids}, 分段删除的接口返回结果为：{jsonarr},{es_result},{es_cc_result},{es_snippet_result}")
            return jsonarr
        else:
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},chunks:{chunk_ids}, 分段删除时发生错误：{es_result},{es_cc_result},{es_snippet_result}")
            result = {
                "code": 1,
                "message": es_result.get("error", "") + es_cc_result.get("error", "") + es_snippet_result.get("error", ""),
                "data": {
                    "success_count": 0
                }
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(f"当前用户:{userId},知识库:{kb_id},chunks:{chunk_ids}, 分段删除的接口返回结果为：{jsonarr}")
            return jsonarr

    except Exception as e:
        logger.info(f"用户{index_name},对应的知识库:{kb_id},chunks:{chunk_ids}, 分段删除时发生错误：{e}")
        result = {
            "code": 1,
            "message": str(e),
            "data": {
                "success_count": 0
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},chunks:{chunk_ids}, 分段删除的接口返回结果为：{jsonarr}")
        return jsonarr

@app.route('/rag/kn/delete_child_chunks', methods=['POST'])
@validate_request
def delete_child_chunks(request_json=None):
    logger.info("--------------------------删除子分段---------------------------\n")
    userId = request_json.get("userId")
    index_name = INDEX_NAME_PREFIX + userId
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    chunk_id = request_json.get("chunk_id")
    chunk_current_num = request_json.get("chunk_current_num")
    child_chunk_current_nums = request_json.get("child_chunk_current_nums")
    snippet_index_name = SNIPPET_INDEX_NAME_PREFIX + userId.replace('-', '_')

    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, "
            f"chunk_id: {chunk_id}, chunk_current_num: {chunk_current_num}, "
            f"child_chunk_current_nums: {child_chunk_current_nums}")

        es_result = es_ops.delete_child_chunks(index_name, kb_id, chunk_id, chunk_current_num, child_chunk_current_nums)
        es_snippet_result = es_ops.delete_child_chunks(snippet_index_name, kb_id, chunk_id, chunk_current_num, child_chunk_current_nums)

        # 删除关联的子分段图片向量
        es_image_result = es_ops.delete_image_chunks(index_name, kb_id, chunk_current_num, child_chunk_current_nums)
        logger.info(f"用户:{userId},知识库:{kb_id}, 删除子分段图片向量结果: {es_image_result}")

        if es_result["success"] and es_snippet_result["success"]:
            logger.info(f"用户{index_name},对应的知识库{kb_id}, chunk: {chunk_id}, "
                        f"child_chunk_current_nums: {child_chunk_current_nums} 记录子分段删除成功")
            result = {
                "code": 0,
                "message": "success"
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},chunk:{chunk_id}, child_chunk_current_nums: {child_chunk_current_nums} "
                f"子分段删除的接口返回结果为：{jsonarr},{es_result},{es_snippet_result}")
            return jsonarr
        else:
            logger.info(
                f"当前用户:{userId},知识库:{kb_id},chunk:{chunk_id}, child_chunk_current_nums: {child_chunk_current_nums} "
                f"子分段删除时发生错误：{es_result},{es_snippet_result}")
            result = {
                "code": 1,
                "message": es_result.get("error", "") + es_snippet_result.get("error", ""),
            }
            jsonarr = json.dumps(result, ensure_ascii=False)
            logger.info(f"当前用户:{userId},知识库:{kb_id},chunk:{chunk_id}, "
                        f"child_chunk_current_nums: {child_chunk_current_nums} 子分段删除的接口返回结果为：{jsonarr}")
            return jsonarr

    except Exception as e:
        logger.info(f"用户{index_name},对应的知识库:{kb_id},chunk:{chunk_id}, "
                    f"child_chunk_current_nums: {child_chunk_current_nums} 子分段删除时发生错误：{e}")
        result = {
            "code": 1,
            "message": str(e),
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_id},chunk:{chunk_id}, "
                    f"child_chunk_current_nums: {child_chunk_current_nums} 子分段删除的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/update_chunk_labels', methods=['POST'])
@validate_request
def update_chunk_labels(request_json=None):
    logger.info("--------------------------根据fileName和chunk_id更新标签---------------------------\n")
    userId = request_json.get("userId")
    index_name = INDEX_NAME_PREFIX + userId
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    chunk_id = request_json.get("chunk_id")
    labels = request_json.get("labels")
    content_index_name = 'content_control_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, chunk_id: {chunk_id}, labels: {labels}")

        index_actions = {
            content_index_name: es_ops.get_cc_index_update_label_actions(content_index_name, kb_id, file_name, labels, chunk_id=chunk_id)
        }
        result = es_ops.update_chunk_labels(index_actions)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},根据fileName和chunk_id更新知识库chunk 标签的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="根据fileName和chunk_id更新知识库chunk 标签时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},根据fileName和chunk_id更新知识库chunk 标签的接口返回结果为：{jsonarr}")
        return jsonarr

@app.route('/rag/kn/get_content_by_ids', methods=['POST'])
@validate_request
def get_content_by_ids(request_json=None):
    """ 根据content_id获取知识库文件片段 """
    logger.info("--------------------------根据content_id获取知识库文件片段信息---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    content_ids = request_json.get("content_ids")
    kb_id = kb_info.get("kb_id")
    content_type = request_json.get("content_type", "text")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},请求的kb_id为:{kb_id},content_ids:{content_ids}")
        searched_index_name = ""
        if content_type == "text":
            searched_index_name = 'content_control_' + index_name
        elif content_type == "community_report":
            searched_index_name = 'community_report_' + index_name
        contents = es_ops.get_contents_by_ids(searched_index_name, kb_id, content_ids)
        for item in contents:  # 将 kb_id 转换为 kb_name
            item["kb_id"] = kb_id
            item["kb_name"] = display_kb_name
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "contents": contents
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},content_ids:{content_ids}, 根据content_ids获取片段信息的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, "根据content_ids获取分段信息时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},content_ids:{content_ids}, 根据content_ids获取片段信息的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/update_content_status', methods=['POST'])
@validate_request
def update_content_status(request_json=None):
    """ 根据content_id更新知识库文件片段状态 """
    logger.info("--------------------------根据content_id更新知识库文件片段状态---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    file_name = request_json.get("file_name")
    content_id = request_json.get("content_id")
    status = request_json.get("status")
    on_off_switch = request_json.get("on_off_switch", -1)
    kb_id = kb_info.get("kb_id")
    content_index_name = 'content_control_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{userId},请求的kb_id为:{kb_id},file_name:{file_name},content_id:{content_id},status:{status},on_off_switch:{on_off_switch}")
        result = es_ops.update_cc_content_status(content_index_name, kb_id, file_name, content_id, status,
                                                 on_off_switch)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},content_id:{content_id},search_after:{status},on_off_switch:{on_off_switch}根据content_id更新知识库文件片段状态的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="根据content_id更新知识库文件片段状态时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},file_name:{file_name},根据content_id更新知识库文件片段状态的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/get_useful_content_status', methods=['POST'])
@validate_request
def get_content_status(request_json=None):
    """ 获取文本分块状态用于进行检索后过滤。 """
    logger.info("--------------------------获取文本分块状态用于进行检索后过滤---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    content_id_list = request_json.get("content_id_list")
    content_index_name = 'content_control_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(f"用户:{userId},请求的kb_id为:{kb_id},content_id_list:{content_id_list}")
        useful_content_id_list = es_ops.get_cc_content_status(content_index_name, kb_id, content_id_list)
        result = {'code': 0, 'message': 'success', 'data': {'useful_content_id_list': useful_content_id_list}}
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},content_id_list:{content_id_list},获取文本分块状态用于进行检索后过滤的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="获取文本分块状态用于进行检索后过滤时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{userId},知识库:{kb_id},content_id_list:{content_id_list},获取文本分块状态用于进行检索后过滤的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/get_kb_id', methods=['POST'])
@validate_request
def get_kb_id(request_json=None):
    """ 获取某个知识库映射的 kb_id接口 (kb_id 必须在请求中直接传入) """
    logger.info("--------------------------获取知识库映射的 kb_id接口---------------------------\n")
    userId = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    kb_name = kb_info.get("kb_name")
    kb_id = kb_info.get("kb_id")
    logger.info(f"用户:{userId},请求的kb_name为:{kb_name},kb_id为:{kb_id}")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        result = {'code': 0, 'message': 'success', 'data': {'kb_id': kb_id}}
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_name},获取知识库映射的 kb_id接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="获取知识库映射的 kb_id接口发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_name},获取知识库映射的 kb_id接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/rag/kn/update_kb_name', methods=['POST'])
@validate_request
def update_uk_kb_name(request_json=None):
    """ 更新 uk映射表 知识库名接口 """
    logger.info("--------------------------更新 uk映射表 知识库名接口---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    userId = request_json.get("userId")
    old_kb_name = request_json.get("old_kb_name")
    new_kb_name = request_json.get("new_kb_name")
    logger.info(f"用户:{userId},请求的ole_kb_name为:{old_kb_name},请求的new_kb_name为:{new_kb_name}")
    try:
        result = kb_info_ops.update_uk_kb_name(userId, old_kb_name, new_kb_name)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},{old_kb_name},{new_kb_name},更新uk映射表知识库名接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="更新uk映射表知识库名接口发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},{old_kb_name},{new_kb_name},更新uk映射表知识库名接口返回结果为：{jsonarr}")
        return jsonarr


# ***************** 老的 ES snippet API servers **********************

@app.route('/api/v1/rag/es/bulk_add', methods=['POST'])
@validate_request
def snippet_bulk_add(request_json=None):
    logger.info("request: /api/v1/rag/es/bulk_add")
    # logger.info('bulk_add request_params: '+ json.dumps(data, indent=4,ensure_ascii=False))

    # index_name = data.get('index_name') 之前拼接好的，弃用
    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get('kb_info', {})
    kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    doc_list = request_json.get('doc_list')
    logger.info(f"request: bulk_add_data len:{len(doc_list)}")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        es_ops.create_index_if_not_exists(index_name, mappings=es_mapping.snippet_mappings)
        result = es_ops.snippet_bulk_add_index_data(index_name, kb_id, doc_list)
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
        logger.info("bulk_add response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="snippet_bulk_add 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("bulk_add response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/bulk_add end")


@app.route('/api/v1/rag/es/add_file', methods=['POST'])
@validate_request
def add_file(request_json=None):
    logger.info("--------------------------新增文件---------------------------\n")
    user_id = request_json.get("user_id")
    index_name = INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    file_meta = request_json.get("file_meta")
    file_index_name = 'file_control_' + index_name

    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{user_id},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, file_meta: {file_meta}")

        es_ops.create_index_if_not_exists(file_index_name, mappings=es_mapping.file_mappings)
        result = es_ops.add_file(file_index_name, kb_id, file_name, file_meta)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{kb_id},file_name:{file_name},新增文件返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, "新增文件时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{display_kb_name},file_name:{file_name},新增文件返回结果为：{jsonarr}")
        return jsonarr

@app.route('/api/v1/rag/es/allocate_chunks', methods=['POST'])
@validate_request
def allocate_chunks(request_json=None):
    logger.info("--------------------------新增分段时分配chunk---------------------------\n")
    user_id = request_json.get("user_id")
    index_name = INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    count = request_json.get("count")
    chunk_type = request_json.get("chunk_type", "text")
    content_index_name = 'content_control_' + index_name
    file_index_name = 'file_control_' + index_name
    report_index_name = 'community_report_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{user_id},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, insert chunk count: {count}")

        es_ops.create_index_if_not_exists(file_index_name, mappings=es_mapping.file_mappings)
        result = {}
        if chunk_type == "text":
            es_ops.create_index_if_not_exists(content_index_name, mappings=es_mapping.cc_mappings)
            result = es_ops.allocate_chunk_nums(file_index_name, content_index_name, kb_id, file_name, count)
        elif chunk_type == "community_report":
            es_ops.create_index_if_not_exists(report_index_name, mappings=es_mapping.community_report_mappings)
            result = es_ops.allocate_chunk_nums(file_index_name, report_index_name, kb_id, file_name, count)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{kb_id},file_name:{file_name},新增分段分配chunk的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="新增分段分配chunk时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{display_kb_name},file_name:{file_name},新增分段分配chunk返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/allocate_child_chunks', methods=['POST'])
@validate_request
def allocate_child_chunks(request_json=None):
    logger.info("--------------------------新增子分段时分配chunk---------------------------\n")
    user_id = request_json.get("user_id")
    index_name = INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get("kb_info", {})
    display_kb_name = kb_info.get("kb_name")  # 显示的名字
    kb_id = kb_info.get("kb_id")
    file_name = request_json.get("file_name")
    chunk_id = request_json.get("chunk_id")
    count = request_json.get("count")
    content_index_name = 'content_control_' + index_name
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        logger.info(
            f"用户:{user_id},display_kb_name: {display_kb_name},请求的kb_id为:{kb_id},file_name:{file_name}, insert chunk count: {count}")

        es_ops.create_index_if_not_exists(content_index_name, mappings=es_mapping.cc_mappings)
        result = es_ops.allocate_child_chunk_nums(content_index_name, kb_id, file_name, chunk_id, count)
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{kb_id},file_name:{file_name},新增子分段分配chunk的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="新增子分段分配chunk时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{display_kb_name},file_name:{file_name},新增子分段分配chunk返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/search', methods=['POST'])
@validate_request
def snippet_search(request_json=None):
    logger.info("request: /api/v1/rag/es/search")
    logger.info('search request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    content_index_name = 'content_control_' + INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get('kb_info', {})
    kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    query = request_json.get('query')
    top_k = int(request_json.get('top_k', 10))
    min_score = float(request_json.get('min_score', 0.0))
    search_by = request_json.get('search_by', "snippet")
    filter_file_name_list = request_json.get("filter_file_name_list", [])
    metadata_filtering_conditions = request_json.get("metadata_filtering_conditions", [])
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")

        final_conditions = []
        for condition in metadata_filtering_conditions:
            if condition["filtering_kb_name"] == kb_name:
                condition["filtering_kb_name"] = kb_id
                final_conditions.append(deepcopy(condition))

        meta_filter_file_name_list = []
        if final_conditions:
            meta_filter_file_name_list = meta_ops.search_with_doc_meta_filter(content_index_name, final_conditions)
            logger.info(f"用户请求的query为:{query}, filter_file_name_list: {filter_file_name_list}, meta_filter_file_name_list: {meta_filter_file_name_list}")
            if len(meta_filter_file_name_list) == 0:
                result = {
                    "search_list": [],
                    "scores": []
                }
                response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
                logger.info("search response: %s", response)
                return Response(response, mimetype='application/json', status=200)

        if meta_filter_file_name_list:
            filter_file_name_list = filter_file_name_list + meta_filter_file_name_list

        result = es_ops.search_data_text_recall(index_name, kb_id, query, top_k, min_score, search_by,
                                                filter_file_name_list=filter_file_name_list)
        search_list = result["search_list"]
        for item in search_list:  # 将 kb_id 转换为 kb_name
            item["kb_id"] = kb_id
            item["kb_name"] = kb_name
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="snippet search 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/search end")



@app.route('/api/v1/rag/es/keyword_search', methods=['POST'])
@validate_request
def keyword_search(request_json=None):
    logger.info("request: /api/v1/rag/es/keyword_search")
    logger.info('search request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))
    user_id = request_json.get('user_id')
    index_name = INDEX_NAME_PREFIX + request_json.get('user_id')
    content_index_name = 'content_control_' + index_name
    kb_info = request_json.get('kb_info', {})
    display_kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    keywords = request_json.get('keywords')
    top_k = int(request_json.get('top_k', 10))
    min_score = float(request_json.get('min_score', 0.0))
    search_by = request_json.get('search_by', "labels")
    filter_file_name_list = request_json.get('filter_file_name_list', [])
    metadata_filtering_conditions = request_json.get('metadata_filtering_conditions', [])
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")

        final_conditions = []
        for condition in metadata_filtering_conditions:
            if condition["filtering_kb_name"] == display_kb_name:
                condition["filtering_kb_name"] = kb_id
                final_conditions.append(deepcopy(condition))

        meta_filter_file_name_list = []
        if final_conditions:
            meta_filter_file_name_list = meta_ops.search_with_doc_meta_filter(content_index_name, final_conditions)
            logger.info(
                f"filter_file_name_list: {filter_file_name_list}, meta_filter_file_name_list: {meta_filter_file_name_list}")
            if len(meta_filter_file_name_list) == 0:
                result = {
                    "search_list": [],
                    "scores": []
                }
                response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
                logger.info("search response: %s", response)
                return Response(response, mimetype='application/json', status=200)

        if meta_filter_file_name_list:
            filter_file_name_list = filter_file_name_list + meta_filter_file_name_list

        result = es_ops.search_data_keyword_recall(content_index_name, kb_id, keywords, top_k, min_score, search_by,
                                                   filter_file_name_list=filter_file_name_list)
        search_list = result["search_list"]
        for item in search_list:  # 将 kb_id 转换为 kb_name
            item["kb_name"] = display_kb_name
            item["kb_id"] = kb_id
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="keyword search 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/keyword_search end")


@app.route('/api/v1/rag/es/rescore', methods=['POST'])
@validate_request
def snippet_rescore(request_json=None):
    logger.info("request: /api/v1/rag/es/rescore")
    logger.info('rescore request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    search_by = request_json.get('search_by', "snippet")
    search_list_infos = request_json.get("search_list_infos")
    query = request_json.get('query')
    weights = request_json.get('weights')

    logger.info(f"query: {query}, weights: {weights}, search_list_infos:{search_list_infos}")
    try:
        def normalize_to_01(scores):
            if len(scores) == 1:
                return [1.0]  # 单个分数归一化为1
            min_score = min(scores)
            max_score = max(scores)
            if min_score == max_score:
                return [1.0 for _ in scores]  # 所有分数相同，统一设为1
            return [(score - min_score) / (max_score - min_score) for score in scores]

        search_list = []
        bm25_scores = []
        cosine_scores = []
        for user_id, search_list_info in search_list_infos.items():
            kb_id_2_kb_name = {}
            index_name = SNIPPET_INDEX_NAME_PREFIX + user_id.replace('-', '_')
            base_infos = search_list_info["base_infos"]
            temp_search_list = search_list_info["search_list"]
            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, base_infos[0]["kb_name"])
            logger.info(
                f"用户:{user_id},请求查询的base_infos为:{base_infos},embedding_model_id:{embedding_model_id}")

            for base_info in base_infos:
                kb_name = base_info["kb_name"]
                kb_id = base_info["kb_id"]
                kb_id_2_kb_name[kb_id] = kb_name

            regular_list = [item for item in temp_search_list if item.get("chunk_type") != "graph"]
            graph_list = [item for item in temp_search_list if item.get("chunk_type") == "graph"]

            if regular_list:
                result = es_ops.rescore_bm25_score(index_name, query, search_by, regular_list)
                rescored = result["search_list"]
                for item in rescored:
                    item["kb_id"] = item["kb_name"]
                    item["kb_name"] = kb_id_2_kb_name[item["kb_name"]]
                    item["user_id"] = user_id
                search_list.extend(rescored)
                bm25_scores.extend(result["scores"])

            if graph_list:
                graph_result = es_ops.rescore_bm25_score(index_name, query, "meta_data.reference_snippet", graph_list)
                graph_rescored = graph_result["search_list"]
                for item in graph_rescored:
                    item["kb_id"] = item["kb_name"]
                    item["kb_name"] = kb_id_2_kb_name[item["kb_name"]]
                    item["user_id"] = user_id
                    item["snippet"] = item["meta_data"]["reference_snippet"]
                search_list.extend(graph_rescored)
                bm25_scores.extend(graph_result["scores"])

            contents = [item["snippet"] for item in search_list]
            cosine_scores.extend(emb_util.calculate_cosine(query, contents, embedding_model_id))
            logger.info(f"user_id: {user_id}, rescore bm25_scores: {bm25_scores}, cosine_scores: {cosine_scores}")

        bm25_normalized = normalize_to_01(bm25_scores)
        cosine_normalized = normalize_to_01(cosine_scores)

        if len(bm25_normalized) != len(cosine_normalized):
            raise ValueError("BM25 scores and Cosine scores length mismatch")

        final_search_list = []
        for item, text_score, vector_score in zip(search_list, bm25_normalized, cosine_normalized):
            score = weights["vector_weight"] * vector_score + weights["text_weight"] * text_score
            item["score"] = score
            final_search_list.append(item)

        final_search_list.sort(key=lambda x: x["score"], reverse=True)
        final_results = {
            "search_list": final_search_list,
            "scores": [item["score"] for item in final_search_list]
        }
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': final_results}, indent=4, ensure_ascii=False)
        logger.info("rescore response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="snippet rescore 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("rescore response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/rescore end")


@app.route('/api/v1/rag/es/search_text_title_list', methods=['POST'])
@validate_request
def search_title_list(request_json=None):
    logger.info("request: /api/v1/rag/es/search_text_title_list")
    logger.info('search request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    # index_name = data.get('index_name') 之前拼接好的，弃用
    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get('kb_info', {})
    kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    query = request_json.get('query')
    top_k = int(request_json.get('top_k', 10))
    min_score = float(request_json.get('min_score', 0.0))
    kb_id_2_kb_name = {}
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        kb_id_2_kb_name[kb_id] = kb_name
        result = es_ops.search_text_title_list(index_name, kb_id, query, top_k, min_score)
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': result}, indent=4, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="search_text_title_list 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("search response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/search_text_title_list end")


@app.route('/api/v1/rag/es/fetch_all', methods=['POST'])
@validate_request
def snippet_fetch_all(request_json=None):
    logger.info("request: /api/v1/rag/es/fetch_all")
    logger.info('fetch_all request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    # index_name = data.get('index_name') 之前拼接好的，弃用
    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    kb_name = request_json.get('kb_name')
    try:
        documents = es_ops.fetch_all_documents(index_name)
        documents_list = list(documents)
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': documents_list}, indent=4, ensure_ascii=False)
        logger.info("fetch_all response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="snippet fetch_all 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("fetch_all response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/fetch_all end")


@app.route('/api/v1/rag/es/delete_doc', methods=['POST'])
@validate_request
def snippet_delete_doc_by_kbname_title(request_json=None):
    logger.info("request: /api/v1/rag/es/delete_doc")
    logger.info('delete_doc_by_title request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    # index_name = data.get('index_name') 之前拼接好的，弃用
    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get('kb_info', {})
    kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    title = request_json.get('title')
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        status = es_ops.delete_data_by_kbname_title(index_name, kb_id, title)
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': status}, indent=4, ensure_ascii=False)
        logger.info("delete_doc_by_title response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="delete_doc_by_title 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("delete_doc_by_title response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/delete_doc end")


@app.route('/api/v1/rag/es/delete_index', methods=['POST'])
@validate_request
def snippet_delete_index_kb_name(request_json=None):
    logger.info("request: /api/v1/rag/es/delete_index")
    logger.info('delete_index request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    user_id = request_json.get('user_id')
    user_id = user_id.replace('-', '_')
    index_name = SNIPPET_INDEX_NAME_PREFIX + user_id
    kb_info = request_json.get('kb_info', {})
    kb_name = kb_info.get('kb_name')
    kb_id = kb_info.get('kb_id')
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")
        status = es_ops.delete_data_by_kbname(index_name, kb_id)
        response = json.dumps({'code': 200, 'msg': 'Success', 'result': status}, indent=4, ensure_ascii=False)
        logger.info("delete_index response: %s", response)
        return Response(response, mimetype='application/json', status=200)
    except Exception as e:
        log_exception_with_trace(e, msg="snippet_delete_index_kb_name 发生错误")
        response = json.dumps({'code': 400, 'msg': str(e), 'result': None}, ensure_ascii=False)
        logger.info("delete_index response: %s", response)
        return Response(response, mimetype='application/json', status=400)
    finally:
        logger.info("request: /api/v1/rag/es/delete_index end")


@app.route('/rag/kn/add_community_reports', methods=['POST'])
@validate_request
def add_community_reports_data(request_json=None):
    logger.info("--------------------------启动community reports数据添加---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    report_index_name = 'community_report_' + index_name
    file_index_name = 'file_control_' + index_name
    user_id = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    kb_name = kb_info.get("kb_name")
    kb_id = kb_info.get("kb_id")
    embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, kb_name)
    doc_list = request_json.get("data")
    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")

        es_ops.create_index_if_not_exists(report_index_name, mappings=es_mapping.community_report_mappings)
        es_ops.create_index_if_not_exists(file_index_name, mappings=es_mapping.file_mappings)

        # 初始化启停状态
        for doc in doc_list:
            doc["status"] = True  # 初始化启停状态

        # ========= 将 embedding_content 编码好向量 =============
        for batch_doc in batch_list(doc_list, batch_size=EMBEDDING_BATCH_SIZE):
            if is_multimodal_model(embedding_model_id):  # 多模态模型则按多模态去编码
                res = emb_util.get_multimodal_embs([{"text": x["embedding_content"]} for x in batch_doc], embedding_model_id=embedding_model_id)
            else:  # 非多模态知识库则按之前的文本去编码
                res = emb_util.get_embs([x["embedding_content"] for x in batch_doc], embedding_model_id=embedding_model_id)
            dense_vector_dim = len(res["result"][0]["dense_vec"]) if res["result"] else 1024
            field_name = f"q_{dense_vector_dim}_content_vector"

            for i, x in enumerate(batch_doc):
                if len(batch_doc) != len(res["result"]):
                    raise RuntimeError(f"Error getting embeddings:{batch_doc}")
                x[field_name] = res["result"][i]["dense_vec"]
        es_result = es_ops.bulk_add_index_data(report_index_name, kb_id, doc_list)  # 注意 存储的时候传入 kb_id
        if not es_result["success"]:
            logger.info(f"当前用户:{user_id},知识库:{kb_name},add_community_report失败：{es_result}")
            raise RuntimeError(es_result.get("error", ""))
        file_name = "社区报告"
        file_result = es_ops.add_file(file_index_name, kb_id, file_name, doc_list[0]["meta_data"])
        jsonarr = json.dumps(file_result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},知识库:{kb_name},add_community_reports的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="add_community_reports 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},知识库:{kb_name},add_community_reports的接口返回结果为：{jsonarr}")
        return jsonarr
    finally:
        logger.info(f"{user_id},{kb_name},add_community_reports end")


@app.route('/rag/kn/del_community_reports', methods=['POST'])
@validate_request
def del_community_reports(request_json=None):
    logger.info("--------------------------启动community reports删除---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    user_id = request_json.get("userId")
    kb_info = request_json.get("kb_info", {})
    kb_id = kb_info.get("kb_id")
    kb_name = kb_info.get("kb_name")  # 显示的名字
    report_index_name = 'community_report_' + index_name
    file_index_name = 'file_control_' + index_name
    clear_reports = request_json.get("clear_reports", False)
    content_ids = request_json.get("content_ids", [])

    try:
        if not kb_id:
            raise ValueError("kb_info.kb_id is required")

        es_ops.create_index_if_not_exists(report_index_name, mappings=es_mapping.community_report_mappings)
        es_ops.create_index_if_not_exists(file_index_name, mappings=es_mapping.file_mappings)

        file_name = "社区报告"
        if clear_reports:
            er_result = es_ops.delete_data_by_kbname_file_name(report_index_name, kb_id, file_name)
        else:
            er_result = es_ops.delete_chunks_by_content_ids(report_index_name, kb_id, content_ids)
        if not er_result["success"]:
            logger.info(
                f"当前用户:{user_id},知识库:{kb_name},community reports删除时发生错误：{er_result}")
            raise RuntimeError(er_result.get("error", ""))
        if clear_reports:
            es_file_result = es_ops.delete_data_by_kbname_file_name(file_index_name, kb_id, file_name)
            if not es_file_result["success"]:
                logger.info(
                    f"当前用户:{user_id},知识库:{kb_name},file index删除社区报告时发生错误：{es_file_result}")
                raise RuntimeError(es_file_result.get("error", ""))
        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},知识库:{kb_name},community reports删除的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="del_community_reports 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},知识库:{kb_name},community reports删除的接口返回结果为：{jsonarr}")
        return jsonarr

@app.route('/rag/kn/search_community_reports', methods=['POST'])
@validate_request
def search_community_reports(request_json=None):
    """ 多知识库 KNN检索 """
    logger.info("--------------------------启动community reports检索---------------------------\n")
    index_name = INDEX_NAME_PREFIX + request_json.get('userId')
    report_index_name = 'community_report_' + index_name
    userId = request_json.get("userId")
    kb_infos = request_json.get("kb_infos")  # list of {kb_name, kb_id}
    top_k = request_json.get("topk", 10)
    query = request_json.get("question")
    min_score = request_json.get("threshold", 0)
    kb_id_2_kb_name = {}
    emb_id2kb_names = {}
    logger.info(f"用户:{index_name},请求查询的kb_infos为:{kb_infos}")
    logger.info(f"用户请求的query为:{query}")
    try:
        if not kb_infos:
            raise ValueError("kb_infos is required")
        for kb_info_item in kb_infos:
            kb_name = kb_info_item["kb_name"]
            kb_id = kb_info_item["kb_id"]
            if not kb_id:
                raise ValueError(f"kb_id is required for kb_name: {kb_name}")
            kb_id_2_kb_name[kb_id] = kb_name
            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(userId, kb_name)
            if embedding_model_id not in emb_id2kb_names:
                emb_id2kb_names[embedding_model_id] = []
            emb_id2kb_names[embedding_model_id].append(kb_id)

        # ============= 开始检索召回 ===============
        es_ops.create_index_if_not_exists(report_index_name, mappings=es_mapping.community_report_mappings)

        search_list = []
        scores = []
        for embedding_model_id, kb_names in emb_id2kb_names.items():
            logger.info(f"用户:{index_name},请求查询的kb_names为:{kb_names},embedding_model_id:{embedding_model_id}")
            result_dict = es_ops.search_data_knn_recall(report_index_name, kb_names, query, top_k, min_score, embedding_model_id=embedding_model_id)
            search_list.extend(result_dict["search_list"])
            scores.extend(result_dict["scores"])

        if len(search_list) > top_k:
            # 合并search_list和scores，按score降序排序
            combined_results = list(zip(search_list, scores))
            combined_results.sort(key=lambda x: x[1], reverse=True)

            # 取前top_k个结果
            top_results = combined_results[:top_k]
            search_list = [item[0] for item in top_results]
            scores = [item[1] for item in top_results]

        for item in search_list:  # 将 kb_id 转换为 kb_name
            item["kb_id"] = item["kb_name"]
            item["kb_name"] = kb_id_2_kb_name[item["kb_name"]]
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "search_list": search_list,
                "scores": scores
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_infos},query:{query},向量库检索的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="查询知识库时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{userId},知识库:{kb_infos},query:{query},向量库检索的接口返回结果为：{jsonarr}")
        return jsonarr

#-------------------------------       问答库       ------------------------------------
@app.route('/api/v1/rag/es/init_QA_base', methods=['POST'])
@validate_request
def init_qa_base(request_json=None):
    """ 初始化 init_qa 接口"""
    logger.info("--------------------------启动问答库初始化---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    embedding_model_id = request_json["embedding_model_id"]
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}, embedding_model_id: {embedding_model_id}")

        judge_time = time.time()

        es_ops.create_index_if_not_exists(KBNAME_MAPPING_INDEX, mappings=es_mapping.uk_mappings)
        es_ops.create_index_if_not_exists(qa_index_name, mappings=es_mapping.qa_mappings)
        qa_base_names = kb_info_ops.get_uk_qa_name_list(user_id)  # 从映射表中获取
        logger.info(f"当前用户:{user_id},共有问答库：{len(qa_base_names)}个，分别为{qa_base_names}")
        judge_time = time.time() - judge_time
        logger.info(f"--------------------------查询qa_map时间:{judge_time}---------------------------\n")
        if qa_base_name in qa_base_names:
            raise RuntimeError(f"已存在同名问答库{qa_base_name}")

        utc_now = datetime.utcnow()
        formatted_time = utc_now.strftime('%Y-%m-%d %H:%M:%S')
        uk_data = [
            {"index_name": qa_index_name, "userId": user_id, "kb_name": qa_base_name,
             "creat_time": formatted_time, "kb_id": qa_base_id, "embedding_model_id": embedding_model_id,
             "is_qa": True}
        ]
        kb_info_ops.bulk_add_uk_index_data(KBNAME_MAPPING_INDEX, uk_data)
        # ====== 新建完成，验证 kb_id 是否写入成功 ======
        qa_id_list = kb_info_ops.get_uk_kb_id_list(KBNAME_MAPPING_INDEX, user_id)
        if qa_base_id not in qa_id_list:  # 新建失败，返回错误
            raise RuntimeError("ini问答库失败，ES写入失败")

        # 新建成功，返回
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}")
        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},ini知识库的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="ini_qa_base 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},ini知识库的接口返回结果为：{jsonarr}")
        return jsonarr



@app.route('/api/v1/rag/es/delete_QA_base', methods=['POST'])
@validate_request
def del_qa_base(request_json=None):
    logger.info("--------------------------启动问答库删除---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}")

        es_result = qa_ops.delete_data_by_qa_info(qa_index_name, qa_base_name, qa_base_id)
        if not es_result["success"]:
            logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, qa_index_name: {qa_index_name}, 问答库删除时发生错误：{es_result}")
            raise RuntimeError(es_result.get("error", ""))

        es_uk_result = es_ops.delete_uk_data_by_kbname(user_id, qa_base_name)
        if not es_uk_result["success"]:
            logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, uk_index_name: {KBNAME_MAPPING_INDEX}, 问答库删除时发生错误：{es_uk_result}")
            raise RuntimeError(es_uk_result.get("error", ""))

        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name},问答库删除的接口返回结果为：{jsonarr},{es_result},{es_uk_result}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="del_qa_base 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},问答库删除的接口返回结果为：{jsonarr}")
        return jsonarr



@app.route('/api/v1/rag/es/add-QAs', methods=['POST'])
@validate_request
def add_qa_data(request_json=None):
    """ 往 ES 中建向量索引数据"""
    logger.info("--------------------------启动问答库数据添加---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, qa_base_name)
    qa_list = request_json.get("data")

    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}")

        # ========= 将 embedding_content 编码好向量 =============
        for batch_doc in batch_list(qa_list, batch_size=EMBEDDING_BATCH_SIZE):
            if is_multimodal_model(embedding_model_id):  # 多模态模型则按多模态去编码
                res = emb_util.get_multimodal_embs([{"text": x["question"]} for x in batch_doc], embedding_model_id=embedding_model_id)
            else:  # 非多模态知识库则按之前的文本去编码
                res = emb_util.get_embs([x["question"] for x in batch_doc], embedding_model_id=embedding_model_id)
            dense_vector_dim = len(res["result"][0]["dense_vec"]) if res["result"] else 1024
            field_name = f"q_{dense_vector_dim}_content_vector"

            for i, x in enumerate(batch_doc):
                if len(batch_doc) != len(res["result"]):
                    raise RuntimeError(f"Error getting embeddings:{batch_doc}")
                x[field_name] = res["result"][i]["dense_vec"]
        # ========= 将 embedding_content 编码好向量 =============
        es_result = qa_ops.bulk_add_index_data(qa_index_name, qa_base_name, qa_list)
        if not es_result["success"]:
            raise RuntimeError(es_result.get("error", ""))

        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},add的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="add_qa_data 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name},add的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/batch-delete-QAs', methods=['POST'])
@validate_request
def batch_delete_qas(request_json=None):
    logger.info("--------------------------根据qa pair ids 删除问答对---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    qa_pair_ids = request_json["QAPairIds"]
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}, qa_pair_ids: {qa_pair_ids}")

        es_result = qa_ops.delete_qa_ids(qa_index_name, qa_base_name, qa_base_id, qa_pair_ids)
        if not es_result["success"]:
            logger.info(
                f"当前用户:{user_id},问答库:{qa_base_name}, 问答对删除时发生错误：{es_result}")
            raise RuntimeError(es_result.get("error", ""))

        result = {
            "code": 0,
            "message": "success",
            "data": {
                "success_count": es_result["deleted"]
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, 问答对删除的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="batch_delete_qas 发生错误")
        result = {
            "code": 1,
            "message": str(e),
            "data": {
                "success_count": 0
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, 问答对删除的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/update_QA', methods=['POST'])
@validate_request
def update_qa(request_json=None):
    """ 根据id更新问答片段状态 """
    logger.info("--------------------------根据qa pair id 更新问答对--------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    qa_pair_id = request_json["QAPairId"]
    update_data = request_json["data"]
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}, qa_pair_id: {qa_pair_id}, update_data:{update_data}")

        if "question" in update_data:
            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, qa_base_name)
            if is_multimodal_model(embedding_model_id):  # 多模态模型则按多模态去编码
                res = emb_util.get_multimodal_embs([{"text": update_data["question"]}], embedding_model_id=embedding_model_id)
            else:  # 非多模态知识库则按之前的文本去编码
                res = emb_util.get_embs([update_data["question"]], embedding_model_id=embedding_model_id)
            if len(res["result"]) != 1:
                raise RuntimeError(f"Error getting embeddings:{update_data}")
            dense_vector_dim = len(res["result"][0]["dense_vec"]) if res["result"] else 1024
            field_name = f"q_{dense_vector_dim}_content_vector"
            update_data[field_name] = res["result"][0]["dense_vec"]
        es_result = qa_ops.update_qa_data(qa_index_name, qa_base_name, qa_pair_id, update_data)
        if not es_result["success"]:
            logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, 问答对更新时发生错误：{es_result}")
            raise RuntimeError(es_result.get("error", ""))

        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name},qa_pair_id:{qa_pair_id}, 更新的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="update_qa 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name},qa_pair_id:{qa_pair_id}, 更新的接口返回结果为：{jsonarr}")
        return jsonarr

@app.route('/api/v1/rag/es/get_QA_list', methods=['POST'])
@validate_request
def get_qa_list(request_json=None):
    """ 获取 分页展示 """
    logger.info("--------------------------获取问答对的分页展示---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    page_size = request_json.get("page_size")
    search_after = request_json.get("search_after")
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}, page_size: {page_size}, search_after:{search_after}")

        qa_result = qa_ops.get_qa_list(qa_index_name, qa_base_name, qa_base_id, page_size, search_after)
        result = {
            "code": 0,
            "message": "success",
            "data": qa_result
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name},page_size:{page_size},search_after:{search_after},分页查询的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="获取问答对的分页展示时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name},分页展示的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/update_QA_metas', methods=['POST'])
@validate_request
def update_qa_metas(request_json=None):
    logger.info("--------------------------更新问答库元数据---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_info = request_json.get("qa_info", {})
    qa_base_name = qa_info.get("QABase")
    qa_base_id = qa_info.get("QAId")
    metas = request_json.get("metas")
    update_type = request_json["update_type"]
    if not qa_base_id:
        raise ValueError("qa_info.QAId is required")
    try:
        logger.info(f"用户:{user_id},问答库:{qa_base_name},qa_base_id:{qa_base_id}, metas: {metas}")

        es_result = {}
        if update_type == "update_metas":
            es_result = qa_ops.update_meta_datas(qa_index_name, qa_base_name, qa_base_id, metas)
        elif update_type == "delete_keys":
            es_result = qa_ops.delete_meta_by_key(qa_index_name, qa_base_name, qa_base_id, metas)
        elif update_type == "rename_keys":
            es_result = qa_ops.rename_metas(qa_index_name, qa_base_name, qa_base_id, metas)
        if not es_result["success"]:
            logger.info(f"当前用户:{user_id},问答库:{qa_base_name}, 问答对更新元数据时发生错误：{es_result}")
            raise RuntimeError(es_result.get("error", ""))

        result = {
            "code": 0,
            "message": "success"
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name}, 更新元数据的接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="update_qa_metas 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(
            f"当前用户:{user_id},问答库:{qa_base_name}, 更新元数据的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/qa_rescore', methods=['POST'])
@validate_request
def qa_rescore(request_json=None):
    logger.info("request: /api/v1/rag/es/qa_rescore")
    logger.info('qa rescore request_params: ' + json.dumps(request_json, indent=4, ensure_ascii=False))

    search_list_infos = request_json.get("search_list_infos")
    query = request_json.get('query')
    weights = request_json.get('weights')

    try:
        search_list = []
        bm25_scores = []
        cosine_scores = []
        for user_id, search_list_info in search_list_infos.items():
            qa_index_name = get_qa_index_name(user_id)
            base_infos = search_list_info["base_infos"]
            temp_search_list = search_list_info["search_list"]
            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, base_infos[0]["QABase"])

            result = qa_ops.qa_rescore_bm25_score(qa_index_name, query, temp_search_list)
            temp_search_list = result["search_list"]
            search_list.extend(temp_search_list)
            bm25_scores.extend(result["scores"])
            contents = [item["question"] for item in temp_search_list]
            cosine_scores.extend(emb_util.calculate_cosine(query, contents, embedding_model_id))
            logger.info(f"rescore bm25_scores: {bm25_scores}, cosine_scores: {cosine_scores}")

        def normalize_to_01(scores):
            if len(scores) == 1:
                return [1.0]  # 单个分数归一化为1
            min_score = min(scores)
            max_score = max(scores)
            if min_score == max_score:
                return [1.0 for _ in scores]  # 所有分数相同，统一设为1
            return [(score - min_score) / (max_score - min_score) for score in scores]

        bm25_normalized = normalize_to_01(bm25_scores)
        cosine_normalized = normalize_to_01(cosine_scores)

        final_search_list = []
        for item, text_score, vector_score in zip(search_list, bm25_normalized, cosine_normalized):
            score = weights["vector_weight"] * vector_score + weights["text_weight"] * text_score
            item["score"] = score
            final_search_list.append(item)

        final_search_list.sort(key=lambda x: x["score"], reverse=True)
        result = {
            "code": 0,
            "message": "success",
            "data": {
                "search_list": final_search_list,
                "scores": [item["score"] for item in final_search_list]
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"qa_rescore接口返回结果为：{jsonarr}")
        return jsonarr
    except Exception as e:
        log_exception_with_trace(e, msg="qa_rescore 发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"qa_rescore接口返回结果为：{jsonarr}")
        return jsonarr



@app.route('/api/v1/rag/es/vector_search', methods=['POST'])
@validate_request
def vector_search(request_json=None):
    """ 多知识库 KNN检索 """
    logger.info("--------------------------启动问答库向量检索---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_infos = request_json.get("qa_infos", [])
    all_base_names = [item["QABase"] for item in qa_infos]
    top_k = request_json.get("topk", 10)
    query = request_json.get("question")
    min_score = request_json.get("threshold", 0)
    metadata_filtering_conditions = request_json.get("metadata_filtering_conditions", [])
    emb_id2base_names = {}
    logger.info(f"用户:{user_id},请求查询的qa_infos为:{qa_infos}, query: {query}, topK: {top_k}, "
                f"threshold: {min_score}, metadata_filtering_conditions: {metadata_filtering_conditions}")
    try:
        if not qa_infos:
            raise ValueError("qa_infos is required")

        exists_base_names = kb_info_ops.get_uk_qa_name_list(user_id)  # 从映射表中获取
        filtering_conditions = {}
        for condition in metadata_filtering_conditions:
            base_name = condition["filtering_qa_base_name"]
            filtering_conditions[base_name] = condition

        final_conditions = []
        for base_name in all_base_names:
            if base_name not in exists_base_names:
                raise RuntimeError(f"用户:{user_id}, {base_name}问答库不存在")

            if base_name in filtering_conditions:
                condition = filtering_conditions[base_name]
                final_conditions.append(deepcopy(condition))

            embedding_model_id = kb_info_ops.get_uk_kb_emb_model_id(user_id, base_name)
            if embedding_model_id not in emb_id2base_names:
                emb_id2base_names[embedding_model_id] = []
            emb_id2base_names[embedding_model_id].append(base_name)

        search_list = []
        scores = []
        for embedding_model_id, base_names in emb_id2base_names.items():
            logger.info(f"用户:{user_id},请求查询的base_names为:{base_names}, query: {query}, embedding_model_id:{embedding_model_id}")
            result_dict = qa_ops.vector_search(qa_index_name, base_names, query, top_k, min_score,
                                               embedding_model_id=embedding_model_id, meta_filter_list=final_conditions)
            search_list.extend(result_dict["search_list"])
            scores.extend(result_dict["scores"])

        if len(search_list) > top_k:
            # 合并search_list和scores，按score降序排序
            combined_results = list(zip(search_list, scores))
            combined_results.sort(key=lambda x: x[1], reverse=True)

            # 取前top_k个结果
            top_results = combined_results[:top_k]
            search_list = [item[0] for item in top_results]
            scores = [item[1] for item in top_results]

        result = {
            "code": 0,
            "message": "success",
            "data": {
                "search_list": search_list,
                "scores": scores
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{all_base_names},query:{query},向量检索的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e,msg="查询问答库时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{all_base_names},query:{query},向量检索的接口返回结果为：{jsonarr}")
        return jsonarr


@app.route('/api/v1/rag/es/text_search', methods=['POST'])
@validate_request
def text_search(request_json=None):
    """ 多问答库库 text检索 """
    logger.info("--------------------------启动问答库全文检索---------------------------\n")
    user_id = request_json.get("userId")
    qa_index_name = get_qa_index_name(user_id)
    qa_infos = request_json.get("qa_infos", [])
    base_names = [item["QABase"] for item in qa_infos]
    top_k = request_json.get("topk", 10)
    query = request_json.get("question")
    min_score = request_json.get("threshold", 0)
    metadata_filtering_conditions = request_json.get("metadata_filtering_conditions", [])
    logger.info(f"用户:{user_id},请求查询的qa_infos为:{qa_infos}, query: {query}, topK: {top_k}, "
                f"threshold: {min_score}, metadata_filtering_conditions: {metadata_filtering_conditions}")
    try:
        if not qa_infos:
            raise ValueError("qa_infos is required")

        exists_base_names = kb_info_ops.get_uk_qa_name_list(user_id)  # 从映射表中获取
        filtering_conditions = {}
        for condition in metadata_filtering_conditions:
            base_name = condition["filtering_qa_base_name"]
            filtering_conditions[base_name] = condition

        final_conditions = []
        for base_name in base_names:
            if base_name not in exists_base_names:
                raise RuntimeError(f"用户:{user_id}, {base_name}问答库不存在")

            if base_name in filtering_conditions:
                condition = filtering_conditions[base_name]
                final_conditions.append(deepcopy(condition))

        result_dict = qa_ops.text_search(qa_index_name, base_names, query, top_k, min_score,
                                         meta_filter_list=final_conditions)
        search_list = result_dict["search_list"]
        scores = result_dict["scores"]

        result = {
            "code": 0,
            "message": "success",
            "data": {
                "search_list": search_list,
                "scores": scores
            }
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{base_names},query:{query},全文检索的接口返回结果为：{jsonarr}")
        return jsonarr

    except Exception as e:
        log_exception_with_trace(e, msg="查询问答库时发生错误")
        result = {
            "code": 1,
            "message": str(e)
        }
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f"当前用户:{user_id},问答库:{base_names},query:{query},全文检索的接口返回结果为：{jsonarr}")
        return jsonarr

if __name__ == '__main__':
    app.run()  # debug=True
