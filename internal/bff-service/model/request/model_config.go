package request

import (
	"encoding/json"
	"fmt"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
)

type BaseModelRequest struct {
	ModelId string `json:"modelId" form:"modelId" validate:"required"`
}
type ModelConfig struct {
	ModelId     string                  `json:"modelId"`
	Provider    string                  `json:"provider" validate:"required" enums:"OpenAI-API-compatible,YuanJing"` // model supplier
	Model       string                  `json:"model" validate:"required"`                                           // Model name
	ModelType   string                  `json:"modelType" validate:"required" enums:"llm,embedding,rerank"`          // Model type
	DisplayName string                  `json:"displayName" validate:"required"`                                     // Model display name
	Avatar      Avatar                  `json:"avatar" `                                                             // Model icon path
	PublishDate string                  `json:"publishDate"`                                                         // Model release time
	Config      interface{}             `json:"config"`
	ModelDesc   string                  `json:"modelDesc"`          // Model description
	Examples    *mp.ProviderModelConfig `json:"examples,omitempty"` // Only used for swagger display; the corresponding llm, embedding or rerank structure in the model corresponding supplier is the actual parameter of config
}

func (cfg *ModelConfig) Check() error {
	_, err := cfg.ConfigString()
	return err
}

func (cfg *ModelConfig) ConfigString() (string, error) {
	if cfg.Config == nil {
		return "", nil
	}
	b, err := json.Marshal(cfg.Config)
	if err != nil {
		return "", fmt.Errorf("marshal model config err: %v", err)
	}
	modelConfig, err := mp.ToModelConfig(cfg.Provider, cfg.ModelType, string(b))
	if err != nil {
		return "", err
	}
	b, err = json.Marshal(modelConfig)
	if err != nil {
		return "", fmt.Errorf("marshal model config err: %v", err)
	}
	return string(b), nil
}
