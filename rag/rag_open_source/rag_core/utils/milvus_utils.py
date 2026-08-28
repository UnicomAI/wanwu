from collections import defaultdict
import requests
import json
import uuid
import os
from typing import List
import re
import logging
import threading
import copy
from threading import Thread

from settings import MILVUS_BASE_URL, TIME_OUT
from utils.minio_utils import check_files_size
from utils.es_utils import allocate_child_chunks

logger = logging.getLogger(__name__)

def make_request(url: str, data: dict):
    response_info = {'code': 0, "message": "成功"}
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            return response_info
        final_response = response.json()
        if final_response['code'] != 0:
            response_info['code'] = final_response['code']
            response_info['message'] = final_response['message']
            return response_info
        # ======== 正常返回 =======
        return final_response
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        return response_info

def generate_chunks_bacth(user_id: str, kb_name: str, chunks: list, batch_size=1000, extract_multimodal_file: bool = True):
    """ 将chunks 按chunk_current_num分组并生成批次数据，每个list的长度为batch_size"""
    batch_data = []
    # 使用defaultdict来聚合数据
    aggregated_data = defaultdict(list)
    # 遍历列表，按照chunk_current_num字段聚合
    temp_num = 0
    for item in chunks:
        temp_num += 1
        if item['meta_data'].get('chunk_current_num', -1) == -1:
            item['meta_data']['chunk_current_num'] = temp_num // 100
        chunk_current_num = item['meta_data']['chunk_current_num']
        aggregated_data[chunk_current_num].append(item)
    # 将聚合后的数据转换为普通字典，以便查看
    aggregated_data = dict(aggregated_data)
    print(aggregated_data)
    is_multimodal = False
    kb_info = get_kb_info(user_id, kb_name)
    if kb_info and "is_multimodal" in kb_info and kb_info.get("is_multimodal"):
        is_multimodal = True
    emb_model_id = kb_info.get("embedding_model_id")
    for key, value in aggregated_data.items():
        # 从 aggregated_data 里提取多模态数据
        if value and is_multimodal and extract_multimodal_file:
            # 需要提取图片链接并校验图片size是否符合emb模型input规格
            image_urls = extract_minio_markdown_images(value[0].get("content", ""))
            if image_urls:
                check_size_result = check_files_size(image_urls, emb_model_id)
                logger.info(f"发现提取到了多模态文件信息:{image_urls},检验大小结果：{check_size_result}")
                for idx, image_url in enumerate(image_urls):  # 遍历图片链接
                    if not check_size_result[idx]:
                        continue
                    image_chunk = {"embedding_content": image_url, "content_type": "image",
                                   "content": value[0]["content"], "meta_data": value[0]["meta_data"]}
                    # 兼容父子分段
                    if "is_parent" in  value[0]:
                        image_chunk["is_parent"] = value[0]["is_parent"]
                    batch_data.append(image_chunk)
        batch_data.extend(value)
        if len(batch_data) >= batch_size:
            yield batch_data
            batch_data = []
    # 最后一个batch
    if batch_data:
        yield batch_data


