import logging
import time

from opentelemetry import trace as otel_trace

logger = logging.getLogger(__name__)


class TraceLoggingMiddleware:
    """FastAPI/ASGI access log 中间件。

    使用纯 ASGI 实现，避免 Starlette BaseHTTPMiddleware 缓冲 SSE 流式响应导致
    问答接口失效的问题。trace_id / span_id 由 logging_config.TraceIdFilter 在
    所有日志行上自动注入，本中间件只负责输出 access log（cost / status / 路径）。
    """

    def __init__(self, app):
        self.app = app

    async def __call__(self, scope, receive, send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        start_time = time.time()
        method = scope.get("method", "-")
        path = scope.get("path", "-")
        query_string = scope.get("query_string", b"").decode("utf-8", errors="ignore")
        full_path = f"{path}?{query_string}" if query_string else path

        if "/apidocs" in full_path or "/docs" in full_path or "/openapi.json" in full_path:
            await self.app(scope, receive, send)
            return

        status_holder = {"code": 0}

        async def send_wrapper(message):
            if message.get("type") == "http.response.start":
                status_holder["code"] = message.get("status", 0)
            await send(message)

        try:
            await self.app(scope, receive, send_wrapper)
        finally:
            # access log 仅用于观察，任何异常都不得影响下游 ASGI 行为
            try:
                cost = round((time.time() - start_time) * 1000, 2)
                status = status_holder["code"]

                # trace_id / span_id 已由 logging_config.TraceIdFilter 统一注入，此处不再重复拼接
                log_msg = (
                    f"{cost}ms | {status} | "
                    f"{method} | {full_path}"
                )
                if status and status < 400:
                    logger.info(log_msg)
                else:
                    logger.error(log_msg)
            except Exception:
                logger.exception("TraceLoggingMiddleware access log failed (response untouched)")


class FirstChunkSpanMiddleware:
    """只为流式响应的首个 body 分片建一个 span，其余分片不建 span。

    配合 FastAPIInstrumentor(..., exclude_spans=["send", "receive"]) 使用：
    instrumentation 不再为每个 ASGI send 事件建 "http send" 子 span，避免一条
    流式（SSE）trace 产生成百上千个 span；本中间件改为只记录首个非空分片，
    span 时长 = 请求开始 -> 首分片，可用于观测首字节 / 首 token 耗时。

    纯 ASGI 实现，与 TraceLoggingMiddleware 一致，不缓冲流式响应。
    span 不显式指定 parent：send_wrapper 仍运行在 OTel 请求 span 的 contextvar
    作用域内，会自动挂在根请求 span 下。
    """

    def __init__(self, app):
        self.app = app
        self._tracer = otel_trace.get_tracer(__name__)

    async def __call__(self, scope, receive, send):
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        start_ns = time.time_ns()
        state = {"done": False}

        async def send_wrapper(message):
            if (not state["done"]
                    and message.get("type") == "http.response.body"
                    and message.get("body")):
                state["done"] = True
                # 建一个 span 即时结束，时长覆盖 请求开始 -> 首分片。
                # 任何埋点异常都不得影响流式响应本身，故 try 兜底。
                try:
                    self._tracer.start_span("http first_output", start_time=start_ns).end()
                except Exception:
                    logger.exception("FirstChunkSpanMiddleware create span failed (response untouched)")
            await send(message)

        await self.app(scope, receive, send_wrapper)
