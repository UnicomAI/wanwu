import os
import logging
import datetime
from logging.handlers import TimedRotatingFileHandler


###Convert to Beijing time
current_date = datetime.datetime.now()
# beijing_tz = datetime.timezone(datetime.timedelta(hours=8))
# beijing_time = date.astimezone(beijing_tz)
current_time = current_date.strftime('%Y%m%d')
#current_time = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")  

# Global variable definition
LOG_DIRECTORY = f'./logs/{current_time}'
LOG_LEVEL = logging.INFO
INTERVAL = 1  # Time interval for log rollback (rollback by week)
BACKUP_COUNT = 10  # Number of log files to keep

def setup_logging(app_name):
    """
    初始化日志配置。

    参数:
    app_name (str): 应用名称，用于日志文件命名
    """
    # Make sure the log directory exists
    if not os.path.exists(LOG_DIRECTORY):
        os.makedirs(LOG_DIRECTORY,exist_ok=True)
 
    # Define the full path of the log file, log file naming rule {app_name}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # Create a weekly rollback file log processor
    handler = TimedRotatingFileHandler(log_file_path, when='W0', interval=INTERVAL, backupCount=BACKUP_COUNT, encoding='utf-8')
    
    # Set log format
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    # Configure root logger
    logger = logging.getLogger()
    logger.setLevel(LOG_LEVEL)
    
    # Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()

    # Add log processor to logger
    logger.addHandler(handler)
