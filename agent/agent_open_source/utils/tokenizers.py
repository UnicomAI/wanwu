import transformers
from typing import Union, List
import torch
from transformers import AutoTokenizer
import time
import os
import logging
logger = logging.getLogger(__name__)

class CustomTokenizer:
    # Use class variables to store model instances, ensuring there is only one instance of each model
    _instances = {}
    # Supported model list
    _supported_models = [
        "deepseek-r1",
        "deepseek-r1-distill-llama-8b",
        "deepseek-r1-distill-llama-70b",
        "deepseek-r1-distill-qwen-1.5b",
        "deepseek-r1-distill-qwen-14b",
        "deepseek-r1-distill-qwen-32b",
        "deepseek-r1-distill-qwen-7b",
        "deepseek-v3",
        "qwq-32b",
        "yuanjing-34b-chat",
        "yuanjing-70b-chat",
        "qwen3-8b"
    ]
    # Default character to token conversion ratio
    _default_char_to_token_ratio = 1.7

    def __new__(cls, tokenizer_dir: str = "/agent/agent_open_source/utils/tokenizers", model_name: str = "deepseek-v3", device: str = None):
        """确保每个模型实例是单例"""
        if model_name not in cls._instances:
            cls._instances[model_name] = super().__new__(cls)
        return cls._instances[model_name]

    def __init__(self, tokenizer_dir: str = "/agent/agent_open_source/utils/tokenizers", model_name: str = "deepseek-v3", device: str = None):
        """初始化tokenizer"""
        model_name = "deepseek-v3"
        if not hasattr(self, 'tokenizer'):  # Make sure to initialize only once
            self.tokenizer = None
            self.device = None
            self.model_name = model_name
            self.is_supported_model = model_name in self._supported_models
            self._initialize(tokenizer_dir, model_name, device)

    def _initialize(self, tokenizer_dir: str, model_name: str, device: str):
        """内部初始化方法"""
        start_time = time.time()
        self.device = device or ('cuda' if torch.cuda.is_available() else 'cpu')
        
        if self.is_supported_model:
            tokenizer_path = os.path.join(tokenizer_dir, model_name)
            # print(f"Tokenizer path: {tokenizer_path}")
            self.tokenizer = AutoTokenizer.from_pretrained(
                tokenizer_path, trust_remote_code=True
            )
            logger.info(f"Tokenizer for {model_name} initialized in {time.time() - start_time:.4f}s")
        else:
            logger.info(f"Model {model_name} not in supported list, using character-based estimation")

    def count_tokens(self, text: Union[str, List[str]]) -> int:
        """计算文本的token数量"""
        start_time = time.time()
        result = None
        
        if not self.is_supported_model:
            # For unsupported models, use character length estimates
            if isinstance(text, str):
                result = int(len(text) / self._default_char_to_token_ratio)
            elif isinstance(text, list):
                result = sum(int(len(t) / self._default_char_to_token_ratio) for t in text)
            else:
                raise ValueError("输入必须是字符串或字符串列表")
        else:
            # For supported models, use tokenizer calculations
            if isinstance(text, str):
                result = len(self.tokenizer.encode(text))
            elif isinstance(text, list):
                result = sum(len(self.tokenizer.encode(t)) for t in text)
            else:
                raise ValueError("输入必须是字符串或字符串列表")
                
        # logger.info(f"Calculating the number of tokens takes: {time.time() - start_time:.4f} seconds")
        return result

    def encode(self, text: str) -> List[int]:
        """将文本编码为token ID列表"""
        start_time = time.time()
        
        if not self.is_supported_model:
            # For unsupported models, return a simulated token list
            tokens = [0] * int(len(text) / self._default_char_to_token_ratio)
            logger.warning(f"Model {self.model_name} not supported for actual encoding, returning estimated token count")
        else:
            tokens = self.tokenizer.encode(text)
            if self.device == 'cuda':
                tokens = torch.tensor(tokens).to(self.device)
                
        logger.info(f"编码耗时: {time.time() - start_time:.4f}秒")
        return tokens

    def decode(self, token_ids: List[int]) -> str:
        """将token ID列表解码为文本"""
        start_time = time.time()
        
        if not self.is_supported_model:
            # For unsupported models, return a warning message
            result = "[不支持的模型无法进行实际解码]"
            logger.warning(f"Model {self.model_name} not supported for decoding")
        else:
            result = self.tokenizer.decode(token_ids)
            
        logger.info(f"解码耗时: {time.time() - start_time:.4f}秒")
        return result

    
    # Remove instance method is_model_supported
    
    @classmethod
    def is_model_in_supported_list(cls, model_name: str) -> bool:
        """
        检查指定的模型是否在支持列表中
        
        Args:
            model_name: 要检查的模型名称
            
        Returns:
            bool: 如果模型在支持列表中返回True，否则返回False
        """
        return model_name in cls._supported_models

    def truncate_text(self, text: str, max_tokens: int) -> str:
        """按照指定的最大token数量截断文本"""
        start_time = time.time()
        if max_tokens < 1:
            raise ValueError("max_tokens必须大于0")

        if not self.is_supported_model:
            # For unsupported models, use character estimate truncation
            max_chars = int(max_tokens * self._default_char_to_token_ratio)
            result = text[:max_chars]
            logger.warning(f"Model {self.model_name} not supported for actual truncation, using character estimation")
        else:
            token_ids = self.tokenizer.encode(text, add_special_tokens=False)
            if len(token_ids) <= max_tokens:
                result = text
            else:
                truncated_ids = token_ids[:max_tokens]
                decoded_text = self.tokenizer.decode(truncated_ids, skip_special_tokens=True)
                # Remove redundant special prefixes (such as BOS, etc.)
                result = decoded_text.lstrip("<|begin▁of▁sentence|>").lstrip()  # If you need more flexibility, you can use regular processing

        logger.info(f"文本截断总耗时: {time.time() - start_time:.4f}秒")
        return result
    
    # def truncate_text(self, text: str, max_tokens: int) -> str:
    #     """Truncate text according to the specified maximum number of tokens"""
    #     start_time = time.time()
    #     if max_tokens < 1:
    #         raise ValueError("max_tokens must be greater than 0")

    #     if not self.is_supported_model:
    #         # For unsupported models, use character estimate truncation
    #         max_chars = int(max_tokens * self._default_char_to_token_ratio)
    #         result = text[:max_chars]
    #         logger.warning(f"Model {self.model_name} not supported for actual truncation, using character estimation")
    #     else:
    #         token_ids = self.encode(text)
    #         if len(token_ids) <= max_tokens:
    #             result = text
    #         else:
    #             truncated_ids = token_ids[:max_tokens]
    #             result = self.decode(truncated_ids)

    #     logger.info(f"Total time spent on text truncation: {time.time() - start_time:.4f} seconds")
    #     return result

