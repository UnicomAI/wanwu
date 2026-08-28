import os
import posixpath
import uuid

from configs.config import config
from extensions.minio import minio_client
from utils.log import logger


def upload_file_to_minio(
    file_stream, original_filename, overwrite_filename=None, bucket_name=None
):
    """
    上传文件到 MinIO 存储桶
    :param file_stream: 文件流
    :param original_filename: 原始文件名
    :param overwrite_filename: (可选) 重写保存的文件名
    :param bucket_name: (可选) 目标存储桶名称
    :return: 上传后的object_path (bucket_name/object_name)
    """
    if not bucket_name:
        bucket_name = config.callback_cfg["MINIO"]["BUCKET_NAME"]
    filename = original_filename
    if overwrite_filename:
        _, file_extension = os.path.splitext(original_filename)
        filename = overwrite_filename + file_extension
    object_name = posixpath.join(uuid.uuid4().hex[:16], filename)
    minio_client.create_public_bucket_if_not_exist(bucket_name)
    minio_client.put_object_from_stream(bucket_name, object_name, file_stream)
    object_path = posixpath.join(bucket_name, object_name)
    return object_path
