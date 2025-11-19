import requests
import json
import uuid
import os

from logging_config import setup_logging
from settings import ES_BASE_URL, TIME_OUT

logger_name = 'rag_es_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))


def add_file(user_id, kb_name, file_name, file_meta, kb_id=""):
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/add_file'
    headers = {'Content-Type': 'application/json'}

    req_data = {'user_id': user_id, 'kb_name': kb_name, 'kb_id': kb_id, 'file_name': file_name, 'file_meta': file_meta}

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + '新增文件请求结果：' + repr(response.text))
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


def allocate_chunks(user_id, kb_name, file_name, count, chunk_type="text", kb_id=""):
    url = ES_BASE_URL + '/api/v1/rag/es/allocate_chunks'
    headers = {'Content-Type': 'application/json'}

    req_data = {'user_id': user_id, 'kb_name': kb_name, 'kb_id': kb_id, 'file_name': file_name, 'count': count, "chunk_type":chunk_type}

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + 'allocate_chunks请求结果：' + repr(response.text))
        if response.status_code != 200:  # 抛出报错 [EN] Throw an error
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = json.loads(response.text)
        return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def allocate_child_chunks(user_id, kb_name, file_name, chunk_id, count, kb_id=""):
    url = ES_BASE_URL + '/api/v1/rag/es/allocate_child_chunks'
    headers = {'Content-Type': 'application/json'}

    req_data = {
        'user_id': user_id,
        'kb_name': kb_name,
        'kb_id': kb_id,
        'file_name': file_name,
        'chunk_id': chunk_id,
        'count': count
    }

    try:
        response = requests.post(url, headers=headers, json=req_data, timeout=TIME_OUT)
        logger.info(repr(file_name) + 'allocate_child_chunks请求结果：' + repr(response.text))
        if response.status_code != 200:  # 抛出报错 [EN] Throw an error
            err = str(response.text)
            return {'code': 1, "message": f"{err}"}
        final_response = json.loads(response.text)
        return final_response
    except Exception as e:
        return {'code': 1, "message": f"{e}"}


def add_es(user_id, kb_name, docs, file_name, kb_id=""):
    batch_size = 1000
    response_info = {'code': 0, "message": "成功"}
    url = ES_BASE_URL + '/api/v1/rag/es/bulk_add'
    headers = {'Content-Type': 'application/json'}

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
        es_data['kb_name'] = kb_name
        es_data['kb_id'] = kb_id

        for doc in docs[i:i + batch_size]:
            chunk_dict = {
                "title": file_name,
                "source_type": "RAG_KB",
                "meta_data": doc["meta_data"]
            }
            # 普通文档切片 [EN] Ordinary document slicing
            if "text" in doc:
                chunk_dict["snippet"] = doc["text"]

            if "graph_data_text" in doc:  # 图谱数据切片 [EN] Atlas data slices
                chunk_dict["graph_data_text"] = doc["graph_data_text"]
                chunk_dict["graph_data"] = doc["graph_data"]

            # 图谱数据切片和社区报告 [EN] Graph data slicing and community reporting
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
                result_data = json.loads(response.text)
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
                if str(json.loads(response.text)) not in error_reason: error_reason.append(
                    str(json.loads(response.text)))

        except Exception as e:
            logger.error(repr(file_name) + "分批添加es请求异常: " + repr(e))
            fail_count = fail_count + 1
            if str(e) not in error_reason: error_reason.append(str(e))

    # print('add_es方法调用接口批量建库，总批次:%s次，成功:%s次,失败:%s次' % (batch_count, success_count, fail_count)) [EN] print('add_es method calls interface to build database in batches, total batch: %s times, success: %s times, failure: %s times' % (batch_count, success_count, fail_count))
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


def get_weighted_rerank(user_id, kb_names, query, weights, milvus_list, es_list, top_k):
    raw_search_list = []
    tmp_content = []

    for i in milvus_list:
        if i["content"] in tmp_content: continue
        raw_search_list.append(
            {"title": i["file_name"], "snippet": i["content"], "kb_name": i["kb_name"], "content_id": i["content_id"],"meta_data": i["meta_data"]})
        tmp_content.append(i["content"])

    for i in es_list:
        if i["snippet"] in tmp_content: continue
        raw_search_list.append(i)
        tmp_content.append(i["snippet"])

    return combine_rescore_es(user_id, kb_names, query, weights, top_k, raw_search_list)


