import json
import logging
import time

from flask import Flask, g, request
from opentelemetry import trace

logger = logging.getLogger(__name__)


def register_tracing(app: Flask):
    """封装路由追踪，集成 OpenTelemetry trace_id/span_id 用于日志关联"""

    @app.before_request
    def start_trace():
        g.start_time = time.time()

        span = trace.get_current_span()
        if span and span.is_recording():
            ctx = span.get_span_context()
            g.trace_id = format(ctx.trace_id, "032x")
            g.span_id = format(ctx.span_id, "016x")
        else:
            g.trace_id = "-"
            g.span_id = "-"

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

        g.request_log = {
            "method": request.method,
            "full_path": request.full_path,
            "body": req_body,
        }

    @app.after_request
    def end_trace(response):
        # access log 仅用于观察，任何异常都不得污染业务响应（避免 Flask
        # 把 after_request 的异常转成默认 HTML 500，丢弃 view 已正常返回的 body）
        try:
            request_log = g.get("request_log", {})
            if "/apidocs" in request_log.get("full_path", ""):
                return response

            cost = round((time.time() - g.get("start_time", time.time())) * 1000, 2)

            try:
                if response.is_streamed:
                    resp_body = "<流式响应，暂无记录>"
                else:
                    resp_body = response.get_data(as_text=True)
            except Exception:
                resp_body = "<无法读取响应体>"

            method = request_log.get("method", "-")
            full_path = request_log.get("full_path", "-")

            # trace_id / span_id 已由 logging_config.TraceIdFilter 统一注入，此处不再重复拼接
            log_msg = (
                f"{cost}ms | {response.status_code} | "
                f"{method} | {full_path} | {resp_body.rstrip(chr(10))}"
            )
            if response.status_code < 400:
                logger.info(log_msg)
            else:
                logger.error(log_msg)
        except Exception:
            logger.exception("end_trace failed (access log only, response untouched)")
        return response
