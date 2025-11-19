from collections import defaultdict
import requests
import json
import uuid
import os
import threading
from threading import Thread

from logging_config import setup_logging
from settings import MILVUS_BASE_URL, TIME_OUT

logger_name = 'rag_milvus_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))


def make_request(url: str, data: dict):
    response_info = {'code': 0, "message": "成功"}
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            return response_info
        final_response = json.loads(response.text)
        if final_response['code'] != 0:
            response_info['code'] = final_response['code']
            response_info['message'] = final_response['message']
            return response_info
        # ======== 正常返回 ======= [EN] ======== Normal return =======
        return final_response
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        return response_info

def generate_chunks_bacth(chunks: list, batch_size=1000):
    """ 将chunks 按chunk_current_num分组并生成批次数据，每个list的长度为batch_size"""
    batch_data = []
    # 使用defaultdict来聚合数据 [EN] Use defaultdict to aggregate data
    aggregated_data = defaultdict(list)
    # 遍历列表，按照chunk_current_num字段聚合 [EN] Traverse the list and aggregate according to the chunk_current_num field
    temp_num = 0
    for item in chunks:
        temp_num += 1
        if item['meta_data'].get('chunk_current_num', -1) == -1:
            item['meta_data']['chunk_current_num'] = temp_num // 100
        chunk_current_num = item['meta_data']['chunk_current_num']
        aggregated_data[chunk_current_num].append(item)
    # 将聚合后的数据转换为普通字典，以便查看 [EN] Convert aggregated data to a normal dictionary for viewing
    aggregated_data = dict(aggregated_data)
    print(aggregated_data)
    for key, value in aggregated_data.items():
        batch_data.extend(value)
        if len(batch_data) >= batch_size:
            yield batch_data
            batch_data = []
    # 最后一个batch [EN] The last batch
    if batch_data:
        yield batch_data


