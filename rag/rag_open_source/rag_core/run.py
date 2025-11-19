#!/usr/bin/env python
# -*- encoding: utf-8 -*-
import os
import json
import requests
import time

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
from know_sse import get_query_dict_cache, query_rewrite
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
from logging_config import setup_logging
from settings import MONGO_URL, USE_DATA_FLYWHEEL
# 定义路径 [EN] define path
paths = ["./parser_data"]
# 遍历路径列表 [EN] Traverse a list of paths
for path in paths:
    # 检查路径是否存在 [EN] Check if the path exists
    if not os.path.exists(path):
        # 如果不存在，则创建目录 [EN] Create the directory if it does not exist
        os.makedirs(path)
        print(f"目录 {path} 已创建。")
    else:
        print(f"目录 {path} 已存在。")

logger_name = 'rag_kb_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

app = Flask(__name__)
CORS(app, resources={r"/*": {"origins": "*"}})

app.config['JSON_AS_ASCII'] = False
app.config['JSONIFY_MIMETYPE'] = 'application/json;charset=utf-8'
# 初始化 MongoDB 客户端 [EN] Initialize the MongoDB client
client = MongoClient(MONGO_URL, 0, connectTimeoutMS=5000, serverSelectionTimeoutMS=3000)
collection = client['rag']['rag_user_logs']
redis_client = redis_utils.get_redis_connection()
chunk_label_redis_client = redis_utils.get_redis_connection(redis_db=5)

@app.route('/rag/init-knowledge-base', methods=['POST'])  # 初始化 done [EN] initialization done
def init_kb():
    logger.info('---------------初始化知识库---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")
        embedding_model_id = init_info.get("embedding_model_id", "")
        enable_knowledge_graph = init_info.get("enable_knowledge_graph", False)
        logger.info(repr(init_info))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(embedding_model_id) > 0

        result_data = kb_utils.init_knowledge_base(user_id, kb_name,
                                                   kb_id=kb_id,
                                                   embedding_model_id=embedding_model_id,
                                                   enable_knowledge_graph=enable_knowledge_graph)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False))
        # response = make_response(json.dumps(result_data, ensure_ascii=False),headers)

    except Exception as e:
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False))
    response.headers['Access-Control-Allow-Origin'] = '*'
    return response


# # ************************* 同步上传 API 接口，关闭不使用 ****************************** [EN] # ************************* Synchronous upload API interface, close and not in use ****************************

@app.route("/rag/add-knowledge-temp", methods=["POST", "GET"])  # 添加单个文件 [EN] Add a single file
def add_konwledge_temp():
    logger.info('---------------上传文件---------------')
    response_info = {
        'code': 0,
        "message": "成功"
    }
    try:
        file = request.files['file']
        user_id = request.form["userId"]
        kb_name = request.form["knowledgeBase"]
        sentence_size = int(request.form.get("sentenceSize", 500))
        separators = list(request.form.get("separators", ['。']))
        chunk_type = str(request.form.get("chunk_type", 'split_by_default'))
        overlap_size = float(request.form.get("overlap_size", 0))
        is_enhanced = request.form.get("is_enhanced", 'false')
        parser_choices = request.form.getlist("parser_choices") or ['text']
        ocr_model_id = request.form.get("ocr_model_id", "")
        pre_process = request.form.get("pre_process") or []
        meta_data_rules = request.form.get("meta_data") or []

        if file is None:
            response_info["code"] = 1
            response_info["message"] = "文件上传失败"
            json_str = json.dumps(response_info, ensure_ascii=False)
            response = make_response(json_str)
            response.headers['Access-Control-Allow-Origin'] = '*'
            return response

        # 保存上传文件 [EN] Save uploaded file
        files = [file]

        logger.info(repr(files))
        logger.info(repr(request.form))

        response_info = kb_utils.add_files(user_id, kb_name, files, sentence_size, overlap_size, chunk_type, separators,
                                  is_enhanced, parser_choices, ocr_model_id, pre_process, meta_data_rules)

        json_str = json.dumps(response_info, ensure_ascii=False)
        response = make_response(json_str)
    except Exception as e:
        import traceback
        print("====> add_konwledge error %s" % e)
        print(traceback.format_exc())
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False))
    response.headers['Access-Control-Allow-Origin'] = '*'
    return response


