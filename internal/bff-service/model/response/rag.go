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
