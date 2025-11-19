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

from easyofd.ofd import OFD
from ofdparser import OfdParser
import base64
from datetime import datetime, timedelta
from enum import Enum

import nltk
# 设置NLTK数据路径 [EN] Set NLTK data path
# 获取当前文件的绝对路径 [EN] Get the absolute path of the current file
current_file_path = os.path.abspath(__file__)
# 获取当前文件所在的目录 [EN] Get the directory where the current file is located
current_dir = os.path.dirname(current_file_path)
# 拼接nltk_data文件夹的路径 [EN] Splice the path to the nltk_data folder
nltk_data_path = os.path.join(current_dir, 'nltk_data')
nltk.data.path.append(nltk_data_path)
nltk.data.path.append("/opt/nltk_data")

# 验证设置是否成功 [EN] Verify setup is successful
from utils import milvus_utils
from utils import es_utils
from utils import file_utils
from utils import rerank_utils
from utils import minio_utils
from utils import redis_utils
from utils import graph_utils
from utils import timing
import time

from logging_config import setup_logging
from settings import REPLACE_MINIO_DOWNLOAD_URL
from settings import USE_POST_FILTER
from settings import GRAPH_SERVER_URL
from utils.constant import USER_DATA_PATH

logger_name = 'rag_kb_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

user_data_path = Path(USER_DATA_PATH)
chunk_label_redis_client = redis_utils.get_redis_connection(redis_db=5)


