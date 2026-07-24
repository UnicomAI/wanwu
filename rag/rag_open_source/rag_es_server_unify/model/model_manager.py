import requests
import os

from typing import Optional, Dict, Any, Type

from enum import Enum
from settings import MODEL_PROVIDER_URL
from log.logger import logger

TIMEOUT_SECONDS = 50

class ModelType(Enum):
    """模型类型枚举"""
    LLM = "llm"
    TEXT_EMBEDDING = "embedding"
    MULTIMODAL_EMBEDDING = "multimodal-embedding"
    RERANK = "rerank"
    MULTIMODAL_RERANK = "multimodal-rerank"
    OCR = "ocr"
    PDF_PARSER = "pdf-parser"
    ASR = "sync-asr"


class BaseModelConfig:
    """模型配置基类，包含所有模型共有的字段"""

    def __init__(self,
                 model_type: ModelType,
                 model_id: str,
                 model_name: str,
                 provider: str,
                 api_key: str,
                 endpoint_url: str,
                 original_args: Dict[str, Any]):
        self.model_type = model_type
        self.model_id = model_id
        self.model_name = model_name
        self.provider = provider
        self.api_key = api_key
        self.endpoint_url = endpoint_url
        self.original_args = original_args

    @property
    def is_vision_support(self) -> bool:
        """是否是图文问答模型"""
        return self.original_args.get("visionSupport", "noSupport") == "support"

    @property
    def is_multimodal(self) -> bool:
        """是否是多模态模型"""
        if self.model_type == ModelType.MULTIMODAL_EMBEDDING:
            return True
        if self.model_type == ModelType.MULTIMODAL_RERANK:
            return True
        if self.model_type == ModelType.LLM and self.is_vision_support:
            return True

        return False

    @property
    def context_window(self) -> Optional[int]:
        """模型最大上下文长度"""
        return self.original_args.get("context_window", 8000)

    @property
    def max_tokens(self) -> Optional[int]:
        """模型 max_token"""
        return self.original_args.get("max_token", 8000)

    @property
    def max_image_size(self) -> Optional[int]:
        """模型输入图片最大 size"""
        return self.original_args.get("max_image_size", 3 * 1024 * 1024)

    def __str__(self):
        return (f"<{self.__class__.__name__} type={self.model_type}, id={self.model_id}, "
                f"name={self.model_name}, provider={self.provider}>, config={self.original_args}")

    @staticmethod
    def _parse_common_data(data: Dict[str, Any]) -> Dict[str, Any]:
        """解析 API 响应中的通用字段"""
        config_data = data.get("config", {})
        return {
            "model_id": data.get("modelId"),
            "model_name": data.get("model"),
            "provider": data.get("provider"),
            "api_key": config_data.get("apiKey", ""),
            "endpoint_url": config_data.get("endpointUrl", ""),
            "original_args": config_data
        }

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        raise NotImplementedError("Subclasses must implement from_api_response")


class LlmModelConfig(BaseModelConfig):
    """LLM 模型配置"""

    def __init__(self, function_calling: str, **kwargs):
        super().__init__(model_type=ModelType.LLM, **kwargs)
        self.function_calling = function_calling

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        common_data = cls._parse_common_data(data)
        function_calling = common_data["original_args"].get("functionCalling", "noSupport")

        return cls(
            function_calling=function_calling,
            **common_data
        )


class EmbeddingModelConfig(BaseModelConfig):
    """Embedding 模型配置"""

    def __init__(self, model_type=ModelType.TEXT_EMBEDDING, **kwargs):
        super().__init__(model_type=model_type, **kwargs)

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        model_type = data.get("modelType")
        common_data = cls._parse_common_data(data)
        # Embedding 强制覆盖 provider 为 OpenAI-API-compatible
        common_data["provider"] = "OpenAI-API-compatible"
        if model_type == ModelType.MULTIMODAL_EMBEDDING.value:
            return cls(ModelType.MULTIMODAL_EMBEDDING, **common_data)
        return cls(**common_data)


class RerankModelConfig(BaseModelConfig):
    """Rerank 模型配置"""

    def __init__(self, model_type=ModelType.RERANK, **kwargs):
        super().__init__(model_type=model_type, **kwargs)

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        model_type = data.get("modelType")
        common_data = cls._parse_common_data(data)
        # Rerank 强制覆盖 provider 为 OpenAI-API-compatible
        common_data["provider"] = "OpenAI-API-compatible"
        if model_type == ModelType.MULTIMODAL_RERANK.value:
            return cls(ModelType.MULTIMODAL_RERANK, **common_data)
        return cls(**common_data)


class OcrModelConfig(BaseModelConfig):
    """OCR 模型配置"""

    def __init__(self, **kwargs):
        super().__init__(model_type=ModelType.OCR, **kwargs)

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        common_data = cls._parse_common_data(data)
        # OCR 强制覆盖 provider 为 OpenAI-API-compatible
        common_data["provider"] = "OpenAI-API-compatible"
        return cls(**common_data)


class AsrModelConfig(BaseModelConfig):
    """asr 模型配置"""

    def __init__(self, **kwargs):
        super().__init__(model_type=ModelType.ASR, **kwargs)

    @classmethod
    def from_api_response(cls, data: Dict[str, Any]):
        common_data = cls._parse_common_data(data)
        return cls(**common_data)

def get_model_configure(model_id: str) -> BaseModelConfig:
    """
    根据模型 ID 获取模型配置信息
    """
    url = f"{MODEL_PROVIDER_URL}/callback/v1/model/{model_id}"
    headers = {
        'Content-Type': 'application/json',
    }

    try:
        response = requests.get(url=url, headers=headers, timeout=TIMEOUT_SECONDS)
        response.raise_for_status()
        data = response.json().get("data")
        if not data:
            raise RuntimeError("No model data returned!")
        logger.info(f"get_model_configure response: model_id={model_id}, data={data}")
        model_type = data.get("modelType")
        if model_type == ModelType.LLM.value:
            return LlmModelConfig.from_api_response(data)
        elif model_type in [ModelType.TEXT_EMBEDDING.value, ModelType.MULTIMODAL_EMBEDDING.value]:
            return EmbeddingModelConfig.from_api_response(data)
        elif model_type in [ModelType.RERANK.value, ModelType.MULTIMODAL_RERANK.value]:
            return RerankModelConfig.from_api_response(data)
        elif model_type in [ModelType.OCR.value, ModelType.PDF_PARSER.value]:
            return OcrModelConfig.from_api_response(data)
        elif model_type == ModelType.ASR.value:
            return AsrModelConfig.from_api_response(data)
        else:
            raise ValueError(f"Unsupported modelType: {model_type}")

    except Exception as e:
        logger.error(f"Failed to fetch model config for ID {model_id}: {repr(e)}")
        raise RuntimeError(f"Failed to get model configuration: {e}")

def is_multimodal_model(model_id: str) -> bool:
    """
    判断模型是否支持多模态
    """
    model_config = get_model_configure(model_id)
    return model_config.is_multimodal