def init_knowledge_base(user_id: str,
                        kb_info: dict,
                        embedding_model_id: str = "",
                        enable_knowledge_graph: bool = False,
                        is_multimodal: bool = False):
    """
    :param kb_info: {"kb_name": "...", "kb_id": "..."} — kb_id 必须非空（由上游生成或前端提供）
    """
    response_info = {'code': 0, "message": '成功'}
    url = MILVUS_BASE_URL + '/rag/kn/init_kb'
    headers = {'Content-Type': 'application/json'}
    data = {
        "userId": user_id,
        "kb_info": kb_info,
        "embedding_model_id": embedding_model_id,
        "enable_knowledge_graph": enable_knowledge_graph,
        "is_multimodal": is_multimodal
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus初始化请求失败：" + repr(response.text))
            return response_info

        init_response = response.json()
        if init_response['code'] != 0:
            response_info['code'] = init_response['code']
            response_info['message'] = init_response['message']
            logger.error("milvus初始化请求失败：" + repr(init_response))
            return response_info
        else:
            logger.info("milvus初始化请求成功")
            return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus初始化请求异常：" + repr(e))
        return response_info


def list_knowledge_base(user_id):
    response_info = {'code': 0, "message": '成功', "data": {"knowledge_base_names": []}}

    # url='http://localhost:6098/list_kb_names'
    url = MILVUS_BASE_URL + '/rag/kn/list_kb_names'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus查询用户所有知识库请求失败：" + repr(response.text))
            return response_info
        result_data = response.json()
        if result_data['code'] != 0:
            response_info['code'] = result_data['code']
            response_info['message'] = result_data['message']
            logger.error("milvus查询用户所有知识库请求失败：" + repr(result_data))
            return response_info
        else:
            response_info['data']['knowledge_base_names'] = result_data['data']['kb_names']
            logger.info("milvus查询用户所有知识库请求成功")
            return response_info

    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus查询用户所有知识库请求异常：" + repr(e))
        return response_info


def list_knowledge_file(user_id, kb_info):
    response_info = {'code': 0, "message": "成功", "data": {"knowledge_file_names": []}}
    # url='http://localhost:6098/list_file_names'
    url = MILVUS_BASE_URL + '/rag/kn/list_file_names'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_info': kb_info}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus查询知识库所有文档请求失败：" + repr(response.text))
            return response_info
        result_data = response.json()
        if result_data['code'] != 0:
            response_info['code'] = result_data['code']
            response_info['message'] = result_data['message']
            logger.error("milvus查询知识库所有文档请求失败：" + repr(result_data))
            return response_info
        else:
            response_info['data']['knowledge_file_names'] = result_data['data']['file_names']
            logger.info("milvus查询知识库所有文档请求成功")
            return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus查询知识库所有文档请求异常：" + repr(e))
        return response_info


def list_knowledge_file_download_link(user_id, kb_info):
    response_info = {'code': 0, "message": "成功", "data": {"file_download_links": []}}
    # url='http://localhost:6098/list_file_names'
    url = MILVUS_BASE_URL + '/rag/kn/list_file_download_links'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_info': kb_info}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus获取知识库里所有文档的下载链接请求失败：" + repr(response.text))
            return response_info
        result_data = response.json()
        if result_data['code'] != 0:
            response_info['code'] = result_data['code']
            response_info['message'] = result_data['message']
            logger.error("milvus获取知识库里所有文档的下载链接请求失败：" + repr(result_data))
            return response_info
        else:
            response_info['data']['file_download_links'] = result_data['data']['file_download_links']
            logger.info("milvus获取知识库里所有文档的下载链接请求成功")
            return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus获取知识库里所有文档的下载链接请求异常：" + repr(e))
        return response_info


KNN_SEARCH_URL = MILVUS_BASE_URL + '/rag/kn/search'
KNN_COMMUNITY_SEARCH_URL = MILVUS_BASE_URL + '/rag/kn/search_community_reports'
def search_milvus(user_id, kb_infos, top_k, question, threshold, search_field, emb_model="bge",
                  filter_file_name_list=[], metadata_filtering_conditions = [], milvus_url = KNN_SEARCH_URL,
                  enable_vision=False, attachment_files=[]):
    """
    :param kb_infos: [{"kb_name": "...", "kb_id": "..."}, ...]
    :param emb_model:  "bge", "bce", "conna"
    """
    post_data = {}
    post_data["userId"] = user_id
    post_data["kb_infos"] = kb_infos
    post_data["topk"] = top_k * 4
    post_data["question"] = question
    post_data["threshold"] = threshold
    post_data["emb_model"] = emb_model
    post_data["filter_file_name_list"] = filter_file_name_list
    post_data["metadata_filtering_conditions"] = metadata_filtering_conditions
    post_data["enable_vision"] = enable_vision
    post_data["attachment_files"] = attachment_files

    response_info = {'code': 0, "message": "成功", "data": {"prompt": "", "search_list": []}}
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(milvus_url, headers=headers, data=json.dumps(post_data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus问题检索请求失败：" + repr(response.text))
            return response_info
        result_data = response.json()
        if result_data['code'] != 0:
            response_info['code'] = 1
            response_info['message'] = result_data['message']
            logger.error("milvus问题检索请求失败：" + repr(result_data))
            return response_info
        response_info['code'] = result_data['code']
        response_info['message'] = result_data['message']
        milvus_return = result_data['data']['search_list']

        if search_field == 'embedding_content':
            if len(milvus_return) == 0:
                response_info['data']['search_list'] = milvus_return
            else:
                response_info['data']['search_list'] = milvus_return[:top_k]
            return response_info
        else:
            if len(milvus_return) == 0:
                response_info['data']['search_list'] = milvus_return
            else:
                deduplication_list = []
                tmp_content = []
                for search_item in milvus_return:
                    item_content = search_item['content']
                    if item_content not in tmp_content:
                        deduplication_list.append(search_item)
                        tmp_content.append(item_content)
                    else:
                        continue
                del tmp_content
                deduplication_list = deduplication_list[:top_k]
                response_info['data']['search_list'] = deduplication_list
                # response_info['data']['search_list']=milvus_return
            logger.info("milvus问题检索请求成功")
            return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus问题检索请求异常：" + repr(e))
        return response_info


ADD_URL = MILVUS_BASE_URL + '/rag/kn/add'
ADD_COMMUNItY_REPORT_URL = MILVUS_BASE_URL + '/rag/kn/add_community_reports'

def add_milvus(user_id, kb_info, sub_chunk, add_file_name, add_file_path, milvus_url = ADD_URL, extract_multimodal_file: bool = True):
    batch_size = 200
    response_info = {'code': 0, "message": "成功"}
    kb_name = kb_info["kb_name"]
    batch_count = 0
    success_count = 0
    fail_count = 0
    error_reason = []
    # sub_chunk 批次生成器,按 按chunk_current_num分组并生成批次数据
    chunk_gen = generate_chunks_bacth(user_id, kb_name, sub_chunk, batch_size=batch_size, extract_multimodal_file=extract_multimodal_file)
    for batch in chunk_gen:
        insert_data = {}
        insert_data['userId'] = user_id
        insert_data['kb_info'] = kb_info
        chunks_data = []
        for chunk in batch:
            chunk_dict = {
                "content": chunk['content'],
                "embedding_content": chunk['embedding_content'],
                "chunk_id": str(uuid.uuid4()),
                "file_name": add_file_name,
                "oss_path": add_file_path,
                "meta_data": chunk['meta_data']
            }

            if "title" in chunk:
                chunk_dict["title"] = chunk["title"]

            if "create_time" in chunk:
                chunk_dict["create_time"] = chunk["create_time"]

            if "is_parent" in chunk:
                chunk_dict["is_parent"] = chunk["is_parent"]

            if "content_type" in chunk:
                chunk_dict["content_type"] = chunk["content_type"]

            if 'labels' in chunk:
                chunk_dict['labels'] = chunk['labels']
            chunks_data.append(chunk_dict)
        insert_data['data'] = chunks_data
        headers = {"Content-Type": "application/json"}
        batch_count = batch_count + 1
        try:
            response = requests.post(milvus_url, headers=headers, json=insert_data, timeout=TIME_OUT)
            logger.info(repr(add_file_name) + '批量写入milvus请求结果:' + repr(batch_count) + repr(response.text))
            if response.status_code != 200:
                logger.error(repr(add_file_name) + repr(batch_count) + '批量写入milvus请求失败')
                fail_count = fail_count + 1
                if str(response.text) not in error_reason: error_reason.append(str(response.text))
                # ========= 报错直接返回结束 =======
                response_info['code'] = 1
                response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
                return response_info

            result_data = response.json()
            if result_data['code'] != 0:
                fail_count = fail_count + 1
                if str(result_data['message']) not in error_reason: error_reason.append(str(result_data['message']))
                logger.error(repr(add_file_name) + repr(batch_count) + '批量写入milvus请求失败')
                # ========= 报错直接返回结束 =======
                response_info['code'] = 1
                response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
                return response_info
            else:
                success_count = success_count + 1
                logger.info(repr(add_file_name) + repr(batch_count) + '批量写入milvus请求成功')

        except Exception as e:
            logger.error(repr(add_file_name) + repr(batch_count) + '批量写入milvus请求异常：' + repr(e))
            fail_count = fail_count + 1
            if repr(e) not in error_reason: error_reason.append(repr(e))
            # ========= 报错直接返回结束 =======
            response_info['code'] = 1
            response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
            return response_info


    # print('add_milvus方法调用接口批量建库，总批次:%s次，成功:%s次,失败:%s次' % (batch_count, success_count, fail_count))
    logger.info('add_milvus方法调用接口批量建库')
    logger.info('总批次：' + repr(batch_count))
    logger.info('成功：' + repr(success_count))
    logger.info('失败：' + repr(fail_count))

    if batch_count == success_count:
        response_info['code'] = 0
        response_info['message'] = '成功'
    else:
        response_info['code'] = 1
        response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
    return response_info


def del_milvus_kbs(user_id, kb_info):
    response_info = {'code': 0, "message": "成功"}
    url = MILVUS_BASE_URL + '/rag/kn/del_kb'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_info': kb_info}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus删除知识库请求失败: " + repr(response.text))
            return response_info
        del_response = response.json()
        if del_response['code'] != 0:
            response_info['code'] = del_response['code']
            response_info['message'] = del_response['message']
            logger.error("milvus删除知识库请求失败: " + repr(del_response))
            return response_info
        logger.info("milvus删除知识库请求成功")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus删除知识库请求异常: " + repr(e))
        return response_info


def del_milvus_files(user_id, kb_info, file_names):
    response_info = {'code': 0, "message": "成功"}
    url = MILVUS_BASE_URL + '/rag/kn/del_files'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'kb_info': kb_info, 'file_names': file_names}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus删除知识库文档请求失败: " + repr(response.text))
            return response_info
        del_files_response = response.json()
        if del_files_response['code'] != 0:
            response_info['code'] = del_files_response['code']
            response_info['message'] = del_files_response['message']
            logger.error("milvus删除知识库文档请求失败: " + repr(del_files_response))
            return response_info
        logger.info("milvus删除知识库文档请求成功")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error("milvus删除知识库文档请求异常: " + repr(e))
        return response_info


def get_milvus_file_content_list(user_id: str, kb_info: dict, file_name: str, page_size: int,
                                 search_after: int, content_type="text", query_text: str = None):
    """
        获取知识库文件片段列表,用于分页展示
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_content_list'
    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'page_size': page_size,
        'search_after': search_after,
        "content_type": content_type,
        'query_text': query_text
    }

    return make_request(url, data)


def get_milvus_file_child_content_list(user_id: str, kb_info: dict, file_name: str, chunk_id: int,
                                       child_chunk_current_num:int=None):
    """
        获取知识库文件子片段列表
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_child_content_list'
    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'child_chunk_current_num': child_chunk_current_num
    }

    return make_request(url, data)


def list_file_names_after_filtering(user_id, kb_info, filtering_conditions):
    """
        根据file_name更新知识库文件元数据
    """
    url = MILVUS_BASE_URL + '/rag/kn/list_file_names_after_filtering'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'filtering_conditions': filtering_conditions,
    }

    return make_request(url, data)


