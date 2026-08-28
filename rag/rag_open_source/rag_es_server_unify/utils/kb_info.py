import requests
import json
import uuid

from utils.config_util import es
from log.logger import logger
from elasticsearch import helpers
from settings import KBNAME_MAPPING_INDEX


def is_multimodal_kb(user_id: str,
                     kb_name: str):
    """ 判断是否是多模态知识库 """
    kb_info = get_uk_kb_info(user_id, kb_name)
    if kb_info and "is_multimodal" in kb_info and kb_info["is_multimodal"]:
        return True

    return False


def get_uk_kb_info(userId, kb_name):
    """ 获取知识库info  """
    kb_info = {}
    logger.info(f"userId:{userId},kb_name:{kb_name} ====== get_uk_kb_info")
    # 查询条件
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": userId}},
                    {"term": {"kb_name": kb_name}},
                ]
            }
        }
    }
    response = es.search(index=KBNAME_MAPPING_INDEX, body=query)
    if len(response["hits"]["hits"]) > 1:
        raise ValueError("存在多条kb info 记录")
    for hit in response["hits"]["hits"]:
        kb_id = hit['_source']["kb_id"]
        kb_info["kb_id"] = kb_id
        kb_info["embedding_model_id"] = hit['_source']["embedding_model_id"]
        if "enable_graph" in hit['_source']:
            kb_info["enable_knowledge_graph"] = hit['_source']["enable_graph"]
        if "is_multimodal" in hit['_source']:
            kb_info["is_multimodal"] = hit['_source']["is_multimodal"]
        else:
            kb_info["is_multimodal"] = False
    logger.info(f"userId:{userId},kb_name:{kb_name} 对应的 kb_info 为:{kb_info}")
    return kb_info

def get_uk_kb_name_list(index_name, user_id):
    """ 获取 userid 的所有 kb_name 映射表下 某个 user_id 所有的知识库名称的集合"""
    body = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": user_id}},
                ],
                "must_not": [
                    {"exists": { "field": "is_qa" }}
                ]
            }
        },
        "aggs": {
            "unique_res": {
                "terms": {
                    "field": "kb_name",
                    "size": 100000,  # 根据需要设置大小
                }
            }
        },
        "size": 0  # 不需要原始文档，只用于聚合
    }

    response = es.search(index=index_name, body=body)
    unique_res = [bucket['key'] for bucket in response['aggregations']['unique_res']['buckets']]
    return unique_res


def get_uk_kb_id_list(index_name, user_id):
    """获取某个 user_id 在映射表中所有知识库的 kb_id 列表（含 QA 库）"""
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": user_id}},
                ]
            }
        },
        "_source": ["kb_id"],
    }
    return [
        doc["_source"]["kb_id"]
        for doc in helpers.scan(
            es,
            query=query,
            index=index_name,
            scroll="5m",
            size=1000,
        )
    ]


def get_uk_kb_emb_model_id(userId, kb_name):
    """ 获取知识库映射的 embedding_model_id  """
    embedding_model_id = ""
    logger.info(f"userId:{userId},kb_name:{kb_name} ====== get_uk_kb_emb_model_id")
    # 查询条件
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": userId}},
                    {"term": {"kb_name": kb_name}},
                ]
            }
        }
    }
    response = es.search(index=KBNAME_MAPPING_INDEX, body=query)
    # 遍历搜索结果，获取 kb_id
    for hit in response["hits"]["hits"]:
        embedding_model_id = hit['_source']["embedding_model_id"]
    logger.info(f"userId:{userId},kb_name:{kb_name} 对应的 embedding_model_id 为:{embedding_model_id}")
    return embedding_model_id


def update_uk_kb_name(userId, old_kb_name, new_kb_name):
    """ 更新 uk映射表 知识库名 """
    # 查询条件
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": userId}},
                    {"term": {"kb_name": old_kb_name}},
                ]
            }
        }
    }
    response = es.search(index=KBNAME_MAPPING_INDEX, body=query)
    # 遍历搜索结果，将actions
    actions = []
    for hit in response["hits"]["hits"]:
        # 往索引里插数据，以index的方式，若_id已存在则先删除再添加
        doc_id = hit['_id']
        # print(doc_id)
        action = {
            "_op_type": "update",
            "_index": KBNAME_MAPPING_INDEX,
            "_id": doc_id,
            "doc": {"kb_name": new_kb_name}
        }
        actions.append(action)
    if len(actions) < 1:
        return {'code': 1, 'message': f'没有找到对应的知识库:{old_kb_name}'}
    # 执行更新操作,并返回
    try:
        helpers.bulk(es, actions)
        es.indices.refresh(index=KBNAME_MAPPING_INDEX)
        return {'code': 0, 'message': 'success'}
    except Exception as e:
        # 如果批量操作失败，返回失败状态和错误信息
        return {'code': 1, 'message': f'{e}'}

def get_uk_qa_name_list(user_id):
    """ 获取 userid 的所有 qa_name 映射表下 某个 user_id 所有的问答库名称的集合"""
    body = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"userId": user_id}},
                    {"term": {"is_qa": True}},
                ]
            }
        },
        "aggs": {
            "unique_res": {
                "terms": {
                    "field": "kb_name",
                    "size": 100000,  # 根据需要设置大小
                }
            }
        },
        "size": 0  # 不需要原始文档，只用于聚合
    }

    response = es.search(index=KBNAME_MAPPING_INDEX, body=body)
    unique_res = [bucket['key'] for bucket in response['aggregations']['unique_res']['buckets']]
    return unique_res


def bulk_add_uk_index_data(index_name, data):
    """(用于userid 的所有 kb_name映射表索引添加数据) 使用 helpers.bulk() 批量上传数据到指定的 Elasticsearch 索引，并返回操作状态"""
    actions = []
    # ============== 直接往里添加，固定 id  ==============
    try:
        for item in data:  # 往索引里插数据，以index的方式，若_id已存在则先删除再添加
            if not item["kb_id"]:  # 如果不传递，则生成一个
                # cont_str = item["index_name"] + item["userId"] + item["kb_name"]
                # doc_id = generate_md5(cont_str)
                doc_id = uuid.uuid4()  # 不关注重复
                # print(doc_id)
                item['item_id'] = doc_id
                item['kb_id'] = doc_id
            else:
                doc_id = item["kb_id"]
                item['item_id'] = doc_id
            action = {
                "_op_type": "index",  # 使用index,已存在就覆盖
                "_index": index_name,
                "_id": doc_id,
                "_source": item
            }
            actions.append(action)

        # 执行批量操作
        helpers.bulk(es, actions)
        res = es.indices.refresh(index=index_name)

        # logger.info(f"{res}： bulk_add_uk_index_data  ----- {data}")
        return {"success": True, "uploaded": len(actions), "error": None}
    except Exception as e:
        # 如果批量操作失败，返回失败状态和错误信息
        return {"success": False, "uploaded": len(actions), "error": str(e)}
