// Package trace_util 提供 WGA agent 级别的 OTel trace 回调处理器。
package trace_util

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WgaGlobalTracing 注册 WGA agent 级别 trace 回调处理器
func WgaGlobalTracing() {
	callbacks.AppendGlobalHandlers(NewWgaTracingHandler())
}

// WgaTracingHandler 实现 callbacks.Handler，为 Eino 组件调用创建 OTel span。
// 它统一了原 EnioTracingHandler 的功能，并添加了 WGA 特有的 span 属性。
//
// 此处理器应通过 WgaGlobalTracing() 注册为全局回调，而非手动注册。
type WgaTracingHandler struct {
	tracer trace.Tracer
}

// NewWgaTracingHandler 创建 WGA trace 回调处理器。
func NewWgaTracingHandler() *WgaTracingHandler {
	return &WgaTracingHandler{
		tracer: _tracer.tp.Tracer("wga"),
	}
}

// Needed 实现 callbacks.TimingChecker 接口。
// 仅当 TracerProvider 已初始化时才处理回调，
// 避免在 trace 未启用时产生无意义开销。
func (h *WgaTracingHandler) Needed(_ context.Context, _ *callbacks.RunInfo, _ callbacks.CallbackTiming) bool {
	return _tracer.tp != nil
}

// OnStart 在组件开始执行时创建 span。
func (h *WgaTracingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	spanName := h.buildSpanName(info)
	ctx, span := h.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	// 基础属性
	span.SetAttributes(
		attribute.String("wga.component.type", string(info.Component)),
		attribute.String("wga.component.name", info.Name),
	)

	// WGA 上下文属性
	h.setWgaAttributes(ctx, span)

	// 组件特定输入属性
	h.recordInputAttributes(span, info, input)

	return ctx
}

// OnEnd 在组件执行结束时记录输出属性并结束 span。
func (h *WgaTracingHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ctx
	}

	h.recordOutputAttributes(span, info, output)
	span.End()
	return ctx
}

// OnError 在组件执行出错时记录错误并结束 span。
func (h *WgaTracingHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
	}
	return ctx
}

// OnStartWithStreamInput 在流式输入开始时创建 span。
func (h *WgaTracingHandler) OnStartWithStreamInput(
	ctx context.Context,
	info *callbacks.RunInfo,
	_ *schema.StreamReader[callbacks.CallbackInput],
) context.Context {
	spanName := "wga.stream." + h.componentTag(info) + "." + info.Name
	ctx, span := h.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	span.SetAttributes(
		attribute.String("wga.component.type", string(info.Component)),
		attribute.String("wga.component.name", info.Name),
		attribute.Bool("wga.streaming", true),
		attribute.String("wga.stream.direction", "input"),
	)

	h.setWgaAttributes(ctx, span)

	return ctx
}

// OnEndWithStreamOutput 在流式输出结束时结束 span。
func (h *WgaTracingHandler) OnEndWithStreamOutput(
	ctx context.Context,
	info *callbacks.RunInfo,
	_ *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(
			attribute.String("wga.stream.direction", "output"),
			attribute.Bool("wga.stream.completed", true),
		)
		span.End()
	}
	return ctx
}

// buildSpanName 根据组件类型构建 span 名称。
func (h *WgaTracingHandler) buildSpanName(info *callbacks.RunInfo) string {
	tag := h.componentTag(info)
	name := info.Name
	if name == "" {
		name = string(info.Component)
	}
	return "wga." + tag + "." + name
}

// componentTag 根据 Eino 组件类型返回简短标签。
func (h *WgaTracingHandler) componentTag(info *callbacks.RunInfo) string {
	switch info.Component {
	case components.ComponentOfChatModel:
		return "llm"
	case components.ComponentOfTool:
		return "tool"
	default:
		// Graph, Chain, Lambda, Prompt, Embedding, Retriever, Indexer, Loader, Transformer 等
		return string(info.Component)
	}
}

// setWgaAttributes 从 context 读取 WGA trace 元数据并设为 span 属性。
func (h *WgaTracingHandler) setWgaAttributes(ctx context.Context, span trace.Span) {
	wgaCtx := GetWgaTraceContext(ctx)
	if wgaCtx == nil {
		return
	}
	span.SetAttributes(
		attribute.String("wga.agent.id", wgaCtx.AgentID),
		attribute.String("wga.agent.type", wgaCtx.AgentType),
		attribute.String("wga.agent.name", wgaCtx.AgentName),
		attribute.String("wga.thread.id", wgaCtx.ThreadID),
		attribute.String("wga.run.id", wgaCtx.RunID),
	)
	if wgaCtx.Model != "" {
		span.SetAttributes(attribute.String("wga.model", wgaCtx.Model))
	}
}