def update_child_chunk(user_id, kb_info, file_name, chunk_id, chunk_current_num, child_chunk):
    """
        更新知识库子段
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_child_chunk'
    kb_name = kb_info["kb_name"]

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'chunk_id': chunk_id,
        'chunk_current_num': chunk_current_num,
        'child_chunk': child_chunk
    }

    response = make_request(url, data)
    if response['code'] == 0:
        is_multimodal = False
        kb_meta = get_kb_info(user_id, kb_name)
        if kb_meta and "is_multimodal" in kb_meta and kb_meta.get("is_multimodal"):
            is_multimodal = True
        emb_model_id = kb_meta.get("embedding_model_id")

        child_content = child_chunk["child_content"]
        child_chunk_current_num = child_chunk["child_chunk_current_num"]
        try:
            # 从 aggregated_data 里提取多模态数据
            if child_content and is_multimodal:
                batch_data = []
                # 需要提取图片链接并校验图片size是否符合emb模型input规格
                image_urls = extract_minio_markdown_images(child_content)
                if image_urls:
                    check_size_result = check_files_size(image_urls, emb_model_id)
                    logger.info(f"发现提取到了多模态文件信息:{image_urls},检验大小结果：{check_size_result}")
                    for idx, image_url in enumerate(image_urls):  # 遍历图片链接
                        if not check_size_result[idx]:
                            continue
                        image_chunk = {"embedding_content": image_url, "content_type": "image"}
                        batch_data.append(image_chunk)
                if batch_data:
                    content_response = get_content_by_ids(user_id, kb_info, [chunk_id])
                    logger.info(f"获取父分段 content_id: {chunk_id}, 结果: {content_response}")
                    if content_response['code'] != 0:
                        raise RuntimeError(
                            f"获取父分段信息失败， user_id: {user_id},kb_name: {kb_name}, content_id: {chunk_id}")

                    parent_chunk = content_response["data"]["contents"][0]
                    parent_content = parent_chunk["content"]
                    meta_data = parent_chunk["meta_data"]
                    child_chunk_total_num = parent_chunk["child_chunk_total_num"]

                    for item in batch_data:
                        item["content"] = parent_content
                        item["meta_data"] = copy.deepcopy(meta_data)
                        item["is_parent"] = False
                        item["meta_data"]["child_chunk_current_num"] = child_chunk_current_num
                        item["meta_data"]["child_chunk_total_num"] = child_chunk_total_num

                    insert_result = add_milvus(user_id, kb_info, batch_data, file_name, "",
                                               extract_multimodal_file=False)
                    logger.info(f"update_child_chunk添加image到milvus结果：{insert_result}")
                    if insert_result['code'] != 0:
                        raise RuntimeError(f"update_child_chunk添加image到milvus失败, insert_result: {insert_result}")
        except Exception as e:
            # skip add image failure
            logger.error(str(e))

    return response


def update_file_metas(user_id, kb_info, update_datas):
    """
        更新知识库元数据
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_file_metas'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'update_datas': update_datas,
    }

    return make_request(url, data)