# @app.route("/rag/batch-add-knowledge", methods=["POST","GET"]) #添加多个文件 [EN] @app.route("/rag/batch-add-knowledge", methods=["POST","GET"]) #Add multiple files
# def batch_add_konwledge():
#     logger.info('---------------批量上传文件---------------') [EN] logger.info('---------------Batch upload files---------------')
#     response_info = {
#         'code': 0,
#         "message": "成功" [EN] "message": "success"
#     }
#     try:
#         files = request.files.getlist('files')
#         user_id = request.form["userId"]
#         kb_name = request.form["knowledgeBase"]
#         sentence_size = int(request.form.get("sentenceSize", 500))
#         overlap_size = float(request.form.get("overlap_size", 0))

#         filepath = os.path.join(user_data_path, user_id, kb_name, 'content')

#         logger.info(repr(files))
#         logger.info(repr(request.form))

#         if files is None:
#             response_info["code"] = 1
#             response_info["message"] = "文件上传失败" [EN] response_info["message"] = "File upload failed"
#             json_str = json.dumps(response_info, ensure_ascii=False)
#             response = make_response(json_str)
#             response.headers['Access-Control-Allow-Origin'] = '*'
#             return response

#         response_info = add_files(user_id,kb_name,files,sentence_size,overlap_size)
#         json_str = json.dumps(response_info, ensure_ascii=False)
#         response = make_response(json_str)
#         response.headers['Access-Control-Allow-Origin'] = '*'
#         return response

#     except Exception as e:
#         logger.info(repr(e))
#         response_info["code"] = 1
#         response_info["msg"] = repr(e)

#         json_str = json.dumps(response_info, ensure_ascii=False)
#         response = make_response(json_str)
#     response.headers['Access-Control-Allow-Origin'] = '*'
#     return response

# # ************************* 同步上传 API 接口，关闭不使用 ****************************** [EN] # ************************* Synchronous upload API interface, close and not in use ****************************

@app.route("/rag/del-knowledge-base", methods=["POST"])  # 删除知识库 done [EN] Delete knowledge base done
def del_kb():
    logger.info('---------------删除知识库---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")

        logger.info(repr(init_info))

        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        result_data = kb_utils.del_konwledge_base(user_id, kb_name, kb_id=kb_id)
        # 在批量删除文件中补充增加删除reids逻辑 begin [EN] Add deletion reids logic to the batch deletion file begin
        if USE_DATA_FLYWHEEL:
            try:
                # redis_client = redis_utils.get_redis_connection()
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"del-knowledge-base Failed to get redis connection: {err}")
                import traceback
                logger.error(traceback.format_exc())
        # 在批量删除文件中补充增加删除reids逻辑 end [EN] Add deletion reids logic to the batch deletion file end
        # ========== chunk labels 删除的逻辑 ========== [EN] ========== Chunk labels deletion logic ==========
        try:
            if not kb_id:
                kb_id = kb_utils.get_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
            # 删除chunk_labels [EN] Delete chunk_labels
            redis_utils.delete_chunk_labels(chunk_label_redis_client, kb_id)
        except Exception as err:
            logger.error(f"del-knowledge-base Failed to delete_chunk_labels: {err}")
            import traceback
            logger.error(traceback.format_exc())
        # ========== chunk labels 删除的逻辑 ========== [EN] ========== Chunk labels deletion logic ==========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-file-tags", methods=['POST'])