# -----------------
# 初始化知识库 [EN] Initialize knowledge base
def init_knowledge_base(user_id, kb_name, kb_id="", embedding_model_id="", enable_knowledge_graph = False):
    response_info = {'code': 0, "message": "成功"}
    # ----------------1、检测向量库名称是否合法 [EN] ----------------1. Check whether the vector library name is legal.
    kb_is_legal = is_valid_string(user_id + kb_name)
    if not kb_is_legal:
        response_info['code'] = 1
        response_info['message'] = '知识库名称仅能包括大小写英文、数字、中文和_符号'
        logger.error('向量库命名不符合规范')
        return response_info
    # ----------------2、check 向量库 是否有重复的 [EN] ----------------2. Check the vector library to see if there are duplicates
    milvus_data = list_knowledge_base(user_id)
    logger.info('向量库已有知识库查询结果：')
    logger.info(repr(milvus_data))

    if milvus_data['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = '向量库校验失败'
        return response_info
    if kb_name in milvus_data['data']['knowledge_base_names']:
        response_info['code'] = 1
        response_info['message'] = '已存在相同名字的向量知识库'
        return response_info
    # ----------------2、建立向量库 [EN] ----------------2. Create vector library
    milvus_init_result = milvus_utils.init_knowledge_base(user_id, kb_name,
                                                          kb_id = kb_id,
                                                          embedding_model_id = embedding_model_id,
                                                          enable_knowledge_graph = enable_knowledge_graph)
    logger.info('向量库初始化结果：')
    logger.info(repr(milvus_init_result))

    if milvus_init_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = milvus_init_result['message']
        return response_info

    # ----------------3、建立路径 [EN] ----------------3. Create a path
    if not os.path.exists(os.path.join(user_data_path, user_id)):
        os.mkdir(os.path.join(user_data_path, user_id))
    if os.path.exists(os.path.join(user_data_path, user_id, kb_name)):
        shutil.rmtree(os.path.join(user_data_path, user_id, kb_name))
    if not os.path.exists(os.path.join(user_data_path, user_id, kb_name)):
        os.mkdir(os.path.join(user_data_path, user_id, kb_name))
    return response_info


# -----------------
# 查询所有知识库 [EN] Query all knowledge bases
def list_knowledge_base(user_id):
    milvus_list_kb_result = milvus_utils.list_knowledge_base(user_id)
    logger.info('用户知识库查询结果：' + repr(milvus_list_kb_result))
    return milvus_list_kb_result


# -----------------
# 查询所有文档 [EN] Query all documents
def list_knowledge_file(user_id, kb_name, kb_id=""):
    milvus_list_file_result = milvus_utils.list_knowledge_file(user_id, kb_name, kb_id=kb_id)
    logger.info('用户知识库文档查询结果：' + repr(milvus_list_file_result))
    return milvus_list_file_result


def list_knowledge_file_download_link(user_id, kb_name, kb_id=""):
    """ 获取知识库里所有文档的下载链接 """
    milvus_list_file_result = milvus_utils.list_knowledge_file_download_link(user_id, kb_name, kb_id=kb_id)
    logger.info('获取知识库里所有文档的下载链接结果：' + repr(milvus_list_file_result))
    if milvus_list_file_result['code'] == 0:  # 替换好 minio下载链接 [EN] Replace the minio download link
        file_download_links = []
        for url in milvus_list_file_result['data']['file_download_links']:
            # 正则表达式匹配 https://ip:port/minio/download/api/ 部分 [EN] Regular expression matching https://ip:port/minio/download/api/ part
            pattern = r'http?://[^/]+/minio/download/api/'
            # 替换文本中的URL [EN] Replace URL in text
            file_download_links.append(re.sub(pattern, REPLACE_MINIO_DOWNLOAD_URL, url))
        milvus_list_file_result['data']['file_download_links'] = file_download_links

    return milvus_list_file_result

# -----------------
# 校验知识库是否存在 [EN] Verify whether the knowledge base exists
def check_knowledge_base(user_id, kb_name, kb_id=""):
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


# -----------------删除知识库 [EN] ------------------Delete knowledge base
def del_konwledge_base(user_id, kb_name, kb_id=""):
    kb_path = os.path.join(user_data_path, user_id, kb_name)
    response_info = {'code': 0, "message": "成功"}
    # ====== check 知识库是否存在 === [EN] ====== check whether the knowledge base exists ===
    milvus_data = list_knowledge_base(user_id)
    if kb_name not in milvus_data['data']['knowledge_base_names']:
        response_info['code'] = 1
        response_info['message'] = f'{kb_name},知识库不存在'
        return response_info

     #删除 知识图谱 [EN] Delete knowledge graph
    kb_info = milvus_utils.get_kb_info(user_id, kb_name)
    if "enable_knowledge_graph" in kb_info and kb_info["enable_knowledge_graph"]:
        try:
            graph_utils.delete_kb_graph(user_id, kb_name)
            logger.info(f"知识图谱删除成功, kb_name:{kb_name}")
            graph_redis_client = redis_utils.get_redis_connection()
            kb_id = kb_info["id"]
            redis_utils.delete_graph_vocabulary_set(graph_redis_client, kb_id)
        except Exception as e:
            logger.error(f"知识图谱删除失败, error: {repr(e)}")

    # --------------1、删除es库 (必须先删除es库，否则会报错) [EN] ---------------1. Delete the es library (the es library must be deleted first, otherwise an error will be reported)
    del_es_result = es_utils.del_es_kb(user_id, kb_name, kb_id=kb_id)
    logger.info('用户es库删除结果：' + repr(del_es_result))
    if del_es_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = del_es_result['message']
        if '不存在' in del_es_result['message']:
            if os.path.exists(kb_path): shutil.rmtree(kb_path)
        return response_info

    # --------------2、删除向量库 [EN] ---------------2. Delete vector library
    del_milvus_result = milvus_utils.del_milvus_kbs(user_id, kb_name, kb_id=kb_id)
    logger.info('用户milvus库删除结果：' + repr(del_milvus_result))
    if del_milvus_result['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = del_milvus_result['message']
        if '不存在' in del_milvus_result['message']:
            if os.path.exists(kb_path): shutil.rmtree(kb_path)
        return response_info

    # --------------3、删除路径 [EN] ---------------3. Delete path
    kb_path = os.path.join(user_data_path, user_id, kb_name)
    if os.path.exists(kb_path):
        shutil.rmtree(kb_path)
    return response_info


# -----------------删除多个文档 [EN] ------------------Delete multiple documents
def del_knowledge_base_files(user_id, kb_name, file_names, kb_id=""):
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

    # --------------2、删除向量库、es库中文档 [EN] ---------------2. Delete documents in vector library and es library
    success_files = []
    failed_files = []
    for file_name in file_names:
        # 删除milvus [EN] Remove milvus
        del_milvus_result = milvus_utils.del_milvus_files(user_id, kb_name, [file_name], kb_id=kb_id)
        logger.info('向量库文档删除结果：' + repr(del_milvus_result))

        if del_milvus_result['code'] != 0:
            failed_files.append([file_name, del_milvus_result['message']])
            continue
        else:
            success_files.append(file_name)
        # 删除es [EN] Delete
        del_es_result = es_utils.del_es_file(user_id, kb_name, file_name, kb_id=kb_id)
        logger.info('es库文档删除结果：' + repr(del_es_result))

        if del_es_result['code'] != 0:
            failed_files.append([file_name, del_es_result['message']])
            continue
        else:
            success_files.append(file_name)

     #删除 知识图谱 [EN] Delete knowledge graph
    kb_info = milvus_utils.get_kb_info(user_id, kb_name)
    if "enable_knowledge_graph" in kb_info and kb_info["enable_knowledge_graph"]:
        try:
            for file_name in success_files:
                graph_utils.delete_file_from_graph(user_id, kb_name, file_name)
                logger.info(f"知识图谱删除成功, file_name:{file_name}")
        except Exception as e:
            failed_files.append([file_name, f"知识图谱删除文件失败, error: {repr(e)}"])
            logger.error(f"知识图谱删除失败, file_name:{file_name}, error: {repr(e)}")

    # --------------2、路径文档 [EN] ---------------2. Path document
    for file_name in success_files:
        del_file_path = os.path.join(filepath, file_name)
        if os.path.isfile(del_file_path): os.remove(del_file_path)
    for i in failed_files:
        if '文档不存在' in i[1]:
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


def add_files(user_id, kb_name, files, sentence_size, overlap_size, chunk_type, separators, is_enhanced,
              parser_choices, ocr_model_id, pre_process, meta_data_rules):
    response_info = {'code': 0, "message": "成功"}
    filepath = os.path.join(user_data_path, user_id, kb_name)
    if not os.path.exists(filepath): os.makedirs(filepath)

    duplicate_files = []
    unique_files = []
    add_files = []
    failed_files = []
    success_files = []

    # --------------1、check milvus
    files_in_milvus = list_knowledge_file(user_id, kb_name)
    logger.info('向量库已有文档查询结果：' + repr(files_in_milvus))

    if files_in_milvus['code'] != 0:
        response_info['code'] = 1
        response_info['message'] = '文档向量库重复查询校验失败'
        return response_info
    filenames_in_milvus = files_in_milvus['data']['knowledge_file_names']
    # filenames_in_milvus=[]
    for f in files:
        if f.filename in filenames_in_milvus:
            duplicate_files.append(f.filename)
        else:
            unique_files.append(f.filename)

    # --------------2、save

    for f in files:
        if f.filename not in unique_files: continue

        # --------------2.1、save to local
        add_file_path = os.path.join(filepath, f.filename)
        f.save(add_file_path)
        logger.info('文件路径是：' + (add_file_path))
        # 检查文件是否存在 [EN] Check if the file exists
        if os.path.exists(add_file_path):
            logger.info('文件已成功保存存在本地, 文件路径是：' + (add_file_path))
        else:
            logger.info(add_file_path + ",文件在本地不存在，未保存成功")

        # --------------2.2、save to minio
        start_time = int(round(time.time() * 1000))
        minio_result = minio_utils.upload_local_file(add_file_path)
        cost1 = int(round(time.time() * 1000)) - start_time

        logger.info(repr(f.filename) + '上传minio花费时间：' + repr(cost1))
        logger.info(repr(f.filename) + '上传minio结果：' + repr(minio_result))

        if minio_result['code'] != 0:
            failed_files.append([f.filename, '上传minio失败'])
            if os.path.exists(add_file_path): os.remove(add_file_path)
            continue
        else:
            download_link = minio_result['download_link']
            add_files.append([f.filename, download_link])

    # --------------3、split chunk
    for pairs in add_files:

        add_file_name = pairs[0]
        download_link = pairs[1]

        add_file_path = os.path.join(filepath, add_file_name)
        split_config = file_utils.SplitConfig(
            sentence_size=sentence_size,
            overlap_size=overlap_size,
            chunk_type=chunk_type,
            separators=separators,
            parser_choices=parser_choices,
            ocr_model_id=ocr_model_id
        )
        sub_chunk, chunks = file_utils.split_text_file(add_file_path, download_link, split_config)

        if is_enhanced == 'true' and len(chunks) > 0:
            logger.info(f'is_enhanced:{is_enhanced}')

        logger.info(repr(add_file_name) + '文档切分长度：' + repr(len(chunks)))
        logger.info(repr(add_file_name) + '文档递归切分长度：' + repr(len(sub_chunk)))

        if len(chunks) == 0:
            failed_files.append([add_file_name, '文档切分失败'])
            continue
        if len(sub_chunk) == 0:
            failed_files.append([add_file_name, '文档递归切分失败'])
            continue
        with open("./data/%s_chunk.txt" % add_file_name, 'w', encoding='utf-8') as chunks_file:
            for item in chunks:
                chunks_file.write(json.dumps(item, ensure_ascii=False))
                chunks_file.write("\n")
        with open("./data/%s_subchunk.txt" % add_file_name, 'w', encoding='utf-8') as sub_chunk_file:
            for item in sub_chunk:
                sub_chunk_file.write(json.dumps(item, ensure_ascii=False))
                sub_chunk_file.write("\n")

        # --------------4、insert milvus
        insert_milvus_result = milvus_utils.add_milvus(user_id, kb_name, sub_chunk, add_file_name, add_file_path)
        logger.info(repr(add_file_name) + '添加milvus结果：' + repr(insert_milvus_result))
        if insert_milvus_result['code'] != 0:
            failed_files.append([add_file_name, insert_milvus_result['message']])
            continue

        # --------------5、insert es
        insert_es_result = es_utils.add_es(user_id, kb_name, chunks, add_file_name)
        logger.info(repr(add_file_name) + '添加es结果：' + repr(insert_es_result))

        if insert_es_result['code'] != 0:
            failed_files.append([add_file_name, insert_es_result['message']])
            continue
    # --------------6、后处理 [EN] ---------------6. Post-processing
    if len(duplicate_files) == 0 and len(failed_files) == 0:
        return response_info
    else:
        for ff in failed_files:
            del_failed_name = ff[0]
            del_file_path = os.path.join(filepath, del_failed_name)
            if os.path.isfile(del_file_path):
                os.remove(del_file_path)
        m1 = ''
        if len(duplicate_files) > 0: m1 = ','.join(duplicate_files) + '上传文件重复。'
        m2 = ''
        if len(failed_files) > 0:
            m2 = '。'.join([i[0] + '上传失败，' + i[1] for i in failed_files])
        response_info = {'code': 1, "message": m1 + m2}
        return response_info


def get_file_content_list(user_id: str, kb_name: str, file_name: str, page_size: int, search_after: int, kb_id=""):
    """
    获取知识库文件片段列表,用于分页展示
    """
    logger.info(f"get_file_content_list start: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, file_name: {file_name}, "
                f"page_size:{page_size}, search_after:{search_after}")
    response_info = milvus_utils.get_milvus_file_content_list(user_id, kb_name, file_name, page_size,
                                                              search_after, kb_id=kb_id)
    logger.info(f"get_file_content_list end: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, file_name: {file_name}, "
                f"page_size:{page_size}, search_after:{search_after}, response: {response_info}")
    return response_info

def get_file_child_content_list(user_id: str, kb_name: str, file_name: str, chunk_id: int, kb_id=""):
    """
    获取知识库文件子片段列表
    """
    logger.info(f"get_file_child_content_list start: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id:{chunk_id}")
    response_info = milvus_utils.get_milvus_file_child_content_list(user_id, kb_name, file_name, chunk_id, kb_id=kb_id)
    logger.info(f"get_file_child_content_list end: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id:{chunk_id}, response: {response_info}")
    return response_info

class MetadataOperation(Enum):
    """
    元数据操作类型枚举
    """
    UPDATE_METAS = "update_metas"
    DELETE_KEYS = "delete_keys"
    RENAME_KEYS = "rename_keys"

def manage_kb_metadata(user_id: str, kb_name: str, operation: MetadataOperation, data: dict, kb_id=""):
    """
    知识库元数据操作
    """
    if not data:
        logger.warning("未提供操作数据")
        return {'code': 1, 'message': '未提供操作数据'}

    logger.info(f"metadata operation start, user_id: {user_id}, kb_name:{kb_name}, "
                f"kb_id:{kb_id}, operation: {operation.value}, data: {data}")

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
    response_info = milvus_utils.update_file_metas(user_id, kb_name, data, kb_id=kb_id)
    logger.info(f"metadata operation end, user_id: {user_id}, kb_name:{kb_name}, "
                f"kb_id:{kb_id}, operation: {operation.value}, data: {data}, response: {response_info}")

    return response_info


def update_content_status(user_id: str, kb_name: str, file_name: str, content_id: str, status: bool,
                          on_off_switch=None, kb_id=""):
    """
    根据content_id更新知识库文件片段状态
    """
    logger.info('========= update_content_status start：' + repr(user_id) + '，' + repr(kb_name) + '，' + repr(kb_id) +
                '，' + repr(file_name) + '，' + repr(content_id) + '，' + repr(status) + '，' + repr(on_off_switch))
    response_info = milvus_utils.update_milvus_content_status(user_id, kb_name, file_name, content_id, status,
                                                              on_off_switch, kb_id=kb_id)
    logger.info('========= update_content_status end：' + repr(user_id) + '，' + repr(kb_name) + '，' + repr(kb_id) +
                '，' + repr(file_name) + '，' + repr(content_id) + '，' + repr(status) + '，' + repr(on_off_switch) +
                ' ====== response:' + repr(
        response_info))
    return response_info


def get_kb_name_id(user_id: str, kb_name: str):
    """
    获取某个知识库映射的 kb_id接口
    """
    logger.info('========= get_kb_name_id start：' + repr(user_id) + '，' + repr(kb_name))
    response_info = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)
    logger.info('========= get_kb_name_id end：' + repr(user_id) + '，' + repr(kb_name) + ' ====== response:' + repr(response_info))
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


def get_knowledge_based_answer(user_id, kb_names, question, rate, top_k, chunk_conent, chunk_size, return_meta=False,
                               prompt_template='', search_field='content', default_answer='根据已知信息，无法回答您的问题。',
                               auto_citation=False, retrieve_method="hybrid_search", kb_ids=[],
                               filter_file_name_list=[], rerank_model_id='', rerank_mod="rerank_model",
                               weights: Optional[dict] | None = None, term_weight_coefficient=1,
                               metadata_filtering_conditions=[], knowledge_base_info={}, use_graph=False):
    """ knowledge_base_info: {"user_id1": [{ "kb_id": "","kb_name": ""}, { "kb_id": "","kb_name": ""}]}"""
    try:
        if search_field == 'emc':
            search_field = 'embedding_content'
        else:
            search_field = 'content'

        # 向量召回 [EN] vector recall
        response_info = {'code': 0, "message": "成功", "data": {"prompt": "", "searchList": []}}

        if top_k == 0:
            response_info['data']["prompt"] = question
            response_info['data']["searchList"] = []
            return response_info
        if knowledge_base_info:  # 整理格式 [EN] Format
            for user_id, kb_info_list in knowledge_base_info.items():
                knowledge_base_info[user_id] = [kb_info["kb_name"] for kb_info in kb_info_list]
        else:
            knowledge_base_info = {user_id: kb_names}
        milvus_useful_list = []  # 后过滤有效的知识片段 [EN] Post-filter valid knowledge fragments
        es_useful_list = []  # 后过滤有效的知识片段 [EN] Post-filter valid knowledge fragments
        label_useful_list = []  # 后过滤有效的知识片段 [EN] Post-filter valid knowledge fragments
        graph_search_list = []  # 知识图谱关联增强片段 [EN] Knowledge graph association enhancement fragments
        graph_data_list = []  # SPO及社区报告置顶片段 [EN] SPO and community report top clip

        for user_id, kb_names in knowledge_base_info.items():
            if retrieve_method in {"semantic_search", "hybrid_search"}:
                # 向量召回 [EN] vector recall
                search_result = milvus_utils.search_milvus(user_id, kb_names, top_k, question, threshold=rate,
                                                           search_field=search_field, kb_ids=[],
                                                           filter_file_name_list=filter_file_name_list,
                                                           metadata_filtering_conditions = metadata_filtering_conditions)

                logger.info(repr(user_id) + repr(kb_names) + repr(question) + '问题向量库查询结果：' + json.dumps(repr(search_result), ensure_ascii=False))

                if search_result['code'] != 0:
                    response_info['code'] = search_result['code']
                    response_info['message'] = search_result['message']
                    return response_info
                milvus_search_list = search_result['data']["search_list"]
                if retrieve_method == "semantic_search" and search_field == "content":  # 只召回向量库 [EN] Recall only vector libraries
                    tmp_content = []
                    search_list = []
                    for i in milvus_search_list:  # 去重 [EN] Remove duplicates
                        if i["content"] in tmp_content:
                            continue
                        search_list.append(i)
                        tmp_content.append(i["content"])
                    milvus_search_list = search_list[:top_k]
            else:
                milvus_search_list = []
            # es召回 [EN] es recall
            if retrieve_method in {"full_text_search", "hybrid_search"}:
                # es召回 [EN] es recall
                es_search_list = []
                es_search_list = es_utils.search_es(user_id, kb_names, question, top_k, kb_ids=[],
                                                    filter_file_name_list=filter_file_name_list,
                                                    metadata_filtering_conditions=metadata_filtering_conditions)
                logger.info(repr(user_id) + repr(kb_names) + repr(question) + '问题es库查询结果：' + json.dumps(repr(es_search_list), ensure_ascii=False))
                if retrieve_method == "full_text_search" and search_field == "content":  # 只召回es库 [EN] Only recall the es library
                    tmp_content = []
                    search_list = []
                    for i in es_search_list:  # 去重 [EN] Remove duplicates
                        if i["snippet"] in tmp_content:
                            continue
                        search_list.append(i)
                        tmp_content.append(i["snippet"])
                    es_search_list = search_list[:top_k]
            else:
                es_search_list = []
            # ========== 标签召回通道判断及调用========== [EN] ========== Label recall channel judgment and call ==========
            unique_labels = set()   # 获取到所有的chunk标签 [EN] Get all chunk tags
            for kb_name in kb_names:
                kb_id = get_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
                unique_labels.update(redis_utils.get_all_chunk_labels(chunk_label_redis_client, kb_id))
            unique_labels_list = list(unique_labels)
            # 初始化一个字典来存储每个标签词的出现次数 [EN] Initialize a dictionary to store the number of occurrences of each tag word
            label_counts = {}
            # 遍历每个标签词，统计其在查询字符串中的出现次数 [EN] Traverse each tag word and count the number of times it appears in the query string
            for label in unique_labels_list:
                if label in question:
                    label_counts[label] = question.count(label)
            # 开始调用标签召回 [EN] Start calling tag recall
            if label_counts:
                label_scores = []
                # label_search_list = []
                label_search_list = es_utils.search_keyword(user_id, kb_names, label_counts, top_k,
                                                            metadata_filtering_conditions=metadata_filtering_conditions)
            else:
                label_scores = []
                label_search_list = []

            if USE_POST_FILTER:
                # **************************** 后过滤 ****************************** [EN] ****************************** Post-filtration ****************************
                try:
                    logger.info(f"user_id: {user_id}, kb_names: {kb_names}, question: {question}, 后过滤start")
                    # 向量召回和es召回做启停用后过滤,注意多个kb_names时，需要做区分 [EN] Vector recall and es recall are used for filtering after activation and deactivation. Note that when multiple kb_names are used, distinction needs to be made.
                    content_status_json = {}
                    search_lists = [milvus_search_list, es_search_list, label_search_list]
                    for search_list in search_lists:
                        for i in search_list:
                            content_status_json[i["kb_name"]] = content_status_json.get(i["kb_name"], [])
                            if i['content_id'] not in content_status_json[i["kb_name"]]:
                                content_status_json[i["kb_name"]].append(i['content_id'])
                    for k in content_status_json:  # 多个kb_names时，需要做区分 [EN] When there are multiple kb_names, distinction needs to be made
                        useful_content_id_list = milvus_utils.get_milvus_content_status(user_id, k, content_status_json[k])
                        logger.info(
                            repr(user_id) + repr(k) + repr(content_status_json[k]) + '======== get_milvus_content_status：' + repr(
                                useful_content_id_list))
                        for c in milvus_search_list:
                            if c['kb_name'] == k and c['content_id'] in useful_content_id_list:
                                milvus_useful_list.append(c)
                        for c in es_search_list:
                            if c['kb_name'] == k and c['content_id'] in useful_content_id_list:
                                es_useful_list.append(c)
                        for c in label_search_list:
                            if c['kb_name'] == k and c['content_id'] in useful_content_id_list:
                                label_useful_list.append(c)
                    logger.info(f"question: {question}, es_useful_list: {es_useful_list}")
                    logger.info(f"question: {question}, milvus_useful_list: {milvus_useful_list}")
                    logger.info(f"question: {question}, label_counts:{label_counts}, label_useful_list: {label_useful_list}")
                except Exception as e:
                    logger.info(repr(user_id) + repr(kb_names) + repr(question) + '后过滤 == have err：' + repr(e))
                    milvus_useful_list.extend(milvus_search_list)
                    es_useful_list.extend(es_search_list)
                    label_useful_list.extend(label_search_list)
                # **************************** 后过滤 ****************************** [EN] ****************************** Post-filtration ****************************
            else:
                milvus_useful_list.extend(milvus_search_list)
                es_useful_list.extend(es_search_list)
                label_useful_list.extend(label_search_list)

            # ========= 图谱召回---增强关联片段以及三元组以及社区报告 start ========= [EN] ========= Graph Recall---Enhanced Correlation Fragments and Triplets and Community Reports start =========
            if use_graph:  # 如果使用图检索 [EN] If using graph search
                # ======== 将graph检索的结果 和 两路检索的结果进行融合，并重新再过一遍rerank ======== [EN] ======== Fuse the results of graph retrieval with the results of two-way retrieval, and rerank again ========
                temp_graph_search_list, temp_graph_dat_list = graph_utils.get_graph_search_list(user_id, kb_names, question, top_k,
                                                                             kb_ids=[],
                                                                             filter_file_name_list=filter_file_name_list)
                graph_search_list.extend(temp_graph_search_list)  # 直接放进去先 [EN] Just put it in first
                graph_data_list.extend(temp_graph_dat_list)  # 直接放进去先 [EN] Just put it in first

        # 多路召回融合 [EN] Multiple recall fusion
        # reank重排 [EN] rearrange
        if not milvus_useful_list and not es_useful_list:  # 都为空不走重排,直接返回 [EN] If all are empty, there will be no re-arrangement and return directly.
            response_info = {'code': 0, "message": "成功", "data": {"prompt": question, "searchList": [], "score": []}}
            logger.info('useful_list is None 重排结果：' + json.dumps(repr(response_info),ensure_ascii=False))
            return response_info
        if rerank_mod == "rerank_model":
            sorted_scores, sorted_search_list = rerank_utils.get_model_rerank(question, top_k, milvus_useful_list,
                                                                              es_useful_list, rerank_model_id,
                                                                              term_weight_coefficient=term_weight_coefficient)
        elif rerank_mod == "weighted_score":
            sorted_scores, sorted_search_list = es_utils.get_weighted_rerank(user_id, kb_names, question, weights,
                                                                             milvus_useful_list, es_useful_list, top_k)
        else:
            raise Exception("rerank_mod is not valid")


        # ========= 标签召回的结果需要置顶到最前面---去重并取topK start ========= [EN] ========= The results of tag recall need to be moved to the front --- remove duplicates and take topK start =========
        if label_useful_list:
            new_search_list = []
            new_scores = []
            tmp_sl_content = {}  # 去重使用 [EN] Remove duplicates and use
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

            # 先按 sorted_scores 排序 search_list 再取 topk [EN] First sort search_list by sorted_scores and then take topk
            sorted_search_list, sorted_scores = zip(*sorted(zip(new_search_list, new_scores), key=lambda x: x[1], reverse=True))
            if len(sorted_search_list) > top_k:  # 取topK [EN] Take topK
                sorted_search_list = sorted_search_list[:top_k]
                sorted_scores = sorted_scores[:top_k]
        # ========= 标签召回的结果需要置顶到最前面---去重并取topK  end ========= [EN] ========= The results of tag recall need to be moved to the front --- remove duplicates and take topK end =========

        sorted_scores, sorted_search_list, has_child = aggregate_chunks(user_id, sorted_scores, sorted_search_list)
        logger.info(f"aggregate_chunks result, has_child: {has_child}, sorted_scores: {sorted_scores}, sorted_search_list: {sorted_search_list}")
        # ======= 将SPO及社区报告置顶 start ======= [EN] ======= Pin SPO and community reports to the top start =======
        if graph_data_list:
            new_search_list = []
            new_scores = []
            for item in graph_data_list:  # 将SPO及社区报告置顶 [EN] Pin SPO and community reports to the top
                new_search_list.append(item)
                new_scores.append(1)
            for s, x in zip(sorted_scores, sorted_search_list):
                new_search_list.append(x)
                new_scores.append(s)
            sorted_search_list = new_search_list[:top_k]
            sorted_scores = new_scores[:top_k]

        rerank_result = rerank_utils.rerank_search(question, sorted_scores, sorted_search_list, rate, return_meta,
                                                   prompt_template, default_answer, auto_citation)

        rerank_result = replace_minio_ip(rerank_result)
        logger.info('重排结果：' + repr(rerank_result))

        if rerank_result['code'] != 0:
            response_info['code'] = rerank_result['code']
            response_info['message'] = rerank_result['message']
            return response_info
        if len(rerank_result['data']['searchList']) == 0:
            response_info['data']["prompt"] = question
            response_info['data']["searchList"] = []
            return response_info
    except Exception as err:
        logger.warn(f"------>knowledge-file Failed: {err}")
        import traceback
        logger.error(traceback.format_exc())

    return rerank_result


def aggregate_chunks(user_id, sorted_scores, sorted_search_list):
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
            parent_items[content_id] = item
            if content_id not in parent_score:
                parent_score[content_id] = sorted_scores[index]
            parent_score[content_id] = max(sorted_scores[index], parent_score[content_id])

    if not parent_child_map:
        return sorted_scores, sorted_search_list, False

    # 处理有子片段的父片段 [EN] Handling parent fragments that have child fragments
    for content_id, children in parent_child_map.items():
        if content_id in parent_items:
            continue
        # 获取父片段信息 [EN] Get parent fragment information
        kb_name = children["search_list"][0]["kb_name"]
        content_response = milvus_utils.get_content_by_ids(user_id, kb_name, [content_id])
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
            "child_score": child_score_list,
            "score": max_score,
            "is_parent": True,
        }

        parent_score[content_id] = max_score

    # 按分数降序排序后返回 [EN] Return after sorting by score in descending order
    sorted_parent_items = sorted(parent_items.items(), key=lambda x: parent_score[x[0]], reverse=True)
    sorted_scores_list = [parent_score[item[0]] for item in sorted_parent_items]
    sorted_items_list = [item[1] for item in sorted_parent_items]

    return sorted_scores_list, sorted_items_list, True


