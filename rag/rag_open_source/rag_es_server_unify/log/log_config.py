import os
import time
import glob
import logging
import sys
from logging.handlers import RotatingFileHandler

# Global variable definition
LOG_DIRECTORY = os.path.abspath(os.path.join(os.path.dirname(__file__), 'logs'))
LOG_LEVEL = logging.INFO
INTERVAL = 1  # Log rollback interval
BACKUP_COUNT = 10  # Number of log files to keep
def get_log_directory():
    """动态获取日志目录路径"""
    # Determine whether it is a packaging environment
    if getattr(sys, 'frozen', False):
        # Packaging environment: use the directory where the executable file is located
        base_dir = os.path.dirname(sys.executable)
    else:
        # Development environment: use the directory where the current file is located
        #base_dir = os.path.dirname(os.path.abspath(__file__))
        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

    # Create a logs subdirectory under the base directory
    log_dir = os.path.join(base_dir, 'logs')
    os.makedirs(log_dir, exist_ok=True)
    return log_dir

# Define a function to clean old log files
def clean_old_logs():
    retention_days = 30
    # Make sure the log directory exists
    if not os.path.exists(LOG_DIRECTORY):
        return
    
    # Get all matching log files
    log_files = glob.glob(os.path.join(LOG_DIRECTORY, '*.*'))
    for log_file in log_files:
        try:
            # Get the modification time of a file
            mod_time = os.path.getmtime(log_file)
            # Calculate the number of days since a file was last modified
            days_old = (time.time() - mod_time) / (24 * 3600)
            # Delete files if they exceed retention days
            if days_old > retention_days:
                os.remove(log_file)
                print(f"Deleted old log file: {log_file}")
        except Exception as e:
            print(f"Error deleting log file {log_file}: {str(e)}")

def setup_logging(app_name, logger_name):
    """
    初始化日志配置。
    """
    # Use dynamic path to get log directory
    LOG_DIRECTORY = get_log_directory()

    # Define the full path of the log file, the log file naming rule {APP_NAME}_{port}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # Create logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # Make sure the log directory exists
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

    # Create a handler for writing to the log file
    file_handler = RotatingFileHandler(log_file_path, maxBytes=1024*1024*5, backupCount=5, encoding='utf-8') 
    file_handler.setLevel(logging.INFO)

    # Create another handler for output to the console
    console_handler = logging.StreamHandler()  
    console_handler.setLevel(logging.INFO)

    # Define the output format of the handler
    formatter = logging.Formatter('%(asctime)s - %(filename)s:%(funcName)s:%(lineno)d - %(levelname)s - %(message)s',  datefmt='%Y-%m-%d %H:%M:%S')  
    # Set time zone to local time
    formatter.converter = time.localtime  # Use local time zone (default)
    
    file_handler.setFormatter(formatter)  
    console_handler.setFormatter(formatter) 
    
    # Clear existing processors to prevent repeated additions
    if logger.hasHandlers():
        logger.handlers.clear()

    # Add log processor to logger
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)  # Add console output handler

    return logger
