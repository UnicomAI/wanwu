"""PDF 解析器：用 pdfplumber 逐页提取文本与表格，达到上限即停止。

表格处理参照 rag PDFLoader：检测到表格时，把表格区域从普通文本中过滤掉，
单独将表格转为 markdown 文本拼入，避免表格内容被 extract_text 抽成散乱字符。
"""

import pdfplumber

from utils.log import logger

from .base import DocumentParser


def _table_to_markdown(rows) -> str:
    """把 pdfplumber 表格二维数据转为 markdown 文本。"""
    if not rows:
        return ""
    lines = []
    for r_idx, row in enumerate(rows):
        cells = [("" if c is None else str(c).replace("\n", " ").strip()) for c in row]
        lines.append("| " + " | ".join(cells) + " |")
        if r_idx == 0:
            lines.append("| " + " | ".join(["---"] * len(cells)) + " |")
    return "\n".join(lines)


class PdfParser(DocumentParser):
    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        limit = self.max_chars(max_tokens)
        parts = []
        total = 0
        with pdfplumber.open(file_path) as pdf:
            for page in pdf.pages:
                page_parts = self._extract_page(page)
                if not page_parts:
                    continue
                for snippet in page_parts:
                    if not snippet:
                        continue
                    parts.append(snippet)
                    parts.append("\n")
                    total += len(snippet) + 1
                    if limit is not None and total >= limit:
                        return self.truncate("".join(parts), max_tokens)
        text = "".join(parts)
        logger.info(f"PDF 解析完成, 文本长度={len(text)}")
        return self.truncate(text, max_tokens)

    @staticmethod
    def _extract_page(page) -> list:
        """提取单页文本：表格转为 markdown，表格区域从普通文本中过滤掉。"""
        tables = page.find_tables()
        bboxes = [table.bbox for table in tables]

        if not bboxes:
            return [page.extract_text() or ""]

        def _not_within_bboxes(obj):
            """判断对象是否落在任一表格 bbox 内（用于过滤掉表格区域的散字）。"""
            for bbox in bboxes:
                x0, top, x1, bottom = bbox
                if obj.get("x0", 0) <= x1 and obj.get("x1", 0) >= x0 \
                        and obj.get("top", 0) <= bottom and obj.get("bottom", 0) >= top:
                    return False
            return True

        # 表格区域外的文本
        outside_text = page.filter(_not_within_bboxes).extract_text() or ""
        result = [outside_text] if outside_text else []

        # 各表格转 markdown
        for table in tables:
            md = _table_to_markdown(table.extract())
            if md:
                result.append(md)
        return result