def update_chunk_labels(user_id, kb_info, file_name, chunk_id, labels):
    """
        根据file_name和chunk_id更新标签
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_chunk_labels'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'labels': labels,
    }

    return make_request(url, data)


def get_content_by_ids(user_id, kb_info, content_ids, content_type= "text"):
    """
        根据file_name和chunk_id获取分段信息
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_content_by_ids'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'content_ids': content_ids,
        "content_type": content_type
    }

    return make_request(url, data)

def batch_delete_chunks(user_id, kb_info, file_name, chunk_ids):
    """
        根据file_name和chunk_ids删除分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/batch_delete_chunks'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'chunk_ids': chunk_ids,
    }

    return make_request(url, data)


def batch_delete_child_chunks(user_id, kb_info, file_name, chunk_id, chunk_current_num,
                        child_chunk_current_nums):
    """
        根据chunk_id和child_chunk_current_nums删除子分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/delete_child_chunks'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'chunk_id': chunk_id,
        "chunk_current_num": chunk_current_num,
        "child_chunk_current_nums": child_chunk_current_nums
    }

    return make_request(url, data)


def update_milvus_content_status(user_id: str, kb_info: dict, file_name: str, content_id: str, status: bool,
                                 on_off_switch=None):
    """
        根据content_id更新知识库文件片段状态
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_content_status'
    if on_off_switch in [True, False]:  # 前端传递了 on_off_switch 参数
        data = {
            'userId': user_id,
            'kb_info': kb_info,
            'file_name': file_name,
            'content_id': content_id,
            'status': status,
            'on_off_switch': on_off_switch,
        }
    else:
        data = {
            'userId': user_id,
            'kb_info': kb_info,
            'file_name': file_name,
            'content_id': content_id,
            'status': status,
        }

    return make_request(url, data)


def del_community_reports(user_id, kb_info, clear_reports=False, content_ids= []):
    """
        根据chunk_id和child_chunk_current_nums删除子分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/del_community_reports'

    data = {
        'userId': user_id,
        'kb_info': kb_info,
        "clear_reports": clear_reports,
        "content_ids": content_ids
    }

    return make_request(url, data)


