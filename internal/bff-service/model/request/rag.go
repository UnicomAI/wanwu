package request

import (
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type RagUpdateReq struct {
	RagID string `json:"ragId" validate:"required"`
	AppBriefConfig
}

type RagCreateReq struct {
	AppBriefConfig
}

func (r *RagCreateReq) Check() error {
	return util.ValidateBriefCreate(&r.Name, &r.Desc, util.SubjectRag)
}

type RagConfig struct {
	RagID                 string                    `json:"ragId" validate:"required"`
	ModelConfig           *AppModelConfig           `json:"modelConfig" validate:"required"`           // 模型
	RerankConfig          *AppModelConfig           `json:"rerankConfig" validate:"required"`          // 知识库Rerank模型
	QARerankConfig        *AppModelConfig           `json:"qaRerankConfig" validate:"required"`        // 问答库Rerank模型
	KnowledgeBaseConfig   *AppKnowledgebaseConfig   `json:"knowledgeBaseConfig" validate:"required"`   // 知识库
	QAKnowledgeBaseConfig *AppQAKnowledgebaseConfig `json:"qaKnowledgeBaseConfig" validate:"required"` // 问答库（不用传知识图谱开关）
	SafetyConfig          *AppSafetyConfig          `json:"safetyConfig"`                              // 敏感词表配置
	VisionConfig          *VisionConfig             `json:"visionConfig"`                              // 视觉开关配置
	RecommendQuestion     []string                  `json:"recommendQuestion"`                         // 推荐问题
}

type ChatRagRequest struct {
	RagID          string                 `json:"ragId" validate:"required"`
	Question       string                 `json:"question" validate:"required"`
	History        []*History             `json:"history"`
	FileInfo       []ConversionStreamFile `json:"fileInfo" form:"fileInfo"` //上传文档列表
	ConversationID string                 `json:"conversationId"`           // 已发布问答必填（须先建会话），草稿不传则服务端 get-or-create
	DetailID       string                 `json:"-"`                        // 本轮明细id，bff 内部生成后透传给 rag-service，不接受调用方传入
}

type RagConversationCreateReq struct {
	RagID  string `json:"ragId" validate:"required"`
	Prompt string `json:"prompt" validate:"required"` // 首次提问，用作会话标题
}

func (r *RagConversationCreateReq) Check() error { return nil }

type RagConversationIDReq struct {
	RagID          string `json:"ragId" form:"ragId" validate:"required"`
	ConversationID string `json:"conversationId" form:"conversationId" validate:"required"`
	DetailID       string `json:"detailId" form:"detailId"` // 可选，传值则只删单条对话，不传则清空全部
}

func (r *RagConversationIDReq) Check() error { return nil }

type RagConversationListReq struct {
	RagID      string `form:"ragId" validate:"required"`
	SearchText string `form:"searchText"`
	PageNo     int    `form:"pageNo" validate:"required"`
	PageSize   int    `form:"pageSize" validate:"required"`
}

func (r *RagConversationListReq) Check() error { return nil }

// RagConversationDraftReq 草稿态按ragId 定位会话
// 分页参数仅查询详情时使用，删除时不传
type RagConversationDraftReq struct {
	RagID    string `json:"ragId" form:"ragId" validate:"required"`
	DetailID string `json:"detailId" form:"detailId"` // 删除时可选，传值则只删单条对话
	PageNo   int    `json:"pageNo" form:"pageNo"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

func (r *RagConversationDraftReq) Check() error { return nil }

type RagConversationDetailListReq struct {
	RagID          string `form:"ragId" validate:"required"`
	ConversationID string `form:"conversationId" validate:"required"`
	PageNo         int    `form:"pageNo" validate:"required"`
	PageSize       int    `form:"pageSize" validate:"required"`
}

func (r *RagConversationDetailListReq) Check() error { return nil }

type RagUploadParams struct {
	Markdown bool `json:"markdown"` // 是否返回markdown格式url
	CommonCheck
}

type History struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type RagReq struct {
	RagID   string `form:"ragId" json:"ragId" validate:"required"`
	Version string `form:"version" json:"version"`
}

func (r RagUpdateReq) Check() error {
	// name/desc 校验在 service 层处理
	return nil
}

func (r *RagConfig) Check() error {
	return nil
}

func (c ChatRagRequest) Check() error {
	for _, f := range c.FileInfo {
		if f.FileUrl != "" {
			if err := url_util.ValidateURL(f.FileUrl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r RagReq) Check() error {
	return nil
}
