package service

type AgentChatReq struct {
	Input          string   `json:"input"`
	Stream         bool     `json:"stream"`
	UploadFile     []string `json:"uploadFile"`
	AssistantId    string   `json:"assistantId"`
	ConversationId string   `json:"conversationId"`
	UserId         string   `json:"userId"`
	OrgId          string   `json:"orgId"`
	Draft          bool     `json:"draft"`
	DetailId       string   `json:"detailId"`
}

type MultiAgentChatReq struct {
	Input            string   `json:"input"`
	Stream           bool     `json:"stream"`
	UploadFile       []string `json:"uploadFile"`
	MultiAssistantId string   `json:"multiAgentId"`
	ConversationId   string   `json:"conversationId"`
	UserId           string   `json:"userId"`
	OrgId            string   `json:"orgId"`
	Draft            bool     `json:"draft"`
	DetailId         string   `json:"detailId"`
}

type AgentUserInputParams struct {
	Draft          bool
	Input          string
	Stream         bool
	UploadFile     []string
	ConversationId string
	UserId         string
	OrgId          string
	DetailId       string
}

func BuildAgentChatReq(agentUserInputParams *AgentUserInputParams, assistantId string) *AgentChatReq {
	var req = &AgentChatReq{
		Input:          agentUserInputParams.Input,
		Stream:         agentUserInputParams.Stream,
		UploadFile:     agentUserInputParams.UploadFile,
		AssistantId:    assistantId,
		ConversationId: agentUserInputParams.ConversationId,
		UserId:         agentUserInputParams.UserId,
		OrgId:          agentUserInputParams.OrgId,
		Draft:          agentUserInputParams.Draft,
		DetailId:       agentUserInputParams.DetailId,
	}
	return req
}

func BuildMultiAgentChatReq(agentUserInputParams *AgentUserInputParams, assistantId string) *MultiAgentChatReq {
	var req = &MultiAgentChatReq{
		Input:            agentUserInputParams.Input,
		Stream:           agentUserInputParams.Stream,
		UploadFile:       agentUserInputParams.UploadFile,
		MultiAssistantId: assistantId,
		ConversationId:   agentUserInputParams.ConversationId,
		UserId:           agentUserInputParams.UserId,
		OrgId:            agentUserInputParams.OrgId,
		Draft:            agentUserInputParams.Draft,
		DetailId:         agentUserInputParams.DetailId,
	}
	return req
}
