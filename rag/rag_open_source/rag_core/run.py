#!/usr/bin/env python
# -*- encoding: utf-8 -*-
import os
import json
import requests
import time
import logging

from flask import Flask, jsonify, request, make_response
from flask_cors import CORS

from textsplitter import ChineseTextSplitter
from pymongo import MongoClient
import argparse
from utils import redis_utils
from utils import file_utils
from utils import kafka_utils
from utils import chunk_utils
from utils import graph_utils
import utils.knowledge_base_utils as kb_utils
from utils.constant import CHUNK_SIZE
import urllib.parse
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
from logging_config import init_logging
from settings import MONGO_URL, USE_DATA_FLYWHEEL
from qa import index as qa_index
from qa import search as qa_search
from utils.http_util import validate_request
from model_manager.model_config import get_model_configure
from utils.otel import init_tracer
from utils.trace import register_tracing
from utils.tools import get_query_dict_cache, query_rewrite

# 初始化 OpenTelemetry（必须在 Flask app 创建之前）
init_tracer("rag-wanwu-core-service")

# 自动 instrument requests 库（出站 HTTP 传播 traceparent）
from opentelemetry.instrumentation.requests import RequestsInstrumentor

RequestsInstrumentor().instrument()

# 定义路径
paths = ["./parser_data"]
# 遍历路径列表
for path in paths:
    # 检查路径是否存在
    if not os.path.exists(path):
        # 如果不存在，则创建目录
        os.makedirs(path)
        print(f"目录 {path} 已创建。")
    else:
        print(f"目录 {path} 已存在。")

app = Flask(__name__)
init_logging()
logger = logging.getLogger(__name__)

# 自动 instrument Flask（入站 HTTP 提取 traceparent，创建 Span）
from opentelemetry.instrumentation.flask import FlaskInstrumentor

FlaskInstrumentor().instrument_app(app)

# 添加日志记录（含 trace_id/span_id 关联）
register_tracing(app)

CORS(app, resources={r"/*": {"origins": "*"}})

app.config['JSON_AS_ASCII'] = False
app.config['JSONIFY_MIMETYPE'] = 'application/json;charset=utf-8'
# 初始化 MongoDB 客户端
client = MongoClient(MONGO_URL, 0, connectTimeoutMS=5000, serverSelectionTimeoutMS=3000)
collection = client['rag']['rag_user_logs']
redis_client = redis_utils.get_redis_connection()
chunk_label_redis_client = redis_utils.get_redis_connection(redis_db=5)

@app.route('/rag/init-knowledge-base', methods=['POST'])
@validate_request
def init_kb(request_json=None):
    logger.info('---------------初始化知识库---------------')
    try:
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        embedding_model_id = request_json.get("embedding_model_id", "")
        enable_knowledge_graph = request_json.get("enable_knowledge_graph", False)
        is_multimodal = request_json.get("is_multimodal", False)
        logger.info(repr(request_json))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(embedding_model_id) > 0

        result_data = kb_utils.init_knowledge_base(user_id, kb_info,
                                                   embedding_model_id=embedding_model_id,
                                                   enable_knowledge_graph=enable_knowledge_graph,
                                                   is_multimodal=is_multimodal)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False))
        # response = make_response(json.dumps(result_data, ensure_ascii=False),headers)

    except Exception as e:
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False))
    response.headers['Access-Control-Allow-Origin'] = '*'
    return response


@app.route("/rag/del-knowledge-base", methods=["POST"])  # 删除知识库 done
@validate_request
def del_kb(request_json=None):
    logger.info('---------------删除知识库---------------')
    try:
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        logger.info(repr(request_json))

        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        result_data = kb_utils.del_konwledge_base(user_id, kb_info)
        # 在批量删除文件中补充增加删除reids逻辑 begin
        if USE_DATA_FLYWHEEL:
            try:
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"del-knowledge-base Failed to get redis connection: {err}")
                import traceback
                logger.error(traceback.format_exc())
        # 在批量删除文件中补充增加删除reids逻辑 end
        # ========== chunk labels 删除的逻辑 ==========
        try:
            # 删除chunk_labels
            redis_utils.delete_chunk_labels(chunk_label_redis_client, kb_id)
        except Exception as err:
            logger.error(f"del-knowledge-base Failed to delete_chunk_labels: {err}")
            import traceback
            logger.error(traceback.format_exc())
        # ========== chunk labels 删除的逻辑 ==========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-file-tags", methods=['POST'])
