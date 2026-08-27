from utils.config_util import es
from elasticsearch import helpers
from utils.es_util import check_index_exists
from log.logger import logger
import utils.mapping_util as es_mapping

from settings import DELETE_BACTH_SIZE
from utils.util import validate_index_name
from utils.meta_util import retype_meta_datas, build_doc_meta_query
from utils.emb_util import get_embs, get_multimodal_embs
from model.model_manager import is_multimodal_model

def delete_data_by_qa_info(index_name: str, qa_name: str, qa_id: str):
    """根据索引名和 qa_name, qa_id字段 精确匹配删除文档，并返回删除操作的状态"""
    # 构建查询条件
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_name}},
                    {"term": {"QAId": qa_id}},
                ]
            }
        }
    }
    try:
        deleted_num = 0
        # 使用 scan API 获取匹配的文档 ID
        scan_kwargs = {
            "index": index_name,
            "query": query,
            "scroll": "1m",
            "size": 100  # 每次返回的文档数量
        }
        if check_index_exists(index_name): #兼容老知识库没有file_control_xxx索引
            delete_actions = []
            for doc in helpers.scan(es, **scan_kwargs):
                delete_actions.append({
                    "_op_type": "delete",
                    "_index": index_name,
                    "_id": doc['_id']
                })
                if len(delete_actions) >= DELETE_BACTH_SIZE:
                    logger.info(f"索引 '{index_name}' qa_name:{qa_name} , 删除文档数量: {deleted_num}")
                    # 使用 bulk API 批量删除
                    res = helpers.bulk(es, delete_actions)
                    deleted_num += res[0]
                    delete_actions = []  # 清空 delete_actions
            if len(delete_actions) > 0:
                logger.info(f"索引 '{index_name}' qa_name:{qa_name} , 删除文档数量: {deleted_num}")
                # 最后的残留 bulk API 也批量删除
                res = helpers.bulk(es, delete_actions)
                deleted_num += res[0]
            es.indices.refresh(index=index_name)
        delete_status = {
            "success": True,
            "deleted": deleted_num
        }
    except Exception as e:
        delete_status = {
            "success": False,
            "error": str(e),
        }

    return delete_status


def bulk_add_index_data(index_name, qa_base_name, data):
    """使用 helpers.bulk() 批量上传数据到指定的 Elasticsearch 索引，并返回操作状态"""
    actions = []
    # 首先校验index命名是否合法
    is_index_valid, reason = validate_index_name(index_name)  # 创建普通文本类型索引
    if not is_index_valid:
        print("index invalid")
        return {"success": False, "uploaded": len(data), "error": reason}

    for item in data:
        doc_id = item['qa_pair_id']
        action = {
            "_op_type": "index",
            "_index": index_name,
            "_id": doc_id,
            "_source": item
        }
        actions.append(action)
    # 补齐缺失的向量动态模板（如 vector_2560），使旧索引新写入的向量字段自动按模板建为可 knn 检索
    es_mapping.update_vector_dynamic_templates(index_name)
    # 执行批量操作
    try:
        helpers.bulk(es, actions)
        # es.indices.refresh(index=index_name)
        logger.info(
            f"bulk_add_index_data, index_name:'{index_name}', qa_base_name:'{qa_base_name}' 添加成功。文档数量: {len(actions)}")
        return {"success": True, "uploaded": len(actions), "error": None}
    except Exception as e:
        # 专门处理批量索引错误
        error_count = len(e.errors)
        logger.error(f"批量索引失败！共 {error_count}/{len(actions)} 个文档索引失败")
        # 打印每个失败文档的详细原因
        for i, error in enumerate(e.errors[:5]):  # 最多打印前5个错误
            doc_id = error['index'].get('_id', '未指定ID')
            reason = error['index']['error']['reason']
            error_type = error['index']['error']['type']
            logger.error(f"失败文档 #{i + 1} - ID: {doc_id}")
            logger.error(f"    → 错误类型: {error_type}")
            logger.error(f"    → 原因: {reason}")
        if error_count > 5:
            logger.error(f"...... 另有 {error_count - 5} 个错误未显示 ......")

        # 如果批量操作失败，返回失败状态和错误信息
        logger.info(f"bulk_add_index_data have err, index_name:'{index_name}',qa_base_name:{qa_base_name}, item:{item}")
        import traceback
        logger.error(traceback.format_exc())
        return {"success": False, "uploaded": len(actions), "error": str(e)}


