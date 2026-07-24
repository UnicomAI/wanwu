from pathlib import Path
from typing import Optional
from urllib.parse import urlparse, urlunparse
import subprocess
import os
import shutil
import argparse
import logging
import datetime
import sys
import requests
import json
import time
import re
import uuid
import copy
import traceback

from easyofd.ofd import OFD
from ofdparser import OfdParser
import base64
from datetime import datetime, timedelta
from enum import Enum

# 验证设置是否成功
from utils import milvus_utils
from utils import es_utils
from utils import file_utils
from utils import rerank_utils
from utils import minio_utils
from utils import redis_utils
from utils import graph_utils
from utils import timing
import time

from settings import REPLACE_MINIO_DOWNLOAD_URL
from settings import USE_POST_FILTER
from settings import GRAPH_SERVER_URL
from utils.constant import USER_DATA_PATH
from model_manager.model_config import get_model_configure

logger = logging.getLogger(__name__)

user_data_path = Path(USER_DATA_PATH)
chunk_label_redis_client = redis_utils.get_redis_connection(redis_db=5)


def is_safe_filename(name: str) -> bool:
    if "/" in name or "\\" in name:
        return False
    if ".." in name:
        return False
    return True


# -----------------
# 初始化知识库
def init_knowledge_base(user_id: str,
                        kb_info: dict,
                        embedding_model_id: str = "",
                        enable_knowledge_graph: bool = False,
                        is_multimodal: bool = False) -> dict:
    """
    初始化知识库

    :param user_id: 用户ID
    :param kb_info: {"kb_name": "...", "kb_id": "..."}
    :param embedding_model_id: 嵌入模型ID (可选)
    :param enable_knowledge_graph: 是否启用知识图谱 (默认 False)
    :param is_multimodal: 是否多模态知识库 (默认 False)
    :return: 操作结果字典，包含 'code' 和 'message'
    """
    kb_name = kb_info["kb_name"]
    response_info = {'code': 0, "message": "成功"}
    try:
        # ----------------0、参数校验
        if is_multimodal and not get_model_configure(embedding_model_id).is_multimodal:
                raise ValueError("multimodal model is needed for initializing multimodal knowledge base")
        # ----------------1、检测向量库名称是否合法
        if not is_valid_string(user_id + kb_name):
            raise ValueError(f'知识库名称仅能包括大小写英文、数字、中文和_符号, input: {kb_name}')
        # ----------------2、check 向量库 是否有重复的
        milvus_data = list_knowledge_base(user_id)
        logger.info(f'向量库已有知识库查询结果：{milvus_data}')
        if milvus_data['code'] != 0:
            raise RuntimeError(f'向量库校验失败, details: {milvus_data["message"]}')
        if kb_name in milvus_data['data']['knowledge_base_names']:
            raise ValueError('已存在相同名字的向量知识库')
        # ----------------2、建立向量库
        milvus_init_result = milvus_utils.init_knowledge_base(user_id, kb_info,
                                                              embedding_model_id = embedding_model_id,
                                                              enable_knowledge_graph = enable_knowledge_graph,
                                                              is_multimodal=is_multimodal)
        logger.info(f'向量库初始化结果：{milvus_init_result}')
        if milvus_init_result['code'] != 0:
            raise RuntimeError(milvus_init_result['message'])
        # ----------------3、建立路径
        if not os.path.exists(os.path.join(user_data_path, user_id)):
            os.mkdir(os.path.join(user_data_path, user_id))
        if os.path.exists(os.path.join(user_data_path, user_id, kb_name)):
            shutil.rmtree(os.path.join(user_data_path, user_id, kb_name))
        if not os.path.exists(os.path.join(user_data_path, user_id, kb_name)):
            os.mkdir(os.path.join(user_data_path, user_id, kb_name))
        return response_info
    except ValueError as e:
        logger.warning(f'参数错误: {e}')
        response_info["code"] = 1
        response_info["message"] = repr(e)
        return response_info
    except Exception as e:
        logger.error(f'初始化运行错误: {e}')
        response_info["code"] = 1
        response_info["message"] = repr(e)
        return response_info


# -----------------
# 查询所有知识库
def list_knowledge_base(user_id):
    milvus_list_kb_result = milvus_utils.list_knowledge_base(user_id)
    logger.info('用户知识库查询结果：' + repr(milvus_list_kb_result))
    return milvus_list_kb_result


# -----------------
# 查询所有文档
def list_knowledge_file(user_id, kb_info):
    milvus_list_file_result = milvus_utils.list_knowledge_file(user_id, kb_info)
    logger.info('用户知识库文档查询结果：' + repr(milvus_list_file_result))
    return milvus_list_file_result


