import redis
import json
import os
import hashlib
from logging_config import setup_logging
logger_name='rag_redis_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))
from settings import REDIS_ADDRESS, REDIS_PORT, REDIS_PASSWD, REDIS_DB



def get_redis_connection(redis_db=REDIS_DB):
    """
    获取 Redis 连接
    :return: Redis 客户端实例
    """
    try:
        # 创建连接池 [EN] Create connection pool
        # redis_pool = redis.ConnectionPool(
        #     host=redis_host,
        #     port=redis_port,
        #     password=redis_passwd,
        #     decode_responses=True,
        #     db=redis_db
        # )
        # 创建 Redis 客户端 [EN] Create a Redis client
        pool = redis.ConnectionPool(host=REDIS_ADDRESS, port=REDIS_PORT, password=REDIS_PASSWD,
                                    decode_responses=True, db=int(redis_db))
        # 测试连接 [EN] test connection
        # redis_client.ping()
        r = redis.Redis(connection_pool=pool)
        logger.info("Connected to Redis successfully!")
        return r
    except Exception as e:
        import traceback
        logger.error("====> conn redis error %s" % e)
        logger.error(traceback.format_exc())

def set_cache(redis_client, key, value):
    """
    设置缓存
    :param redis_client: Redis 客户端实例
    :param key: 缓存键
    :param value: 缓存值
    """
    try:
        # 将数据转换为 JSON 字符串 [EN] Convert data to JSON string
        value_json = json.dumps(value,ensure_ascii=False)
        # 将数据存入 Redis [EN] Save data to Redis
        redis_client.set(key, value_json)
        logger.info(f"Data stored in Redis with key: {key}")
    except Exception as e:
        logger.error(f"Failed to store data in Redis: {e}")
        # raise

def get_cache(redis_client, key):
    """
    根据键查询缓存并打印结果
    :param redis_client: Redis 客户端实例
    :param key: 缓存键
    :return: 缓存值
    """
    try:
        
        # 查询缓存 [EN] Query cache
        cached_data = redis_client.get(key)
        if cached_data:
            logger.info(f"Data retrieved from Redis with key: {key}")
            logger.info(f"Value: {cached_data}")
            # 将 JSON 字符串转换回字典 [EN] Convert JSON string back to dictionary
            value = json.loads(cached_data)
            logger.info(f"Search List: {value}")
            return value
        else:
            logger.info(f"No data found for key: {key}")
            return None
    except Exception as e:
        logger.error(f"Failed to retrieve data from Redis: {e}")
        # raise


def delete_cache_by_prefix(redis_client, prefix):
    """
        按前缀删除 Redis 缓存
        :param redis_client: Redis 客户端实例
        :param prefix: 缓存键的前缀
        """
    try:
        # 使用 scan 命令按前缀模糊查询键 [EN] Use the scan command to fuzzy query keys by prefix
        keys_to_delete = []
        cursor = '0'
        while cursor != 0:
            cursor, keys = redis_client.scan(cursor=cursor, match=f"{prefix}*")
            keys_to_delete.extend(keys)

        if keys_to_delete:
            # 删除找到的所有键 [EN] Delete all keys found
            redis_client.delete(*keys_to_delete)
            logger.info(f"Deleted {len(keys_to_delete)} keys with prefix: {prefix}")
            response_info = {'code': 0, "message": "redis缓存删除成功"}
        else:
            logger.info(f"No keys found with prefix: {prefix}")
            response_info = {'code': 0, "message": "当前知识库无redis缓存信息"}
    except Exception as e:
        response_info = {'code': 1, "message": "redis缓存删除失败"}
        logger.error(f"Failed to delete cache by prefix: {e}")

    return response_info

def delete_cache_by_key(redis_client, key):
    """
    根据键删除 Redis 中的缓存
    :param key: 要删除的键
    """
    try:
        # 删除单个键 [EN] Delete a single key
        result = redis_client.delete(key)
        if result == 1:
            logger.info(f"Key '{key}' deleted successfully.")
            print(f"Key '{key}' deleted successfully.")
        else:
            logger.info(f"Key '{key}' does not exist.")
            print(f"Key '{key}' does not exist.")
    except Exception as e:
        logger.error(f"Error deleting key '{key}': {e}")
        print(f"Error deleting key '{key}': {e}")


