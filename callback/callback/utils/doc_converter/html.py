"""Markdown 转 HTML"""

import markdown

from callback.utils.doc_converter.common import MD_EXTENSIONS, HTML_STYLE


def markdown_to_html(md_content: str) -> str:
    """将 Markdown 转换为带样式的 HTML（用于 .html 文件导出）"""
    html_content = markdown.markdown(
        md_content, extensions=MD_EXTENSIONS, output_format="html5"
    )

    return f"""<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
{HTML_STYLE}
    </style>
</head>
<body>
{html_content}
</body>
</html>"""
