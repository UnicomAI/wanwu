import json
from abc import ABC, abstractmethod
from utils.config_util import es
from log.logger import logger
from utils.es_util import update_index_data
from utils.mapping_util import update_doc_meta_mapping
from utils.util import *
from elasticsearch import helpers

def construct_empty_query(query: dict):
    base_query = {
        "bool": {
            "must_not": query
        }
    }

    return base_query

class MetaConditionBuilder(ABC):
    """元数据条件构建器抽象基类"""

    def __init__(self, meta_name, operator, value):
        self.meta_name = meta_name
        self.operator = operator
        self.value = value

    @abstractmethod
    def build_condition(self):
        """构建具体条件查询"""
        pass


class StringConditionBuilder(MetaConditionBuilder):
    """字符串类型条件构建器"""

    def build_condition(self):
        base_query = {
            "nested": {
                "path": "meta_data.doc_meta",
                "query": {
                    "bool": {
                        "must": [
                            {"term": {"meta_data.doc_meta.key": self.meta_name}}
                        ]
                    }
                }
            }
        }

        if self.operator == "contains":
            base_query["nested"]["query"]["bool"]["must"].append({
                "wildcard": {"meta_data.doc_meta.string_value": f"*{self.value}*"}
            })
        elif self.operator == "not contains":
            base_query["nested"]["query"]["bool"]["must"].append({
                "wildcard": {"meta_data.doc_meta.string_value": f"*{self.value}*"}
            })
            base_query = { "bool": { "must_not": [base_query]}}
        elif self.operator == "start with":
            base_query["nested"]["query"]["bool"]["must"].append({
                "wildcard": {"meta_data.doc_meta.string_value": f"{self.value}*"}
            })
        elif self.operator == "end with":
            base_query["nested"]["query"]["bool"]["must"].append({
                "wildcard": {"meta_data.doc_meta.string_value": f"*{self.value}"}
            })
        elif self.operator == "is":
            base_query["nested"]["query"]["bool"]["must"].append({
                "term": {"meta_data.doc_meta.string_value": self.value}
            })
        elif self.operator == "is not":
            base_query["nested"]["query"]["bool"]["must"].append({
                "term": {"meta_data.doc_meta.string_value": self.value}
            })
            base_query = { "bool": { "must_not": [base_query]}}
        elif self.operator == "empty":
            base_query = construct_empty_query(base_query)
        elif self.operator == "not empty":
            pass

        return base_query


class NumberConditionBuilder(MetaConditionBuilder):
    """数字类型条件构建器"""

    def build_condition(self):
        base_query = {
            "nested": {
                "path": "meta_data.doc_meta",
                "query": {
                    "bool": {
                        "must": [
                            {"term": {"meta_data.doc_meta.key": self.meta_name}}
                        ]
                    }
                }
            }
        }

        if self.operator == "=":
            base_query["nested"]["query"]["bool"]["must"].append({
                "term": {"meta_data.doc_meta.int_value": self.value}
            })
        elif self.operator == "≠":
            base_query["nested"]["query"]["bool"]["must_not"] = [
                {"term": {"meta_data.doc_meta.int_value": self.value}}
            ]
        elif self.operator == ">":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"gt": self.value}}
            })
        elif self.operator == "<":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"lt": self.value}}
            })
        elif self.operator == "≥":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"gte": self.value}}
            })
        elif self.operator == "≤":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"lte": self.value}}
            })
        elif self.operator == "empty":
            base_query = construct_empty_query(base_query)
        elif self.operator == "not empty":
            pass

        return base_query


class TimeConditionBuilder(MetaConditionBuilder):
    """时间类型条件构建器"""

    def build_condition(self):
        base_query = {
            "nested": {
                "path": "meta_data.doc_meta",
                "query": {
                    "bool": {
                        "must": [
                            {"term": {"meta_data.doc_meta.key": self.meta_name}}
                        ]
                    }
                }
            }
        }

        if self.operator == "is":
            base_query["nested"]["query"]["bool"]["must"].append({
                "term": {"meta_data.doc_meta.int_value": self.value}
            })
        elif self.operator == "before":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"lt": self.value}}
            })
        elif self.operator == "after":
            base_query["nested"]["query"]["bool"]["must"].append({
                "range": {"meta_data.doc_meta.int_value": {"gt": self.value}}
            })
        elif self.operator == "empty":
            base_query = construct_empty_query(base_query)
        elif self.operator == "not empty":
            pass

        return base_query


