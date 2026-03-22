package mp

import (
	"encoding/json"
	"fmt"
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type ILLMParams interface {
	GetParams() map[string]interface{}
}

// ToModelEndpoint 返回model、model_url、model_id的kv
func ToModelEndpoint(modelId, model string) map[string]interface{} {
	ret := make(map[string]interface{})
	if modelId != "" && model != "" {
		modelUrl, _ := url.JoinPath(_callbackUrl, "/callback/v1/model", modelId)
		ret["model"] = model
		ret["model_url"] = modelUrl
		ret["model_id"] = modelId
	}
	return ret
}

// ToModelParams 返回ILLMParams、IEmbeddingParams或IRerankParams，与对应实际传给模型的参数
func ToModelParams(provider, modelType, cfg string) (interface{}, map[string]interface{}, error) {
	params := make(map[string]interface{})
	if cfg == "" {
		return nil, params, nil
	}
	var ret interface{} // 前端需要的结构体
	var err error
	switch provider {
	case ProviderOpenAICompatible:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderYuanJing:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		case ModelTypeGui:
		case ModelTypeOcr:
		case ModelTypePdfParser:
		case ModelTypeSyncAsr:
		case ModelTypeText2Image:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderHuoshan:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderOllama:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderQwen:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderInfini:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderQianfan:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		case ModelTypeTextRerank:
		case ModelTypeTextEmbedding:
		case ModelTypeOcr:
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderDeepSeek:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	case ProviderMiniMax:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_common.LLMParams{}
			if err = json.Unmarshal([]byte(cfg), llm); err == nil {
				ret = llm
				params = llm.GetParams()
			}
		default:
			return nil, nil, fmt.Errorf("invalid model type: %v", modelType)
		}
	default:
		return nil, nil, fmt.Errorf("invalid provider: %v", modelType)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal model params err: %v", err)
	}
	return ret, params, nil
}

type AppModelParams struct {
	ProviderOpenAICompatible AppModelParamsOpenAICompatible `json:"providerOpenAICompatible"` // OpenAI-API-compatible模型配置
	ProviderYuanJing         AppModelParamsYuanjing         `json:"providerYuanjing"`         // YuanJing模型配置
	ProviderHuoshan          AppModelParamsHuoShan          `json:"providerHuoshan"`
	ProviderQwen             AppModelParamsQwen             `json:"providerQwen"`
	ProviderOllama           AppModelParamsOllama           `json:"providerOllama"`
	ProviderInfini           AppModelParamsInfini           `json:"providerInfini"`
	ProviderQianFan          AppModelParamsQianFan          `json:"providerQianFan"`
	ProviderDeepSeek         AppModelParamsDeepSeek         `json:"providerDeepSeek"`
	ProviderMiniMax          AppModelParamsMiniMax          `json:"providerMiniMax"` // MiniMax模型配置
}

type AppModelParamsOpenAICompatible struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsYuanjing struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsHuoShan struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsQwen struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsOllama struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsInfini struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsQianFan struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsDeepSeek struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}

type AppModelParamsMiniMax struct {
	LLM mp_common.LLMParams `json:"llm"` // 大语言模型配置
}