def delete_qa_ids(index_name, qa_base_name, qa_base_id, qa_pair_ids):
    """ delete_qa_ids """
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_base_name}},
                    {"term": {"QAId": qa_base_id}},
                    {"terms": {"qa_pair_id": qa_pair_ids}},
                ]
            }
        }
    }

    try:
        deleted_num = 0
        scan_kwargs = {
            "index": index_name,
            "query": query,
            "scroll": "1m",
            "size": 100  # 每次返回的文档数量
        }

        delete_actions = []
        for doc in helpers.scan(es, **scan_kwargs):
            delete_actions.append({
                "_op_type": "delete",
                "_index": index_name,
                "_id": doc['_id']
            })
            if len(delete_actions) >= DELETE_BACTH_SIZE:
                logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 删除文档数量: {deleted_num}")
                # 使用 bulk API 批量删除
                res = helpers.bulk(es, delete_actions)
                deleted_num += res[0]
                delete_actions = []  # 清空 delete_actions
        if len(delete_actions) > 0:
            logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 删除文档数量: {deleted_num}")
            # 最后的残留 bulk API 也批量删除
            res = helpers.bulk(es, delete_actions)
            deleted_num += res[0]
        es.indices.refresh(index=index_name)
        delete_status = {
            "success": True,
            "deleted": deleted_num
        }
    except Exception as e:
        delete_status = {
            "success": False,
            "error": str(e),
        }

    return delete_status


def update_qa_data(index_name, qa_base_name, qa_pair_id, upsert_data):
    """ update_qa_status"""

    actions = [{
        "_op_type": "update",
        "_index": index_name,
        "_id": qa_pair_id,
        "doc": upsert_data
    }]

    # 补齐缺失的向量动态模板（如 vector_2560），使旧索引新写入的向量字段自动按模板建为可 knn 检索
    es_mapping.update_vector_dynamic_templates(index_name)

    # 执行批量操作
    try:
        helpers.bulk(es, actions)
        es.indices.refresh(index=index_name)
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 更新成功: {upsert_data}")
        return {"success": True, "upserted": len(actions), "error": None}
    except Exception as e:
        # 如果批量操作失败，返回失败状态和错误信息
        return {"success": False, "upserted": len(actions), "error": str(e)}


def get_qa_list(index_name, qa_base_name, qa_base_id, page_size: int, search_after: int):
    """ 获取分页展示 """
    # ======== 分页查询参数 =============
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_base_name}},
                    {"term": {"QAId": qa_base_id}},
                ]
            }
        },
        #"search_after": [search_after],  # 初始化search_after参数
        "from": search_after,
        "size": page_size,
        "sort": {"qa_pair_id": {"order": "asc"}},  # 确保按照文档ID升序排序
        "_source": {
            "excludes": [
                "content_vector",
                "q_768_content_vector",
                "q_1024_content_vector",
                "q_1536_content_vector",
                "q_2048_content_vector"
            ]
        } #排除embedding数据
    }
    # 执行查询
    response = es.search(
        index=index_name,
        body=query
    )

    # 获取当前页的文档列表
    page_hits = response['hits']['hits']
    qa_list = []
    for doc in page_hits:
        qa_list.append(doc['_source'])

    # 获取匹配总数
    total_hits = response['hits']['total']['value']

    return {
        "qa_list": qa_list,
        "qa_pair_total_num": int(total_hits)
    }


def update_meta_datas(index_name, qa_base_name, qa_base_id, metas):
    """ 更新操作列表 """
    id2meta = {}
    qa_pair_ids = []

    for meta in metas:
        qa_pair_id = meta["qa_pair_id"]
        qa_pair_ids.append(qa_pair_id)
        id2meta[qa_pair_id] = meta["metadata_list"]

    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_base_name}},
                    {"term": {"QAId": qa_base_id}},
                    {"terms": {"qa_pair_id": qa_pair_ids}},
                ]
            }
        }
    }

    try:
        scan_kwargs = {
            "index": index_name,
            "query": query,
            "scroll": "1m",
            "size": 100
        }

        actions = []
        for doc in helpers.scan(es, **scan_kwargs):
            qa_pair_id = doc["_source"]["qa_pair_id"]
            data = {
                "qa_pair_id": qa_pair_id,
                "meta_data": doc["_source"].get("meta_data", {})
            }

            nested_doc_meta = retype_meta_datas(id2meta[qa_pair_id])
            data["meta_data"]["doc_meta"] = nested_doc_meta

            actions.append({
                "_op_type": "update",
                "_index": index_name,
                "_id": qa_pair_id,
                "doc": data
            })

        helpers.bulk(es, actions)
        es.indices.refresh(index=index_name)
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 更新元数据成功: {metas}")
        return {"success": True, "error": None}
    except Exception as e:
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 更新元数据失败, error: {str(e)}")
        return {"success": False, "error": str(e)}


