import traceback

import requests
import json
import uuid
import os
import html2text
import datetime
import requests
from collections import defaultdict
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
import os, time, traceback
# import urllib3
from datetime import datetime, timedelta
from logging_config import setup_logging
logger_name='rag_ocr_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))
from utils.constant import MAX_SENTENCE_SIZE, OCR_MAX_WORKERS
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


def ocr_parser_text(add_file_path, ocr_model_id):
    """
    :param add_file_path:本地文件路径
    Args:
        add_file_path:

    Returns:

    """
    logger.info("-----视觉影印扫描版pdf处理，解析返回text-------")
    file_name = os.path.split(add_file_path)[-1]
    text = ""
    if not ocr_model_id:
        logger.error("ocr_model_id为空，无法进行图片OCR解析")
        return text

    model_config = get_model_configure(ocr_model_id)
    wanwu_ocr_url = ""
    api_key = ""
    if isinstance(model_config, OcrModelConfig):
        wanwu_ocr_url = model_config.endpoint_url + "/ocr"
        api_key = model_config.api_key

    logger.info(f"构造OCR请求URL: {wanwu_ocr_url}")
    files = {"file": open(add_file_path, 'rb').read()}
    headers = {"Authorization": f"Bearer {api_key}"}
    file_data = {
        "filename": file_name
    }
    # 记录请求详细信息 [EN] Logging request details
    logger.info(f"发起OCR请求 | URL: {wanwu_ocr_url} | 文件名: {file_name} | 模型ID: {ocr_model_id}")

    try:
        start_time = time.time()
        r = requests.post(wanwu_ocr_url, files=files, headers=headers, data=file_data, timeout=600)
        elapsed = time.time() - start_time
        # 记录响应基本信息 [EN] Record basic response information
        logger.info(f"收到OCR响应 | 状态码: {r.status_code} | 耗时: {elapsed:.2f}秒")
        if r.status_code == 200:
            ret_json = json.loads(r.text)
            logger.info(json.dumps(ret_json, ensure_ascii=False))
            if ret_json.get("code") == 0:
                parser_chunks = ret_json["data"]
                for item in parser_chunks:
                    text += item["text"]

    except requests.exceptions.Timeout:
        # 处理超时异常 [EN] Handling timeout exceptions
        logger.error("ocr request timed out.")
    except requests.exceptions.RequestException as e:
        # 其他类型的异常处理 [EN] Other types of exception handling
        logger.error(f"An error occurred on ocrcall: {e}")
    return text


