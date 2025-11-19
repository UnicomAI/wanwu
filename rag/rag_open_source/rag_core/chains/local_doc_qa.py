import os
# from langchain.embeddings.huggingface import HuggingFaceEmbeddings
from chains.excel_loader import ExcelLoader
from chains.pdf_loader import PDFLoader
from chains.pptx_loader import PPTXLoader
from chains.image_loader import ImageLoader
from chains.doc_loader import DOCLoader, DOCXLoader
from chains.html_loader import HTMLLoader
from chains.markdown_loader import MarkdownLoader
from langchain_community.document_loaders import UnstructuredFileLoader, TextLoader, CSVLoader
from utils import model_parser_utils
from utils.constant import SENTENCE_SIZE
from utils.prompts import PROMPT_TEMPLATE

from textsplitter import ChineseTextSplitter
from typing import List

from langchain.docstore.document import Document

import chardet
import xml.etree.ElementTree as ET
from logging_config import setup_logging
logger_name = 'rag_local_doc_qa'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))

def torch_gc():
    pass

def extract_text_from_xml(input_xml_path):
    """
    从输入的 XML 文件中提取文本内容，并将其写入同名的 TXT 文件中。

    参数:
    - input_xml_path: str，输入的 XML 文件路径
    """
    # 获取输入文件的目录和文件名（不含扩展名） [EN] Get the directory and file name of the input file (without extension)
    base_name = os.path.splitext(input_xml_path)[0]
    # 生成输出的 TXT 文件路径 [EN] Generate output TXT file path
    output_txt_path = base_name + '.txt'

    # 解析 XML 文件 [EN] Parse XML files
    tree = ET.parse(input_xml_path)
    root = tree.getroot()

    # 存储提取的文本 [EN] Store extracted text
    text_contents = set()

    # 递归地遍历 XML 树，并提取文本内容 [EN] Recursively traverse the XML tree and extract the text content
    def recurse_extract_text(element):
        # 收集当前元素的文本内容（如果有） [EN] Collects the text content of the current element (if any)
        if element.text and element.text.strip():
            text_contents.add(element.text.strip())
        # 递归地处理子元素 [EN] Process child elements recursively
        for child in element:
            recurse_extract_text(child)
    
    # 开始递归遍历 [EN] Start recursive traversal
    recurse_extract_text(root)

    # 将文本内容写入 TXT 文件 [EN] Write text content to TXT file
    with open(output_txt_path, 'w', encoding='utf-8') as f:
        for text in sorted(text_contents):
            f.write(text + '\n')
    return output_txt_path




# patch HuggingFaceEmbeddings to make it hashable
# def _embeddings_hash(self):
#     return hash(self.model_name)
#
#
# HuggingFaceEmbeddings.__hash__ = _embeddings_hash




# def tree(filepath, ignore_dir_names=None, ignore_file_names=None):
#     """返回两个列表，第一个列表为 filepath 下全部文件的完整路径, 第二个为对应的文件名""" [EN] """Returns two lists. The first list is the full path of all files under filepath, and the second is the corresponding file name"""
#     if ignore_dir_names is None:
#         ignore_dir_names = []
#     if ignore_file_names is None:
#         ignore_file_names = []
#     ret_list = []
#     if isinstance(filepath, str):
#         if not os.path.exists(filepath):
#             print("路径不存在") [EN] print("Path does not exist")
#             return None, None
#         elif os.path.isfile(filepath) and os.path.basename(filepath) not in ignore_file_names:
#             return [filepath], [os.path.basename(filepath)]
#         elif os.path.isdir(filepath) and os.path.basename(filepath) not in ignore_dir_names:
#             for file in os.listdir(filepath):
#                 fullfilepath = os.path.join(filepath, file)
#                 if os.path.isfile(fullfilepath) and os.path.basename(fullfilepath) not in ignore_file_names:
#                     ret_list.append(fullfilepath)
#                 if os.path.isdir(fullfilepath) and os.path.basename(fullfilepath) not in ignore_dir_names:
#                     ret_list.extend(tree(fullfilepath, ignore_dir_names, ignore_file_names)[0])
#     return ret_list, [os.path.basename(p) for p in ret_list]


