
from log.log_config import setup_logging
from settings import APP_NAME, LOGGER_NAME

# 设置日志,获取日志记录器 [EN] Set up logs and get loggers
logger = setup_logging(APP_NAME, LOGGER_NAME)