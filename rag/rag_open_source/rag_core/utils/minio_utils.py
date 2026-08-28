from typing import List

from minio import Minio
import os
import re
import tempfile
import json
import time
import requests
from datetime import datetime, timedelta
import uuid
from pathlib import Path
from PIL import Image
# import oss_utils
import logging

logger = logging.getLogger(__name__)

from settings import MINIO_ADDRESS, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, SECURE
from settings import USE_OSS, BUCKET_NAME
from settings import MINIO_UPLOAD_BUCKET_NAME, REPLACE_MINIO_DOWNLOAD_URL
from model_manager.model_config import get_model_configure

try:
    GLOBAL_MINIO_CLIENT = Minio(
        MINIO_ADDRESS,
        access_key=MINIO_ACCESS_KEY,
        secret_key=MINIO_SECRET_KEY,
        secure=SECURE
    )
except Exception as e:
    logger.error(f"FATAL: Failed to initialize global Minio client: {e}")
    raise


max_retries = 3

def upload_local_file(file_path):
    """
    上传本地文件到 MinIO，并返回预签名的下载链接。

    :param file_path: 本地文件路径
    :return: 预签名的下载链接
    """
    bucket_name = MINIO_UPLOAD_BUCKET_NAME # 指定上传到的桶名
    # 获取文件名和扩展名
    _, filename_no_path = os.path.split(os.path.abspath(file_path))  # 提取文件名（包含后缀）
    base_filename, file_extension = os.path.splitext(filename_no_path)  # 分离文件名和后缀
    # # ======== 如果是图片，检验一下像素尺寸，太小的就返回空 ==========
    # if file_extension.lower() in ['.png', '.jpg', '.jpeg', '.bmp', '.gif']:
    #     try:
    #         img = Image.open(file_path)
    #         width, height = img.size
    #         if width < 50 or height < 50:
    #             return {"code": 0, 'message': '图片尺寸太小，download_link置为空，请注意', "download_link": ''}
    #     except Exception as e:
    #         logger.error(f"Error opening image file: {e}")  # 只记录报错也继续往下上传，等于不校验
    # # ======== 如果
    # 生成一个唯一的 UUID 作为临时文件名
    temp_file_name = str(uuid.uuid4())
    object_name = temp_file_name + file_extension  # 使用文件名作为对象名
    try:
        # # 检查桶是否存在，如果不存在则创建
        # if not minio_client.bucket_exists(bucket_name):
        #     minio_client.make_bucket(bucket_name)
        # 上传文件
        GLOBAL_MINIO_CLIENT.fput_object(bucket_name, object_name, file_path)
        logger.info(f"文件 {file_path} 已成功上传到 MinIO 桶 {bucket_name}，对象名 {object_name}")
        # # 生成预签名下载链接
        # presigned_url = minio_client.presigned_get_object(bucket_name, object_name, expires=timedelta(days=1))
        # print(f"预签名下载链接: {presigned_url}")
        # 直接拼接链接
        download_link = REPLACE_MINIO_DOWNLOAD_URL + '/' + bucket_name + '/' + object_name
        return {"code": 0, 'message': '成功', "download_link": download_link}
    except Exception as e:
        print(f"上传文件或生成预签名链接失败: {e}")
        return {"code": 1, 'message': f'Minio 上传失败{e}', "download_link": ''}


def remove_minio_file(object_name, bucket_name=None):
    """
    删除 MinIO 上的对象。

    :param object_name: 对象名（不含桶名）
    :param bucket_name: 桶名，默认取 MINIO_UPLOAD_BUCKET_NAME
    :return: True=删除成功，False=删除失败
    """
    bucket = bucket_name or MINIO_UPLOAD_BUCKET_NAME
    try:
        GLOBAL_MINIO_CLIENT.remove_object(bucket, object_name)
        logger.info(f"MinIO 删除成功: {bucket}/{object_name}")
        return True
    except Exception as e:
        logger.error(f"MinIO 删除失败: {bucket}/{object_name}, 错误: {e}")
        return False


def create_download_url(bucket_name, object_name, expire=timedelta(days=1)):
    """生成预签名下载链接"""
    # 生成预签名下载链接
    try:
        # 初始化 MinIO 客户端
        minio_client = Minio(
            MINIO_ADDRESS,
            access_key=MINIO_ACCESS_KEY,
            secret_key=MINIO_SECRET_KEY,
            secure=SECURE
        )
        presigned_url = minio_client.presigned_get_object(bucket_name, object_name, expires=expire)
        # 将内网 MINIO_ADDRESS 替换为对外暴露的 REPLACE_MINIO_DOWNLOAD_URL
        presigned_url = re.sub(
            r'https?://' + re.escape(MINIO_ADDRESS),
            REPLACE_MINIO_DOWNLOAD_URL.rstrip('/'),
            presigned_url,
            count=1
        )
        logger.info(f"{bucket_name},{object_name},预签名下载链接: {presigned_url}")
        return presigned_url
    except Exception as e:
        logger.info(f"{bucket_name},{object_name},生成预签名链接失败: {e}")
        return ""

