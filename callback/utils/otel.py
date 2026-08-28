import os

from opentelemetry import trace
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.sampling import ALWAYS_ON

from utils.log import logger

SERVICE_NAME_CALLBACK = "callback-wanwu"


def init_tracer():
    """初始化 OpenTelemetry TracerProvider，与 Go 侧对齐。

    - 采样策略: AlwaysSample（与 Go 侧一致）
    - 传播格式: W3C TraceContext + Baggage（OTel 默认，与 Go 侧一致）
    - 通过 JAEGER_ENABLE 控制是否启用导出，JAEGER_OTLP_ENDPOINT 指定导出地址
      （与 Go 服务的环境变量命名一致）
    """
    jaeger_enable = os.environ.get("JAEGER_ENABLE", "").lower() in ("true", "1", "yes")
    endpoint = os.environ.get("JAEGER_OTLP_ENDPOINT", "")

    if jaeger_enable and endpoint:
        # Go 侧的环境变量值不含 http:// 前缀（如 jaeger-wanwu:4318），
        # 但 OTLP HTTP Exporter 需要完整 URL，此处自动补全
        if not endpoint.startswith("http://") and not endpoint.startswith("https://"):
            endpoint = f"http://{endpoint}"

        # Go SDK 的 WithEndpoint 会自动拼接 /v1/traces，
        # 但 Python OTLPSpanExporter 的 endpoint 参数是完整路径，不会自动拼接，
        # 需要手动补上
        if not endpoint.rstrip("/").endswith("/v1/traces"):
            endpoint = endpoint.rstrip("/") + "/v1/traces"

        from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
        from opentelemetry.sdk.trace.export import BatchSpanProcessor

        resource = Resource.create({SERVICE_NAME: SERVICE_NAME_CALLBACK})
        provider = TracerProvider(sampler=ALWAYS_ON, resource=resource)
        exporter = OTLPSpanExporter(endpoint=endpoint)
        provider.add_span_processor(BatchSpanProcessor(exporter))
    else:
        provider = TracerProvider(sampler=ALWAYS_ON)

    trace.set_tracer_provider(provider)

    mode = (
        f"export to {endpoint}, service={SERVICE_NAME_CALLBACK}"
        if jaeger_enable and endpoint
        else "noop (no exporter)"
    )
    logger.info(f"tracer initialized: mode={mode}")

    return provider
