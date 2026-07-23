"""文档解析器统一接口。

各格式解析器实现 ``DocumentParser.parse``，只负责把文档读取为纯文本，
不做切分、OCR、ASR 等额外处理。大文件可通过 ``max_tokens`` 提前终止读取。
"""

from abc import ABC, abstractmethod

from utils.tokenizers import CustomTokenizer


class DocumentParser(ABC):
    """文档解析器抽象基类。"""

    # 字符到 token 的转换比例，与 CustomTokenizer 保持一致，避免两处漂移。
    _CHAR_TO_TOKEN_RATIO = CustomTokenizer._default_char_to_token_ratio

    @abstractmethod
    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        """提取文档纯文本。

        Args:
            file_path: 本地文件路径。
            max_tokens: 最大 token 数；``<=0`` 表示不截断，读取全部内容。
                        ``>0`` 时按 ``max_tokens * _CHAR_TO_TOKEN_RATIO`` 字符数截断，
                        实现应尽量在达到上限时提前终止读取以提升效率。

        Returns:
            提取出的纯文本。
        """

    @classmethod
    def max_chars(cls, max_tokens: int):
        """根据 max_tokens 计算最大字符数；``<=0`` 返回 ``None`` 表示不限。"""
        if max_tokens is None or max_tokens <= 0:
            return None
        return int(max_tokens * cls._CHAR_TO_TOKEN_RATIO)

    @classmethod
    def truncate(cls, text: str, max_tokens: int) -> str:
        """对已读取的文本按 max_tokens 截断；用于无法流式提前终止的格式。"""
        limit = cls.max_chars(max_tokens)
        if limit is None or len(text) <= limit:
            return text
        return text[:limit]
