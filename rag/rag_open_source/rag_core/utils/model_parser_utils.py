import traceback
import requests
import json
import uuid
import os
import html2text
import datetime
import requests
# import urllib3

from datetime import datetime, timedelta
from logging_config import setup_logging
logger_name='rag_model_parser_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))
from utils.constant import MODEL_PARSER_MAX_WORKERS
from model_manager import get_model_configure, OcrModelConfig

hl2txt = html2text.HTML2Text()

from concurrent.futures import ThreadPoolExecutor, as_completed
# from PyPDF2 import PdfReader, PdfWriter
import time
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
from collections import defaultdict
import fitz
from pathlib import Path


def get_page_data(page_num, add_file_path, ocr_model_id):
    """
    获取单页的数据并调用模型解析服务
    :param page_num: 页码
    :param add_file_path: 文件路径
    :return: 模型解析结果
    """
    # file_name = os.path.split(add_file_path)[-1]
    directory = os.path.dirname(add_file_path)
    path_obj = Path(add_file_path)

    file_name = path_obj.stem
    full_file_name = path_obj.name  # 带扩展名的完整文件名（用于formData的fileName） [EN] Full filename with extension (fileName for formData)

    try:
        # 打开PDF文件 [EN] Open PDF file

        pdf_document = fitz.open(add_file_path)

        if page_num > len(pdf_document) or page_num < 1:
            logger.error(f"Page number {page_num} is out of range.")
            return None, page_num

        # 创建一个新的PDF文档并将指定页添加到其中 [EN] Create a new PDF document and add specified pages to it

        page_pdf_path = f"{file_name}_page_{page_num}.pdf"
        # 组合成新的文件路径 [EN] Combined into a new file path
        output_pdf_path = os.path.join(directory, page_pdf_path)
        logger.info("======>model_parser_utils,get_page_data=%s" % output_pdf_path)
        new_pdf = fitz.open()  # 新建一个空的PDF文档 [EN] Create a new empty PDF document
        new_pdf.insert_pdf(pdf_document, from_page=page_num - 1, to_page=page_num - 1)
        new_pdf.save(output_pdf_path)
        new_pdf.close()

        files = {"file": (page_pdf_path, open(output_pdf_path, 'rb'))}

        data = {
            "file_name": page_pdf_path,
            "extract_image": 1
        }

        if ocr_model_id == "":
            logger.error("ocr_model_id为空")
            return None, page_num

        model_config = get_model_configure(ocr_model_id)
        wanwu_ocr_url = ""
        api_key = ""
        if isinstance(model_config, OcrModelConfig):
            wanwu_ocr_url = model_config.endpoint_url + "/pdf-parser"
            api_key = model_config.api_key
        headers = {"Authorization": f"Bearer {api_key}"}

        rate_limit_backoff = [10, 20, 40, 60]  # 限流退避 [EN] Current limiting backoff
        other_error_max_retries = 2  # 其他错误最多重试2次 [EN] Other errors will be retried up to 2 times
        other_error_wait = 0.5  # 每次0.5s [EN] 0.5s each time

        attempt = 0
        while attempt < max(len(rate_limit_backoff), other_error_max_retries) + 1:
            try:
                r = requests.post(wanwu_ocr_url, files=files, headers=headers, data=data, timeout=60)
                ret_json = r.json()
                # logger.info(f"model_parser_utils.get_page_data result: {ret_json}")
                r.raise_for_status()  # 触发HTTP错误状态码的异常 [EN] Exceptions that trigger HTTP error status codes
                if ret_json.get("code") == "200":
                    text = ret_json["content"]
                    # logger.info(f"get_paged_data page:%s, result:%s" % (page_num, text))
                    return text, page_num

                else:
                    logger.error(f"页 {page_num} PDF模型解析失败：{ret_json.get('message', '未知错误')}")
                    return None, page_num
            except requests.exceptions.HTTPError as e:
                error_details = f"HTTPError: {type(e).__name__} - {str(e)}"
                if hasattr(e, 'response'):
                    try:
                        status_code = getattr(e.response, "status_code", "N/A")
                        error_body = e.response.text if hasattr(e.response, "text") else "N/A"
                        error_details += f" | HTTP {status_code}: {error_body}"
                    except Exception as parse_err:
                        error_details += f" | Failed to parse error: {parse_err}"

                logger.error(f"页 {page_num} HTTP错误 (attempt {attempt + 1}): {error_details}")

                # 判断是否限流(429) [EN] Determine whether to limit current (429)
                is_rate_limited = error_details and "429" in error_details
                if is_rate_limited:
                    if attempt < len(rate_limit_backoff):
                        wait_time = rate_limit_backoff[attempt]
                        logger.warning(f"Rate limited (429). Retrying after {wait_time}s...")
                        time.sleep(wait_time)
                        attempt += 1
                        continue
                    else:
                        logger.error("Exceeded max retries due to rate limiting or server error.")
                        break
                else:
                    if attempt < other_error_max_retries:
                        logger.warning(f"Non-429 error. Retrying after {other_error_wait}s...")
                        time.sleep(other_error_wait)
                        attempt += 1
                        continue
                    else:
                        logger.error("Exceeded max retries for non-429/5xx errors.")
                        break

            except requests.exceptions.Timeout:
                error_details = "Timeout Error"
                logger.error(f"页 {page_num} 请求超时 (attempt {attempt + 1}): {error_details}")

                if attempt < other_error_max_retries:
                    logger.warning(f"Timeout error. Retrying after {other_error_wait}s...")
                    time.sleep(other_error_wait)
                    attempt += 1
                    continue
                else:
                    logger.error("Exceeded max retries for timeout errors.")
                    break

            except requests.exceptions.RequestException as e:
                error_details = f"RequestException: {type(e).__name__} - {str(e)}"
                logger.error(f"页 {page_num} 请求异常 (attempt {attempt + 1}): {error_details}")
                break

            except Exception as e:
                error_details = f"Unexpected Error: {type(e).__name__} - {str(e)}"
                logger.error(f"页 {page_num} 未预期错误 (attempt {attempt + 1}): {error_details}")
                break
            finally:
                # 清理临时文件 [EN] Clean temporary files
                if os.path.exists(output_pdf_path):
                    os.remove(output_pdf_path)
                time.sleep(0.1)
    except Exception as e:
        logger.error(f"处理页 {page_num} 失败：{e}")
        logger.error(traceback.format_exc())
        return None, page_num

    # 最终错误处理 [EN] final error handling
    logger.error(f"Failed to process page {page_num} after retries.")
    return None, page_num


