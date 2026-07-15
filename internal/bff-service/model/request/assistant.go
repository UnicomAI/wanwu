package request

import (
	"errors"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"
)

type AssistantUpdateReq struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	AppBriefConfig
}

func (a *AssistantUpdateReq) Check() error {
	// name/desc 校验在 service 层处理
	return nil
}

type AssistantCreateReq struct {
	Category int `json:"category"` // 1:单智能体 2:多智能体
	AppBriefConfig
}

func (a *AssistantCreateReq) Check() error {
	return util.ValidateBriefCreate(&a.Name, &a.Desc, util.SubjectAssistant)
}

type AssistantConfig struct {
	AssistantId         string                  `json:"assistantId"  validate:"required"`
	Prologue            string                  `json:"prologue"`            // 开场白
	Instructions        string                  `json:"instructions"`        // 系统提示词
	RecommendQuestion   []string                `json:"recommendQuestion"`   // 推荐问题
	ModelConfig         *AppModelConfig         `json:"modelConfig"`         // 模型
	KnowledgeBaseConfig *AppKnowledgebaseConfig `json:"knowledgeBaseConfig"` // 知识库
	SafetyConfig        *AppSafetyConfig        `json:"safetyConfig"`        // 敏感词表配置
	RerankConfig        *AppModelConfig         `json:"rerankConfig"`        // Rerank模型
	VisionConfig        *VisionConfig           `json:"visionConfig"`        // 视觉配置
	MemoryConfig        *MemoryConfig           `json:"memoryConfig"`        // 记忆配置
	RecommendConfig     *RecommendConfig        `json:"recommendConfig"`     // 追问配置
}

type MemoryConfig struct {
	MaxHistoryLength int32 `json:"maxHistoryLength"`
}

func (a *AssistantConfig) Check() error {
	return nil
}

type AssistantPublish struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	Scope       int32  `json:"scope"  validate:"required"`
}

func (a *AssistantPublish) Check() error { return nil }

type AssistantDeleteRequest struct {
	AssistantId string `json:"assistantId"  validate:"required"`
}

func (a *AssistantDeleteRequest) Check() error { return nil }

type AssistantIdRequest struct {
	AssistantId string `json:"assistantId" form:"assistantId" validate:"required"`
	Version     string `json:"version" form:"version"`
}

func (a *AssistantIdRequest) Check() error { return nil }

type AssistantWorkFlowAddRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	WorkFlowId  string `json:"workFlowId" validate:"required"`
}

func (w *AssistantWorkFlowAddRequest) Check() error { return nil }

type AssistantWorkFlowDelRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	WorkFlowId  string `json:"workFlowId" validate:"required"`
}

func (w *AssistantWorkFlowDelRequest) Check() error { return nil }

type AssistantWorkFlowToolEnableRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	WorkFlowId  string `json:"workFlowId" validate:"required"`
	Enable      bool   `json:"enable"`
}

func (w *AssistantWorkFlowToolEnableRequest) Check() error { return nil }

type AssistantMCPToolAddRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	MCPId       string `json:"mcpId" validate:"required"`
	MCPType     string `json:"mcpType" validate:"required,oneof=mcp mcpserver"`
	ActionName  string `json:"actionName" validate:"required"`
}

func (m *AssistantMCPToolAddRequest) Check() error { return nil }

type AssistantMCPToolDelRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	MCPId       string `json:"mcpId" validate:"required"`
	MCPType     string `json:"mcpType" validate:"required,oneof=mcp mcpserver"`
	ActionName  string `json:"actionName" validate:"required"`
}

func (w *AssistantMCPToolDelRequest) Check() error { return nil }

type AssistantMCPToolEnableRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	MCPId       string `json:"mcpId" validate:"required"`
	MCPType     string `json:"mcpType" validate:"required,oneof=mcp mcpserver"`
	ActionName  string `json:"actionName" validate:"required"`
	Enable      bool   `json:"enable"`
}

func (a *AssistantMCPToolEnableRequest) Check() error { return nil }

type AssistantSkillAddRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	SkillId     string `json:"skillId" validate:"required"`
	SkillType   string `json:"skillType" validate:"required,oneof=builtin custom acquired"`
}

func (a *AssistantSkillAddRequest) Check() error { return nil }

type AssistantSkillDelRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	SkillId     string `json:"skillId" validate:"required"`
	SkillType   string `json:"skillType" validate:"required,oneof=builtin custom acquired"`
}

func (a *AssistantSkillDelRequest) Check() error { return nil }

type AssistantSkillEnableSwitchRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	SkillId     string `json:"skillId" validate:"required"`
	SkillType   string `json:"skillType" validate:"required,oneof=builtin custom acquired"`
	Enable      bool   `json:"enable"`
}

func (a *AssistantSkillEnableSwitchRequest) Check() error { return nil }

type ConversationCreateRequest struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	Prompt      string `json:"prompt"  validate:"required"`
}

func (c *ConversationCreateRequest) Check() error { return nil }

type ConversationDeleteRequest struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	DetailId    string `json:"detailId"` // 可选，传值则删除单条对话，不传则删除全部对话
}

func (c *ConversationDeleteRequest) Check() error { return nil }

type ConversationIdRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
	DetailId       string `json:"detailId" form:"detailId"` // 可选，传值则删除单条对话，不传则删除全部对话
}

func (c *ConversationIdRequest) Check() error { return nil }

