import requests
import json
import uuid
import os
import logging

from settings import ES_BASE_URL, TIME_OUT

logger = logging.getLogger(__name__)


def add_file(user_id, kb_info, file_name, file_meta):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/add_file'
    headers = {'Content-Type': 'application/json'}

    req_data = {'user_id': user_id, 'kb_info': kb_info, 'file_name': file_name, 'file_meta': file_meta}

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + '新增文件请求结果：' + repr(response.text))
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


def allocate_chunks(user_id, kb_info, file_name, count, chunk_type="text"):
    url = ES_BASE_URL + '/api/v1/rag/es/allocate_chunks'
    headers = {'Content-Type': 'application/json'}

    req_data = {'user_id': user_id, 'kb_info': kb_info, 'file_name': file_name, 'count': count, "chunk_type":chunk_type}

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + 'allocate_chunks请求结果：' + repr(response.text))
        if response.status_code != 200:  # 抛出报错
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = response.json()
        return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def allocate_child_chunks(user_id, kb_info, file_name, chunk_id, count):
    url = ES_BASE_URL + '/api/v1/rag/es/allocate_child_chunks'
    headers = {'Content-Type': 'application/json'}

    req_data = {
        'user_id': user_id,
        'kb_info': kb_info,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'count': count
    }

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + 'allocate_child_chunks请求结果：' + repr(response.text))
        if response.status_code != 200:  # 抛出报错
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = response.json()
        return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def add_es(user_id, kb_info, docs, file_name):
    batch_size = 1000
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/bulk_add'
    headers = {'Content-Type': 'application/json'}

    kb_name = kb_info["kb_name"]
    batch_count = 0
    success_count = 0
    fail_count = 0
    error_reason = []

    for i in range(0, len(docs), batch_size):
        es_data = {}
        es_data['index_name'] = 'rag2_' + user_id + '_' + kb_name
        es_data['index_name'] = es_data['index_name'].lower()
        es_data['doc_list'] = []
        es_data['user_id'] = user_id
        es_data['kb_info'] = kb_info

        for doc in docs[i:i + batch_size]:
            chunk_dict = {
                "content_type": "text",
                "title": file_name,
                "source_type": "RAG_KB",
                "meta_data": doc["meta_data"]
            }
            # 普通文档切片
            if "text" in doc:
                chunk_dict["snippet"] = doc["text"]

            if "graph_data_text" in doc:  # 图谱数据切片
                chunk_dict["graph_data_text"] = doc["graph_data_text"]
                chunk_dict["graph_data"] = doc["graph_data"]

            # 图谱数据切片和社区报告
            if "chunk_type" in doc:
                chunk_dict["chunk_type"] = doc["chunk_type"]

            if "parent_text" in doc:
                 chunk_dict["parent_snippet"] = doc["parent_text"]
            if "is_parent" in doc:
                chunk_dict["is_parent"] = doc["is_parent"]

            es_data['doc_list'].append(chunk_dict)

        batch_count = batch_count + 1
        try:
            response = requests.post(url, headers=headers, json=es_data, timeout=TIME_OUT)
            logger.info(repr(file_name) + '分批写入es请求结果：' + repr(batch_count) + repr(response.text))
            if response.status_code == 200:
                result_data = response.json()
                if result_data['result']['success']:
                    success_count = success_count + 1
                    logger.info(repr(file_name) + "分批添加es请求成功")
                else:
                    fail_count = fail_count + 1
                    if str(result_data['result']['error']) not in error_reason: error_reason.append(
                        str(result_data['result']['error']))
                    logger.error(repr(file_name) + "分批添加es请求失败")
            else:
                logger.error(repr(file_name) + "分批添加es请求失败")
                fail_count = fail_count + 1
                if str(response.json()) not in error_reason: error_reason.append(
                    str(response.json()))

        except Exception as e:
            logger.error(repr(file_name) + "分批添加es请求异常: " + repr(e))
            fail_count = fail_count + 1
            if str(e) not in error_reason: error_reason.append(str(e))

    # print('add_es方法调用接口批量建库，总批次:%s次，成功:%s次,失败:%s次' % (batch_count, success_count, fail_count))
    logger.info('add_es方法调用接口批量建库')
    logger.info('总批次：' + repr(batch_count))
    logger.info('成功：' + repr(success_count))
    logger.info('失败：' + repr(fail_count))

    if batch_count == success_count:
        response_info['code'] = 0
        response_info['message'] = '成功'
    else:
        response_info['code'] = 0
        response_info['message'] = '部分文件添加es失败: ' + '/t'.join(error_reason)
    return response_info


