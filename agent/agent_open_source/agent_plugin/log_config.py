import os
import logging
import datetime
from logging.handlers import TimedRotatingFileHandler


###转化为北京时间 [EN] ##Convert to Beijing time
current_date = datetime.datetime.now()
# beijing_tz = datetime.timezone(datetime.timedelta(hours=8))
# beijing_time = date.astimezone(beijing_tz)
current_time = current_date.strftime('%Y%m%d')
#current_time = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")  

# 全局变量定义 [EN] Global variable definition
LOG_DIRECTORY = f'./logs/{current_time}'
LOG_LEVEL = logging.INFO
INTERVAL = 1  # 日志回滚的时间间隔（按周回滚） [EN] Time interval for log rollback (rollback by week)
BACKUP_COUNT = 10  # 保留的日志文件数量 [EN] Number of log files to keep

def setup_logging(app_name):
    """
    初始化日志配置。

    参数:
    app_name (str): 应用名称，用于日志文件命名
    """
    # 确保日志目录存在 [EN] Make sure the log directory exists
    if not os.path.exists(LOG_DIRECTORY):
        os.makedirs(LOG_DIRECTORY,exist_ok=True)
 
    # 定义日志文件的完整路径，日志文件命名规则 {app_name}.log [EN] Define the full path of the log file, log file naming rule {app_name}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # 创建一个按周回滚的文件日志处理器 [EN] Create a weekly rollback file log processor
    handler = TimedRotatingFileHandler(log_file_path, when='W0', interval=INTERVAL, backupCount=BACKUP_COUNT, encoding='utf-8')
    
    # 设置日志格式 [EN] Set log format
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    # 配置根日志记录器 [EN] Configure root logger
    logger = logging.getLogger()
    logger.setLevel(LOG_LEVEL)
    
    # 清除已存在的处理器，防止重复添加 [EN] Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()

    # 将日志处理器添加到日志记录器 [EN] Add log processor to logger
    logger.addHandler(handler)