# 方法：根据 user_id 和条目 id 增加单条数据 [EN] Method: Add a single piece of data based on user_id and entry id
def add_query_dict_entry(redis_client, user_id, term_entry, knowledgebase):
    """
    向 Redis 中添加单条 term_dict 数据。
    :param user_id: 用户ID
    :param term_entry: 要添加的单条 term_dict 数据（字典）
    """

    for kb in knowledgebase:
        redis_key = f"query_dict:{user_id}:{kb}"
        field = str(term_entry['id'])  # 使用 id 作为字段 [EN] Use id as field
        value = json.dumps(term_entry)  # 将条目转为 JSON 字符串 [EN] Convert entries to JSON string
        # 添加到 Redis 哈希表 [EN] Add to Redis hash table
        redis_client.hset(redis_key, field, value)
        print(f"已添加条目：{term_entry}")
        logger.info(f"已添加条目：{term_entry}")


# 方法：根据 user_id 和条目 id 删除单条数据 [EN] Method: Delete a single piece of data based on user_id and entry id
def delete_query_dict_entry(redis_client, user_id, term_entry_id, knowledgebase):
    """
    删除 Redis 中指定 user_id 和条目 id 的 term_dict 数据。
    :param user_id: 用户ID
    :param term_entry_id: 要删除的条目 id
    """

    for kb in knowledgebase:
        redis_key = f"query_dict:{user_id}:{kb}"
        field = str(term_entry_id)  # 使用 id 作为字段 [EN] Use id as field

        # 从 Redis 哈希表中删除指定字段 [EN] Delete the specified field from the Redis hash table
        redis_client.hdel(redis_key, field)
        print(f"已删除条目 id: {term_entry_id}")
        logger.info(f"已删除条目 id: {term_entry_id}")


# 方法：根据 user_id 和条目 id 修改单条数据 [EN] Method: Modify a single piece of data based on user_id and entry id
def update_query_dict_entry(redis_client, user_id, term_entry_id, new_term_entry, knowledgebase):
    """
    修改 Redis 中指定 user_id 和条目 id 的 term_dict 数据。
    :param user_id: 用户ID
    :param term_entry_id: 要修改的条目 id
    :param new_term_entry: 更新后的 term_dict 数据（字典）
    """

    for kb in knowledgebase:
        redis_key = f"query_dict:{user_id}:{kb}"
        field = str(term_entry_id)  # 使用 id 作为字段 [EN] Use id as field
        value = json.dumps(new_term_entry)  # 将条目转为 JSON 字符串 [EN] Convert entries to JSON string

        # 更新 Redis 哈希表中的字段 [EN] Update fields in Redis hash table
        redis_client.hset(redis_key, field, value)
        print(f"已修改条目 id: {term_entry_id} 为：{new_term_entry}")
        logger.info(f"已修改条目 id: {term_entry_id} 为：{new_term_entry}")

def get_query_dict_cache(redis_client, user_id, knowledgebases):
    """
    根据 user_id,查询的知识库knowledgebase列表 查询 Redis 中的缓存，将哈希表字段的值解析为 query_dict。
    :param user_id: 用户ID
    :return: 完整的 query_dict 数据（列表形式），如果缓存不存在则返回 None。
    """
    all_query_dicts = []

    redis_key_list = []
    for knowledgebase in knowledgebases:
        redis_key = f"query_dict:{user_id}:{knowledgebase}"
        redis_key_list.append(redis_key)
    for redis_key in redis_key_list:
        # 获取整个哈希表，返回一个字典，字段是 id，值是对应的条目 JSON 字符串 [EN] Get the entire hash table and return a dictionary. The field is id and the value is the corresponding entry JSON string.
        term_dict_hash = redis_client.hgetall(redis_key)
        if term_dict_hash:
            # 将每个字段的 JSON 字符串转换为 Python 对象（字典） [EN] Convert each field's JSON string to a Python object (dictionary)
            term_dict = [json.loads(value) for value in term_dict_hash.values()]
            all_query_dicts.extend(term_dict)
    # 此处请将all_query_dicts相同元素去重 [EN] Please deduplicate the same elements in all_query_dicts here
    # 去重：将所有字典转换为 JSON 字符串，存入集合中，集合自动去重 [EN] Deduplication: Convert all dictionaries into JSON strings and store them in the collection, and the collection will automatically deduplicate
    unique_query_dicts = {json.dumps(query_dict, sort_keys=True): query_dict for query_dict in all_query_dicts}
    # 返回去重后的字典列表 [EN] Return the dictionary list after deduplication
    return list(unique_query_dicts.values())


