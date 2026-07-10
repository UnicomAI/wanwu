package mp

import (
	"context"
	"encoding/json"
	"fmt"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	mp_deepseek "github.com/UnicomAI/wanwu/pkg/model-provider/mp-deepseek"
	mp_huoshan "github.com/UnicomAI/wanwu/pkg/model-provider/mp-huoshan"
	mp_infini "github.com/UnicomAI/wanwu/pkg/model-provider/mp-infini"
	mp_jina "github.com/UnicomAI/wanwu/pkg/model-provider/mp-jina"
	mp_ollama "github.com/UnicomAI/wanwu/pkg/model-provider/mp-ollama"
	mp_openai_compatible "github.com/UnicomAI/wanwu/pkg/model-provider/mp-openai-compatible"
	mp_qianfan "github.com/UnicomAI/wanwu/pkg/model-provider/mp-qianfan"
	mp_qwen "github.com/UnicomAI/wanwu/pkg/model-provider/mp-qwen"
	mp_yuanjing "github.com/UnicomAI/wanwu/pkg/model-provider/mp-yuanjing"
	mp_zhipu "github.com/UnicomAI/wanwu/pkg/model-provider/mp-zhipu"
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

type IMultiModalEmbedding interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.MultiModalEmbeddingReq) (mp_common.IMultiModalEmbeddingReq, error)
	MultiModalEmbeddings(ctx context.Context, req mp_common.IMultiModalEmbeddingReq, headers ...mp_common.Header) (mp_common.IMultiModalEmbeddingResp, error)
}

type IRerank interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.TextRerankReq) (mp_common.ITextRerankReq, error)
	Rerank(ctx context.Context, req mp_common.ITextRerankReq, headers ...mp_common.Header) (mp_common.ITextRerankResp, error)
}

type IMultiModalRerank interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.MultiModalRerankReq) (mp_common.IMultiModalRerankReq, error)
	MultiModalRerank(ctx context.Context, req mp_common.IMultiModalRerankReq, headers ...mp_common.Header) (mp_common.IMultiModalRerankResp, error)
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

type ISyncAsr interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.SyncAsrReq) (mp_common.ISyncAsrReq, error)
	SyncAsr(ctx context.Context, req mp_common.ISyncAsrReq, headers ...mp_common.Header) (mp_common.ISyncAsrResp, error)
}

type IText2Image interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.Text2ImageReq) (mp_common.IText2ImageReq, error)
	Text2Image(ctx *gin.Context, req mp_common.IText2ImageReq, headers ...mp_common.Header) (mp_common.IText2ImageResp, error)
}

type IGui interface {
	Tags() []mp_common.Tag
	NewReq(req *mp_common.GuiReq) (mp_common.IGuiReq, error)
	Gui(ctx context.Context, req mp_common.IGuiReq, headers ...mp_common.Header) (mp_common.IGuiResp, error)
}