def list_knowledge_file_download_link(user_id, kb_info):
    """ 获取知识库里所有文档的下载链接 """
    milvus_list_file_result = milvus_utils.list_knowledge_file_download_link(user_id, kb_info)
    logger.info('获取知识库里所有文档的下载链接结果：' + repr(milvus_list_file_result))
    if milvus_list_file_result['code'] == 0:  # 替换好 minio下载链接
        file_download_links = []
        for url in milvus_list_file_result['data']['file_download_links']:
            # 正则表达式匹配 https://ip:port/minio/download/api/ 部分
            pattern = r'http?://[^/]+/minio/download/api/'
            # 替换文本中的URL
            file_download_links.append(re.sub(pattern, REPLACE_MINIO_DOWNLOAD_URL, url))
        milvus_list_file_result['data']['file_download_links'] = file_download_links

    return milvus_list_file_result

# -----------------
# 校验知识库是否存在
def check_knowledge_base(user_id, kb_info):
    kb_name = kb_info["kb_name"]
    response_info = {'code': 0, "message": "成功", "data": {"kb_exists": True}}
    milvus_list_kb_result = milvus_utils.list_knowledge_base(user_id)
    logger.info('用户知识库查询结果：' + repr(milvus_list_kb_result))
    if milvus_list_kb_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = milvus_list_kb_result['message']
        response_info['data']['kb_exists'] = False
        return response_info
    else:
        kb_list = milvus_list_kb_result['data']['knowledge_base_names']
        if len(kb_list) > 0 and kb_name in kb_list:
            return response_info
        else:
            response_info['data']['kb_exists'] = False
            return response_info


# -----------------删除知识库
def del_konwledge_base(user_id, kb_info):
    kb_name = kb_info["kb_name"]
    kb_id = kb_info["kb_id"]
    kb_path = os.path.join(user_data_path, user_id, kb_name)
    response_info = {'code': 0, "message": "成功"}
    # ====== check 知识库是否存在 ===
    milvus_data = list_knowledge_base(user_id)
    if kb_name not in milvus_data['data']['knowledge_base_names']:
        response_info['code'] = 1
        response_info['message'] = f'{kb_name},知识库不存在'
        return response_info

     #删除 知识图谱
    kb_meta = milvus_utils.get_kb_info(user_id, kb_name)
    if "enable_knowledge_graph" in kb_meta and kb_meta["enable_knowledge_graph"]:
        try:
            graph_utils.delete_kb_graph(user_id, kb_name)
            logger.info(f"知识图谱删除成功, kb_name:{kb_name}")
            graph_redis_client = redis_utils.get_redis_connection()
            redis_utils.delete_graph_vocabulary_set(graph_redis_client, kb_id)
        except Exception as e:
            logger.error(f"知识图谱删除失败, error: {repr(e)}")

    # --------------1、删除es库 (必须先删除es库，否则会报错)
    del_es_result = es_utils.del_es_kb(user_id, kb_info)
    logger.info('用户es库删除结果：' + repr(del_es_result))
    if del_es_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = del_es_result['message']
        if '不存在' in del_es_result['message']:
            if os.path.exists(kb_path): shutil.rmtree(kb_path)
        return response_info

    # --------------2、删除向量库
    del_milvus_result = milvus_utils.del_milvus_kbs(user_id, kb_info)
    logger.info('用户milvus库删除结果：' + repr(del_milvus_result))
    if del_milvus_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = del_milvus_result['message']
        if '不存在' in del_milvus_result['message']:
            if os.path.exists(kb_path): shutil.rmtree(kb_path)
        return response_info

    # --------------3、删除路径
    kb_path = os.path.join(user_data_path, user_id, kb_name)
    if os.path.exists(kb_path):
        shutil.rmtree(kb_path)
    return response_info


