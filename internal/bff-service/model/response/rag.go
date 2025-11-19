package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type RagInfo struct {
	RagID string `json:"ragId" validate:"required"`
	request.AppBriefConfig
	ModelConfig         request.AppModelConfig         `json:"modelConfig" validate:"required"`         // 模型 [EN] Model
	RerankConfig        request.AppModelConfig         `json:"rerankConfig" validate:"required"`        // Rerank模型 [EN] Rerank model
	KnowledgeBaseConfig request.AppKnowledgebaseConfig `json:"knowledgeBaseConfig" validate:"required"` // 知识库 [EN] knowledge base
	SafetyConfig        request.AppSafetyConfig        `json:"safetyConfig"`                            // 敏感词表配置 [EN] Sensitive word list configuration
}
