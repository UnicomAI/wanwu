package util

import (
	"encoding/json"

	"github.com/cloudwego/eino/schema"
)

const (
	AgentSearchKnowledgeName = "智能体知识库检索"
	AgentSkillPrefix         = "skill-"
	AgentSkillWgaStop        = "skill_wga_stop" //沙箱输出结束，但未完全输出停止，因为硬编码了结束事件
	AgentStartLabel          = "transfer_to_agent"
	MainAgentExitLabel       = "exit"
)

var stopFinishReason = map[string]bool{
	"stop":   true, //正常停止
	"length": true, //截断停止
}

// StopMessage 判断是否是停止消息
func StopMessage(chatMessage *schema.Message) bool {
	return chatMessage.ResponseMeta != nil && stopFinishReason[chatMessage.ResponseMeta.FinishReason]
}

// WgaStopMessage 判断是否是沙箱停止消息
func WgaStopMessage(chatMessage *schema.Message) bool {
	return chatMessage.ResponseMeta != nil && AgentSkillWgaStop == chatMessage.ResponseMeta.FinishReason
}

func BuildAssistantMessage(content string, extra map[string]any) string {
	message := &schema.Message{
		Role:    schema.Assistant,
		Content: content,
		Extra:   extra,
	}
	marshal, _ := json.Marshal(message)
	return string(marshal)
}

func BuildToolParamsMessage(toolCall []schema.ToolCall) *schema.Message {
	return &schema.Message{ //模拟智能体切换消息
		Role:      schema.Assistant,
		ToolCalls: toolCall,
		ResponseMeta: &schema.ResponseMeta{
			FinishReason: "tool_calls",
		},
	}
}

func BuildToolFinishMessage(content string) string {
	message := &schema.Message{
		Role:    schema.Assistant,
		Content: content,
		ResponseMeta: &schema.ResponseMeta{
			FinishReason: "stop",
		},
	}
	marshal, _ := json.Marshal(message)
	return string(marshal)
}
