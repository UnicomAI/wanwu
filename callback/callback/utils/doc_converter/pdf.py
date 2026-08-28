"""Markdown 转 PDF（使用 reportlab platypus 直接渲染）"""

import io
import os
import re

from reportlab.lib.colors import HexColor
from reportlab.lib.enums import TA_JUSTIFY
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle
from reportlab.lib.units import mm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (
    HRFlowable,
    Paragraph,
    SimpleDocTemplate,
    Spacer,
    Table,
    TableStyle,
)

from callback.utils.doc_converter.common import parse_table


def _xml_escape(text: str) -> str:
    """转义 reportlab Paragraph 中的 XML 特殊字符"""
    return (
        text.replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
    )


def _xml_escape_preserve_spaces(text: str) -> str:
    """转义 XML 特殊字符，并将空格转为 &nbsp; 以保留缩进（用于代码块）"""
    return (
        text.replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
        .replace("  ", "&nbsp;&nbsp;")  # 成对空格转为 &nbsp;
        .replace(" \t", "&nbsp;&nbsp;&nbsp;&nbsp;")  # 空格+tab
        .replace("\t", "&nbsp;&nbsp;&nbsp;&nbsp;")  # tab 转 4 个 &nbsp;
    )


def _md_inline_to_reportlab(text: str) -> str:
    """将 Markdown 行内格式转换为 reportlab Paragraph 的 XML 标记"""
    parts = []
    last_end = 0
    inline_re = re.compile(r"(\*\*\*(.+?)\*\*\*|\*\*(.+?)\*\*|\*(.+?)\*|`(.+?)`)")

    for match in inline_re.finditer(text):
        if match.start() > last_end:
            parts.append(_xml_escape(text[last_end : match.start()]))

        if match.group(2):  # ***粗斜体***
            parts.append(f"<b><i>{_xml_escape(match.group(2))}</i></b>")
        elif match.group(3):  # **粗体**
            parts.append(f"<b>{_xml_escape(match.group(3))}</b>")
        elif match.group(4):  # *斜体*
            parts.append(f"<i>{_xml_escape(match.group(4))}</i>")
        elif match.group(5):  # `行内代码`
            parts.append(
                f'<font name="Courier" size="10">{_xml_escape(match.group(5))}</font>'
            )

        last_end = match.end()

    if last_end < len(text):
        parts.append(_xml_escape(text[last_end:]))

    return "".join(parts) if parts else _xml_escape(text)


def _build_pdf_styles():
    """构建 PDF 文档样式集"""
    return {
        "h1": ParagraphStyle(
            "H1", fontName="SimHei", fontSize=22, leading=28,
            spaceAfter=12, textColor=HexColor("#333333"),
        ),
        "h2": ParagraphStyle(
            "H2", fontName="SimHei", fontSize=18, leading=24,
            spaceAfter=10, textColor=HexColor("#333333"),
        ),
        "h3": ParagraphStyle(
            "H3", fontName="SimHei", fontSize=14, leading=20,
            spaceAfter=8, textColor=HexColor("#333333"),
        ),
        "h4": ParagraphStyle(
            "H4", fontName="SimHei", fontSize=12, leading=16,
            spaceAfter=6, textColor=HexColor("#333333"),
        ),
        "normal": ParagraphStyle(
            "Normal", fontName="SimHei", fontSize=12, leading=18,
            alignment=TA_JUSTIFY,
        ),
        "code": ParagraphStyle(
            "Code", fontName="Courier", fontSize=10, leading=14,
            backColor=HexColor("#F5F5F5"), leftIndent=10, rightIndent=10,
            spaceBefore=8, spaceAfter=8,
        ),
        "quote": ParagraphStyle(
            "Quote", fontName="SimHei", fontSize=12, leading=18,
            leftIndent=20, textColor=HexColor("#666666"),
        ),
        "bullet": ParagraphStyle(
            "Bullet", fontName="SimHei", fontSize=12, leading=18,
            leftIndent=20, bulletIndent=8,
        ),
    }


def _build_table_style():
    """构建 PDF 表格样式"""
    return TableStyle([
        ("FONTNAME", (0, 0), (-1, -1), "SimHei"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("ALIGN", (0, 0), (-1, -1), "LEFT"),
        ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
        ("GRID", (0, 0), (-1, -1), 0.5, HexColor("#666666")),
        ("BACKGROUND", (0, 0), (-1, 0), HexColor("#F0F0F0")),
        ("BOLD", (0, 0), (-1, 0), True),
        ("TOPPADDING", (0, 0), (-1, -1), 6),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 6),
        ("LEFTPADDING", (0, 0), (-1, -1), 8),
        ("RIGHTPADDING", (0, 0), (-1, -1), 8),
    ])


