package request

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
)

// UpdateGeneralAgentConfigReq 更新通用智能体配置请求
type UpdateGeneralAgentConfigReq struct {
	Tool      []GeneralAgentConfigToolItem `json:"tool"`
	Mcp       []GeneralAgentConfigItem     `json:"mcp"`
	Workflow  []GeneralAgentConfigItem     `json:"workflow"`
	Skill     []GeneralAgentConfigItem     `json:"skill"`
	Assistant []GeneralAgentConfigItem     `json:"assistant"`
	Knowledge []GeneralAgentConfigItem     `json:"knowledge"`
	Ontology  []GeneralAgentConfigItem     `json:"ontology" validate:"max=1"` // 目前限制只能选一个知识网络
}

// GeneralAgentConfigItem 配置项（带type）
type GeneralAgentConfigItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// GeneralAgentConfigToolItem tool配置项
type GeneralAgentConfigToolItem struct {
	ID   string `json:"toolId"`
	Type string `json:"toolType"`
}

func (c *UpdateGeneralAgentConfigReq) Check() error { return nil }

type GetGeneralAgentConversationConfigReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *GetGeneralAgentConversationConfigReq) Check() error { return nil }

type CreateGeneralAgentConversationReq struct {
	Title       string          `json:"title" validate:"required"`       // 标题
	ModelConfig *AppModelConfig `json:"modelConfig" validate:"required"` // 模型
}

func (c *CreateGeneralAgentConversationReq) Check() error { return nil }

// OpenAPIWgaCreateConversationReq OpenAPI WGA 创建对话请求（与 v1 handler 分离，不传 ModelConfig 而使用 modelUuid）
type OpenAPIWgaCreateConversationReq struct {
	Title     string `json:"title" validate:"required"`     // 标题
	ModelUuid string `json:"modelUuid" validate:"required"` // 模型UUID，必填
}

func (c *OpenAPIWgaCreateConversationReq) Check() error { return nil }

type CreateGeneralAgentSkillConversationReq CreateGeneralAgentConversationReq

func (c *CreateGeneralAgentSkillConversationReq) Check() error {
	return (*CreateGeneralAgentConversationReq)(c).Check()
}

type ImportGeneralAgentSkillConversationReq struct {
	CreateCustomSkillReq
	ModelConfig *AppModelConfig `json:"modelConfig" validate:"required"`
}

func (c *ImportGeneralAgentSkillConversationReq) Check() error { return nil }

type ConvertGeneralAgentSkillConversationReq struct {
	ID          string          `json:"id" validate:"required"`
	Type        string          `json:"type" validate:"required"`
	ModelConfig *AppModelConfig `json:"modelConfig" validate:"required"`
}

func (c *ConvertGeneralAgentSkillConversationReq) Check() error {
	switch strings.TrimSpace(strings.ToLower(c.Type)) {
	case "mcp", "tool", "agent", "workflow", "rag":
		return nil
	default:
		return fmt.Errorf("unsupported type: %s", c.Type)
	}
}

type RefreshGeneralAgentSkillConversationReq struct {
	SkillID string `json:"skillId" validate:"required"`
}

func (c *RefreshGeneralAgentSkillConversationReq) Check() error { return nil }

type DeleteGeneralAgentConversationReq struct {
	ThreadID string `json:"threadId" validate:"required"` // 对话ID
}

func (c *DeleteGeneralAgentConversationReq) Check() error { return nil }

type GetGeneralAgentConversationListReq struct {
	PageNo     int    `json:"pageNo" form:"pageNo" validate:"required"`     // 页码
	PageSize   int    `json:"pageSize" form:"pageSize" validate:"required"` // 每页数量
	SearchText string `json:"searchText" form:"searchText"`                 // 标题关键词，模糊匹配，空则不过滤
}

func (c *GetGeneralAgentConversationListReq) Check() error { return nil }

type GetGeneralAgentConversationDetailReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *GetGeneralAgentConversationDetailReq) Check() error { return nil }

type GetGeneralAgentSkillPreviewConversationDetailReq struct {
	PreviewID string `json:"previewId" form:"previewId" validate:"required"`
}

func (c *GetGeneralAgentSkillPreviewConversationDetailReq) Check() error { return nil }

type GeneralAgentConfigCheckRequest struct {
	AgentID  string `json:"agentId"`                                      // 子智能体ID
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *GeneralAgentConfigCheckRequest) Check() error { return nil }

type UpdateGeneralAgentConversationConfigReq struct {
	ThreadID    string          `json:"threadId" validate:"required"`    // 对话ID
	ModelConfig *AppModelConfig `json:"modelConfig" validate:"required"` // 模型
}

type ToolSelected struct {
	ToolID   string `json:"toolId" validate:"required"`   // 工具ID
	ToolType string `json:"toolType" validate:"required"` // 工具类型
}

type AssistantSelected struct {
	AssistantID string `json:"assistantId" validate:"required"` // 智能体ID
}

func (c *UpdateGeneralAgentConversationConfigReq) Check() error { return nil }

type GeneralAgentConversationChatReq struct {
	AgentID  string                            `json:"agentId"`                      // 智能体ID
	ThreadID string                            `json:"threadId" validate:"required"` // 对话ID
	Messages []GeneralAgentConversationMessage `json:"messages" validate:"required"` // 消息
}

func (c *GeneralAgentConversationChatReq) Check() error { return nil }

// OpenAPIWgaConversationChatReq OpenAPI WGA 对话流请求（与 v1 handler 分离，modelUuid 必填）
type OpenAPIWgaConversationChatReq struct {
	AgentID   string                            `json:"agentId"`                      // 智能体ID，可选，为空时使用默认Supervisor
	ThreadID  string                            `json:"threadId" validate:"required"` // 对话ID
	Messages  []GeneralAgentConversationMessage `json:"messages" validate:"required"` // 消息列表
	ModelUuid string                            `json:"modelUuid"`                    // 模型UUID，可选
}

func (c *OpenAPIWgaConversationChatReq) Check() error { return nil }

// GeneralAgentSkillConversationChatReq Skill对话请求
type GeneralAgentSkillConversationChatReq struct {
	CustomSkillID string                            `json:"customSkillId" validate:"required"` // 自定义技能ID，用于workspace隔离
	Mode          string                            `json:"mode"`                              // 模式：normal/import/convert/preview，默认 normal
	PreviewID     string                            `json:"previewId"`                         // preview模式对话ID，用于历史记录隔离
	ThreadID      string                            `json:"threadId" validate:"required"`      // 对话ID
	Messages      []GeneralAgentConversationMessage `json:"messages" validate:"required"`      // 消息
}

func (c *GeneralAgentSkillConversationChatReq) Check() error {
	mode := strings.TrimSpace(strings.ToLower(c.Mode))
	switch mode {
	case "", "normal", "import", "convert", "preview":
	default:
		return fmt.Errorf("unsupported mode: %s", c.Mode)
	}
	if mode == "preview" && strings.TrimSpace(c.PreviewID) == "" {
		return fmt.Errorf("previewId is required when mode is preview")
	}
	return nil
}

type GeneralAgentConversationMessage struct {
	ID      string      `json:"id"`                          // 消息id
	Role    string      `json:"role" validate:"required"`    // 角色 user
	Content interface{} `json:"content" validate:"required"` // 内容 string 或者 [{"type":"text","text":"这张图片是什么？"},{"type":"binary","mimeType":"image/png","url":"https://..."}]
}

func (m *GeneralAgentConversationMessage) GetURLs() map[string]string {
	urls := make(map[string]string)
	switch v := m.Content.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if m["type"] == "binary" {
					urlStr, _ := m["url"].(string)
					fileName, _ := m["fileName"].(string)
					if urlStr != "" && fileName != "" {
						urls[fileName] = urlStr
					}
				}
			}
		}
	}
	return urls
}

