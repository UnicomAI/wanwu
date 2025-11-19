package mp

import (
	"context"
	"encoding/json"
	"fmt"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	mp_huoshan "github.com/UnicomAI/wanwu/pkg/model-provider/mp-huoshan"
	mp_infini "github.com/UnicomAI/wanwu/pkg/model-provider/mp-infini"
	mp_ollama "github.com/UnicomAI/wanwu/pkg/model-provider/mp-ollama"
	mp_openai_compatible "github.com/UnicomAI/wanwu/pkg/model-provider/mp-openai-compatible"
	mp_qwen "github.com/UnicomAI/wanwu/pkg/model-provider/mp-qwen"
	mp_yuanjing "github.com/UnicomAI/wanwu/pkg/model-provider/mp-yuanjing"
	"github.com/gin-gonic/gin"
)

type ILLM interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.LLMReq) (mp_common.ILLMReq, error)
	ChatCompletions(ctx context.Context, req mp_common.ILLMReq, headers ...mp_common.Header) (mp_common.ILLMResp, <-chan mp_common.ILLMResp, error)
}

type IEmbedding interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.EmbeddingReq) (mp_common.IEmbeddingReq, error)
	Embeddings(ctx context.Context, req mp_common.IEmbeddingReq, headers ...mp_common.Header) (mp_common.IEmbeddingResp, error)
}

type IRerank interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.RerankReq) (mp_common.IRerankReq, error)
	Rerank(ctx context.Context, req mp_common.IRerankReq, headers ...mp_common.Header) (mp_common.IRerankResp, error)
}

type IOcr interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.OcrReq) (mp_common.IOcrReq, error)
	Ocr(ctx *gin.Context, req mp_common.IOcrReq, headers ...mp_common.Header) (mp_common.IOcrResp, error)
}

type IPdfParser interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.PdfParserReq) (mp_common.IPdfParserReq, error)
	PdfParser(ctx *gin.Context, req mp_common.IPdfParserReq, headers ...mp_common.Header) (mp_common.IPdfParserResp, error)
}

type IGui interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.GuiReq) (mp_common.IGuiReq, error)
	Gui(ctx context.Context, req mp_common.IGuiReq, headers ...mp_common.Header) (mp_common.IGuiResp, error)
}

// ToModelTags ILLM、IEmbedding或IRerank 的 标签列表 [EN] ToModelTags tag list for ILLM, IEmbedding or IRerank
func ToModelTags(provider, modelType, cfg string) ([]mp_common.Tag, error) {
	if cfg == "" {
		return nil, nil
	}
	var tags []mp_common.Tag
	switch provider {
	case ProviderOpenAICompatible:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_openai_compatible.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeRerank:
			rerank := &mp_openai_compatible.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_openai_compatible.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderYuanJing:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_yuanjing.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeRerank:
			rerank := &mp_yuanjing.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_yuanjing.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeOcr:
			ocr := &mp_yuanjing.Ocr{}
			if err := json.Unmarshal([]byte(cfg), ocr); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = ocr.Tags()
		case ModelTypeGui:
			gui := &mp_yuanjing.Gui{}
			if err := json.Unmarshal([]byte(cfg), gui); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = gui.Tags()
		case ModelTypePdfParser:
			pdfParser := &mp_yuanjing.PdfParser{}
			if err := json.Unmarshal([]byte(cfg), pdfParser); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = pdfParser.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderHuoshan:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_huoshan.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_huoshan.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderQwen:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_qwen.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeRerank:
			rerank := &mp_qwen.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_qwen.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderOllama:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_ollama.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_ollama.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderInfini:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_infini.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeRerank:
			rerank := &mp_infini.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeEmbedding:
			embedding := &mp_infini.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	default:
		return nil, fmt.Errorf("ToModelTags:invalid provider: %v", provider)
	}
	return tags, nil
}

