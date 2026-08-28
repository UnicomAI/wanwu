import os
import logging
import datetime
from logging.handlers import TimedRotatingFileHandler
from logging.handlers import RotatingFileHandler
import sys

from opentelemetry import trace as otel_trace

###转化为北京时间
current_date = datetime.datetime.now()
current_time = current_date.strftime('%Y%m%d')

# 全局变量定义
LOG_DIRECTORY = f'./logs'
LOG_LEVEL = logging.INFO
INTERVAL = 1
LOG_FILE_MAX_SIZE = 20
LOG_FILE_BACKUP_COUNT = 10
LOG_FORMAT = '%(asctime)s - %(name)s - %(levelname)s - trace_id=%(trace_id)s span_id=%(span_id)s - %(message)s'  # 日志格式
LOG_DATEFORMAT = '%Y-%m-%d %H:%M:%S'  # 日期格式


class TraceIdFilter(logging.Filter):
    """从 OpenTelemetry 当前 Span 提取 trace_id / span_id 注入 LogRecord。

    一份代码同时适配 Flask(run.py) / FastAPI(know_sse.py) / Kafka worker，
    所有入口都会通过 Instrumentor 或手动 start_as_current_span 让当前 span 可读。
    无 span（启动期日志、入口外调用）时兜底为 "-"。
    """

    def filter(self, record):
        try:
            span = otel_trace.get_current_span()
            if span and span.is_recording():
                ctx = span.get_span_context()
                record.trace_id = format(ctx.trace_id, "032x")
                record.span_id = format(ctx.span_id, "016x")
            else:
                record.trace_id = "-"
                record.span_id = "-"
        except Exception:
            record.trace_id = "-"
            record.span_id = "-"
        return True


def get_log_directory():
    """动态获取日志目录路径"""
    # 判断是否为打包环境
    if getattr(sys, 'frozen', False):
        # 打包环境：使用可执行文件所在目录
        base_dir = os.path.dirname(sys.executable)
    else:
        # 开发环境：使用当前文件所在目录
        base_dir = os.path.dirname(os.path.abspath(__file__))

    # 在基础目录下创建 logs 子目录
    log_dir = os.path.join(base_dir, 'logs')
    os.makedirs(log_dir, exist_ok=True)
    return log_dir

def setup_logging(app_name,logger_name):
    """
    初始化日志配置。

    参数:
    app_name (str): 应用名称，用于日志文件命名
    """
    # 使用动态路径获取日志目录
    LOG_DIRECTORY = get_log_directory()

    # 定义日志文件的完整路径，日志文件命名规则 {app_name}.log
    log_file_path = os.path.join(LOG_DIRECTORY, f'{app_name}.log')

    # 创建logger
    logger = logging.getLogger(logger_name)
    logger.setLevel(LOG_LEVEL)

    # 确保日志目录存在
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

    # 创建一个handler，用于写入日志文件  
    # file_handler = TimedRotatingFileHandler(log_file_path, when='D', interval=INTERVAL, backupCount=BACKUP_COUNT, encoding='utf-8')
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

    trace_filter = TraceIdFilter()
    file_handler.addFilter(trace_filter)
    console_handler.addFilter(trace_filter)

    file_handler.setFormatter(formatter)
    console_handler.setFormatter(formatter)

    # 清除已存在的处理器，防止重复添加
    if logger.hasHandlers():
        logger.handlers.clear()

    # 给logger添加handler
    logger.addHandler(file_handler)

    return logger


def init_logging():
    log_handlers = []
    # 强制重配置 stdout 为行缓冲模式，确保每一行日志都能即时输出到 Docker Console
    if hasattr(sys.stdout, 'reconfigure'):
        sys.stdout.reconfigure(line_buffering=True)

    log_file = os.getenv("LOG_FILE")

    if log_file:
        # 使用动态路径获取日志目录
        LOG_DIRECTORY = get_log_directory()
        os.makedirs(LOG_DIRECTORY, exist_ok=True)
        # 定义日志文件的完整路径，日志文件命名规则 {app_name}.log
        pid = os.getpid()
        unique_log_file = os.path.join(LOG_DIRECTORY, f'{log_file}.log')

        log_handlers.append(
            RotatingFileHandler(
                filename=unique_log_file,
                maxBytes=LOG_FILE_MAX_SIZE * 1024 * 1024,
                backupCount=LOG_FILE_BACKUP_COUNT,
            )
        )

    # Console Handler
    sh = logging.StreamHandler(sys.stdout)
    log_handlers.append(sh)

    # 给所有 handler 挂 TraceIdFilter（不能挂在 root logger 上：child logger 触发
    # 的 record 在传播时只过 root 的 handlers，不再过 root logger 的 filter()）
    trace_filter = TraceIdFilter()
    for h in log_handlers:
        h.addFilter(trace_filter)

    logging.basicConfig(
        level=LOG_LEVEL,
        format=LOG_FORMAT,
        datefmt=LOG_DATEFORMAT,
        handlers=log_handlers,
        force=True,
    )
