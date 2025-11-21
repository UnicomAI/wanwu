import requests
import os

from enum import Enum
from logging_config import setup_logging
from settings import MODEL_PROVIDER_URL, MODEL_PROVIDER_ACCESS_TOKEN

logger_name = 'model_manager'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))


class ModelType(Enum):
    """
    Enum class for model type.
    """

    LLM = "llm"
    TEXT_EMBEDDING = "embedding"
    RERANK = "rerank"
    OCR = "ocr"


class ModelConfigure:
    """
    Model configuration class.
    """

    def __init__(self, model_type: ModelType, model_id: str, model_name: str, model_provider: str, model_args: dict):
        self.model_type = model_type
        self.model_id = model_id
        self.model_name = model_name
        self.model_args = model_args
        self.model_provider = model_provider

    def __str__(self):
        return f"ModelConfigure(model_type={self.model_type}, model_id={self.model_id}, model_name={self.model_name}, model_provider={self.model_provider}, model_args={self.model_args})"


class LlmModelConfig(ModelConfigure):
    """
    model_args: {
        "apiKey": {
            "description": "ApiKey",
            "type": "string"
        },
        "endpointUrl": {
            "description": "推理url",
            "type": "string"
        },
        "functionCalling": {
            "description": "函数调用是否支持",
            "type": "string",
            "enum": [
                "noSupport",
                "toolCall",
                "functionCall"
            ]
        }
    }
    """

    def __init__(self, model_id: str, model_name: str, model_provider: str, api_key: str, endpoint_url: str,
                 function_calling: str, model_args: dict):
        super().__init__(ModelType.LLM, model_id, model_name, model_provider, model_args)
        self.api_key = api_key
        self.endpoint_url = endpoint_url
        self.function_calling = function_calling

    @classmethod
    def from_api_response(cls, data: dict):
        provider = data.get("provider")
        llm_cfg = data.get("config", {})
        # if provider == "OpenAI-API-compatible" or provider == "YuanJing":
        #     llm_cfg = data.get("config", {})
        # else:
        #     raise ValueError(f"Unsupported provider: {provider}")
        return cls(
            model_id=data.get("modelId"),
            model_name=data.get("model"),
            model_provider=provider,
            api_key=llm_cfg.get("apiKey"),
            endpoint_url=llm_cfg.get("endpointUrl"),
            function_calling=llm_cfg.get("functionCalling"),
            model_args=llm_cfg
        )


class EmbeddingModelConfig(ModelConfigure):
    """
    model_args: {
        "apiKey": {
            "description": "ApiKey",
            "type": "string"
        },
        "endpointUrl": {
            "description": "推理url",
            "type": "string"
        }
    }
    """

    def __init__(self, model_id: str, model_name: str, model_provider: str, api_key: str, endpoint_url: str,
                 model_args: dict):
        super().__init__(ModelType.TEXT_EMBEDDING, model_id, model_name, model_provider, model_args)
        self.api_key = api_key
        self.endpoint_url = endpoint_url

    @classmethod
    def from_api_response(cls, data: dict):
        # provider = data.get("provider")
        emb_cfg = data.get("config", {})
        # if provider == "OpenAI-API-compatible" or provider == "YuanJing":
        #     emb_cfg = data.get("config", {})
        # else:
        #     raise ValueError(f"Unsupported provider: {provider}")
        return cls(
            model_id=data.get("modelId"),
            model_name=data.get("model"),
            model_provider="OpenAI-API-compatible",
            api_key=emb_cfg.get("apiKey"),
            endpoint_url=emb_cfg.get("endpointUrl"),
            model_args=emb_cfg
        )


class RerankModelConfig(ModelConfigure):
    """
    model_args: {
        "apiKey": {
            "description": "ApiKey",
            "type": "string"
        },
        "endpointUrl": {
            "description": "推理url",
            "type": "string"
        }
    }
    """

    def __init__(self, model_id: str, model_name: str, model_provider: str, api_key: str, endpoint_url: str,
                 model_args: dict):
        super().__init__(ModelType.RERANK, model_id, model_name, model_provider, model_args)
        self.api_key = api_key
        self.endpoint_url = endpoint_url

    @classmethod
    def from_api_response(cls, data: dict):
        # provider = data.get("provider")
        rerank_cfg = data.get("config", {})
        # if provider == "OpenAI-API-compatible" or provider == "YuanJing":
        #     rerank_cfg = data.get("config", {})
        # else:
        #     raise ValueError(f"Unsupported provider: {provider}")
        return cls(
            model_id=data.get("modelId"),
            model_name=data.get("model"),
            model_provider="OpenAI-API-compatible",  # The openapi format is hard-coded and specifically adapted by the bff layer.
            api_key=rerank_cfg.get("apiKey"),
            endpoint_url=rerank_cfg.get("endpointUrl"),
            model_args=rerank_cfg
        )


class OcrModelConfig(ModelConfigure):
    """
    model_args: {
        "apiKey": {
            "description": "ApiKey",
            "type": "string"
        },
        "endpointUrl": {
            "description": url",
            "type": "string"
        }
    }
    """

    def __init__(self, model_id: str, model_name: str, model_provider: str, api_key: str, endpoint_url: str,
                 model_args: dict):
        super().__init__(ModelType.OCR, model_id, model_name, model_provider, model_args)
        self.api_key = api_key
        self.endpoint_url = endpoint_url

    @classmethod
    def from_api_response(cls, data: dict):
        ocr_cfg = data.get("config", {})
        return cls(
            model_id=data.get("modelId"),
            model_name=data.get("model"),
            model_provider="OpenAI-API-compatible",  # The openapi format is hard-coded and specifically adapted by the bff layer.
            api_key=ocr_cfg.get("apiKey"),
            endpoint_url=ocr_cfg.get("endpointUrl"),
            model_args=ocr_cfg
        )


def get_model_configure(model_id: str) -> ModelConfigure:
    """
    Get model configuration by model id.
    """

    header = {
        'Content-Type': 'application/json',
        # "Authorization": "Bearer " + MODEL_PROVIDER_ACCESS_TOKEN
    }
    url = f"{MODEL_PROVIDER_URL}/callback/v1/model/{model_id}"
    try:
        resp = requests.get(url=url, headers=header, timeout=50)
        resp.raise_for_status()
        data = resp.json().get("data")
        if not data:
            raise RuntimeError("No model data returned!")

        model_type = data["modelType"]
        if model_type == "llm":
            model_config = LlmModelConfig.from_api_response(data)
        elif model_type == "embedding":
            model_config = EmbeddingModelConfig.from_api_response(data)
        elif model_type == "rerank":
            model_config = RerankModelConfig.from_api_response(data)
        elif model_type in ["ocr", "pdf-parser"]:
            model_config = OcrModelConfig.from_api_response(data)
        else:
            raise RuntimeError(f"Unsupported modelType: {model_type}")
        return model_config
    except Exception as e:
        logger.error("模型参数请求异常：" + repr(e))
        raise RuntimeError(f"Failed to get model configuration: {e}")
