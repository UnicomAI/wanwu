package mp

// model type
const (
	ModelTypeLLM            = "llm"
	ModelTypeTextEmbedding  = "embedding"
	ModelTypeTextRerank     = "rerank"
	ModelTypeMultiEmbedding = "multimodal-embedding"
	ModelTypeMultiRerank    = "multimodal-rerank"
	ModelTypeOcr            = "ocr"
	ModelTypeGui            = "gui"
	ModelTypePdfParser      = "pdf-parser"
	ModelTypeSyncAsr        = "sync-asr"
	ModelTypeText2Image     = "text2image"
	//ModelTypeOcrDs      = "ocr-deepseek"
	//ModelTypeOcrPaddle  = "ocr-paddle"
)

// model provider
const (
	ProviderOpenAICompatible = "OpenAI-API-compatible"
	ProviderYuanJing         = "YuanJing"
	ProviderHuoshan          = "HuoShan"
	ProviderOllama           = "Ollama"
	ProviderQwen             = "Qwen"
	ProviderInfini           = "Infini"
	ProviderQianfan          = "QianFan"
	ProviderDeepSeek         = "DeepSeek"
	ProviderJina             = "Jina"
	ProviderZhipu            = "ZhiPu"
	ProviderTencent          = "Tencent"
)

const (
	MTNameLLM            = "文本生成"
	MTNameTextEmbedding  = "文本向量化"
	MTNameTextRerank     = "文本重排序"
	MTNameMultiEmbedding = "多模态向量化"
	MTNameMultiRerank    = "多模态重排序"
	MTNameOcr            = "OCR"
	MTNameGui            = "GUI"
	MTNamePdfParser      = "PDF文档解析"
	MTNameSyncAsr        = "短语音识别"

	PNameOpenAICompatible = "OpenAI-API-compatible"
	PNameYuanJing         = "联通元景"
	PNameHuoshan          = "火山引擎"
	PNameOllama           = "Ollama"
	PNameQwen             = "通义千问"
	PNameInfini           = "无问芯穹"
	PNameQianfan          = "百度千帆"
	PNameDeepSeek         = "DeepSeek"
	PNameJina             = "Jina"
	PNameZhipu            = "智谱"
	PNameTencent          = "腾讯"
)

var (
	_callbackUrl string
)

func Init(callbackUrl string) {
	if _callbackUrl != "" {
		panic("model provider already init")
	}
	_callbackUrl = callbackUrl
}