# -----------------删除多个文档
def del_knowledge_base_files(user_id, kb_info, file_names):
    kb_name = kb_info["kb_name"]
    filepath = os.path.join(user_data_path, user_id, kb_name)
    response_info = {'code': 0, "message": "成功"}
    # --------------1、check file_names
    if len(file_names) == 0:
        response_info['code'] = 1
        response_info['message'] = '未指定需要删除的文档'
        return response_info
    if all(not s for s in file_names):
        response_info['code'] = 1
        response_info['message'] = '未指定需要删除的文档'
        return response_info

    # --------------2、删除向量库、es库中文档
    success_files = []
    failed_files = []
    for file_name in file_names:
        # 删除milvus
        del_milvus_result = milvus_utils.del_milvus_files(user_id, kb_info, [file_name])
        logger.info('向量库文档删除结果：' + repr(del_milvus_result))

        if del_milvus_result['code'] != 0:
            failed_files.append([file_name, del_milvus_result['message']])
            continue
        else:
            success_files.append(file_name)
        # 删除es
        del_es_result = es_utils.del_es_file(user_id, kb_info, file_name)
        logger.info('es库文档删除结果：' + repr(del_es_result))

        if del_es_result['code'] != 0:
            failed_files.append([file_name, del_es_result['message']])
            continue
        else:
            success_files.append(file_name)

     #删除 知识图谱
    kb_meta = milvus_utils.get_kb_info(user_id, kb_name)
    if "enable_knowledge_graph" in kb_meta and kb_meta["enable_knowledge_graph"]:
        try:
            for file_name in success_files:
                graph_utils.delete_file_from_graph(user_id, kb_name, file_name)
                logger.info(f"知识图谱删除成功, file_name:{file_name}")
        except Exception as e:
            failed_files.append([file_name, f"知识图谱删除文件失败, error: {repr(e)}"])
            logger.error(f"知识图谱删除失败, file_name:{file_name}, error: {repr(e)}")

    # --------------2、路径文档
    for file_name in success_files:
        if is_safe_filename(file_name):
            del_file_path = os.path.join(filepath, file_name)
            if os.path.isfile(del_file_path): os.remove(del_file_path)
    for i in failed_files:
        if '文档不存在' in i[1]:
            if is_safe_filename(i[0]):
                del_file_path = os.path.join(filepath, i[0])
                if os.path.isfile(del_file_path): os.remove(del_file_path)

    if len(failed_files) == 0:
        return response_info
    else:
        m2 = ''
        if len(failed_files) > 0:
            m2 = '。'.join([i[0] + '删除失败，' + i[1] for i in failed_files])
        response_info['code'] = 1
        response_info['message'] = m2
        return response_info


def get_file_content_list(user_id: str, kb_info: dict, file_name: str, page_size: int, search_after: int,
                          query_text: str = None):
    """
    获取知识库文件片段列表,用于分页展示
    """
    logger.info(f"get_file_content_list start: {user_id}, kb_info: {kb_info}, file_name: {file_name}, "
                f"page_size:{page_size}, search_after:{search_after}, query_text:{query_text}")
    response_info = milvus_utils.get_milvus_file_content_list(user_id, kb_info, file_name, page_size,
                                                              search_after, query_text=query_text)
    logger.info(f"get_file_content_list end: {user_id}, kb_info: {kb_info}, file_name: {file_name}, "
                f"page_size:{page_size}, search_after:{search_after}, query_text:{query_text}, response: {response_info}")
    return response_info

def get_file_child_content_list(user_id: str, kb_info: dict, file_name: str, chunk_id: int):
    """
    获取知识库文件子片段列表
    """
    logger.info(f"get_file_child_content_list start: {user_id}, kb_info: {kb_info}, "
                f"file_name: {file_name}, chunk_id:{chunk_id}")
    response_info = milvus_utils.get_milvus_file_child_content_list(user_id, kb_info, file_name, chunk_id)
    logger.info(f"get_file_child_content_list end: {user_id}, kb_info: {kb_info}"
                f"file_name: {file_name}, chunk_id:{chunk_id}, response: {response_info}")
    return response_info

class MetadataOperation(Enum):
    """
    元数据操作类型枚举
    """
    UPDATE_METAS = "update_metas"
    DELETE_KEYS = "delete_keys"
    RENAME_KEYS = "rename_keys"

def manage_kb_metadata(user_id: str, kb_info: dict, operation: MetadataOperation, data: dict):
    """
    知识库元数据操作
    """
    if not data:
        logger.warning("未提供操作数据")
        return {'code': 1, 'message': '未提供操作数据'}

    logger.info(f"metadata operation start, user_id: {user_id}, kb_info: {kb_info}, "
                f"operation: {operation.value}, data: {data}")

    if operation == MetadataOperation.UPDATE_METAS:
        if 'metas' not in data or not data['metas']:
            logger.warning("更新元数据操作未提供元数据")
            return {'code': 1, 'message': '未提供更新元数据'}
    elif operation == MetadataOperation.DELETE_KEYS:
        if 'keys' not in data or not data['keys']:
            logger.warning("删除操作未提供keys")
            return {'code': 1, 'message': '未提供要删除的keys'}
    elif operation == MetadataOperation.RENAME_KEYS:
        if 'key_mappings' not in data or not data['key_mappings']:
            logger.warning("重命名元数据未提供key mappings")
            return {'code': 1, 'message': '未提供key mappings'}
        else:
            for mapping in data['key_mappings']:
                if (not isinstance(mapping, dict)
                        or 'old_key' not in mapping
                        or 'new_key' not in mapping
                        or mapping["old_key"] == mapping['new_key']):
                    logger.warning(f"无效的key mapping: {mapping}")
                    return {'code': 1, 'message': f'无效的key mapping: {mapping}'}
    else:
        logger.warning(f"元数据不支持的操作类型: {operation.value}")
        return {'code': 1, 'message': f'不支持的操作类型: {operation.value}'}

    data["operation"] = operation.value
    response_info = milvus_utils.update_file_metas(user_id, kb_info, data)
    logger.info(f"metadata operation end, user_id: {user_id}, kb_info: {kb_info}, "
                f"operation: {operation.value}, data: {data}, response: {response_info}")

    return response_info