def init_knowledge_base(user_id, kb_name, kb_id="", embedding_model_id="", enable_knowledge_graph = False):
    response_info = {'code': 0, "message": '成功'}
    url = MILVUS_BASE_URL + '/rag/kn/init_kb'
    headers = {'Content-Type': 'application/json'}
    if not kb_id:
        kb_id = str(uuid.uuid4())
    data = {
        "userId": user_id,
        "kb_name": kb_name,
        "kb_id": kb_id,
        "embedding_model_id": embedding_model_id,
        "enable_knowledge_graph": enable_knowledge_graph
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus初始化请求失败：" + repr(response.text))
            return response_info

        init_response = json.loads(response.text)
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
        result_data = json.loads(response.text)
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


def list_knowledge_file(user_id, kb_name, kb_id=""):
    response_info = {'code': 0, "message": "成功", "data": {"knowledge_file_names": []}}
    # url='http://localhost:6098/list_file_names'
    url = MILVUS_BASE_URL + '/rag/kn/list_file_names'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_name': kb_name, 'kb_id': kb_id}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus查询知识库所有文档请求失败：" + repr(response.text))
            return response_info
        result_data = json.loads(response.text)
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


def list_knowledge_file_download_link(user_id, kb_name, kb_id=""):
    response_info = {'code': 0, "message": "成功", "data": {"file_download_links": []}}
    # url='http://localhost:6098/list_file_names'
    url = MILVUS_BASE_URL + '/rag/kn/list_file_download_links'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_name': kb_name, 'kb_id': kb_id}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus获取知识库里所有文档的下载链接请求失败：" + repr(response.text))
            return response_info
        result_data = json.loads(response.text)
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
def search_milvus(user_id, kb_names, top_k, question, threshold, search_field, emb_model="bge", kb_ids=[],
                  filter_file_name_list=[], metadata_filtering_conditions = [], milvus_url = KNN_SEARCH_URL):
    """
    :param emb_model:  "bge", "bce", "conna"
    """
    post_data = {}
    post_data["userId"] = user_id
    post_data["kb_names"] = kb_names
    post_data["topk"] = top_k * 4
    post_data["question"] = question
    post_data["threshold"] = threshold
    post_data["emb_model"] = emb_model
    post_data["kb_ids"] = kb_ids
    post_data["filter_file_name_list"] = filter_file_name_list
    post_data["metadata_filtering_conditions"] = metadata_filtering_conditions

    response_info = {'code': 0, "message": "成功", "data": {"prompt": "", "search_list": []}}
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(milvus_url, headers=headers, data=json.dumps(post_data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus问题检索请求失败：" + repr(response.text))
            return response_info
        result_data = json.loads(response.text)
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

def add_milvus(user_id, kb_name, sub_chunk, add_file_name, add_file_path, kb_id="", milvus_url = ADD_URL):
    batch_size = 200
    response_info = {'code': 0, "message": "成功"}
    batch_count = 0
    success_count = 0
    fail_count = 0
    error_reason = []
    # sub_chunk 批次生成器,按 按chunk_current_num分组并生成批次数据 [EN] sub_chunk batch generator, group by chunk_current_num and generate batch data
    chunk_gen = generate_chunks_bacth(sub_chunk, batch_size=batch_size)
    for batch in chunk_gen:
        insert_data = {}
        insert_data['userId'] = user_id
        insert_data['kb_name'] = kb_name
        insert_data['kb_id'] = kb_id
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

            if "create_time" in chunk:
                chunk_dict["create_time"] = chunk["create_time"]

            if "is_parent" in chunk:
                chunk_dict["is_parent"] = chunk["is_parent"]

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
                # ========= 报错直接返回结束 ======= [EN] ========= Report an error and return directly to the end =======
                response_info['code'] = 1
                response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
                return response_info

            result_data = json.loads(response.text)
            if result_data['code'] != 0:
                fail_count = fail_count + 1
                if str(result_data['message']) not in error_reason: error_reason.append(str(result_data['message']))
                logger.error(repr(add_file_name) + repr(batch_count) + '批量写入milvus请求失败')
                # ========= 报错直接返回结束 ======= [EN] ========= Report an error and return directly to the end =======
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
            # ========= 报错直接返回结束 ======= [EN] ========= Report an error and return directly to the end =======
            response_info['code'] = 1
            response_info['message'] = '部分文件添加milvus失败: ' + '/t'.join(error_reason)
            return response_info


    # print('add_milvus方法调用接口批量建库，总批次:%s次，成功:%s次,失败:%s次' % (batch_count, success_count, fail_count)) [EN] print('add_milvus method calls interface to build database in batches, total batch: %s times, success: %s times, failure: %s times' % (batch_count, success_count, fail_count))
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


def del_milvus_kbs(user_id, kb_name, kb_id):
    response_info = {'code': 0, "message": "成功"}
    url = MILVUS_BASE_URL + '/rag/kn/del_kb'
    headers = {'Content-Type': 'application/json', }
    data = {'userId': user_id, 'kb_name': kb_name, 'kb_id': kb_id}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus删除知识库请求失败: " + repr(response.text))
            return response_info
        del_response = json.loads(response.text)
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


def del_milvus_files(user_id, kb_name, file_names, kb_id=""):
    response_info = {'code': 0, "message": "成功"}
    url = MILVUS_BASE_URL + '/rag/kn/del_files'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'kb_name': kb_name, 'file_names': file_names, 'kb_id': kb_id}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            response_info['code'] = 1
            response_info['message'] = str(response.text)
            logger.error("milvus删除知识库文档请求失败: " + repr(response.text))
            return response_info
        del_files_response = json.loads(response.text)
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


def get_milvus_file_content_list(user_id: str, kb_name: str, file_name: str, page_size: int,
                                 search_after: int, kb_id="", content_type="text"):
    """
        获取知识库文件片段列表,用于分页展示
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_content_list'
    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'file_name': file_name,
        'page_size': page_size,
        'search_after': search_after,
        'kb_id': kb_id,
        "content_type": content_type
    }

    return make_request(url, data)


def get_milvus_file_child_content_list(user_id: str, kb_name: str, file_name: str, chunk_id: int, kb_id=""):
    """
        获取知识库文件子片段列表
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_child_content_list'
    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'kb_id': kb_id
    }

    return make_request(url, data)


def list_file_names_after_filtering(user_id, kb_name, filtering_conditions, kb_id=""):
    """
        根据file_name更新知识库文件元数据
    """
    url = MILVUS_BASE_URL + '/rag/kn/list_file_names_after_filtering'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'filtering_conditions': filtering_conditions,
        'kb_id': kb_id
    }

    return make_request(url, data)


