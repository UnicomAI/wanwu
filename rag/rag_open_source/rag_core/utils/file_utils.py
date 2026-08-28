import copy
import os
import urllib3
import logging
from dataclasses import dataclass, field
from typing import List

from chains.pdf_loader import PDFLoader
from chains.excel_loader import ExcelLoader
from chains.doc_loader import DOCXLoader
from chains.local_doc_qa import load_file, ChineseTextSplitter

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

from utils.constant import MAX_SENTENCE_SIZE, MIN_SENTENCE_SIZE, MIN_EMBEDDING_SIZE
from utils import model_parser_utils
from utils import asr_utils
from utils import keyframe_extract

logger = logging.getLogger(__name__)

@dataclass
class SplitConfig:
    """文档分割配置"""
    sentence_size: int
    overlap_size: int
    chunk_type: str
    separators: List[str]
    parser_choices: List[str]
    ocr_model_id: str
    asr_model_id: str
    multimodal_model_id: str
    split_type: str = "common"
    child_chunk_config: dict = field(default_factory=dict)


def split_parent_child_chunks(filepath: str,config: SplitConfig):
    docs = []
    chunks = []
    sub_chunks = []
    file_name = os.path.split(filepath)[-1]
    if isinstance(filepath, str):
        if not os.path.exists(filepath):
            logger.error(f"filepath:{filepath},路径不存在")
            return chunks, sub_chunks
        elif os.path.isfile(filepath):
            file = os.path.split(filepath)[-1]
            logger.info(f"切分类型是: {config.chunk_type}")
            docs = load_file(filepath,
                             config.separators,
                             config.sentence_size,
                             config.chunk_type,
                             config.overlap_size,
                             config.parser_choices,
                             config.ocr_model_id,
                             config.asr_model_id,
                             config.multimodal_model_id
                             )

    child_sentence_size = max(int(config.child_chunk_config["chunk_size"]), MIN_SENTENCE_SIZE)
    child_separators=config.child_chunk_config["separators"]
    child_textsplitter = ChineseTextSplitter(
        chunk_type = "split_by_design",
        pdf=False,
        sentence_size=child_sentence_size,
        separators=child_separators
    )

    for p_index, document in enumerate(docs):
        if len(document.page_content) > 0:
            text = document.page_content[:MAX_SENTENCE_SIZE]
            child_list = child_textsplitter.split_text(text)
            if len(child_list) == 0:
                continue
            for index, child_text in enumerate(child_list):
                if len(child_text.strip()) == 0:
                    continue
                meta_data = {
                    "file_name": file_name,
                    "child_chunk_current_num": index+1,
                    "child_chunk_total_num": len(child_list),
                    "chunk_current_num": p_index+1,
                    "chunk_total_num": len(docs)
                }
                sub_chunks.append({
                    'content': text,
                    'embedding_content': child_text,
                    'meta_data': meta_data,
                    "is_parent": False
                })

                chunks.append({
                    "text": child_text,
                    "parent_text": text,
                    'meta_data': meta_data,
                    "is_parent": False
                })

    return chunks, sub_chunks


def split_child_chunks(chunks: list, child_chunk_config: dict):
    new_chunks = []
    sub_chunks = []

    child_sentence_size = max(int(child_chunk_config["chunk_size"]), MIN_SENTENCE_SIZE)
    child_separators=child_chunk_config["separators"]
    child_textsplitter = ChineseTextSplitter(
        chunk_type = "split_by_design",
        pdf=False,
        sentence_size=child_sentence_size,
        separators=child_separators
    )

    for chunk in chunks:
        text = chunk["text"][:MAX_SENTENCE_SIZE]
        if len(text) > 0:
            child_list = child_textsplitter.split_text(text)
            for index, child_text in enumerate(child_list):
                if len(child_text.strip()) > 0:
                    meta_data = copy.deepcopy(chunk["meta_data"])
                    meta_data["child_chunk_current_num"] = index
                    meta_data["child_chunk_total_num"] = len(child_list)

                    sub_chunks.append({
                        'content': text,
                        'embedding_content': child_text,
                        'meta_data': meta_data,
                        "is_parent": False
                    })

                    new_chunks.append({
                        "text": child_text,
                        "parent_text": text,
                        'meta_data': meta_data,
                        "is_parent": False
                    })

    return new_chunks, sub_chunks