def is_valid_string(s):
    pattern = r'^[0-9a-zA-Z\u4e00-\u9fa5_-]+$'
    return re.match(pattern, s) is not None


def replace_minio_ip(rerank_result):
    if 'data' not in rerank_result:
        return rerank_result
    if 'prompt' in rerank_result['data']:
        # prompt 中的 minio url 更新替换 [EN] Minio url in prompt updated and replaced
        text = rerank_result['data']['prompt']
        # 正则表达式匹配 https://ip:port/minio/download/api/ 部分 [EN] Regular expression matching https://ip:port/minio/download/api/ part
        pattern = r'http?://[^/]+/minio/download/api/'
        # 替换文本中的URL [EN] Replace URL in text
        replaced_text = re.sub(pattern, REPLACE_MINIO_DOWNLOAD_URL, text)
        rerank_result['data']['prompt'] = replaced_text
    if 'searchList' not in rerank_result['data']:
        return rerank_result
    for i in range(len(rerank_result['data']['searchList'])):
        # content中的 minio url 更新替换 [EN] Minio url in content updated and replaced
        text = rerank_result['data']['searchList'][i]['snippet']
        # 正则表达式匹配 https://ip:port/minio/download/api/ 部分 [EN] Regular expression matching https://ip:port/minio/download/api/ part
        pattern = r'http?://[^/]+/minio/download/api/'
        # 替换文本中的URL [EN] Replace URL in text
        replaced_text = re.sub(pattern, REPLACE_MINIO_DOWNLOAD_URL, text)
        rerank_result['data']['searchList'][i]['snippet'] = replaced_text

        if 'meta_data' not in rerank_result['data']['searchList'][i]:
            continue
        if ('bucket_name' not in rerank_result['data']['searchList'][i]['meta_data'] or
                'object_name' not in rerank_result['data']['searchList'][i]['meta_data']):
            continue
        # 获取原始的 bucket_name 和 object_name 去拿取预签名下载链接 [EN] Get the original bucket_name and object_name to get the pre-signed download link
        bucket_name = rerank_result['data']['searchList'][i]['meta_data']['bucket_name']
        object_name = rerank_result['data']['searchList'][i]['meta_data']['object_name']
        new_url = minio_utils.craete_download_url(bucket_name, object_name, expire=timedelta(days=1))
        rerank_result['data']['searchList'][i]['meta_data']['download_link'] = new_url


    return rerank_result