def get_file_from_minio(object_name, download_path):
    stat = False
    download_link = ''
    """从 MinIO 获取文件并保存到本地"""
    retries = 0
    while retries < max_retries:
        try:
            minio_res = GLOBAL_MINIO_CLIENT.fget_object(BUCKET_NAME, object_name, download_path)
            logger.info(f'minio 下载到本地：{BUCKET_NAME},{object_name},{download_path}====mio_res：{minio_res}')
            # 检查文件是否存在
            if os.path.exists(download_path):
                # 文件大小检查（如果已知原始文件大小）
                original_size = minio_res.size  # 原始文件大小从返回处取
                local_size = os.path.getsize(download_path)
                while local_size < original_size:
                    logger.info(
                        f"{download_path},===== original_size:{original_size}- local_size:{local_size},文件大小不匹配，可能下载不完整")
                    local_size = os.path.getsize(download_path)
                    retries += 1
                    time.sleep(3)
                    if retries >= max_retries:  # 超过重试时间
                        break
                if local_size == original_size:
                    logger.info(
                        f"{download_path},===== original_size:{original_size}- local_size:{local_size},文件大小匹配，下载正确")
                    # ================ 检查文件大小完毕 ===============
                logger.info('文件已成功保存存在本地, 文件路径是：' + (download_path))
                stat = True
                download_link = f"{REPLACE_MINIO_DOWNLOAD_URL}/{BUCKET_NAME}/{object_name}"
                logger.info(repr(object_name) + ' minio文件下载成功')
                return stat, download_link
            else:  # 重试
                logger.info(download_path + ",文件在本地不存在，未保存成功")
                retries += 1
                time.sleep(3)
        except Exception as err:
            logger.info(repr(object_name) + ' minio文件下载失败，正在重试...错误：' + repr(err))
            retries += 1
            time.sleep(3)
    return stat, download_link

def get_file_size_from_minio(object_name):
    """从 MinIO 获取文件大小"""
    stat = False
    file_size = 0
    retries = 0
    while retries < max_retries:
        try:
            file_stat = GLOBAL_MINIO_CLIENT.stat_object(MINIO_UPLOAD_BUCKET_NAME, object_name)
            file_size = file_stat.size
            logger.info(f'{object_name} 文件大小：{file_size} bytes')
            stat = True
            return stat, file_size
        except Exception as err:
            logger.info(repr(object_name) + ' 获取文件大小失败，正在重试...错误：' + repr(err))
            retries += 1
            time.sleep(3)
    return stat, file_size

def check_files_size(file_urls: List[str], embedding_model_id: str) -> List[bool]:
    """检查 MinIO 中的文件大小是否与预期匹配"""
    emb_model_info = get_model_configure(embedding_model_id)
    if not emb_model_info.is_multimodal:
        raise ValueError(f"Model {emb_model_info.model_name} does not support multimodal.")

    max_image_size = emb_model_info.max_image_size
    result = []
    for url in file_urls:
        stat, file_size = get_file_size_from_minio(url.split('/')[-1])
        if not stat:
            raise RuntimeError(f"获取文件大小失败，file url = {url}")
        if file_size > max_image_size:
            logger.info(f"文件大小超过最大限制，file url = {url}， max_image_size = {max_image_size}")
            result.append(False)
        else:
            result.append(True)

    return result


def replace_minio_url(context: str, image_url_prefix: str = "") -> (str, list):
    """
    提取Markdown中的图片URL，下载到/tmp临时文件，并替换为本地路径
    只处理图片格式：jpg, jpeg, png, gif, bmp, webp, svg
    """
    ALLOWED_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.svg'}
    temp_files_to_cleanup = []  # 记录需要清理的临时文件
    # 提取Markdown图片URL正则：![alt](url)
    pattern = r"!\[.*?\]\((.*?)\)"
    matches = list(re.finditer(pattern, context))
    replace_info = []
    if not matches:
        return context, replace_info

    # 从后向前替换（避免替换后影响匹配位置）

    for match in reversed(matches):
        url = match.group(1)
        if url.startswith(image_url_prefix):
            ext = Path(url.split('?')[0].lower()).suffix  # 获取文件扩展名
            if ext not in ALLOWED_EXTENSIONS:
                continue
            # 检查URL是否是图片（通过扩展名初步判断）
            try:
                r = requests.get(url, timeout=10)
                if r.status_code == 200:
                    # 创建临时文件，确保在with块外操作已关闭的文件
                    tmp_fd, img_tmp_path = tempfile.mkstemp(suffix=ext)
                    temp_files_to_cleanup.append(img_tmp_path)
                    # 使用文件描述符写入，确保完全写入并关闭
                    with os.fdopen(tmp_fd, 'wb') as tmp_file:
                        tmp_file.write(r.content)
                    logger.info(f"已下载: {url[:60]}... -> {img_tmp_path}")
                    # 文件已关闭，现在可以安全上传
                    minio_result = upload_local_file(img_tmp_path)
                    if minio_result['code'] == 0:
                        placeholder = minio_result['download_link']
                        logger.info("====>image_download_link=%s" % placeholder)
                        context = context.replace(url, placeholder)
                        replace_info.append((url, placeholder))
                    else:
                        logger.error(f"====>get image_download_link err:{minio_result}")
                else:
                    logger.error(f"====>get image_download_link err:{r.text}")

            except Exception as e:
                logger.error(f"下载失败 {url[:50]}...: {e}")
                continue
    # 确保所有临时文件都被清理
    for tmp_file in temp_files_to_cleanup:
        try:
            if os.path.exists(tmp_file):
                os.remove(tmp_file)
        except Exception as e:
            logger.error(f"清理临时文件失败 {tmp_file}: {e}")
    return context, replace_info

