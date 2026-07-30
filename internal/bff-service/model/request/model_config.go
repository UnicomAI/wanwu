package request

import (
	"encoding/json"
	"fmt"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
)

type BaseModelRequest struct {
	ModelId string `json:"modelId" form:"modelId" validate:"required"`
}

type ModelConfig struct {
	ModelId      string                  `json:"modelId"`
	Provider     string                  `json:"provider" validate:"required" enums:"OpenAI-API-compatible,YuanJing"` // 模型供应商
	Model        string                  `json:"model" validate:"required"`                                           // 模型名称
	ModelType    string                  `json:"modelType" validate:"required" enums:"llm,embedding,rerank"`          // 模型类型
	DisplayName  string                  `json:"displayName" validate:"required"`                                     // 模型显示名称
	Avatar       Avatar                  `json:"avatar" `                                                             // 模型图标路径
	PublishDate  string                  `json:"publishDate"`                                                         // 模型发布时间
	Config       interface{}             `json:"config"`
	ModelDesc    string                  `json:"modelDesc"`          // 模型描述
	Examples     *mp.ProviderModelConfig `json:"examples,omitempty"` // 仅用于swagger展示；模型对应供应商中的对应llm、embedding或rerank结构是config实际的参数
	ScopeType    string                  `json:"scopeType"`
	ImportSource string                  `json:"importSource"` // 模型导入来源(builtin=平台内置,external=外部URL，默认external)
}

func (cfg *ModelConfig) Check() error {
	_, err := cfg.ConfigString()
	if err != nil {
		return err
	}
	if cfg.Config != nil {
		configBytes, err := json.Marshal(cfg.Config)
		if err != nil {
			return err
		}
		var endpointCfg struct {
			EndpointUrl string `json:"endpointUrl"`
		}
		if err := json.Unmarshal(configBytes, &endpointCfg); err != nil {
			return err
		}
		if endpointCfg.EndpointUrl != "" {
			if err := url_util.ValidateURL(endpointCfg.EndpointUrl); err != nil {
				return err
			}
		}
	}
	return nil
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