def markdown_to_pdf(md_content: str) -> bytes:
    """
    将 Markdown 转换为 PDF（使用 reportlab platypus 直接渲染）

    支持：标题、粗体/斜体/行内代码、无序/有序列表、表格、代码块、引用、分隔线
    """
    # 注册中文字体
    # __file__ = .../callback/callback/utils/doc_converter/pdf.py
    # 需要上溯到 callback/callback/，再拼 static/simhei.ttf
    font_path = os.path.join(
        os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))),
        "static",
        "simhei.ttf",
    )
    pdfmetrics.registerFont(TTFont("SimHei", font_path))

    styles = _build_pdf_styles()

    # 构建文档
    pdf_buffer = io.BytesIO()
    doc = SimpleDocTemplate(
        pdf_buffer,
        pagesize=A4,
        leftMargin=20 * mm,
        rightMargin=20 * mm,
        topMargin=20 * mm,
        bottomMargin=20 * mm,
    )

    story = []
    lines = md_content.split("\n")
    i = 0

    while i < len(lines):
        line = lines[i]

        # 空行
        if not line.strip():
            i += 1
            continue

        # 标题
        if line.startswith("#"):
            level = min(len(line) - len(line.lstrip("#")), 4)
            text = line.lstrip("#").strip()
            story.append(Paragraph(_md_inline_to_reportlab(text), styles[f"h{level}"]))
            story.append(Spacer(1, 4))
            i += 1
            continue

        # 无序列表
        if line.strip().startswith(("- ", "* ", "+ ")):
            indent_level = (len(line) - len(line.lstrip())) // 2
            text = line.strip()[2:]
            
            left_indent = 20 + indent_level * 15
            bullet_indent = left_indent - 12  # bullet 在文字左侧 12pt 处

            style = ParagraphStyle(
                f"BulletIndented_{indent_level}",  # 注意：不同层级用不同name，避免样式缓存冲突
                parent=styles["bullet"],
                leftIndent=left_indent,
                bulletIndent=bullet_indent,
                bulletText="•",
            )
            story.append(Paragraph(_md_inline_to_reportlab(text), style))
            i += 1
            continue

        # 有序列表
        ordered_match = re.match(r"^(\d+)\.\s+(.*)$", line.strip())
        if ordered_match:
            list_num = int(ordered_match.group(1))
            list_text = ordered_match.group(2)
            indent_level = (len(line) - len(line.lstrip())) // 2
            num_text = f"{list_num}. {_md_inline_to_reportlab(list_text)}"
            style = ParagraphStyle(
                "OrderedIndented", parent=styles["bullet"],
                leftIndent=20 + indent_level * 15,
            )
            story.append(Paragraph(num_text, style))
            i += 1
            continue

        # 表格
        if "|" in line and i + 1 < len(lines) and re.match(
            r"^[\|\s\-:]+$", lines[i + 1].strip()
        ):
            table_data, rows_consumed = parse_table(lines, i)
            if table_data:
                table = Table(table_data)
                table.setStyle(_build_table_style())
                story.append(table)
                story.append(Spacer(1, 8))
            i += rows_consumed
            continue

        # 代码块
        if line.strip().startswith("```"):
            code_lines = []
            i += 1
            while i < len(lines) and not lines[i].strip().startswith("```"):
                code_lines.append(lines[i])
                i += 1
            if i < len(lines):
                i += 1  # 跳过结束的 ```
            # 用 <br/> 换行，&nbsp; 保留缩进，避免 Preformatted + 中文字体导致 \n 显示为 n
            code_html = "<br/>".join(
                _xml_escape_preserve_spaces(code_line) for code_line in code_lines
            )
            code_style = ParagraphStyle(
                "CodeBlock", parent=styles["code"],
                fontName="Courier", fontSize=10, leading=14,
            )
            story.append(Paragraph(code_html, code_style))
            continue

        # 引用块
        if line.strip().startswith(">"):
            quote_parts = []
            while i < len(lines) and lines[i].strip().startswith(">"):
                quote_text = lines[i].strip()[1:].strip()
                if quote_text:
                    quote_parts.append(_md_inline_to_reportlab(quote_text))
                i += 1
            quote_text = "<br/>".join(quote_parts)
            # 用 Table 实现左侧竖线效果
            inner = Paragraph(quote_text, styles["quote"])
            quote_table = Table(
                [[inner]],
                colWidths=[doc.width - 24],
            )
            quote_table.setStyle(TableStyle([
                ("FONTNAME", (0, 0), (-1, -1), "SimHei"),
                ("LINEBEFORE", (0, 0), (0, -1), 4, HexColor("#DDDDDD")),
                ("LEFTPADDING", (0, 0), (-1, -1), 15),
                ("TOPPADDING", (0, 0), (-1, -1), 5),
                ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
                ("RIGHTPADDING", (0, 0), (-1, -1), 5),
            ]))
            story.append(quote_table)
            story.append(Spacer(1, 4))
            continue

        # 分隔线
        if line.strip() in ("---", "***", "___"):
            story.append(HRFlowable(width="100%", thickness=1, color=HexColor("#CCCCCC")))
            story.append(Spacer(1, 8))
            i += 1
            continue

        # 普通段落
        story.append(Paragraph(_md_inline_to_reportlab(line), styles["normal"]))
        story.append(Spacer(1, 4))
        i += 1

    doc.build(story)
    return pdf_buffer.getvalue()
