package request

import (
	"fmt"

	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// 知识问答 OpenAPI 请求模型，对外一律用 uuid 指代应用（uuid 即 ragId）。

// --- 应用管理 ---

// OpenAPICreateRagRequest 创建知识问答
type OpenAPICreateRagRequest struct {
	AppBriefConfig // 名称、描述、图标
}

func (req *OpenAPICreateRagRequest) Check() error {
	return util.ValidateBriefCreate(&req.Name, &req.Desc, util.SubjectRag)
}

// OpenAPIUpdateRagRequest 更新知识问答基本信息
type OpenAPIUpdateRagRequest struct {
	UUID           string `json:"uuid" validate:"required"` // 知识问答UUID
	AppBriefConfig        // 名称、描述、图标
}

func (req *OpenAPIUpdateRagRequest) Check() error { return nil }

// OpenAPIRagDeleteRequest 删除知识问答
type OpenAPIRagDeleteRequest struct {
	UUID string `form:"uuid" validate:"required"`
}

func (req *OpenAPIRagDeleteRequest) Check() error { return nil }

// OpenAPIRagListRequest 知识问答列表
type OpenAPIRagListRequest struct {
	Name string `form:"name"` // 按名称模糊筛选，可选
}

func (req *OpenAPIRagListRequest) Check() error { return nil }

// OpenAPIGetRagInfoRequest 知识问答详情
type OpenAPIGetRagInfoRequest struct {
	UUID      string `form:"uuid" validate:"required"`
	Published bool   `form:"published"` // true=已发布态，false=草稿态（默认）
}

func (req *OpenAPIGetRagInfoRequest) Check() error { return nil }

// OpenAPIRagConfigUpdateRequest 更新知识问答配置，除 uuid 外均可选，未传的沿用当前草稿配置
type OpenAPIRagConfigUpdateRequest struct {
	UUID                  string                    `json:"uuid" validate:"required"`
	ModelConfig           *AppModelConfig           `json:"modelConfig"`           // 对话模型，modelId 传模型UUID
	RerankConfig          *AppModelConfig           `json:"rerankConfig"`          // 知识库Rerank模型，modelId 传模型UUID
	QARerankConfig        *AppModelConfig           `json:"qaRerankConfig"`        // 问答库Rerank模型，modelId 传模型UUID
	KnowledgeBaseConfig   *AppKnowledgebaseConfig   `json:"knowledgeBaseConfig"`   // 知识库
	QAKnowledgeBaseConfig *AppQAKnowledgebaseConfig `json:"qaKnowledgeBaseConfig"` // 问答库
	SafetyConfig          *AppSafetyConfig          `json:"safetyConfig"`          // 安全护栏
	VisionConfig          *VisionConfig             `json:"visionConfig"`          // 视觉开关
	RecommendQuestion     []string                  `json:"recommendQuestion"`     // 推荐问题
}

func (req *OpenAPIRagConfigUpdateRequest) Check() error { return nil }

// OpenAPIRagPublishRequest 发布知识问答
type OpenAPIRagPublishRequest struct {
	UUID        string `json:"uuid" validate:"required"`
	Version     string `json:"version" validate:"required"`
	Desc        string `json:"desc" validate:"required"`
	PublishType string `json:"publishType" validate:"required"` // public:系统公开发布 organization:组织公开发布 private:私密发布
}

func (req *OpenAPIRagPublishRequest) Check() error {
	if !versionRegexp.MatchString(req.Version) {
		return fmt.Errorf("version must be in format 'vX.Y.Z'")
	}
	return nil
}

// --- 会话管理 ---

// OpenAPIRagCreateConversationRequest 创建会话
type OpenAPIRagCreateConversationRequest struct {
	UUID  string `json:"uuid" validate:"required"`
	Title string `json:"title" validate:"required"` // 会话标题，一般取首个问题
}

func (req *OpenAPIRagCreateConversationRequest) Check() error { return nil }

// OpenAPIRagConversationListRequest 会话列表
type OpenAPIRagConversationListRequest struct {
	UUID       string `form:"uuid" validate:"required"`
	SearchText string `form:"searchText"` // 按会话标题模糊筛选，可选
	PageNo     int    `form:"pageNo" validate:"required"`
	PageSize   int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIRagConversationListRequest) Check() error { return nil }

// OpenAPIRagConversationDetailRequest 会话历史消息
type OpenAPIRagConversationDetailRequest struct {
	UUID           string `form:"uuid" validate:"required"`
	ConversationID string `form:"conversation_id" validate:"required"`
	PageNo         int    `form:"pageNo" validate:"required"`
	PageSize       int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIRagConversationDetailRequest) Check() error { return nil }

// OpenAPIRagConversationDeleteRequest 删除会话（含全部消息）
type OpenAPIRagConversationDeleteRequest struct {
	UUID           string `form:"uuid" validate:"required"`
	ConversationID string `form:"conversation_id" validate:"required"`
}

func (req *OpenAPIRagConversationDeleteRequest) Check() error { return nil }

// OpenAPIRagConversationClearRequest 清空/按条删除会话消息，会话本身保留
type OpenAPIRagConversationClearRequest struct {
	UUID           string `form:"uuid" validate:"required"`
	ConversationID string `form:"conversation_id" validate:"required"`
	DetailID       string `form:"detail_id"` // 不传则清空该会话全部消息
}

func (req *OpenAPIRagConversationClearRequest) Check() error { return nil }

// OpenAPIRagDraftConversationDetailRequest 草稿会话历史消息（草稿态每个应用仅一条会话）
type OpenAPIRagDraftConversationDetailRequest struct {
	UUID     string `form:"uuid" validate:"required"`
	PageNo   int    `form:"pageNo" validate:"required"`
	PageSize int    `form:"pageSize" validate:"required"`
}

func (req *OpenAPIRagDraftConversationDetailRequest) Check() error { return nil }

// OpenAPIRagDraftConversationDeleteRequest 删除草稿会话消息
type OpenAPIRagDraftConversationDeleteRequest struct {
	UUID     string `form:"uuid" validate:"required"`
	DetailID string `form:"detail_id"` // 不传则清空全部
}

func (req *OpenAPIRagDraftConversationDeleteRequest) Check() error { return nil }

// --- 草稿态问答 ---

// OpenAPIRagDraftChatRequest 草稿态问答，基于草稿配置，不要求已发布。
// 草稿态每个知识问答只有一条会话，由服务端按 uuid 解析，故不接受调用方传 conversation_id
type OpenAPIRagDraftChatRequest struct {
	UUID     string                 `json:"uuid" validate:"required"`
	Query    string                 `json:"query" validate:"required"`
	Stream   bool                   `json:"stream"`
	FileInfo []ConversionStreamFile `json:"file_info"` // 随问题携带的文件，当前仅支持 1 张图片(png/jpg/jpeg)
}

func (req *OpenAPIRagDraftChatRequest) Check() error {
	return checkRagChatFileInfo(req.FileInfo)
}

// checkRagChatFileInfo 校验问答附件，只管数量与地址合法性，附件类型由 rag-service buildAttachmentList 判定
func checkRagChatFileInfo(fileInfoList []ConversionStreamFile) error {
	if len(fileInfoList) > 1 {
		return fmt.Errorf("file_info 当前仅支持传 1 个文件")
	}
	for _, f := range fileInfoList {
		if f.FileUrl == "" {
			continue
		}
		if err := url_util.ValidateURL(f.FileUrl); err != nil {
			return err
		}
	}
	return nil
}