def update_chunk_labels(redis_client, kb_id, file_name, chunk_id, labels):
    """
    更新指定知识库中某个chunk的标签
    :param redis_client: Redis连接
    :param kb_id: 知识库ID
    :param file_name: 文件名
    :param chunk_id: chunk的ID
    :param labels: 标签列表，类型为list
    """
    try:
        # 获取Redis连接 [EN] Get Redis connection
        # redis_client = get_redis_connection(redis_db=5)  # 使用固定DB 5 [EN] redis_client = get_redis_connection(redis_db=5) # Use fixed DB 5
        # 构造key [EN] Construct key
        hash_file_name = hashlib.md5(file_name.encode('utf-8')).hexdigest()  # 规避特殊字符 [EN] Avoid special characters
        key = f"{kb_id}{hash_file_name}{chunk_id}"
        # 将标签列表转换为JSON字符串存储 [EN] Convert list of tags to JSON string storage
        value = json.dumps({"labels": labels})
        # 更新或新增记录 [EN] Update or add records
        redis_client.set(key, value)
        logger.info(f"Updated chunk labels successfully: {key}")
    except Exception as e:
        logger.error(f"Failed to update chunk labels: {e}")
        import traceback
        logger.error(traceback.format_exc())


def delete_chunk_labels(redis_client, kb_id, file_name=""):
    """
        按前缀删除 Redis 缓存
        :param redis_client: Redis 客户端实例
        :param kb_id: 知识库ID
        :param file_name: 文件名，如果指定了文件名，则删除该文件名对应的缓存
    """
    try:
        # # 获取Redis连接 [EN] # Get Redis connection
        # redis_client = get_redis_connection(redis_db=5)  # 使用固定DB 5 [EN] redis_client = get_redis_connection(redis_db=5) # Use fixed DB 5
        # 使用scan命令查找所有匹配的key [EN] Use the scan command to find all matching keys
        if file_name:  # 如果指定了文件名，则使用文件名生成前缀 [EN] If a filename is specified, the filename is used to generate the prefix
            hash_file_name = hashlib.md5(file_name.encode('utf-8')).hexdigest()  # 规避特殊字符 [EN] Avoid special characters
            prefix = f"{kb_id}{hash_file_name}"
        else:
            prefix = f"{kb_id}"
        cursor = "0"
        while cursor != 0:
            cursor, keys = redis_client.scan(cursor=cursor, match=f"{prefix}*")
            if keys:
                # 删除找到的所有key [EN] Delete all keys found
                redis_client.delete(*keys)
        logger.info(f"Deleted prefix chunk labels successfully: {prefix}")
    except Exception as e:
        logger.error(f"Failed to delete prefix chunk labels: {e}")
        import traceback
        logger.error(traceback.format_exc())


