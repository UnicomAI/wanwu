from minio import Minio
import os
import re
import tempfile
import json
import time
import requests
from datetime import datetime, timedelta
import uuid
# import oss_utils

from logging_config import setup_logging

logger_name = 'rag_minio_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

from settings import MINIO_ADDRESS, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, SECURE
from settings import USE_OSS, BUCKET_NAME
from settings import MINIO_UPLOAD_BUCKET_NAME, REPLACE_MINIO_DOWNLOAD_URL


max_retries = 3

def upload_local_file(file_path):
    """
    上传本地文件到 MinIO，并返回预签名的下载链接。

    :param file_path: 本地文件路径
    :return: 预签名的下载链接
    """
    bucket_name = MINIO_UPLOAD_BUCKET_NAME # 指定上传到的桶名 [EN] Specify the bucket name to upload to
    # 获取文件名和扩展名 [EN] Get file name and extension
    _, filename_no_path = os.path.split(os.path.abspath(file_path))  # 提取文件名（包含后缀） [EN] Extract file name (including suffix)
    base_filename, file_extension = os.path.splitext(filename_no_path)  # 分离文件名和后缀 [EN] Separate filename and suffix
    # 生成一个唯一的 UUID 作为临时文件名 [EN] Generate a unique UUID as a temporary file name
    temp_file_name = str(uuid.uuid4())
    object_name = temp_file_name + file_extension  # 使用文件名作为对象名 [EN] Use filename as object name
    try:
        # 初始化 MinIO 客户端 [EN] Initialize the MinIO client
        minio_client = Minio(
            MINIO_ADDRESS,
            access_key=MINIO_ACCESS_KEY,
            secret_key=MINIO_SECRET_KEY,
            secure=SECURE
        )
        # # 检查桶是否存在，如果不存在则创建 [EN] # Check if the bucket exists, create it if it does not exist
        # if not minio_client.bucket_exists(bucket_name):
        #     minio_client.make_bucket(bucket_name)
        # 上传文件 [EN] Upload files
        minio_client.fput_object(bucket_name, object_name, file_path)
        logger.info(f"文件 {file_path} 已成功上传到 MinIO 桶 {bucket_name}，对象名 {object_name}")
        # # 生成预签名下载链接 [EN] # Generate pre-signed download link
        # presigned_url = minio_client.presigned_get_object(bucket_name, object_name, expires=timedelta(days=1))
        # print(f"预签名下载链接: {presigned_url}") [EN] print(f"Pre-signed download link: {presigned_url}")
        # 直接拼接链接 [EN] Direct splicing link
        download_link = REPLACE_MINIO_DOWNLOAD_URL + '/' + bucket_name + '/' + object_name
        return {"code": 0, 'message': '成功', "download_link": download_link}
    except Exception as e:
        print(f"上传文件或生成预签名链接失败: {e}")
        return {"code": 1, 'message': f'Minio 上传失败{e}', "download_link": ''}


def craete_download_url(bucket_name, object_name, expire=timedelta(days=1)):
    """生成预签名下载链接"""
    # 生成预签名下载链接 [EN] Generate pre-signed download link
    try:
        # 初始化 MinIO 客户端 [EN] Initialize the MinIO client
        minio_client = Minio(
            MINIO_ADDRESS,
            access_key=MINIO_ACCESS_KEY,
            secret_key=MINIO_SECRET_KEY,
            secure=SECURE
        )
        presigned_url = minio_client.presigned_get_object(bucket_name, object_name, expires=expire)
        # 正则表达式匹配 https://ip:port/minio/download/api/ 部分 [EN] Regular expression matching https://ip:port/minio/download/api/ part
        pattern = r'http?://[^/]+/minio/download/api/'
        # 替换文本中的URL [EN] Replace URL in text
        presigned_url = re.sub(pattern, REPLACE_MINIO_DOWNLOAD_URL, presigned_url)
        logger.info(f"{bucket_name},{object_name},预签名下载链接: {presigned_url}")
        return presigned_url
    except Exception as e:
        logger.info(f"{bucket_name},{object_name},生成预签名链接失败: {e}")
        return ""

def get_file_from_minio(object_name, download_path):
    if USE_OSS:
        stat, download_link = oss_utils.get_file_from_oss(object_name, download_path)
        return stat, download_link
    else:
        # 初始化 MinIO 客户端 [EN] Initialize the MinIO client
        minio_client = Minio(
            MINIO_ADDRESS,
            access_key=MINIO_ACCESS_KEY,
            secret_key=MINIO_SECRET_KEY,
            secure=SECURE
        )
        stat = False
        download_link = ''
        """从 MinIO 获取文件并保存到本地"""
        retries = 0
        while retries < max_retries:
            try:
                minio_res = minio_client.fget_object(BUCKET_NAME, object_name, download_path)
                logger.info(f'minio 下载到本地：{BUCKET_NAME},{object_name},{download_path}====mio_res：{minio_res}')
                # 检查文件是否存在 [EN] Check if the file exists
                if os.path.exists(download_path):
                    # 文件大小检查（如果已知原始文件大小） [EN] File size check (if original file size is known)
                    original_size = minio_res.size  # 原始文件大小从返回处取 [EN] The original file size is taken from the return
                    local_size = os.path.getsize(download_path)
                    while local_size < original_size:
                        logger.info(
                            f"{download_path},===== original_size:{original_size}- local_size:{local_size},文件大小不匹配，可能下载不完整")
                        local_size = os.path.getsize(download_path)
                        retries += 1
                        time.sleep(3)
                        if retries >= max_retries:  # 超过重试时间 [EN] Retry time exceeded
                            break
                    if local_size == original_size:
                        logger.info(
                            f"{download_path},===== original_size:{original_size}- local_size:{local_size},文件大小匹配，下载正确")
                        # ================ 检查文件大小完毕 =============== [EN] ================ Checking file size completed ===============
                    logger.info('文件已成功保存存在本地, 文件路径是：' + (download_path))
                    stat = True
                    download_link = f"{REPLACE_MINIO_DOWNLOAD_URL}/{BUCKET_NAME}/{object_name}"
                    logger.info(repr(object_name) + ' minio文件下载成功')
                    return stat, download_link
                else:  # 重试 [EN] Try again
                    logger.info(download_path + ",文件在本地不存在，未保存成功")
                    retries += 1
                    time.sleep(3)
            except Exception as err:
                logger.info(repr(object_name) + ' minio文件下载失败，正在重试...错误：' + repr(err))
                retries += 1
                time.sleep(3)
        return stat, download_link

