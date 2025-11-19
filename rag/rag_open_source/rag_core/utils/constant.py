import os

LLM_DEVICE = ""
EMBEDDING_DEVICE = ""

# supported LLM models
# llm_model_dict handles some preset behaviors of the loader, such as loading location, model name, and model processor instance
# Modify the property values ​​in the following dictionary to specify the local LLM model storage location
# For example, change the "local_model_path" of "chatglm-6b" from None to "User/Downloads/chatglm-6b"
# Please write absolute path here
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

    # For models called through fastchat, please refer to the following format
    "fastchat-chatglm-6b": {
        "name": "chatglm-6b",  # "name" is changed to "model_name" in the fastchat service
        "pretrained_model_name": "chatglm-6b",
        "local_model_path": "/root/exec/chatglm-6b",
        "provides": "FastChatOpenAILLM",  # When using fastchat api, make sure "provides" is "FastChatOpenAILLM"
        "api_base_url": "http://localhost:30000/v1"  # "name" is changed to "api_base_url" in the fastchat service
    },

    # For models called through fastchat, please refer to the following format
    "fastchat-vicuna-13b-hf": {
        "name": "vicuna-13b-hf",  # "name" is changed to "model_name" in the fastchat service
        "pretrained_model_name": "vicuna-13b-hf",
        "local_model_path": None,
        "provides": "FastChatOpenAILLM",  # When using fastchat api, make sure "provides" is "FastChatOpenAILLM"
        "api_base_url": "http://localhost:8000/v1"  # "name" is changed to "api_base_url" in the fastchat service
    },
}

# LLM name
LLM_MODEL = "chatglm2-6b"
# Quantitatively load 8bit model
LOAD_IN_8BIT = False
# Load the model with bfloat16 precision. Requires NVIDIA Ampere GPU.
BF16 = False
# The location where local lora is stored
LORA_DIR = "loras/"

# Number of cached knowledge bases
CACHED_VS_NUM = 1

# text clause length
SENTENCE_SIZE = 100

# The length of the single segment context after matching
CHUNK_SIZE = 300

# The maximum length of text in a chunk. Chunks whose len(text) exceeds this length will be discarded.
MAX_SENTENCE_SIZE = 4048
# The minimum length of text in the chunk for cascade splitting. Chunks whose len(text) is smaller than this length will not be cascaded.
MIN_SENTENCE_SIZE = 100

OCR_MAX_WORKERS = 1
#Model parsing service
MODEL_PARSER_MAX_WORKERS = 1

#add file
USER_DATA_PATH = "./user_data"
CONVERT_DIR = "/model_extend/convert"