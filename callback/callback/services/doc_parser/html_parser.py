"""HTML 解析器：用 beautifulsoup4 提取可见文本。

HTML 无法流式提取文本，因此一次性读取后截断（非提前终止）。
"""

from bs4 import BeautifulSoup

from .base import DocumentParser
from .text_parser import _detect_encoding


class HtmlParser(DocumentParser):
    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        encoding = _detect_encoding(file_path)
        with open(file_path, "r", encoding=encoding, errors="ignore") as f:
            html = f.read()
        soup = BeautifulSoup(html, "html.parser")
        # 移除 script/style 等非正文标签
        for tag in soup(["script", "style"]):
            tag.decompose()
        text = soup.get_text(separator="\n")
        # 合并多余空行
        lines = [line.strip() for line in text.splitlines() if line.strip()]
        text = "\n".join(lines)
        return self.truncate(text, max_tokens)
