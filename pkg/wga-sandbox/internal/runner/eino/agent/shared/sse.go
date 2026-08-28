package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudwego/eino/adk"
)

type SSEWriter interface {
	WriteAgentEvent(event *adk.AgentEvent)
}

// sseDebugEnabled 控制 marshalAgentEvent 是否输出完整 payload。
// 默认仅打长度摘要；设置 SSE_DEBUG=1（或任意非空值）启用 full payload。
var sseDebugEnabled = os.Getenv("SSE_DEBUG") != ""

// httpSSEWriter 把 AgentEvent 序列化后以 `data: ...\n\n` 形式刷出。
type httpSSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewHTTPSSEWriter(w http.ResponseWriter, flusher http.Flusher) SSEWriter {
	return &httpSSEWriter{w: w, flusher: flusher}
}

func (h *httpSSEWriter) WriteAgentEvent(event *adk.AgentEvent) {
	jsonData, _ := marshalAgentEvent(event)
	_, _ = fmt.Fprintf(h.w, "data: %s\n\n", jsonData)
	h.flusher.Flush()
}

func marshalAgentEvent(event *adk.AgentEvent) ([]byte, error) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("[SSE] marshal AgentEvent failed: %v", err)
		errEvent := &adk.AgentEvent{Err: fmt.Errorf("内部序列化错误")}
		return json.Marshal(errEvent)
	}
	role := eventRole(event)
	if sseDebugEnabled {
		log.Printf("[SSE] >>> role=%s bytes=%d payload=%s", role, len(jsonData), string(jsonData))
	} else {
		log.Printf("[SSE] >>> role=%s bytes=%d", role, len(jsonData))
	}
	return jsonData, nil
}

func eventRole(event *adk.AgentEvent) string {
	if event == nil || event.Output == nil || event.Output.MessageOutput == nil {
		return ""
	}
	return string(event.Output.MessageOutput.Role)
}

// --- 公共事件处理 ---

// ProcessEvents 消费 adk 事件迭代器并经 SSEWriter 写出。
//
// 返回 sentFinal 表示已经写出一条 Role=Assistant + FinishReason=stop 的兜底/正常收尾消息，
// 调用方据此决定是否还需要补发兜底，避免重复。
//
// sentFinal=true 的三种来源：
//  1. forwardMessageEvent 转发的消息已经是 assistant+stop（包括 length 被改写后的 stop）；
//  2. 迭代器自然结束时由本函数补发一条 BuildFinalAgentEvent；
//  3. 任何错误路径（event.Err / forwardMessageEvent 物化失败 / iter.Err 等价的 ctx 错）
//     都先写出诊断 event，再补一条 BuildFinalAgentEvent。
//
// 不变量：函数返回时 sentFinal 一定为 true，下游 handler 不需要再补 stop 帧。
func ProcessEvents(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent], w SSEWriter) (eventCount int, sentFinal bool) {
	writeFinal := func(err error) {
		w.WriteAgentEvent(BuildFinalAgentEvent(FinalErrorSourceAgent, err))
		sentFinal = true
	}

	for {
		event, ok := iter.Next()
		if !ok {
			log.Printf("[Events] iterator closed, total=%d sentFinal=%v", eventCount, sentFinal)
			if !sentFinal {
				// 迭代器自然结束：无论 ctx 是否 cancel，都补一条 stop 兜底。
				// ctx.Err() 可能为 nil —— BuildFinalAssistantMessage 会用
				// StreamEndedWithoutFinishMsg 给一个有意义的非空 content。
				var ctxErr error
				if ctx != nil {
					ctxErr = ctx.Err()
				}
				writeFinal(ctxErr)
			}
			return
		}

		if event.Err != nil {
			log.Printf("[Events] error event #%d agent=%s role=%s tool=%s err=%v",
				eventCount, event.AgentName, eventRole(event), eventToolName(event), event.Err)
			// 返回原诊断错误事件给客户端（保持现有逻辑），再补一条兜底 assistant+stop。
			w.WriteAgentEvent(event)
			writeFinal(event.Err)
			return
		}

		eventCount++

		if event.Output == nil || event.Output.MessageOutput == nil {
			log.Printf("[Events] event #%d empty output, skipping", eventCount)
			continue
		}

		isFinal, err := forwardMessageEvent(event, w, eventCount)
		if err != nil {
			writeFinal(err)
			return
		}
		if isFinal {
			// 这条已经满足 assistant+stop（含 length 改写后的 stop）：
			// 直接收尾，由 caller 的 sentFinal=true 通路跳过 handler defer 的二次补发。
			sentFinal = true
			return
		}
	}
}

