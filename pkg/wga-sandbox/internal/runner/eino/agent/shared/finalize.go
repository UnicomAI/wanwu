package shared

import (
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// FinishReasonStop 是 OpenAI / eino 约定的正常停止 FinishReason 值。
// 上游 internal/agent-service/pkg/util.StopMessage 依赖此值判定沙箱收尾。
const FinishReasonStop = "stop"

// FinishReasonLength 是模型因长度上限截断时的 FinishReason 值。
// 沙箱出口处会把它统一改写为 stop，并把原值保留到 Message.Extra[OriginalFinishReasonKey]。
const FinishReasonLength = "length"

// FinishReasonToolCalls 是 OpenAI 规范下模型要继续调用工具时的 FinishReason 值。
// 部分国产模型在该场景仍发 "stop" + tool_calls，沙箱出口会统一标准化为此值。
const FinishReasonToolCalls = "tool_calls"

// OriginalFinishReasonKey 是 NormalizeFinishReason 写入 Message.Extra 的 key，
// 供下游识别"这条 stop 其实是因 length 截断"或"这条 tool_calls 原本是 stop 偏差"。
const OriginalFinishReasonKey = "original_finish_reason"

// StreamEndedWithoutFinishMsg 是迭代器自然结束但模型未给出 stop / length
// 时的兜底 content；保证收尾 stop 帧 content 非空、且语义有信息量。
const StreamEndedWithoutFinishMsg = "stream ended without explicit finish"

// FinalErrorSource 标记兜底消息的产生层，作为前缀写入 content：
// "error[<source>]: <err.Error()>"。调用方可用 strings.HasPrefix
// 分别识别 agent 业务错误与 runner 通道错误。
type FinalErrorSource string

const (
	// FinalErrorSourceAgent 表示沙箱容器内 eino-agent 自报的错误
	// （LLM 调用失败、初始化失败、handler panic、ctx 超时等）。
	FinalErrorSourceAgent FinalErrorSource = "agent"
	// FinalErrorSourceRunner 表示宿主机 wga-sandbox runner 观察到的
	// 通道/容器异常（unexpected EOF、connection reset、i/o timeout 等）。
	FinalErrorSourceRunner FinalErrorSource = "runner"
)

// BuildFinalAssistantMessage 构造 Role=Assistant + FinishReason=stop 的兜底消息。
//
// content 取值：
//   - err != nil：       "error[<source>]: <err.Error()>"
//   - err == nil：       StreamEndedWithoutFinishMsg
//
// 不再返回空字符串——下游统计/UI 不需要为"空 content stop 帧"做特殊处理。
func BuildFinalAssistantMessage(source FinalErrorSource, err error) *schema.Message {
	var content string
	if err != nil {
		content = fmt.Sprintf("error[%s]: %s", source, err.Error())
	} else {
		content = StreamEndedWithoutFinishMsg
	}
	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
		ResponseMeta: &schema.ResponseMeta{
			FinishReason: FinishReasonStop,
		},
	}
}

// NormalizeFinishReason 在沙箱 SSE 出口做一次 finish_reason 标准化，统一收尾语义：
//
//  1. 任何 stop / length + 非空 tool_calls 的帧 → tool_calls。
//     修正部分模型在"还要调工具"时仍发 finish_reason=stop 的协议偏差，让
//     下游"finish_reason==stop"仅对应"agent 真正终结"——react 中间步、
//     transfer_to_agent 第一帧、其它 tool 路由帧不会误触发 stop 通路。
//
//  2. length 且无 tool_calls → stop。保留"length 视为正常收尾"的历史行为。
//
// 原值统一保留到 Message.Extra[OriginalFinishReasonKey]，供下游识别截断或
// 协议偏差。无 ResponseMeta / nil 消息一律不动；重复调用幂等。
//
// 注意 switch 顺序：先判 tool_calls 非空,避免 length + tool_calls 这种罕见
// 组合被误改为 stop 后让 ProcessEvents 提前 sentFinal=true 收尾。
func NormalizeFinishReason(msg *schema.Message) {
	if msg == nil || msg.ResponseMeta == nil {
		return
	}
	original := msg.ResponseMeta.FinishReason
	var normalized string
	switch {
	case len(msg.ToolCalls) > 0 && (original == FinishReasonStop || original == FinishReasonLength):
		normalized = FinishReasonToolCalls
	case original == FinishReasonLength:
		normalized = FinishReasonStop
	default:
		return
	}
	if msg.Extra == nil {
		msg.Extra = make(map[string]any)
	}
	msg.Extra[OriginalFinishReasonKey] = original
	msg.ResponseMeta.FinishReason = normalized
}

// BuildFinalAgentEvent 把兜底 assistant+stop 消息包成 AgentEvent，
// 便于 server 层通过 SSEWriter.WriteAgentEvent 直接写出。
func BuildFinalAgentEvent(source FinalErrorSource, err error) *adk.AgentEvent {
	return &adk.AgentEvent{
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message:     BuildFinalAssistantMessage(source, err),
				Role:        schema.Assistant,
			},
		},
	}
}

// IsFinalStopMessage 判断 schema.Message 是否已经是 assistant+stop 的收尾消息。
func IsFinalStopMessage(m *schema.Message) bool {
	if m == nil {
		return false
	}
	if m.Role != schema.Assistant {
		return false
	}
	if m.ResponseMeta == nil {
		return false
	}
	return m.ResponseMeta.FinishReason == FinishReasonStop
}

// ValidateToolCallArguments 校验 msg 中所有 tool_calls 的 arguments 是否为合法 JSON。
// 空串视为不合法（部分模型在截断时输出空 arguments）。
// 返回首个不合法的 tool_call 描述，全合法返回 nil。
func ValidateToolCallArguments(msg *schema.Message) error {
	for _, tc := range msg.ToolCalls {
		arg := tc.Function.Arguments
		if arg == "" || !json.Valid([]byte(arg)) {
			name := tc.Function.Name
			if name == "" {
				name = "(unnamed)"
			}
			id := tc.ID
			if id == "" {
				id = "(no-id)"
			}
			return fmt.Errorf("tool call %s(%s) arguments invalid: %s", name, id, arg)
		}
	}
	return nil
}
