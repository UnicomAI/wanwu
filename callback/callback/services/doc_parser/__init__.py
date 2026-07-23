"""文档解析器工厂包：按文件后缀分发到对应解析器，只提取纯文本。

用法：
    from callback.services.doc_parser import parse_file, download_to_tempfile
    text = parse_file("/path/to/file.pdf", max_tokens=20480)
"""

import os
import urllib.parse

import requests

from utils.log import logger
from utils.response import BizError

from .base import DocumentParser
from .docx_parser import DocParser, DocxParser
from .excel_parser import XlsParser, XlsxParser
from .factory import ParserFactory
from .html_parser import HtmlParser
from .pdf_parser import PdfParser
from .pptx_parser import PptxParser
from .text_parser import CsvParser, TextParser, XmlParser


def _register_defaults() -> None:
    """注册默认支持的文件后缀到对应解析器。"""
    mapping = {
        ".txt": TextParser,
        ".md": TextParser,
        ".csv": CsvParser,
        ".xml": XmlParser,
        ".pdf": PdfParser,
        ".docx": DocxParser,
        ".doc": DocParser,
        ".xlsx": XlsxParser,
        ".xls": XlsParser,
        ".html": HtmlParser,
        ".htm": HtmlParser,
        ".pptx": PptxParser,
    }
    for ext, parser_cls in mapping.items():
        ParserFactory.register(ext, parser_cls)


_register_defaults()


def parse_file(file_path: str, max_tokens: int = 0) -> str:
    """按文件后缀解析本地文档，返回纯文本。"""
    return ParserFactory.parse_file(file_path, max_tokens)


def download_to_tempfile(file_url: str):
    """把 URL 文件流式下载到本地临时文件。

    Returns:
        (file_path, file_name): 本地临时文件路径与解码后的文件名。
        调用方负责在使用后删除 file_path。

    Raises:
        BizError: URL 无效或下载失败。
    """
    if not file_url:
        raise BizError("文件下载链接为空")

    parsed = urllib.parse.urlparse(file_url)
    file_name = urllib.parse.unquote(parsed.path.split("/")[-1]) or "download"
    suffix = os.path.splitext(file_name)[1] or ""

    fd, file_path = _mkstemp(suffix)
    try:
        # verify=False 与原 parse_doc 行为一致，适配 MinIO 内网自签证书场景
        response = requests.get(file_url, stream=True, timeout=(10, 30), verify=False)
        response.raise_for_status()
        with os.fdopen(fd, "wb") as f:
            for chunk in response.iter_content(chunk_size=8192):
                if chunk:
                    f.write(chunk)
    except Exception as err:
        # 下载失败需自行清理临时文件
        _safe_unlink(file_path)
        logger.error(f"下载文件失败 {file_url}: {err}")
        raise BizError(f"下载文件失败: {err}")

    return file_path, file_name


def _mkstemp(suffix: str):
    """创建临时文件，返回 (fd, path)。隔离以便测试 mock。"""
    import tempfile

    return tempfile.mkstemp(suffix=suffix)


def _safe_unlink(path: str) -> None:
    try:
        os.unlink(path)
    except OSError:
        pass


__all__ = [
    "DocumentParser",
    "ParserFactory",
    "parse_file",
    "download_to_tempfile",
]
