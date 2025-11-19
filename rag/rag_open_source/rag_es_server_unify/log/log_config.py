import os
import time
import glob
import logging
import sys
from logging.handlers import RotatingFileHandler

# 全局变量定义 [EN] Global variable definition
LOG_DIRECTORY = os.path.abspath(os.path.join(os.path.dirname(__file__), 'logs'))
LOG_LEVEL = logging.INFO
INTERVAL = 1  # 日志回滚的时间间隔 [EN] Log rollback interval
BACKUP_COUNT = 10  # 保留的日志文件数量 [EN] Number of log files to keep
def get_log_directory():
    """动态获取日志目录路径"""
    # 判断是否为打包环境 [EN] Determine whether it is a packaging environment
    if getattr(sys, 'frozen', False):
        # 打包环境：使用可执行文件所在目录 [EN] Packaging environment: use the directory where the executable file is located
        base_dir = os.path.dirname(sys.executable)
    else:
        # 开发环境：使用当前文件所在目录 [EN] Development environment: use the directory where the current file is located
        #base_dir = os.path.dirname(os.path.abspath(__file__))
        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

    # 在基础目录下创建 logs 子目录 [EN] Create a logs subdirectory under the base directory
    log_dir = os.path.join(base_dir, 'logs')
    os.makedirs(log_dir, exist_ok=True)
    return log_dir

# 定义一个函数来清理旧的日志文件 [EN] Define a function to clean old log files
def clean_old_logs():
    retention_days = 30
    # 确保日志目录存在 [EN] Make sure the log directory exists
    if not os.path.exists(LOG_DIRECTORY):
        return
    
    # 获取所有匹配的日志文件 [EN] Get all matching log files
    log_files = glob.glob(os.path.join(LOG_DIRECTORY, '*.*'))
    for log_file in log_files:
        try:
            # 获取文件的修改时间 [EN] Get the modification time of a file
            mod_time = os.path.getmtime(log_file)
            # 计算文件的最后修改时间距离现在的天数 [EN] Calculate the number of days since a file was last modified
            days_old = (time.time() - mod_time) / (24 * 3600)
            # 如果文件超过保留天数，则删除 [EN] Delete files if they exceed retention days
            if days_old > retention_days:
                os.remove(log_file)
                print(f"Deleted old log file: {log_file}")
        except Exception as e:
            print(f"Error deleting log file {log_file}: {str(e)}")

def setup_logging(app_name, logger_name):
    """
    初始化日志配置。
    """
    # 使用动态路径获取日志目录 [EN] Use dynamic path to get log directory
    LOG_DIRECTORY = get_log_directory()

    # 定义日志文件的完整路径，日志文件命名规则 {APP_NAME}_{port}.log [EN] Define the full path of the log file, the log file naming rule {APP_NAME}_{port}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # 创建logger [EN] Create logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # 确保日志目录存在 [EN] Make sure the log directory exists
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

    # 创建一个handler，用于写入日志文件 [EN] Create a handler for writing to the log file  
    file_handler = RotatingFileHandler(log_file_path, maxBytes=1024*1024*5, backupCount=5, encoding='utf-8') 
    file_handler.setLevel(logging.INFO)

    # 再创建一个handler，用于输出到控制台 [EN] Create another handler for output to the console  
    console_handler = logging.StreamHandler()  
    console_handler.setLevel(logging.INFO)

    # 定义handler的输出格式 [EN] Define the output format of the handler  
    formatter = logging.Formatter('%(asctime)s - %(filename)s:%(funcName)s:%(lineno)d - %(levelname)s - %(message)s',  datefmt='%Y-%m-%d %H:%M:%S')  
    # 设置时区为本地时间 [EN] Set time zone to local time
    formatter.converter = time.localtime  # 使用本地时区（默认） [EN] Use local time zone (default)
    
    file_handler.setFormatter(formatter)  
    console_handler.setFormatter(formatter) 
    
    # 清除已存在的处理器，防止重复添加 [EN] Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()

    # 将日志处理器添加到日志记录器 [EN] Add log processor to logger
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)  # 添加控制台输出处理器 [EN] Add console output handler

    return logger
