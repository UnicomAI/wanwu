import json
import time
import requests
from typing import List

import numpy as np

from model.model_manager import get_model_configure
from settings import MINIO_ADDRESS, REPLACE_MINIO_DOWNLOAD_URL, EMBEDDING_BATCH_SIZE

from log.logger import logger


def batch_list(lst: list, batch_size=32):
    for i in range(0, len(lst), batch_size):
        yield lst[i:i + batch_size]

def _execute_embedding(request_func, log_prefix="Embedding"):
    rate_limit_backoff = [10, 20, 40, 60]
    other_error_max_retries = 2
    other_error_wait = 0.5

    attempt = 0
    last_error = None

    while attempt < max(len(rate_limit_backoff), other_error_max_retries) + 1:
        try:
            start_time = time.time()
            response_json = request_func()
            dense_vec_data = response_json["data"]
            response_metadata = {
                "object": response_json.get("object"),
                "model": response_json.get("model"),
                "usage": response_json.get("usage"),
                "data_count": len(dense_vec_data)
            }
            logger.info(f"Response metadata: {json.dumps(response_metadata)}")

            # 调试日志：记录前3个向量的维度
            if dense_vec_data:
                sample_info = [
                    {"index": i, "vec_len": len(item["embedding"])}
                    for i, item in enumerate(dense_vec_data[:3])
                ]
                logger.debug(f"Sample vector dimensions: {sample_info}")
            latency = time.time() - start_time
            logger.info(f"Received {log_prefix} response in {latency:.2f}s")

            # 构建结果
            result_list = [
                {"dense_vec": emb_vec["embedding"]}
                for emb_vec in dense_vec_data
            ]
            return {"result": result_list}

        except Exception as e:
            # 增强错误日志
            error_details = f"Error: {type(e).__name__} - {str(e)}"
            last_error = error_details

            # 尝试获取OpenAI错误详情
            if hasattr(e, 'response'):
                try:
                    status_code = getattr(e.response, "status_code", "N/A")
                    error_body = e.response.text if hasattr(e.response, "text") else "N/A"
                    error_details += f" | HTTP {status_code}: {error_body[:200]}"
                except Exception as parse_err:
                    error_details += f" | Failed to parse error: {parse_err}"

            logger.error(f"{log_prefix} request failed (attempt {attempt + 1}): {error_details}")

            # 判断是否限流
            is_rate_limited = error_details and "429" in error_details
            if is_rate_limited:
                if attempt < len(rate_limit_backoff):
                    wait_time = rate_limit_backoff[attempt]
                    logger.warning(f"Rate limited (429). Retrying after {wait_time}s...")
                    time.sleep(wait_time)
                    attempt += 1
                    continue
                else:
                    logger.error("Exceeded max retries due to rate limiting.")
                    break
            else:
                if attempt < other_error_max_retries:
                    logger.warning(f"Non-429 error. Retrying after {other_error_wait}s...")
                    time.sleep(other_error_wait)
                    attempt += 1
                    continue
                else:
                    logger.error("Exceeded max retries for non-429 errors.")
                    break

    # 最终错误处理
    raise RuntimeError(f"Failed to get {log_prefix.lower()}s after retries, last error: {last_error}")


def get_embs(texts: list, embedding_model_id=""):
    """ 使用 openai embedding 协议（HTTP）获取文本向量。

    走 raw requests.post 以让 RequestsInstrumentor 自动注入 W3C traceparent，
    打通到上游 embedding 服务的 trace 链路（OpenAI SDK 底层 httpx 不在覆盖范围内）。
    """
    emb_info = get_model_configure(embedding_model_id)
    logger.info(f"Starting embedding request for {len(texts)} texts, model: {emb_info.model_name}")

    api_key = emb_info.api_key or "fake api key"
    # 安全记录API Key（仅显示部分）
    masked_key = api_key[:4] + "****" + api_key[-4:] if len(api_key) > 8 else "****"

    emb_model_url = emb_info.endpoint_url
    if not emb_model_url.endswith("/embeddings"):
        emb_model_url = emb_model_url.rstrip("/") + "/embeddings"

    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}",
    }
    payload = {
        "model": emb_info.model_name,
        "input": texts,
        "encoding_format": "float",
    }

    # 安全的请求日志
    request_details = {
        "url": emb_model_url,
        "model": emb_info.model_name,
        "api_key": masked_key,  # 使用脱敏后的key
        "text_count": len(texts),
        "input": texts
    }
    logger.info(f"Sending embedding request: {json.dumps(request_details, ensure_ascii=False)}")

    def request_func():
        response = requests.post(emb_model_url, headers=headers, json=payload, timeout=120)
        if response.status_code != 200:
            raise Exception(f"HTTP {response.status_code}: {response.text}")
        return response.json()

    return _execute_embedding(request_func, log_prefix="Embedding")