def get_kb_info(user_id, kb_name):
    url = MILVUS_BASE_URL + '/rag/kn/get_kb_info'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, "kb_name": kb_name}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        response.raise_for_status()
        result_data = response.json()
        if result_data.get('code') != 0:
            raise RuntimeError(result_data['message'])
        else:
            kb_info = result_data.get('data', {}).get("kb_info", {})
            logger.info(f"milvus查询知识库详情成功:{kb_info}")
            return kb_info
    except Exception as e:
        logger.error(f"milvus查询知识库详情失败: {e}")
        raise RuntimeError(f"milvus查询知识库详情失败: {e}") from e


def is_multimodal_kb(user_id: str,
                     kb_name: str):
    kb_info = get_kb_info(user_id, kb_name)
    if kb_info and "is_multimodal" in kb_info and kb_info.get("is_multimodal"):
        return True

    return False


def get_milvus_content_status(user_id: str, kb_info: dict, content_id_list: list):
    """
        获取文本分块状态用于进行检索后过滤。
    """
    response_info = {'code': 0, "message": "成功"}
    kb_name = kb_info["kb_name"]
    # url = "http://localhost:30041/rag/kn/get_useful_content_status"  # 临时地址
    url = MILVUS_BASE_URL + '/rag/kn/get_useful_content_status'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'kb_info': kb_info, 'content_id_list': content_id_list}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:  # 抛出报错
            err = str(response.text)
            raise RuntimeError(f"{kb_name}-{content_id_list},Error get_milvus_content_status: {err}")
        final_response = response.json()
        if final_response['code'] == 0:  # 正常获取到了结果
            res_list = final_response['data']["useful_content_id_list"]
            return res_list
        else:  # 抛出报错
            raise RuntimeError(
                f"{kb_name}-{content_id_list},Error get_milvus_content_status: {final_response}")
    except Exception as e:
        raise RuntimeError(f"{e}")