// ToModelConfig 返回ILLM、IEmbedding或IRerank [EN] ToModelConfig returns ILLM, IEmbedding or IRerank
func ToModelConfig(provider, modelType, cfg string) (interface{}, error) {
	if cfg == "" {
		return nil, nil
	}
	var ret interface{} // 前端需要的结构体 [EN] Structure required by the front end
	switch provider {
	case ProviderOpenAICompatible:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_openai_compatible.LLM{}
		case ModelTypeRerank:
			ret = &mp_openai_compatible.Rerank{}
		case ModelTypeEmbedding:
			ret = &mp_openai_compatible.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderYuanJing:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_yuanjing.LLM{}
		case ModelTypeRerank:
			ret = &mp_yuanjing.Rerank{}
		case ModelTypeEmbedding:
			ret = &mp_yuanjing.Embedding{}
		case ModelTypeOcr:
			ret = &mp_yuanjing.Ocr{}
		case ModelTypeGui:
			ret = &mp_yuanjing.Gui{}
		case ModelTypePdfParser:
			ret = &mp_yuanjing.PdfParser{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderHuoshan:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_huoshan.LLM{}
		case ModelTypeEmbedding:
			ret = &mp_huoshan.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderQwen:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_qwen.LLM{}
		case ModelTypeRerank:
			ret = &mp_qwen.Rerank{}
		case ModelTypeEmbedding:
			ret = &mp_qwen.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderOllama:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_ollama.LLM{}
		case ModelTypeEmbedding:
			ret = &mp_ollama.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderInfini:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_infini.LLM{}
		case ModelTypeRerank:
			ret = &mp_infini.Rerank{}
		case ModelTypeEmbedding:
			ret = &mp_infini.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	default:
		return nil, fmt.Errorf("ToModelConfig:invalid provider: %v", modelType)
	}

	if err := json.Unmarshal([]byte(cfg), ret); err != nil {
		return nil, fmt.Errorf("unmarshal model config err: %v", err)
	}
	return ret, nil
}

type ProviderModelConfig struct {
	ProviderYuanJing         ProviderModelByYuanjing         `json:"providerYuanJing"`
	ProviderOpenAICompatible ProviderModelByOpenAICompatible `json:"providerOpenAICompatible"`
	ProviderHuoshan          ProviderModelByHuoshan          `json:"providerHuoshan"`
	ProviderQwen             ProviderModelByQwen             `json:"providerQwen"`
	ProviderOllama           ProviderModelByOllama           `json:"providerOllama"`
	ProviderInfini           ProviderModelByInfini           `json:"providerModelByInfini"`
}

type ProviderModelByOpenAICompatible struct {
	Llm       mp_openai_compatible.LLM       `json:"llm"`
	Rerank    mp_openai_compatible.Rerank    `json:"rerank"`
	Embedding mp_openai_compatible.Embedding `json:"embedding"`
}

type ProviderModelByYuanjing struct {
	Llm       mp_yuanjing.LLM       `json:"llm"`
	Rerank    mp_yuanjing.Rerank    `json:"rerank"`
	Embedding mp_yuanjing.Embedding `json:"embedding"`
	Ocr       mp_yuanjing.Ocr       `json:"ocr"`
	Gui       mp_yuanjing.Gui       `json:"gui"`
	PdfParser mp_yuanjing.PdfParser `json:"pdf-parser"`
}

type ProviderModelByHuoshan struct {
	Llm       mp_huoshan.LLM       `json:"llm"`
	Embedding mp_huoshan.Embedding `json:"embedding"`
}

type ProviderModelByQwen struct {
	Llm       mp_qwen.LLM       `json:"llm"`
	Rerank    mp_qwen.Rerank    `json:"rerank"`
	Embedding mp_qwen.Embedding `json:"embedding"`
}

type ProviderModelByOllama struct {
	Llm       mp_ollama.LLM       `json:"llm"`
	Embedding mp_ollama.Embedding `json:"embedding"`
}

type ProviderModelByInfini struct {
	Llm       mp_infini.LLM       `json:"llm"`
	Rerank    mp_infini.Rerank    `json:"rerank"`
	Embedding mp_infini.Embedding `json:"embedding"`
}