def kb_rescore(query, weights, search_list, top_k):
    response_info = {"code": 0, "message": "", "data": {"sorted_scores": [], "sorted_search_list": []}}
    if len(search_list) <= 0:
        return response_info

    es_data = {
        "query": query,
        "weights": weights,
        "search_list_infos": search_list
    }

    es_url = ES_BASE_URL + "/api/v1/rag/es/rescore"
    headers = {'Content-Type': 'application/json'}
    try:
        if not search_list:
            return response_info
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code == 200:
            result_data = response.json()
            response_info["data"]["sorted_search_list"] = result_data['result']['search_list'][:top_k]
            response_info["data"]["sorted_scores"] = result_data['result']['scores'][:top_k]
            logger.info("query：" + repr(query) + ", es重评分请求成功")

            return response_info
        else:
            logger.error("query：" + repr(query) + ", es重评分请求失败" + repr(response.text))
            raise RuntimeError(repr(response.text))
    except Exception as e:
        logger.error(" query：" + repr(query) + ", es重评分请求异常：" + repr(e))
        return {"code": 1, "message": str(e)}


def search_es(user_id, kb_infos, query, top_k, filter_file_name_list=[], metadata_filtering_conditions = []):
    """
    :param kb_infos: [{"kb_name": "...", "kb_id": "..."}, ...]
    """
    search_list = []
    for kb_info in kb_infos:
        kb_name = kb_info["kb_name"]
        kb_id = kb_info["kb_id"]
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_info'] = kb_info
        es_data['query'] = query
        es_data['top_k'] = top_k
        es_data['min_score'] = 0
        es_data['filter_file_name_list'] = filter_file_name_list
        es_data['metadata_filtering_conditions'] = metadata_filtering_conditions
        es_url = ES_BASE_URL + "/api/v1/rag/es/search"
        headers = {'Content-Type': 'application/json'}
        try:
            response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
            if response.status_code == 200:
                tmp_sl = response.json()['result']['search_list']
                search_list.extend(tmp_sl)
                logger.info("知识库：" + repr(kb_name) + "es检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + "es检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + "es检索请求异常：" + repr(e))
    return search_list


def search_graph_es(user_id, kb_infos, query, top_k, filter_file_name_list=[]):
    """
    :param kb_infos: [{"kb_name": "...", "kb_id": "..."}, ...]
    """
    search_list = []
    for kb_info in kb_infos:
        kb_name = kb_info["kb_name"]
        kb_id = kb_info["kb_id"]
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_info'] = kb_info
        es_data['query'] = query
        es_data['top_k'] = top_k
        es_data['search_by'] = "graph_data_text"
        es_data['min_score'] = 0
        es_data['filter_file_name_list'] = filter_file_name_list[:10]  # 限制最多10个，以免掉蹦
        es_url = ES_BASE_URL + "/api/v1/rag/es/search"
        headers = {'Content-Type': 'application/json'}
        try:
            response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
            if response.status_code == 200:
                tmp_sl = response.json()['result']['search_list']
                search_list.extend(tmp_sl)
                logger.info("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求异常：" + repr(e))
    return search_list


def search_keyword(user_id, kb_infos, keywords, top_k, filter_file_name_list=[], metadata_filtering_conditions = []):
    """
    :param kb_infos: [{"kb_name": "...", "kb_id": "..."}, ...]
    """
    search_list = []
    for kb_info in kb_infos:
        kb_name = kb_info["kb_name"]
        kb_id = kb_info["kb_id"]
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_info'] = kb_info
        es_data['keywords'] = keywords
        es_data['top_k'] = top_k
        es_data['min_score'] = 0
        es_data['filter_file_name_list'] = filter_file_name_list
        es_data['metadata_filtering_conditions'] = metadata_filtering_conditions
        es_url = ES_BASE_URL + "/api/v1/rag/es/keyword_search"
        headers = {'Content-Type': 'application/json'}
        try:
            response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
            if response.status_code == 200:
                tmp_sl = response.json()['result']['search_list']
                for x in range(len(tmp_sl)):
                    tmp_sl[x]['kb_name'] = kb_name
                    tmp_sl[x]['kb_id'] = kb_id
                search_list = search_list + tmp_sl
                logger.info("知识库：" + repr(kb_name) + "es keyword 检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + "es keyword检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + "es keyword检索请求异常：" + repr(e))
    return search_list


