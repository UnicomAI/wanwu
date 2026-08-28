package response

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

// GetGeneralAgentConfigResp 通用智能体配置响应
type GetGeneralAgentConfigResp []*GeneralAgentConfigList

// GeneralAgentConfigList 配置列表
type GeneralAgentConfigList struct {
	ListType string      `json:"listType"` // 类型: tool, mcp, workflow, skill, assistant, knowledge, ontology
	List     interface{} `json:"list"`     // 列表项
}

// GeneralAgentConfigItem 配置项（mcp/assistant/skill/workflow 用）
type GeneralAgentConfigItem struct {
	ID   string `json:"id"`   // ID
	Type string `json:"type"` // 类型
}

// GeneralAgentConfigToolItem tool配置项
type GeneralAgentConfigToolItem struct {
	ID   string `json:"toolId"`   // 工具ID
	Type string `json:"toolType"` // 工具类型
}

type GeneralAgentInfo struct {
	AgentID     string         `json:"agentId"`     // 子智能体ID
	AgentName   string         `json:"agentName"`   // 子智能体名称
	Avatar      request.Avatar `json:"avatar"`      // logo
	Placeholder string         `json:"placeholder"` // 占位提示文本
}

type GetGeneralAgentSubListResp struct {
	WgaAgentList []GeneralAgentInfo `json:"wgaAgentList"` // 子智能体列表
}

type GetGeneralAgentConversationConfigResp struct {
	ThreadID    string                 `json:"threadId"`    // 对话ID
	ModelConfig request.AppModelConfig `json:"modelConfig"` // 模型
}

type CreateGeneralAgentConversationResp struct {
	ThreadID string `json:"threadId"` // 对话ID
}

type CreateGeneralAgentSkillConversationResp struct {
	CustomSkillID string `json:"customSkillId"`
	ThreadID      string `json:"threadId"`
	PreviewID     string `json:"previewId"`
}

type ImportGeneralAgentSkillConversationResp struct {
	CustomSkillID string `json:"customSkillId"`
	ThreadID      string `json:"threadId"`
	PreviewID     string `json:"previewId"`
}

type ConvertGeneralAgentSkillConversationResp struct {
	CustomSkillID string `json:"customSkillId"`
	ThreadID      string `json:"threadId"`
	PreviewID     string `json:"previewId"`
}

type RefreshGeneralAgentSkillConversationResp struct {
	CustomSkillID string `json:"customSkillId"`
	ThreadID      string `json:"threadId"`
	PreviewID     string `json:"previewId"`
}

type GeneralAgentConversationInfo struct {
	ThreadID            string `json:"threadId"`            // 对话ID
	Title               string `json:"title"`               // 对话标题
	CreatedAt           string `json:"createdAt"`           // 创建时间
	UpdatedAt           string `json:"updatedAt"`           // 更新时间
	IsSkillConversation bool   `json:"isSkillConversation"` // 是否为skill对话
	SkillID             string `json:"skillId,omitempty"`   // custom skill ID
	PreviewID           string `json:"previewId,omitempty"` // skill preview conversation ID
}

type GetGeneralAgentToolSelectResp struct {
	Category  string     `json:"category"`  // 类型
	Condition string     `json:"condition"` // 条件 none | optional | required
	ToolList  []ToolInfo `json:"toolList"`  // 工具列表
}

type GeneralAgentToolInfoResp struct {
	Actions []*protocol.Tool `json:"actions"` // action列表
	ToolInfo
}

type GeneralAgentConfigCheckResponse struct {
	Meet      bool                         `json:"meet"`      // 是否符合要求
	ModelMeet bool                         `json:"modelMeet"` // 是否符合模型要求 FIXME 去掉模型检查
	ToolsMeet []GeneralAgentToolCategories `json:"toolsMeet"` // 工具是否符合要求
}

type GeneralAgentToolCategories struct {
	Category  string                  `json:"category"`  // 工具类别类型
	Condition string                  `json:"condition"` // 工具类别条件
	Meet      bool                    `json:"meet"`      // 是否满足条件
	Tools     []GeneralAgentCheckTool `json:"tools"`     // 工具检查结果
}

type GeneralAgentCheckTool struct {
	ToolID string `json:"toolId"` // 工具ID
	Meet   bool   `json:"meet"`   // 是否符合要求
}

type GeneralAgentFileNode struct {
	Name     string                  `json:"name"`
	Type     string                  `json:"type"` // "file" or "directory"
	Size     int64                   `json:"size,omitempty"`
	MimeType string                  `json:"mimeType,omitempty"`
	Children []*GeneralAgentFileNode `json:"children,omitempty"`
}

type GeneralAgentWorkspaceResp struct {
	GeneralAgentConversationWorkspaceInfo
	Path  string                  `json:"path"`
	Files []*GeneralAgentFileNode `json:"files"`
}

type GeneralAgentConversationDetailInfo struct {
	ThreadID  string        `json:"threadId"`
	RunID     string        `json:"runId"`
	CreatedAt int64         `json:"createdAt"`
	Events    []interface{} `json:"events"`
}

type GeneralAgentConversationWorkspaceInfo struct {
	ThreadID  string `json:"threadId"`
	RunID     string `json:"runId"`
	FileCount int32  `json:"fileCount"`
	TotalSize int64  `json:"totalSize"`
	IsDisplay bool   `json:"isDisplay"`
}

type GeneralAgentUploadLimitResp struct {
	UploadLimitList []*GeneralAgentUploadLimit `json:"uploadLimitList"`
}

type GeneralAgentUploadLimit struct {
	FileType string   `json:"fileType"` // 文件类型，如：image、video、audio、document
	MaxSize  int      `json:"maxSize"`  // 文件大小限制，单位MB
	ExtList  []string `json:"extList"`  // 支持的文件后缀列表
}

type GeneralAgentOntologyEmployee struct {
	ID     string         `json:"id"`     // 数字员工ID
	Name   string         `json:"name"`   // 数字员工姓名
	Desc   string         `json:"desc"`   // 数字员工描述
	Avatar request.Avatar `json:"avatar"` // 数字员工头像
}

type GeneralAgentResourceSelectItem struct {
	ID     string         `json:"id"`     // ID
	Name   string         `json:"name"`   // 名称
	Desc   string         `json:"desc"`   // 描述
	Avatar request.Avatar `json:"avatar"` // 头像
	Type   string         `json:"type"`   // 类型
	Author string         `json:"author"` // 作者
}

type GeneralAgentResourceSelectList struct {
	ListType string                            `json:"listType"` // 列表类型: mcp, workflow, skill, assistant, knowledge, ontology
	List     []*GeneralAgentResourceSelectItem `json:"list"`     // 列表项
}

type WgaConversationPendingResp struct {
	ThreadID               string                                    `json:"threadId"`               // 对话ID
	HasPendingConversation bool                                      `json:"hasPendingConversation"` // 是否有运行中的会话
	Messages               []request.GeneralAgentConversationMessage `json:"messages"`               // 运行中会话的用户消息
}
