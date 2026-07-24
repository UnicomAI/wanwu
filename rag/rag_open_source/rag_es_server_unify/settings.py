import os
from config.config_parser import Config

config = Config('./config/config.ini')

#索引前缀
INDEX_NAME_PREFIX = config.get('DEFAULT', 'INDEX_NAME_PREFIX')  # 测试环境索引前缀
SNIPPET_INDEX_NAME_PREFIX = config.get('DEFAULT', 'SNIPPET_INDEX_NAME_PREFIX')  # 老的ES snippet 测试环境索引前缀
KBNAME_MAPPING_INDEX = config.get('DEFAULT', 'KBNAME_MAPPING_INDEX')  # userid 的所有 kb_name映射表

#日志名称
APP_NAME = config.get('DEFAULT', 'APP_NAME')  # 应用名称
LOGGER_NAME = config.get('DEFAULT', 'LOGGER_NAME')  # 日志器名称

# embedding
EMBEDDING_BATCH_SIZE = os.getenv("EMBEDDING_BATCH_SIZE", 10)

# ES
ES_HOSTS = [os.getenv("ES_HOSTS")]
ES_USER = os.getenv("ES_USER")
ES_PASSWORD = os.getenv("ES_PASSWORD")
ES_VERIFY_CERTS = os.getenv("ES_VERIFY_CERTS")
if ES_HOSTS is None or ES_USER is None or ES_PASSWORD is None:
    ES_HOSTS = [config.get('ES', 'Hosts')]
    ES_USER = config.get('ES', 'User')
    ES_PASSWORD = config.get('ES', 'Password')
if ES_VERIFY_CERTS is None:
    ES_VERIFY_CERTS = config.getboolean('ES', 'Verify_certs')
DELETE_BACTH_SIZE = config.getint('ES', 'DELETE_BACTH_SIZE')

# minio
MINIO_ADDRESS = os.getenv("MINIO_ADDRESS")
if MINIO_ADDRESS is None:
    MINIO_ADDRESS = config.getstr('MINIO', 'MINIO_ADDRESS')
REPLACE_MINIO_DOWNLOAD_URL = os.getenv("REPLACE_MINIO_DOWNLOAD_URL")
if not REPLACE_MINIO_DOWNLOAD_URL:
    raise ValueError("REPLACE_MINIO_DOWNLOAD_URL 环境变量未设置或为空，请检查部署配置")

#model
MODEL_PROVIDER_URL = os.getenv("MODEL_PROVIDER_URL")
MODEL_PROVIDER_ACCESS_TOKEN = ""
if MODEL_PROVIDER_URL is None:
    MODEL_PROVIDER_URL = config.getstr('MODEL_PROVIDER', 'MODEL_PROVIDER_URL')
    MODEL_PROVIDER_ACCESS_TOKEN = config.getstr('MODEL_PROVIDER', 'MODEL_PROVIDER_ACCESS_TOKEN')