def update_content_status(user_id: str, kb_info: dict, file_name: str, content_id: str, status: bool,
                          on_off_switch=None):
    """
    根据content_id更新知识库文件片段状态
    """
    logger.info('========= update_content_status start：' + repr(user_id) + '，' + repr(kb_info) +
                '，' + repr(file_name) + '，' + repr(content_id) + '，' + repr(status) + '，' + repr(on_off_switch))
    response_info = milvus_utils.update_milvus_content_status(user_id, kb_info, file_name, content_id, status,
                                                              on_off_switch)
    logger.info('========= update_content_status end：' + repr(user_id) + '，' + repr(kb_info) +
                '，' + repr(file_name) + '，' + repr(content_id) + '，' + repr(status) + '，' + repr(on_off_switch) +
                ' ====== response:' + repr(
        response_info))
    return response_info


def update_kb_name(user_id: str, old_kb_name: str, new_kb_name: str):
    """
    更新知识库名接口
    """
    logger.info('========= update_kb_name start：' + repr(user_id) + '，' + repr(old_kb_name) + '，' + repr(new_kb_name))
    response_info = milvus_utils.update_milvus_kb_name(user_id, old_kb_name, new_kb_name)
    logger.info('========= update_kb_name end：' + repr(user_id) + '，' + repr(old_kb_name) + '，' +
                 repr(new_kb_name) + ' ====== response:' + repr(response_info))
    return response_info