def get_multimodal_embs(inputs: List[dict], embedding_model_id=""):
    """
    获取多模态向量，返回向量列表
    :param inputs: 输入列表，元素可以是：
                         [{"text": "xxx", "image": "url"}, {"text": "xxx"}, {"image": "url"}]
                         即支持纯文本、纯图片或图文对（文+图）
    :param embedding_model_id: 指定的模型ID
    """
    # 过滤和校验 inputs
    filtered_inputs = []
    for idx, item in enumerate(inputs):
        if not isinstance(item, dict):
            raise ValueError(f"Input item at index {idx} is invalid. Expected dict, but got {type(item)}.")

        clean_item = {}
        # 仅保留 text 和 image key
        if "text" in item:
            clean_item["text"] = item["text"]
        if "image" in item:
            if not item["image"].startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in item["image"]:
                suffix = item["image"].replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
                clean_item["image"] = f"http://{MINIO_ADDRESS}/{suffix}"
            else:
                clean_item["image"] = item["image"]
        if not clean_item:
            raise ValueError(f"Input item at index {idx} is invalid. must contain 'text' or 'image' key. Item: {item}")

        filtered_inputs.append(clean_item)

    if not filtered_inputs:
        raise ValueError("Input list is empty after filtering. Each item must contain 'text' or 'image' key.")

    inputs = filtered_inputs
    logger.info(f"get_multimodal_embs inputs: {inputs}")

    emb_info = get_model_configure(embedding_model_id)
    if not emb_info.is_multimodal:
        raise ValueError(f"Model {emb_info.model_name} does not support multimodal embedding.")
    logger.info(f"Starting multimodal embedding request for {len(inputs)} inputs, model: {emb_info.model_name}")

    api_key = emb_info.api_key or "fake api key"
    # 安全记录API Key（仅显示部分）
    masked_key = api_key[:4] + "****" + api_key[-4:] if len(api_key) > 8 else "****"

    emb_model_url = emb_info.endpoint_url
    if not emb_model_url.endswith("/multimodal-embeddings"):
        emb_model_url = emb_model_url.rstrip("/") + "/multimodal-embeddings"
    model_name = emb_info.model_name

    # for test
    # emb_model_url = "https://api.jina.ai/v1/embeddings"
    # api_key = "jina_xxxx"
    # masked_key = api_key
    # model_name = "jina-embeddings-v4"

    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {api_key}"
    }

    payload = {
        "model": model_name,
        "input": inputs
    }

    request_details = {
        "url": emb_model_url,
        "model": model_name,
        "api_key": masked_key,
        "input_count": len(inputs)
    }
    logger.info(f"Sending multimodal embedding request: {json.dumps(request_details, ensure_ascii=False)}")

    def request_func():
        response = requests.post(emb_model_url, headers=headers, json=payload, timeout=120)
        if response.status_code != 200:
            raise Exception(f"HTTP {response.status_code}: {response.text}")
        response_json = response.json()
        return response_json

    try:
        # 尝试批量执行
        return _execute_embedding(request_func, log_prefix="Multimodal embedding")
    except Exception as batch_error:
        # 如果批量失败且输入中有图片，开启逐个重试模式
        has_image = any("image" in item for item in inputs)
        if not has_image:
            raise batch_error  # 如果全是文本也失败了，直接抛出

        logger.warning(f"Batch embedding failed: {batch_error}. Switching to serial mode...")

        final_results = []
        for idx, single_input in enumerate(inputs):
            # 定义单个请求的闭包
            def single_request_func():
                inner_payload = {"model": model_name, "input": [single_input]}
                resp = requests.post(emb_model_url, headers=headers, json=inner_payload, timeout=60)
                if resp.status_code != 200:
                    raise Exception(f"HTTP {resp.status_code}: {resp.text}")
                return resp.json()

            try:
                # 依然复用 _execute_embedding 的重试逻辑，但只针对这一个 input
                single_res = _execute_embedding(single_request_func, log_prefix=f"Single-item-{idx}")
                # _execute_embedding 返回的是 {"result": [{"dense_vec": [...]}]}
                final_results.append(single_res["result"][0])
            except Exception as single_error:
                logger.error(f"Failed to embed item at index {idx}: {single_error}")
                # 如果包含图片且失败，填充 None
                if "image" in single_input:
                    final_results.append({"dense_vec": None})
                else:
                    # 如果纯文本也失败
                    raise single_error

        return {"result": final_results}


def calculate_cosine(query, contents, embedding_model_id="") -> list[float]:
    query_vector_scores = []
    emb_info = get_model_configure(embedding_model_id)

    contents_vector = []
    if emb_info.is_multimodal:
        query_vector = get_multimodal_embs([{"text": query}], embedding_model_id=embedding_model_id)["result"][0]["dense_vec"]
        for batch_doc in batch_list(contents, batch_size=EMBEDDING_BATCH_SIZE):
            contents_vector.extend(
                get_multimodal_embs([{"text": item} for item in batch_doc], embedding_model_id=embedding_model_id)["result"]
            )
    else:
        query_vector = get_embs([query], embedding_model_id=embedding_model_id)["result"][0]["dense_vec"]
        for batch_doc in batch_list(contents, batch_size=EMBEDDING_BATCH_SIZE):
            contents_vector.extend(
                get_embs(batch_doc, embedding_model_id=embedding_model_id)["result"]
            )

    for item in contents_vector:
        vec1 = np.array(query_vector)
        vec2 = np.array(item["dense_vec"])

        # calculate dot product
        dot_product = np.dot(vec1, vec2)

        # calculate norm
        norm_vec1 = np.linalg.norm(vec1)
        norm_vec2 = np.linalg.norm(vec2)

        # calculate cosine similarity
        cosine_sim = dot_product / (norm_vec1 * norm_vec2)
        query_vector_scores.append(cosine_sim)

    return query_vector_scores

if __name__ == '__main__':
    input = [
        {
            "text": "海滩上美丽的日落"
        },
        {
            "text": "A beautiful sunset over the beach",
            "image": "iVBORw0KGgoAAAANSUhEUgAAABwAAAA4CAIAAABhUg/jAAAAMklEQVR4nO3MQREAMAgAoLkoFreTiSzhy4MARGe9bX99lEqlUqlUKpVKpVKpVCqVHksHaBwCA2cPf0cAAAAASUVORK5CYII="
        },
        {
            "image": "https://i.ibb.co/r5w8hG8/beach2.jpg"
        },
    ]
    get_multimodal_embs(input, 5)
