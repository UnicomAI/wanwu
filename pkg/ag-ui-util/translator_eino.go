package ag_ui_util

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// AgentActivitySimple 表示单个智能体的活动状态。
type AgentActivitySimple struct {
	*MessageState
	activityType               string
	activityID                 string
	agentName                  string
	toolCallStarted            map[string]bool
	currentStreamingToolCallID string
	streamingToolCallArgs      map[string]string
	streamingToolCalls         map[string]bool
}

const (
	DefaultAgentName     = "default"
	subAgentActivityType = "sub_agent"

	ExtraKeyQuestionType = "questionType"
	ExtraKeyQuestionData = "questionData"
)

func NewAgentActivitySimple(agentName string) *AgentActivitySimple {
	return &AgentActivitySimple{
		MessageState:          NewMessageState(),
		activityType:          subAgentActivityType,
		activityID:            aguievents.GenerateStepID(),
		agentName:             agentName,
		toolCallStarted:       make(map[string]bool),
		streamingToolCallArgs: make(map[string]string),
		streamingToolCalls:    make(map[string]bool),
	}
}

// EinoTranslator 将 eino AgentEvent 转换为 AG-UI 事件，用于多智能体场景。
// 通过 ACTIVITY_SNAPSHOT 事件标识当前运行的智能体。
//
// ActivitySnapshot 结构示例：
//
//	{"type":"ACTIVITY_SNAPSHOT","messageId":"step-xxx","activityType":"sub_agent",
//	 "content":{"agentName":"Plan Agent","instanceNum":1,"status":"started"}}
type EinoTranslator struct {
	threadID           string
	runID              string
	runStarted         bool
	runFinished        bool
	agentActivities    []*AgentActivitySimple
	currentActivity    *AgentActivitySimple
	agentInstanceCount map[string]int
}

// NewEinoTranslator 创建多智能体转换器。
func NewEinoTranslator(threadID, runID string) *EinoTranslator {
	return &EinoTranslator{
		threadID:           threadID,
		runID:              runID,
		agentActivities:    make([]*AgentActivitySimple, 0),
		agentInstanceCount: make(map[string]int),
	}
}