def get_knowledge_based_answer(knowledge_base_info, question, rate, top_k, chunk_conent, chunk_size, return_meta=False,
                               prompt_template='', search_field='content', default_answer='根据已知信息，无法回答您的问题。',
                               auto_citation=False, retrieve_method="hybrid_search",
                               filter_file_name_list=[], rerank_model_id='', rerank_mod="rerank_model",
                               weights: Optional[dict] | None = None, metadata_filtering_conditions=[], use_graph=False,
                               enable_vision=False, attachment_files=[]):
    """ knowledge_base_info: {"user_id1": [{ "kb_id": "","kb_name": ""}, { "kb_id": "","kb_name": ""}]}"""
    response_info = {'code': 0, "message": "成功", "data": {"prompt": "", "searchList": [], "score": []}}
    try:
        if search_field == 'emc':
            search_field = 'embedding_content'
        else:
            search_field = 'content'

        if top_k == 0:
            response_info['data']["prompt"] = question
            response_info['data']["searchList"] = []
            return response_info

        duplicate_set = set()
        vector_text_search_list = []
        label_useful_list = []  # 后过滤有效的知识片段
        graph_data_list = []  # SPO及社区报告置顶片段
        file_search_list = []
        for user_id, base_info_list in knowledge_base_info.items():
            temp_duplicate_set = set()
            user_search_list = []
            kb_names = [kb_info["kb_name"] for kb_info in base_info_list]
            kb_ids = [kb_info["kb_id"] for kb_info in base_info_list]
            kb_infos = [{"kb_name": n, "kb_id": i} for n, i in zip(kb_names, kb_ids)]
            if retrieve_method in {"semantic_search", "hybrid_search"}:
                search_result = milvus_utils.search_milvus(user_id, kb_infos, top_k, question, threshold=rate,
                                                           search_field=search_field,
                                                           filter_file_name_list=filter_file_name_list,
                                                           metadata_filtering_conditions = metadata_filtering_conditions,
                                                           enable_vision=enable_vision, attachment_files=attachment_files)

                logger.info(repr(user_id) + repr(kb_names) + repr(question) + '问题向量库查询结果：' + json.dumps(repr(search_result), ensure_ascii=False))

                if search_result['code'] != 0:
                    response_info['code'] = search_result['code']
                    response_info['message'] = search_result['message']
                    return response_info
                milvus_search_list = search_result['data']["search_list"]

                for item in milvus_search_list:
                    content = {
                        "title": item["file_name"],
                        "snippet": item["content"],
                        "kb_name": item["kb_name"],
                        "kb_id": item["kb_id"],
                        "content_id": item["content_id"],
                        "meta_data": item["meta_data"],
                        "user_id": user_id
                    }
                    check_repeated_text = item["content"]
                    if enable_vision and "content_type" in item and item["content_type"] == "image":
                        check_repeated_text = item["embedding_content"]
                        content["content_type"] = item["content_type"]
                        content["file_url"] = item["embedding_content"]

                    if check_repeated_text in temp_duplicate_set:
                        continue

                    if "is_parent" in item:
                        content["is_parent"] = item["is_parent"]
                    user_search_list.append(content)
                    temp_duplicate_set.add(check_repeated_text)

            if retrieve_method in {"full_text_search", "hybrid_search"} and question and len(str(question).strip()) > 0:
                # es召回
                es_search_list = es_utils.search_es(user_id, kb_infos, question, top_k,
                                                    filter_file_name_list=filter_file_name_list,
                                                    metadata_filtering_conditions=metadata_filtering_conditions)
                logger.info(repr(user_id) + repr(kb_names) + repr(question) + '问题es库查询结果：' + json.dumps(repr(es_search_list), ensure_ascii=False))
                for item in es_search_list:
                    if item["snippet"] in temp_duplicate_set: continue
                    item["user_id"] = user_id
                    user_search_list.append(item)
                    temp_duplicate_set.add(item["snippet"])

            # ========== 标签召回通道判断及调用==========
            unique_labels = set()   # 获取到所有的chunk标签
            for kb_id in kb_ids:
                unique_labels.update(redis_utils.get_all_chunk_labels(chunk_label_redis_client, kb_id))
            unique_labels_list = list(unique_labels)
            # 初始化一个字典来存储每个标签词的出现次数
            label_counts = {}
            # 遍历每个标签词，统计其在查询字符串中的出现次数
            for label in unique_labels_list:
                if label in question:
                    label_counts[label] = question.count(label)

            # 开始调用标签召回
            label_search_list = []
            if label_counts:
                label_search_list = es_utils.search_keyword(user_id, kb_infos, label_counts, top_k,
                                                            metadata_filtering_conditions=metadata_filtering_conditions)

            # 后过滤 status
            user_post_search_list = []
            if USE_POST_FILTER:
                logger.info(f"user_id: {user_id}, kb_names: {kb_names}, question: {question}, 后过滤start")
                # 向量召回和es召回做启停用后过滤,注意多个kb_names时，需要做区分
                content_status_json = {}
                search_lists = [user_search_list, label_search_list]
                for search_list in search_lists:
                    for i in search_list:
                        content_status_json[i["kb_name"]] = content_status_json.get(i["kb_name"], [])
                        if i['content_id'] not in content_status_json[i["kb_name"]]:
                            content_status_json[i["kb_name"]].append(i['content_id'])
                kb_name_to_kb_id = {n: i for n, i in zip(kb_names, kb_ids)}
                for kb_name in content_status_json:  # 多个kb_names时，需要做区分
                    cur_kb_info = {"kb_name": kb_name, "kb_id": kb_name_to_kb_id[kb_name]}
                    useful_content_id_list = milvus_utils.get_milvus_content_status(user_id, cur_kb_info, content_status_json[kb_name])
                    logger.info(
                        repr(user_id) + repr(kb_name) + repr(content_status_json[kb_name]) + '======== get_milvus_content_status：' + repr(
                            useful_content_id_list))
                    for item in user_search_list:
                        if item['kb_name'] == kb_name and item['content_id'] in useful_content_id_list:
                            user_post_search_list.append(item)
                    for c in label_search_list:
                        if c['kb_name'] == kb_name and c['content_id'] in useful_content_id_list:
                            label_useful_list.append(c)
                logger.info(f"question: {question}, user_id: {user_id}, user_post_search_list: {user_post_search_list}")
                logger.info(f"question: {question}, user_id: {user_id}, label_counts:{label_counts}, label_useful_list: {label_useful_list}")
            else:
                user_post_search_list = user_search_list
                label_useful_list.extend(label_search_list)

            #去重合并
            for item in user_post_search_list:
                if "content_type" in item and item["content_type"] == "image":
                    if item["file_url"] in duplicate_set: continue
                    duplicate_set.add(item["file_url"])
                else:
                    if item["snippet"] in duplicate_set: continue
                    duplicate_set.add(item["snippet"])
                vector_text_search_list.append(item)

            # ========= 图谱召回---增强关联片段以及三元组以及社区报告 start =========
            if use_graph:  # 如果使用图检索
                # ======== 将graph检索的结果 和 两路检索的结果进行融合，并重新再过一遍rerank ========
                temp_graph_search_list, temp_graph_dat_list = graph_utils.get_graph_search_list(user_id, kb_infos, question, top_k,
                                                                             threshold=rate,
                                                                             filter_file_name_list=filter_file_name_list)
                graph_data_list.extend(temp_graph_dat_list)  # 社区报告等直接放进去先
                # 根据 duplicate_set 去重，将图谱关联出来的chunk 再加入 vector_text_search_list
                for item in temp_graph_search_list:
                    item["user_id"] = user_id
                    if item["snippet"] in duplicate_set: continue
                    vector_text_search_list.append(item)
                    duplicate_set.add(item["snippet"])

        # 多路召回融合
        # reank重排
        if not vector_text_search_list:  # 都为空不走重排,直接返回
            response_info = {'code': 0, "message": "成功", "data": {"prompt": question, "searchList": [], "score": []}}
            logger.info('useful_list is None 重排结果：' + json.dumps(repr(response_info),ensure_ascii=False))
            return response_info


        if rerank_mod == "rerank_model":
            model_config = get_model_configure(rerank_model_id)
            is_support_multimodal = model_config.is_multimodal
            documents = []
            for item in vector_text_search_list:
                if is_support_multimodal:
                    if enable_vision and "content_type" in item and item["content_type"] == "image":
                        documents.append({item["content_type"]: item["file_url"]})
                    else:
                        documents.append({"text": item["snippet"]})
                else:
                    documents.append(item["snippet"])

            query = question
            if is_support_multimodal:
                query = {}
                if len(str(question).strip()) > 0:
                    query = {"text": question}
                if enable_vision and is_support_multimodal and attachment_files:
                    for item in attachment_files:
                        query.update(item)
                    if model_config.provider == "YuanJing":  #rernak供应商剔除 text
                        query.pop("text", None)

            rerank_result = rerank_utils.model_rerank(query,
                                                      top_k,
                                                      documents,
                                                      vector_text_search_list,
                                                      rerank_model_id,
                                                      model_config) # type: ignore
        elif rerank_mod == "weighted_score":
            if not question:
                rerank_result = vector_text_search_list
            else:
                rerank_result = rerank_utils.get_weighted_rerank(question, weights,
                                                                 vector_text_search_list, top_k)
        else:
            raise Exception("rerank_mod is not valid")
        if rerank_result["code"] != 0:
            logger.warn(f"rerank failed, rerank method: {rerank_mod}, rerank result: {rerank_result}")
            raise RuntimeError(rerank_result["message"])
        sorted_scores = rerank_result['data']["sorted_scores"]
        sorted_search_list = rerank_result['data']["sorted_search_list"]


        # ========= 标签召回的结果需要置顶到最前面---去重并取topK start =========
        if label_useful_list:
            new_search_list = []
            new_scores = []
            tmp_sl_content = {}  # 去重使用
            for item in label_useful_list:
                item["snippet"] = item["content"]
                del item["content"]
                item["title"] = item["file_name"]
                del item["file_name"]
                if item["content_id"] not in tmp_sl_content:
                    new_search_list.append(item)
                    new_scores.append(1)
                    tmp_sl_content[item['content_id']] = item['snippet']

            for s, x in zip(sorted_scores, sorted_search_list):
                if x['content_id'] not in tmp_sl_content:
                    tmp_sl_content[x['content_id']] = x['snippet']
                    new_search_list.append(x)
                    new_scores.append(s)

            # 先按 sorted_scores 排序 search_list 再取 topk
            sorted_search_list, sorted_scores = zip(*sorted(zip(new_search_list, new_scores), key=lambda x: x[1], reverse=True))
            if len(sorted_search_list) > top_k:  # 取topK
                sorted_search_list = sorted_search_list[:top_k]
                sorted_scores = sorted_scores[:top_k]
        # ========= 标签召回的结果需要置顶到最前面---去重并取topK  end =========

        sorted_scores, sorted_search_list, has_child = aggregate_chunks(sorted_scores, sorted_search_list)
        logger.info(f"aggregate_chunks result, has_child: {has_child}, sorted_scores: {sorted_scores}, sorted_search_list: {sorted_search_list}")
        # ======= 将SPO及社区报告置顶 start =======
        if graph_data_list:
            new_search_list = []
            new_scores = []
            for item in graph_data_list:  # 将SPO及社区报告置顶
                new_search_list.append(item)
                new_scores.append(1)
            for s, x in zip(sorted_scores, sorted_search_list):
                new_search_list.append(x)
                new_scores.append(s)
            sorted_search_list = new_search_list[:top_k]
            sorted_scores = new_scores[:top_k]

        replace_minio_ip(sorted_search_list)
        response_info = rerank_utils.assemble_search_result(question, sorted_scores, sorted_search_list, rate, return_meta,
                                                            prompt_template, default_answer, auto_citation)

        logger.info('重排结果：' + repr(response_info))

        if response_info['code'] != 0:
            raise RuntimeError(response_info['message'])

        if len(response_info['data']['searchList']) == 0:
            response_info['data']["prompt"] = question
            response_info['data']["searchList"] = []

        return response_info
    except Exception as e:
        logger.warn(f"get_knowledge_based_answer Failed: {e}")
        logger.error(traceback.format_exc())
        response_info["code"] = 1
        response_info["message"] = str(e)
        return response_info


