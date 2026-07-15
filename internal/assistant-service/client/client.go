package client

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
)

type IClient interface {
	//================Assistant================
	CreateAssistant(ctx context.Context, assistant *model.Assistant) *err_code.Status
	UpdateAssistant(ctx context.Context, assistant *model.Assistant) *err_code.Status
	DeleteAssistant(ctx context.Context, assistantID uint32) *err_code.Status
	GetAssistant(ctx context.Context, assistantID uint32, userID, orgID string) (*model.Assistant, *err_code.Status)
	GetAssistantsByIDs(ctx context.Context, assistantIDs []uint32) ([]*model.Assistant, *err_code.Status)
	GetAssistantByUuid(ctx context.Context, uuid string) (*model.Assistant, *err_code.Status)
	GetAssistantList(ctx context.Context, userID, orgID string, name string) ([]*model.Assistant, int64, *err_code.Status)
	CheckSameAssistantName(ctx context.Context, userID, orgID, name, assistantID string) *err_code.Status
	CopyAssistant(ctx context.Context, assistant *model.Assistant, workflows []*model.AssistantWorkflow, mcps []*model.AssistantMCP, customTools []*model.AssistantTool, subAgents []*model.MultiAgentRelation, skills []*model.AssistantSkill) (uint32, *err_code.Status)

	//================AssistantSnapshot================
	CreateAssistantSnapshot(ctx context.Context, assistantSnapshot *model.AssistantSnapshot) (uint32, *err_code.Status)
	UpdateAssistantSnapshot(ctx context.Context, assistantID uint32, desc, extra string, userID, orgID string) *err_code.Status
	GetAssistantSnapshotList(ctx context.Context, assistantID uint32, userID, orgID string) ([]*model.AssistantSnapshot, *err_code.Status)
	GetAssistantSnapshot(ctx context.Context, assistantID uint32, version string) (*model.AssistantSnapshot, *err_code.Status)
	GetAssistantSnapshotListByAssistantIds(ctx context.Context, assistantIds []uint32) ([]*model.AssistantSnapshot, *err_code.Status)
	RollbackAssistantSnapshot(ctx context.Context, assistant *model.Assistant, tools []*model.AssistantTool, mcps []*model.AssistantMCP, workflows []*model.AssistantWorkflow, subAgents []*model.MultiAgentRelation, userID, orgID string) *err_code.Status

	//================AssistantWorkflow================
	CreateAssistantWorkflow(ctx context.Context, workflow *model.AssistantWorkflow) *err_code.Status
	DeleteAssistantWorkflow(ctx context.Context, assistantId uint32, workflowId string) *err_code.Status
	UpdateAssistantWorkflow(ctx context.Context, workflow *model.AssistantWorkflow) *err_code.Status
	GetAssistantWorkflow(ctx context.Context, assistantId uint32, workflowId string) (*model.AssistantWorkflow, *err_code.Status)
	GetAssistantWorkflowsByAssistantID(ctx context.Context, assistantId uint32) ([]*model.AssistantWorkflow, *err_code.Status)
	DeleteAssistantWorkflowByWorkflowId(ctx context.Context, workflowId string) *err_code.Status

	//================AssistantMCP================
	CreateAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string, userId, orgID string) *err_code.Status
	DeleteAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string) *err_code.Status
	GetAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string) (*model.AssistantMCP, *err_code.Status)
	DeleteAssistantMCPByMCPId(ctx context.Context, mcpId string, mcpType string) *err_code.Status
	GetAssistantMCPList(ctx context.Context, assistantId uint32) ([]*model.AssistantMCP, *err_code.Status)
	UpdateAssistantMCP(ctx context.Context, mcp *model.AssistantMCP) *err_code.Status

	//================AssistantTool================
	CreateAssistantTool(ctx context.Context, assistantId uint32, toolId, toolType string, actionName string, userId, orgID string) *err_code.Status
	DeleteAssistantTool(ctx context.Context, assistantId uint32, toolId string, toolType string, actionName string) *err_code.Status
	UpdateAssistantTool(ctx context.Context, tool *model.AssistantTool) *err_code.Status
	UpdateAssistantToolConfig(ctx context.Context, assistantId uint32, toolId, toolConfig string) *err_code.Status
	GetAssistantTool(ctx context.Context, assistantId uint32, toolId, toolType string, actionName string) (*model.AssistantTool, *err_code.Status)
	GetAssistantToolList(ctx context.Context, assistantId uint32) ([]*model.AssistantTool, *err_code.Status)
	DeleteAssistantToolByToolId(ctx context.Context, toolId string, toolType string) *err_code.Status

	//================Assistant Skill================
	CreateAssistantSkill(ctx context.Context, assistantId uint32, skillId, skillType, userId, orgId string) *err_code.Status
	DeleteAssistantSkill(ctx context.Context, assistantId uint32, skillId, skillType string) *err_code.Status
	GetAssistantSkillById(ctx context.Context, assistantId uint32, skillId, skillType string) (*model.AssistantSkill, *err_code.Status)
	GetAssistantSkillList(ctx context.Context, assistantId uint32) ([]*model.AssistantSkill, *err_code.Status)
	UpdateAssistantSkillEnable(ctx context.Context, assistantId uint32, skillId, skillType string, enable bool) *err_code.Status

	//================Conversation================
	CreateConversation(ctx context.Context, conversation *model.Conversation) *err_code.Status
	UpdateConversation(ctx context.Context, conversation *model.Conversation) *err_code.Status
	DeleteConversation(ctx context.Context, conversationID uint32) *err_code.Status
	GetConversationByAssistantID(ctx context.Context, assistantID, conversationType string) (*model.Conversation, *err_code.Status)
	GetConversationList(ctx context.Context, assistantID, conversationType, userID, orgID, searchText string, offset, limit int32) ([]*model.Conversation, int64, *err_code.Status)

	//================CustomPrompt================
	CreateCustomPrompt(ctx context.Context, avatarPath, name, desc, prompt, userId, orgID string) (string, *err_code.Status)
	DeleteCustomPrompt(ctx context.Context, customPromptID uint32) *err_code.Status
	UpdateCustomPrompt(ctx context.Context, info *assistant_service.CustomPromptUpdateReq) *err_code.Status
	GetCustomPrompt(ctx context.Context, customPromptID uint32) (*model.CustomPrompt, *err_code.Status)
	GetCustomPromptList(ctx context.Context, userID, orgID string, name string) ([]*model.CustomPrompt, int64, *err_code.Status)
	CopyCustomPrompt(ctx context.Context, customPromptID uint32, userId, orgID string) (string, *err_code.Status)

	//================MultiAssistant================
	GetMultiAssistant(ctx context.Context, multiAssistantID uint32, userID, orgID string, draft bool, version string, filterSubEnable bool) (multiAgent *model.Assistant, multiAgentSnapshot *model.AssistantSnapshot, subAgents []*model.AssistantSnapshot, err error)
	CreateMultiAssistantRelation(ctx context.Context, assistant *model.MultiAgentRelation) *err_code.Status
	FetchMultiAssistantRelationList(ctx context.Context, multiAssistantID uint32, version string, draft bool) ([]*model.MultiAgentRelation, *err_code.Status)
	FetchMultiAssistantRelationFirst(ctx context.Context, multiAssistantID, agentID uint32) (*model.MultiAgentRelation, *err_code.Status)
	DeleteMultiAssistantRelation(ctx context.Context, multiAssistantID, agentID uint32) *err_code.Status
	UpdateMultiAssistantRelation(ctx context.Context, assistant *model.MultiAgentRelation) *err_code.Status
	BatchCreateMultiAssistantRelation(ctx context.Context, assistants []*model.MultiAgentRelation, version string) *err_code.Status

	//=================WgaConversationConfig================
	GetWgaConversationConfig(ctx context.Context, threadId string, userId, orgId string) (*model.WgaConversationConfig, *err_code.Status)
	UpdateWgaConversationConfig(ctx context.Context, config *model.WgaConversationConfig) *err_code.Status
	CreateWgaConversationConfig(ctx context.Context, config *model.WgaConversationConfig) *err_code.Status
	DeleteWgaConversationConfig(ctx context.Context, threadId string) *err_code.Status
	GetWgaConversationConfigList(ctx context.Context, userID, orgID, searchText string, offset, limit int32) ([]*model.WgaConversationConfig, int64, *err_code.Status)
	WgaConversationConfigExists(ctx context.Context, threadId, userID, orgID string) (bool, *err_code.Status)

	//=================WgaConfig================
	GetWgaConfig(ctx context.Context, userId, orgId string) (*model.WgaConfig, *err_code.Status)
	UpdateWgaConfig(ctx context.Context, config *model.WgaConfig) *err_code.Status
}
