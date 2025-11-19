import os

LLM_DEVICE = ""
EMBEDDING_DEVICE = ""

# supported LLM models
# llm_model_dict 处理了loader的一些预设行为，如加载位置，模型名称，模型处理器实例 [EN] llm_model_dict handles some preset behaviors of the loader, such as loading location, model name, and model processor instance
# 在以下字典中修改属性值，以指定本地 LLM 模型存储位置 [EN] Modify the property values ​​in the following dictionary to specify the local LLM model storage location
# 如将 "chatglm-6b" 的 "local_model_path" 由 None 修改为 "User/Downloads/chatglm-6b" [EN] For example, change the "local_model_path" of "chatglm-6b" from None to "User/Downloads/chatglm-6b"
# 此处请写绝对路径 [EN] Please write absolute path here
llm_model_dict = {
    "chatglm-6b-int4-qe": {
        "name": "chatglm-6b-int4-qe",
        "pretrained_model_name": "THUDM/chatglm-6b-int4-qe",
        "local_model_path": None,
        "provides": "ChatGLM"
    },
    "chatglm-6b-int4": {
        "name": "chatglm-6b-int4",
        "pretrained_model_name": "THUDM/chatglm-6b-int4",
        "local_model_path": None,
        "provides": "ChatGLM"
    },
    "chatglm-6b-int8": {
        "name": "chatglm-6b-int8",
        "pretrained_model_name": "THUDM/chatglm-6b-int8",
        "local_model_path": None,
        "provides": "ChatGLM"
    },
    "chatglm-6b": {
        "name": "chatglm-6b",
        "pretrained_model_name": "THUDM/chatglm-6b",
        "local_model_path": "/root/exec/chatglm-6b",
        "provides": "ChatGLM"
    },
    "chatglm2-6b": {
        "name": "chatglm2-6b",
        "pretrained_model_name": "THUDM/chatglm2-6b",
        "local_model_path": "/root/exec/chatglm2-6b",
        "provides": "ChatGLM"
    },

    "chatyuan": {
        "name": "chatyuan",
        "pretrained_model_name": "ClueAI/ChatYuan-large-v2",
        "local_model_path": None,
        "provides": None
    },
    "moss": {
        "name": "moss",
        "pretrained_model_name": "fnlp/moss-moon-003-sft",
        "local_model_path": None,
        "provides": "MOSSLLM"
    },
    "vicuna-13b-hf": {
        "name": "vicuna-13b-hf",
        "pretrained_model_name": "vicuna-13b-hf",
        "local_model_path": None,
        "provides": "LLamaLLM"
    },

    # 通过 fastchat 调用的模型请参考如下格式 [EN] For models called through fastchat, please refer to the following format
    "fastchat-chatglm-6b": {
        "name": "chatglm-6b",  # "name"修改为fastchat服务中的"model_name" [EN] "name" is changed to "model_name" in the fastchat service
        "pretrained_model_name": "chatglm-6b",
        "local_model_path": "/root/exec/chatglm-6b",
        "provides": "FastChatOpenAILLM",  # 使用fastchat api时，需保证"provides"为"FastChatOpenAILLM" [EN] When using fastchat api, make sure "provides" is "FastChatOpenAILLM"
        "api_base_url": "http://localhost:30000/v1"  # "name"修改为fastchat服务中的"api_base_url" [EN] "name" is changed to "api_base_url" in the fastchat service
    },

    # 通过 fastchat 调用的模型请参考如下格式 [EN] For models called through fastchat, please refer to the following format
    "fastchat-vicuna-13b-hf": {
        "name": "vicuna-13b-hf",  # "name"修改为fastchat服务中的"model_name" [EN] "name" is changed to "model_name" in the fastchat service
        "pretrained_model_name": "vicuna-13b-hf",
        "local_model_path": None,
        "provides": "FastChatOpenAILLM",  # 使用fastchat api时，需保证"provides"为"FastChatOpenAILLM" [EN] When using fastchat api, make sure "provides" is "FastChatOpenAILLM"
        "api_base_url": "http://localhost:8000/v1"  # "name"修改为fastchat服务中的"api_base_url" [EN] "name" is changed to "api_base_url" in the fastchat service
    },
}

# LLM 名称 [EN] LLM name
LLM_MODEL = "chatglm2-6b"
# 量化加载8bit 模型 [EN] Quantitatively load 8bit model
LOAD_IN_8BIT = False
# Load the model with bfloat16 precision. Requires NVIDIA Ampere GPU.
BF16 = False
# 本地lora存放的位置 [EN] The location where local lora is stored
LORA_DIR = "loras/"

# 缓存知识库数量 [EN] Number of cached knowledge bases
CACHED_VS_NUM = 1

# 文本分句长度 [EN] text clause length
SENTENCE_SIZE = 100

# 匹配后单段上下文长度 [EN] The length of the single segment context after matching
CHUNK_SIZE = 300

# chunk中text的最大长度上限，len(text)超过此长度的chunk会被抛弃掉 [EN] The maximum length of text in a chunk. Chunks whose len(text) exceeds this length will be discarded.
MAX_SENTENCE_SIZE = 4048
# chunk中的text级联切分的最小长度，len(text)小于此长度的chunk不会被级联切分 [EN] The minimum length of text in the chunk for cascade splitting. Chunks whose len(text) is smaller than this length will not be cascaded.
MIN_SENTENCE_SIZE = 100

OCR_MAX_WORKERS = 1
#模型解析服务 [EN] Model parsing service
MODEL_PARSER_MAX_WORKERS = 1

#add file
USER_DATA_PATH = "./user_data"
CONVERT_DIR = "/model_extend/convert"