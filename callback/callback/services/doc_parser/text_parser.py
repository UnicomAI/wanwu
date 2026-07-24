"""纯文本类文档解析器：txt / md / csv / xml。

只读取文本，不做任何结构化切分。txt/md 按块流式读取以支持提前终止；
csv 逐行拼接；xml 用 iterparse 流式提取文本节点。
"""

import csv
import os
import xml.etree.ElementTree as ET

import chardet

from .base import DocumentParser
from .excel_parser import _escape_cell, _rows_to_markdown, _trim_trailing_empty


def _detect_encoding(file_path: str, sample_size: int = 10485760) -> str:
    """检测文件编码，移植自 rag 的 detect_file_encoding 逻辑。

    顺序：UTF-8 BOM → 严格 UTF-8 → chardet（含 gb18030 纠偏与黑名单）。
    """
    try:
        with open(file_path, "rb") as f:
            raw = f.read(sample_size)
    except Exception:
        return "utf-8"

    if not raw:
        return "utf-8"

    # 1. 优先检查 UTF-8 BOM
    if raw.startswith(b"\xef\xbb\xbf"):
        return "utf-8-sig"

    # 2. 尝试严格 UTF-8 解码
    try:
        raw.decode("utf-8")
        return "utf-8"
    except UnicodeDecodeError:
        pass

    # 3. 使用 chardet 分析
    result = chardet.detect(raw)
    encoding = (result.get("encoding") or "").lower()
    confidence = result.get("confidence") or 0

    if encoding:
        # 黑名单：这些编码在中文环境下 99% 都是误报
        misleading = ["windows-1254", "windows-1252", "macroman", "iso-8859-1", "ascii"]
        # 符合 GB 系列 / cp936 / 黑名单 / 置信度不足 → 纠偏为 gb18030
        if (
            encoding.startswith("gb")
            or encoding == "cp936"
            or encoding in misleading
            or confidence < 0.95
        ):
            return "gb18030"
        return encoding

    return "gb18030"


class TextParser(DocumentParser):
    """txt / md 解析器：按块流式读取，达到上限即停止。"""

    _CHUNK_SIZE = 65536  # 64KB

    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        limit = self.max_chars(max_tokens)
        encoding = _detect_encoding(file_path)
        parts = []
        total = 0
        with open(file_path, "r", encoding=encoding, errors="ignore") as f:
            while True:
                chunk = f.read(self._CHUNK_SIZE)
                if not chunk:
                    break
                parts.append(chunk)
                total += len(chunk)
                if limit is not None and total >= limit:
                    break
        text = "".join(parts)
        return self.truncate(text, max_tokens)


class CsvParser(DocumentParser):
    """csv 解析器：输出 markdown 表格。

    多行时第一行作为表头，其余作数据行，渲染为 markdown 表格；只有 1 行时，
    单列直接输出原始值，多列则该行作表头行渲染为只有表头的表格。
    """

    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        limit = self.max_chars(max_tokens)
        encoding = _detect_encoding(file_path)
        parts = []
        total = 0

        def append(text: str) -> bool:
            nonlocal total
            parts.append(text)
            total += len(text)
            return limit is not None and total >= limit

        # newline='' 避免 csv 行内换行被二次处理
        with open(file_path, "r", encoding=encoding, errors="ignore", newline="") as f:
            rows = list(csv.reader(f))
        rows = [r for r in rows if any(c.strip() for c in r)]
        if not rows:
            return ""

        matrix = _trim_trailing_empty(rows)
        ncol = max((len(r) for r in matrix), default=0)

        # 单行单列：直接输出原始值，不套表格
        if len(matrix) == 1 and ncol <= 1:
            append(_escape_cell(matrix[0][0] if matrix[0] else None) + "\n")
        else:
            headers = [_escape_cell(v) for v in matrix[0]]
            data_rows = [[_escape_cell(v) for v in r] for r in matrix[1:]]
            for line in _rows_to_markdown(data_rows, headers):
                if append(line):
                    break
            append("\n")
        text = "".join(parts)
        return self.truncate(text, max_tokens)


class XmlParser(DocumentParser):
    """xml 解析器：用 iterparse 流式提取文本节点，避免整树载入内存。"""

    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        limit = self.max_chars(max_tokens)
        parts = []
        total = 0
        # iterparse 逐事件消费，按 XML 声明的编码解析，clear 释放已处理节点
        context = ET.iterparse(file_path, events=("end",))
        for _event, elem in context:
            if elem.text and elem.text.strip():
                snippet = elem.text.strip() + "\n"
                parts.append(snippet)
                total += len(snippet)
                if limit is not None and total >= limit:
                    break
            elem.clear()
        text = "".join(parts)
        return self.truncate(text, max_tokens)