// forwardMessageEvent 把一个带 MessageOutput 的 event 物化为非流式后转发。
// 流式与非流式共用同一条转发路径：MessageVariant.GetMessage() 内部已经处理
// schema.ConcatMessageStream，无需我们手动 Recv 循环。
//
// 返回 isFinal 表示写出的这条消息已是 assistant+stop（含 length 被原地改写为 stop
// 的情况），ProcessEvents 据此置 sentFinal 以避免 handler defer 二次补发。
func forwardMessageEvent(event *adk.AgentEvent, w SSEWriter, eventNum int) (isFinal bool, err error) {
	mv := event.Output.MessageOutput
	msg, err := mv.GetMessage()
	if err != nil {
		log.Printf("[Events] event #%d get message failed: agent=%s role=%s tool=%s err=%v",
			eventNum, event.AgentName, string(mv.Role), mv.ToolName, err)
		w.WriteAgentEvent(buildDiagnosticEvent(event, fmt.Errorf("failed to materialize message: %w", err)))
		return false, err
	}
	if msg == nil {
		log.Printf("[Events] event #%d materialized message is nil: agent=%s role=%s tool=%s",
			eventNum, event.AgentName, string(mv.Role), mv.ToolName)
		err := fmt.Errorf("materialized message is nil")
		w.WriteAgentEvent(buildDiagnosticEvent(event, err))
		return false, err
	}

	NormalizeFinishReason(msg)

	// 校验 tool_calls arguments 合法性：finish_reason == tool_calls 时，若 arguments 为空或非法 JSON，
	// 说明模型输出被截断或协议偏差，直接走 writeFinal(err) 输出 error stop 帧，
	// 避免下游 tools_node 尝试 unmarshal 损坏参数后报更难追踪的错误。
	if msg.ResponseMeta != nil && msg.ResponseMeta.FinishReason == FinishReasonToolCalls && len(msg.ToolCalls) > 0 {
		if validateErr := ValidateToolCallArguments(msg); validateErr != nil {
			log.Printf("[Events] event #%d invalid tool_calls arguments: %v", eventNum, validateErr)
			w.WriteAgentEvent(buildDiagnosticEvent(event, validateErr))
			return false, validateErr
		}
	}

	var finish string
	if msg.ResponseMeta != nil {
		finish = msg.ResponseMeta.FinishReason
	}
	log.Printf("[Events] event #%d role=%s tool=%s streaming=%v content=%d tool_calls=%d finish=%s",
		eventNum, string(mv.Role), mv.ToolName, mv.IsStreaming, len(msg.Content), len(msg.ToolCalls), finish)

	w.WriteAgentEvent(&adk.AgentEvent{
		AgentName: event.AgentName,
		RunPath:   event.RunPath,
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Message:     msg,
				Role:        mv.Role,
				ToolName:    mv.ToolName,
			},
		},
		Action: event.Action,
	})
	return IsFinalStopMessage(msg), nil
}

func buildDiagnosticEvent(src *adk.AgentEvent, err error) *adk.AgentEvent {
	return &adk.AgentEvent{
		AgentName: src.AgentName,
		RunPath:   src.RunPath,
		Output: &adk.AgentOutput{
			MessageOutput: &adk.MessageVariant{
				IsStreaming: false,
				Role:        src.Output.MessageOutput.Role,
				ToolName:    src.Output.MessageOutput.ToolName,
			},
		},
		Err: err,
	}
}

func eventToolName(event *adk.AgentEvent) string {
	if event == nil || event.Output == nil || event.Output.MessageOutput == nil {
		return ""
	}
	return event.Output.MessageOutput.ToolName
}