def combine_rescore_es(user_id, kb_names, query, weights, top_k, search_list = []):
    rescored_search_list = []
    sorted_score_list = []
    es_data = {}
    es_data['user_id'] = user_id
    es_data['query'] = query
    es_data['search_list'] = search_list
    es_data["weights"] = weights
    es_data["kb_names"] = kb_names
    es_url = ES_BASE_URL + "/api/v1/rag/es/rescore"
    headers = {'Content-Type': 'application/json'}
    try:
        if not search_list:
            return sorted_score_list, rescored_search_list
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            rescored_search_list = result_data['result']['search_list'][:top_k]
            sorted_score_list = result_data['result']['scores'][:top_k]
            logger.info("user_id：" + repr(user_id) + ", query：" + repr(query) + ", es重评分请求成功")
        else:
            logger.error("user_id：" + repr(user_id) + ", query：" + repr(query) + ", es重评分请求失败" + repr(response.text))
    except Exception as e:
        logger.error("user_id：" + repr(user_id) + ", query：" + repr(query) + ", es重评分请求异常：" + repr(e))
    return sorted_score_list, rescored_search_list


def search_es(user_id, kb_names, query, top_k, kb_ids=[], filter_file_name_list=[], metadata_filtering_conditions = []):
    search_list = []
    for kb_name in kb_names:
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_name'] = kb_name
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
                tmp_sl = json.loads(response.text)['result']['search_list']
                for x in range(len(tmp_sl)):
                    tmp_sl[x]['kb_name'] = kb_name
                search_list = search_list + tmp_sl
                logger.info("知识库：" + repr(kb_name) + "es检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + "es检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + "es检索请求异常：" + repr(e))
    return search_list


def search_graph_es(user_id, kb_names, query, top_k, kb_ids=[], filter_file_name_list=[]):
    search_list = []
    for kb_name in kb_names:
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_name'] = kb_name
        es_data['query'] = query
        es_data['top_k'] = top_k
        es_data['search_by'] = "graph_data_text"
        es_data['min_score'] = 0
        es_data['filter_file_name_list'] = filter_file_name_list[:10]  # 限制最多10个，以免掉蹦 [EN] Limit to 10 at most to avoid jumping
        es_url = ES_BASE_URL + "/api/v1/rag/es/search"
        headers = {'Content-Type': 'application/json'}
        try:
            response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
            if response.status_code == 200:
                tmp_sl = json.loads(response.text)['result']['search_list']
                for x in range(len(tmp_sl)):
                    tmp_sl[x]['kb_name'] = kb_name
                search_list = search_list + tmp_sl
                logger.info("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + f" query:{query}" + "graph_es检索请求异常：" + repr(e))
    return search_list


def search_keyword(user_id, kb_names, keywords, top_k, kb_ids=[], filter_file_name_list=[], metadata_filtering_conditions = []):
    search_list = []
    for kb_name in kb_names:
        es_data = {}
        es_data['user_id'] = user_id
        es_data['kb_name'] = kb_name
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
                tmp_sl = json.loads(response.text)['result']['search_list']
                for x in range(len(tmp_sl)):
                    tmp_sl[x]['kb_name'] = kb_name
                search_list = search_list + tmp_sl
                logger.info("知识库：" + repr(kb_name) + "es keyword 检索请求成功")
            else:
                logger.error("知识库：" + repr(kb_name) + "es keyword检索请求失败：" + repr(response.text))
        except Exception as e:
            logger.error("知识库：" + repr(kb_name) + "es keyword检索请求异常：" + repr(e))
    return search_list


def del_es_file(user_id, kb_name, file_name, kb_id=""):
    response_info = {'code': 0, "message": "成功"}
    es_data = {}
    es_data['index_name'] = 'rag2_' + user_id + '_' + kb_name
    es_data['index_name'] = es_data['index_name'].lower()
    es_data['user_id'] = user_id
    es_data['kb_name'] = kb_name
    es_data['title'] = file_name
    es_data['kb_id'] = kb_id
    es_url = ES_BASE_URL + "/api/v1/rag/es/delete_doc"
    headers = {'Content-Type': 'application/json'}
    try:
        response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
        if response.status_code == 200:
            result_data = json.loads(response.text)
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


def del_es_kb(user_id, kb_name, kb_id=""):
    response_info = {'code': 0, "message": "成功"}
    es_data = {}
    es_data['index_name'] = 'rag2_' + user_id + '_' + kb_name
    es_data['index_name'] = es_data['index_name'].lower()
    es_data['user_id'] = user_id
    es_data['kb_name'] = kb_name
    es_data['kb_id'] = kb_id
    es_url = ES_BASE_URL + "/api/v1/rag/es/delete_index"
    headers = {'Content-Type': 'application/json'}
    response = requests.post(es_url, headers=headers, json=es_data, timeout=TIME_OUT)
    try:
        if response.status_code == 200:
            result_data = json.loads(response.text)
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
            print(response.text)  # 打印API返回的JSON数据 [EN] Print the JSON data returned by the API
            return True
        else:
            print("请求失败")
            return False
            print(response.text)  # 打印错误信息 [EN] Print error message
    except Exception as e:
        import traceback
        print("====> add_es error %s" % e)
        print(traceback.format_exc())
        return False


if __name__ == '__main__':
    keywords = {"商飞测试": 100,"杭州": 10}
    result = search_keyword("1", ["gx_test"], keywords, 5, 0)
    print(result)



