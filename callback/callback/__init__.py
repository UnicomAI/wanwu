import logging

from flasgger import Swagger
from flask import Flask

from configs.config import load_config
from extensions.minio import init_minio
from extensions.redis import init_redis
from utils.otel import init_tracer
from utils.response import register_error_handlers
from utils.trace import register_tracing


def create_app():
    # 初始化 OpenTelemetry（必须在 Flask app 创建之前）
    init_tracer()

    # 自动 instrument requests 库（出站 HTTP 传播 traceparent）
    from opentelemetry.instrumentation.requests import RequestsInstrumentor

    RequestsInstrumentor().instrument()

    app = Flask(__name__)
    app.json.ensure_ascii = False
    app.config["SWAGGER"] = {"openapi": "3.0.1"}
    # 初始化 swagger
    Swagger(app)

    # 自动 instrument Flask（入站 HTTP 提取 traceparent，创建 Span）
    from opentelemetry.instrumentation.flask import FlaskInstrumentor

    FlaskInstrumentor().instrument_app(app)

    # init config
    load_config()

    # init redis
    init_redis()

    # init minio
    init_minio()

    # 添加日志记录（含 trace_id/span_id 关联）
    register_tracing(app)

    # 注册异常处理
    register_error_handlers(app)

    # 注册蓝图
    from callback.routes import callback_bp

    app.register_blueprint(callback_bp, url_prefix="/v1")

    return app