def update_child_chunk(user_id, kb_name, chunk_id, chunk_current_num, child_chunk, kb_id=""):
    """
        更新知识库子段
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_child_chunk'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'chunk_id': chunk_id,
        'chunk_current_num': chunk_current_num,
        'child_chunk': child_chunk
    }

    return make_request(url, data)


def update_file_metas(user_id, kb_name, update_datas, kb_id=""):
    """
        更新知识库元数据
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_file_metas'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'update_datas': update_datas,
        'kb_id': kb_id
    }

    return make_request(url, data)


def update_chunk_labels(user_id, kb_name, file_name, chunk_id, labels, kb_id=""):
    """
        根据file_name和chunk_id更新标签
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_chunk_labels'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'labels': labels,
        'kb_id': kb_id
    }

    return make_request(url, data)


def get_content_by_ids(user_id, kb_name, content_ids, content_type= "text", kb_id=""):
    """
        根据file_name和chunk_id获取分段信息
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_content_by_ids'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'content_ids': content_ids,
        'kb_id': kb_id,
        "content_type": content_type
    }

    return make_request(url, data)

def batch_delete_chunks(user_id, kb_name, file_name, chunk_ids, kb_id=""):
    """
        根据file_name和chunk_ids删除分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/batch_delete_chunks'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'file_name': file_name,
        'chunk_ids': chunk_ids,
        'kb_id': kb_id
    }

    return make_request(url, data)


def batch_delete_child_chunks(user_id, kb_name, file_name, chunk_id, chunk_current_num,
                        child_chunk_current_nums, kb_id=""):
    """
        根据chunk_id和child_chunk_current_nums删除子分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/delete_child_chunks'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'kb_id': kb_id,
        "chunk_current_num": chunk_current_num,
        "child_chunk_current_nums": child_chunk_current_nums
    }

    return make_request(url, data)


def update_milvus_content_status(user_id: str, kb_name: str, file_name: str, content_id: str, status: bool,
                                 on_off_switch=None, kb_id=""):
    """
        根据content_id更新知识库文件片段状态
    """
    url = MILVUS_BASE_URL + '/rag/kn/update_content_status'
    if on_off_switch in [True, False]:  # 前端传递了 on_off_switch 参数 [EN] The front end is passed the on_off_switch parameter
        data = {
            'userId': user_id,
            'kb_name': kb_name,
            'file_name': file_name,
            'content_id': content_id,
            'status': status,
            'on_off_switch': on_off_switch,
            'kb_id': kb_id
        }
    else:
        data = {
            'userId': user_id,
            'kb_name': kb_name,
            'file_name': file_name,
            'content_id': content_id,
            'status': status,
            'kb_id': kb_id
        }

    return make_request(url, data)


def del_community_reports(user_id, kb_name, clear_reports=False, content_ids= [], kb_id=""):
    """
        根据chunk_id和child_chunk_current_nums删除子分段
    """
    url = MILVUS_BASE_URL + '/rag/kn/del_community_reports'

    data = {
        'userId': user_id,
        'kb_name': kb_name,
        'kb_id': kb_id,
        "clear_reports": clear_reports,
        "content_ids": content_ids
    }

    return make_request(url, data)


