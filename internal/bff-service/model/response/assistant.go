package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type Assistant struct {
	request.AppBriefConfig                                // 基本信息 [EN] Basic information
	AssistantId            string                         `json:"assistantId"  validate:"required"`
	Prologue               string                         `json:"prologue"`            // 开场白 [EN] opening remarks
	Instructions           string                         `json:"instructions"`        // 系统提示词 [EN] System prompt word
	RecommendQuestion      []string                       `json:"recommendQuestion"`   // 推荐问题 [EN] Recommended questions
	ModelConfig            request.AppModelConfig         `json:"modelConfig"`         // 模型 [EN] Model
	KnowledgeBaseConfig    request.AppKnowledgebaseConfig `json:"knowledgeBaseConfig"` // 知识库 [EN] knowledge base
	RerankConfig           request.AppModelConfig         `json:"rerankConfig"`        // Rerank模型 [EN] Rerank model
	SafetyConfig           request.AppSafetyConfig        `json:"safetyConfig"`        // 敏感词表配置 [EN] Sensitive word list configuration
	VisionConfig           VisionConfig                   `json:"visionConfig"`        // 视觉配置 [EN] visual configuration
	Scope                  int32                          `json:"scope"`               // 作用域 [EN] Scope
	WorkFlowInfos          []*AssistantWorkFlowInfo       `json:"workFlowInfos"`       // 工作流信息 [EN] Workflow information
	MCPInfos               []*AssistantMCPInfo            `json:"mcpInfos"`            // MCP信息 [EN] MCP information
	ToolInfos              []*AssistantToolInfo           `json:"toolInfos"`           // 自定义工具、内置工具 [EN] Custom tools, built-in tools
	CreatedAt              string                         `json:"createdAt"`           // 创建时间 [EN] creation time
	UpdatedAt              string                         `json:"updatedAt"`           // 更新时间 [EN] Update time
}

type AssistantWorkFlowInfo struct {
	UniqueId     string         `json:"uniqueId"`
	WorkFlowId   string         `json:"workFlowId"`
	ApiName      string         `json:"apiName"`
	Enable       bool           `json:"enable"`
	AvatarPath   request.Avatar `json:"avatar"`
	WorkFlowName string         `json:"name"`
	WorkFlowDesc string         `json:"workFlowDesc"`
}

type AssistantMCPInfo struct {
	UniqueId   string         `json:"uniqueId"`
	MCPId      string         `json:"mcpId"`
	MCPType    string         `json:"mcpType" validate:"required,oneof=mcp mcpserver"`
	MCPName    string         `json:"mcpName"`
	ActionName string         `json:"actionName"`
	Enable     bool           `json:"enable"`
	Valid      bool           `json:"valid"`
	Avatar     request.Avatar `json:"avatar"`
}

type AssistantToolInfo struct {
	UniqueId   string                      `json:"uniqueId"`
	ToolId     string                      `json:"toolId"`
	ToolType   string                      `json:"toolType" validate:"required,oneof=builtin custom"`
	ToolName   string                      `json:"toolName"`
	ActionName string                      `json:"actionName"`
	Enable     bool                        `json:"enable"`
	Valid      bool                        `json:"valid"`
	ToolConfig request.AssistantToolConfig `json:"toolConfig"`
	Avatar     request.Avatar              `json:"avatar"`
}

type ConversationInfo struct {
	ConversationId string `json:"conversationId"`
	AssistantId    string `json:"assistantId"`
	Title          string `json:"title"`
	CreatedAt      string `json:"createdAt"`
}

type ConversationDetailInfo struct {
	Id             string                 `json:"id"`
	AssistantId    string                 `json:"assistantId"`
	ConversationId string                 `json:"conversationId"`
	Prompt         string                 `json:"prompt"`
	SysPrompt      string                 `json:"sysPrompt"`
	Response       string                 `json:"response"`
	SearchList     interface{}            `json:"searchList"`
	QaType         int32                  `json:"qa_type"`
	CreatedBy      string                 `json:"createdBy"`
	CreatedAt      int64                  `json:"createdAt"`
	UpdatedAt      int64                  `json:"updatedAt"`
	RequestFiles   []AssistantRequestFile `json:"requestFiles"`
	FileSize       int64                  `json:"fileSize"`
	FileName       string                 `json:"fileName"`
}
type AssistantRequestFile struct {
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	FileUrl  string `json:"fileUrl"`
}
type ConversationCreateResp struct {
	ConversationId string `json:"conversationId"`
}

type AssistantCreateResp struct {
	AssistantId string `json:"assistantId"`
}

type AssistantTemplateInfo struct {
	AssistantTemplateId string `json:"assistantTemplateId"` // 智能体模板Id [EN] Agent template ID
	AppType             string `json:"appType"`             // 应用类型(固定值: agentTemplate) [EN] Application type (fixed value: agentTemplate)
	Category            string `json:"category"`            // 种类(gov:政务,industry:工业,edu:文教,medical:医疗) [EN] Category (gov: government affairs, industry: industry, edu: culture and education, medical: medical)
	request.AppBriefConfig
	Prologue                  string   `json:"prologue"`            // 开场白 [EN] opening remarks
	Instructions              string   `json:"instructions"`        // 系统提示词 [EN] System prompt word
	RecommendQuestion         []string `json:"recommendQuestion"`   // 推荐问题 [EN] Recommended questions
	Summary                   string   `json:"summary"`             // 使用概述 [EN] Usage overview
	Feature                   string   `json:"feature"`             // 特性说明 [EN] Feature description
	Scenario                  string   `json:"scenario"`            // 应用场景 [EN] Application scenarios
	WorkFlowConfigInstruction string   `json:"workFlowInstruction"` // 工作流配置说明 [EN] Workflow configuration instructions
}