type ConversationGetListRequest struct {
	AssistantId string `json:"assistantId" form:"assistantId"  validate:"required"`
	PageSize    int    `json:"pageSize" form:"pageSize"  validate:"required"`
	PageNo      int    `json:"pageNo" form:"pageNo"  validate:"required"`
	SearchText  string `json:"searchText" form:"searchText"` // 标题关键词，模糊匹配，空则不过滤
}

func (c *ConversationGetListRequest) Check() error { return nil }

type ConversationGetDetailListRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
	PageSize       int    `json:"pageSize" form:"pageSize"  validate:"required"`
	PageNo         int    `json:"pageNo" form:"pageNo"  validate:"required"`
}

func (c *ConversationGetDetailListRequest) Check() error { return nil }

type ConversionStreamRequest struct {
	AssistantId    string                 `json:"assistantId" form:"assistantId"  validate:"required"`
	ConversationId string                 `json:"conversationId" form:"conversionId"`
	FileInfo       []ConversionStreamFile `json:"fileInfo" form:"fileInfo"`
	Prompt         string                 `json:"prompt" form:"prompt"  validate:"required"`
	SystemPrompt   string                 `json:"systemPrompt" form:"systemPrompt"`
	IsCompare      bool                   `json:"isCompare" form:"isCompare"`
	SseHold        bool                   `json:"sseHold" form:"sseHold"`
}

type PendingConversionRequest struct {
	AssistantId    string `json:"assistantId" form:"assistantId"  validate:"required"`
	Draft          bool   `json:"draft" form:"draft"`                 //是否草稿态
	ConversationId string `json:"conversationId" form:"conversionId"` //当非草稿时，conversationId必填
}

func (p *PendingConversionRequest) Check() error {
	if !p.Draft && p.ConversationId == "" {
		return fmt.Errorf("conversationId is required")
	}
	return nil
}

type ConversionStreamConnectRequest struct {
	AssistantId    string `json:"assistantId" form:"assistantId"  validate:"required"`
	ConversationId string `json:"conversationId" form:"conversionId"  validate:"required"`
	CommonCheck
}

type ConversionStreamCancelRequest struct {
	PendingConversionRequest
}

func (p *ConversionStreamCancelRequest) Check() error {
	if !p.Draft && p.ConversationId == "" {
		return fmt.Errorf("conversationId is required")
	}
	return nil
}

func (c *ConversionStreamRequest) Check() error {
	return nil
}

type ConversionStreamFile struct {
	FileName string `json:"fileName" form:"fileName"`
	FileSize int64  `json:"fileSize" form:"fileSize"`
	FileUrl  string `json:"fileUrl" form:"fileUrl"`
}

type AssistantTemplateRequest struct {
	AssistantTemplateId string `json:"assistantTemplateId" form:"assistantTemplateId"  validate:"required"`
}

func (a *AssistantTemplateRequest) Check() error { return nil }

type AssistantToolAddRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	ToolId      string `json:"toolId" validate:"required"`
	ToolType    string `json:"toolType" validate:"required,oneof=builtin custom"`
	ActionName  string `json:"actionName" validate:"required"`
}

func (c *AssistantToolAddRequest) Check() error { return nil }

type AssistantToolDelRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	ToolId      string `json:"toolId" validate:"required"`
	ToolType    string `json:"toolType" validate:"required,oneof=builtin custom"`
	ActionName  string `json:"actionName" validate:"required"`
}

func (c *AssistantToolDelRequest) Check() error { return nil }

type AssistantToolEnableRequest struct {
	AssistantId string `json:"assistantId" validate:"required"`
	ToolId      string `json:"toolId" validate:"required"`
	ToolType    string `json:"toolType" validate:"required,oneof=builtin custom"`
	ActionName  string `json:"actionName" validate:"required"`
	Enable      bool   `json:"enable"`
}

func (c *AssistantToolEnableRequest) Check() error { return nil }

type AssistantToolConfigRequest struct {
	AssistantId string              `json:"assistantId" validate:"required"`
	ToolId      string              `json:"toolId" validate:"required"`
	ToolConfig  AssistantToolConfig `json:"toolConfig" validate:"required"`
}

type AssistantToolConfig struct {
	RerankId string `json:"rerankId,omitempty"`
}

func (c *AssistantToolConfigRequest) Check() error { return nil }

type MultiAgentCreateReq struct {
	AssistantId string `json:"assistantId" validate:"required"` // 多智能体id
	AgentId     string `json:"agentId" validate:"required"`     // 子智能体id
	CommonCheck
}

type MultiAgentEnableSwitchReq struct {
	AssistantId string `json:"assistantId" validate:"required"` // 多智能体id
	AgentId     string `json:"agentId" validate:"required"`     // 子智能体id
	Enable      bool   `json:"enable"`                          // 开关状态
	CommonCheck
}

type MultiAgentConfigUpdateReq struct {
	AssistantId string `json:"assistantId" validate:"required"` // 多智能体id
	AgentId     string `json:"agentId" validate:"required"`     // 子智能体id
	Desc        string `json:"desc" validate:"required"`        // 子智能体描述
	CommonCheck
}

type QuestionRecommendRequest struct {
	Query          string `json:"query" form:"query"  validate:"required"`             //用户问题
	AssistantId    string `json:"assistantId" form:"assistantId"  validate:"required"` //智能体id
	ConversationId string `json:"conversationId" form:"conversationId"`                //会话id，如果非试用则不可为空
	Trial          bool   `json:"trial" form:"trial"`
}

func (c *QuestionRecommendRequest) Check() error {
	if !c.Trial && c.ConversationId == "" {
		return errors.New("conversationId can not be empty")
	}
	return nil
}