@validate_request
def updateFileTags(request_json=None):
    logger.info('---------------更新文件元数据---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        tags = request_json.get("tags", None)
        logger.info(repr(request_json))

        if tags is None:
            raise ValueError("tags must be not None")
        if not isinstance(tags, list):
            raise ValueError("tags must be a list or None")
        metas = {
            "metas": [{
                "file_name": file_name,
                "metadata_list": tags
            }]
        }
        response_info = kb_utils.manage_kb_metadata(user_id, kb_info, kb_utils.MetadataOperation.UPDATE_METAS, metas)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-file-metas", methods=['POST'])
@validate_request
def updateFileMetas(request_json=None):
    logger.info('---------------批量更新文件元数据---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        metas = request_json.get("metas", [])
        logger.info(repr(request_json))

        if not isinstance(metas, list):
            raise ValueError("metas must be a list")
        if not metas:
            raise ValueError("metas must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_info, kb_utils.MetadataOperation.UPDATE_METAS, {"metas": metas})
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/delete-meta-by-keys", methods=['POST'])
@validate_request
def deleteMetaByKeys(request_json=None):
    logger.info('---------------知识库删除元数据key---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        keys = request_json.get("keys", [])
        logger.info(repr(request_json))

        if not isinstance(keys, list):
            raise ValueError("keys must be a list")
        if not keys:
            raise ValueError("keys must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_info, kb_utils.MetadataOperation.DELETE_KEYS, {"keys": keys})
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/rename-meta-keys", methods=['POST'])
@validate_request
def renameMetaKeys(request_json=None):
    logger.info('---------------重命名知识库元数据key---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        key_mappings = request_json.get("mappings", [])
        logger.info(repr(request_json))

        if not isinstance(key_mappings, list):
            raise ValueError("key_mappings must be a list")
        if not key_mappings:
            raise ValueError("key_mappings must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_info, kb_utils.MetadataOperation.RENAME_KEYS, {"key_mappings": key_mappings})
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-chunk-labels", methods=['POST'])
@validate_request
def updateChunkLabels(request_json=None):
    logger.info('---------------更新分片标签---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        chunk_id = request_json.get("chunk_id")
        labels = request_json.get("labels", None)
        logger.info(repr(request_json))

        if labels is None or not isinstance(labels, list):
            raise ValueError("labels must specified as an array")

        response_info = chunk_utils.update_chunk_labels(user_id, kb_info, file_name, chunk_id, labels)
        # ======= chunk labels 更新的逻辑 ========
        redis_utils.update_chunk_labels(chunk_label_redis_client, kb_id, file_name, chunk_id, labels)
        # ======= chunk labels 更新的逻辑 ========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/search-knowledge-base", methods=["POST"])  # 查询 done
@validate_request
def search_knowledge_base(request_json=None):
    logger.info('---------------问题查询---------------')
    response_info = {
        'code': 0,
        "message": "成功",
        "data": {
            "prompt": "",
            "searchList": []
        }
    }
    try:
        enable_vision = request_json.get("enable_vision", False)
        attachment_files = request_json.get("attachment_files", [])
        return_meta = request_json.get("return_meta", False)
        prompt_template = request_json.get("prompt_template", '')
        knowledge_base_info = request_json.get("knowledge_base_info", {})
        question = request_json.get('question')
        rate = float(request_json.get('threshold', 0))
        top_k = int(request_json.get('topK', 5))
        chunk_conent = int(request_json.get('extend', '1'))
        chunk_size = int(request_json.get('extendedLength', CHUNK_SIZE))
        search_field = request_json.get('search_field', 'con')
        # if user_id == '': user_id = str(request.headers.get('X-Uid', ''))
        default_answer = request_json.get("default_answer", '根据已知信息，无法回答您的问题。')
        # 是否开启自动引文，此参数与prompt_template互斥，当开启auto_citation时，prompt_template用户传参不生效
        auto_citation = request_json.get("auto_citation", False)
        # 是否query改写
        rewrite_query = request_json.get("rewrite_query", False)
        use_graph = request_json.get("use_graph", False)
        filter_file_name_list = request_json.get("filter_file_name_list", [])
        rerank_mod = request_json.get("rerank_mod", "rerank_model")
        # Dify开源版本问答时需指定rerank模型
        rerank_model_id = request_json.get("rerank_model_id", '')
        weights = request_json.get("weights", None)
        retrieve_method = request_json.get("retrieve_method", "hybrid_search")

        #metadata filtering params
        metadata_filtering = request_json.get("metadata_filtering", False)
        metadata_filtering_conditions = request_json.get("metadata_filtering_conditions", [])
        if not metadata_filtering:
            metadata_filtering_conditions = []
        logger.info(repr(request_json))

        # 检查 knowledge_base_info 是否为空
        if not knowledge_base_info:
            raise ValueError("knowledge_base_info cannot be empty")
        # 检查 rerank_model_id 是否为空
        if rerank_mod == "rerank_model" and not rerank_model_id:
            raise ValueError("rerank_model_id cannot be empty when using model-based reranking.")

        if rerank_mod == "weighted_score" and weights is None:
            raise ValueError("weights cannot be empty when using weighted score reranking.")
        if weights is not None and not isinstance(weights, dict):
            raise ValueError("weights must be a dictionary or None.")

        if rerank_mod == "weighted_score" and retrieve_method != "hybrid_search":
            raise ValueError("Weighted score reranking is only supported in hybrid search mode.")

        filter_attachment_files = []
        for item in attachment_files:
            file_type = item.get("file_type")
            if file_type == "image":
                file_url = item["file_url"]
                parsed_url = urllib.parse.urlparse(file_url)
                # 校验 URL 是否包含协议头和网络位置
                if not all([parsed_url.scheme, parsed_url.netloc]):
                    raise ValueError(f"Invalid attachment file URL: {file_url}")
                filter_attachment_files.append({"image": file_url})
            else:
                raise ValueError(f"attachment_file type {file_type} not support")
        if len(filter_attachment_files) > 1:
            raise ValueError("Multiple attachment files are not supported.")
        attachment_files = filter_attachment_files

        # assert len(user_id) > 0
        # assert len(kb_name) > 0 or len(kb_id) > 0
        if not attachment_files and not (question and len(str(question).strip()) > 0):
            raise ValueError("Invalid input: Attachment and question cannot both be empty")

        if attachment_files and not (question and len(str(question).strip()) > 0):
            if rerank_mod != "rerank_model":
                raise ValueError("rerank_mod must be rerank_model when only attachment_files is provided.")
            model_config = get_model_configure(rerank_model_id)
            if not model_config.is_multimodal:
                raise ValueError("The specified rerank_model_id does not support multimodal input.")

        if rewrite_query:
            for user_id, kb_info_list in knowledge_base_info.items():
                kb_names = [kb_info['kb_name'] for kb_info in kb_info_list]
                kb_ids = [kb_info['kb_id'] for kb_info in kb_info_list]
                query_dict_list = get_query_dict_cache(redis_client, user_id, kb_ids)
                if query_dict_list:
                    rewritten_queries = query_rewrite(question, query_dict_list)
                    logger.info("对query进行改写,原问题:%s 改写后问题:%s" % (question, ",".join(rewritten_queries)))
                    if len(rewritten_queries) > 0:
                        question = rewritten_queries[0]
                        logger.info("按新问题:%s 进行召回" % question)
                else:
                    logger.info("未启用或维护转名词表,query未改写,按原问题:%s 进行召回" % question)

        response_info = kb_utils.get_knowledge_based_answer(knowledge_base_info, question, rate, top_k, chunk_conent, chunk_size,
                                                   return_meta, prompt_template, search_field, default_answer,
                                                   auto_citation, retrieve_method = retrieve_method,
                                                   filter_file_name_list=filter_file_name_list,
                                                   rerank_model_id=rerank_model_id, rerank_mod=rerank_mod,
                                                   weights=weights, metadata_filtering_conditions=metadata_filtering_conditions,
                                                   use_graph=use_graph, enable_vision=enable_vision,
                                                   attachment_files=attachment_files)
        json_str = json.dumps(response_info, ensure_ascii=False)

        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"prompt": "", "searchList": []}}
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    return response


@app.route("/rag/list-knowledge-base", methods=["POST"])  # 查询用户下所有的知识库名称 done
@validate_request
def list_kb(request_json=None):
    logger.info('---------------查询所有知识库---------------')
    try:
        user_id = request_json["userId"]
        logger.info(repr(request_json))

        assert len(user_id) > 0

        response_info = kb_utils.list_knowledge_base(user_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"knowledge_base_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/list-knowledge-file", methods=["POST"])  # 显示用户知识库下所有的文件 done
@validate_request
def list_file(request_json=None):
    logger.info('---------------查询所有知识库文件---------------')
    try:
        user_id = request_json["userId"]
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        logger.info(repr(request_json))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        response_info = kb_utils.list_knowledge_file(user_id, kb_info)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": "repr(e)", "data": {"knowledge_file_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response



@app.route("/rag/list-knowledge-file-download-link", methods=["POST"])  # 显示用户知识库下所有的文件 done
@validate_request
def list_file_download_link(request_json=None):
    logger.info('---------------查询所有知识库文件的 download_link---------------')
    try:
        user_id = request_json["userId"]
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        logger.info(repr(request_json))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        response_info = kb_utils.list_knowledge_file_download_link(user_id, kb_info)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": "repr(e)", "data": {"knowledge_file_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/del-knowledge-file", methods=["POST"])  # 删除某个知识库下的单个文件 done
@validate_request
def del_file(request_json=None):
    logger.info('---------------删除知识库文件---------------')
    try:
        file_name = request_json.get("fileName")
        if not kb_utils.is_safe_filename(file_name):
            raise ValueError("fileName is not safe")
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        logger.info(repr(request_json))

        assert len(file_name) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(user_id) > 0

        result_data = kb_utils.del_knowledge_base_files(user_id, kb_info, [file_name])
        # 在批量删除文件中补充增加删除reids逻辑 begin
        if USE_DATA_FLYWHEEL:
            try:
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"del-knowledge-file Failed to get redis connection: {err}")
                import traceback
                logger.error(traceback.format_exc())
        # 在批量删除文件中补充增加删除reids逻辑 end
        # ========== chunk labels 删除的逻辑 ==========
        try:
            # 删除chunk_labels
            redis_utils.delete_chunk_labels(chunk_label_redis_client, kb_id, file_name=file_name)
        except Exception as err:
            logger.error(f"del-knowledge-file Failed to delete_chunk_labels: {err}")
            import traceback
            logger.error(traceback.format_exc())
        # ========== chunk labels 删除的逻辑 ==========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)

    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/batch_del_knowfiles", methods=["POST"])  # 删除某个知识库下的多个文件 done
@validate_request
def del_files(request_json=None):
    logger.info('---------------批量删除知识库文件---------------')
    try:
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_names = request_json.get("fileNames")
        for file_name in file_names:
            if not kb_utils.is_safe_filename(file_name):
                raise ValueError("fileName is not safe")

        logger.info(repr(request_json))

        assert len(file_names) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(user_id) > 0

        result_data = kb_utils.del_knowledge_base_files(user_id, kb_info, file_names)
        # 在批量删除文件中补充增加删除reids逻辑 begin
        if USE_DATA_FLYWHEEL:
            try:
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"Failed to get redis connection maybe not use dataflywheel or uninstall redis: {err}")
        # 在批量删除文件中补充增加删除reids逻辑 end
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)

    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/check-knowledge-base", methods=["POST"])  # 查询某个知识库是否在某个用户下 done
@validate_request
def check_kb(request_json=None):
    logger.info('---------------校验知识库是否存在---------------')
    try:
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        logger.info(repr(request_json))

        assert len(user_id) > 0

        response_info = kb_utils.check_knowledge_base(user_id, kb_info)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"knowledge_base_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/split_text", methods=['POST'])
@validate_request
def split_text(request_json=None):
    text = request_json.get('text', '')
    chunk_type = request_json.get('chunk_type')
    sentence_size = request_json.get('sentence_size', 500)
    overlap_size = request_json.get('overlap_size', 0.2)
    separators = request_json.get('separators', ["。", "！", "？", ".", "!", "?", "……", "|\n"])
    pdf = request_json.get('pdf', False)
    excel = request_json.get('excel', False)

    splitter = ChineseTextSplitter(chunk_type=chunk_type, sentence_size=sentence_size, overlap_size=overlap_size,
                                   pdf=pdf, excel=excel, separators=separators)
    result = splitter.split_text(text)

    return jsonify(result)


@app.route("/rag/get-content-list", methods=['POST'])
@validate_request
def getContentList(request_json=None):
    logger.info('---------------获取某个文件的文本分块列表---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        page_size = request_json.get('page_size')
        search_after = request_json.get('search_after')
        query_text = request_json.get('query_text')
        # 获取分页文件内容列表
        response_info = kb_utils.get_file_content_list(user_id, kb_info, file_name, page_size, search_after,
                                                       query_text=query_text)
        headers = {'Access-Control-Allow-Origin': '*'}
        if response_info['code'] == 0:
            kb_utils.replace_minio_ip(response_info['data']['content_list'], only_meta=True)
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/get-child-content-list", methods=['POST'])
@validate_request
def getChildContentList(request_json=None):
    logger.info('---------------获取某个文件的子文本分块列表---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('file_name')
        chunk_id = request_json.get('chunk_id')

        response_info = kb_utils.get_file_child_content_list(user_id, kb_info, file_name, chunk_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-add-chunks", methods=['POST'])
@validate_request
def batchAddChunks(request_json=None):
    logger.info('---------------批量新增文本分块---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        max_sentence_size = request_json.get('max_sentence_size')
        chunks = request_json.get('chunks')
        split_type = request_json.get("split_type", "common")
        child_chunk_config = request_json.get("child_chunk_config", None)

        if not chunks or not isinstance(chunks, list):
            raise ValueError("chunks must be a list and not empty")
        if split_type == "parent_child" and not child_chunk_config:
            raise ValueError("child_chunk_config should not be None when split_type is parent_child")
        response_info = chunk_utils.batch_add_chunks(user_id, kb_info, file_name, max_sentence_size, chunks,
                                         split_type = split_type,
                                         child_chunk_config=child_chunk_config)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-add-child-chunks", methods=['POST'])
@validate_request
def batchAddChildChunks(request_json=None):
    logger.info('---------------批量新增文本子分块---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        chunk_id = request_json.get('chunk_id')
        child_contents = request_json.get("child_contents", None)

        if not child_contents or not isinstance(child_contents, list):
            raise ValueError("child_contents must be a list and not empty")

        response_info = chunk_utils.batch_add_child_chunks(user_id, kb_info, file_name, chunk_id,
                                                           child_contents = child_contents)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-chunk", methods=['POST'])
@validate_request
def updateChunk(request_json=None):
    logger.info('---------------更新分段---------------')
    logger.info(f"更新分段参数：{request_json}")
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        max_sentence_size = request_json.get('max_sentence_size')
        chunk = request_json.get('chunk', None)
        split_type = request_json.get("split_type", "common")
        child_chunk_config = request_json.get("child_chunk_config", None)

        if not chunk or not isinstance(chunk, dict):
            raise ValueError("chunk must be a dict and not empty")

        if "labels" in chunk and not isinstance(chunk["labels"], list):
            raise ValueError("labels must be a list")

        if split_type == "parent_child" and not child_chunk_config:
            raise ValueError("child_chunk_config should not be None when split_type is parent_child")

        response_info = chunk_utils.update_chunk(user_id, kb_info, file_name, max_sentence_size, chunk,
                                                 split_type=split_type,
                                                 child_chunk_config=child_chunk_config)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-child-chunk", methods=['POST'])
@validate_request
def updateChildChunk(request_json=None):
    logger.info('---------------更新子分段---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        child_chunk = request_json.get('child_chunk', None)
        chunk_id = request_json.get('chunk_id')
        chunk_current_num = request_json.get('chunk_current_num')

        if not child_chunk or not isinstance(child_chunk, dict):
            raise ValueError("child_chunk must be a dict and not empty")

        response_info = chunk_utils.update_child_chunk(user_id, kb_info, file_name, chunk_id, chunk_current_num, child_chunk)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-chunks", methods=['POST'])
@validate_request
def batchDeleteChunks(request_json=None):
    logger.info('---------------批量删除文本分段---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        chunk_ids = request_json.get('chunk_ids', [])

        if not chunk_ids or not isinstance(chunk_ids, list):
            raise ValueError("chunk_ids must be a list and not empty")
        response_info = chunk_utils.batch_delete_chunks(user_id, kb_info, file_name, chunk_ids)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-child-chunks", methods=['POST'])
@validate_request
def batchDeleteChildChunks(request_json=None):
    logger.info('---------------批量删除文本子分段---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        chunk_id = request_json.get('chunk_id')
        chunk_current_num = request_json.get('chunk_current_num')
        child_chunk_current_nums = request_json.get('child_chunk_current_nums', [])

        if not child_chunk_current_nums or not isinstance(child_chunk_current_nums, list):
            raise ValueError("child_chunk_current_nums must be a list and not empty")
        response_info = chunk_utils.batch_delete_child_chunks(user_id, kb_info, file_name, chunk_id, chunk_current_num,
                                                           child_chunk_current_nums)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-content-status", methods=['POST'])
@validate_request
def updateContentStatus(request_json=None):
    logger.info('---------------更新文本分块状态---------------')
    try:

        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        file_name = request_json.get('fileName')
        content_id = request_json.get('content_id')
        status = request_json.get('status')
        on_off_switch = request_json.get('on_off_switch', None)  # 没有传递则默认为 None
        response_info = kb_utils.update_content_status(user_id, kb_info, file_name, content_id, status, on_off_switch)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-kb-name", methods=['POST'])
@validate_request
def updateKbName(request_json=None):
    logger.info('---------------更新知识库名接口---------------')
    try:
        user_id = request_json.get('userId')
        old_kb_name = request_json.get('old_kb_name')
        new_kb_name = request_json.get('new_kb_name')
        response_info = kb_utils.update_kb_name(user_id, old_kb_name, new_kb_name)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/del-knowledge-cache", methods=["POST"])
@validate_request
def del_knowledge_cache(request_json=None):
    logger.info('---------------删除知识库数据飞轮缓存---------------')
    try:
        user_id = request_json.get("userId")
        kb_name = request_json.get("knowledgeBase")

        logger.info(repr(request_json))

        assert len(user_id) > 0
        assert len(kb_name) > 0
        prefix = "%s^%s^" % (user_id, kb_name)
        result_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)

    except Exception as e:
        logger.error(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

def truncate_filename(filename, max_length=200):
    """
    从后往前截取文件名，确保其长度不超过 max_length 并保留扩展名
    :param filename: 原始文件名
    :param max_length: 最大允许的文件名长度，默认为 200
    :return: 截断后的文件名
    """
    base, ext = os.path.splitext(filename)

    if len(base) + len(ext) <= max_length:
        return filename

    # 从后往前截取255个字符，确保保留扩展名
    truncated_base = base[-(max_length - len(ext)):]
    return truncated_base + ext

@app.route('/rag/doc_parser', methods=['POST'])
@validate_request
def doc_parser(request_json=None):
    logger.info('---------------给定文件解析内容并切分，返回切分的chunklist---------------')
    response_info = {
        'code': 200,
        "message": "",
        "docs": []
    }
    parser_data_path = './parser_data/'
    max_length = 200
    try:
        download_link = request_json.get("url", '')
        if not download_link:
            response_info['code'] = 0
            response_info['message'] = "文件下载链接为空！"
            json_str = json.dumps(response_info, ensure_ascii=False)
            response = make_response(json_str)
            response.headers['Access-Control-Allow-Origin'] = '*'
            return response
        parsed_url = urllib.parse.urlparse(download_link)
        file_name = parsed_url.path.split('/')[-1]
        # 截断文件名:当文件名过长，超出系统允许的最大长度时，请从后往前截取200个字符
        truncated_file_name = truncate_filename(file_name)
        logger.info("---------->truncated_file_name=%s" % truncated_file_name)
        # file_path = os.path.join(parser_data_path, user_id, kb_name)
        file_path = parser_data_path + truncated_file_name

        file_response = requests.get(download_link, verify=False)

        with open(file_path, "wb") as file:
            file.write(file_response.content)
        overlap_size = request_json.get('overlap_size', 0)
        sentence_size = request_json.get('sentence_size', 8096)
        separators = request_json.get('separators', ['。'])
        parser_choices = request_json.get('parser_choices', ['text', 'model'])
        ocr_model_id = request_json.get('ocr_model_id',"")
        chunk_type = 'split_by_design'

        split_config = file_utils.SplitConfig(
            sentence_size=sentence_size,
            overlap_size=overlap_size,
            chunk_type=chunk_type,
            separators=separators,
            parser_choices=parser_choices,
            ocr_model_id=ocr_model_id,
            asr_model_id = "",
            multimodal_model_id = ""
        )
        status, chunks, filename = file_utils.split_chunks_for_parser(file_path, split_config)

        if status:
            response_info['code'] = 200
            response_info['message'] = "解析成功！"
            response_info['docs'] = chunks
        else:
            response_info['code'] = 0
            response_info['message'] = "解析失败！"
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'
    except Exception as err:
        import traceback
        print("====> call error %s" % err)
        print(traceback.format_exc())
        logger.info(traceback.format_exc())
        logger.error('doc_parser请求异常：' + repr(err))
        response_info['message'] = traceback.format_exc()
        # response_info = {'code': 1, "message": repr(e), "data": {"prompt": "", "searchList": []}}
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    return response


@app.route('/rag/user_feedback', methods=['POST'])
@validate_request
def user_feedback(request_json=None):
    logger.info('--------------rag点赞与/点踩,用户反馈接口---------------')
    response_info = {
        'code': 200,
        "message": ""
    }

    try:
        msg_id = request_json.get("msg_id", "")
        action = request_json.get("action", "")  # like:点赞；stomp：点踩; cancel：取消
        answer = request_json.get("answer", "") # 答案
        error_type = request_json.get("error_type", "") #all_error:全部错误; part_error:部分错误; other:其他
        other_reason = request_json.get("other_reason", "") # 其他原因说明
        source = request_json.get("source", "") # 调用来源: ChatConsult 或 Agent 值为空可能API调用
        # 是否开启数据飞轮
        data_flywheel = request_json.get("data_flywheel", False)
        if msg_id and action:
            u_condition = {'id': msg_id}
            data = {}
            data["action"] = action
            data["error_type"] = error_type
            data["other_reason"] = other_reason
            data["source"] = source
            data["answer"] = answer
            # data["status"] = 1
            data["update_time"] = int(round(time.time() * 1000))
            cur_count = collection.count_documents(u_condition)

            if cur_count == 0:
                update_count = 0
            elif cur_count == 1:
                result = collection.update_one(u_condition, {'$set': data})
                update_count = result.modified_count
            elif cur_count > 1:
                result = collection.update_many(u_condition, {'$set': data})
                update_count = result.modified_count
                logger.warn("---->user_feedback,msg_id=%s,更新了%s条记录，请检查！" % (msg_id, update_count))
            response_info['msg_id'] = msg_id
            if update_count > 0:
                response_info['code'] = 200
                response_info['message'] = "反馈成功！"
            else:
                response_info['code'] = 0
                response_info['message'] = "msg_id未找到问答记录，请重新提问后再反馈！"
            if data_flywheel and action == "stomp" and cur_count > 0:
                try:
                    message = collection.find_one(u_condition,{"_id": 0})
                    status = int(message["status"])
                    if status == 0:
                        kafka_utils.push_kafka_msg(message)
                        collection.update_many(u_condition, {'$set': {'status': 1, 'update_time': int(round(time.time() * 1000))}})
                        logger.info("--->反馈badcase:msg_id:%s,已推送kakfa数据 %s" % (msg_id, json.dumps(message, ensure_ascii=False)))
                    else:
                        logger.info("user_feedback,msg_id=%s,上次已推送过kafka，不再重复推送" % msg_id)
                except Exception as err:
                    import traceback
                    # print("====> call error %s" % err)
                    # print(traceback.format_exc())
                    logger.info('user_feedback push data error:%s' % traceback.format_exc())
                    logger.error('user_feedback msg_id %s,推送kafka异常：' + msg_id)
        else:
            response_info['code'] = 0
            response_info['message'] = "反馈id和动作不能为空！"

        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    except Exception as err:
        import traceback
        # print("====> call error %s" % err)
        # print(traceback.format_exc())
        logger.info(traceback.format_exc())
        logger.error('user_feedback请求异常：' + repr(err))
        response_info['message'] = "操作失败，请稍后重试！"
        response_info['code'] = 0
        # response_info = {'code': 1, "message": repr(e), "data": {"prompt": "", "searchList": []}}
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    return response


@app.route('/rag/proper_noun', methods=['POST'])
@validate_request
def proper_noun(request_json=None):
    logger.info('--------------平台专名词表同步更新至redis接口---------------')
    response_info = {
        'code': 200,
        "message": ""
    }

    try:
        msg_id = int(request_json.get("id", "-1"))
        # user_id = init_info.get("user_id", "")
        action = request_json.get("action", "")  # add：新增；delete:删除; update:修改
        name = request_json.get("name", "")  # 专名词
        alias = request_json.get("alias", [])  # 别名词表
        knowledge_base_info = request_json.get("knowledge_base_info", {})
        if knowledge_base_info:  # 整理格式
            for user_id, kb_info_list in knowledge_base_info.items():
                knowledge_base_info[user_id] = [kb_info['kb_id'] for kb_info in kb_info_list]
        logger.info(f"edit knowledge_base_info:{knowledge_base_info}")
        if msg_id and action and knowledge_base_info:  # 注意 knowledge_base 里是 kb_ids
            for user_id, knowledge_base in knowledge_base_info.items():
                try:
                    item_entry = {"id": msg_id, "name": name, "alias": alias}
                    if action == "add":
                        redis_utils.add_query_dict_entry(redis_client, user_id, item_entry, knowledge_base)
                    elif action == "delete":
                        redis_utils.delete_query_dict_entry(redis_client, user_id, msg_id, knowledge_base)
                    elif action == "update":
                        redis_utils.update_query_dict_entry(redis_client, user_id, msg_id, item_entry, knowledge_base)
                    response_info['code'] = 200
                    response_info['message'] = "操作成功！"
                    logger.info("proper_noun already update redis-cache,user_id=%s,action=%s,item_entry=%s" %
                                (user_id,action,json.dumps(item_entry, ensure_ascii=False)))
                except Exception as err:
                    logger.warn(f"syn proper_noun cache Failed: {err}")
                    response_info['code'] = 0
                    response_info['message'] = "同步专名词缓存异常！"
                    import traceback
                    logger.error(traceback.format_exc())
                    break
        else:
            response_info['code'] = 0
            response_info['message'] = "必选参数缺失，请检查！"

        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    except Exception as err:
        import traceback
        # print("====> call error %s" % err)
        # print(traceback.format_exc())
        logger.info(traceback.format_exc())
        logger.error('proper_noun请求异常：' + repr(err))
        response_info['message'] = "操作失败，请稍后重试！"
        response_info['code'] = 0
        # response_info = {'code': 1, "message": repr(e), "data": {"prompt": "", "searchList": []}}
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    return response


@app.route("/rag/batch-add-reports", methods=['POST'])
@validate_request
def batchAddReports(request_json=None):
    logger.info('---------------批量新增社区报告---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        reports = request_json.get("reports", None)

        if not reports or not isinstance(reports, list):
            raise ValueError("reports must be a list and not empty")

        response_info = graph_utils.batch_add_community_reports(user_id, kb_info, reports)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-report", methods=['POST'])
@validate_request
def updateReport(request_json=None):
    logger.info('---------------更新community report---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        report = request_json.get('reports', None)

        if not report or not isinstance(report, dict):
            raise ValueError("reports must be a dict and not empty")

        response_info = graph_utils.update_community_reports(user_id, kb_info, report)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-reports", methods=['POST'])
@validate_request
def batchDeleteReports(request_json=None):
    logger.info('---------------批量删除community reports---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        report_ids = request_json.get('report_ids', [])

        if not report_ids or not isinstance(report_ids, list):
            raise ValueError("report_ids must be a list and not empty")
        response_info = graph_utils.batch_delete_community_reports(user_id, kb_info, report_ids)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/get-community-report-list", methods=['POST'])
@validate_request
def getReportsList(request_json=None):
    logger.info('---------------获取community reports列表---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}
        page_size = request_json.get('page_size')
        search_after = request_json.get('search_after')
        # 获取分页文件内容列表
        response_info = graph_utils.get_community_report_list(user_id, kb_info, page_size, search_after)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/knowledgeBase-graph", methods=['POST'])
@validate_request
def knowledgeBaseGraph(request_json=None):
    logger.info('---------------获取知识库知识图谱---------------')
    try:
        user_id = request_json.get('userId')
        kb_name = request_json.get("knowledgeBase", "")
        kb_id = request_json.get("kb_id")
        if not kb_id:
            raise ValueError("kb_id is required")
        kb_info = {"kb_name": kb_name, "kb_id": kb_id}

        graph_data = graph_utils.get_kb_graph_data(user_id, kb_info)
        response_info = {'code': 0, "message": "", "data": graph_data}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# ================= QA 知识库相关接口 =================
# 1. 创建问答库
@app.route("/rag/init-QA-base", methods=['POST'])
@validate_request
def init_qa_base(request_json=None):
    """ 创建问答库 """
    logger.info('---------------初始化问答库---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        embedding_model_id = request_json.get("embedding_model_id")
        logger.info(f"[init_qa_base] uid={user_id}, qa_base={qa_base}, qa_id={qa_id}, embed_id={embedding_model_id}")
        response_info = qa_index.init_qa_base(user_id, qa_base, qa_id, embedding_model_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 2. 删除问答库
@app.route("/rag/delete-QA-base", methods=['POST'])
@validate_request
def delete_qa_base(request_json=None):
    """ 删除问答库 """
    logger.info('---------------删除问答库---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        logger.info(f"[delete_qa_base] uid={user_id}, base={qa_base}, qaid={qa_id}")
        response_info = qa_index.delete_qa_base(user_id, qa_base, qa_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 3. 批量新增问答对
@app.route("/rag/batch-add-QAs", methods=['POST'])
@validate_request
def batch_add_qas(request_json=None):
    """ 批量新增问答对 """
    logger.info('-------------- 批量新增问答对---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        qa_pairs = request_json.get("QAPairs")
        logger.info(f"[batch_add_qas] uid={user_id}, base={qa_base}, qaid={qa_id}, count={len(qa_pairs)}")
        response_info = qa_index.batch_add_qas(user_id, qa_base, qa_id, qa_pairs)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 4. 查看问答对列表
@app.route("/rag/get-QA-list", methods=['POST'])
@validate_request
def get_qa_list(request_json=None):
    """ 查看问答对列表 """
    logger.info('---------------分页获取问答对列表---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        page_size = request_json.get("page_size")
        search_after = request_json.get("search_after")
        logger.info(
            f"[get_qa_list] uid={user_id}, base={qa_base}, qaid={qa_id}, size={page_size}, after={search_after}")
        response_info = qa_index.get_qa_list(user_id, qa_base, qa_id, page_size, search_after)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 5. 更新问答对
@app.route("/rag/update-QA", methods=['POST'])
@validate_request
def update_qa(request_json=None):
    """ 更新问答对 """
    logger.info('---------------更新问答对---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        qa_pair = request_json.get("QAPair")
        logger.info(f"[update_qa] uid={user_id}, base={qa_base}, qaid={qa_id}, qa_pair={qa_pair}")
        response_info = qa_index.update_qa(user_id, qa_base, qa_id, qa_pair)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 6. 删除问答对
@app.route("/rag/batch-delete-QAs", methods=['POST'])
@validate_request
def batch_delete_qas(request_json=None):
    """ 批量删除问答对 """
    logger.info('---------------批量删除问答对---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        qa_pair_ids = request_json.get("QAPairIds")
        logger.info(f"[batch_delete_qas] uid={user_id}, base={qa_base}, qaid={qa_id}, ids={qa_pair_ids}")
        response_info = qa_index.batch_delete_qas(user_id, qa_base, qa_id, qa_pair_ids)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 7. 启停问答对
@app.route("/rag/update-QA-status", methods=['POST'])
@validate_request
def update_qa_status(request_json=None):
    """ 启停问答对 """
    logger.info('---------------更新问答对启停状态---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        qa_pair_id = request_json.get("QAPairId")
        status = request_json.get("status")
        logger.info(
            f"[update_qa_status] uid={user_id}, base={qa_base}, qaid={qa_id}, pair={qa_pair_id}, status={status}")
        response_info = qa_index.update_qa_status(user_id, qa_base, qa_id, qa_pair_id, status)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 8.1 批量更新元数据
@app.route("/rag/update-QA-metas", methods=['POST'])
@validate_request
def update_qa_metas(request_json=None):
    """ 批量更新元数据 """
    logger.info('---------------更新问答对元数据---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        metas = request_json.get("metas")
        logger.info(f"[update_qa_metas] uid={user_id}, base={qa_base}, qaid={qa_id}, metas={len(metas)}")
        response_info = qa_index.update_qa_metas(user_id, qa_base, qa_id, metas)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 8.2 批量删除元数据
@app.route("/rag/delete-QA-meta-by-keys", methods=['POST'])
@validate_request
def delete_qa_meta_by_keys(request_json=None):
    """ 批量删除元数据 """
    logger.info('---------------删除问答库元数据key---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        keys = request_json.get("keys")
        logger.info(f"[delete_meta_by_keys] uid={user_id}, base={qa_base}, qaid={qa_id}, keys={keys}")
        response_info = qa_index.delete_meta_by_keys(user_id, qa_base, qa_id, keys)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


# 8.3 批量重命名元数据 Key
@app.route("/rag/rename-QA-meta-keys", methods=['POST'])
@validate_request
def rename_qa_meta_keys(request_json=None):
    """ 批量重命名元数据 Key """
    logger.info('---------------重命名问答库元数据key name---------------')
    try:
        user_id = request_json.get('userId')
        qa_base = request_json.get("QABase")
        qa_id = request_json.get("QAId")
        if not qa_id:
            raise ValueError("QAId is required")
        mappings = request_json.get("mappings")
        logger.info(f"[rename_meta_keys] uid={user_id}, base={qa_base}, qaid={qa_id}, mappings={mappings}")
        response_info = qa_index.rename_meta_keys(user_id, qa_base, qa_id, mappings)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/search-QA-base", methods=["POST"])  # 查询 done
@validate_request
def search_qa_base(request_json=None):
    logger.info('---------------问答库问题查询---------------')
    try:
        return_meta = request_json.get("returnMeta", False)
        qa_base_info = request_json.get("QABaseInfo", {})
        question = request_json.get('question')
        rate = float(request_json.get('threshold', 0))
        top_k = int(request_json.get('topK', 5))
        # 是否query改写
        rewrite_query = request_json.get("rewriteQuery", False)
        rerank_mod = request_json.get("rerankMod", "rerank_model")
        rerank_model_id = request_json.get("rerankModelId", '')
        weights = request_json.get("weights", None)
        retrieve_method = request_json.get("retrieveMethod", "hybrid_search")

        #metadata filtering params
        metadata_filtering = request_json.get("metadataFiltering", False)
        metadata_filtering_conditions = request_json.get("metadataFilteringConditions", [])
        if not metadata_filtering:
            metadata_filtering_conditions = []
        logger.info(repr(request_json))

        # 检查 qa_base_info 是否为空
        if not qa_base_info:
            raise ValueError("qa_base_info cannot be empty")
        for _uid, _qa_info_list in qa_base_info.items():
            for _qa_info in _qa_info_list:
                if not _qa_info.get("QAId"):
                    raise ValueError(f"QAId is required in qa_base_info for user {_uid}")
        # 检查 rerank_model_id 是否为空
        if rerank_mod == "rerank_model" and not rerank_model_id:
            raise ValueError("rerank_model_id cannot be empty when using model-based reranking.")

        if rerank_mod == "weighted_score" and weights is None:
            raise ValueError("weights cannot be empty when using weighted score reranking.")
        if weights is not None and not isinstance(weights, dict):
            raise ValueError("weights must be a dictionary or None.")

        if rerank_mod == "weighted_score" and retrieve_method != "hybrid_search":
            raise ValueError("Weighted score reranking is only supported in hybrid search mode.")

        if len(question) <= 0:
            raise ValueError("empty question")

        if rewrite_query:
            for user_id, qa_info_list in qa_base_info.items():
                qa_base_names = [qa_info['QABase'] for qa_info in qa_info_list]
                qa_base_ids = [qa_info['QAId']  for qa_info in qa_info_list]
                query_dict_list = get_query_dict_cache(redis_client, user_id, qa_base_ids)
                if query_dict_list:
                    rewritten_queries = query_rewrite(question, query_dict_list)
                    logger.info("对query进行改写,原问题:%s 改写后问题:%s" % (question, ",".join(rewritten_queries)))
                    if len(rewritten_queries) > 0:
                        question = rewritten_queries[0]
                        logger.info("按新问题:%s 进行召回" % question)
                else:
                    logger.info("未启用或维护转名词表,query未改写,按原问题:%s 进行召回" % question)

        response_info = qa_search.search_qa_base(question, top_k, rate, return_meta, retrieve_method = retrieve_method,
                                                   rerank_model_id=rerank_model_id, rerank_mod=rerank_mod,
                                                   weights=weights, metadata_filtering_conditions=metadata_filtering_conditions,
                                                   qa_base_info=qa_base_info)
        json_str = json.dumps(response_info, ensure_ascii=False)

        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"prompt": "", "searchList": []}}
        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
        response.headers['Access-Control-Allow-Origin'] = '*'

    return response


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int)
    args = parser.parse_args()
    app.run(host='0.0.0.0', port=args.port)