def split_chunks(filepath: str,config: SplitConfig):
    #单个文件进行切分成块
    docs = []
    file_name = os.path.split(filepath)[-1]
    if isinstance(filepath, str):
        if not os.path.exists(filepath):
            logger.error(f"filepath:{filepath},路径不存在")
            return False, docs, ''
        elif os.path.isfile(filepath):
            file = os.path.split(filepath)[-1]
            logger.info(f"切分类型是: {config.chunk_type}")
            docs = load_file(filepath,
                             config.separators,
                             config.sentence_size,
                             config.chunk_type,
                             config.overlap_size,
                             config.parser_choices,
                             config.ocr_model_id,
                             config.asr_model_id,
                             config.multimodal_model_id)
            if len(docs) > 0:
                chunks = []
                for document in docs:
                    if len(document.page_content) > 0:
                        doc_dict = {
                            "text": document.page_content[:MAX_SENTENCE_SIZE],
                            "meta_data": {"file_name": file_name}
                        }
                        chunks.append(doc_dict)
                        # chunks.append(document.page_content)
                if len(chunks)>0:
                    return True,chunks,file
                else:
                    return False,docs,''
            else:
                return False,docs,''
        else:
            return False,docs,''
    return False,docs,''


def split_chunks_for_parser(filepath: str, config: SplitConfig):
    docs = []
    file_name = os.path.split(filepath)[-1]
    if isinstance(filepath, str):
        if not os.path.exists(filepath):
            logger.info("路径不存在")
            return False, docs, ''
        elif os.path.isfile(filepath):
            file = os.path.split(filepath)[-1]
            logger.info(f"切分类型是: {config.chunk_type}, overlap_size: {config.overlap_size}")
            docs = load_file(filepath,
                             config.separators,
                             config.sentence_size,
                             config.chunk_type,
                             config.overlap_size,
                             config.parser_choices,
                             config.ocr_model_id)
            if len(docs) > 0:
                chunks = []
                for document in docs:
                    doc_dict = {"text": document.page_content, "metadata": {"file_name": file_name}}
                    chunks.append(doc_dict)
                if len(chunks) > 0:
                    return True, chunks, file
                else:
                    return False, docs, ''
            else:
                return False, docs, ''
        else:
            return False, docs, ''
    return False, docs, ''


def split_doc(chunks, sentence_size):
    # overlap_size = 0
    textsplitter = ChineseTextSplitter(chunk_type =1, pdf=False, sentence_size=sentence_size)
    sub_doc = []
    layer = 2

    for i in chunks:
        text = i["text"][:MAX_SENTENCE_SIZE]

        if len(i["text"].strip()) > 0:
            chunk_dict = {
                'content': text,
                'embedding_content': text,
                'meta_data': copy.deepcopy(i['meta_data'])
            }
            # 如果labels存在，则添加到chunk_dict中
            if "labels" in i and i["labels"]:
                chunk_dict["labels"] = copy.deepcopy(i["labels"])
            sub_doc.append(chunk_dict)

    while True:
        if sentence_size < MIN_SENTENCE_SIZE: break
        if layer == 0: break
        sentence_size = int(sentence_size/2)
        layer=layer-1
        textsplitter.sentence_size=sentence_size
        for i in chunks:
            text = i["text"][:MAX_SENTENCE_SIZE]

            sub_list = textsplitter.split_text1(text)
            if len(sub_list) == 1:continue
            for short_text in sub_list:
                if len(short_text.strip()) > 0:
                    chunk_dict = {
                        'content': text,
                        'embedding_content': short_text,
                        'meta_data': copy.deepcopy(i['meta_data'])
                    }
                    # 如果labels存在，则添加到chunk_dict中
                    if "labels" in i and i["labels"]:
                        chunk_dict["labels"] = copy.deepcopy(i["labels"])
                    sub_doc.append(chunk_dict)
    return sub_doc