def get_kb_info(user_id, kb_name):
    error_message = ""
    url = MILVUS_BASE_URL + '/rag/kn/get_kb_info'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, "kb_name": kb_name}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            error_message = str(response.text)
        else:
            result_data = json.loads(response.text)
            if result_data['code'] != 0:
                error_message = result_data['message']
            else:
                logger.info("milvus查询知识库graph状态请求成功")
                return result_data['data']["kb_info"]
    except Exception as e:
        error_message = str(e)

    logger.error(f"milvus查询知识库graph状态请求失败：{error_message}")
    return {}


def get_milvus_content_status(user_id: str, kb_name: str, content_id_list: list):
    """
        获取文本分块状态用于进行检索后过滤。
    """
    response_info = {'code': 0, "message": "成功"}
    # url = "http://localhost:30041/rag/kn/get_useful_content_status"  # 临时地址 [EN] url = "http://localhost:30041/rag/kn/get_useful_content_status" # Temporary address
    url = MILVUS_BASE_URL + '/rag/kn/get_useful_content_status'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'kb_name': kb_name, 'content_id_list': content_id_list}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:  # 抛出报错 [EN] Throw an error
            err = str(response.text)
            raise RuntimeError(f"{kb_name}-{content_id_list},Error get_milvus_content_status: {err}")
        final_response = json.loads(response.text)
        if final_response['code'] == 0:  # 正常获取到了结果 [EN] The results were obtained normally
            res_list = final_response['data']["useful_content_id_list"]
            return res_list
        else:  # 抛出报错 [EN] Throw an error
            raise RuntimeError(
                f"{kb_name}-{content_id_list},Error get_milvus_content_status: {final_response}")
    except Exception as e:
        raise RuntimeError(f"{e}")


def get_milvus_kb_name_id(user_id: str, kb_name: str):
    """
        获取某个知识库映射的 kb_id接口
    """
    url = MILVUS_BASE_URL + '/rag/kn/get_kb_id'
    headers = {'Content-Type': 'application/json'}
    data = {'userId': user_id, 'kb_name': kb_name}
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:  # 抛出报错 [EN] Throw an error
            err = str(response.text)
            raise RuntimeError(f"{kb_name}-,Error get_kb_name_id: {err}")
        final_response = json.loads(response.text)
        if final_response['code'] == 0:  # 正常获取到了结果 [EN] The results were obtained normally
            kb_id = final_response['data']["kb_id"]
            return kb_id
        else:  # 抛出报错 [EN] Throw an error
            raise RuntimeError(
                f"{kb_name},Error get_kb_name_id: {final_response}")
    except Exception as e:
        raise RuntimeError(f"Error get_kb_name_id:{e}")


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
        if response.status_code != 200:  # 抛出报错 [EN] Throw an error
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = json.loads(response.text)
        if final_response['code'] == 0:  # 正常获取到了结果 [EN] The results were obtained normally
            return response_info
        else:  # 抛出报错 [EN] Throw an error
            return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def get_extend_content_item(user_id, kb_name, knowledge_item, extend_num=1):
    """
        获取扩展上下文接口
    """
    file_name = knowledge_item["meta_data"]["file_name"]
    search_after = max(knowledge_item["meta_data"]["chunk_current_num"] - extend_num - 1, 0)
    page_size = 2*extend_num + 1
    res = get_milvus_file_content_list(user_id, kb_name, file_name, page_size, search_after)
    content_list = res["data"]["content_list"]
    extend_content = ""
    for item in content_list:
        extend_content += item["content"]
    # ====== 正常返回 ===== [EN] ====== Normal return =====
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
    add_file_name = "4-手册-SDS 21 空调 (2).pdf"
    add_file_path = "user_data/18ef6f66-b82b-43d8-b934-d46b10acbecb/8155ef14-80d4-4600-9b7e-6359a1fac98b/4-手册-SDS 21 空调 (2).pdf"
    insert_milvus_result = add_milvus(userId, knowledgeBase, sub_chunk, add_file_name, add_file_path)

    if insert_milvus_result['code'] != 0:
        print('失败')
    else:
        print('成功')
