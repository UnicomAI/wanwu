// Package trace_util 提供 WGA agent 级别的 trace 上下文传播工具函数。
package trace_util

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// wgaTraceKeyType 是 WGA trace 上下文的 context key 类型。
type wgaTraceKeyType struct{}

// WgaTraceContext 存储 WGA agent 级别的 trace 元数据，
// 通过 context 传递给 Eino 回调处理器，用于丰富 span 属性。
type WgaTraceContext struct {
	AgentID   string // agent 配置 ID
	AgentType string // react / sandbox / sequential / loop / parallel / supervisor
	AgentName string // agent 名称
	ThreadID  string // 对话会话 ID
	RunID     string // 执行请求 ID
	Model     string // 模型标识
}

// SetWgaTraceContext 存储 WGA trace 元数据到 context。
func SetWgaTraceContext(ctx context.Context, wgaCtx *WgaTraceContext) context.Context {
	return context.WithValue(ctx, wgaTraceKeyType{}, wgaCtx)
}

// GetWgaTraceContext 从 context 读取 WGA trace 元数据。
// 如果不存在返回 nil。
func GetWgaTraceContext(ctx context.Context) *WgaTraceContext {
	val := ctx.Value(wgaTraceKeyType{})
	if val == nil {
		return nil
	}
	wgaCtx, ok := val.(*WgaTraceContext)
	if !ok {
		return nil
	}
	return wgaCtx
}

// ExtractTraceHeaders 从 context 提取 W3C trace 传播头（traceparent, tracestate, baggage）。
// 用于在跨进程调用（如 sandbox HTTP 请求）时传播 trace 上下文。
func ExtractTraceHeaders(ctx context.Context) map[string]string {
	headers := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
	return headers
}

// InjectTraceHeaders 将 trace 传播头注入到新 context。
// 用于从存储的 trace 头恢复 trace 上下文。
func InjectTraceHeaders(ctx context.Context, headers map[string]string) context.Context {
	if len(headers) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(headers))
}

// StartAgentSpan 创建顶层 agent 执行 span。
// span 名称为 "wga.agent.{agentType}"，并设置 agent 相关属性。
func StartAgentSpan(ctx context.Context, agentType, agentName, agentID string) (context.Context, trace.Span) {
	tracer := _tracer.tp.Tracer("wga")
	ctx, span := tracer.Start(ctx, "wga.agent."+agentType,
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	span.SetAttributes(
		attribute.String("wga.agent.id", agentID),
		attribute.String("wga.agent.type", agentType),
		attribute.String("wga.agent.name", agentName),
	)

	// 附加 WGA trace 上下文中的属性
	if wgaCtx := GetWgaTraceContext(ctx); wgaCtx != nil {
		span.SetAttributes(
			attribute.String("wga.thread.id", wgaCtx.ThreadID),
			attribute.String("wga.run.id", wgaCtx.RunID),
		)
		if wgaCtx.Model != "" {
			span.SetAttributes(attribute.String("wga.model", wgaCtx.Model))
		}
	}

	return ctx, span
}

// WrapIteratorWithSpan 包装 AsyncIterator，在迭代结束时自动结束 span。
func WrapIteratorWithSpan(iter *adk.AsyncIterator[*adk.AgentEvent], span trace.Span) *adk.AsyncIterator[*adk.AgentEvent] {
	outIter, outGen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer span.End()
		defer outGen.Close()
		for {
			event, ok := iter.Next()
			if !ok {
				return
			}
			outGen.Send(event)
		}
	}()

	return outIter
}