def split_short_doc(file_path, download_link, chunks, sentence_size):
    textsplitter = ChineseTextSplitter(chunk_type=1, pdf=False, sentence_size=sentence_size)
    sub_doc=[]
    file_name = os.path.split(file_path)[-1]


    layer=2

    for chunk in chunks:
        text = chunk["text"][:MAX_SENTENCE_SIZE]

        chunk_type = chunk["type"]

        embedding_chunks = chunk.get("embedding_chunks", [])
        if len(text.strip()) > 0:
            # 对于text和table类型至少保证一条进入subchunk
            chunk_dict = {
                "content": text,
                "embedding_content": text,
                "meta_data": copy.deepcopy(chunk['meta_data'])
            }
            sub_doc.append(chunk_dict)
            # if chunk_type == "table":
            #     # 仅对pdf不走自定义例如走ocr解析的表格再按换行做一次切分(注意：此种切分方式会引入embedding_content为空的情况)
            #     if len(embedding_chunks) == 0 and file_name.lower().endswith(".pdf") and text.count("\n") >= 2:
            #         embedding_chunks = text.split("\n")
            # 若有定制化构建的embedding_chunks则将此作为语义索引
            for item in embedding_chunks:
                embedding_content = item
                if len(embedding_content.strip()) > MIN_EMBEDDING_SIZE and len(embedding_content) < MAX_SENTENCE_SIZE:
                    embed_chunk_dict = {
                        "content": text,
                        "embedding_content": embedding_content,
                        "meta_data": copy.deepcopy(chunk['meta_data'])
                    }
                    sub_doc.append(embed_chunk_dict)
    while True:
        if sentence_size < MIN_SENTENCE_SIZE:break
        if layer==0:break
        sentence_size = int(sentence_size/2)
        layer=layer-1
        textsplitter.sentence_size=sentence_size
        for chunk in chunks:
            text = chunk["text"]
            chunk_type = chunk["type"]

            if chunk_type == "text":
                sub_list = textsplitter.split_text1(text)
                if len(sub_list) == 1:continue
                for short_text in sub_list:
                    if len(short_text.strip()) > MIN_EMBEDDING_SIZE and len(short_text) < MAX_SENTENCE_SIZE and len(chunk["text"]) < MAX_SENTENCE_SIZE:
                        chunk_dict = {
                            "content": text,
                            "embedding_content": short_text,
                            "meta_data": copy.deepcopy(chunk['meta_data'])
                        }
                        sub_doc.append(chunk_dict)

    return sub_doc


