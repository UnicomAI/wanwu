package ag_ui_util

import (
	"context"
	"encoding/json"
	"time"

	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/sst/opencode-sdk-go"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type opencodeEventType string

const (
	opencodeEventTypeText             opencodeEventType = "text"
	opencodeEventTypeToolUse          opencodeEventType = "tool_use"
	opencodeEventTypeReasoning        opencodeEventType = "reasoning"
	opencodeEventTypeQuestionAsked    opencodeEventType = "question.asked"
	opencodeEventTypeQuestionReplied  opencodeEventType = "question.replied"
	opencodeEventTypeQuestionRejected opencodeEventType = "question.rejected"
)

type opencodeEvent struct {
	Type opencodeEventType `json:"type"`
	Part json.RawMessage   `json:"part"`
}

type opencodeErrorPart struct {
	Error struct {
		Name string `json:"name"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	} `json:"error"`
}

// OpencodeTranslator 将 opencode JSON 事件流转换为 AG-UI 事件。
// 用于 wga-sandbox 输出转换。
//
// opencode 事件类型映射：
//   - "text" -> TEXT_MESSAGE_CONTENT
//   - "reasoning" -> REASONING_MESSAGE_CONTENT
//   - "tool_use" -> TOOL_CALL_START/ARGS/END/RESULT
//   - "error" -> TEXT_MESSAGE_CONTENT（带 [error] 前缀）
//
// AG-UI 协议要求：
//   - ToolCall 使用 parentMessageId 关联到当前文本消息
//   - ToolResult 使用新的 messageId
//   - Reasoning 在 TextMessage 之前发送，使用独立的消息流
type OpencodeTranslator struct {
	BaseState
	activeToolCalls map[string]bool
}

// NewOpencodeTranslator 创建 opencode 转换器。
func NewOpencodeTranslator(threadID, runID string) *OpencodeTranslator {
	return &OpencodeTranslator{
		BaseState:       NewBaseState(threadID, runID),
		activeToolCalls: make(map[string]bool),
	}
}

func (t *OpencodeTranslator) TranslateStream(ctx context.Context, in <-chan string) <-chan aguievents.Event {
	out := make(chan aguievents.Event, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		defer func() {
			for _, evt := range t.FinishBase() {
				select {
				case out <- evt:
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				for _, evt := range t.translate(ctx, line) {
					select {
					case out <- evt:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}

func (t *OpencodeTranslator) translate(_ context.Context, line string) []aguievents.Event {
	var evt opencodeEvent
	if err := json.Unmarshal([]byte(line), &evt); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse opencode event: %v", t.RunID(), err)
		return nil
	}

	if t.MessageID() == "" {
		t.SetMessageID(aguievents.GenerateMessageID())
	}

	var events []aguievents.Event

	switch evt.Type {
	case opencodeEventTypeText:
		events = t.translateText(evt.Part)
	case opencodeEventTypeToolUse:
		events = t.translateToolUse(evt.Part)
	case opencodeEventTypeReasoning:
		events = t.translateReasoning(evt.Part)
	case opencodeEventTypeQuestionAsked, opencodeEventTypeQuestionReplied, opencodeEventTypeQuestionRejected:
		events = t.translateQuestion(evt.Part, string(evt.Type))
	case "error":
		events = t.translateError(evt.Part)
	}

	return events
}

func (t *OpencodeTranslator) translateText(partData json.RawMessage) []aguievents.Event {
	var part opencode.TextPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse text part: %v", t.RunID(), err)
		return nil
	}

	if part.Text == "" {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.EndReasoningMessage()...)
	events = append(events, t.EndReasoning()...)
	events = append(events, t.StartTextMessage()...)
	events = append(events, aguievents.NewTextMessageContentEvent(t.MessageID(), part.Text))
	return events
}

func (t *OpencodeTranslator) translateReasoning(partData json.RawMessage) []aguievents.Event {
	var part opencode.ReasoningPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse reasoning part: %v", t.RunID(), err)
		return nil
	}

	if part.Text == "" {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.EndTextMessage()...)
	events = append(events, t.StartReasoning()...)
	events = append(events, t.StartReasoningMessage()...)
	events = append(events, aguievents.NewReasoningMessageContentEvent(t.ReasoningMessageID(), part.Text))
	return events
}

func (t *OpencodeTranslator) translateError(partData json.RawMessage) []aguievents.Event {
	var part opencodeErrorPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse error part: %v", t.RunID(), err)
		return nil
	}

	msg := "[error] " + part.Error.Name
	if part.Error.Data.Message != "" {
		msg += ": " + part.Error.Data.Message
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.EndReasoningMessage()...)
	events = append(events, t.EndReasoning()...)
	events = append(events, t.StartTextMessage()...)
	events = append(events, aguievents.NewTextMessageContentEvent(t.MessageID(), msg))
	return events
}

func (t *OpencodeTranslator) translateToolUse(partData json.RawMessage) []aguievents.Event {
	var part opencode.ToolPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse tool_use part: %v", t.RunID(), err)
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)

	parentMsgID := t.MessageID()
	events = append(events, t.EndAll()...)

	toolCallID := part.CallID
	if toolCallID == "" {
		toolCallID = part.ID
	}

	switch part.State.Status {
	case opencode.ToolPartStateStatusPending, opencode.ToolPartStateStatusRunning:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(parentMsgID)))
			t.activeToolCalls[toolCallID] = true
			if input := t.getToolInput(part.State); input != "" {
				events = append(events, aguievents.NewToolCallArgsEvent(toolCallID, input))
			}
		}

	case opencode.ToolPartStateStatusCompleted:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(parentMsgID)))
			if input := t.getToolInput(part.State); input != "" {
				events = append(events, aguievents.NewToolCallArgsEvent(toolCallID, input))
			}
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
		} else {
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
			delete(t.activeToolCalls, toolCallID)
		}
		resultMessageID := aguievents.GenerateMessageID()
		events = append(events, aguievents.NewToolCallResultEvent(resultMessageID, toolCallID, part.State.Output))

	case opencode.ToolPartStateStatusError:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(parentMsgID)))
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
		} else {
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
			delete(t.activeToolCalls, toolCallID)
		}
		resultMessageID := aguievents.GenerateMessageID()
		events = append(events, aguievents.NewToolCallResultEvent(resultMessageID, toolCallID, part.State.Error))
	}

	return events
}

func (t *OpencodeTranslator) getToolInput(state opencode.ToolPartState) string {
	if state.Input == nil {
		return ""
	}
	switch v := state.Input.(type) {
	case string:
		return v
	case map[string]interface{}:
		data, _ := json.Marshal(v)
		return string(data)
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}

type questionPart struct {
	Type       string         `json:"type"`
	QuestionID string         `json:"questionId"`
	SessionID  string         `json:"sessionID"`
	Status     string         `json:"status,omitempty"`
	Questions  []questionItem `json:"questions"`
	Answers    [][]string     `json:"answers,omitempty"`
}

type questionItem struct {
	Question string           `json:"question"`
	Header   string           `json:"header"`
	Options  []questionOption `json:"options"`
	Multiple bool             `json:"multiple"`
	Custom   bool             `json:"custom"`
}

type questionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

func (t *OpencodeTranslator) translateQuestion(partData json.RawMessage, eventType string) []aguievents.Event {
	var part questionPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse question part: %v", t.RunID(), err)
		return nil
	}

	status := part.Status
	if status == "" {
		switch eventType {
		case string(opencodeEventTypeQuestionAsked):
			status = "pending"
		case string(opencodeEventTypeQuestionReplied):
			status = "answered"
		case string(opencodeEventTypeQuestionRejected):
			status = "rejected"
		}
	}

	questions := make([]any, 0, len(part.Questions))
	for _, q := range part.Questions {
		options := make([]any, 0, len(q.Options))
		for _, opt := range q.Options {
			options = append(options, map[string]any{
				"label":       opt.Label,
				"description": opt.Description,
			})
		}
		questions = append(questions, map[string]any{
			"question": q.Question,
			"header":   q.Header,
			"options":  options,
			"multiple": q.Multiple,
			"custom":   q.Custom,
		})
	}

	content := map[string]any{
		"questionId": part.QuestionID,
		"runId":      t.RunID(),
		"threadId":   t.ThreadID(),
		"status":     status,
		"questions":  questions,
		"timestamp":  time.Now().UnixMilli(),
	}
	if len(part.Answers) > 0 {
		content["answers"] = part.Answers
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.EndAll()...)
	events = append(events, aguievents.NewActivitySnapshotEvent(
		aguievents.GenerateStepID(),
		ActivityTypeQuestion,
		content,
	))
	return events
}
