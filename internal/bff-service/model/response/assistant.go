package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type Assistant struct {
	request.AppBriefConfig                                // 基本信息
	AssistantId            string                         `json:"assistantId"  validate:"required"`
	UUID                   string                         `json:"uuid"`
	Prologue               string                         `json:"prologue"`            // 开场白
	Instructions           string                         `json:"instructions"`        // 系统提示词
	RecommendQuestion      []string                       `json:"recommendQuestion"`   // 推荐问题
	ModelConfig            request.AppModelConfig         `json:"modelConfig"`         // 模型
	KnowledgeBaseConfig    request.AppKnowledgebaseConfig `json:"knowledgeBaseConfig"` // 知识库
	RerankConfig           request.AppModelConfig         `json:"rerankConfig"`        // Rerank模型
	SafetyConfig           request.AppSafetyConfig        `json:"safetyConfig"`        // 敏感词表配置
	VisionConfig           VisionConfig                   `json:"visionConfig"`        // 视觉配置
	MemoryConfig           request.MemoryConfig           `json:"memoryConfig"`        // 记忆配置
	RecommendConfig        RecommendConfig                `json:"recommendConfig"`     // 追问配置
	Scope                  int32                          `json:"scope"`               // 作用域
	WorkFlowInfos          []*AssistantWorkFlowInfo       `json:"workFlowInfos"`       // 工作流信息
	MCPInfos               []*AssistantMCPInfo            `json:"mcpInfos"`            // MCP信息
	ToolInfos              []*AssistantToolInfo           `json:"toolInfos"`           // 自定义工具、内置工具
	SkillInfos             []*AssistantSkillInfo          `json:"skillInfos"`          // 技能配置
	MultiAgentInfos        []*AssistantAgentInfo          `json:"multiAgentInfos"`     // 多智能体配置
	CreatedAt              string                         `json:"createdAt"`           // 创建时间
	UpdatedAt              string                         `json:"updatedAt"`           // 更新时间
	NewAgent               bool                           `json:"newAgent"`            // 是否是新版本智能体
	PublishType            string                         `json:"publishType"`         // 发布类型
	Category               int32                          `json:"category"`            // 智能体分类 1.单智能体 2.多智能体
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
	UniqueId    string         `json:"uniqueId"`
	MCPId       string         `json:"mcpId"`
	MCPType     string         `json:"mcpType" validate:"required,oneof=mcp mcpserver"`
	MCPName     string         `json:"mcpName"`
	ActionName  string         `json:"actionName"`
	Enable      bool           `json:"enable"`
	Valid       bool           `json:"valid"`
	Avatar      request.Avatar `json:"avatar"`
	Description string         `json:"description"`
}

type AssistantToolInfo struct {
	UniqueId    string                      `json:"uniqueId"`
	ToolId      string                      `json:"toolId"`
	ToolType    string                      `json:"toolType" validate:"required,oneof=builtin custom"`
	ToolName    string                      `json:"toolName"`
	ActionName  string                      `json:"actionName"`
	Enable      bool                        `json:"enable"`
	Valid       bool                        `json:"valid"`
	ToolConfig  request.AssistantToolConfig `json:"toolConfig"`
	Avatar      request.Avatar              `json:"avatar"`
	Description string                      `json:"description"`
}

type AssistantSkillInfo struct {
	SkillId     string         `json:"skillId"`
	SkillType   string         `json:"skillType" validate:"required,oneof=builtin custom acquired"`
	SkillName   string         `json:"skillName"`
	Author      string         `json:"author"`
	Enable      bool           `json:"enable"`
	Valid       bool           `json:"valid"`
	Avatar      request.Avatar `json:"avatar"`
	Description string         `json:"description"`
}

type AssistantAgentInfo struct {
	AgentId string         `json:"agentId"`
	Name    string         `json:"name"`
	Desc    string         `json:"desc"`
	Enable  bool           `json:"enable"`
	Avatar  request.Avatar `json:"avatar"`
}

// --- conversation ---