def model_parser(add_file_path, ocr_model_id):
    """
    处理整个PDF文档，按页并发调用OCR文档解析工具服务
    :param add_file_path: 文件路径
    :param ocr_model_id: 模型服务id
    返回切分chunks
    """
    logger.info("----->模型解析:pdf按页解析处理本地文件%s" % add_file_path)
    # merged_data = defaultdict(lambda: {"type": "text", "text": "", "page_num": [], "length": 0})
    merged_list = []
    sorted_result = []
    # full_text = ""
    # file_name = os.path.split(add_file_path)[-1]
    try:
        # 使用fitz打开PDF文件并获取总页数 [EN] Open PDF file using fitz and get total number of pages
        pdf_document = fitz.open(add_file_path)
        num_pages = len(pdf_document)

        with ThreadPoolExecutor(max_workers=MODEL_PARSER_MAX_WORKERS) as executor:  # 调整max_workers以适应你的需求 [EN] Adjust max_workers to suit your needs
            futures = {executor.submit(get_page_data, page_num, add_file_path, ocr_model_id): page_num for page_num in
                       range(1, num_pages + 1)}

            for future in as_completed(futures):
                page_data, page_num = future.result()
                if page_data is not None:
                    merged_list.append({
                        "type": "text",
                        "text": page_data,
                        "page_num": [page_num],
                        "length": len(page_data)
                    })


        # 按 page_num 对合并后的结果进行排序 [EN] Sort merged results by page_num
        sorted_result = sorted(merged_list,
                               key=lambda item: item["page_num"][0] if isinstance(item["page_num"], list) else item[
                                   "page_num"])
        # for item in sorted_result:
        #     full_text += item["text"]
        # with open("./parser_data/%s.txt" % file_name, 'w', encoding='utf-8') as c_file:
        #     c_file.write(full_text)
    except Exception as err:
        logger.error("====> model_parser error %s" % err)
        logger.error("Failed to process the entire PDF document.")
        import traceback
        logger.error(traceback.format_exc())
    return sorted_result


def model_parser_file(add_file_path, ocr_model_id):
    """
    处理整个PDF文档，按页并发调用OCR文档解析工具服务
    :param add_file_path: 文件路径
    :param ocr_model_id: 模型服务id
    返回文件路径
    """
    logger.info("----->模型解析:pdf处理本地文件%s" % add_file_path)
    # merged_data = defaultdict(lambda: {"type": "text", "text": "", "page_num": [], "length": 0})
    merged_list = []
    # sorted_result = []
    full_text = ""
    file_name = os.path.split(add_file_path)[-1]
    try:
        # 使用fitz打开PDF文件并获取总页数 [EN] Open PDF file using fitz and get total number of pages
        pdf_document = fitz.open(add_file_path)
        num_pages = len(pdf_document)

        with ThreadPoolExecutor(max_workers=MODEL_PARSER_MAX_WORKERS) as executor:  # 调整max_workers以适应你的需求 [EN] Adjust max_workers to suit your needs
            futures = {executor.submit(get_page_data, page_num, add_file_path, ocr_model_id): page_num for page_num in
                       range(1, num_pages + 1)}

            for future in as_completed(futures):
                page_data, page_num = future.result()
                if page_data is not None:
                    merged_list.append({
                        "type": "text",
                        "text": page_data,
                        "page_num": [page_num],
                        "length": len(page_data)
                    })


        # 按 page_num 对合并后的结果进行排序 [EN] Sort merged results by page_num
        sorted_result = sorted(merged_list,
                               key=lambda item: item["page_num"][0] if isinstance(item["page_num"], list) else item[
                                   "page_num"])
        for item in sorted_result:
            full_text += item["text"]
        output_file_path = "./parser_data/%s.txt" % file_name
        with open(output_file_path, 'w', encoding='utf-8') as c_file:
            c_file.write(full_text)
    except Exception as err:
        logger.error("PDF模型解析服务异常 %s" % err)
        logger.error(traceback.format_exc())
    return output_file_path



