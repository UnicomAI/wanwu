import requests
import json
import time
import os
from logging_config import setup_logging
from settings import MQ_REL_URL, MQ_URL_URL, MQ_URLINSERT_URL, MQ_KB_STATUS_URL
logger_name='rag_mq_rel_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))


def update_doc_status(doc_id, status, meta_datas = [], api_url=MQ_REL_URL):
    """
    :param api_url: API的URL
    :param doc_id: 文档ID
    :param status: 文档状态 int
    :return: API响应的字典

    文档状态：
    10 文档添加成功
    31 文档查重完成
    32 文档下载完成
    33 文档解析完成
    34 文档添加向量库完成
    35 文档添加ES库完成
    
    51文档向量库重复查询校验失败
    52 文档已存在该知识库
    53文档下载失败
    54文档解析失败
    55 文档添加向量库失败
    56 文档添加ES库失败
    """
    
    # Request body
    # meta_data_list = [
    #     {"key": "release date", "value": "2025-08-07"}
    # ]
    meta_data_list = []
    for item in meta_datas:
        callback_data = {"key": item["key"], "value": item["value"], "metaId": item["meta_id"], "valueType": item["value_type"]}
        meta_data_list.append(callback_data)

    payload = {
        "id": doc_id,
        "status": status,
        "metaDataList": meta_data_list
    }
    # Request header
    headers = {'Content-Type': 'application/json'}
    retries=0
    max_retries = 5
    while retries < max_retries:
        try:
            # Send POST request
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info(f"---->回调Mass平台返回结果: {response}")
            logger.info('----->已回调Maas平台接口-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # Handle response
            if response.status_code == 200:
                response_data = response.json()
                return {"code": 0,"message":"文档状态更新回调成功"}
                
        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")  
        retries += 1  
        time.sleep(1)  # Wait for a period of time before retrying, for example 1 second
    return {"code": -1,"message":"文档状态更新回调错误"}



def update_url_status(id, status, fileSize=0, fileName='',api_url=MQ_URL_URL):
    """
    :param api_url: API的URL
    :param doc_id: 文档ID
    :param status: 文档状态 int
    :return: API响应的字典

    解析状态：
    10 解析成功
    57 解析失败

    """
    
    # Request body
    payload = {"id": id,"status": status,'fileSize':fileSize,'fileName':fileName}
    # Request header
    headers = {'Content-Type': 'application/json'}
    retries=0
    max_retries = 5
    while retries < max_retries:
        try:
            # Send POST request
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info('----->已回调Maas平台接口-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # Handle response
            if response.status_code == 200:
                response_data = response.json()
                return {"code": 0,"message":"解析状态更新回调成功"}
                
        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")  
        retries += 1  
        time.sleep(1)  # Wait for a period of time before retrying, for example 1 second
    return {"code": -1,"message":"解析状态更新回调错误"}



def update_urlinsert_status(doc_id, status, api_url=MQ_URLINSERT_URL):
    
    # Request body
    payload = {"id": doc_id,"status": status}
    # Request header
    headers = {'Content-Type': 'application/json'}
    retries=0
    max_retries = 5
    while retries < max_retries:
        try:
            # Send POST request
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info('----->已回调Maas平台接口url入库-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # Handle response
            if response.status_code == 200:
                response_data = response.json()
                return {"code": 0,"message":"文档状态更新回调成功"}
                
        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")  
        retries += 1  
        time.sleep(1)  # Wait for a period of time before retrying, for example 1 second
    return {"code": -1,"message":"文档状态更新回调错误"}


def update_kb_status(kb_id, status, api_url=MQ_KB_STATUS_URL):
    payload = {
        "knowledgeId": kb_id,
        "reportStatus": status
    }
    # Request header
    headers = {'Content-Type': 'application/json'}
    retries = 0
    max_retries = 5
    while retries < max_retries:
        try:
            # Send POST request
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info(f"---->回调Mass平台返回结果: {response}")
            logger.info('----->已回调Maas平台接口-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # Handle response
            if response.status_code == 200:
                return {"code": 0, "message": "文档状态更新回调成功"}

        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")
        retries += 1
        time.sleep(1)  # Wait for a period of time before retrying, for example 1 second
    return {"code": -1, "message": "文档状态更新回调错误"}
