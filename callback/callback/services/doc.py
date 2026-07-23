import concurrent.futures
import io
import json
import os
import posixpath

import requests

from callback.services import minio as minio_service
from callback.services.doc_parser import download_to_tempfile, parse_file
from callback.services.doc_parser.factory import ParserFactory
from callback.utils.doc_converter import markdown_to_html, markdown_to_docx, markdown_to_pdf
from configs.config import config
from extensions.minio import minio_client
from utils.build_prompt import build_docqa_prompt_from_search_list
from utils.log import logger
from utils.response import BizError


def process_documents(query, file_urls):
    """
    解析文档并生成 Prompt
    """
    if not file_urls:
        raise BizError("No file URLs provided.")

    # 统一处理为列表
    file_urls = [file_urls] if isinstance(file_urls, str) else file_urls
    all_docs = []

    with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
        future_to_url = {executor.submit(parse_doc, url): url for url in file_urls}

        for future in concurrent.futures.as_completed(future_to_url):
            url = future_to_url[future]
            try:
                docs = future.result()
                all_docs.extend(docs)
            except Exception as e:
                # 这里可以记录日志
                logger.error(f"解析文档失败 {url}: {str(e)}")

    if not all_docs:
        raise BizError("No document content parsed.")

    # 构建文档列表
    doc_list = [
        {
            "snippet": doc.get("text"),
            "file_name": doc.get("metadata", {}).get("file_name"),
        }
        for doc in all_docs
    ]

    # 构建提示词
    prompt = build_docqa_prompt_from_search_list(query, doc_list)
    return prompt


def generate_file_to_minio(formatted_markdown, filename, to_format="txt"):

    with io.BytesIO() as file_buffer:
        # 1. 初始化变量
        full_filename = filename + ".txt"

        # 2. 根据格式生成文件内容
        if to_format == "pdf":
            full_filename = filename + ".pdf"
            pdf_bytes = markdown_to_pdf(formatted_markdown)
            file_buffer.write(pdf_bytes)

        elif to_format == "docx":
            full_filename = filename + ".docx"
            doc = markdown_to_docx(formatted_markdown)
            doc.save(file_buffer)

        elif to_format == "md":
            full_filename = filename + ".md"
            file_buffer.write(formatted_markdown.encode("utf-8"))

        elif to_format == "html":
            full_filename = filename + ".html"
            html_content = markdown_to_html(formatted_markdown)
            file_buffer.write(html_content.encode("utf-8"))

        elif to_format == "txt":
            full_filename = filename + ".txt"
            file_buffer.write(formatted_markdown.encode("utf-8"))

        # 3. 上传逻辑
        object_path = minio_service.upload_file_to_minio(file_buffer, full_filename)

        download_link = posixpath.join(
            config.callback_cfg["URL"]["MINIO_DOWNLOAD"], object_path
        )

        # 4. 返回结果
        return full_filename, object_path, download_link


def parse_doc(file_url):
    """
    解析单个文档

    本地工厂支持的格式直接解析；不支持的格式(如 .wps/.ofd)回退到 rag 的
    /rag/doc_parser 接口，并将返回的 chunks 拼接为整篇全文。

    参数:
    file_url (str): 文档URL

    返回:
    list: 解析后的文档片段列表，每项 {"text": ..., "metadata": {"file_name": ...}}
    """
    file_name = _url_filename(file_url)

    if not _rag_only_enabled() and ParserFactory.is_supported(file_name):
        file_path, _ = download_to_tempfile(file_url)
        try:
            text = parse_file(file_path, max_tokens=0)
        finally:
            _safe_unlink(file_path)
    else:
        if _rag_only_enabled():
            logger.info(f"RAG_ONLY 已开启，强制走 rag 解析: {file_url}")
        else:
            logger.info(f"本地不支持 {file_name}，回退 rag 解析: {file_url}")
        text = _parse_via_rag(file_url, max_tokens=0)

    if not text:
        raise BizError("No document content parsed.")

    return [{"text": text, "metadata": {"file_name": file_name}}]


def parse_doc_only(file_url, max_token):
    """
    解析单个文档，不进行切分，按 max_token 截断返回完整文本

    本地工厂支持的格式直接解析；不支持的格式回退到 rag 接口并拼接全文后截断。

    参数:
    file_url (str): 文件URL
    max_token (int): 返回文本的最大token数；<=0 时使用默认上限

    返回:
    str: 解析后(并按需截断)的文档内容
    """
    limit_max_token = int(config.callback_cfg["DOC"]["DEFAULT_LIMIT_MAX_TOKEN"])
    if max_token <= 0:
        max_token = limit_max_token

    file_name = _url_filename(file_url)

    if not _rag_only_enabled() and ParserFactory.is_supported(file_name):
        file_path, _ = download_to_tempfile(file_url)
        try:
            text = parse_file(file_path, max_tokens=max_token)
        finally:
            _safe_unlink(file_path)
    else:
        if _rag_only_enabled():
            logger.info(f"RAG_ONLY 已开启，强制走 rag 解析: {file_url}")
        else:
            logger.info(f"本地不支持 {file_name}，回退 rag 解析: {file_url}")
        text = _parse_via_rag(file_url, max_tokens=max_token)

    if not text:
        raise BizError("No document content parsed.")

    return text


def _parse_via_rag(file_url, max_tokens=0):
    """回退到 rag 的 /rag/doc_parser 接口解析文档。

    rag 接口为切分语义，这里传 separators=[]、较大的 sentence_size 以尽量减少切分，
    再把返回的 chunks 拼接为整篇全文，与本地解析器的"只读文本"语义对齐。
    最后按 max_tokens 截断。verify=False 与原 parse_doc 行为一致，适配内网自签证书。
    """
    from callback.services.doc_parser.base import DocumentParser

    url = config.callback_cfg["URL"]["RAG_DOC_PARSER"]
    if not url:
        raise BizError("本地不支持该文件类型，且未配置 RAG_DOC_PARSER 兜底地址")

    payload = json.dumps(
        {
            "url": file_url,
            # sentence_size 给一个较大值，配合 separators=[] 让 rag 尽量不切分
            "sentence_size": 100000,
            "overlap_size": 0,
            "separators": [],
        }
    )
    headers = {"Content-Type": "application/json;charset=utf-8"}
    response = requests.post(url, headers=headers, data=payload, verify=False, timeout=300)
    docs = response.json().get("docs", [])

    full_text = "\n".join(doc.get("text", "") for doc in docs)
    return DocumentParser.truncate(full_text, max_tokens)


def _rag_only_enabled() -> bool:
    """是否强制只走 rag 兜底解析。

    读 ``[DOC] RAG_ONLY``：``1``/``true`` 表示强制只走 rag（本地支持格式也跳过本地解析），
    ``0``/未配置走正常逻辑（本地支持走本地，不支持回退 rag）。
    """
    raw = (config.callback_cfg["DOC"].get("RAG_ONLY", "") or "").strip().lower()
    return raw in ("1", "true")


def _url_filename(file_url):
    """从 URL 中提取解码后的文件名（含后缀），用于后缀预判。"""
    import urllib.parse

    parsed = urllib.parse.urlparse(file_url)
    return urllib.parse.unquote(parsed.path.split("/")[-1]) or "download"


def _safe_unlink(path: str) -> None:
    """删除临时文件，忽略不存在等错误。"""
    try:
        os.unlink(path)
    except OSError:
        pass
