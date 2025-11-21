package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type Assistant struct {
	request.AppBriefConfig                                // Basic information
	AssistantId            string                         `json:"assistantId"  validate:"required"`
	Prologue               string                         `json:"prologue"`            // opening remarks
	Instructions           string                         `json:"instructions"`        // System prompt word
	RecommendQuestion      []string                       `json:"recommendQuestion"`   // Recommended questions
	ModelConfig            request.AppModelConfig         `json:"modelConfig"`         // Model
	KnowledgeBaseConfig    request.AppKnowledgebaseConfig `json:"knowledgeBaseConfig"` // knowledge base
	RerankConfig           request.AppModelConfig         `json:"rerankConfig"`        // Rerank model
	SafetyConfig           request.AppSafetyConfig        `json:"safetyConfig"`        // Sensitive word list configuration
	VisionConfig           VisionConfig                   `json:"visionConfig"`        // visual configuration
	Scope                  int32                          `json:"scope"`               // Scope
	WorkFlowInfos          []*AssistantWorkFlowInfo       `json:"workFlowInfos"`       // Workflow information
	MCPInfos               []*AssistantMCPInfo            `json:"mcpInfos"`            // MCP information
	ToolInfos              []*AssistantToolInfo           `json:"toolInfos"`           // Custom tools, built-in tools
	CreatedAt              string                         `json:"createdAt"`           // creation time
	UpdatedAt              string                         `json:"updatedAt"`           // Update time
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
	AssistantTemplateId string `json:"assistantTemplateId"` // Agent template ID
	AppType             string `json:"appType"`             // Application type (fixed value: agentTemplate)
	Category            string `json:"category"`            // Category (gov: government affairs, industry: industry, edu: culture and education, medical: medical)
	request.AppBriefConfig
	Prologue                  string   `json:"prologue"`            // opening remarks
	Instructions              string   `json:"instructions"`        // System prompt word
	RecommendQuestion         []string `json:"recommendQuestion"`   // Recommended questions
	Summary                   string   `json:"summary"`             // Usage overview
	Feature                   string   `json:"feature"`             // Feature description
	Scenario                  string   `json:"scenario"`            // Application scenarios
	WorkFlowConfigInstruction string   `json:"workFlowInstruction"` // Workflow configuration instructions
}
