package trace_util

import (
	"context"

	"github.com/UnicomAI/wanwu/pkg/log"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// DetachContext extracts trace propagation info from src and injects it into
// a new context.Background(). The returned context is NOT cancelled when src is
// cancelled, but preserves the trace chain (traceID, spanID) so downstream
// calls (e.g. gRPC) can propagate the trace correctly.
//
// Typical use-case: goroutines that must outlive the HTTP request lifecycle
// (e.g. async statistics recording) but still need trace correlation.
func DetachContext(src context.Context) context.Context {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(src, carrier)

	dst := context.Background()
	dst = otel.GetTextMapPropagator().Extract(dst, carrier)

	return dst
}

func InjectContext(ctx context.Context, traceIDStr, spanIDStr string) context.Context {
	// 2. 将十六进制字符串解析为 TraceID 类型
	traceID, err := trace.TraceIDFromHex(traceIDStr)
	if err != nil {
		log.Errorf("解析 TraceID 失败: %v\n", err)
		return ctx
	}

	// 3. 生成一个临时的 SpanID
	// 必须提供一个 16 位十六进制字符串，此处仅供示例，生产环境需自行生成
	spanID, err := trace.SpanIDFromHex(spanIDStr)
	if err != nil {
		log.Errorf("解析 SpanID 失败: %v\n", err)
		return ctx
	}

	// 4. 构建一个 SpanContext，必须同时包含 TraceID 和 SpanID
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
		// 如果需要采样，可以设置 Remote: true 或 TraceFlags: trace.FlagsSampled
	})

	// 5. 将 SpanContext 注入到 context 中
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}
