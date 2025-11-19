import os
import logging
import datetime
from logging.handlers import TimedRotatingFileHandler
from logging.handlers import RotatingFileHandler
import sys

###Convert to Beijing time
current_date = datetime.datetime.now()
current_time = current_date.strftime('%Y%m%d')

# Global variable definition
LOG_DIRECTORY = f'./logs'
LOG_LEVEL = logging.INFO
INTERVAL = 1 
BACKUP_COUNT = 10  # Number of log files to keep


def get_log_directory():
    """动态获取日志目录路径"""
    # Determine whether it is a packaging environment
    if getattr(sys, 'frozen', False):
        # Packaging environment: use the directory where the executable file is located
        base_dir = os.path.dirname(sys.executable)
    else:
        # Development environment: use the directory where the current file is located
        base_dir = os.path.dirname(os.path.abspath(__file__))

    # Create a logs subdirectory under the base directory
    log_dir = os.path.join(base_dir, 'logs')
    os.makedirs(log_dir, exist_ok=True)
    return log_dir

def setup_logging(app_name,logger_name):
    """
    初始化日志配置。

    参数:
    app_name (str): 应用名称，用于日志文件命名
    """
    # Use dynamic path to get log directory
    LOG_DIRECTORY = get_log_directory()

    # Define the full path of the log file, log file naming rule {app_name}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # Create logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # Make sure the log directory exists
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

    # Create a handler for writing to the log file
    # file_handler = TimedRotatingFileHandler(log_file_path, when='D', interval=INTERVAL, backupCount=BACKUP_COUNT, encoding='utf-8')
    file_handler = RotatingFileHandler(log_file_path, maxBytes=1024*1024*5, backupCount=5, encoding='utf-8') 
    file_handler.setLevel(logging.INFO)
  
    # Create another handler for output to the console
    console_handler = logging.StreamHandler()  
    console_handler.setLevel(logging.INFO)
    
    # Define the output format of the handler
    formatter = logging.Formatter('%(asctime)s - %(filename)s:%(funcName)s:%(lineno)d - %(levelname)s - %(message)s',  datefmt='%Y-%m-%d %H:%M:%S')   
    
    file_handler.setFormatter(formatter)  
    console_handler.setFormatter(formatter) 

    # Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()
        
    # Add handler to logger
    logger.addHandler(file_handler)  

    return logger
