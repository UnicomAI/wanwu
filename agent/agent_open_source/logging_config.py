import os
import logging
import datetime
from logging.handlers import TimedRotatingFileHandler
from logging.handlers import RotatingFileHandler


###转化为北京时间 [EN] ##Convert to Beijing time
current_date = datetime.datetime.now()
current_time = current_date.strftime('%Y%m%d')

# 全局变量定义 [EN] Global variable definition
LOG_DIRECTORY = f'./logs'
LOG_LEVEL = logging.INFO
INTERVAL = 1 
BACKUP_COUNT = 10  # 保留的日志文件数量 [EN] Number of log files to keep

def setup_logging(app_name,logger_name):
    """
    初始化日志配置。

    参数:
    app_name (str): 应用名称，用于日志文件命名
    """
    # 确保日志目录存在 [EN] Make sure the log directory exists
    if not os.path.exists(LOG_DIRECTORY):
        os.makedirs(LOG_DIRECTORY)
 
    # 定义日志文件的完整路径，日志文件命名规则 {app_name}.log [EN] Define the full path of the log file, log file naming rule {app_name}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # 创建logger [EN] Create logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # 创建一个handler，用于写入日志文件 [EN] Create a handler for writing to the log file  
    # file_handler = TimedRotatingFileHandler(log_file_path, when='D', interval=INTERVAL, backupCount=BACKUP_COUNT, encoding='utf-8')
    file_handler = RotatingFileHandler(log_file_path, maxBytes=1024*1024*5, backupCount=5, encoding='utf-8') 
    file_handler.setLevel(logging.INFO)
  
    # 再创建一个handler，用于输出到控制台 [EN] Create another handler for output to the console  
    console_handler = logging.StreamHandler()  
    console_handler.setLevel(logging.INFO)
    
    # 定义handler的输出格式 [EN] Define the output format of the handler  
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')  
    file_handler.setFormatter(formatter)  
    console_handler.setFormatter(formatter) 

    # 清除已存在的处理器，防止重复添加 [EN] Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()
        
    # 给logger添加handler [EN] Add handler to logger  
    logger.addHandler(file_handler)  

    return logger