def aggregate_chunks(sorted_scores, sorted_search_list):
    """
    聚合子片段到父片段中
    """

    parent_child_map = {}
    parent_items = {}
    parent_score = {}

    for index, item in enumerate(sorted_search_list):
        content_id = item["content_id"]
        if 'is_parent' in item and item['is_parent'] is False:
            if content_id not in parent_child_map:
                parent_child_map[content_id] = {"search_list":[], "score":[]}

            parent_child_map[content_id]["search_list"].append(item)
            parent_child_map[content_id]["score"].append(sorted_scores[index])
        else:
            if content_id not in parent_items:
                parent_items[content_id] = copy.deepcopy(item)
                parent_items[content_id]["rerank_info"]= []

            if "content_type" in item and item["content_type"] == "image":
                parent_items[content_id]["rerank_info"].append({
                    "type": "image",
                    "file_url": item["file_url"],
                    "score": sorted_scores[index]
                })
                # reset parent content type, output type only [graph, community_report, text]
                parent_items[content_id]["content_type"] = "text"
            else:
                parent_items[content_id]["rerank_info"].append({
                    "type": "text",
                    "content": item["snippet"],
                    "score": sorted_scores[index]
                })

            if content_id not in parent_score:
                parent_score[content_id] = sorted_scores[index]
            parent_score[content_id] = max(sorted_scores[index], parent_score[content_id])

    has_child = True if parent_child_map else False

    # 处理有子片段的父片段
    for content_id, children in parent_child_map.items():
        if content_id not in parent_items:
            # 获取父片段信息
            kb_name = children["search_list"][0]["kb_name"]
            kb_id = children["search_list"][0].get("kb_id", "")
            user_id = children["search_list"][0]["user_id"]
            content_response = milvus_utils.get_content_by_ids(user_id, {"kb_name": kb_name, "kb_id": kb_id}, [content_id])
            logger.info(f"获取父分段 content_id: {content_id}, 结果: {content_response}")
            if content_response['code'] != 0:
                logger.error(f"获取分段信息失败， user_id: {user_id},kb_name: {kb_name}, content_id: {content_id}")
                continue

            parent_content = content_response["data"]["contents"][0]

            child_score_list = []
            for index, item in enumerate(children["search_list"]):
                item["child_snippet"] = item["snippet"]
                child_score_list.append(children["score"][index])

            max_score = max(child_score_list)
            parent_items[content_id] = {
                "title": parent_content["file_name"],
                "snippet": parent_content["content"],
                "kb_name": parent_content["kb_name"],
                "content_id": parent_content["content_id"],
                "meta_data": parent_content["meta_data"],
                "child_content_list": children["search_list"],
                "rerank_info": [],
                "child_score": child_score_list,
                "score": max_score,
                "is_parent": True,
            }

        for index, item in enumerate(children["search_list"]):
            if "content_type" in item and item["content_type"] == "image":
                parent_items[content_id]["rerank_info"].append({
                    "type": "image",
                    "file_url": item["file_url"],
                    "score": sorted_scores[index]
                })
            else:
                parent_items[content_id]["rerank_info"].append({
                    "type": "text",
                    "content": item["snippet"],
                    "score": sorted_scores[index]
                })

        parent_score[content_id] = max_score

    # 按分数降序排序后返回
    sorted_parent_items = sorted(parent_items.items(), key=lambda x: parent_score[x[0]], reverse=True)
    sorted_scores_list = [parent_score[item[0]] for item in sorted_parent_items]
    sorted_items_list = [item[1] for item in sorted_parent_items]

    return sorted_scores_list, sorted_items_list, has_child