def update_milvus_kb_name(user_id: str, old_kb_name: str, new_kb_name: str):
    """
        更新知识库名接口
    """
    response_info = {'code': 0, "message": "成功"}
    url = MILVUS_BASE_URL + '/rag/kn/update_kb_name'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'old_kb_name': old_kb_name, 'new_kb_name': new_kb_name}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:  # 抛出报错
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = response.json()
        if final_response['code'] == 0:  # 正常获取到了结果
            return response_info
        else:  # 抛出报错
            return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def extract_minio_markdown_images(text: str) -> List[str]:
    """
    Extract the markdown images from the text. Only minio image urls.
    """
    # 匹配 markdown 图片语法中的 MinIO 下载链接
    # 模式说明: https?://[host]/minio/download/api/[bucket]/[object]
    pattern = r'!\[.*?\]\((https?://[^/]+/minio/download/api/[^)]+)\)'
    res = re.findall(pattern, text)
    return res


def get_extend_content_item(user_id, kb_info, knowledge_item, extend_num=1):
    """
        获取扩展上下文接口
    """
    file_name = knowledge_item["meta_data"]["file_name"]
    search_after = max(knowledge_item["meta_data"]["chunk_current_num"] - extend_num - 1, 0)
    page_size = 2*extend_num + 1
    res = get_milvus_file_content_list(user_id, kb_info, file_name, page_size, search_after)
    content_list = res["data"]["content_list"]
    extend_content = ""
    for item in content_list:
        extend_content += item["content"]
    # ====== 正常返回 =====
    knowledge_item["extend_content"] = extend_content
    return knowledge_item



if __name__ == "__main__":
    sub_chunk = []
    chunk = {
        "content": "空调 - 描述\n2. 详细描述\nA. 空调系统-IASC\n(3) 接口\nIASC通过ARINC429数据总线与DCU（31-41-05）通信。电源系统通过左直流汇流条（L DC BUS）为IASC1-通道A供电，通过左直流重要汇流条（L\nDC ESS BUS）为IASC1-通道B供电。\n电源系统通过右直流汇流条（R DC BUS）为IASC2-通道B供电，通过右直流重要汇流条（R\nDC ESS BUS）为IASC2-通道A供电。\n",
        "embedding_content": "空调 - 描述\n2. 详细描述\nA. 空调系统-IASC\n(3) 接口\nIASC通过ARINC429数据总线与DCU（31-41-05）通信。电源系统通过左直流汇流条（L DC BUS）为IASC1-通道A供电，通过左直流重要汇流条（L\nDC ESS BUS）为IASC1-通道B供电。\n电源系统通过右直流汇流条（R DC BUS）为IASC2-通道B供电，通过右直流重要汇流条（R\nDC ESS BUS）为IASC2-通道A供电。\n"
    }
    sub_chunk.append(chunk)
    userId = "18ef6f66-b82b-43d8-b934-d46b10acbecb",
    knowledgeBase = "8155ef14-80d4-4600-9b7e-6359a1fac98b"
    kb_id = "8155ef14-80d4-4600-9b7e-6359a1fac98b"
    add_file_name = "4-手册-SDS 21 空调 (2).pdf"
    add_file_path = "user_data/18ef6f66-b82b-43d8-b934-d46b10acbecb/8155ef14-80d4-4600-9b7e-6359a1fac98b/4-手册-SDS 21 空调 (2).pdf"
    kb_info = {"kb_name": knowledgeBase, "kb_id": kb_id}
    insert_milvus_result = add_milvus(userId, kb_info, sub_chunk, add_file_name, add_file_path)

    if insert_milvus_result['code'] != 0:
        print('失败')
    else:
        print('成功')
