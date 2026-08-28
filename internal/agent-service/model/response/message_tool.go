package response

import (
	"sort"

	"github.com/cloudwego/eino/schema"
)

type MessageTool struct {
	ChatMessage *schema.Message
	RespContext *AgentChatRespContext
}

func CreateMessageTool(chatMessage *schema.Message, respContext *AgentChatRespContext) *MessageTool {
	return &MessageTool{
		ChatMessage: chatMessage,
		RespContext: respContext,
	}
}

func (m *MessageTool) ToolMessage() bool {
	return m.ToolStart() || m.ToolParamsEnd() || m.ToolEnd()
}

func (m *MessageTool) ToolStart() bool {
	return len(m.ChatMessage.ToolCalls) > 0
}

func (m *MessageTool) ToolParamsEnd() bool {
	responseMeta := m.ChatMessage.ResponseMeta
	if responseMeta == nil {
		return false
	}
	return responseMeta.FinishReason == "tool_calls"
}

func (m *MessageTool) ToolEnd() bool {
	return m.ChatMessage.Role == schema.Tool
}

// ToolId 构造toolId
// case1:工具同步调用结果，或者模型处理较好会直接返回模型id
// case2:触发了工具的并发调用即，先输出了两此工具参数，此时输出工具调用结果，如果没有toolId就默认按顺序填充结果
// case3:参数输出过程中，或者工具同步调用结果 没有toolId 标识，则返回当前toolId（上次参数输出的toolId）
func (m *MessageTool) ToolId() string {
	if len(m.ChatMessage.ToolCallID) > 0 {
		return m.ChatMessage.ToolCallID
	}
	toolIdList := FilerToolByStep(m.RespContext, ToolResultFinishStep, false)
	if len(toolIdList) > 1 { //此处表示有多个工具并发调用了
		var agentToolList []*AgentTool
		for _, toolId := range toolIdList {
			tool := m.RespContext.AgentToolContext.GetTool(toolId)
			toolIndex := BuildToolIndex(m.ChatMessage)
			if toolIndex != nil && tool.ToolIndex != nil && *toolIndex == *tool.ToolIndex {
				return tool.ToolId
			}
			agentToolList = append(agentToolList, tool)
		}
		sort.Slice(agentToolList, func(i, j int) bool {
			return agentToolList[i].Order > agentToolList[j].Order
		})
		return agentToolList[0].ToolId
	}
	return m.RespContext.AgentToolContext.GetCurrentToolId()
}

func (m *MessageTool) NewTool(tool schema.ToolCall) bool {
	return len(tool.ID) > 0 && !m.RespContext.AgentToolContext.ExistTool(tool.ID)
}