def updateFileTags():
    logger.info('---------------更新文件元数据---------------')
    try:
        req_info = json.loads(request.get_data())
        user_id = req_info.get('userId')
        kb_name = req_info.get("knowledgeBase", "")
        kb_id = req_info.get("kb_id", "")
        file_name = req_info.get('fileName')
        tags = req_info.get("tags", None)
        logger.info(repr(req_info))

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
        response_info = kb_utils.manage_kb_metadata(user_id, kb_name, kb_utils.MetadataOperation.UPDATE_METAS, metas, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-file-metas", methods=['POST'])
def updateFileMetas():
    logger.info('---------------批量更新文件元数据---------------')
    try:
        req_info = json.loads(request.get_data())
        user_id = req_info.get('userId')
        kb_name = req_info.get("knowledgeBase", "")
        kb_id = req_info.get("kb_id", "")
        metas = req_info.get("metas", [])
        logger.info(repr(req_info))

        if not isinstance(metas, list):
            raise ValueError("metas must be a list")
        if not metas:
            raise ValueError("metas must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_name, kb_utils.MetadataOperation.UPDATE_METAS, {"metas": metas}, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/delete-meta-by-keys", methods=['POST'])
def deleteMetaByKeys():
    logger.info('---------------知识库删除元数据key---------------')
    try:
        req_info = json.loads(request.get_data())
        user_id = req_info.get('userId')
        kb_name = req_info.get("knowledgeBase", "")
        kb_id = req_info.get("kb_id", "")
        keys = req_info.get("keys", [])
        logger.info(repr(req_info))

        if not isinstance(keys, list):
            raise ValueError("keys must be a list")
        if not keys:
            raise ValueError("keys must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_name, kb_utils.MetadataOperation.DELETE_KEYS, {"keys": keys}, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/rename-meta-keys", methods=['POST'])
def renameMetaKeys():
    logger.info('---------------重命名知识库元数据key---------------')
    try:
        req_info = json.loads(request.get_data())
        user_id = req_info.get('userId')
        kb_name = req_info.get("knowledgeBase", "")
        kb_id = req_info.get("kb_id", "")
        key_mappings = req_info.get("mappings", [])
        logger.info(repr(req_info))

        if not isinstance(key_mappings, list):
            raise ValueError("key_mappings must be a list")
        if not key_mappings:
            raise ValueError("key_mappings must be not empty")
        response_info = kb_utils.manage_kb_metadata(user_id, kb_name, kb_utils.MetadataOperation.RENAME_KEYS, {"key_mappings": key_mappings}, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-chunk-labels", methods=['POST'])
def updateChunkLabels():
    logger.info('---------------更新分片标签---------------')
    try:
        req_info = json.loads(request.get_data())
        user_id = req_info.get('userId')
        kb_name = req_info.get("knowledgeBase", "")
        kb_id = req_info.get("kb_id", "")
        file_name = req_info.get('fileName')
        chunk_id = req_info.get("chunk_id")
        labels = req_info.get("labels", None)
        logger.info(repr(req_info))

        if labels is None or not isinstance(labels, list):
            raise ValueError("labels must specified as an array")

        response_info = chunk_utils.update_chunk_labels(user_id, kb_name, file_name, chunk_id, labels, kb_id=kb_id)
        # ======= chunk labels 更新的逻辑 ======== [EN] ======= Chunk labels update logic ========
        if not kb_id:  # kb_id为空，则根据kb_name获取kb_id [EN] If kb_id is empty, get kb_id based on kb_name
            kb_id = kb_utils.get_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
        redis_utils.update_chunk_labels(chunk_label_redis_client, kb_id, file_name, chunk_id, labels)
        # ======= chunk labels 更新的逻辑 ======== [EN] ======= Chunk labels update logic ========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/search-knowledge-base", methods=["POST"])  # 查询 done [EN] Query done
def search_knowledge_base():
    logger.info('---------------问题查询---------------')
    response_info = {
        'code': 0,
        "message": "成功",
        "data": {"prompt": "",
                 "searchList": []}
    }
    try:
        init_info = json.loads(request.get_data())

        return_meta = init_info.get("return_meta", False)
        prompt_template = init_info.get("prompt_template", '')
        # user_id = init_info['userId']
        # kb_name = init_info.get("knowledgeBase", "")
        # kb_id = init_info.get("kb_id", "")
        knowledge_base_info = init_info.get("knowledge_base_info", {})
        question = init_info['question']
        rate = float(init_info.get('threshold', 0))
        top_k = int(init_info.get('topK', 5))
        chunk_conent = int(init_info.get('extend', '1'))
        chunk_size = int(init_info.get('extendedLength', CHUNK_SIZE))
        search_field = init_info.get('search_field', 'con')
        # if user_id == '': user_id = str(request.headers.get('X-Uid', ''))
        default_answer = init_info.get("default_answer", '根据已知信息，无法回答您的问题。')
        # 是否开启自动引文，此参数与prompt_template互斥，当开启auto_citation时，prompt_template用户传参不生效 [EN] Whether to enable automatic citation. This parameter is mutually exclusive with prompt_template. When auto_citation is enabled, prompt_template user-passed parameters will not take effect.
        auto_citation = init_info.get("auto_citation", False)
        # 是否query改写 [EN] Whether to rewrite query
        rewrite_query = init_info.get("rewrite_query", False)
        use_graph = init_info.get("use_graph", False)
        filter_file_name_list = init_info.get("filter_file_name_list", [])
        rerank_mod = init_info.get("rerank_mod", "rerank_model")
        # Dify开源版本问答时需指定rerank模型 [EN] The rerank model needs to be specified when answering questions in the open source version of Dify
        rerank_model_id = init_info.get("rerank_model_id", '')
        weights = init_info.get("weights", None)
        retrieve_method = init_info.get("retrieve_method", "hybrid_search")

        #metadata filtering params
        metadata_filtering = init_info.get("metadata_filtering", False)
        metadata_filtering_conditions = init_info.get("metadata_filtering_conditions", [])
        if not metadata_filtering:
            metadata_filtering_conditions = []
        logger.info(repr(init_info))

        # 检查 knowledge_base_info 是否为空 [EN] Check if knowledge_base_info is empty
        if not knowledge_base_info:
            raise ValueError("knowledge_base_info cannot be empty")
        # 检查 rerank_model_id 是否为空 [EN] Check if rerank_model_id is empty
        if rerank_mod == "rerank_model" and not rerank_model_id:
            raise ValueError("rerank_model_id cannot be empty when using model-based reranking.")

        if rerank_mod == "weighted_score" and weights is None:
            raise ValueError("weights cannot be empty when using weighted score reranking.")
        if weights is not None and not isinstance(weights, dict):
            raise ValueError("weights must be a dictionary or None.")

        if rerank_mod == "weighted_score" and retrieve_method != "hybrid_search":
            raise ValueError("Weighted score reranking is only supported in hybrid search mode.")

        # assert len(user_id) > 0
        # assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(question) > 0

        # if isinstance(kb_name, str):
        #     kb_names = [kb_name]
        # else:
        #     kb_names = kb_name
        # if isinstance(kb_id, str):
        #     kb_ids = [kb_id]
        # else:
        #     kb_ids = kb_id
        if rewrite_query:
            for user_id, kb_info_list in knowledge_base_info.items():
                kb_names = [kb_info['kb_name'] for kb_info in kb_info_list]
                kb_ids = [kb_info['kb_id'] if kb_info.get('kb_id') else kb_utils.get_kb_name_id(user_id, kb_info['kb_name']) for kb_info in kb_info_list]
                query_dict_list = get_query_dict_cache(redis_client, user_id, kb_ids)
                if query_dict_list:
                    rewritten_queries = query_rewrite(question, query_dict_list)
                    logger.info("对query进行改写,原问题:%s 改写后问题:%s" % (question, ",".join(rewritten_queries)))
                    if len(rewritten_queries) > 0:
                        question = rewritten_queries[0]
                        logger.info("按新问题:%s 进行召回" % question)
                else:
                    logger.info("未启用或维护转名词表,query未改写,按原问题:%s 进行召回" % question)

        response_info = kb_utils.get_knowledge_based_answer("", "", question, rate, top_k, chunk_conent, chunk_size,
                                                   return_meta, prompt_template, search_field, default_answer,
                                                   auto_citation, retrieve_method = retrieve_method, kb_ids=[],
                                                   filter_file_name_list=filter_file_name_list,
                                                   rerank_model_id=rerank_model_id, rerank_mod=rerank_mod,
                                                   weights=weights, metadata_filtering_conditions=metadata_filtering_conditions,
                                                   knowledge_base_info=knowledge_base_info, use_graph=use_graph)
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


@app.route("/rag/list-knowledge-base", methods=["POST"])  # 查询用户下所有的知识库名称 done [EN] Query all knowledge base names under the user done
def list_kb():
    logger.info('---------------查询所有知识库---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]

        logger.info(repr(init_info))

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


@app.route("/rag/list-knowledge-file", methods=["POST"])  # 显示用户知识库下所有的文件 done [EN] Show all files under user knowledge base done
def list_file():
    logger.info('---------------查询所有知识库文件---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")

        logger.info(repr(init_info))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        response_info = kb_utils.list_knowledge_file(user_id, kb_name, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": "repr(e)", "data": {"knowledge_file_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response



@app.route("/rag/list-knowledge-file-download-link", methods=["POST"])  # 显示用户知识库下所有的文件 done [EN] Show all files under user knowledge base done
def list_file_download_link():
    logger.info('---------------查询所有知识库文件的 download_link---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")

        logger.info(repr(init_info))
        assert len(user_id) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0

        response_info = kb_utils.list_knowledge_file_download_link(user_id, kb_name, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": "repr(e)", "data": {"knowledge_file_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/del-knowledge-file", methods=["POST"])  # 删除某个知识库下的单个文件 done [EN] Delete a single file under a knowledge base done
def del_file():
    logger.info('---------------删除知识库文件---------------')
    try:
        init_info = json.loads(request.get_data())
        file_name = init_info["fileName"]
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")

        logger.info(repr(init_info))

        assert len(file_name) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(user_id) > 0

        result_data = kb_utils.del_knowledge_base_files(user_id, kb_name, [file_name], kb_id=kb_id)
        # 在批量删除文件中补充增加删除reids逻辑 begin [EN] Add deletion reids logic to the batch deletion file begin
        if USE_DATA_FLYWHEEL:
            try:
                # redis_client = redis_utils.get_redis_connection()
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"del-knowledge-file Failed to get redis connection: {err}")
                import traceback
                logger.error(traceback.format_exc())
        # 在批量删除文件中补充增加删除reids逻辑 end [EN] Add deletion reids logic to the batch deletion file end
        # ========== chunk labels 删除的逻辑 ========== [EN] ========== Chunk labels deletion logic ==========
        try:
            kb_id = kb_utils.get_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
            # 删除chunk_labels [EN] Delete chunk_labels
            redis_utils.delete_chunk_labels(chunk_label_redis_client, kb_id, file_name=file_name)
        except Exception as err:
            logger.error(f"del-knowledge-file Failed to delete_chunk_labels: {err}")
            import traceback
            logger.error(traceback.format_exc())
        # ========== chunk labels 删除的逻辑 ========== [EN] ========== Chunk labels deletion logic ==========
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)

    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/batch_del_knowfiles", methods=["POST"])  # 删除某个知识库下的多个文件 done [EN] Delete multiple files under a knowledge base done
def del_files():
    logger.info('---------------批量删除知识库文件---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")
        file_names = init_info["fileNames"]

        logger.info(repr(init_info))

        assert len(file_names) > 0
        assert len(kb_name) > 0 or len(kb_id) > 0
        assert len(user_id) > 0

        result_data = kb_utils.del_knowledge_base_files(user_id, kb_name, file_names, kb_id=kb_id)
        # 在批量删除文件中补充增加删除reids逻辑 begin [EN] Add deletion reids logic to the batch deletion file begin
        if USE_DATA_FLYWHEEL:
            try:
                # redis_client = redis_utils.get_redis_connection()
                prefix = "%s^%s^" % (user_id, kb_name)
                redis_data = redis_utils.delete_cache_by_prefix(redis_client, prefix)
                logger.info("clean flywheel cache result:%s" % json.dumps(redis_data, ensure_ascii=False))
            except Exception as err:
                logger.warn(f"Failed to get redis connection maybe not use dataflywheel or uninstall redis: {err}")
        # 在批量删除文件中补充增加删除reids逻辑 end [EN] Add deletion reids logic to the batch deletion file end
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(result_data, ensure_ascii=False), headers)

    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/check-knowledge-base", methods=["POST"])  # 查询某个知识库是否在某个用户下 done [EN] Query whether a certain knowledge base is under a certain user done
def check_kb():
    logger.info('---------------校验知识库是否存在---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info.get("knowledgeBase", "")
        kb_id = init_info.get("kb_id", "")

        logger.info(repr(init_info))

        assert len(user_id) > 0

        response_info = kb_utils.check_knowledge_base(user_id, kb_name, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"knowledge_base_names": []}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/split_text", methods=['POST'])
def split_text():
    data = request.get_json()
    text = data.get('text', '')
    chunk_type = data.get('chunk_type')
    sentence_size = data.get('sentence_size', 500)
    overlap_size = data.get('overlap_size', 0.2)
    separators = data.get('separators', ["。", "！", "？", ".", "!", "?", "……", "|\n"])
    pdf = data.get('pdf', False)
    excel = data.get('excel', False)

    splitter = ChineseTextSplitter(chunk_type=chunk_type, sentence_size=sentence_size, overlap_size=overlap_size,
                                   pdf=pdf, excel=excel, separators=separators)
    result = splitter.split_text(text)

    return jsonify(result)


@app.route("/rag/get-content-list", methods=['POST'])
def getContentList():
    logger.info('---------------获取某个文件的文本分块列表---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        page_size = data['page_size']
        search_after = data['search_after']
        # 获取分页文件内容列表 [EN] Get a list of paginated file contents
        response_info = kb_utils.get_file_content_list(user_id, kb_name, file_name, page_size, search_after, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/get-child-content-list", methods=['POST'])
def getChildContentList():
    logger.info('---------------获取某个文件的子文本分块列表---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['file_name']
        chunk_id = data['chunk_id']

        response_info = kb_utils.get_file_child_content_list(user_id, kb_name, file_name, chunk_id, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-add-chunks", methods=['POST'])
def batchAddChunks():
    logger.info('---------------批量新增文本分块---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        max_sentence_size = data['max_sentence_size']
        chunks = data['chunks']
        split_type = data.get("split_type", "common")
        child_chunk_config = data.get("child_chunk_config", None)

        if not chunks or not isinstance(chunks, list):
            raise ValueError("chunks must be a list and not empty")
        if split_type == "parent_child" and not child_chunk_config:
            raise ValueError("child_chunk_config should not be None when split_type is parent_child")
        response_info = chunk_utils.batch_add_chunks(user_id, kb_name, file_name, max_sentence_size, chunks,
                                         split_type = split_type,
                                         child_chunk_config=child_chunk_config,
                                         kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-add-child-chunks", methods=['POST'])
def batchAddChildChunks():
    logger.info('---------------批量新增文本子分块---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        chunk_id = data['chunk_id']
        child_contents = data.get("child_contents", None)

        if not child_contents or not isinstance(child_contents, list):
            raise ValueError("child_contents must be a list and not empty")

        response_info = chunk_utils.batch_add_child_chunks(user_id, kb_name, file_name, chunk_id,
                                                           child_contents = child_contents, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-chunk", methods=['POST'])
def updateChunk():
    logger.info('---------------更新分段---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        max_sentence_size = data['max_sentence_size']
        chunk = data.get('chunk', None)
        split_type = data.get("split_type", "common")
        child_chunk_config = data.get("child_chunk_config", None)

        if not chunk or not isinstance(chunk, dict):
            raise ValueError("chunk must be a dict and not empty")

        if "labels" in chunk and not isinstance(chunk["labels"], list):
            raise ValueError("labels must be a list")

        if split_type == "parent_child" and not child_chunk_config:
            raise ValueError("child_chunk_config should not be None when split_type is parent_child")

        response_info = chunk_utils.update_chunk(user_id, kb_name, file_name, max_sentence_size, chunk,
                                                 split_type=split_type,
                                                 child_chunk_config=child_chunk_config,
                                                 kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-child-chunk", methods=['POST'])
def updateChildChunk():
    logger.info('---------------更新子分段---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        child_chunk = data.get('child_chunk', None)
        chunk_id = data.get('chunk_id')
        chunk_current_num = data.get('chunk_current_num')

        if not child_chunk or not isinstance(child_chunk, dict):
            raise ValueError("child_chunk must be a dict and not empty")

        response_info = chunk_utils.update_child_chunk(user_id, kb_name, file_name, chunk_id, chunk_current_num, child_chunk, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-chunks", methods=['POST'])
def batchDeleteChunks():
    logger.info('---------------批量删除文本分段---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        chunk_ids = data.get('chunk_ids', [])

        if not chunk_ids or not isinstance(chunk_ids, list):
            raise ValueError("chunk_ids must be a list and not empty")
        response_info = chunk_utils.batch_delete_chunks(user_id, kb_name, file_name, chunk_ids, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-child-chunks", methods=['POST'])
def batchDeleteChildChunks():
    logger.info('---------------批量删除文本子分段---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        chunk_id = data.get('chunk_id')
        chunk_current_num = data.get('chunk_current_num')
        child_chunk_current_nums = data.get('child_chunk_current_nums', [])

        if not child_chunk_current_nums or not isinstance(child_chunk_current_nums, list):
            raise ValueError("child_chunk_current_nums must be a list and not empty")
        response_info = chunk_utils.batch_delete_child_chunks(user_id, kb_name, file_name, chunk_id, chunk_current_num,
                                                           child_chunk_current_nums, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-content-status", methods=['POST'])
def updateContentStatus():
    logger.info('---------------更新文本分块状态---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        file_name = data['fileName']
        content_id = data['content_id']
        status = data['status']
        on_off_switch = data.get('on_off_switch', None)  # 没有传递则默认为 None [EN] Defaults to None if not passed
        response_info = kb_utils.update_content_status(user_id, kb_name, file_name, content_id, status, on_off_switch, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/update-kb-name", methods=['POST'])
def updateKbName():
    logger.info('---------------更新知识库名接口---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        old_kb_name = data['old_kb_name']
        new_kb_name = data['new_kb_name']
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
def del_knowledge_cache():
    logger.info('---------------删除知识库数据飞轮缓存---------------')
    try:
        init_info = json.loads(request.get_data())
        user_id = init_info["userId"]
        kb_name = init_info["knowledgeBase"]

        logger.info(repr(init_info))

        assert len(user_id) > 0
        assert len(kb_name) > 0
        # redis_client = redis_utils.get_redis_connection()
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

    # 从后往前截取255个字符，确保保留扩展名 [EN] Trim 255 characters from back to front, making sure to keep the extension
    truncated_base = base[-(max_length - len(ext)):]
    return truncated_base + ext

@app.route('/rag/doc_parser', methods=['POST'])
def doc_parser():
    logger.info('---------------给定文件解析内容并切分，返回切分的chunklist---------------')
    response_info = {
        'code': 200,
        "message": "",
        "docs": []
    }
    parser_data_path = './parser_data/'
    max_length = 200
    try:
        init_info = json.loads(request.get_data())
        download_link = init_info.get("url", '')
        if not download_link:
            response_info['code'] = 0
            response_info['message'] = "文件下载链接为空！"
            json_str = json.dumps(response_info, ensure_ascii=False)
            response = make_response(json_str)
            response.headers['Access-Control-Allow-Origin'] = '*'
            return response
        parsed_url = urllib.parse.urlparse(download_link)
        file_name = parsed_url.path.split('/')[-1]
        # 截断文件名:当文件名过长，超出系统允许的最大长度时，请从后往前截取200个字符 [EN] Truncate the file name: When the file name is too long and exceeds the maximum length allowed by the system, please truncate 200 characters from the back to the front.
        truncated_file_name = truncate_filename(file_name)
        logger.info("---------->truncated_file_name=%s" % truncated_file_name)
        # file_path = os.path.join(parser_data_path, user_id, kb_name)
        file_path = parser_data_path + truncated_file_name

        file_response = requests.get(download_link, verify=False)

        with open(file_path, "wb") as file:
            file.write(file_response.content)
        overlap_size = init_info.get('overlap_size', 0)
        sentence_size = init_info.get('sentence_size', 8096)
        separators = init_info.get('separators', ['。'])
        parser_choices = init_info.get('parser_choices', ['text','ocr'])
        ocr_model_id = init_info.get('ocr_model_id',"")
        chunk_type = 'split_by_design'

        split_config = file_utils.SplitConfig(
            sentence_size=sentence_size,
            overlap_size=overlap_size,
            chunk_type=chunk_type,
            separators=separators,
            parser_choices=parser_choices,
            ocr_model_id=ocr_model_id
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
def user_feedback():
    logger.info('--------------rag点赞与/点踩,用户反馈接口---------------')
    response_info = {
        'code': 200,
        "message": ""
    }

    try:
        init_info = json.loads(request.get_data())
        msg_id = init_info.get("msg_id", "")
        action = init_info.get("action", "")  # like:点赞；stomp：点踩; cancel：取消 [EN] like: like; stomp: click; cancel: cancel
        answer = init_info.get("answer", "") # 答案 [EN] Answer
        error_type = init_info.get("error_type", "") #all_error:全部错误; part_error:部分错误; other:其他 [EN] all_error: all errors; part_error: partial errors; other: other
        other_reason = init_info.get("other_reason", "") # 其他原因说明 [EN] Other reasons
        source = init_info.get("source", "") # 调用来源: ChatConsult 或 Agent 值为空可能API调用 [EN] Call source: ChatConsult or Agent. If the value is empty, it may be an API call.
        # 是否开启数据飞轮 [EN] Whether to enable data flywheel
        data_flywheel = init_info.get("data_flywheel", False)
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
def proper_noun():
    logger.info('--------------平台专名词表同步更新至redis接口---------------')
    response_info = {
        'code': 200,
        "message": ""
    }

    try:
        init_info = json.loads(request.get_data())
        msg_id = int(init_info.get("id", "-1"))
        # user_id = init_info.get("user_id", "")
        action = init_info.get("action", "")  # add：新增；delete:删除; update:修改 [EN] add: add; delete: delete; update: modify
        name = init_info.get("name", "")  # 专名词 [EN] proper noun
        alias = init_info.get("alias", [])  # 别名词表 [EN] alias list
        # apply_type = init_info.get("apply_type", [])  # 作用域：user 或 knowledgebase 或 user + knowledgebase [EN] apply_type = init_info.get("apply_type", []) # Scope: user or knowledgebase or user + knowledgebase
        # knowledge_base = init_info.get("knowledge_base_list", []) # 若作用域为knowledgebase需传 知识库名称列表 [EN] knowledge_base = init_info.get("knowledge_base_list", []) # If the scope is knowledgebase, a list of knowledge base names needs to be passed
        knowledge_base_info = init_info.get("knowledge_base_info", {})
        if knowledge_base_info:  # 整理格式 [EN] Format
            for user_id, kb_info_list in knowledge_base_info.items():
                knowledge_base_info[user_id] = [kb_info['kb_id'] if kb_info.get('kb_id') else kb_utils.get_kb_name_id(user_id, kb_info['kb_name']) for kb_info in kb_info_list]
        logger.info(f"edit knowledge_base_info:{knowledge_base_info}")
        if msg_id and action and knowledge_base_info:  # 注意 knowledge_base 里是 kb_ids [EN] Note that knowledge_base contains kb_ids
            for user_id, knowledge_base in knowledge_base_info.items():
                try:
                    # redis_client = redis_utils.get_redis_connection()
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
def batchAddReports():
    logger.info('---------------批量新增社区报告---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        reports = data.get("reports", None)

        if not reports or not isinstance(reports, list):
            raise ValueError("reports must be a list and not empty")

        response_info = graph_utils.batch_add_community_reports(user_id, kb_name, reports, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/update-report", methods=['POST'])
def updateReport():
    logger.info('---------------更新community report---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        report = data.get('reports', None)

        if not report or not isinstance(report, dict):
            raise ValueError("reports must be a dict and not empty")

        response_info = graph_utils.update_community_reports(user_id, kb_name, report, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/batch-delete-reports", methods=['POST'])
def batchDeleteReports():
    logger.info('---------------批量删除community reports---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        report_ids = data.get('report_ids', [])

        if not report_ids or not isinstance(report_ids, list):
            raise ValueError("report_ids must be a list and not empty")
        response_info = graph_utils.batch_delete_community_reports(user_id, kb_name, report_ids, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e), "data": {"success_count": 0}}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response


@app.route("/rag/get-community-report-list", methods=['POST'])
def getReportsList():
    logger.info('---------------获取community reports列表---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")
        page_size = data['page_size']
        search_after = data['search_after']
        # 获取分页文件内容列表 [EN] Get a list of paginated file contents
        response_info = graph_utils.get_community_report_list(user_id, kb_name, page_size, search_after, kb_id=kb_id)
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

@app.route("/rag/knowledgeBase-graph", methods=['POST'])
def knowledgeBaseGraph():
    logger.info('---------------获取知识库知识图谱---------------')
    try:
        data = request.get_json()
        user_id = data['userId']
        kb_name = data.get("knowledgeBase", "")
        kb_id = data.get("kb_id", "")

        graph_data = graph_utils.get_kb_graph_data(user_id, kb_name, kb_id=kb_id)
        response_info = {'code': 0, "message": "", "data": graph_data}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    except Exception as e:
        logger.info(repr(e))
        response_info = {'code': 1, "message": repr(e)}
        headers = {'Access-Control-Allow-Origin': '*'}
        response = make_response(json.dumps(response_info, ensure_ascii=False), headers)
    return response

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int)
    args = parser.parse_args()
    app.run(host='0.0.0.0', port=args.port)
