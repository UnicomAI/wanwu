package service_model

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

// BffResponse bff返回结果
type BffResponse struct {
	Code int64      `json:"code"`
	Data *ModelInfo `json:"data"`
	Msg  string     `json:"msg"`
}

type ModelInfo struct {
	ModelId     string                  `json:"modelId"`
	Provider    string                  `json:"provider" validate:"required" enums:"OpenAI-API-compatible,YuanJing"` // 模型供应商
	Model       string                  `json:"model" validate:"required"`                                           // 模型名称
	ModelType   string                  `json:"modelType" validate:"required" enums:"llm,embedding,rerank"`
	DisplayName string                  `json:"displayName"` // 模型显示名称
	Avatar      request.Avatar          `json:"avatar" `     // 模型图标路径
	PublishDate string                  `json:"publishDate"` // 模型发布时间
	IsActive    bool                    `json:"isActive"`    // 启用状态（true: 启用，false: 禁用）
	UserId      string                  `json:"userId"`
	OrgId       string                  `json:"orgId"`
	CreatedAt   string                  `json:"createdAt"`
	UpdatedAt   string                  `json:"updatedAt"`
	ModelDesc   string                  `json:"modelDesc"`
	Tags        []mp_common.Tag         `json:"tags"`
	Config      *LLMModelConfig         `json:"config"`
	Examples    *mp.ProviderModelConfig `json:"examples,omitempty"` // 仅用于swagger展示；模型对应供应商中的对应llm、embedding或rerank结构是config实际的参数
}

type LLMModelConfig struct {
	ApiKey          string `json:"apiKey"`
	EndpointUrl     string `json:"endpointUrl"`     // 模型名称
	FunctionCalling string `json:"functionCalling"` // 是否支持functionCall
	VisionSupport   string `json:"visionSupport"`   // 是否支持多模态
	MaxTokens       int    `json:"maxTokens"`       // 最大token限制
	ContextSize     int    `json:"contextSize"`     // 模型上下文限制
}