def is_valid_string(s):
    pattern = r'^[0-9a-zA-Z\u4e00-\u9fa5_-]+$'
    return re.match(pattern, s) is not None


def _parse_minio_url(url):
    """从 MinIO 下载 URL 中解析出 (bucket_name, object_name)，失败返回 (None, None)。"""
    prefix = REPLACE_MINIO_DOWNLOAD_URL.rstrip('/') + '/'
    if url.startswith(prefix):
        # lstrip('/') 处理 api// 双斜杠的情况
        path = url[len(prefix):].lstrip('/').split('?')[0]
        parts = path.split('/', 1)
        if len(parts) == 2 and parts[0]:
            return parts[0], parts[1]
    # 兜底：匹配内网 IP 格式，/* 兼容双斜杠
    m = re.match(r'https?://[^/]+/minio/download/api//*([^/]+)/(.+?)(?:\?.*)?$', url)
    if m:
        return m.group(1), m.group(2)
    return None, None


def replace_minio_ip(search_list, only_meta=False):
    """将 search_list 中所有 MinIO URL 替换为预签名 URL（原地修改）。
    相同的原始 URL 只调用一次 create_download_url，其余复用缓存结果。
    应在 assemble_search_result 之前调用，使 prompt 中的 URL 也随 snippet 一并处理。
    """
    # 匹配含 /minio/download/api/ 的 URL（内网 IP 或外网地址均可），/* 兼容双斜杠
    minio_url_re = re.compile(r'https?://[^\s\)"\'>\]]+/minio/download/api//*[^\s\)"\'>\]]*')
    url_presigned_cache = {}

    def get_presigned(url):
        if url in url_presigned_cache:
            return url_presigned_cache[url]
        bucket, obj = _parse_minio_url(url)
        if not bucket:
            url_presigned_cache[url] = url
            return url
        new_url = minio_utils.create_download_url(bucket, obj, expire=timedelta(days=1))
        result = new_url if new_url else url
        url_presigned_cache[url] = result
        return result

    for item in search_list:
        # meta_data download_link：用规范化 key 走同一缓存
        meta = item.get('meta_data', {})
        if 'bucket_name' in meta and 'object_name' in meta:
            raw = f"{REPLACE_MINIO_DOWNLOAD_URL.rstrip('/')}/{meta['bucket_name']}/{meta['object_name']}"
            meta['download_link'] = get_presigned(raw)

        if only_meta: continue

        # snippet：替换文本内嵌的所有 MinIO URL 为预签名 URL
        item['snippet'] = minio_url_re.sub(lambda m: get_presigned(m.group(0)), item['snippet'])

        # rerank_info 里 type==image 的 file_url / type==text 的 content
        for rerank_item in item.get('rerank_info', []):
            if rerank_item.get('type') == 'image' and 'file_url' in rerank_item:
                rerank_item['file_url'] = get_presigned(rerank_item['file_url'])
            elif rerank_item.get('type') == 'text' and 'content' in rerank_item:
                rerank_item['content'] = minio_url_re.sub(lambda m: get_presigned(m.group(0)), rerank_item['content'])


