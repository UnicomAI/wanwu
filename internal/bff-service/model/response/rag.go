package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type RagInfo struct {
	RagID string `json:"ragId" validate:"required"`
	request.AppBriefConfig
	ModelConfig         request.AppModelConfig         `json:"modelConfig" validate:"required"`         // Model
	RerankConfig        request.AppModelConfig         `json:"rerankConfig" validate:"required"`        // Rerank model
	KnowledgeBaseConfig request.AppKnowledgebaseConfig `json:"knowledgeBaseConfig" validate:"required"` // knowledge base
	SafetyConfig        request.AppSafetyConfig        `json:"safetyConfig"`                            // Sensitive word list configuration
}