def ocr_parser_native(add_file_path, ocr_model_id):
    """
    :param add_file_path: 本地文件路径
    :param ocr_model_id: OCR模型ID
    :return: OCR解析结果(JSON格式)
    """
    logger.info("-----视觉影印扫描版pdf处理，解析返回text-------")
    file_name = os.path.split(add_file_path)[-1]
    ret_json = {}

    if not ocr_model_id:
        logger.error("ocr_model_id为空，无法进行OCR解析")
        return None


    model_config = get_model_configure(ocr_model_id)
    wanwu_ocr_url = ""
    if isinstance(model_config, OcrModelConfig):
        wanwu_ocr_url = model_config.endpoint_url + "/ocr"
    logger.info(f"构造OCR请求URL: {wanwu_ocr_url}")

    # 读取文件并记录基本信息 [EN] Read the file and record basic information
    try:
        with open(add_file_path, 'rb') as f:
            file_content = f.read()
        file_size = len(file_content)
        logger.info(f"读取文件成功 | 文件名: {file_name} | 文件大小: {file_size} bytes")
    except Exception as e:
        logger.error(f"文件读取失败 | 路径: {add_file_path} | 错误: {str(e)}")
        return None

    files = {"file": (file_name, file_content, "application/pdf")}

    # 记录请求详细信息 [EN] Logging request details
    logger.info(f"发起OCR请求 | URL: {wanwu_ocr_url} | 文件名: {file_name} | 模型ID: {ocr_model_id}")
    logger.debug(f"请求头信息: {files.keys()}")  # 调试级别记录详细头信息 [EN] Debug level logging detailed header information

    try:
        start_time = time.time()
        r = requests.post(wanwu_ocr_url, files=files, timeout=600)
        elapsed = time.time() - start_time

        # 记录响应基本信息 [EN] Record basic response information
        logger.info(f"收到OCR响应 | 状态码: {r.status_code} | 耗时: {elapsed:.2f}秒")

        if r.status_code == 200:
            # 记录成功响应摘要 [EN] Logging a summary of successful responses
            logger.info(f"OCR请求成功 | 文件名: {file_name}")
            logger.debug(f"完整响应内容: {r.text}")  # 调试级别记录完整响应 [EN] Debug level logs full response

            try:
                ret_json = r.json()
                # 记录解析后的关键信息摘要 [EN] Record the summary of key information after parsing
                if isinstance(ret_json, dict):
                    logger.info(f"解析JSON成功 | 返回键值: {list(ret_json.keys())}")
                    if 'data' in ret_json and 'ocr_result' in ret_json['data']:
                        result_count = len(ret_json['data']['ocr_result'])
                        logger.info(f"识别结果统计 | 页面数: {result_count}")
            except json.JSONDecodeError:
                logger.error(f"响应JSON解析失败 | 响应文本: {r.text[:200]}...")
                return None
        else:
            # 记录错误响应详细信息 [EN] Logging error response details
            logger.error(f"OCR请求失败 | 状态码: {r.status_code} | 错误信息: {r.text[:500]}...")
            if r.status_code >= 500:
                logger.error("服务器端错误，请检查OCR服务状态")
            elif r.status_code == 404:
                logger.error("接口路径错误，请验证OCR模型ID")

    except requests.exceptions.Timeout:
        logger.error("OCR请求超时 | 已超过600秒等待时间")
    except requests.exceptions.RequestException as e:
        logger.error(f"OCR请求异常 | 错误类型: {type(e).__name__} | 详细信息: {str(e)}")
    except Exception as e:
        logger.error(f"处理OCR响应时发生未预期错误 | 错误: {str(e)}")

    return ret_json



