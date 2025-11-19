import time
import functools
import logging
import inspect
from datetime import datetime
from typing import Optional, Callable, Any

logger = logging.getLogger(__name__)

def advanced_timing_decorator(task_name: Optional[str] = None,
                              log_level: str = 'INFO',
                              include_args: bool = False) -> Callable:
    """
    增强型计时装饰器，支持普通函数与流式生成器函数。
    对流式函数，结束时统计并输出：
      - First token delay: 第一个 token 延迟（秒）
      - Total time: 总耗时（秒）
      - Total tokens: 处理的 token 数量
      - Tokens per second: 每秒处理的 token 数
    """
    def decorator(func: Callable) -> Callable:
        func_logger = logging.getLogger(func.__module__)
        
        @functools.wraps(func)
        def wrapper(*args, **kwargs) -> Any:
            original_name = task_name or func.__name__
            start_time = time.perf_counter()
            start_datetime = datetime.now()
            
            try:
                result = func(*args, **kwargs)
            except Exception as e:
                total_time = time.perf_counter() - start_time
                log_msg = f"任务 [{original_name}] 执行失败"
                if include_args:
                    log_msg += f" 参数: args={args}, kwargs={kwargs}"
                log_msg += f" 耗时: {total_time:.3f}秒 (开始于: {start_datetime.strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]})"
                getattr(func_logger, log_level.lower(), func_logger.info)(log_msg)
                raise

            # 如果返回值是生成器，则包装之以记录流式处理的统计数据 [EN] If the return value is a generator, wrap it to record streaming statistics
            if inspect.isgenerator(result):
                def generator_wrapper():
                    item_count = 0
                    first_token_time = None
                    try:
                        for item in result:
                            if first_token_time is None:
                                first_token_time = time.perf_counter() - start_time
                            item_count += 1
                            yield item
                    finally:
                        total_time = time.perf_counter() - start_time
                        tokens_per_second = item_count / total_time if total_time > 0 else 0
                        log_msg = (
                            f"任务 [{original_name}] ---> "
                            f"First token delay: {(first_token_time if first_token_time is not None else 0):.3f} s, "
                            f"Total time: {total_time:.3f} s, "
                            f"Total tokens: {item_count}, "
                            f"Tokens per second: {tokens_per_second:.1f}"
                        )
                        getattr(func_logger, log_level.lower(), func_logger.info)(log_msg)
                return generator_wrapper()
            else:
                # 非生成器函数：直接记录总耗时 [EN] Non-generator function: directly record the total time taken
                total_time = time.perf_counter() - start_time
                # log_msg = ""
                log_msg = f"任务 [{original_name}]"
                if include_args:
                    log_msg += f" 参数: args={args}, kwargs={kwargs}"
                log_msg += f"耗时: {total_time:.3f}秒 (开始于: {start_datetime.strftime('%Y-%m-%d %H:%M:%S.%f')[:-3]})"
                getattr(func_logger, log_level.lower(), func_logger.info)(log_msg)
                return result                
        return wrapper
    return decorator

# 预设的不同日志级别的装饰器 [EN] Preset decorators for different log levels
def advanced_timing_debug(task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    return advanced_timing_decorator(task_name, 'DEBUG', include_args)

def advanced_timing_info(task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    return advanced_timing_decorator(task_name, 'INFO', include_args)

def advanced_timing_warning(task_name: Optional[str] = None, include_args: bool = False) -> Callable:
    return advanced_timing_decorator(task_name, 'WARNING', include_args)

# 使用示例 [EN] Usage example
if __name__ == "__main__":
    # 配置日志输出 [EN] Configure log output
    logging.basicConfig(level=logging.DEBUG, format='%(levelname)s:%(name)s:%(message)s')
    
    @advanced_timing_decorator("普通函数示例")
    def normal_function(x, y):
        time.sleep(0.5)
        return x + y
    
    @advanced_timing_decorator("流式函数示例")
    def stream_function():
        for i in range(3):
            time.sleep(0.1)
            yield i
    
    # 测试普通函数 [EN] Test normal functions
    result = normal_function(1, 2)
    logger.debug("普通函数结果: %s", result)
    
    # 测试流式生成器函数 [EN] Test streaming generator functions
    for item in stream_function():
        logger.debug("流式处理项目: %s", item)