def load_file(filepath,separators=['。'],sentence_size=SENTENCE_SIZE, chunk_type='split_by_default',overlap_size=0.2,parser_choices=["text"],ocr_model_id = ""):
    if filepath.lower().endswith(".md"):
        # loader = UnstructuredFileLoader(filepath, mode="elements")
        # docs = loader.load()
        encoding = detect_file_encoding(filepath)
        loader = MarkdownLoader(filepath, encoding=encoding, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type=chunk_type, sentence_size=sentence_size,
                                           overlap_size=overlap_size, separators=separators)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".txt"):
        encoding = detect_file_encoding(filepath)
        loader = TextLoader(filepath, encoding=encoding, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type=chunk_type, sentence_size=sentence_size,overlap_size=overlap_size,separators=separators)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".pdf"):
        if "model" in parser_choices:
            markdown_file_path = model_parser_utils.model_parser_file(filepath,ocr_model_id)
            encoding = detect_file_encoding(markdown_file_path)
            loader = TextLoader(markdown_file_path, encoding=encoding, autodetect_encoding=True)
            textsplitter = ChineseTextSplitter(chunk_type=chunk_type, sentence_size=sentence_size,
                                               overlap_size=overlap_size, separators=separators)
            docs = loader.load_and_split(textsplitter)
        else:
            loader = PDFLoader(file_path=filepath,parser_choices=parser_choices, ocr_model_id = ocr_model_id)
            # print("=========>load_file,chunk_type=%s" % str(loader.get_chunk_type()))
            textsplitter = ChineseTextSplitter(chunk_type=chunk_type, pdf=True, sentence_size=sentence_size,overlap_size=overlap_size, separators=separators)
            docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith((".jpg", ".jpeg", ".png")):
        # loader = UnstructuredPaddleImageLoader(filepath, mode="elements")
        loader = ImageLoader(file_path=filepath, ocr_model_id = ocr_model_id)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size)
        docs = loader.load_and_split(text_splitter=textsplitter)
    elif filepath.lower().endswith(".csv"):
        encoding = detect_file_encoding(filepath)
        logger.info("csv detect encoding is:%s" % encoding)
        loader = CSVLoader(filepath, encoding=encoding, autodetect_encoding=True)
        docs = loader.load()
    elif filepath.lower().endswith(".xlsx"):
        # filepath = trans_excel(filepath)
        # print(filepath)
        encoding = detect_file_encoding(filepath)
        logger.info("xlsx detect encoding is:%s" % encoding)
        loader = ExcelLoader(filepath, encoding=encoding, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, excel=True, sentence_size=sentence_size)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".xml"):
        filepath = extract_text_from_xml(filepath)
        loader = TextLoader(filepath, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size)
    elif filepath.lower().endswith(".doc"):
        loader = DOCLoader(filepath, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size,overlap_size=overlap_size, separators=separators)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".docx"):
        loader = DOCXLoader(filepath, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size,overlap_size=overlap_size, separators=separators)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".html"):
        encoding = detect_file_encoding(filepath)
        loader = HTMLLoader(filepath, encoding=encoding, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size, overlap_size=overlap_size,
                                           separators=separators)
        docs = loader.load_and_split(textsplitter)
    elif filepath.lower().endswith(".pptx"):
        loader = PPTXLoader(filepath, autodetect_encoding=True)
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size, overlap_size=overlap_size, separators=separators)
        docs = loader.load_and_split(textsplitter)
    else:
        loader = UnstructuredFileLoader(filepath, mode="elements")
        textsplitter = ChineseTextSplitter(chunk_type, sentence_size=sentence_size, overlap_size=overlap_size, separators=separators)
        docs = loader.load_and_split(text_splitter=textsplitter)
    write_check_file(filepath, docs)
    return docs

def write_check_file(filepath, docs):
    folder_path = os.path.join(os.path.dirname(filepath), "tmp_files")
    if not os.path.exists(folder_path):
        os.makedirs(folder_path)
    fp = os.path.join(folder_path, 'load_file.txt')
    with open(fp, 'a+', encoding='utf-8') as fout:
        fout.write("filepath=%s,len=%s" % (filepath, len(docs)))
        fout.write('\n')
        for i in docs:
            fout.write(str(i))
            fout.write('\n')
        fout.close()


def search_result2docs(search_results):
    docs = []
    for result in search_results:
        doc = Document(page_content=result["snippet"] if "snippet" in result.keys() else "",
                       metadata={"source": result["link"] if "link" in result.keys() else "",
                                 "filename": result["title"] if "title" in result.keys() else ""})
        docs.append(doc)
    return docs

            
def detect_file_encoding(file_path):
    with open(file_path, 'rb') as f:
        raw_data = f.read()
    encoding = chardet.detect(raw_data)['encoding']
    if encoding is not None:
        if encoding.lower().startswith('gb'):
            encoding = 'gb18030'
    
    return encoding        