def build_single_condition(condition):
    """根据条件类型构建查询"""
    meta_name = condition["meta_name"]
    meta_type = condition["meta_type"].lower()
    operator = condition["comparison_operator"]
    value = condition.get("value", "")

    # 创建对应的构建器
    if meta_type == "string":
        builder = StringConditionBuilder(meta_name, operator, value)
    elif meta_type == "number":
        builder = NumberConditionBuilder(meta_name, operator, value)
    elif meta_type == "time":
        builder = TimeConditionBuilder(meta_name, operator, value)
    else:
        raise ValueError(f"Unsupported meta_type: {meta_type}")

    return builder.build_condition()


def build_conditions_group(group):
    """构建条件组查询"""
    base_name = []
    field_name = ""
    if "filtering_kb_name" in group:
        base_name = group["filtering_kb_name"]
        field_name = "kb_name"

    if "filtering_qa_base_name" in group:
        base_name = group["filtering_qa_base_name"]
        field_name = "QABase"

    logical_op = group["logical_operator"].lower()
    conditions = group["conditions"]

    query = {
        "bool": {
            "must": [
                {"term": {field_name: base_name}}
            ]
        }
    }

    if logical_op == "and":
        # 对于AND操作，每个条件都作为must子句
        for condition in conditions:
            query["bool"]["must"].append(build_single_condition(condition))
    elif logical_op == "or":
        # 对于OR操作，所有条件作为should子句
        should_clauses = []
        for condition in conditions:
            should_clauses.append(build_single_condition(condition))
        query["bool"]["must"].append({
            "bool": {
                "should": should_clauses,
                "minimum_should_match": 1
            }
        })

    return query


def build_doc_meta_query(filtering_conditions):
    """
    构建元数据过滤查询

    参数:
    filtering_conditions: 过滤条件列表

    返回:
    Elasticsearch查询DSL
    """
    # 构建完整的查询
    if len(filtering_conditions) == 1:
        # 单个条件组
        return build_conditions_group(filtering_conditions[0])
    else:
        # 多个条件组，使用should连接（OR关系）
        should_clauses = []
        for group in filtering_conditions:
            should_clauses.append(build_conditions_group(group))

        return {
            "bool": {
                "should": should_clauses,
                "minimum_should_match": 1
            }
        }


def search_with_doc_meta_filter(index_name, filtering_conditions):
    """
    使用元数据过滤条件进行搜索

    参数:
    index_name: 索引名称
    filtering_conditions: 过滤条件

    返回:
    搜索结果
    """
    query_body = build_doc_meta_query(filtering_conditions)

    logger.info('search_with_doc_meta_filter, query body: ' + json.dumps(query_body, indent=4, ensure_ascii=False))

    # 添加搜索参数
    search_body = {
        "query": query_body,
    }

    # 初始化scroll
    response = es.search(
        index=index_name,
        body=search_body,
        scroll='2m'
    )

    scroll_id = response['_scroll_id']
    hits = response['hits']['hits']
    file_names = set()

    # 添加第一批结果
    for hit in hits:
        hit_data = hit['_source']
        file_names.add(hit_data["file_name"])

    # 继续scroll直到没有更多结果
    while len(hits) > 0:
        response = es.scroll(scroll_id=scroll_id, scroll='2m')
        hits = response['hits']['hits']
        scroll_id = response['_scroll_id']

        for hit in hits:
            hit_data = hit['_source']
            file_names.add(hit_data["file_name"])

    # 清理scroll
    if scroll_id:
        es.clear_scroll(scroll_id=scroll_id)

    return list(file_names)


def retype_meta_datas(meta_datas: list):
    result = []
    for item in meta_datas:
        if item["value_type"] == "string":
            result.append({
                "key": str(item["key"]),
                "string_value": str(item["value"]),
                "value_type": str(item["value_type"])
            })
        elif item["value_type"] == "number" or item["value_type"] == "time":
            result.append({
                "key": str(item["key"]),
                "int_value": int(item["value"]),
                "value_type": str(item["value_type"])
            })

    return result