def convert_office_file(file_path, target_dir, target_format):
    # 检查文件夹是否存在，如果不存在则创建 [EN] Check if the folder exists, if not create it
    if not os.path.exists(target_dir):
        os.makedirs(target_dir)
    # 获取文件名和扩展名 [EN] Get file name and extension
    _, filename_no_path = os.path.split(os.path.abspath(file_path))  # 提取文件名（包含后缀） [EN] Extract file name (including suffix)
    base_filename, file_extension = os.path.splitext(filename_no_path)  # 分离文件名和后缀 [EN] Separate filename and suffix
    # ===== 首先把文件另存为英文临时文件 ===== [EN] ===== First save the file as an English temporary file =====
    # 生成一个唯一的 UUID 作为临时文件名 [EN] Generate a unique UUID as a temporary file name
    temp_file_name = str(uuid.uuid4())
    # 构造临时文件的完整路径 [EN] Construct the full path to the temporary file
    temp_file_path = os.path.join(target_dir, temp_file_name + file_extension)
    # 将原始文件复制为临时文件 [EN] Copy original file to temporary file
    shutil.copy(file_path, temp_file_path)
    logger.info(f"{file_path}文件已成功另存为临时文件：{temp_file_path}")
    if file_extension in [".ofd"]:  # ofd格式文件转换 [EN] ofd format file conversion
        dst_path = os.path.join(target_dir, f"{temp_file_name}.{target_format}")
        # print(temp_file_path, "======", dst_path)
        try:
            with open(temp_file_path, "rb") as f:
                ofdb64 = str(base64.b64encode(f.read()), "utf-8")
            try:
                # ============ 第一种方法，easyofd  ============= [EN] ============ The first method, easyofd =============
                ofd = OFD()  # 初始化OFD 工具类 [EN] Initialize OFD tool class
                ofd.read(ofdb64, save_xml=True, xml_name=f"{temp_file_name}_xml")  # 读取ofdb64 [EN] Read ofdb64
                # print("ofd.data", ofd.data) # ofd.data 为程序解析结果 [EN] print("ofd.data", ofd.data) # ofd.data is the program analysis result
                pdf_bytes = ofd.to_pdf()  # 转pdf [EN] Convert to pdf
                # img_np = ofd.to_jpg()  # 转图片 [EN] img_np = ofd.to_jpg() # Transfer pictures
                ofd.del_data()
                # ============ 第一种方法，easyofd ============= [EN] ============ The first method, easyofd =============
            except Exception as e:
                logger.info(f"easyofd Error ofd2pdf: {e}")
                # ============ 第二种方法，ofdparser ============= [EN] ============ The second method, ofdparser =============
                parser = OfdParser(ofdb64)
                pdf_bytes = parser.ofd2pdf()
                # ============ 第二种方法，ofdparser ============= [EN] ============ The second method, ofdparser =============

            with open(dst_path, "wb") as f:
                f.write(pdf_bytes)
        except Exception as e:
            # print(e)
            logger.info(f"Error ofd2pdf: {e}")
    else:  # 使用 soffice 转换 [EN] Convert using soffice
        # 构造命令 [EN] construction command
        command = f"/usr/bin/soffice --headless --convert-to {target_format} {temp_file_path} --outdir {target_dir}"
        # 执行命令并等待完成 [EN] Execute the command and wait for completion
        try:
            # 设置命令运行超时时间 [EN] Set command execution timeout
            result = subprocess.run(command, shell=True, check=True, capture_output=True, text=True, timeout=300)
        except subprocess.TimeoutExpired:
            logger.info(f"{command}命令超时，已尝试终止进程。")
        except subprocess.CalledProcessError as e:
            logger.info(f"Error during command execution: {e}")
    res_filename = os.path.join(target_dir, f"{temp_file_name}.{target_format}")
    # 检查文件是否存在 [EN] Check if the file exists
    if os.path.exists(res_filename):
        logger.info(f"{file_path} convert_office_file successfully => {res_filename}")
        return res_filename
    else:
        logger.info(f"convert_office_file err => {file_path} ,res_filename:{res_filename}")
        return False