def get_page_data(page_num, add_file_path, ocr_model_id):
    """
    获取单页的数据并调用OCR服务
    :param page_num: 页码
    :param add_file_path: 文件路径
    :return: OCR结果
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
        logger.info("======>ocr_utils,get_page_data=%s" % output_pdf_path)
        new_pdf = fitz.open()  # 新建一个空的PDF文档 [EN] Create a new empty PDF document
        new_pdf.insert_pdf(pdf_document, from_page=page_num - 1, to_page=page_num - 1)
        new_pdf.save(output_pdf_path)
        new_pdf.close()

        # 构造请求参数（符合formData要求） [EN] Construct request parameters (in line with formData requirements)
        files = {"file": (page_pdf_path, open(output_pdf_path, 'rb').read(), "application/pdf")}
        data = {"fileName": full_file_name}  # 显式传递原始文件名 [EN] Explicitly pass the original file name

        if ocr_model_id == "":
            logger.error("ocr_model_id为空")
            return None, page_num

        model_config = get_model_configure(ocr_model_id)
        wanwu_ocr_url = ""
        api_key = ""
        if isinstance(model_config, OcrModelConfig):
            wanwu_ocr_url = model_config.endpoint_url + "/ocr"
            api_key = model_config.api_key
        headers = {"Authorization": f"Bearer {api_key}"}

        # 使用重试机制 [EN] Use retry mechanism
        session = requests.Session()
        retry_strategy = Retry(
            total=3,
            backoff_factor=1,
            status_forcelist=[500, 502, 503, 504],
            method_whitelist=["POST"]
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        session.mount("http://", adapter)

        try:
            r = session.post(wanwu_ocr_url, files=files, headers=headers, data=data, timeout=60)
            r.raise_for_status()  # 触发HTTP错误状态码的异常 [EN] Exceptions that trigger HTTP error status codes
            ret_json = r.json()
            # logger.info("page_num:%s,ocr_result:%s" % (page_num, json.dumps(ret_json, ensure_ascii=False)))
            # 解析返回结果（符合新的JSON结构） [EN] Parse the return result (conform to the new JSON structure)
            if ret_json.get("code") == 0:
                page_data = ret_json.get("data", [])
                # 补充当前页码到返回数据中（若接口返回的page_num不正确） [EN] Add the current page number to the returned data (if the page_num returned by the interface is incorrect)
                for item in page_data:
                    item["page_num"] = [page_num]  # 确保page_num字段为列表格式 [EN] Make sure the page_num field is in list format
                return page_data, page_num
            else:
                logger.error(f"页 {page_num} OCR失败：{ret_json.get('message', '未知错误')}")
                return None, page_num
        except requests.exceptions.HTTPError as e:
            logger.error(f"页 {page_num} HTTP错误：{e}")
            return None, page_num
        except requests.exceptions.Timeout:
            logger.error(f"页 {page_num} 请求超时")
            return None, page_num
        except requests.exceptions.RequestException as e:
            logger.error(f"页 {page_num} 请求异常：{e}")
            return None, page_num
        finally:
            # 清理临时文件 [EN] Clean temporary files
            if os.path.exists(output_pdf_path):
                os.remove(output_pdf_path)
            time.sleep(0.1)

    except Exception as e:
        logger.error(f"处理页 {page_num} 失败：{e}")
        logger.error(traceback.format_exc())
        return None, page_num

def ocr_parser(add_file_path, ocr_model_id):
    """
    处理整个PDF文档，按页并发调用OCR服务
    :param add_file_path: 文件路径
    """
    #logger.info("-----视觉影印扫描版pdf处理add_file_path=%s" % add_file_path) [EN] logger.info("-----Visual photocopying scanned version pdf processing add_file_path=%s" % add_file_path)
    merged_data = defaultdict(lambda: {"type": "text", "text": "", "page_num": [], "length": 0})
    merged_list = []
    sorted_result = []

    try:
        # 使用fitz打开PDF文件并获取总页数 [EN] Open PDF file using fitz and get total number of pages
        pdf_document = fitz.open(add_file_path)
        num_pages = len(pdf_document)
        pdf_document.close()  # 提前关闭，避免文件占用 [EN] Close in advance to avoid file occupation

        with ThreadPoolExecutor(max_workers=OCR_MAX_WORKERS) as executor:  # 调整max_workers以适应你的需求 [EN] Adjust max_workers to suit your needs
            futures = {executor.submit(get_page_data, page_num, add_file_path, ocr_model_id): page_num for page_num in
                       range(1, num_pages + 1)}

            for future in as_completed(futures):
                page_data, page_num = future.result()
                if page_data is not None:
                    for item in page_data:
                        if item["type"] == 'table':
                            md_text = hl2txt.handle(item["text"])  # 将html_string转换为markdown [EN] Convert html_string to markdown
                            item["text"] = md_text
                        if item["type"] not in ['page-header', 'page-footer']:
                            if len(merged_data[page_num]['text'] + item["text"]) < MAX_SENTENCE_SIZE:
                                merged_data[page_num]["text"] += item["text"]  # 拼接文本 [EN] Splice text
                                merged_data[page_num]['page_num'].append(page_num)
                                merged_data[page_num]['length'] = len(merged_data[page_num]['text'])
                            else:
                                merged_list.append({
                                    "type": item["type"],
                                    "text": item["text"],
                                    "page_num": [page_num],
                                    "length": len(item["text"])
                                })

        # 将merged_data中的值添加到merged_list [EN] Add values ​​from merged_data to merged_list
        merged_list.extend(merged_data.values())

        # 按 page_num 对合并后的结果进行排序 [EN] Sort merged results by page_num
        sorted_result = sorted(merged_list,
                               key=lambda item: item["page_num"][0] if isinstance(item["page_num"], list) else item[
                                   "page_num"])
    except Exception as err:
        logger.error("====> ocr_parser error %s" % err)
        logger.error("Failed to process the entire PDF document.")
        import traceback
        logger.error(traceback.format_exc())
    return sorted_result


if __name__ == "__main__":
    add_file_path ="./parser_data/2019年联通学院一号楼三层多功能厅扩声系统采购.pdf"
    ocr_model_id = "89"
    sort_list = ocr_parser(add_file_path, ocr_model_id)
