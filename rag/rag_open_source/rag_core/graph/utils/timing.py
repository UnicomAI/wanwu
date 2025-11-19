# coding=utf-8
import time
import functools
import logging
from typing import Optional, Callable, Any
from datetime import datetime


def timing_decorator(logger,
                     task_name: Optional[str] = None,
                     log_level: str = 'INFO',
                     include_args: bool = False,
                     ) -> Callable:
    """
    用于测量函数执行时间的装饰器

    参数:
        task_name: 任务名称，如果不提供则使用函数名
        log_level: 日志级别 ('DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL')
        include_args: 是否在日志中包含函数参数

    用法示例:
        @timing_decorator()
        def my_function():
            pass

        @timing_decorator("自定义任务名")
        def another_function():
            pass

        @timing_decorator(log_level='DEBUG', include_args=True)
        def function_with_args(x, y):
            pass
    """

    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> Any:
            # 确定任务名称 [EN] Determine task name
            actual_task_name = task_name if task_name is not None else func.__name__

            # 获取开始时间 [EN] Get start time
            start_time = time.perf_counter()
            start_datetime = datetime.now()

            try:
                # 执行原函数 [EN] Execute original function
                result = func(*args, **kwargs)
                success = True
            except Exception as e:
                success = False
                raise e
            finally:
                # 计算执行时间 [EN] Calculate execution time
                end_time = time.perf_counter()
                execution_time = end_time - start_time

                # 构建日志消息 [EN] Build log messages
                log_message = f"任务 [{actual_task_name}] "
                if include_args:
                    log_message += f"参数: args={args}, kwargs={kwargs} "
                log_message += f"耗时: {execution_time:.3f}秒"
                if not success:
                    log_message += " (执行失败)"

                # 添加开始时间信息 [EN] Add start time information
                log_message += f" (开始于: {start_datetime.strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]})"

                # 根据指定的日志级别记录 [EN] Logging according to the specified log level
                log_func = getattr(logger, log_level.lower(), logger.info)
                log_func(log_message)

            return result

        return wrapper

    return decorator


# 为了方便使用，提供一些预设的装饰器 [EN] For ease of use, some preset decorators are provided
def timing_debug(logger, task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    """DEBUG级别的计时装饰器"""
    return timing_decorator(logger, task_name, 'DEBUG', include_args)


def timing_info(logger, task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    """INFO级别的计时装饰器"""
    return timing_decorator(logger, task_name, 'INFO', include_args)


def timing_warning(logger, task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    """WARNING级别的计时装饰器"""
    return timing_decorator(logger, task_name, 'WARNING', include_args)


# 使用示例 [EN] Usage example
if __name__ == "__main__":
    pass