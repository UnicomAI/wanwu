import json
import traceback
import os
import html2text
import requests
# import urllib3
import logging

from utils.constant import MODEL_PARSER_MAX_WORKERS
from model_manager.model_config import get_model_configure, OcrModelConfig

hl2txt = html2text.HTML2Text()

from concurrent.futures import ThreadPoolExecutor, as_completed
# from PyPDF2 import PdfReader, PdfWriter
import time
import fitz
from pathlib import Path
from utils import minio_utils
from settings import MINIO_ADDRESS, REPLACE_MINIO_DOWNLOAD_URL

logger = logging.getLogger(__name__)

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
    full_file_name = path_obj.name  # 带扩展名的完整文件名（用于formData的fileName）

    try:
        # 打开PDF文件

        pdf_document = fitz.open(add_file_path)

        if page_num > len(pdf_document) or page_num < 1:
            logger.error(f"Page number {page_num} is out of range.")
            return None, page_num

        # 创建一个新的PDF文档并将指定页添加到其中

        page_pdf_path = f"{file_name}_page_{page_num}.pdf"
        # 组合成新的文件路径
        output_pdf_path = os.path.join(directory, page_pdf_path)
        logger.info("======>model_parser_utils,get_page_data=%s" % output_pdf_path)
        new_pdf = fitz.open()  # 新建一个空的PDF文档
        new_pdf.insert_pdf(pdf_document, from_page=page_num - 1, to_page=page_num - 1)
        new_pdf.save(output_pdf_path)
        new_pdf.close()

        if ocr_model_id == "":
            logger.error("ocr_model_id为空")
            if os.path.exists(output_pdf_path):
                os.remove(output_pdf_path)
            return None, page_num

        # ocr_model_id 有效后再上传 MinIO
        upload_result = minio_utils.upload_local_file(output_pdf_path)
        if upload_result['code'] != 0:
            logger.error(f"页 {page_num} 上传 MinIO 失败")
            if os.path.exists(output_pdf_path):
                os.remove(output_pdf_path)
            return None, page_num

        minio_url = upload_result['download_link']
        # 与 ASR 保持一致：将公开代理 URL 替换为内网 MinIO 地址
        if not minio_url.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in minio_url:
            suffix = minio_url.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
            minio_url = f"http://{MINIO_ADDRESS}/{suffix}"
        object_name = minio_url.split('/')[-1]

        model_config = get_model_configure(ocr_model_id)
        wanwu_ocr_url = ""
        api_key = ""
        model_name = ""
        if isinstance(model_config, OcrModelConfig):
            wanwu_ocr_url = model_config.endpoint_url + "/ocr"
            api_key = model_config.api_key
            model_name = getattr(model_config, "model_name", "") or ""
        headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }

        data = {
            "extract_image": 1,
            "extract_image_content": 1,
            "url": minio_url,
            "fileName": full_file_name,
        }

        rate_limit_backoff = [10, 20, 40, 60]  # 限流退避
        other_error_max_retries = 2  # 其他错误最多重试2次
        other_error_wait = 0.5  # 每次0.5s

        attempt = 0
        while attempt < max(len(rate_limit_backoff), other_error_max_retries) + 1:
            try:

                r = requests.post(wanwu_ocr_url, headers=headers, json=data, timeout=60)
                logger.info("====>wanwu_ocr_url=%s,data=%s" % (wanwu_ocr_url, json.dumps(data, ensure_ascii=False)))
                ret_json = r.json()
                # logger.info(f"model_parser_utils.get_page_data result: {ret_json}")
                r.raise_for_status()  # 触发HTTP错误状态码的异常
                if ret_json.get("code") == 0:
                    data_obj = ret_json.get("data", {}) or {}
                    text = data_obj.get("fullContent", "")
                    image_url_prefix = ret_json.get("prefixImageUrl", "")
                    if image_url_prefix:
                        text, replace_info = minio_utils.replace_minio_url(text, image_url_prefix)
                        logger.info(f"get_page_data replace url info: {replace_info}")
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

                # 判断是否限流(429)
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
                # 只清理本地临时文件，MinIO 文件保留供重试或运维独立清理
                if os.path.exists(output_pdf_path):
                    os.remove(output_pdf_path)
                time.sleep(0.1)
    except Exception as e:
        logger.error(f"处理页 {page_num} 失败：{e}")
        logger.error(traceback.format_exc())
        return None, page_num

    # 最终错误处理
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
        # 使用fitz打开PDF文件并获取总页数
        pdf_document = fitz.open(add_file_path)
        num_pages = len(pdf_document)

        with ThreadPoolExecutor(max_workers=MODEL_PARSER_MAX_WORKERS) as executor:  # 调整max_workers以适应你的需求
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


        # 按 page_num 对合并后的结果进行排序
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
        # 使用fitz打开PDF文件并获取总页数
        pdf_document = fitz.open(add_file_path)
        num_pages = len(pdf_document)

        with ThreadPoolExecutor(max_workers=MODEL_PARSER_MAX_WORKERS) as executor:  # 调整max_workers以适应你的需求
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


        # 按 page_num 对合并后的结果进行排序
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


def model_parser_image(image_file_path, ocr_model_id):
    """
    调用文档解析模型处理图片文件
    :param image_file_path: 图片文件路径（.jpg/.jpeg/.png）
    :param ocr_model_id: 文档解析模型ID
    :return: 解析后的文本描述（Markdown 格式）

    说明：改造前图片在 "ocr" 模式下由 ocr_utils 提取文字。改造后 "ocr" 被删除，
          图片需改由 /pdf-parser（文档解析模型）端点处理，否则 model 模式下图片无文字提取能力。
          此函数上传图片到 MinIO，通过 URL 请求调用 /pdf-parser 端点，解析后清理 MinIO 临时文件。
    """
    if ocr_model_id == "":
        logger.error("ocr_model_id为空")
        return ""

    model_config = get_model_configure(ocr_model_id)
    wanwu_ocr_url = ""
    api_key = ""
    if isinstance(model_config, OcrModelConfig):
        wanwu_ocr_url = model_config.endpoint_url + "/ocr"
        api_key = model_config.api_key

    file_name = os.path.split(image_file_path)[-1]

    # 上传本地图片到 MinIO
    upload_result = minio_utils.upload_local_file(image_file_path)
    if upload_result['code'] != 0:
        logger.error(f"图片 {file_name} 上传 MinIO 失败")
        return ""

    minio_url = upload_result['download_link']
    # 与 ASR 保持一致：将公开代理 URL 替换为内网 MinIO 地址
    if not minio_url.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in minio_url:
        suffix = minio_url.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
        minio_url = f"http://{MINIO_ADDRESS}/{suffix}"
    object_name = minio_url.split('/')[-1]

    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    data = {
        "extract_image": 1,
        "extract_image_content": 1,
        "url": minio_url,
        "fileName": file_name,
    }

    try:
        r = requests.post(wanwu_ocr_url, headers=headers, json=data, timeout=60)
        r.raise_for_status()
        ret_json = r.json()
        if ret_json.get("code") == 0:
            data_obj = ret_json.get("data", {}) or {}
            text = data_obj.get("fullContent", "")
            image_url_prefix = ret_json.get("prefixImageUrl", "")
            if image_url_prefix:
                text, replace_info = minio_utils.replace_minio_url(text, image_url_prefix)
            return text
        else:
            logger.error(f"图片模型解析失败：{ret_json.get('message', '未知错误')}")
    except Exception as e:
        logger.error(f"图片模型解析异常：{e}")
    finally:
        # MinIO 文件保留供运维独立清理
        pass
    return ""

