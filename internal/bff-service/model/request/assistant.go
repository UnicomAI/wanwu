package request

import "fmt"

type AssistantBrief struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	AppBriefConfig
}

func (a *AssistantBrief) Check() error { return nil }

type AssistantConfig struct {
	AssistantId         string                 `json:"assistantId"  validate:"required"`
	Prologue            string                 `json:"prologue"`            // 开场白 [EN] opening remarks
	Instructions        string                 `json:"instructions"`        // 系统提示词 [EN] System prompt word
	RecommendQuestion   []string               `json:"recommendQuestion"`   // 推荐问题 [EN] Recommended questions
	ModelConfig         AppModelConfig         `json:"modelConfig"`         // 模型 [EN] Model
	KnowledgeBaseConfig AppKnowledgebaseConfig `json:"knowledgeBaseConfig"` // 知识库 [EN] knowledge base
	SafetyConfig        AppSafetyConfig        `json:"safetyConfig"`        // 敏感词表配置 [EN] Sensitive word list configuration
	RerankConfig        AppModelConfig         `json:"rerankConfig"`        // Rerank模型 [EN] Rerank model
	VisionConfig        VisionConfig           `json:"visionConfig"`        // 视觉配置 [EN] visual configuration
}

func (a *AssistantConfig) Check() error { return nil }

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

type ConversationCreateRequest struct {
	AssistantId string `json:"assistantId"  validate:"required"`
	Prompt      string `json:"prompt"  validate:"required"`
}

func (c *ConversationCreateRequest) Check() error { return nil }

type ConversationIdRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
}

func (c *ConversationIdRequest) Check() error { return nil }

type ConversationGetListRequest struct {
	AssistantId string `json:"assistantId" form:"assistantId"  validate:"required"`
	PageSize    int    `json:"pageSize" form:"pageSize"  validate:"required"`
	PageNo      int    `json:"pageNo" form:"pageNo"  validate:"required"`
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
	Trial          bool                   `json:"trial" form:"trial"`
	Prompt         string                 `json:"prompt" form:"prompt"  validate:"required"`
}

func (c *ConversionStreamRequest) Check() error {
	// 当Trial=false时，ConversationId必填 [EN] When Trial=false, ConversationId is required
	if !c.Trial && c.ConversationId == "" {
		return fmt.Errorf("conversationId is required when trial is false")
	}
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