// ToModelTags ILLM、IEmbedding或IRerank 的标签列表
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
		case ModelTypeTextRerank:
			rerank := &mp_openai_compatible.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_openai_compatible.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeSyncAsr:
			asr := &mp_openai_compatible.SyncAsr{}
			if err := json.Unmarshal([]byte(cfg), asr); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = asr.Tags()
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
		case ModelTypeTextRerank:
			rerank := &mp_yuanjing.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeMultiRerank:
			rerank := &mp_yuanjing.MultiModalRerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
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
		case ModelTypeSyncAsr:
			asr := &mp_yuanjing.SyncAsr{}
			if err := json.Unmarshal([]byte(cfg), asr); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = asr.Tags()
		case ModelTypeText2Image:
			text2Image := &mp_yuanjing.Text2Image{}
			if err := json.Unmarshal([]byte(cfg), text2Image); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = text2Image.Tags()
		case ModelTypeMultiEmbedding:
			embedding := &mp_yuanjing.MultiModalEmbedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
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
		case ModelTypeTextEmbedding:
			embedding := &mp_huoshan.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeSyncAsr:
			asr := &mp_huoshan.SyncAsr{}
			if err := json.Unmarshal([]byte(cfg), asr); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = asr.Tags()
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
		case ModelTypeTextRerank:
			rerank := &mp_qwen.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_qwen.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeMultiEmbedding:
			embedding := &mp_qwen.MultiModalEmbedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeSyncAsr:
			asr := &mp_qwen.SyncAsr{}
			if err := json.Unmarshal([]byte(cfg), asr); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = asr.Tags()
		case ModelTypeMultiRerank:
			rerank := &mp_qwen.MultiModalRerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
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
		case ModelTypeTextEmbedding:
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
		case ModelTypeTextRerank:
			rerank := &mp_infini.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_infini.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderQianfan:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_qianfan.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeTextRerank:
			rerank := &mp_qianfan.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_qianfan.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderDeepSeek:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_deepseek.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderJina:
		switch modelType {
		case ModelTypeTextRerank:
			rerank := &mp_jina.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeMultiRerank:
			rerank := &mp_jina.MultiModalRerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_jina.Embedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		case ModelTypeMultiEmbedding:
			embedding := &mp_jina.MultiModalEmbedding{}
			if err := json.Unmarshal([]byte(cfg), embedding); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = embedding.Tags()
		default:
			return nil, fmt.Errorf("ToModelTags:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderZhipu:
		switch modelType {
		case ModelTypeLLM:
			llm := &mp_zhipu.LLM{}
			if err := json.Unmarshal([]byte(cfg), llm); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = llm.Tags()
		case ModelTypeTextRerank:
			rerank := &mp_zhipu.Rerank{}
			if err := json.Unmarshal([]byte(cfg), rerank); err != nil {
				return nil, fmt.Errorf("unmarshal model config err: %v", err)
			}
			tags = rerank.Tags()
		case ModelTypeTextEmbedding:
			embedding := &mp_zhipu.Embedding{}
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

var (
	maxTextLength    int64 = 512
	maxVideoClipSize int64 = 10
	maxImageSize     int64 = 3
	maxAsrFileSize   int64 = 10
)

// ToModelConfig 返回ILLM、IEmbedding或IRerank
func ToModelConfig(provider, modelType, cfg string) (interface{}, error) {
	if cfg == "" {
		return nil, nil
	}
	var ret interface{} // 前端需要的结构体
	switch provider {
	case ProviderOpenAICompatible:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_openai_compatible.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_openai_compatible.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_openai_compatible.Embedding{}
		case ModelTypeSyncAsr:
			ret = &mp_openai_compatible.SyncAsr{
				MaxAsrFileSize: &maxAsrFileSize,
			}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderYuanJing:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_yuanjing.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_yuanjing.Rerank{}
		case ModelTypeMultiRerank:
			ret = &mp_yuanjing.MultiModalRerank{
				MaxTextLength:       &maxTextLength,
				MaxVideoClipSize:    &maxVideoClipSize,
				MaxImageSize:        &maxImageSize,
				SupportFileTypes:    []string{"image"},
				SupportImageInQuery: true,
			}
		case ModelTypeTextEmbedding:
			ret = &mp_yuanjing.Embedding{}
		case ModelTypeOcr:
			ret = &mp_yuanjing.Ocr{}
		case ModelTypeGui:
			ret = &mp_yuanjing.Gui{}
		case ModelTypePdfParser:
			ret = &mp_yuanjing.PdfParser{}
		case ModelTypeSyncAsr:
			ret = &mp_yuanjing.SyncAsr{
				MaxAsrFileSize: &maxAsrFileSize,
			}
		case ModelTypeText2Image:
			ret = &mp_yuanjing.Text2Image{}
		case ModelTypeMultiEmbedding:
			ret = &mp_yuanjing.MultiModalEmbedding{
				MaxTextLength:    &maxTextLength,
				MaxImageSize:     &maxImageSize,
				MaxVideoClipSize: &maxVideoClipSize,
				SupportFileTypes: []string{"image"},
			}

		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderHuoshan:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_huoshan.LLM{}
		case ModelTypeTextEmbedding:
			ret = &mp_huoshan.Embedding{}
		case ModelTypeSyncAsr:
			ret = &mp_huoshan.SyncAsr{
				MaxAsrFileSize: &maxAsrFileSize,
			}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderQwen:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_qwen.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_qwen.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_qwen.Embedding{}
		case ModelTypeMultiEmbedding:
			ret = &mp_qwen.MultiModalEmbedding{
				MaxTextLength:    &maxTextLength,
				MaxImageSize:     &maxImageSize,
				MaxVideoClipSize: &maxVideoClipSize,
				SupportFileTypes: []string{"image"},
			}
		case ModelTypeSyncAsr:
			ret = &mp_qwen.SyncAsr{
				MaxAsrFileSize: &maxAsrFileSize,
			}
		case ModelTypeMultiRerank:
			ret = &mp_qwen.MultiModalRerank{
				MaxTextLength:       &maxTextLength,
				MaxImageSize:        &maxImageSize,
				SupportFileTypes:    []string{"image"},
				SupportImageInQuery: true,
			}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderOllama:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_ollama.LLM{}
		case ModelTypeTextEmbedding:
			ret = &mp_ollama.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderInfini:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_infini.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_infini.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_infini.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderQianfan:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_qianfan.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_qianfan.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_qianfan.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderDeepSeek:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_deepseek.LLM{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderJina:
		switch modelType {
		case ModelTypeTextRerank:
			ret = &mp_jina.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_jina.Embedding{}
		case ModelTypeMultiRerank:
			ret = &mp_jina.MultiModalRerank{
				MaxTextLength:       &maxTextLength,
				MaxVideoClipSize:    &maxVideoClipSize,
				MaxImageSize:        &maxImageSize,
				SupportFileTypes:    []string{"image"},
				SupportImageInQuery: false,
			}
		case ModelTypeMultiEmbedding:
			ret = &mp_jina.MultiModalEmbedding{
				MaxTextLength:    &maxTextLength,
				MaxImageSize:     &maxImageSize,
				SupportFileTypes: []string{"image"},
			}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	case ProviderZhipu:
		switch modelType {
		case ModelTypeLLM:
			ret = &mp_zhipu.LLM{}
		case ModelTypeTextRerank:
			ret = &mp_zhipu.Rerank{}
		case ModelTypeTextEmbedding:
			ret = &mp_zhipu.Embedding{}
		default:
			return nil, fmt.Errorf("ToModelConfig:invalid provider %v model type %v", provider, modelType)
		}
	default:
		return nil, fmt.Errorf("ToModelConfig:invalid provider: %v", provider)
	}

	if err := json.Unmarshal([]byte(cfg), ret); err != nil {
		return nil, fmt.Errorf("unmarshal model config err: %v", err)
	}
	return ret, nil
}

type ProviderModelConfig struct {
	ProviderYuanJing         ProviderModelByYuanjing         `json:"providerYuanJing"`
	ProviderOpenAICompatible ProviderModelByOpenAICompatible `json:"providerOpenAICompatible"`
	ProviderHuoshan          ProviderModelByHuoShan          `json:"providerHuoShan"`
	ProviderQwen             ProviderModelByQwen             `json:"providerQwen"`
	ProviderOllama           ProviderModelByOllama           `json:"providerOllama"`
	ProviderInfini           ProviderModelByInfini           `json:"providerInfini"`
	ProviderQianFan          ProviderModelByQianFan          `json:"providerQianFan"`
	ProviderDeepSeek         ProviderModelByDeepSeek         `json:"providerDeepSeek"`
	ProviderZhipu            ProviderModelByZhipu            `json:"providerZhipu"`
}

type ProviderModelByOpenAICompatible struct {
	Llm       mp_openai_compatible.LLM       `json:"llm"`
	Rerank    mp_openai_compatible.Rerank    `json:"rerank"`
	Embedding mp_openai_compatible.Embedding `json:"embedding"`
	Asr       mp_openai_compatible.SyncAsr   `json:"asr"`
}

type ProviderModelByYuanjing struct {
	Llm       mp_yuanjing.LLM       `json:"llm"`
	Rerank    mp_yuanjing.Rerank    `json:"rerank"`
	Embedding mp_yuanjing.Embedding `json:"embedding"`
	Ocr       mp_yuanjing.Ocr       `json:"ocr"`
	Gui       mp_yuanjing.Gui       `json:"gui"`
	PdfParser mp_yuanjing.PdfParser `json:"pdf-parser"`
}

type ProviderModelByHuoShan struct {
	Llm       mp_huoshan.LLM       `json:"llm"`
	Embedding mp_huoshan.Embedding `json:"embedding"`
}

type ProviderModelByQwen struct {
	Llm                 mp_qwen.LLM                 `json:"llm"`
	Rerank              mp_qwen.Rerank              `json:"rerank"`
	Embedding           mp_qwen.Embedding           `json:"embedding"`
	MultiModalEmbedding mp_qwen.MultiModalEmbedding `json:"multiModalEmbedding"`
	MultiModalRerank    mp_qwen.MultiModalRerank    `json:"multiModalRerank"`
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

type ProviderModelByQianFan struct {
	Llm       mp_qianfan.LLM       `json:"llm"`
	Rerank    mp_qianfan.Rerank    `json:"rerank"`
	Embedding mp_qianfan.Embedding `json:"embedding"`
}

type ProviderModelByDeepSeek struct {
	Llm mp_deepseek.LLM `json:"llm"`
}

type ProviderModelByZhipu struct {
	Llm       mp_zhipu.LLM       `json:"llm"`
	Rerank    mp_zhipu.Rerank    `json:"rerank"`
	Embedding mp_zhipu.Embedding `json:"embedding"`
}
