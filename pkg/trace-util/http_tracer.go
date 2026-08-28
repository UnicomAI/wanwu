package trace_util

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func NewTracerGin(serviceName string) *gin.Engine {
	r := gin.Default()
	// 启用 ContextWithFallback，使 gin.Context.Value() 委托到 Request.Context()
	// 这样才能正确获取 otelgin 中间件设置的 trace context
	r.ContextWithFallback = true
	// TraceFromURLMiddleware 必须在 otelgin 之前执行：
	// 它将 URL 路径中的 traceId-spanId 转为 traceparent Header，
	// 这样 otelgin 才能识别并挂到正确的 parent span 上。
	r.Use(TraceFromURLMiddleware())
	r.Use(otelgin.Middleware(serviceName))
	return r
}

// TraceFromURLMiddleware 从 URL 路径中提取 trace 上下文并注入到请求头。
//
// 背景：sandbox 内的 opencode/eino-agent 通过 @ai-sdk/openai-compatible 调模型 API，
// 该 npm 包不支持自定义请求头，因此 traceparent 无法通过 HTTP Header 传递。
// 作为替代方案，trace 信息编码在 URL 路径中：
//
//	/callback/v1/model/{modelId}/trace/{traceId}/span/{spanId}/chat/completions
//
// 本中间件在 otelgin 之前执行，从路径中解析出 traceparent 并注入 Header，
// 确保 otelgin 能将此请求识别为上游 trace 的子 span 而非新 root span。
func TraceFromURLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从路径参数提取（匹配 /model/:modelId/trace/:traceId/span/:spanId/ 模式）
		traceID := c.Param("traceId")
		spanID := c.Param("spanId")
		if traceID != "" && spanID != "" {
			traceparent := fmt.Sprintf("00-%s-%s-01", traceID, spanID)
			c.Request.Header.Set("traceparent", traceparent)
			c.Next()
			return
		}

		// 兜底：从原始路径中按 /model/{id}/trace/{traceId}/span/{spanId}/ 模式解析
		path := c.Request.URL.Path
		parts := strings.Split(strings.Trim(path, "/"), "/")
		for i := 0; i+5 < len(parts); i++ {
			if parts[i] == "model" && parts[i+2] == "trace" && parts[i+4] == "span" {
				candidateTraceID := parts[i+3]
				candidateSpanID := parts[i+5]
				if len(candidateTraceID) >= 32 && len(candidateSpanID) >= 16 {
					traceparent := fmt.Sprintf("00-%s-%s-01", candidateTraceID, candidateSpanID)
					c.Request.Header.Set("traceparent", traceparent)
					break
				}
			}
		}
		c.Next()
	}
}

func NewResty(ctx context.Context) *resty.Client {
	return resty.New().SetPreRequestHook(func(client *resty.Client, request *http.Request) error {
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))
		return nil
	})
}