func (t *EinoTranslator) TranslateStream(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent]) <-chan aguievents.Event {
	out := make(chan aguievents.Event, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		defer func() {
			for _, evt := range t.finishAllAgents() {
				select {
				case out <- evt:
				case <-ctx.Done():
					return
				}
			}
		}()

		select {
		case out <- aguievents.NewRunStartedEvent(t.threadID, t.runID):
			t.runStarted = true
		case <-ctx.Done():
			return
		}

		for {
			event, ok := iter.Next()
			if !ok {
				return
			}

			agentName := event.AgentName
			if agentName == "" {
				agentName = DefaultAgentName
			}

			shouldSwitch := t.currentActivity == nil || t.currentActivity.agentName != agentName
			if shouldSwitch {
				for _, evt := range t.switchAgent(agentName) {
					select {
					case out <- evt:
					case <-ctx.Done():
						return
					}
				}
			}

			if event.Err != nil {
				errMsg := &schema.Message{
					Role:    schema.Assistant,
					Content: "[error] " + strings.ReplaceAll(event.Err.Error(), "%!s(<nil>)", ""),
				}
				for _, evt := range t.translateMessageForCurrentAgent(errMsg) {
					select {
					case out <- evt:
					case <-ctx.Done():
						return
					}
				}
				// 错误事件后不会有后续事件（eino adk 在发送错误事件后会关闭迭代器）
				return
			}

			if event.Action != nil && event.Action.Exit {
				return
			}

			if event.Output == nil || event.Output.MessageOutput == nil {
				continue
			}

			msgOutput := event.Output.MessageOutput

			if msgOutput.IsStreaming {
				t.translateStreamForAgent(ctx, msgOutput, out)
			} else {
				for _, evt := range t.translateMessageForCurrentAgent(msgOutput.Message) {
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

func (t *EinoTranslator) switchAgent(newAgent string) []aguievents.Event {
	var events []aguievents.Event

	if t.currentActivity != nil {
		events = append(events, t.endCurrentAgentActivity()...)
	}

	t.agentInstanceCount[newAgent]++
	instanceNum := t.agentInstanceCount[newAgent]

	activity := NewAgentActivitySimple(newAgent)
	t.agentActivities = append(t.agentActivities, activity)

	content := map[string]interface{}{
		"agentName":   newAgent,
		"instanceNum": instanceNum,
		"status":      "started",
	}
	events = append(events, aguievents.NewActivitySnapshotEvent(
		activity.activityID,
		activity.activityType,
		content,
	))

	t.currentActivity = activity

	return events
}

func (t *EinoTranslator) endCurrentAgentActivity() []aguievents.Event {
	if t.currentActivity == nil {
		return nil
	}

	activity := t.currentActivity
	var events []aguievents.Event

	// 结束所有活跃的消息
	events = append(events, activity.EndAll()...)

	content := map[string]interface{}{
		"agentName": activity.agentName,
		"status":    "finished",
	}
	events = append(events, aguievents.NewActivitySnapshotEvent(
		activity.activityID,
		activity.activityType,
		content,
	))

	return events
}

func (t *EinoTranslator) translateMessageForCurrentAgent(msg *schema.Message) []aguievents.Event {
	if msg == nil {
		return nil
	}

	if msg.Extra != nil {
		if questionType, ok := msg.Extra[ExtraKeyQuestionType].(string); ok && questionType == "question" {
			if questionData, ok := msg.Extra[ExtraKeyQuestionData].(map[string]any); ok {
				return t.translateQuestionEvent(questionData)
			}
		}
	}

	activity := t.currentActivity
	if activity == nil {
		activity = NewAgentActivitySimple(DefaultAgentName)
	}
	return translateMessageWithActivity(msg, activity, false)
}

func (t *EinoTranslator) translateQuestionEvent(data map[string]any) []aguievents.Event {
	questionID, _ := data["questionId"].(string)
	status, _ := data["status"].(string)
	questions := data["questions"]
	answers, _ := data["answers"].([][]string)

	content := map[string]any{
		"questionId": questionID,
		"runId":      t.runID,
		"threadId":   t.threadID,
		"status":     status,
		"questions":  questions,
		"timestamp":  time.Now().UnixMilli(),
	}
	if len(answers) > 0 {
		content["answers"] = answers
	}

	var events []aguievents.Event
	if t.currentActivity != nil {
		events = append(events, t.currentActivity.EndAll()...)
	}
	events = append(events, aguievents.NewActivitySnapshotEvent(
		aguievents.GenerateStepID(),
		ActivityTypeQuestion,
		content,
	))
	return events
}

func (t *EinoTranslator) translateStreamForAgent(ctx context.Context, msgOutput *adk.MessageVariant, out chan<- aguievents.Event) {
	if msgOutput.MessageStream == nil {
		return
	}
	defer msgOutput.MessageStream.Close()

	if t.currentActivity == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		frame, err := msgOutput.MessageStream.Recv()
		if err == io.EOF {
			for toolCallID := range t.currentActivity.streamingToolCalls {
				args := t.currentActivity.streamingToolCallArgs[toolCallID]
				if !t.currentActivity.toolCallStarted[toolCallID] {
					continue
				}
				if args != "" {
					select {
					case out <- aguievents.NewToolCallArgsEvent(toolCallID, args):
					case <-ctx.Done():
						return
					}
				}
				select {
				case out <- aguievents.NewToolCallEndEvent(toolCallID):
				case <-ctx.Done():
					return
				}
				delete(t.currentActivity.toolCallStarted, toolCallID)
				delete(t.currentActivity.streamingToolCallArgs, toolCallID)
				delete(t.currentActivity.streamingToolCalls, toolCallID)
			}
			return
		}
		if err != nil {
			return
		}

		for _, evt := range translateMessageWithActivity(frame, t.currentActivity, true) {
			select {
			case out <- evt:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (t *EinoTranslator) finishAllAgents() []aguievents.Event {
	var events []aguievents.Event

	if t.currentActivity != nil {
		events = append(events, t.endCurrentAgentActivity()...)
	}

	if !t.runFinished {
		t.runFinished = true
		events = append(events, aguievents.NewRunFinishedEvent(t.threadID, t.runID))
	}

	return events
}

// --- function ---

func translateMessageWithActivity(msg *schema.Message, activity *AgentActivitySimple, isStreaming bool) []aguievents.Event {
	if msg == nil {
		return nil
	}

	var events []aguievents.Event

	if msg.Role == schema.Tool && msg.ToolCallID != "" {
		events = append(events, activity.EndAll()...)
		toolResultMessageID := aguievents.GenerateMessageID()
		events = append(events, aguievents.NewToolCallResultEvent(toolResultMessageID, msg.ToolCallID, msg.Content))
		return events
	}

	hasContent := msg.Content != "" || msg.ReasoningContent != "" || len(msg.ToolCalls) > 0
	if !hasContent {
		return nil
	}

	if len(msg.ToolCalls) > 0 {
		parentMsgID := activity.TextMsgID()
		events = append(events, activity.EndAll()...)

		for _, tc := range msg.ToolCalls {

			toolCallID := tc.ID
			if toolCallID == "" && activity.currentStreamingToolCallID != "" {
				toolCallID = activity.currentStreamingToolCallID
			}

			if toolCallID == "" {
				continue
			}

			if tc.ID != "" && tc.Function.Name != "" {
				activity.currentStreamingToolCallID = tc.ID
				if !activity.toolCallStarted[tc.ID] {
					events = append(events, aguievents.NewToolCallStartEvent(tc.ID, tc.Function.Name, aguievents.WithParentMessageID(parentMsgID)))
					activity.toolCallStarted[tc.ID] = true
				}
			}

			if tc.Function.Arguments != "" {
				activity.streamingToolCallArgs[toolCallID] += tc.Function.Arguments
				if !isStreaming {
					if activity.toolCallStarted[toolCallID] {
						events = append(events, aguievents.NewToolCallArgsEvent(toolCallID, activity.streamingToolCallArgs[toolCallID]))
						events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
						delete(activity.toolCallStarted, toolCallID)
						delete(activity.streamingToolCallArgs, toolCallID)
					}
				} else {
					activity.streamingToolCalls[toolCallID] = true
				}
			}
		}
	}

	if msg.ReasoningContent != "" {
		events = append(events, activity.EndTextMessage()...)
		events = append(events, activity.StartReasoning()...)
		events = append(events, activity.StartReasoningMessage()...)
		events = append(events, aguievents.NewReasoningMessageContentEvent(activity.ReasoningMsgID(), msg.ReasoningContent))
	}

	if msg.Content != "" {
		events = append(events, activity.EndReasoningMessage()...)
		events = append(events, activity.EndReasoning()...)
		events = append(events, activity.StartTextMessage()...)
		events = append(events, aguievents.NewTextMessageContentEvent(activity.TextMsgID(), msg.Content))
	}

	return events
}