def get_index_update_actions(index_name, kb_name, file_name, update_data, index_type:IndexType=IndexType.MAIN):
    """
   更新操作函数

    参数:
    index_name: 索引名称
    kb_name: 知识库名称
    file_name: 文件名
    update_data: 更新的数据, list
    index_type: 索引类型 ("main", "snippet", "content_control", "file_control")

    返回:
    actions: 更新操作列表
    """
    # 根据索引类型确定查询字段和ID字段
    if index_type == IndexType.SNIPPET:
        # title 未在 snippet_mappings 中定义，属动态映射出的分词 text 字段；
        # term 精确匹配必须使用 keyword 子字段（与 es_util 中 title.keyword 既有用法一致），
        # 否则该过滤条件永远 0 命中，meta 更新会对 snippet 索引静默失效
        file_field = "title.keyword"
        id_field = "chunk_id"
    elif index_type == IndexType.CONTENT_CONTROL:
        file_field = "file_name"
        id_field = "content_id"
    elif index_type == IndexType.FILE_CONTROL:
        file_field = "file_name"
        id_field = "file_id"
    else:
        file_field = "file_name"
        id_field = "chunk_id"

    must_conditions = [
        {"term": {"kb_name": kb_name}},
        {"term": {file_field: file_name}}
    ]

    query = {
        "query": {
            "bool": {
                "must": must_conditions
            }
        }
    }

    scan_kwargs = {
        "index": index_name,
        "query": query,
        "scroll": "1m",
        "size": 100
    }

    upsert_data = []
    for doc in helpers.scan(es, **scan_kwargs):
        # 处理元数据
        data = {
            id_field: doc["_source"][id_field],
            "meta_data": doc["_source"].get("meta_data", {})
        }

        nested_doc_meta = retype_meta_datas(update_data)
        data["meta_data"]["doc_meta"] = nested_doc_meta

        upsert_data.append(data)

    actions = []
    for item in upsert_data:
        doc_id = item[id_field]
        action = {
            "_op_type": "update",
            "_index": index_name,
            "_id": doc_id,
            "doc": item
        }
        actions.append(action)

    return actions


def get_index_delete_meta_actions(index_name, kb_name, keys, index_type:IndexType=IndexType.MAIN):
    """
    删除包含指定keys的文档元数据

    参数:
    index_name: 索引名称
    kb_name: 知识库名称
    keys: 需要删除的key列表
    index_type: 索引类型 ("main", "snippet", "content_control", "file_control")

    返回:
    actions: 更新操作列表
    """

    # 根据索引类型确定ID字段
    if index_type == IndexType.SNIPPET:
        id_field = "chunk_id"
    elif index_type == IndexType.CONTENT_CONTROL:
        id_field = "content_id"
    elif index_type == IndexType.FILE_CONTROL:
        id_field = "file_id"
    else:
        id_field = "chunk_id"

    # 构建查询条件 - 只返回包含指定keys的文档
    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"kb_name": kb_name}}
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

    upsert_data = []
    for doc in helpers.scan(es, **scan_kwargs):
        # 获取文档当前的元数据
        current_meta_data = doc["_source"].get("meta_data", {})
        current_doc_meta = current_meta_data.get("doc_meta", [])

        # 过滤掉在keys列表中的元数据key
        filtered_doc_meta = [item for item in current_doc_meta if item.get("key") not in keys]

        if len(filtered_doc_meta) != len(current_doc_meta):
            data = {
                id_field: doc["_source"][id_field],
                "meta_data": current_meta_data
            }
            data["meta_data"]["doc_meta"] = filtered_doc_meta
            upsert_data.append(data)

    actions = []
    for item in upsert_data:
        doc_id = item[id_field]
        action = {
            "_op_type": "update",
            "_index": index_name,
            "_id": doc_id,
            "doc": item
        }
        actions.append(action)

    return actions


