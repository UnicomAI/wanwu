"""Word 文档解析器：docx 与 doc。

docx 用 python-docx 遍历段落与表格，异常时回退 docx2txt；
表格解析沿用 wanwu 项目原 rag 实现的合并单元格处理逻辑(_parse_row 等)，保持原状。
doc(老格式)先试 docx2txt，再尽力调用系统 antiword/catdoc，均失败则抛业务异常。
"""

import os
import subprocess

import docx2txt
from docx import Document as DocxDocument

from utils.log import logger
from utils.response import BizError

from .base import DocumentParser


def _table_to_markdown(table) -> str:
    """把 docx 表格转为 markdown 文本，沿用 rag 原版的合并单元格处理。"""

    def _parse_cell_paragraph(paragraph):
        paragraph_content = []
        for run in paragraph.runs:
            paragraph_content.append(run.text)
        return "".join(paragraph_content).strip()

    def _parse_cell(cell):
        cell_content = []
        for paragraph in cell.paragraphs:
            parsed_paragraph = _parse_cell_paragraph(paragraph)
            if parsed_paragraph:
                cell_content.append(parsed_paragraph)
        unique_content = list(dict.fromkeys(cell_content))
        return " ".join(unique_content)

    def _parse_row(row, total_cols):
        # Initialize a row, all of which are empty by default
        row_cells = [""] * total_cols
        processed_cell_ids = set()  # 用于跟踪已经处理过的单元格对象
        col_index = 0

        for cell in row.cells:
            # 使用 id() 来唯一标识每个单元格对象，因为 python-docx 对合并单元格会返回同一个对象
            cell_id = id(cell)

            # 跳过已经处理过的单元格对象
            if cell_id in processed_cell_ids:
                continue

            processed_cell_ids.add(cell_id)

            # 找到下一个空列的位置
            while col_index < total_cols and row_cells[col_index] != "":
                col_index += 1

            # if col_index is out of range the loop is jumped
            if col_index >= total_cols:
                break

            cell_content = _parse_cell(cell).strip()
            cell_colspan = cell.grid_span or 1
            cell_colspan = min(cell_colspan, total_cols - col_index)  # 确保不超过总列数

            # 填充单元格内容到相应列
            for i in range(cell_colspan):
                if col_index + i < total_cols:
                    row_cells[col_index + i] = cell_content if i == 0 else ""

            col_index += cell_colspan

        return row_cells

    # calculate the total number of columns
    total_cols = max(len(row.cells) for row in table.rows)
    header_row = table.rows[0]
    headers = _parse_row(header_row, total_cols)
    markdown = ["| " + " | ".join(headers) + " |"]
    markdown.append("| " + " | ".join(["---"] * total_cols) + " |")

    # 收集所有行，用于后续去重
    all_rows = []
    for row in table.rows[1:]:
        row_cells = _parse_row(row, total_cols)
        all_rows.append(row_cells)

    # 对完全相同的行进行去重
    seen_rows = []
    for row_cells in all_rows:
        row_str = "| " + " | ".join(row_cells) + " |"
        if row_str not in seen_rows:
            seen_rows.append(row_str)
            markdown.append(row_str)
    return "\n".join(markdown)


class DocxParser(DocumentParser):
    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        limit = self.max_chars(max_tokens)
        parts = []
        total = 0
        try:
            doc = DocxDocument(file_path)
            # 按文档顺序遍历段落与表格：通过 body 子元素判断类型
            from docx.oxml.text.paragraph import CT_P
            from docx.oxml.table import CT_Tbl
            from docx.table import Table
            from docx.text.paragraph import Paragraph

            for child in doc.element.body:
                if isinstance(child, CT_P):
                    para = Paragraph(child, doc)
                    text = para.text.strip()
                    if text:
                        # 对齐 rag docx_to_markdown：段落末尾两个空格 + 两换行(markdown 硬换行)
                        parts.append(text)
                        parts.append("  \n\n")
                        total += len(text) + 4
                        if limit is not None and total >= limit:
                            break
                elif isinstance(child, CT_Tbl):
                    table = Table(child, doc)
                    text = _table_to_markdown(table)
                    if text:
                        parts.append("\n")
                        parts.append(text)
                        parts.append("\n")
                        total += len(text) + 2
                        if limit is not None and total >= limit:
                            break
        except Exception as err:
            logger.warning(f"python-docx 解析失败，回退 docx2txt: {err}")
            parts = [docx2txt.process(file_path)]
        text = "".join(parts)
        return self.truncate(text, max_tokens)


class DocParser(DocumentParser):
    """老式 .doc 格式：尽力而为，无可靠纯 Python 库时给清晰错误。"""

    def parse(self, file_path: str, max_tokens: int = 0) -> str:
        text = ""
        # 1. 先试 docx2txt（对伪 docx / RTF 偶尔有效）
        try:
            text = docx2txt.process(file_path) or ""
        except Exception:
            text = ""

        # 2. 试系统命令 antiword / catdoc（尽力，不作依赖）
        if not text.strip():
            for cmd in ("antiword", "catdoc"):
                try:
                    completed = subprocess.run(
                        [cmd, file_path],
                        capture_output=True,
                        timeout=60,
                        text=True,
                    )
                    if completed.returncode == 0 and completed.stdout.strip():
                        text = completed.stdout
                        break
                except FileNotFoundError:
                    continue
                except Exception as err:
                    logger.warning(f"{cmd} 解析 .doc 失败: {err}")

        if not text.strip():
            raise BizError("无法解析 .doc 文件，建议转换为 .docx 后再上传")

        return self.truncate(text, max_tokens)