// recordInputAttributes 记录组件特定的输入属性。
func (h *WgaTracingHandler) recordInputAttributes(span trace.Span, info *callbacks.RunInfo, input callbacks.CallbackInput) {
	switch info.Component {
	case components.ComponentOfChatModel:
		if modelInput := model.ConvCallbackInput(input); modelInput != nil {
			attrs := []attribute.KeyValue{
				attribute.Int("wga.llm.input.messages.count", len(modelInput.Messages)),
			}
			if modelInput.Config != nil && modelInput.Config.Model != "" {
				attrs = append(attrs, attribute.String("wga.llm.model", modelInput.Config.Model))
			}
			if len(modelInput.Tools) > 0 {
				attrs = append(attrs, attribute.Int("wga.llm.input.tools.count", len(modelInput.Tools)))
			}
			span.SetAttributes(attrs...)
		}
	case components.ComponentOfTool:
		if toolInput := tool.ConvCallbackInput(input); toolInput != nil {
			span.SetAttributes(
				attribute.String("wga.tool.name", info.Name),
			)
			if toolInput.ArgumentsInJSON != "" {
				maxLen := 2000
				arg := toolInput.ArgumentsInJSON
				if len(arg) > maxLen {
					arg = arg[:maxLen] + "...(truncated)"
				}
				span.SetAttributes(attribute.String("wga.tool.arguments", arg))
			}
		}
	}
}

// recordOutputAttributes 记录组件特定的输出属性。
func (h *WgaTracingHandler) recordOutputAttributes(span trace.Span, info *callbacks.RunInfo, output callbacks.CallbackOutput) {
	switch info.Component {
	case components.ComponentOfChatModel:
		if modelOutput := model.ConvCallbackOutput(output); modelOutput != nil {
			attrs := []attribute.KeyValue{}
			if modelOutput.TokenUsage != nil {
				attrs = append(attrs,
					attribute.Int("wga.llm.prompt_tokens", modelOutput.TokenUsage.PromptTokens),
					attribute.Int("wga.llm.completion_tokens", modelOutput.TokenUsage.CompletionTokens),
					attribute.Int("wga.llm.total_tokens", modelOutput.TokenUsage.TotalTokens),
				)
				if modelOutput.TokenUsage.CompletionTokensDetails.ReasoningTokens > 0 {
					attrs = append(attrs,
						attribute.Int("wga.llm.reasoning_tokens", modelOutput.TokenUsage.CompletionTokensDetails.ReasoningTokens),
					)
				}
				if modelOutput.TokenUsage.PromptTokenDetails.CachedTokens > 0 {
					attrs = append(attrs,
						attribute.Int("wga.llm.cached_tokens", modelOutput.TokenUsage.PromptTokenDetails.CachedTokens),
					)
				}
			}
			if modelOutput.Config != nil && modelOutput.Config.Model != "" {
				attrs = append(attrs, attribute.String("wga.llm.model", modelOutput.Config.Model))
			}
			if modelOutput.Message != nil {
				attrs = append(attrs,
					attribute.Bool("wga.llm.has_tool_calls", len(modelOutput.Message.ToolCalls) > 0),
				)
			}
			if len(attrs) > 0 {
				span.SetAttributes(attrs...)
			}
		}
	case components.ComponentOfTool:
		if toolOutput := tool.ConvCallbackOutput(output); toolOutput != nil {
			attrs := []attribute.KeyValue{
				attribute.Bool("wga.tool.success", true),
			}
			if toolOutput.Response != "" {
				maxLen := 2000
				resp := toolOutput.Response
				if len(resp) > maxLen {
					resp = resp[:maxLen] + "...(truncated)"
				}
				attrs = append(attrs, attribute.String("wga.tool.response", resp))
			}
			span.SetAttributes(attrs...)
		}
	}
}

// ensure WgaTracingHandler implements callbacks.Handler at compile time.
var _ callbacks.Handler = (*WgaTracingHandler)(nil)

// ensure WgaTracingHandler implements callbacks.TimingChecker at compile time.
var _ callbacks.TimingChecker = (*WgaTracingHandler)(nil)
