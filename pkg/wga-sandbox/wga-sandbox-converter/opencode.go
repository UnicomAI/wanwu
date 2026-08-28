package wga_sandbox_converter

import (
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/log"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	"github.com/cloudwego/eino/schema"
)

type opencodeConverter struct {
	handlers map[wga_sandbox.OpencodeEventType]partParser
}

func newOpencodeConverter() *opencodeConverter {
	return &opencodeConverter{
		handlers: map[wga_sandbox.OpencodeEventType]partParser{
			wga_sandbox.OpencodeEventTypeStepStart:        parseSkipPart,
			wga_sandbox.OpencodeEventTypeStepFinish:       parseSkipPart,
			wga_sandbox.OpencodeEventTypeText:             parseTextPart,
			wga_sandbox.OpencodeEventTypeReasoning:        parseReasoningPart,
			wga_sandbox.OpencodeEventTypeToolUse:          parseToolUsePart,
			wga_sandbox.OpencodeEventTypeFile:             parseFilePart,
			wga_sandbox.OpencodeEventTypeSnapshot:         parseSnapshotPart,
			wga_sandbox.OpencodeEventTypeAgent:            parseAgentPart,
			wga_sandbox.OpencodeEventTypePatch:            parsePatchPart,
			wga_sandbox.OpencodeEventTypeRetry:            parseRetryPart,
			wga_sandbox.OpencodeEventTypeError:            parseErrorPart,
			wga_sandbox.OpencodeEventTypeQuestionAsked:    parseQuestionPart,
			wga_sandbox.OpencodeEventTypeQuestionReplied:  parseQuestionPart,
			wga_sandbox.OpencodeEventTypeQuestionRejected: parseQuestionPart,
		},
	}
}

func (c *opencodeConverter) Convert(line string) ([]*schema.Message, error) {
	event, err := wga_sandbox.ParseOpencodeEvent([]byte(line))
	if err != nil {
		return nil, err
	}

	handler := c.handlers[event.Type]
	if handler == nil {
		log.Warnf("[converter][opencode] unknown event type: %s", event.Type)
		return nil, nil
	}

	content, err := handler(event.Part)
	if err != nil {
		return nil, err
	}
	if content.skip {
		return nil, nil
	}

	var messages []*schema.Message

	if len(content.toolCalls) > 0 || content.content != "" || content.reasoningContent != "" {
		msg := &schema.Message{
			Role:             schema.Assistant,
			Content:          content.content,
			ReasoningContent: content.reasoningContent,
			ToolCalls:        content.toolCalls,
		}
		if content.extra != nil {
			msg.Extra = content.extra
		}
		messages = append(messages, msg)
	}

	if content.toolResult != nil {
		messages = append(messages, &schema.Message{
			Role:       schema.Tool,
			ToolCallID: content.toolResult.callID,
			Content:    content.toolResult.content,
		})
	}

	return messages, nil
}

type messageContent struct {
	content          string
	reasoningContent string
	toolCalls        []schema.ToolCall
	toolResult       *toolResult
	skip             bool
	extra            map[string]any
}

type toolResult struct {
	callID  string
	content string
}

type partParser func(json.RawMessage) (messageContent, error)

func parseSkipPart(_ json.RawMessage) (messageContent, error) {
	return messageContent{skip: true}, nil
}

func parseTextPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeTextPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: p.Text}, nil
}

func parseReasoningPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeReasoningPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{reasoningContent: p.Text}, nil
}

func parseToolUsePart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeToolPart(part)
	if err != nil {
		return messageContent{}, err
	}
	input := ""
	if p.State.Input != nil {
		data, _ := json.Marshal(p.State.Input)
		input = string(data)
	}

	result := messageContent{
		toolCalls: []schema.ToolCall{
			{
				ID:   p.CallID,
				Type: "function",
				Function: schema.FunctionCall{
					Name:      p.Tool,
					Arguments: input,
				},
			},
		},
	}

	switch p.State.Status {
	case "completed":
		output := p.State.Output
		if output == "" {
			output = "success"
		}
		result.toolResult = &toolResult{
			callID:  p.CallID,
			content: output,
		}
	case "error":
		if p.State.Error != "" {
			result.toolResult = &toolResult{
				callID:  p.CallID,
				content: p.State.Error,
			}
		}
	}

	return result, nil
}

func parseFilePart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeFilePart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[file] " + p.Filename}, nil
}

func parseSnapshotPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeSnapshotPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[snapshot] " + p.ID}, nil
}

func parseAgentPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeAgentPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[agent] " + p.Name}, nil
}

func parsePatchPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodePartPatchPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[patch] " + p.ID}, nil
}

func parseRetryPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodePartRetryPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[retry] " + p.ID}, nil
}

func parseErrorPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeErrorPart(part)
	if err != nil {
		return messageContent{}, err
	}
	msg := "[error] " + p.Error.Name
	if p.Error.Data.Message != "" {
		msg += ": " + p.Error.Data.Message
	}
	return messageContent{content: msg}, nil
}

func parseQuestionPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeQuestionPart(part)
	if err != nil {
		return messageContent{}, err
	}

	questions := make([]map[string]any, 0, len(p.Questions))
	for _, q := range p.Questions {
		options := make([]map[string]any, 0, len(q.Options))
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

	questionData := map[string]any{
		"questionId": p.QuestionID,
		"sessionId":  p.SessionID,
		"status":     p.Status,
		"questions":  questions,
	}
	if len(p.Answers) > 0 {
		questionData["answers"] = p.Answers
	}

	return messageContent{
		content: "[question] " + p.QuestionID,
		extra: map[string]any{
			"questionType": "question",
			"questionData": questionData,
		},
	}, nil
}
