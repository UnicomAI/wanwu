package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type ModelInfo struct {
	ModelId     string                  `json:"modelId"`
	Provider    string                  `json:"provider" validate:"required" enums:"OpenAI-API-compatible,YuanJing"` // model supplier
	Model       string                  `json:"model" validate:"required"`                                           // Model name
	ModelType   string                  `json:"modelType" validate:"required" enums:"llm,embedding,rerank"`
	DisplayName string                  `json:"displayName"` // Model display name
	Avatar      request.Avatar          `json:"avatar" `     // Model icon path
	PublishDate string                  `json:"publishDate"` // Model release time
	IsActive    bool                    `json:"isActive"`    // Enabled status (true: enabled, false: disabled)
	UserId      string                  `json:"userId"`
	OrgId       string                  `json:"orgId"`
	CreatedAt   string                  `json:"createdAt"`
	UpdatedAt   string                  `json:"updatedAt"`
	ModelDesc   string                  `json:"modelDesc"`
	Tags        []mp_common.Tag         `json:"tags"`
	Config      interface{}             `json:"config"`
	Examples    *mp.ProviderModelConfig `json:"examples,omitempty"` // Only used for swagger display; the corresponding llm, embedding or rerank structure in the model corresponding supplier is the actual parameter of config
}
