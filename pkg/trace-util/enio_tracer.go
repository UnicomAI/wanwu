package trace_util

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type EnioTracingHandler struct {
	tracer trace.Tracer
}

func EnioGlobalTracing() {
	// 注册全局回调处理器
	callbacks.AppendGlobalHandlers(NewEnioTracingHandler())
}

func NewEnioTracingHandler() *EnioTracingHandler {
	return &EnioTracingHandler{
		tracer: _tracer.tp.Tracer("eino"),
	}
}

func (h *EnioTracingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	spanName := "eino." + info.Type + "." + info.Name

	//log.Infof("Starting span: %s", spanName)
	// 创建新的Span
	ctx, span := h.tracer.Start(ctx, spanName)

	// 记录组件元数据
	span.SetAttributes(
		attribute.String("component.type", info.Type),
		attribute.String("component.name", info.Name),
		attribute.String("node.type", fmt.Sprintf("%T", info.Component)),
	)

	// 记录输入参数（根据实际类型转换）
	if modelInput, ok := input.(*model.CallbackInput); ok {
		span.SetAttributes(
			attribute.Int("input.messages.count", len(modelInput.Messages)),
			attribute.String("input.config.model", modelInput.Config.Model),
		)
	}

	return ctx
}

func (h *EnioTracingHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ctx
	}
	//log.Infof("end span")
	//// 记录执行结果
	//if modelOutput, ok := output.(*model.CallbackOutput); ok {
	//	span.SetAttributes(
	//		attribute.Int("output.messages.count", len(modelOutput.Messages)),
	//		attribute.Bool("output.has_tool_calls", len(modelOutput.ToolCalls) > 0),
	//	)
	//}

	span.End()
	return ctx
}

func (h *EnioTracingHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
	}
	return ctx
}

func (h *EnioTracingHandler) OnStartWithStreamInput(
	ctx context.Context,
	info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput],
) context.Context {
	ctx, span := h.tracer.Start(ctx, "eino.stream."+info.Type)
	log.Infof("strat stream span")
	span.SetAttributes(
		attribute.String("stream.type", "input"),
		attribute.String("component.type", info.Type),
	)

	return ctx
}

func (h *EnioTracingHandler) OnEndWithStreamOutput(
	ctx context.Context,
	info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput],
) context.Context {
	log.Infof("end stream span")
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(
			attribute.String("stream.type", "output"),
			attribute.Bool("stream.completed", true),
		)
		span.End()
	}
	return ctx
}
