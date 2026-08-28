package conversation

import "github.com/UnicomAI/wanwu/internal/assistant-service/client/model"

var knowledge = &Knowledge{}

type Knowledge struct {
}

func init() {
	InitBuilder(knowledge)
}

func (*Knowledge) EventType() int {
	return KnowledgeEventType
}
func (*Knowledge) Build(conversationResp *ConversationResp, conversation, searchResult string, agentChatResp *AgentChatResp) error {
	eventData := agentChatResp.EventData
	if eventData == nil {
		return nil
	}
	resp := conversationResp.ConversationEventMap[eventData.Id]
	if resp == nil {
		resp = CreateConversationResp()
		resp.Order = eventData.Order
		resp.EventType = KnowledgeEventType
		conversationResp.ConversationEventMap[eventData.Id] = resp
	}
	var hasResult = false
	if resp.SearchList == nil && len(searchResult) > 0 {
		resp.SearchList = &searchResult
		hasResult = true
	}
	//终态存储
	if len(conversation) > 0 {
		hasResult = true
		//保存对话
		resp.Write(conversation, eventData.Order)
	}
	//终态存储
	if eventData.Status == model.EventEndStatus {
		if !hasResult {
			eventData.Status = model.EventFailStatus
		}
		resp.EventData = eventData
		resp.Write("", eventData.Order) //这个可能影响历史记录
	}
	return nil
}
