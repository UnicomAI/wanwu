package request

import (
	"fmt"

	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type OpenAPIAgentCreateConversationRequest struct {
	Title string `json:"title"`
	UUID  string `json:"uuid" validate:"required"`
}

func (req *OpenAPIAgentCreateConversationRequest) Check() error {
	return nil
}

type OpenAPIAgentChatRequest struct {
	UUID           string                 `json:"uuid" validate:"required"`
	ConversationID string                 `json:"conversation_id"`
	Query          string                 `json:"query" validate:"required"`
	Stream         bool                   `json:"stream"`
	FileInfo       []ConversionStreamFile `json:"file_info"`
}

func (req *OpenAPIAgentChatRequest) Check() error {
	// file_info 字段保留 array 形态以便未来扩展多文件能力，
	// 当前智能体一次仅支持 1 个文件，超过则显式报错避免静默截断。
	if len(req.FileInfo) > 1 {
		return fmt.Errorf("file_info 当前仅支持传 1 个文件")
	}
	for _, f := range req.FileInfo {
		if f.FileUrl != "" {
			if err := url_util.ValidateURL(f.FileUrl); err != nil {
				return err
			}
		}
	}
	return nil
}

type OpenAPIAgentDraftChatRequest struct {
	UUID           string                 `json:"uuid" validate:"required"`
	ConversationID string                 `json:"conversation_id"`
	Query          string                 `json:"query" validate:"required"`
	FileInfo       []ConversionStreamFile `json:"file_info"`
}

func (req *OpenAPIAgentDraftChatRequest) Check() error {
	if len(req.FileInfo) > 1 {
		return fmt.Errorf("file_info 当前仅支持传 1 个文件")
	}
	for _, f := range req.FileInfo {
		if f.FileUrl != "" {
			if err := url_util.ValidateURL(f.FileUrl); err != nil {
				return err
			}
		}
	}
	return nil
}

type OpenAPIRagChatRequest struct {
	UUID    string     `json:"uuid" validate:"required"`
	Query   string     `json:"query" validate:"required"`
	Stream  bool       `json:"stream"`
	History []*History `json:"history"`
}

func (req *OpenAPIRagChatRequest) Check() error {
	return nil
}

type OpenAPICreateAgentRequest struct {
	Category int `json:"category"` // 1:单智能体 2:多智能体
	AppBriefConfig
}

func (req *OpenAPICreateAgentRequest) Check() error {
	// category：0 默认单智能体，1 单智能体，2 多智能体；负数/越界非法
	if req.Category < 0 || req.Category > 2 {
		return fmt.Errorf("invalid category %d, expected 1 (single) or 2 (multi)", req.Category)
	}
	return util.ValidateBriefCreate(&req.Name, &req.Desc, util.SubjectAssistant)
}

// OpenAPIUpdateAgentRequest 更新智能体基本信息请求
type OpenAPIUpdateAgentRequest struct {
	AssistantUUID  string `json:"assistantUuid" validate:"required"` // 智能体UUID
	AppBriefConfig        // 名称、描述、图标
}

func (req *OpenAPIUpdateAgentRequest) Check() error {
	return nil
}

type OpenAPIWorkflowRunReq struct {
	UUID       string         `json:"uuid" validate:"required"`
	Parameters map[string]any `json:"parameters"`
}

func (req *OpenAPIWorkflowRunReq) Check() error {
	return nil
}

type OpenAPIChatflowCreateConversationRequest struct {
	UUID             string `json:"uuid" validate:"required"`
	ConversationName string `json:"conversation_name"`
}

func (req *OpenAPIChatflowCreateConversationRequest) Check() error {
	return nil
}

type OpenAPIChatflowChatRequest struct {
	UUID           string         `json:"uuid" validate:"required"`
	ConversationId string         `json:"conversation_id" validate:"required"`
	Query          string         `json:"query" validate:"required"`
	Parameters     map[string]any `json:"parameters"`
}

func (req *OpenAPIChatflowChatRequest) Check() error {
	return nil
}

type OpenAPIChatflowGetConversationMessageListRequest struct {
	UUID           string `json:"uuid" validate:"required"`
	ConversationId string `json:"conversation_id" validate:"required"`
	Limit          string `json:"limit"`
}

func (req *OpenAPIChatflowGetConversationMessageListRequest) Check() error {
	return nil
}

type OpenAPIChatflowDeleteConversationRequest struct {
	UUID           string `json:"uuid" validate:"required"`
	ConversationId string `json:"conversation_id" validate:"required"`
}

func (req *OpenAPIChatflowDeleteConversationRequest) Check() error {
	return nil
}

type OpenAPIChatflowGetConversationListRequest struct {
	UUID string `json:"uuid" validate:"required"`
}

func (req *OpenAPIChatflowGetConversationListRequest) Check() error {
	return nil
}

type OpenAPIAgentConfigUpdateRequest struct {
	AssistantUUID       string                  `json:"assistantUuid" validate:"required"`
	Prologue            string                  `json:"prologue"`
	Instructions        string                  `json:"instructions"`
	RecommendQuestion   []string                `json:"recommendQuestion"`
	ModelConfig         *AppModelConfig         `json:"modelConfig"`
	KnowledgeBaseConfig *AppKnowledgebaseConfig `json:"knowledgeBaseConfig"`
	SafetyConfig        *AppSafetyConfig        `json:"safetyConfig"`
	RerankConfig        *AppModelConfig         `json:"rerankConfig"`
	VisionConfig        *VisionConfig           `json:"visionConfig"`
	MemoryConfig        *MemoryConfig           `json:"memoryConfig"`
	RecommendConfig     *RecommendConfig        `json:"recommendConfig"`
}

func (req *OpenAPIAgentConfigUpdateRequest) Check() error {
	return validateAgentConfigUpdate(req)
}

type OpenAPIGetAgentInfoRequest struct {
	UUID      string `form:"uuid" validate:"required"`
	Published bool   `form:"published"`
}

func (req *OpenAPIGetAgentInfoRequest) Check() error {
	return nil
}

type OpenAPIAgentPublishRequest struct {
	AssistantUUID string `json:"assistantUuid" validate:"required"`
	Version       string `json:"version" validate:"required"`
	Desc          string `json:"desc" validate:"required"`
	PublishType   string `json:"publishType" validate:"required"` // public:系统公开发布 organization:组织公开发布 private:私密发布
}

func (req *OpenAPIAgentPublishRequest) Check() error {
	if !versionRegexp.MatchString(req.Version) {
		return fmt.Errorf("version must be in format 'vX.Y.Z'")
	}
	return nil
}

// --- agent list / delete ---

// OpenAPIAgentListRequest 智能体列表请求
type OpenAPIAgentListRequest struct {
	Name string `form:"name"` // 按名称模糊筛选，可选
}

func (req *OpenAPIAgentListRequest) Check() error { return nil }

// OpenAPIAgentDeleteRequest 删除智能体请求
type OpenAPIAgentDeleteRequest struct {
	UUID string `form:"uuid" validate:"required"` // 智能体UUID
}

func (req *OpenAPIAgentDeleteRequest) Check() error { return nil }

// --- conversation management (published) ---

// OpenAPIAgentConversationListRequest 已发布智能体对话列表请求
type OpenAPIAgentConversationListRequest struct {
	UUID     string `form:"uuid" validate:"required"` // 智能体UUID
	PageNo   int    `form:"pageNo" validate:"required"`
	PageSize int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIAgentConversationListRequest) Check() error { return nil }

// OpenAPIAgentConversationDetailRequest 已发布智能体对话历史消息请求
type OpenAPIAgentConversationDetailRequest struct {
	ConversationID string `form:"conversation_id" validate:"required"` // 对话ID（从创建对话接口获取）
	PageNo         int    `form:"pageNo" validate:"required"`
	PageSize       int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIAgentConversationDetailRequest) Check() error { return nil }

// OpenAPIAgentConversationDeleteRequest 删除已发布智能体对话请求 删除整个对话主体（含所有消息）
type OpenAPIAgentConversationDeleteRequest struct {
	ConversationID string `form:"conversation_id" validate:"required"` // 对话ID
}

func (req *OpenAPIAgentConversationDeleteRequest) Check() error { return nil }

// OpenAPIAgentConversationClearRequest 清空/按条删除已发布智能体对话历史（保留对话ID）
// detail_id 不传则清空整个对话的全部消息；传值则只删除该条消息
type OpenAPIAgentConversationClearRequest struct {
	ConversationID string `form:"conversation_id" validate:"required"` // 对话ID
	DetailID       string `form:"detail_id"`                           // 可选：消息ID，不传则清空全部消息
}

func (req *OpenAPIAgentConversationClearRequest) Check() error { return nil }

// --- conversation management (draft) ---

// OpenAPIAgentDraftConversationDetailRequest 草稿态智能体对话历史消息请求
// 草稿态每个智能体只有一条会话，通过 uuid 定位
type OpenAPIAgentDraftConversationDetailRequest struct {
	UUID     string `form:"uuid" validate:"required"` // 智能体UUID
	PageNo   int    `form:"pageNo" validate:"required"`
	PageSize int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIAgentDraftConversationDetailRequest) Check() error { return nil }

// OpenAPIAgentDraftConversationDeleteRequest 删除草稿态智能体对话请求（DELETE 走 query）
// detail_id 不传则清空全部消息，传值则只删除该条
type OpenAPIAgentDraftConversationDeleteRequest struct {
	UUID     string `form:"uuid" validate:"required"` // 智能体UUID
	DetailID string `form:"detail_id"`                // 可选：消息ID，不传则清空全部
}

func (req *OpenAPIAgentDraftConversationDeleteRequest) Check() error { return nil }