def del_es_file(user_id, kb_info, file_name):
    response_info = {'code': 0, "message": "成功"}
    kb_name = kb_info["kb_name"]
    es_data = {}
    es_data['index_name'] = 'rag2_' + user_id + '_' + kb_name
    es_data['index_name'] = es_data['index_name'].lower()
    es_data['user_id'] = user_id
    es_data['kb_info'] = kb_info
    es_data['title'] = file_name
    es_url = ES_BASE_URL + "/api/v1/rag/es/delete_doc"
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code == 200:
            result_data = response.json()
            if result_data['result']['success']:
                logger.info("es删除文件请求成功")
                return response_info
            else:
                logger.error("es删除文件请求失败")
                error_msg = str(result_data['result']['error'])
                if 'no such index' in error_msg:
                    response_info['code'] = 0
                    response_info['message'] = kb_name + 'es知识库不存在'
                    return response_info
                else:
                    response_info['code'] = 1
                    response_info['message'] = error_msg
                    return response_info
        else:
            logger.error("es删除文件请求失败：" + repr(response.text))
            response_info['code'] = 1
            response_info['message'] = repr(response.text)
            return response_info
    except Exception as e:
        logger.error("es删除文件请求异常：" + repr(e))
        response_info['code'] = 1
        response_info['message'] = repr(e)
        return response_info


def del_es_kb(user_id, kb_info):
    response_info = {'code': 0, "message": "成功"}
    kb_name = kb_info["kb_name"]
    es_data = {}
    es_data['user_id'] = user_id
    es_data['kb_info'] = kb_info
    es_url = ES_BASE_URL + "/api/v1/rag/es/delete_index"
    headers = {'Content-Type': 'application/json'}
    response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
    try:
        if response.status_code == 200:
            result_data = response.json()
            if result_data['result']['success']:
                logger.info("es删除知识库请求成功")
                return response_info
            else:
                logger.error("es删除文件请求失败")
                error_msg = str(result_data['result']['error'])
                if 'no such index' in error_msg:
                    response_info['code'] = 0
                    response_info['message'] = kb_name + 'es知识库不存在'
                    return response_info
                else:
                    response_info['code'] = 1
                    response_info['message'] = error_msg
                    return response_info
        else:
            logger.error("es删除知识库请求失败：" + repr(response.text))
            response_info['code'] = 1
            response_info['message'] = repr(response.text)
            return response_info
    except Exception as e:
        logger.error("es删除知识库请求异常：" + repr(e))
        response_info['code'] = 1
        response_info['message'] = repr(e)
        return response_info


def add_es_bak(user_id, kb_name, docs, file_name):
    es_data = {}
    es_data['index_name'] = 'rag2_' + user_id + '_' + kb_name
    es_data['index_name'] = es_data['index_name'].lower()
    doc_list = []
    for doc in docs:
        es_file_path = file_name
        doc_list.append({
            "title": es_file_path,
            "snippet": doc["text"],
            "source_type": "RAG_KB",
            "meta_data": doc["meta_data"]
        })
    es_data['doc_list'] = doc_list
    es_url = ES_BASE_URL + "/api/v1/rag/es/bulk_add"
    headers = {'Content-Type': 'application/json'}
    print(es_data)
    try:
        print("es_data=%s" % json.dumps(es_data, ensure_ascii=False))
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code == 200:
            print("请求成功")
            print(response.json())  # 打印API返回的JSON数据
            return True
        else:
            print("请求失败")
            print(response.json())  # 打印错误信息
            return False
    except Exception as e:
        import traceback
        print("====> add_es error %s" % e)
        print(traceback.format_exc())
        return False

def init_qa_base(user_id, qa_info, embedding_model_id):
    response_info = {'code': 0, "message": '成功'}
    url = ES_BASE_URL + '/api/v1/rag/es/init_QA_base'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info,
        "embedding_model_id": embedding_model_id
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"es问答库初始化请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        init_response = response.json()
        if init_response['code'] != 0:
            logger.error(f"es问答库初始化请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {init_response}")
            raise RuntimeError(init_response['message'])

        logger.info("es问答库初始化请求成功")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"es问答库初始化请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info

