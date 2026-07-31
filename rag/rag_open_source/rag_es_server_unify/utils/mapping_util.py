from utils.config_util import es
from log.logger import logger

vector_dynamic_templates = [
    {
        "vector_512": {
            "match": "*_512_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 512,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_768": {
            "match": "*_768_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 768,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_1024": {
            "match": "*_1024_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 1024,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_1536": {
            "match": "*_1536_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 1536,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_2048": {
            "match": "*_2048_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 2048,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_2560": {
            "match": "*_2560_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 2560,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    },
    {
        "vector_4096": {
            "match": "*_4096_content_vector",
            "mapping": {
                "type": "dense_vector",
                "dims": 4096,
                "element_type": "float",
                "index": True,
                "similarity": "cosine",
                "index_options": {
                    "type": "hnsw",
                    "m": 16,
                    "ef_construction": 100
                }
            }
        }
    }
]
mappings = {
    "dynamic_templates": vector_dynamic_templates,
    "properties": {
        "content_id": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},  # 指定分词方式
        "embedding_content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        # 指定分词方式
    }
}
uk_mappings = {
    "properties": {
        "index_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "userId": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "kb_id": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "embedding_model_id": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合"
        "enable_graph": {"type": "boolean"},
    }
}
# ES 需提前 init_kb 添加 content中控部分索引
cc_mappings = {
    "properties": {
        "content_id": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "chunk_id": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "status": {"type": "boolean"},  # 指定为 keyword，方便用于排序和聚合
        "labels": {"type": "keyword"},
        "content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},  # 指定分词方式
        "child_chunk_total_num": {"type": "long"},
        "meta_data": {
            "properties": {
                "doc_meta": {
                    "type": "nested",
                    "properties": {
                        "key": {"type": "keyword"},
                        "int_value": {"type": "long"},
                        "string_value": {"type": "keyword"},
                        "value_type": {"type": "keyword"}
                    }
                }
            }
        }
    }
}

snippet_mappings = {
    "properties": {
        "snippet": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},  # 指定分词方式
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于排序和聚合
        "content_id": {"type": "keyword"}, # 指定为 keyword，方便用于排序和聚合
    }
}


file_mappings = {
    "properties": {
        "file_id": {"type": "keyword"},
        "kb_name": {"type": "keyword"},
        "file_name": {"type": "keyword"},
        "meta_data" : {
          "properties" : {
            "bucket_name" : {"type" : "keyword"},
            "chunk_total_num" : {"type" : "long"},
            "doc_meta" : {
              "type" : "nested",
              "properties" : {
                "key" : {"type" : "keyword"},
                "int_value" : {"type" : "long"},
                "string_value": {"type": "keyword"},
                "value_type" : {"type" : "keyword"}
              }
            },
            "download_link" : {"type" : "keyword"},
            "object_name" : {"type" : "keyword"}
          }
        }
    }
}

community_report_mappings = {
    "dynamic_templates": vector_dynamic_templates,
    "properties": {
        "content_id": {"type": "keyword"},
        "file_name": {"type": "keyword"},
        "kb_name": {"type": "keyword"},
        "title": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "embedding_content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "chunk_id": {"type": "keyword"},
        "status": {"type": "boolean"},
        "create_time": {"type": "keyword"},
    }
}

qa_mappings = {
    "dynamic_templates": vector_dynamic_templates,
    "properties": {
        "qa_pair_id": {"type": "keyword"},
        "question": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "answer": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "QABase": {"type": "keyword"},
        "QAId": {"type": "keyword"},
        "status": {"type": "boolean"},
        "meta_data": {
            "properties": {
                "doc_meta": {
                    "type": "nested",
                    "properties": {
                        "key": {"type": "keyword"},
                        "int_value": {"type": "long"},
                        "string_value": {"type": "keyword"},
                        "value_type": {"type": "keyword"}
                    }
                }
            }
        }
    }
}

def is_field_exist(index_name:str, field_name:str)-> (bool, dict):
    mapping = es.indices.get_mapping(index=index_name)
    properties = mapping[index_name].get('mappings', {}).get('properties', {})

    if field_name not in properties:
        return False, properties

    return True, properties

def update_doc_meta_mapping(index_name):
    meta_data_exist, properties = is_field_exist(index_name, "meta_data")

    # 取出 doc_meta 的属性
    doc_meta_props = (
        properties.get("meta_data", {})
        .get("properties", {})
        .get("doc_meta", {})
        .get("properties", {})
    )

    # 要新增的字段定义
    new_fields = {
        "int_value": {"type": "long"},
        "string_value": {"type": "keyword"},
    }
    # 检测缺失字段
    missing_fields = {
        k: v for k, v in new_fields.items() if k not in doc_meta_props
    }


    if not meta_data_exist or missing_fields:
        # 如果 meta_data 或者 doc_meta 字段不存在，添加它
        es.indices.put_mapping(
            index=index_name,
            body={
                "properties": {
                    "meta_data": {
                        "properties": {
                            "doc_meta": {
                                "type": "nested",
                                "properties": {
                                    "key": {"type": "keyword"},
                                    "int_value": {"type": "long"},
                                    "string_value": {"type": "keyword"},
                                    "value_type": {"type": "keyword"}
                                }
                            }
                        }
                    }
                }
            }
        )
        logger.info(f"已为索引 '{index_name}' 添加 doc_meta 字段映射")


def update_vector_dynamic_templates(index_name):
    """为已存在的索引补齐缺失的向量动态模板。

    动态模板在索引创建时固化，旧索引不会自动获得新增的向量模板
    （如 vector_2560）。Put Mapping 对 dynamic_templates 按名合并，
    只追加缺失模板、不覆盖已有的 vector_512/768/... ，故补缺失子集是
    安全且幂等的。补齐后，新写入的 q_{dim}_content_vector 会自动按模板
    建为可 knn 检索的 dense_vector 字段。
    """
    mapping = es.indices.get_mapping(index=index_name)
    existing_templates = mapping[index_name].get('mappings', {}).get('dynamic_templates', []) or []

    # 收敛现有模板名；dynamic_templates 每个元素形如 {name: {...}}
    existing_names = set()
    for tpl in existing_templates:
        if isinstance(tpl, dict):
            existing_names.update(tpl.keys())

    # 筛出规范的向量模板里缺失的（按名比对）
    missing_templates = [
        tpl for tpl in vector_dynamic_templates
        if isinstance(tpl, dict) and not (set(tpl.keys()) & existing_names)
    ]

    if not missing_templates:
        return

    missing_names = [list(tpl.keys())[0] for tpl in missing_templates]
    es.indices.put_mapping(
        index=index_name,
        body={"dynamic_templates": missing_templates}
    )
    logger.info(f"已为索引 '{index_name}' 补充向量动态模板: {missing_names}")