def split_file_adapter(add_file_path: str, download_link: str, config: SplitConfig):
    """
    自动解析与切分适配器
    :param add_file_path:文件类型切分适配
    :param download_link:minio下载链接
    :param config: split配置, sentence_size(chunk切分长度阈值),parser_choices(解析方式),ocr_model_id(ocr模型id）

    """
    status = False
    sub_chunk = []
    chunks = []
    file_name = os.path.split(add_file_path)[-1]

    if file_name.lower().endswith(".pdf"):
        if "model" in config.parser_choices:
            logger.info("-----文档解析模型(model)处理PDF-------")
            chunks = model_parser_utils.model_parser(add_file_path, config.ocr_model_id)
        else:
            # text 模式：传统解析，不再根据 chunk_type 自动分发
            logger.info("-----文字提取(text)处理PDF-------")
            loader = PDFLoader(file_path=add_file_path, parser_choices=config.parser_choices,
                               ocr_model_id=config.ocr_model_id, autodetect_encoding=True)
            height_list = loader.get_height_list() if hasattr(loader, 'get_height_list') else loader.get_chunk_type()[1]
            if height_list and len(height_list) >= 2:
                # 含表格/标题的 PDF 走结构化提取
                chunks = loader.load_and_split_doc(height_list)
            else:
                # 普通 PDF 走基础 load()
                docs = loader.load()
                chunks = [{"type": "text", "text": doc.page_content, "page_num": [],
                           "length": len(doc.page_content)} for doc in docs]

    elif file_name.lower().endswith(".xlsx"):
        logger.info("-----定制化xlsx解析+切分：支持多sheet页-------")
        loader = ExcelLoader(add_file_path, autodetect_encoding=True)
        chunks = loader.load_and_split_doc()
    elif file_name.lower().endswith(".xls"):
        logger.info("-----定制化xls解析+切分：支持多sheet页-------")
        loader = ExcelLoader(add_file_path, autodetect_encoding=True)
        chunks = loader.load_and_split_xls()
    elif file_name.lower().endswith(".docx"):
        logger.info("-----定制化word解析+切分：支持复杂表格-------")
        loader = DOCXLoader(add_file_path, autodetect_encoding=True)
        chunks = loader.custom_load_and_split_doc()
    elif file_name.lower().endswith((".wav", ".mp3", ".aac", ".m4a", ".mp4",  ".mov", ".avi", ".wmv")):
        if "asr" in config.parser_choices:
            logger.info("-----执行音视频ASR解析转文本-------")
            chunks = asr_utils.asr_parser_chunk(add_file_path, config.asr_model_id)
        if file_name.lower().endswith((".mp4", ".mov", ".avi", ".wmv")):
            logger.info("-----执行视频画面关键帧解析转文本-------")
            keyframe_chunks = keyframe_extract.exact(add_file_path, config.parser_choices, config.multimodal_model_id)
            if keyframe_chunks:
                # 如果视频中有画面帧择提取画面帧chunk追加
                chunks.extend(keyframe_chunks)

    status = True if len(chunks) > 0 else False
    if status:
        new_chunks = []
        for chunk in chunks:
            text = chunk["text"][:MAX_SENTENCE_SIZE]
            row_num = 0
            sheet_name = ""
            if len(text) > 0:
                chunk_type = chunk.get('type', 'text')
                page_num = chunk.get('page_num', [])
                parent_title = chunk.get('parent_title', [])
                embedding_chunks = chunk.get('embedding_chunks', [])

                if "meta_data" in chunk:
                    chunk_download_link = chunk["meta_data"]["download_link"] if "download_link" in chunk["meta_data"] else ""
                    if chunk_download_link == "":
                        chunk_download_link = download_link
                    chunk_file_name = chunk["meta_data"]["file_name"] if "file_name" in chunk["meta_data"] else file_name
                    row_num = chunk["meta_data"]["row_num"] if "row_num" in chunk["meta_data"] else 0
                    sheet_name = chunk["meta_data"]["sheet_name"] if "sheet_name" in chunk["meta_data"] else ""
                else:
                    chunk_download_link = download_link
                    chunk_file_name = file_name

                doc_dict = {
                    "type": chunk_type,
                    "text": text,
                    "embedding_chunks": embedding_chunks,
                    "meta_data": {
                        "page_num": page_num,
                        "parent_title": parent_title,
                        "file_name": chunk_file_name,
                        "download_link": chunk_download_link,
                        "row_num": row_num,
                        "sheet_name": sheet_name,
                    }
                }
                new_chunks.append(doc_dict)
        chunks = new_chunks
        init_chunk_num(chunks)
        sub_chunk = split_short_doc(add_file_path, download_link, chunks, config.sentence_size)
    return status, sub_chunk, chunks


def init_chunk_num(chunks):
    for index, item in enumerate(chunks):
        if "chunk_current_num" not in item["meta_data"]:
            item["meta_data"]["chunk_current_num"] = index + 1
            item["meta_data"]["chunk_total_num"] = len(chunks)

def init_chunk_role(chunks: list, is_parent: bool):
    for item in chunks:
        item["is_parent"] = is_parent

def split_text_file(add_file_path: str, download_link: str, config: SplitConfig):
    chunks = []
    sub_chunks = []
    status = False

    if config.split_type == "parent_child":
        if config.chunk_type == 'split_by_design':
            chunks, sub_chunks = split_parent_child_chunks(add_file_path, config)
    else:
        if config.chunk_type == 'split_by_design':
            # 分段方式：自定义分段
            status, chunks, filename = split_chunks(add_file_path, config)
            init_chunk_num(chunks)
            sub_chunks = split_doc(chunks, config.sentence_size)
        else:
            # 分段方式：自动分段
            status, sub_chunks, chunks = split_file_adapter(add_file_path, download_link, config)
            if not status:
                status, chunks, filename = split_chunks(add_file_path, config)
                init_chunk_num(chunks)
                sub_chunks = split_doc(chunks, config.sentence_size)

    return sub_chunks, chunks