def del_qa_base(user_id, qa_info):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/delete_QA_base'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"es问答库删除请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        del_response = response.json()
        if del_response['code'] != 0:
            logger.error(f"es问答库删除请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {del_response}")
            raise RuntimeError(del_response['message'])

        logger.info(f"es问答库删除请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"es问答库删除请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info


def del_qas(user_id, qa_info, qa_pair_ids):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/batch-delete-QAs'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info,
        "QAPairIds": qa_pair_ids
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"es删除问答对请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        del_response = response.json()
        if del_response['code'] != 0:
            logger.error(f"es删除问答对请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {del_response}")
            raise RuntimeError(del_response['message'])

        logger.info(f"es删除问答对请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"es删除问答对请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info

def add_qas(user_id, qa_info, qa_list):
    batch_size = 1000
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/add-QAs'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    qa_base_id = qa_info["QAId"]
    batch_count = 0
    qa_pair_ids = []

    try:
        for i in range(0, len(qa_list), batch_size):
            es_data = {"userId": user_id, "qa_info": qa_info, 'data': []}

            for qa in qa_list[i:i + batch_size]:
                qa_dict = {
                    "qa_pair_id": qa["qa_pair_id"],
                    "question": qa["question"],
                    "answer": qa["answer"],
                    "QABase": qa_base_name,
                    "QAId": qa_base_id,
                    "status": True
                }

                es_data['data'].append(qa_dict)
                qa_pair_ids.append(qa["qa_pair_id"])

            batch_count = batch_count + 1
            response = requests.post(url, headers=headers, json=es_data, timeout=TIME_OUT)
            logger.info('问答对分批写入es请求结果：' + repr(batch_count) + repr(response.text))
            if response.status_code != 200:
                logger.error(
                    f"问答对分批写入es请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
                raise RuntimeError(str(response.text))

            result_data = response.json()
            if result_data['code'] != 0:
                logger.error(
                    f"es问答库删除请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {result_data}")
                raise RuntimeError(result_data['message'])

            logger.info(f"问答对分批添加es请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}, batch_count: {batch_count}")

    except Exception as e:
        logger.error(f"问答对分批添加es请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        response_info['code'] = 1
        response_info['message'] = str(e)
        # 回滚
        del_res = del_qas(user_id, qa_info, qa_pair_ids)
        if del_res["code"] != 0:
            del_err_msg = del_res["message"]
            logger.error(f"问答对分批部分添加es失败后, 数据回滚也失败, user_id: {user_id}, qa_base_name: {qa_base_name}, "
                         f"qa_pair_ids: {qa_pair_ids}, error: {del_err_msg}")

    return response_info


def update_qa_data(user_id, qa_info, qa_pair_id, update_data):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/update_QA'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info,
        "QAPairId": qa_pair_id,
        "data": update_data
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"更新问答对请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        del_response = response.json()
        if del_response['code'] != 0:
            logger.error(f"es更新问答对请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {del_response}")
            raise RuntimeError(del_response['message'])

        logger.info(f"es更新问答对请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"es更新问答对请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info


def get_qa_list(user_id, qa_info, page_size, search_after):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/get_QA_list'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info,
        "page_size": page_size,
        "search_after": search_after
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"问答对分页请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        del_response = response.json()
        if del_response['code'] != 0:
            logger.error(f"问答对分页请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {del_response}")
            raise RuntimeError(del_response['message'])

        logger.info(f"问答对分页请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}")
        return del_response
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"问答对分页请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info

def update_qa_metas(user_id, qa_info, metas, update_type):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/update_QA_metas'
    headers = {'Content-Type': 'application/json'}

    qa_base_name = qa_info["QABase"]
    data = {
        "userId": user_id,
        "qa_info": qa_info,
        "metas": metas,
        "update_type": update_type
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'), timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"更新问答对元数据请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        del_response = response.json()
        if del_response['code'] != 0:
            logger.error(f"es更新问答对元数据请求失败, user_id: {user_id}, qa_base_name: {qa_base_name}, response: {del_response}")
            raise RuntimeError(del_response['message'])

        logger.info(f"es更新问答对元数据请求成功, user_id: {user_id}, qa_base_name: {qa_base_name}")
        return response_info
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"es更新问答对元数据请求异常, user_id: {user_id}, qa_base_name: {qa_base_name}, exception: {repr(e)}")
        return response_info

def vector_search(user_id, qa_infos, question, top_k, threshold=0.0, metadata_filtering_conditions = [], base_type="qa"):
    response_info = {'code': 0, "message": "成功", "data": {}}
    url = ES_BASE_URL + '/api/v1/rag/es/vector_search'
    headers = {'Content-Type': 'application/json'}

    data = {
        "userId": user_id,
        "qa_infos": qa_infos,
        "topk": top_k,
        "question": question,
        "threshold": threshold,
        "metadata_filtering_conditions": metadata_filtering_conditions,
        "base_type": base_type
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'),
                                 timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(
                f"问答对向量检索请求失败, user_id: {user_id}, qa_infos: {qa_infos}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        result_data = response.json()
        if result_data['code'] != 0:
            logger.error(
                f"问答对向量检索请求失败, user_id: {user_id}, qa_infos: {qa_infos}, response: {result_data}")
            raise RuntimeError(result_data['message'])

        logger.info(f"问答对向量检索请求成功, user_id: {user_id}, qa_infos: {qa_infos}")
        return result_data
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"问答对向量检索请求异常, user_id: {user_id}, qa_infos: {qa_infos}, exception: {repr(e)}")
        return response_info

def full_text_search(user_id, qa_infos, question, top_k, search_by = "question", threshold=0.0, metadata_filtering_conditions=[], base_type="qa"):
    response_info = {'code': 0, "message": "成功", "data": {}}
    url = ES_BASE_URL + '/api/v1/rag/es/text_search'
    headers = {'Content-Type': 'application/json'}

    data = {
        "userId": user_id,
        "qa_infos": qa_infos,
        "topk": top_k,
        "question": question,
        "threshold": threshold,
        "metadata_filtering_conditions": metadata_filtering_conditions,
        "base_type": base_type
    }
    try:
        response = requests.post(url, headers=headers, data=json.dumps(data, ensure_ascii=False).encode('utf-8'),
                                 timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(
                f"问答对全文检索请求失败, user_id: {user_id}, qa_infos: {qa_infos}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        result_data = response.json()
        if result_data['code'] != 0:
            logger.error(
                f"问答对全文检索请求失败, user_id: {user_id}, qa_infos: {qa_infos}, response: {result_data}")
            raise RuntimeError(result_data['message'])

        logger.info(f"问答对全文检索请求成功, user_id: {user_id}, qa_infos: {qa_infos}")
        return result_data
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"问答对全文检索请求异常, user_id: {user_id}, qa_infos: {qa_infos}, exception: {repr(e)}")
        return response_info

def qa_rescore(query, weights, top_k, search_list_infos):
    response_info = {'code': 0, "message": "成功", "data": {"sorted_scores":[], "sorted_search_list": []}}
    es_data = {
        "query": query,
        "search_list_infos": search_list_infos,
        "weights": weights
    }

    es_url = ES_BASE_URL + "/api/v1/rag/es/qa_rescore"
    headers = {'Content-Type': 'application/json'}

    try:
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code != 200:
            logger.error(f"问答对权重重排序请求失败, query: {query}, search_list_infos: {search_list_infos}, response: {repr(response.text)}")
            raise RuntimeError(str(response.text))

        result_data = response.json()
        if result_data['code'] != 0:
            logger.error(f"问答对权重重排序请求失败, query: {query}, search_list_infos: {search_list_infos}, response: {result_data}")
            raise RuntimeError(result_data['message'])

        sorted_search_list= result_data['data']['search_list'][:top_k]
        sorted_scores = result_data['data']['scores'][:top_k]
        logger.info(f"问答对权重重排序请求成功, query: {query}, sorted_search_list: {sorted_search_list}, sorted_scores: {sorted_scores}")

        return {
            "code": 0,
            "message": "",
            "data": {
                "sorted_scores": sorted_scores,
                "sorted_search_list": sorted_search_list
            }
        }
    except Exception as e:
        response_info['code'] = 1
        response_info['message'] = str(e)
        logger.error(f"答对权重重排序请求异常, query: {query}, exception: {repr(e)}")
        return response_info



if __name__ == '__main__':
    keywords = {"商飞测试": 100,"杭州": 10}
    result = search_keyword("1", ["gx_test"], keywords, 5, 0)
    print(result)



