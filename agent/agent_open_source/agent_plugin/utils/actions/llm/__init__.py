import re
from typing import Dict, Optional

from utils.actions.llm.base import LLM_REGISTRY

from .base import BaseChatModel, ModelServiceError
from .unicom import UnicomLLM
# from .oai import TextChatAtOAI
# from .openvino import OpenVINO
# from .qwen_dashscope import QwenChatAtDS
# from .qwenvl_dashscope import QwenVLChatAtDS


def get_chat_model(cfg: Optional[Dict] = None) -> BaseChatModel:
    """The interface of instantiating LLM objects.

    Args:
        cfg: The LLM configuration, one example is:
          llm_cfg = {
            # Use the model service provided by DashScope:
            'model': 'qwen-max',
            'model_server': 'dashscope',
            # Use your own model service compatible with OpenAI API:
            # 'model': 'Qwen',
            # 'model_server': 'http://127.0.0.1:7905/v1',
            # (Optional) LLM hyper-parameters:
            'generate_cfg': {
                'top_p': 0.8
            }
          }

    Returns:
        LLM object.
    """
    cfg = cfg or {}
    model = cfg.get("model",'')
    model_type = cfg.get("model_type",'')
    model_server = cfg.get("model_server",'')
    generate_cfg = cfg.get("generate_cfg",'')
    if model_type:
        if model_type in LLM_REGISTRY:
            return LLM_REGISTRY[model_type](cfg)
        else:
            raise ValueError(f'Please set model_type from {str(LLM_REGISTRY.keys())}')

    # Deduce model_type from model and model_server if model_type is not provided:

    if 'model_server' in cfg:
        if cfg['model_server'].strip().startswith('http'):
            model_type = 'unicomllm'
            return LLM_REGISTRY[model_type](cfg)

    # model = cfg.get('model', '')
    if 'qwen-vl' in model:
        model_type = 'qwenvl_dashscope'
        return LLM_REGISTRY[model_type](cfg)

    if 'qwen' in model:
        model_type = 'qwen_dashscope'
        return LLM_REGISTRY[model_type](cfg)

    raise ValueError(f'Invalid model cfg: {cfg}')


__all__ = [
    'BaseChatModel',
    'QwenChatAtDS',
    'TextChatAtOAI',
    'QwenVLChatAtDS',
    'OpenVINO',
    'get_chat_model',
    'ModelServiceError',
]
