package wga_sandbox

import (
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/opencode"
	sdk "github.com/sst/opencode-sdk-go"
)

// OpencodeEventType opencode 事件类型。
type OpencodeEventType = opencode.OpencodeEventType

// opencode 事件类型常量。
const (
	OpencodeEventTypeStepStart  OpencodeEventType = opencode.OpencodeEventTypeStepStart
	OpencodeEventTypeStepFinish OpencodeEventType = opencode.OpencodeEventTypeStepFinish
	OpencodeEventTypeText       OpencodeEventType = opencode.OpencodeEventTypeText
	OpencodeEventTypeToolUse    OpencodeEventType = opencode.OpencodeEventTypeToolUse
	OpencodeEventTypeReasoning  OpencodeEventType = opencode.OpencodeEventTypeReasoning
	OpencodeEventTypeSnapshot   OpencodeEventType = opencode.OpencodeEventTypeSnapshot
	OpencodeEventTypePatch      OpencodeEventType = opencode.OpencodeEventTypePatch
	OpencodeEventTypeFile       OpencodeEventType = opencode.OpencodeEventTypeFile
	OpencodeEventTypeAgent      OpencodeEventType = opencode.OpencodeEventTypeAgent
	OpencodeEventTypeRetry      OpencodeEventType = opencode.OpencodeEventTypeRetry
	OpencodeEventTypeSubtask    OpencodeEventType = opencode.OpencodeEventTypeSubtask
	OpencodeEventTypeCompaction OpencodeEventType = opencode.OpencodeEventTypeCompaction
	OpencodeEventTypeError      OpencodeEventType = opencode.OpencodeEventTypeError

	// Question 事件类型（Human-in-the-Loop）
	OpencodeEventTypeQuestionAsked    OpencodeEventType = opencode.OpencodeEventTypeQuestionAsked
	OpencodeEventTypeQuestionReplied  OpencodeEventType = opencode.OpencodeEventTypeQuestionReplied
	OpencodeEventTypeQuestionRejected OpencodeEventType = opencode.OpencodeEventTypeQuestionRejected
)

// OpencodeEvent opencode 事件结构。
type OpencodeEvent = opencode.OpencodeEvent

// ParseOpencodeEvent 解析 opencode JSON 输出为事件结构。
func ParseOpencodeEvent(data []byte) (*OpencodeEvent, error) {
	var event OpencodeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// ParseOpencodeTextPart 解析文本类型事件内容。
func ParseOpencodeTextPart(data []byte) (*sdk.TextPart, error) {
	var part sdk.TextPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeToolPart 解析工具调用类型事件内容。
func ParseOpencodeToolPart(data []byte) (*sdk.ToolPart, error) {
	var part sdk.ToolPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeReasoningPart 解析推理类型事件内容。
func ParseOpencodeReasoningPart(data []byte) (*sdk.ReasoningPart, error) {
	var part sdk.ReasoningPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeStepStartPart 解析步骤开始类型事件内容。
func ParseOpencodeStepStartPart(data []byte) (*sdk.StepStartPart, error) {
	var part sdk.StepStartPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeStepFinishPart 解析步骤结束类型事件内容。
func ParseOpencodeStepFinishPart(data []byte) (*sdk.StepFinishPart, error) {
	var part sdk.StepFinishPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeFilePart 解析文件类型事件内容。
func ParseOpencodeFilePart(data []byte) (*sdk.FilePart, error) {
	var part sdk.FilePart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeSnapshotPart 解析快照类型事件内容。
func ParseOpencodeSnapshotPart(data []byte) (*sdk.SnapshotPart, error) {
	var part sdk.SnapshotPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeAgentPart 解析代理类型事件内容。
func ParseOpencodeAgentPart(data []byte) (*sdk.AgentPart, error) {
	var part sdk.AgentPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodePartPatchPart 解析补丁类型事件内容。
func ParseOpencodePartPatchPart(data []byte) (*sdk.PartPatchPart, error) {
	var part sdk.PartPatchPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodePartRetryPart 解析重试类型事件内容。
func ParseOpencodePartRetryPart(data []byte) (*sdk.PartRetryPart, error) {
	var part sdk.PartRetryPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeErrorPart 解析错误类型事件内容。
func ParseOpencodeErrorPart(data []byte) (*opencode.ErrorPart, error) {
	var part opencode.ErrorPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}

// ParseOpencodeQuestionPart 解析问题类型事件内容。
func ParseOpencodeQuestionPart(data []byte) (*opencode.QuestionPart, error) {
	var part opencode.QuestionPart
	if err := json.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	return &part, nil
}