def delete_meta_by_key(index_name, qa_base_name, qa_base_id, keys):

    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_base_name}},
                    {"term": {"QAId": qa_base_id}},
                ],
                "should": [
                    {
                        "nested": {
                            "path": "meta_data.doc_meta",
                            "query": {
                                "terms": {"meta_data.doc_meta.key": keys}
                            }
                        }
                    }
                ],
                "minimum_should_match": 1
            }
        }
    }

    scan_kwargs = {
        "index": index_name,
        "query": query,
        "scroll": "1m",
        "size": 100
    }

    try:
        actions = []
        for doc in helpers.scan(es, **scan_kwargs):
            qa_pair_id = doc["_source"]["qa_pair_id"]
            # 获取文档当前的元数据
            current_meta_data = doc["_source"].get("meta_data", {})
            current_doc_meta = current_meta_data.get("doc_meta", [])

            # 过滤掉在keys列表中的元数据key
            filtered_doc_meta = [item for item in current_doc_meta if item.get("key") not in keys]

            if len(filtered_doc_meta) != len(current_doc_meta):
                data = {
                    "qa_pair_id": qa_pair_id,
                    "meta_data": current_meta_data
                }
                data["meta_data"]["doc_meta"] = filtered_doc_meta
                actions.append({
                    "_op_type": "update",
                    "_index": index_name,
                    "_id": qa_pair_id,
                    "doc": data
                })
        helpers.bulk(es, actions)
        es.indices.refresh(index=index_name)
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 删除元数据key成功: {keys}")
        return {"success": True, "error": None}
    except Exception as e:
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , 删除元数据key失败, error: {str(e)}")
        return {"success": False, "error": str(e)}



def rename_metas(index_name, qa_base_name, qa_base_id, key_mappings):
    old_keys = [mapping["old_key"] for mapping in key_mappings]
    # 创建key映射快速查找
    key_map = {mapping["old_key"]: mapping["new_key"] for mapping in key_mappings}
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"QABase": qa_base_name}},
                    {"term": {"QAId": qa_base_id}},
                ],
                "should": [
                    {
                        "nested": {
                            "path": "meta_data.doc_meta",
                            "query": {
                                "terms": {"meta_data.doc_meta.key": old_keys}
                            }
                        }
                    }
                ],
                "minimum_should_match": 1
            }
        }
    }

    scan_kwargs = {
        "index": index_name,
        "query": query,
        "scroll": "1m",
        "size": 100
    }
    try:
        actions = []
        for doc in helpers.scan(es, **scan_kwargs):
            qa_pair_id = doc["_source"]["qa_pair_id"]
            # 获取文档当前的元数据
            current_meta_data = doc["_source"].get("meta_data", {})
            current_doc_meta = current_meta_data.get("doc_meta", [])

            # 重命名需要更改的键
            renamed_doc_meta = []
            has_changes = False

            for item in current_doc_meta:
                new_item = item.copy()
                old_key = item.get("key")

                # 如果当前键需要重命名
                if old_key in key_map:
                    new_item["key"] = key_map[old_key]
                    has_changes = True

                renamed_doc_meta.append(new_item)

            # 只有当key确实有被重命名时才添加到更新列表
            if has_changes:
                data = {
                    "qa_pair_id": qa_pair_id,
                    "meta_data": current_meta_data
                }
                data["meta_data"]["doc_meta"] = renamed_doc_meta
                actions.append({
                    "_op_type": "update",
                    "_index": index_name,
                    "_id": qa_pair_id,
                    "doc": data
                })
        helpers.bulk(es, actions)
        es.indices.refresh(index=index_name)
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , rename元数据key成功: {key_mappings}")
        return {"success": True, "error": None}
    except Exception as e:
        logger.info(f"索引 '{index_name}' qa_base_name:{qa_base_name} , rename元数据key失败, error: {str(e)}")
        return {"success": False, "error": str(e)}