type ConversationInfo struct {
	ConversationId string `json:"conversationId"`
	AssistantId    string `json:"assistantId"`
	Title          string `json:"title"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type ConversationResponse struct {
	Response    string `json:"response"`
	Order       int32  `json:"order"`
	ErrResponse string `json:"errResponse"`
	ErrMessage  string `json:"errMessage"`
}

type ConversationDetailInfo struct {
	Id                  string                  `json:"id"`
	AssistantId         string                  `json:"assistantId"`
	ConversationId      string                  `json:"conversationId"`
	Prompt              string                  `json:"prompt"`
	SysPrompt           string                  `json:"sysPrompt"`
	Response            string                  `json:"response"`
	ResponseList        []*ConversationResponse `json:"responseList"`
	SearchList          interface{}             `json:"searchList"`
	QaType              int32                   `json:"qa_type"`
	CreatedBy           string                  `json:"createdBy"`
	CreatedAt           int64                   `json:"createdAt"`
	UpdatedAt           int64                   `json:"updatedAt"`
	RequestFiles        []AssistantRequestFile  `json:"requestFiles"`
	FileSize            int64                   `json:"fileSize"`
	FileName            string                  `json:"fileName"`
	SubConversationList []*SubConversation      `json:"subConversationList"`
	ResponseFiles       []*AgentResponseFile    `json:"responseFiles"`
}

type SubConversation struct {
	Response         string      ` json:"response"`
	SearchList       interface{} `json:"searchList"`
	ParentId         string      ` json:"parentId"`        // 事件挂载id
	Id               string      ` json:"id"`              // 事件id
	Name             string      `json:"name"`             // 事件名称
	Profile          string      `json:"profile"`          // 事件头像
	TimeCost         string      `json:"timeCost"`         // 耗时
	Status           int32       `json:"status"`           // 1:成功，2：失败
	ConversationType string      `json:"conversationType"` // subAgent：子智能体；agentTool：主智能体工具；subAgentTool：子智能体工具
	Order            int32       `json:"order"`
	ErrMessage       string      `json:"errMessage"` //错误信息
}

type AssistantRequestFile struct {
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	FileUrl  string `json:"fileUrl"`
}

type ConversationCreateResp struct {
	ConversationId string `json:"conversationId"`
}

type ConversationIdResp struct {
	ConversationId string `json:"conversationId"`
}

type AssistantCreateResp struct {
	AssistantId string `json:"assistantId"`
}

type PendingConversationResp struct {
	ConversationId         string                 `json:"conversationId"`         // 会话id
	HasPendingConversation bool                   `json:"hasPendingConversation"` // 是否有进行中会话
	Prompt                 string                 `json:"prompt"`
	RequestFiles           []AssistantRequestFile `json:"requestFiles"`
}

// --- conversation sse ---

type ConversationSSEData struct {
	Code           int                  `json:"code"`
	Message        string               `json:"message"`
	Response       string               `json:"response"`
	Order          int                  `json:"order"`
	EventType      int                  `json:"eventType"`
	EventData      interface{}          `json:"eventData"`
	GenFileUrlList []interface{}        `json:"gen_file_url_list"`
	History        []interface{}        `json:"history"`
	Finish         int                  `json:"finish"`
	Usage          ConversationSSEUsage `json:"usage"`
	SearchList     []interface{}        `json:"search_list"`
	QaType         int                  `json:"qa_type"`
	ResponseFiles  []*AgentResponseFile `json:"responseFiles"`
}
type ConversationSSEUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type AssistantResponseFile struct {
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	FileUrl  string `json:"fileUrl"`
	MIMEType string `json:"mimeType"`

	// 扩展信息：
	// skill => {"name":"技能名称", "desc":"技能描述", "author":"技能作者", "avatar":{"path":"技能图标"}, "inResource": bool, "expiredAt": "过期时间7天", "skillSaveId": "保存的技能ID"}
	MetaData map[string]interface{} `json:"metadata"`
}

type AgentResponseFile struct {
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	FileUrl  string `json:"fileUrl"`
	FileType string `json:"fileType"`

	// 扩展信息：
	MetaData *AgentFileMeta `json:"metadata"`
}

type AgentFileMeta struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	CreateAt string `json:"createAt"`
}

// --- assistant template ---

type AssistantTemplateInfo struct {
	AssistantTemplateId string `json:"assistantTemplateId"` // 智能体模板Id
	AppType             string `json:"appType"`             // 应用类型(固定值: agentTemplate)
	Category            string `json:"category"`            // 种类(gov:政务,industry:工业,edu:文教,medical:医疗)
	request.AppBriefConfig
	Prologue                  string   `json:"prologue"`            // 开场白
	Instructions              string   `json:"instructions"`        // 系统提示词
	RecommendQuestion         []string `json:"recommendQuestion"`   // 推荐问题
	Summary                   string   `json:"summary"`             // 使用概述
	Feature                   string   `json:"feature"`             // 特性说明
	Scenario                  string   `json:"scenario"`            // 应用场景
	WorkFlowConfigInstruction string   `json:"workFlowInstruction"` // 工作流配置说明
}
