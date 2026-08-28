package eino

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/shared"
)

// connectSSE 打开到沙箱内 eino-agent /chat 端点的 SSE 长连接，
// 返回 sseCh（仅含 `data:` 行 payload）以及 streamErr —— 后者用于在流阶段
// 异常断开（如沙箱容器重启导致 connection reset / unexpected EOF）时，
// 把底层 IO 错误暴露给 forwardSSEStream，最终通过 BuildFinalAgentEvent
// 回传给调用方。
//
// 连接握手与数据流分两段：
//  1. 同步阶段：发起 POST 请求并读取响应头，若失败立即通过 errCh 返回错误。
//  2. 流阶段：连接握手成功后 close(connected) 唤醒外部调用方，goroutine
//     继续 scan body 把每行 SSE data 投递到 sseCh，直到 EOF 或 ctx 取消。
//     scanner 异常错误写入 streamErr 后再关闭 sseCh。
func (r *Runner) connectSSE(ctx context.Context) (<-chan string, *streamErrHolder, error) {
	sseCh := make(chan string, 1024)
	errCh := make(chan error, 1)
	connected := make(chan struct{})
	streamErr := &streamErrHolder{}

	go func() {
		defer util.PrintPanicStack()
		// sseCh / errCh 的 close 顺序：connected 信号触发后由本 goroutine 负责
		// close sseCh（消费者用 range 退出）；errCh 同时关闭，避免外部死等。
		defer close(sseCh)
		defer close(errCh)

		resp, err := trace_util.NewResty(ctx).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			SetTimeout(0).
			R().
			SetContext(ctx).
			SetQueryParam("workspace", r.sb.WorkDir()).
			SetQueryParam("agent_type", r.agentType).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "text/event-stream").
			SetBody(map[string]any{"messages": r.req.Messages}).
			SetDoNotParseResponse(true).
			Post(r.req.Sandbox.EinoEndpoint() + "/chat")

		if err != nil {
			errCh <- fmt.Errorf("SSE connect failed: %w", err)
			return
		}
		defer func() {
			if resp != nil && resp.RawResponse != nil && resp.RawResponse.Body != nil {
				_ = resp.RawResponse.Body.Close()
			}
		}()

		if resp == nil || resp.RawResponse == nil {
			errCh <- fmt.Errorf("SSE connect failed: empty response")
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if resp.StatusCode() >= 300 {
			b, _ := io.ReadAll(resp.RawResponse.Body)
			errCh <- fmt.Errorf("SSE connect failed: [%d] %s", resp.StatusCode(), string(b))
			return
		}

		// 握手成功：唤醒外部调用方，外部可以开始从 sseCh 消费 data 行。
		close(connected)
		streamErr.set(r.readSSEStream(ctx, resp.RawResponse.Body, sseCh))
	}()

	select {
	case err := <-errCh:
		return nil, nil, err
	case <-connected:
		return sseCh, streamErr, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}

// streamErrHolder 由 readSSEStream 写入、forwardSSEStream 在 sseCh 关闭后读取。
// 同一 goroutine 内先 set 再 close(sseCh)，消费者通过 channel close 形成
// happens-before 同步，无需额外锁。
type streamErrHolder struct {
	err error
}

func (h *streamErrHolder) set(err error) { h.err = err }
func (h *streamErrHolder) get() error {
	if h == nil {
		return nil
	}
	return h.err
}

func (r *Runner) readSSEStream(ctx context.Context, body io.ReadCloser, sseCh chan<- string) error {
	scanner := util.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		select {
		case sseCh <- data:
		case <-ctx.Done():
			return nil
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF && err != context.Canceled {
		log.Warnf("%s SSE stream error: %v", r.logPrefix, err)
		return err
	}
	return nil
}

// forwardSSEStream 把 sseCh 中的 SSE data 行转换并转发到 outputCh。
//
// sawFinal / fatalErr 以指针出参形式传入：caller 的 deferred 兜底必须在事件
// 循环结束后读取它们的最终值，以决定是否补发一条 assistant+stop 兜底行。
//
// streamErr 携带 readSSEStream 阶段的底层 IO 错误（如沙箱容器重启导致的
// connection reset / unexpected EOF），在 sseCh 关闭后回填 fatalErr，
// 避免简化错误分类后调用方收到空 content 的兜底消息。
func (r *Runner) forwardSSEStream(ctx context.Context, sseCh <-chan string, streamErr *streamErrHolder,
	outputCh chan<- string, sawFinal *bool, fatalErr *error) {
	for {
		select {
		case <-ctx.Done():
			if fatalErr != nil && *fatalErr == nil {
				*fatalErr = ctx.Err()
			}
			return
		case data, ok := <-sseCh:
			if !ok {
				if fatalErr != nil && *fatalErr == nil {
					if err := streamErr.get(); err != nil {
						*fatalErr = err
					}
				}
				return
			}
			if line := convertEvent(data); line != "" {
				if sawFinal != nil && !*sawFinal && isFinalStopLine(line) {
					*sawFinal = true
				}
				select {
				case outputCh <- line:
				case <-ctx.Done():
					if fatalErr != nil && *fatalErr == nil {
						*fatalErr = ctx.Err()
					}
					return
				}
			}
			if isDoneEvent(data) {
				return
			}
		}
	}
}

// convertEvent 过滤掉控制类事件（done / error），其余 data 行原样转发。
func convertEvent(data string) string {
	var event struct {
		Role string `json:"role"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return ""
	}
	if event.Role == "done" || event.Role == "error" {
		return ""
	}
	return data
}

func isDoneEvent(data string) bool {
	var event struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return false
	}
	return event.Role == "done"
}

// isFinalStopLine 用最小化解析判断 SSE data 行是否已经携带
// Role=assistant + ResponseMeta.FinishReason="stop" 的消息。
// 与 wga-sandbox-converter/eino_chat_model_converter.go 的解析路径保持一致。
func isFinalStopLine(line string) bool {
	var event struct {
		Output *struct {
			MessageOutput *struct {
				Message *struct {
					Role         string `json:"role"`
					ResponseMeta *struct {
						FinishReason string `json:"finish_reason"`
					} `json:"response_meta"`
				} `json:"Message"`
			} `json:"MessageOutput"`
		} `json:"Output"`
	}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return false
	}
	if event.Output == nil || event.Output.MessageOutput == nil || event.Output.MessageOutput.Message == nil {
		return false
	}
	msg := event.Output.MessageOutput.Message
	if msg.Role != "assistant" {
		return false
	}
	if msg.ResponseMeta == nil {
		return false
	}
	return msg.ResponseMeta.FinishReason == shared.FinishReasonStop
}

// buildFinalSSELine 构造一行可写入 outputCh 的 SSE payload，
// 内容形态与 agent/shared/sse.go:marshalAgentEvent 输出保持一致：
// 直接 json.Marshal(*adk.AgentEvent)，下游 eino_chat_model_converter 即能解析。
func buildFinalSSELine(err error) string {
	event := shared.BuildFinalAgentEvent(shared.FinalErrorSourceRunner, err)
	data, mErr := json.Marshal(event)
	if mErr != nil {
		log.Warnf("[eino] failed to marshal final fallback event: %v", mErr)
		return ""
	}
	return string(data)
}