def qa_rescore_bm25_score(index_name, query, search_list = []):
    qa_pair_ids = []
    for item in search_list:
        qa_pair_ids.append(item['qa_pair_id'])
    """根据content id 进行过滤，重计算bm 25得分，并按分数从高到低排序"""
    search_body = {
        "query": {
            "bool": {
                "filter": [
                    {"terms": {"qa_pair_id": qa_pair_ids}}
                ],
                "must": [
                    {"match": {"question": query}}
                ]
            }
        },
        "size": len(search_list),  # 指定返回的文档数量
        "sort": [
            {"_score": {"order": "desc"}}  # 按分数降序排序
        ],
        "_source": {
            "excludes": [
                "content_vector",
                "q_768_content_vector",
                "q_1024_content_vector",
                "q_1536_content_vector",
                "q_2048_content_vector"
            ]
        }  # 排除embedding数据
    }

    response = es.search(index=index_name, body=search_body)

    search_list = []
    scores = []
    # 遍历搜索结果，填充列表
    for hit in response['hits']['hits']:
        hit_data = hit['_source']
        hit_data["score"] = hit['_score']
        search_list.append(hit_data)
        scores.append(hit['_score'])

    # 构建结果字典
    result_dict = {
        "search_list": search_list,
        "scores": scores
    }

    return result_dict


def vector_search(index_name, base_names, query, top_k, min_score, embedding_model_id="", meta_filter_list=[]):
    """根据查询检索数据，仅返回分数高于 min_score 的文档，并按分数从高到低排序，支持多知识库"""

    if is_multimodal_model(embedding_model_id):
        query_vector = get_multimodal_embs([{"text": query}], embedding_model_id=embedding_model_id)["result"][0]["dense_vec"]
    else:
        query_vector = get_embs([query], embedding_model_id=embedding_model_id)["result"][0]["dense_vec"]
    field_name = f"q_{len(query_vector)}_content_vector"

    search_body = {
        "knn": {
            "field": field_name,
            "query_vector": query_vector,
            "filter": {
                "bool": {
                    "must": [
                        {"terms": {"QABase": base_names}},
                        {"term": {"status": True}},
                        build_doc_meta_query(meta_filter_list)
                    ]
                }
            },
            "k": top_k,
            "num_candidates": max(50, top_k),
        },
        "min_score": min_score,
        "size": top_k,  # 指定返回的文档数量
        "sort": [
            {"_score": {"order": "desc"}}  # 按分数降序排序
        ],
        "_source": {
            "excludes": [
                "content_vector",
                "q_768_content_vector",
                "q_1024_content_vector",
                "q_1536_content_vector",
                "q_2048_content_vector"
            ]
        } #排除embedding数据
    }

    response = es.search(index=index_name, body=search_body)

    search_list = []
    scores = []
    # 遍历搜索结果，填充列表
    for hit in response['hits']['hits']:
        hit_data = hit['_source']
        hit_data["score"] = hit['_score']
        search_list.append(hit_data)
        scores.append(hit['_score'])

    # 构建结果字典
    result_dict = {
        "search_list": search_list,
        "scores": scores
    }

    return result_dict

def text_search(index_name, base_names, query, top_k, min_score, meta_filter_list=[]):
    """根据查询检索数据，仅返回分数高于 min_score 的文档，并按分数从高到低排序，支持多知识库"""

    search_body = {
        "query": {
            "bool": {
                "filter": {
                    "bool": {
                        "must": [
                            {"terms": {"QABase": base_names}},
                            {"term": {"status": True}},
                            build_doc_meta_query(meta_filter_list)
                        ]
                    }
                },
                "must": [
                    {"match": {"question": query}}
                ]
            }
        },
        "min_score": min_score,
        "size": top_k,  # 指定返回的文档数量
        "sort": [
            {"_score": {"order": "desc"}}  # 按分数降序排序
        ],
        "_source": {
            "excludes": [
                "content_vector",
                "q_768_content_vector",
                "q_1024_content_vector",
                "q_1536_content_vector",
                "q_2048_content_vector"
            ]
        }  # 排除embedding数据
    }

    response = es.search(index=index_name, body=search_body)

    search_list = []
    scores = []
    # 遍历搜索结果，填充列表
    for hit in response['hits']['hits']:
        hit_data = hit['_source']
        hit_data["score"] = hit['_score']
        search_list.append(hit_data)
        scores.append(hit['_score'])

    # 构建结果字典
    result_dict = {
        "search_list": search_list,
        "scores": scores
    }

    return result_dict