import hashlib
import re
import json
from itertools import product

def generate_md5(content_str):
    # 创建一个md5 hash对象
    md5_obj = hashlib.md5()

    # 对字符串进行编码，因为md5需要bytes类型的数据
    md5_obj.update(content_str.encode('utf-8'))

    # 获取十六进制的MD5值
    md5_value = md5_obj.hexdigest()

    return md5_value

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
        # 获取整个哈希表，返回一个字典，字段是 id，值是对应的条目 JSON 字符串
        term_dict_hash = redis_client.hgetall(redis_key)
        if term_dict_hash:
            # 将每个字段的 JSON 字符串转换为 Python 对象（字典）
            term_dict = [json.loads(value) for value in term_dict_hash.values()]
            all_query_dicts.extend(term_dict)
    # 此处请将all_query_dicts相同元素去重
    # 去重：将所有字典转换为 JSON 字符串，存入集合中，集合自动去重
    unique_query_dicts = {json.dumps(query_dict, sort_keys=True): query_dict for query_dict in all_query_dicts}
    # 返回去重后的字典列表
    return list(unique_query_dicts.values())

def query_rewrite(question, term_dict):
    """
    根据专名同义词表改写用户问题，支持生成多个改写结果（针对多个别名）。

    参数:
    - question (str): 用户输入问题。
    - term_dict (list): 专名同义词表，每项为字典，包含 'name' 和 'alias'。

    返回:
    - list: 改写后的用户问题列表，每个改写对应一种组合方式。
    """
    # 保存所有的替换项
    replacements = []

    for term in term_dict:
        name = term["name"]  # 标准词
        aliases = term["alias"]  # 别名列表

        # 如果问题中包含标准词，则保存替换方案
        if re.search(re.escape(name), question):
            replacements.append([(name, alias) for alias in aliases])

    # 如果没有匹配到标准词，直接返回原问题
    if not replacements:
        return [question]

    # 使用笛卡尔积计算所有可能的替换组合
    combinations = product(*replacements)

    rewritten_questions = []
    for combo in combinations:
        # 逐个应用替换规则
        new_question = question
        for name, alias in combo:
            new_question = re.sub(re.escape(name), alias, new_question)
        rewritten_questions.append(new_question)

    return rewritten_questions