func (m *GeneralAgentConversationMessage) GetTextContent() string {
	var text string
	switch v := m.Content.(type) {
	case string:
		text = v
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if typ, _ := m["type"].(string); typ == "text" {
					if t, _ := m["text"].(string); t != "" {
						text += t + " "
					}
				}
			}
		}
	}
	return text
}

func (m *GeneralAgentConversationMessage) WithMinioUrlReplaced() GeneralAgentConversationMessage {
	var cp GeneralAgentConversationMessage
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &cp)

	minioEndpoint := config.Cfg().Minio.Endpoint
	downloadURL := config.Cfg().Minio.DownloadURL
	if minioEndpoint == "" || downloadURL == "" {
		return cp
	}

	if v, ok := cp.Content.([]interface{}); ok {
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if typ, _ := m["type"].(string); typ == "binary" {
					if urlStr, _ := m["url"].(string); urlStr != "" {
						m["url"] = strings.ReplaceAll(urlStr, "http://"+minioEndpoint, downloadURL)
					}
				}
			}
		}
	}
	return cp
}

type GeneralAgentSkillConversationPendingReq struct {
	ThreadID  string `json:"threadId" form:"threadId"`   // 对话ID
	PreviewID string `json:"previewId" form:"previewId"` // preview模式对话ID，用于历史记录隔离
}

func (c *GeneralAgentSkillConversationPendingReq) Check() error { return nil }

func (c *GeneralAgentSkillConversationPendingReq) ChatThreadID() string {
	if strings.TrimSpace(c.PreviewID) != "" {
		return c.PreviewID
	}
	return c.ThreadID
}

type GeneralAgentSkillConversationConnectReq = GeneralAgentSkillConversationPendingReq

type GeneralAgentSkillConversationCancelReq = GeneralAgentSkillConversationPendingReq

type GeneralAgentWorkspaceDownloadReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
	RunID    string `json:"runId" form:"runId"`                           // 运行ID
	Path     string `json:"path" form:"path"`                             // workspace中路径
}

func (c *GeneralAgentWorkspaceDownloadReq) Check() error { return nil }

type GeneralAgentWorkspacePreviewReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
	RunID    string `json:"runId" form:"runId"`                           // 运行ID
	Path     string `json:"path" form:"path" validate:"required"`         // 文件路径
}

func (c *GeneralAgentWorkspacePreviewReq) Check() error { return nil }

type GeneralAgentWorkspaceReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
	RunID    string `json:"runId" form:"runId"`                           // 运行ID
}

func (c *GeneralAgentWorkspaceReq) Check() error { return nil }

type GeneralAgentReplyQuestionReq struct {
	RunID      string     `json:"runId" validate:"required"`
	QuestionID string     `json:"questionId" validate:"required"`
	Answers    [][]string `json:"answers" validate:"required"`
}

func (c *GeneralAgentReplyQuestionReq) Check() error { return nil }

type GeneralAgentRejectQuestionReq struct {
	RunID      string `json:"runId" validate:"required"`
	QuestionID string `json:"questionId" validate:"required"`
}

func (c *GeneralAgentRejectQuestionReq) Check() error { return nil }

// WgaConversationPendingReq 查询WGA运行中会话请求
type WgaConversationPendingReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *WgaConversationPendingReq) Check() error { return nil }

// WgaConversationConnectReq WGA流式问答断线重连请求
type WgaConversationConnectReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *WgaConversationConnectReq) Check() error { return nil }

// WgaConversationCancelReq WGA流式问答手动停止请求
type WgaConversationCancelReq struct {
	ThreadID string `json:"threadId" form:"threadId" validate:"required"` // 对话ID
}

func (c *WgaConversationCancelReq) Check() error { return nil }
