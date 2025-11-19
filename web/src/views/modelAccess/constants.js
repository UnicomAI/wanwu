export const LLM = 'llm'
export const RERANK = 'rerank'
export const EMBEDDING = 'embedding'
export const OCR = 'ocr'
export const GUI = 'gui'
export const PDF_PARSER = 'pdf-parser'

export const MODEL_TYPE_OBJ = {
    [LLM]: 'LLM',
    [RERANK]: 'Rerank',
    [EMBEDDING]: 'Embedding',
    [OCR]: 'OCR',
    [GUI]: 'GUI',
    [PDF_PARSER]: 'Document Parsing Service'
}

export const MODEL_TYPE = Object.keys(MODEL_TYPE_OBJ).map(key => ({key, name: MODEL_TYPE_OBJ[key]}))

export const YUAN_JING = 'YuanJing'
export const OPENAI_API = 'OpenAI-API-compatible'
export const OLLAMA = 'Ollama'
export const QWEN = 'Qwen'
export const HUOSHAN = 'Huoshan'
export const INFINI = 'Infini'
export const PROVIDER_OBJ = {
    [OPENAI_API]: 'OpenAI-API-compatible',
    [YUAN_JING]: 'China Unicom Yuanjing',
    [OLLAMA]: 'Ollama',
    [QWEN]: 'Tongyi Qianwen',
    [HUOSHAN]: 'Volcano Engine',
    [INFINI]: 'Infini'
}

export const PROVIDER_IMG_OBJ = {
    [OPENAI_API]: require('@/assets/imgs/openAI.png'),
    [YUAN_JING]: require('@/assets/imgs/yuanjing.png'),
    [OLLAMA]: require('@/assets/imgs/ollama.png'),
    [QWEN]: require('@/assets/imgs/qwen.png'),
    [HUOSHAN]: require('@/assets/imgs/volcano.png'),
    [INFINI]: require('@/assets/imgs/infini.png'),
}

export const PROVIDER_TYPE = Object.keys(PROVIDER_OBJ)
    .map(key => {
        return ({
            key,
            name: PROVIDER_OBJ[key],
            children: key === YUAN_JING
                ? MODEL_TYPE
                : [OLLAMA, HUOSHAN].includes(key)
                    ? MODEL_TYPE.filter(item => [LLM, EMBEDDING].includes(item.key))
                    : MODEL_TYPE.filter(item => [LLM, RERANK, EMBEDDING].includes(item.key))
        })
    })

export const DEFAULT_CALLING = 'noSupport'
export const FUNC_CALLING = [
    {key: 'noSupport', name: 'Not Supported'},
    {key: 'toolCall', name: 'Tool call'},
    /*{key: 'functionCall', name: 'Function call'},*/
]

export const DEFAULT_SUPPORT = 'noSupport'
export const SUPPORT_LIST = [
    {key: 'noSupport', name: 'Not Supported'},
    {key: 'support', name: 'Supported'},
]

export const TYPE_OBJ = {
    apiKey: {
        [YUAN_JING]: 'sk-abc********************xyz',
        [OPENAI_API]: 'sk_7e4*************4s-BpI1l',
        [OLLAMA]: '',
        [QWEN]: 'sk-b************c70d',
        [HUOSHAN]: 'd8008ac0-****-****-****-**************',
        [INFINI]: 'sk-nw****gzjb6'
    },
    inferUrl: {
        [OCR]: 'https://maas-api.ai-yuanjing.com/openapi/v1',
        [GUI]: 'https://maas-api.ai-yuanjing.com/openapi/v1',
        [PDF_PARSER]: 'https://maas-api.ai-yuanjing.com/openapi/v1',
        [YUAN_JING]: 'https://maas.ai-yuanjing.com/openapi/compatible-mode/v1',
        [OPENAI_API]: 'https://api.siliconflow.cn/v1',
        [OLLAMA]: 'https://192.168.21.100:11434',
        [QWEN]: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
        [HUOSHAN]: 'https://ark.cn-beijing.volces.com/api/v3',
        [INFINI]: 'https://cloud.infini-ai.com/maas/v1'
    },
}