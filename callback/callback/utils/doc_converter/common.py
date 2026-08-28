"""Markdown 转换器共享常量和工具函数"""

import re

# Markdown 解析扩展
MD_EXTENSIONS = [
    "tables",
    "fenced_code",
    "codehilite",
    "toc",
    "nl2br",
    "sane_lists",
]

# HTML 导出样式
HTML_STYLE = """
    body {
        font-family: "Microsoft YaHei", "SimSun", "SimHei", Arial, sans-serif;
        font-size: 12pt;
        line-height: 1.6;
        max-width: 800px;
        margin: 20px auto;
        padding: 20px;
    }
    h1 { font-size: 22pt; font-weight: bold; color: #333; margin-top: 20px; }
    h2 { font-size: 18pt; font-weight: bold; color: #333; margin-top: 18px; }
    h3 { font-size: 14pt; font-weight: bold; color: #333; margin-top: 16px; }
    h4 { font-size: 12pt; font-weight: bold; color: #333; margin-top: 14px; }
    p { margin: 10px 0; text-align: justify; }
    ul, ol { margin: 10px 0; padding-left: 30px; }
    li { margin: 5px 0; }
    table {
        border-collapse: collapse;
        width: 100%;
        margin: 15px 0;
    }
    th, td {
        border: 1px solid #666;
        padding: 8px 12px;
        text-align: left;
    }
    th {
        background-color: #f0f0f0;
        font-weight: bold;
    }
    code {
        background-color: #f5f5f5;
        padding: 2px 6px;
        border-radius: 3px;
        font-family: Consolas, "Courier New", monospace;
        font-size: 10pt;
    }
    pre {
        background-color: #f5f5f5;
        padding: 12px;
        border-radius: 5px;
        overflow-x: auto;
        margin: 15px 0;
    }
    pre code {
        background-color: transparent;
        padding: 0;
    }
    blockquote {
        border-left: 4px solid #ddd;
        margin: 15px 0;
        padding-left: 15px;
        color: #666;
    }
    a { color: #0066cc; text-decoration: underline; }
    hr { border: none; border-top: 1px solid #ccc; margin: 20px 0; }
    strong { font-weight: bold; }
    em { font-style: italic; }
"""


def parse_table(lines, start_idx):
    """
    解析 Markdown 表格，返回 (table_data, rows_consumed)

    供 docx 和 pdf 转换器共用
    """
    table_data = []
    i = start_idx
    rows_consumed = 0

    while i < len(lines) and "|" in lines[i]:
        row = [cell.strip() for cell in lines[i].split("|")[1:-1]]
        # 跳过分隔行 (| --- | --- |)
        if row and not all(re.match(r"^[-:]+$", c) for c in row if c):
            table_data.append(row)
        i += 1
        rows_consumed += 1

    return table_data, rows_consumed
