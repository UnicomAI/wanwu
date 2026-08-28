import os
import time
import glob
import logging
import sys
from logging.handlers import RotatingFileHandler


class TraceIdFilter(logging.Filter):
    """从 Flask g 注入 trace_id / span_id 到每条 LogRecord。

    脱离 Flask 请求上下文（如 __main__、后台任务）时兜底为 "-"。
    """

    def filter(self, record):
        try:
            from flask import g, has_request_context
            if has_request_context():
                record.trace_id = g.get("trace_id", "-")
                record.span_id = g.get("span_id", "-")
            else:
                record.trace_id = "-"
                record.span_id = "-"
        except Exception:
            record.trace_id = "-"
            record.span_id = "-"
        return True


# 全局变量定义
LOG_DIRECTORY = os.path.abspath(os.path.join(os.path.dirname(__file__), 'logs'))
LOG_LEVEL = logging.INFO
INTERVAL = 1  # 日志回滚的时间间隔
BACKUP_COUNT = 10  # 保留的日志文件数量
def get_log_directory():
    """动态获取日志目录路径"""
    # 判断是否为打包环境
    if getattr(sys, 'frozen', False):
        # 打包环境：使用可执行文件所在目录
        base_dir = os.path.dirname(sys.executable)
    else:
        # 开发环境：使用当前文件所在目录
        #base_dir = os.path.dirname(os.path.abspath(__file__))
        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

    # 在基础目录下创建 logs 子目录
    log_dir = os.path.join(base_dir, 'logs')
    os.makedirs(log_dir, exist_ok=True)
    return log_dir

# 定义一个函数来清理旧的日志文件
def clean_old_logs():
    retention_days = 30
    # 确保日志目录存在
    if not os.path.exists(LOG_DIRECTORY):
        return
    
    # 获取所有匹配的日志文件
    log_files = glob.glob(os.path.join(LOG_DIRECTORY, '*.*'))
    for log_file in log_files:
        try:
            # 获取文件的修改时间
            mod_time = os.path.getmtime(log_file)
            # 计算文件的最后修改时间距离现在的天数
            days_old = (time.time() - mod_time) / (24 * 3600)
            # 如果文件超过保留天数，则删除
            if days_old > retention_days:
                os.remove(log_file)
                print(f"Deleted old log file: {log_file}")
        except Exception as e:
            print(f"Error deleting log file {log_file}: {str(e)}")

def setup_logging(app_name, logger_name):
    """
    初始化日志配置。
    """
    # 使用动态路径获取日志目录
    LOG_DIRECTORY = get_log_directory()

    # 定义日志文件的完整路径，日志文件命名规则 {APP_NAME}_{port}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # 创建logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # 确保日志目录存在
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

    # 创建一个handler，用于写入日志文件  
    file_handler = RotatingFileHandler(log_file_path, maxBytes=1024*1024*5, backupCount=5, encoding='utf-8') 
    file_handler.setLevel(logging.INFO)

    # 再创建一个handler，用于输出到控制台  
    console_handler = logging.StreamHandler()  
    console_handler.setLevel(logging.INFO)

    # 定义handler的输出格式
    formatter = logging.Formatter(
        '%(asctime)s - %(filename)s:%(funcName)s:%(lineno)d - %(levelname)s - trace_id=%(trace_id)s span_id=%(span_id)s - %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
    )
    # 设置时区为本地时间
    formatter.converter = time.localtime  # 使用本地时区（默认）

    trace_filter = TraceIdFilter()
    file_handler.addFilter(trace_filter)
    console_handler.addFilter(trace_filter)

    file_handler.setFormatter(formatter)
    console_handler.setFormatter(formatter)
    
    # 清除已存在的处理器，防止重复添加
    if logger.hasHandlers():
        logger.handlers.clear()

    # 将日志处理器添加到日志记录器
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)  # 添加控制台输出处理器

    return logger
