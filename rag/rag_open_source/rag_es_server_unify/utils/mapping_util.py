from utils.config_util import es
from log.logger import logger

vector_dynamic_templates = [
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
    }
]
mappings = {
    "dynamic_templates": vector_dynamic_templates,
    "properties": {
        "content_id": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},  # 指定分词方式
        "embedding_content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        # 指定分词方式
    }
}
uk_mappings = {
    "properties": {
        "index_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "userId": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "kb_id": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "embedding_model_id": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合"
        "enable_graph": {"type": "boolean"},
    }
}
# ES 需提Before init_kb 添加 contentMedium控部分Index
cc_mappings = {
    "properties": {
        "content_id": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "chunk_id": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "status": {"type": "boolean"},  # 指定为 keyword，方便用于Sort和聚合
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
        "file_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "kb_name": {"type": "keyword"},  # 指定为 keyword，方便用于Sort和聚合
        "content_id": {"type": "keyword"}, # 指定为 keyword，方便用于Sort和聚合
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
        "content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "embedding_content": {"type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart"},
        "chunk_id": {"type": "keyword"},
        "status": {"type": "boolean"},
        "create_time": {"type": "keyword"},
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

    # 要Add的Field定义
    new_fields = {
        "int_value": {"type": "long"},
        "string_value": {"type": "keyword"},
    }
    # 检测缺失Field
    missing_fields = {
        k: v for k, v in new_fields.items() if k not in doc_meta_props
    }


    if not meta_data_exist or missing_fields:
        # If meta_data Or doc_meta Field不存在，添加它
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