def get_all_chunk_labels(redis_client, kb_id):
    """
    根据 kb_id 查询所有 chunk 的标签，并去重返回一个完整的标签列表
    :param redis_client: Redis连接
    :param kb_id: 知识库ID
    :return: 去重后的标签列表
    """
    try:
        # # 获取Redis连接 [EN] # Get Redis connection
        # redis_client = get_redis_connection(redis_db=5)  # 使用固定DB 5 [EN] redis_client = get_redis_connection(redis_db=5) # Use fixed DB 5
        # 初始化一个集合用于去重 [EN] Initialize a collection for deduplication
        unique_labels = set()
        # 使用scan命令查找所有匹配的key [EN] Use the scan command to find all matching keys
        cursor = "0"
        while cursor != 0:
            cursor, keys = redis_client.scan(cursor=cursor, match=f"{kb_id}*")
            for key in keys:
                # 获取key对应的value [EN] Get the value corresponding to the key
                value = redis_client.get(key)
                if value:
                    # 解析JSON格式的value [EN] Parse value in JSON format
                    chunk_data = json.loads(value)
                    # 将labels添加到集合中去重 [EN] Add labels to the collection to remove duplicates
                    unique_labels.update(chunk_data.get("labels", []))
        # 将集合转换为列表并返回 [EN] Convert a collection to a list and back
        return list(unique_labels)
    except Exception as e:
        logger.error(f"Failed to get {kb_id} all chunk labels: {e}")
        import traceback
        logger.error(traceback.format_exc())
        return []


def delete_graph_vocabulary_set(redis_client, kb_id):
    """
    如果键不存在则跳过，存在则删除 Redis 中的 graph_vocabulary 集合
    :param redis_client: Redis连接
    :param kb_id: 知识库ID
    """
    try:
        # 构造键名 [EN] Construct key name
        graph_key = f"graph_vocabulary_{kb_id}"
        # 如果键不存在，直接跳过 [EN] If the key does not exist, skip directly
        if not redis_client.exists(graph_key):
            pass
        else:
            # 如果键存在，删除旧的集合 [EN] If the key exists, delete the old collection
            redis_client.delete(graph_key)
    except Exception as e:
        logger.error(f"Failed to delete {kb_id} graph vocabulary set: {e}")
        import traceback
        logger.error(traceback.format_exc())


def update_graph_vocabulary_set(redis_client, kb_id, elements_to_add=None, elements_to_remove=None):
    """
    更新 Redis 中的 graph_vocabulary 集合
    :param redis_client: Redis连接
    :param kb_id: 知识库ID
    :param elements_to_add: 要添加的元素列表
    :param elements_to_remove: 要删除的元素列表
    """
    try:
        # 构造键名 [EN] Construct key name
        graph_key = f"graph_vocabulary_{kb_id}"
        if elements_to_add:
            redis_client.sadd(graph_key, *elements_to_add)  # 添加元素 [EN] Add element
        if elements_to_remove:
            redis_client.srem(graph_key, *elements_to_remove)  # 删除元素 [EN] Delete element
    except Exception as e:
        logger.error(f"Failed to update {kb_id} graph vocabulary set: {e}")
        import traceback
        logger.error(traceback.format_exc())


# 查询集合 [EN] Query collection
def query_graph_vocabulary_set(redis_client, kb_id):
    """
    查询 Redis 中的集合
    :param redis_client: Redis连接
    :param kb_id: 知识库ID
    :return: 集合中的所有元素
    """
    try:
        # 构造键名 [EN] Construct key name
        graph_key = f"graph_vocabulary_{kb_id}"
        if redis_client.exists(graph_key):
            res_set = redis_client.smembers(graph_key)
            return res_set
        else:
            return set()
    except Exception as e:
        logger.error(f"Failed to query {kb_id} graph vocabulary set: {e}")
        import traceback
        logger.error(traceback.format_exc())


# 主程序 [EN] main program
if __name__ == "__main__":
    # 获取 Redis 连接 [EN] Get Redis connection
    redis_client = get_redis_connection()

    # 要存储的数据 [EN] data to store
    question = "example_question"
    search_list = [{"id": "1", "value": 2}]  # 修正了重复的键名 [EN] Fixed duplicate key names

    # 设置缓存 [EN] Set up cache
    set_cache(redis_client, question, search_list)

    # 查询缓存 [EN] Query cache
    get_cache(redis_client, question)

    # prefix = "13782ff0-5a73-4a51-8f7d-77015ee79aad^商飞-空调系统^" [EN] prefix = "13782ff0-5a73-4a51-8f7d-77015ee79aad^COMAC-Air Conditioning System^"
    # result = delete_cache_by_prefix(redis_client, prefix)
    # print(json.dumps(result, ensure_ascii=False))