import requests
import json
import time
import os
from settings import MQ_REL_URL, MQ_KB_STATUS_URL
import logging

logger = logging.getLogger(__name__)

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
    
    # 请求体
    # meta_data_list = [
    #     {"key": "发布日期", "value": "2025-08-07"}
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
    # 请求头
    headers = {'Content-Type': 'application/json'}
    retries=0
    max_retries = 5
    while retries < max_retries:
        try:
            # 发送POST请求
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info(f"---->回调Mass平台返回结果: {response}")
            logger.info('----->已回调Maas平台接口-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # 处理响应
            if response.status_code == 200:
                response_data = response.json()
                return {"code": 0,"message":"文档状态更新回调成功"}
                
        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")  
        retries += 1  
        time.sleep(1)  # 每次重试前等待一段时间，例如1秒  
    return {"code": -1,"message":"文档状态更新回调错误"}


def update_kb_status(kb_id, status, api_url=MQ_KB_STATUS_URL):
    payload = {
        "knowledgeId": kb_id,
        "reportStatus": status
    }
    # 请求头
    headers = {'Content-Type': 'application/json'}
    retries = 0
    max_retries = 5
    while retries < max_retries:
        try:
            # 发送POST请求
            response = requests.post(api_url, data=json.dumps(payload), headers=headers)
            logger.info(f"---->回调Mass平台返回结果: {response}")
            logger.info('----->已回调Maas平台接口-同步状态：' + repr(api_url) + repr(json.dumps(payload)))
            # 处理响应
            if response.status_code == 200:
                return {"code": 0, "message": "文档状态更新回调成功"}

        except Exception as e:
            logger.error(f"文档状态更新回调错误：{e}，正在重试...")
        retries += 1
        time.sleep(1)  # 每次重试前等待一段时间，例如1秒
    return {"code": -1, "message": "文档状态更新回调错误"}
