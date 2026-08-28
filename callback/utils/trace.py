import json
import time

from flask import Flask, g, request
from opentelemetry import trace

from utils.log import logger


def register_tracing(app: Flask):
    """封装路由追踪，集成 OpenTelemetry trace_id/span_id 用于日志关联"""

    @app.before_request
    def start_trace():
        # 保存请求开始时间
        g.start_time = time.time()

        # 提取 trace_id / span_id 用于日志关联
        span = trace.get_current_span()
        if span and span.is_recording():
            ctx = span.get_span_context()
            g.trace_id = format(ctx.trace_id, "032x")
            g.span_id = format(ctx.span_id, "016x")
        else:
            g.trace_id = "-"
            g.span_id = "-"

        # 调试：打印入站 traceparent（从上游 Go 服务传来的）和当前 Span 信息
        # incoming_traceparent = request.headers.get("traceparent", "")
        # parent_span_id = "-"
        # if incoming_traceparent:
        #     parts = incoming_traceparent.split("-")
        #     if len(parts) >= 3:
        #         parent_span_id = parts[2]
        # logger.info(
        #     f"[otel-debug] inbound | "
        #     f"traceparent={incoming_traceparent} | "
        #     f"parent_span_id={parent_span_id} | "
        #     f"trace_id={g.trace_id} span_id={g.span_id}"
        # )

        # 尝试获取请求体
        try:
            if request.is_json:
                req_body = request.get_json(silent=True)
            elif request.form:
                req_body = request.form.to_dict()
            elif request.data:
                req_body = request.get_data(as_text=True)
            else:
                req_body = None
        except Exception:
            req_body = "<无法解析请求体>"

        # 记录请求基本信息（此时不记录完整日志，等响应后再统一输出）
        g.request_log = {
            "method": request.method,
            "full_path": request.full_path,
            "header": dict(request.headers),
            "body": req_body,
        }

    @app.after_request
    def end_trace(response):
        request_log = g.get("request_log", {})
        if "/apidocs" in request_log.get("full_path", ""):
            return response  # 跳过 apidocs 的日志记录

        # 耗时ms
        cost = round((time.time() - g.get("start_time", time.time())) * 1000, 2)

        # 获取原始响应体（仅适用于非流式响应）
        try:
            if response.is_streamed:
                resp_body = "<流式响应，暂无记录>"
            else:
                resp_body = response.get_data(as_text=True)
        except Exception:
            resp_body = "<无法读取响应体>"

        trace_id = g.get("trace_id", "-")
        span_id = g.get("span_id", "-")
        method = request_log.get("method", "-")
        full_path = request_log.get("full_path", "-")
        body = json.dumps(request_log.get("body"), ensure_ascii=False)

        log_msg = (
            f"{cost}ms | {response.status_code} | "
            f"trace_id={trace_id} span_id={span_id} | "
            f"{method} | {full_path} | {body} | {resp_body.rstrip(chr(10))}"
        )
        if response.status_code < 400:
            logger.info(log_msg)
        else:
            logger.error(log_msg)
        return response
