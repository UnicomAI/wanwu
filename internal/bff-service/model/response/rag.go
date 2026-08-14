package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type RagInfo struct {
	RagID string `json:"ragId" validate:"required"`
	request.AppBriefConfig
	ModelConfig           request.AppModelConfig           `json:"modelConfig" validate:"required"`           // 模型
	RerankConfig          request.AppModelConfig           `json:"rerankConfig" validate:"required"`          // Rerank模型
	QARerankConfig        request.AppModelConfig           `json:"qaRerankConfig" validate:"required"`        // 问答库Rerank模型
	KnowledgeBaseConfig   request.AppKnowledgebaseConfig   `json:"knowledgeBaseConfig" validate:"required"`   // 知识库
	QAKnowledgeBaseConfig request.AppQAKnowledgebaseConfig `json:"qaKnowledgeBaseConfig" validate:"required"` // 问答库
	SafetyConfig          request.AppSafetyConfig          `json:"safetyConfig"`                              // 敏感词表配置
	AppPublishConfig      request.AppPublishConfig         `json:"appPublishConfig"`                          // 发布配置
	VisionConfig          request.VisionConfig             `json:"visionConfig"`                              // 视觉开关
	RecommendQuestion     []string                         `json:"recommendQuestion"`                         // 推荐问题
	CreatedAt             string                           `json:"createdAt"`                                 // 创建时间
	UpdatedAt             string                           `json:"updatedAt"`                                 // 更新时间
}

type RagUploadResult struct {
	DownloadLink string `json:"download_link"`
	Error        string `json:"error"`
}

type RagUploadFile struct {
	FileIndex int    `json:"fileIndex"`
	FileUrl   string `json:"fileUrl"`
}

type RagUploadResponseWithErr struct {
	RagUploadFile *RagUploadFile `json:"ragUploadFile"`
	Error         error          `json:"error"`
}

func RagUploadError(index int, err error) *RagUploadResponseWithErr {
	return &RagUploadResponseWithErr{RagUploadFile: &RagUploadFile{FileIndex: index}, Error: err}
}

func RagUploadSuccess(index int, ragUploadFile *RagUploadFile) *RagUploadResponseWithErr {
	ragUploadFile.FileIndex = index
	return &RagUploadResponseWithErr{RagUploadFile: ragUploadFile}
}

type RagUploadResponse struct {
	FileList []*RagUploadFile `json:"fileList"`
}

type RagConversationCreateResp struct {
	ConversationId string `json:"conversationId"`
}

type RagConversationInfo struct {
	ConversationId string `json:"conversationId"`
	RagId          string `json:"ragId"`
	Title          string `json:"title"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type RagConversationDetailInfo struct {
	DetailId             string           `json:"detailId"`
	RagId                string           `json:"ragId"`
	ConversationId       string           `json:"conversationId"`
	Prompt               string           `json:"prompt"`
	Response             string           `json:"response"`
	ReasoningContent     *string          `json:"reasoningContent"` // 深度思考过程；null=整轮没思考，""=思考过但无内容
	SearchList           *string          `json:"searchList"`       // 知识库引用段落；null=没走知识库检索，""=检索无召回
	QaSearchList         *string          `json:"qaSearchList"`     // 问答库命中结果，与 searchList 互斥；null=没走问答库检索，""=检索未命中
	ErrMessage           string           `json:"errMessage"`
	QaErrMessage         string           `json:"qaErrMessage"` // 问答库检索失败原因，非空表示这轮是问答库报错后转的知识库
	TraceId              string           `json:"traceId"`      // 本轮问答的链路id，可据此关联应用/模型统计明细
	RequestFiles         []RagRequestFile `json:"requestFiles"`
	CreatedAt            int64            `json:"createdAt"`
	ReasoningTimeCost    RagTimeCost      `json:"reasoningTimeCost"`    // 深度思考起止
	SearchListTimeCost   RagTimeCost      `json:"searchListTimeCost"`   // 知识库检索起止
	QaSearchListTimeCost RagTimeCost      `json:"qaSearchListTimeCost"` // 问答库检索起止
}

// RagTimeCost 一个阶段的起止时刻(毫秒)，两个字段均为0表示该阶段未发生
type RagTimeCost struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type RagRequestFile struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileUrl  string `json:"fileUrl"`
}