def convert_office_file(file_path, target_dir, target_format):
    # 检查文件夹是否存在，如果不存在则创建
    if not os.path.exists(target_dir):
        os.makedirs(target_dir)
    # 获取文件名和扩展名
    _, filename_no_path = os.path.split(os.path.abspath(file_path))  # 提取文件名（包含后缀）
    base_filename, file_extension = os.path.splitext(filename_no_path)  # 分离文件名和后缀
    # ===== 首先把文件另存为英文临时文件 =====
    # 生成一个唯一的 UUID 作为临时文件名
    temp_file_name = str(uuid.uuid4())
    # 构造临时文件的完整路径
    temp_file_path = os.path.join(target_dir, temp_file_name + file_extension)
    # 将原始文件复制为临时文件
    shutil.copy(file_path, temp_file_path)
    logger.info(f"{file_path}文件已成功另存为临时文件：{temp_file_path}")
    if file_extension in [".ofd"]:  # ofd格式文件转换
        dst_path = os.path.join(target_dir, f"{temp_file_name}.{target_format}")
        # print(temp_file_path, "======", dst_path)
        try:
            with open(temp_file_path, "rb") as f:
                ofdb64 = str(base64.b64encode(f.read()), "utf-8")
            try:
                # ============ 第一种方法，easyofd  =============
                ofd = OFD()  # 初始化OFD 工具类
                ofd.read(ofdb64, save_xml=True, xml_name=f"{temp_file_name}_xml")  # 读取ofdb64
                # print("ofd.data", ofd.data) # ofd.data 为程序解析结果
                pdf_bytes = ofd.to_pdf()  # 转pdf
                # img_np = ofd.to_jpg()  # 转图片
                ofd.del_data()
                # ============ 第一种方法，easyofd =============
            except Exception as e:
                logger.info(f"easyofd Error ofd2pdf: {e}")
                # ============ 第二种方法，ofdparser =============
                parser = OfdParser(ofdb64)
                pdf_bytes = parser.ofd2pdf()
                # ============ 第二种方法，ofdparser =============

            with open(dst_path, "wb") as f:
                f.write(pdf_bytes)
        except Exception as e:
            # print(e)
            logger.info(f"Error ofd2pdf: {e}")
    else:  # 使用 soffice 转换
        # 构造命令
        command = f"/usr/bin/soffice --headless --convert-to {target_format} {temp_file_path} --outdir {target_dir}"
        # 执行命令并等待完成
        try:
            # 设置命令运行超时时间
            result = subprocess.run(command, shell=True, check=True, capture_output=True, text=True, timeout=300)
        except subprocess.TimeoutExpired:
            logger.info(f"{command}命令超时，已尝试终止进程。")
        except subprocess.CalledProcessError as e:
            logger.info(f"Error during command execution: {e}")
    res_filename = os.path.join(target_dir, f"{temp_file_name}.{target_format}")
    # 检查文件是否存在
    if os.path.exists(res_filename):
        logger.info(f"{file_path} convert_office_file successfully => {res_filename}")
        return res_filename
    else:
        logger.info(f"convert_office_file err => {file_path} ,res_filename:{res_filename}")
        return False