if __name__ == "__main__":
    r1_tokenizer = CustomTokenizer(model_name="deepseek-r1")
    v3_tokenizer = CustomTokenizer(model_name="deepseek-v3")
    qwq_tokenizer = CustomTokenizer(model_name="qwq-32b")
    deepseek_r1_distill_qwen_32b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-qwen-32b")
    deepseek_r1_distill_qwen_1_5b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-qwen-1.5b")
    deepseek_r1_distill_llama_8b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-llama-8b")
    deepseek_r1_distill_qwen_7b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-qwen-7b")
    deepseek_r1_distill_llama_70b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-llama-70b")
    deepseek_r1_distill_qwen_14b_tokenizer = CustomTokenizer(model_name="deepseek-r1-distill-qwen-14b")
    yuanjing_34b_chat_tokenizer = CustomTokenizer(model_name="yuanjing-34b-chat")
    yuanjing_70b_chat_tokenizer = CustomTokenizer(model_name="yuanjing-70b-chat")
    other_tokenizer = CustomTokenizer(model_name="some-other-model")

    text = "中华人民共和国是一个伟大的国家!"
    
    print(f"deepseek-r1 tokens: {r1_tokenizer.count_tokens(text)}")
    print(f"deepseek-v3 tokens: {v3_tokenizer.count_tokens(text)}")
    print(f"qwq-32b tokens: {qwq_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-qwen-32b tokens: {deepseek_r1_distill_qwen_32b_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-qwen-1.5b tokens: {deepseek_r1_distill_qwen_1_5b_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-llama-8b tokens: {deepseek_r1_distill_llama_8b_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-qwen-7b tokens: {deepseek_r1_distill_qwen_7b_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-llama-70b tokens: {deepseek_r1_distill_llama_70b_tokenizer.count_tokens(text)}")
    print(f"deepseek-r1-distill-qwen-14b tokens: {deepseek_r1_distill_qwen_14b_tokenizer.count_tokens(text)}")
    print(f"yuanjing-34b-chat tokens: {yuanjing_34b_chat_tokenizer.count_tokens(text)}")
    print(f"yuanjing-70b-chat tokens: {yuanjing_70b_chat_tokenizer.count_tokens(text)}")
    print(f"other model tokens: {other_tokenizer.count_tokens(text)}")
    
    # Test model support checks
    print(f"Is deepseek-v3 supported: {CustomTokenizer.is_model_in_supported_list('deepseek-v3')}")
    print(f"Is yuanjing-34b-chat supported: {CustomTokenizer.is_model_in_supported_list('yuanjing-34b-chat')}")
    
    print(f"Is some-other-model supported: {CustomTokenizer.is_model_in_supported_list('some-other-model')}")
    print(f"Is current model supported: {deepseek_r1_distill_qwen_32b_tokenizer.is_supported_model}")  # Using instance properties
