"""文档解析器工厂：按文件后缀分发到对应的解析器实现。"""

import os

from utils.log import logger
from utils.response import BizError

from .base import DocumentParser


class ParserFactory:
    """按文件后缀查找解析器的工厂。"""

    # 后缀(小写，含点) -> 解析器类
    _registry: dict = {}

    @classmethod
    def register(cls, ext: str, parser_cls: type) -> None:
        """注册某个后缀到解析器类。"""
        cls._registry[ext.lower()] = parser_cls

    @classmethod
    def get_parser(cls, file_path: str) -> DocumentParser:
        """根据文件后缀返回解析器实例；未支持的后缀抛 BizError。"""
        ext = os.path.splitext(file_path)[1].lower()
        parser_cls = cls._registry.get(ext)
        if parser_cls is None:
            raise BizError(f"不支持的文件类型: {ext or '无后缀'}")
        return parser_cls()

    @classmethod
    def is_supported(cls, file_path: str) -> bool:
        """判断该文件后缀是否有本地解析器。"""
        ext = os.path.splitext(file_path)[1].lower()
        return ext in cls._registry

    @classmethod
    def parse_file(cls, file_path: str, max_tokens: int = 0) -> str:
        """便捷方法：按后缀分发解析并返回纯文本。"""
        parser = cls.get_parser(file_path)
        logger.info(f"解析文档: {file_path}, max_tokens={max_tokens}")
        return parser.parse(file_path, max_tokens)