def get_index_rename_meta_actions(index_name, kb_name, key_mappings, index_type: IndexType = IndexType.MAIN):
    """
    重命名文档元数据中的key

    参数:
    index_name: 索引名称
    kb_name: 知识库名称
    key_mappings: key映射列表，每个元素包含 {"old_key": "key1", "new_key": "key2"}
    index_type: 索引类型枚举

    返回:
    actions: 更新操作列表
    """

    # 根据索引类型确定ID字段
    if index_type == IndexType.SNIPPET:
        id_field = "chunk_id"
    elif index_type == IndexType.CONTENT_CONTROL:
        id_field = "content_id"
    elif index_type == IndexType.FILE_CONTROL:
        id_field = "file_id"
    else:
        id_field = "chunk_id"

    # 构建查询条件 - 只返回包含需要重命名的键的文档
    old_keys = [mapping["old_key"] for mapping in key_mappings]

    query = {
        "query": {
            "bool": {
                "must": [
                    {"term": {"kb_name": kb_name}}
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

    upsert_data = []
    for doc in helpers.scan(es, **scan_kwargs):
        # 获取文档当前的元数据
        current_meta_data = doc["_source"].get("meta_data", {})
        current_doc_meta = current_meta_data.get("doc_meta", [])

        # 创建key映射快速查找
        key_map = {mapping["old_key"]: mapping["new_key"] for mapping in key_mappings}

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
                id_field: doc["_source"][id_field],
                "meta_data": current_meta_data
            }
            data["meta_data"]["doc_meta"] = renamed_doc_meta
            upsert_data.append(data)

    actions = []
    for item in upsert_data:
        doc_id = item[id_field]
        action = {
            "_op_type": "update",
            "_index": index_name,
            "_id": doc_id,
            "doc": item
        }
        actions.append(action)

    return actions

def update_index_meta_data(index_actions: dict):
    """更新索引元数据"""
    return update_index_data(index_actions, update_doc_meta_mapping)

def update_file_metas(user_id:str, kb_name: str, update_datas: dict):
    operation = update_datas.get("operation")
    main_index_name = get_main_index_name(user_id)
    snippet_index_name = get_snippet_index_name(user_id)
    content_index_name = get_content_control_index_name(user_id)
    file_index_name = get_file_index_name(user_id)

    index_type_mapping = {
        main_index_name: IndexType.MAIN,
        snippet_index_name: IndexType.SNIPPET,
        content_index_name: IndexType.CONTENT_CONTROL,
        file_index_name: IndexType.FILE_CONTROL
    }

    # init
    index_actions = {}
    for index_name in index_type_mapping:
        index_actions[index_name] = []

    if operation == "update_metas":
        meta_datas = update_datas.get("metas", [])
        logger.info(f"请求更新元数据 meta_datas: {meta_datas}")
        for item in meta_datas:
            file_name = item["file_name"]
            metadata_list = item["metadata_list"]
            for index_name, index_type in index_type_mapping.items():
                index_actions[index_name].extend(
                    get_index_update_actions(index_name, kb_name, file_name, metadata_list, index_type))
        return update_index_meta_data(index_actions)
    elif operation == "delete_keys":
        keys = update_datas.get("keys", [])
        logger.info(f"请求删除元数据 keys: {keys}")
        for index_name, index_type in index_type_mapping.items():
            index_actions[index_name].extend(
                get_index_delete_meta_actions(index_name, kb_name, keys, index_type))
        return update_index_meta_data(index_actions)
    elif operation == "rename_keys":
        key_mappings = update_datas.get("key_mappings", [])
        logger.info(f"请求重命名元数据 key_mappings: {key_mappings}")
        for index_name, index_type in index_type_mapping.items():
            index_actions[index_name].extend(
                get_index_rename_meta_actions(index_name, kb_name, key_mappings, index_type))
        return update_index_meta_data(index_actions)
    else:
        logger.warning(f"更新元数据不支持的操作类型: {operation}")
        return {
            "code": 1,
            "message": f"更新元数据不支持的操作类型: {operation}"
        }


# 使用示例
if __name__ == "__main__":
    # 示例1: 单个条件组，AND关系
    filtering_conditions_1 = [
        {
            "filtering_kb_name": "gx_test",
            "logical_operator": "and",
            "conditions": [
                {
                    "meta_name": "int_key",
                    "meta_type": "Number",
                    "comparison_operator": "is",
                    "value": "4444"
                },
                {
                    "meta_name": "str_key",
                    "meta_type": "String",
                    "comparison_operator": "empty"
                }
            ]
        }
    ]

    # 示例2: 单个条件组，OR关系
    filtering_conditions_2 = [
        {
            "filtering_kb_name": "gx_test",
            "logical_operator": "or",
            "conditions": [
                {
                    "meta_name": "int_key",
                    "meta_type": "Number",
                    "comparison_operator": ">",
                    "value": "1000"
                },
                {
                    "meta_name": "str_key",
                    "meta_type": "String",
                    "comparison_operator": "not empty"
                }
            ]
        }
    ]

    # 示例3: 多个条件组
    filtering_conditions_3 = [
        {
            "filtering_kb_name": "gx_test",
            "logical_operator": "and",
            "conditions": [
                {
                    "meta_name": "int_key",
                    "meta_type": "Number",
                    "comparison_operator": ">",
                    "value": "1000"
                }
            ]
        },
        {
            "filtering_kb_name": "gx_test2",
            "logical_operator": "or",
            "conditions": [
                {
                    "meta_name": "str_key",
                    "meta_type": "String",
                    "comparison_operator": "contains",
                    "value": "test"
                }
            ]
        }
    ]

    # query body
    query1 = build_doc_meta_query(filtering_conditions_1)
    print(query1)
    query2 = build_doc_meta_query(filtering_conditions_2)
    print(query2)
    query3 = build_doc_meta_query(filtering_conditions_3)
    print(query3)

    # 执行搜索
    # result1 = search_with_doc_meta_filter("your_index_name", filtering_conditions_1)
    # result2 = search_with_doc_meta_filter("your_index_name", filtering_conditions_2)
    # result3 = search_with_doc_meta_filter("your_index_name", filtering_conditions_3)
