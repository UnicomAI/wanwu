import os
import json
import ssl
import re
import time
import shutil
import requests
import numpy as np
import uuid
import hashlib
import tiktoken
import urllib.parse
from utils.knowledge_base_utils import *
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from sse_starlette.sse import ServerSentEvent, EventSourceResponse
from langchain.prompts import PromptTemplate
from model_manager.model_config import get_model_configure, LlmModelConfig
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

from datetime import datetime, timedelta
from utils.prompts import PROMPT_TEMPLATE, CITATION_INSTRUCTION
from settings import (SSE_USE_MONGO, TEMPERATURE, MONGO_URL, REPLACE_MINIO_DOWNLOAD_URL, MINIO_ADDRESS,
                      TRUNCATE_PROMPT, CONTEXT_LENGTH, RESERVED_FOR_OUTPUT)

from logging_config import init_logging

from pymongo import MongoClient
from utils import redis_utils
from utils.constant import CHUNK_SIZE
from utils.otel import init_tracer
from utils.trace_asgi import TraceLoggingMiddleware, FirstChunkSpanMiddleware
from utils.tools import get_query_dict_cache, query_rewrite

# 初始化 OpenTelemetry（必须在 FastAPI app 创建之前）
init_tracer("rag-wanwu-sse-service")

# 自动 instrument requests 库（出站 HTTP 传播 traceparent）
from opentelemetry.instrumentation.requests import RequestsInstrumentor

RequestsInstrumentor().instrument()

user_data_path = r'./user_data'
app = FastAPI()
init_logging()
logger = logging.getLogger(__name__)

# 自动 instrument FastAPI（入站 HTTP 提取 traceparent，创建 Span）
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

# exclude_spans 关闭按 ASGI send/receive 事件建子 span：流式响应每个分片都会触发一次
# send，否则一条 trace 会产生成百上千个 "http send" span。首分片的 span 由
# FirstChunkSpanMiddleware 单独补回。
FastAPIInstrumentor.instrument_app(app, exclude_spans=["send", "receive"])

# 添加日志记录（含 trace_id/span_id 关联）
app.add_middleware(TraceLoggingMiddleware)
# 流式响应只为首个分片建一个 span（首 token 耗时），其余分片不建 span
app.add_middleware(FirstChunkSpanMiddleware)
# 解决跨域问题
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=False,
    allow_methods=["*"],
    allow_headers=["*"]
)
# 初始化 MongoDB 客户端
client = MongoClient(MONGO_URL, 0, connectTimeoutMS=5000, serverSelectionTimeoutMS=3000)

collection = client['rag']['rag_user_logs']
redis_client = redis_utils.get_redis_connection()

tiktoken_cache_dir = "/opt/tiktoken_cache"
os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir
encoding = tiktoken.encoding_for_model("gpt-4")

def get_prompt(question: str,
               search_list: list,
               default_answer: str,
               auto_citation: bool,
               prompt_template: str = "",
               context_size: int = CONTEXT_LENGTH,
               max_tokens: int = 4096):
    citation = CITATION_INSTRUCTION if auto_citation else ""
    if auto_citation and default_answer:
        default_answer_text = "请仅基于提供的参考信息中上下文提供答案。如果提供的参考信息中的所有上下文对回答问题均无帮助，请直接输出:%s" % default_answer
    else:
        default_answer_text = ""

    # 构造一个没有 context 的模板来估算剩余空间
    template_to_use = prompt_template if (
                len(prompt_template) > 0 and "{question}" in prompt_template) else PROMPT_TEMPLATE

    # 估算除了 {context} 以外占用的 token
    base_prompt_without_context = template_to_use.replace("{context}", "")
    # 填充实际变量（不含 context）
    base_filled = base_prompt_without_context.format(
        question=question,
        citation=citation,
        default_answer=default_answer_text
    )

    base_tokens = len(encoding.encode(base_filled))
    available_tokens_for_context = context_size - base_tokens - RESERVED_FOR_OUTPUT - 50  # 预留50个token缓冲

    # 处理并截取 context
    valid_search_list = []
    res_search_list = []  # 返回的 res_search_list
    current_context_tokens = 0

    for i, x in enumerate(search_list):
        snippet = x['snippet']
        item_text = f"\n【{i + 1}^】\n{snippet}" if auto_citation else snippet
        item_tokens = len(encoding.encode(item_text))

        if current_context_tokens + item_tokens <= available_tokens_for_context:
            valid_search_list.append(item_text)
            res_search_list.append(x)
            current_context_tokens += item_tokens
        else:
            # 计算剩余可用的空位
            remaining_tokens = available_tokens_for_context - current_context_tokens

            if remaining_tokens > 0:
                # 将 item_text 编码，截取前 N 个 token，再解码回文本
                tokens = encoding.encode(item_text)
                truncated_tokens = tokens[:remaining_tokens]
                truncated_text = encoding.decode(truncated_tokens) + "..."  # 添加省略号提示

                valid_search_list.append(truncated_text)
                res_search_list.append(x)
                current_context_tokens += remaining_tokens

            break

    context = "\n".join(valid_search_list)

    formatted_prompt = PromptTemplate(
        template=template_to_use,
        input_variables=["citation", "default_answer", "question", "context"] if "{citation}" in template_to_use else [
            "question", "context"]
    )

    render_kwargs = {"question": question, "context": context}
    if "{citation}" in template_to_use:
        render_kwargs.update({"citation": citation, "default_answer": default_answer_text})

    return formatted_prompt.format(**render_kwargs), res_search_list


def calculate_multimodal_tokens(content_list):
    total_tokens = 0
    for item in content_list:
        if item["type"] == "text":
            total_tokens += len(encoding.encode(item["text"]))
        elif item["type"] == "image_url":
            # 这是一个经验预估值：
            # GPT-4o 低分辨率模式一张图 85 tokens
            # 高分辨率通常 170 - 765+ tokens
            total_tokens += 170
    return total_tokens


@app.post("/rag/knowledge/stream/search")
async def search(request: Request):
    prompt = ''
    history = []

    async def send_request(llm_url:str, api_key: str, llm_data: dict, valid_search_list:list):
        start_time = time.time()
        waitting_response = ""
        answer = ""
        first_output = True
        finish = 0
        current_stream_data = None
        output_str = ""
        try:
            headers = {"Content-Type": "application/json", "Authorization": f"Bearer {api_key}"}
            response = requests.post(llm_url, json=llm_data, headers=headers, verify=False, stream=True)
            logger.info(f'{llm_url} ====== 大模型开始流式输出，发送到大模型参数：' + repr(llm_data))
            response.raise_for_status()
            if response.status_code != 200:
                # 尝试读取一点错误信息
                error_text = response.text[:200] if response.text else "No content"
                raise Exception(f"HTTP Error {response.status_code}: {error_text}")

            has_stream_data = False
            for line in response.iter_lines(decode_unicode=True):
                has_stream_data = True
                current_stream_data = line
                # logger.info(f"raw stream data: {line}") # 调试用
                if not line: continue

                if line.startswith("data:"):
                    # 过滤掉 "data:"
                    line = line[5:]
                # 过滤掉 [DONE] 标记
                if line.strip() == "[DONE]":
                    continue

                datajson = json.loads(line)

                # 优先检查上游错误码 (防御 code:110000 且 data:None 的情况)
                if datajson.get("code") and datajson.get("code") != 0:
                    raise RuntimeError(f"Upstream model error: {datajson.get('msg')}")

                if "choices" in datajson:
                    # 标准格式
                    choices = datajson.get("choices", [{}])
                    content = choices[0].get("delta", {}).get("content", "")
                    reasoning_content = choices[0].get("delta", {}).get("reasoning_content", "")
                else:
                    # 嵌套格式 (data: { choices: ... })
                    # 必须先判断 data 是否为字典，防止 'data': None 导致崩溃
                    data_obj = datajson.get("data")
                    if not isinstance(data_obj, dict):
                        raise ValueError(f"Invalid data structure: 'data' field is not a dictionary. Got: {type(data_obj)}")

                    choices = data_obj.get("choices", [{}])
                    content = choices[0].get("message", {}).get("content", "")
                    reasoning_content = choices[0].get("message", {}).get("reasoning_content", "")

                finish_reason = choices[0].get("finish_reason", "")
                if finish_reason == "stop":
                    finish = 1
                elif finish_reason == "sensitive_cancel":
                    finish = 4
                else:
                    finish = 0

                answer += content
                waitting_response += content
                history_tmp = history.copy()
                history_tmp.append({
                    "query": question,
                    "response": answer,
                    "needHistory": True
                })
                response_info = {
                    'code': 0,
                    "message": "success",
                    "msg_id": msg_id,
                    "data": {"output": content,
                             "reasoning_content": reasoning_content,
                             "searchList": valid_search_list,
                             },
                    "history": history_tmp,
                    "finish": finish
                }
                if score != -1:  # 如果允许返回得分
                    response_info["data"]["score"] = score
                output_str = json.dumps(response_info, ensure_ascii=False)
                yield output_str
                if first_output:
                    end_time = time.time()
                    logger.info(f"question:{question}。开始流式第一个词返回时间：{end_time - start_time}秒")
                    first_output = False

            if not has_stream_data:
                raise Exception("Response body is empty (No stream data received).")

        except Exception as e:  # 如果发生异常，返回错误信息
            logger.error(f"LLM Error url:{llm_url}, current parsed stream data: {current_stream_data}, err: {e}")
            if finish not in [1, 4]:  # 如果模型没有停止输出，则返回错误信息
                response_info = {
                    'code': 1,
                    "message": f"LLM Error:{str(e)}",
                }
                output_str = json.dumps(response_info, ensure_ascii=False)
                yield output_str

        end_time = time.time()
        logger.info(f"question:{question}。流式最后一个词返回时间：{end_time - start_time}秒,返回json:{output_str}")

    async def stream_generate(prompt, history, search_list, question, top_p, repetition_penalty, temperature,
                              custom_model_info, do_sample, score, msg_id, llm_config, enable_thinking):
        model_name = llm_config.model_name
        if isinstance(llm_config, LlmModelConfig):
            llm_url = llm_config.endpoint_url + "/chat/completions"
            api_key = llm_config.api_key
            context_size = llm_config.context_size
            max_tokens = llm_config.max_tokens
        else:
            raise ValueError(f"{model_name} is not llm model")

        messages = []
        available_tokens_for_context = context_size - RESERVED_FOR_OUTPUT - 50  # 50 for buffer
        prompt, valid_search_list = get_prompt(question, search_list, default_answer, auto_citation, prompt_template,
                                               context_size, max_tokens)
        num_tokens = len(encoding.encode(prompt))

        for item in history:
            num_tokens += len(encoding.encode(item["query"]))
            num_tokens += len(encoding.encode(item["response"]))
            if num_tokens > available_tokens_for_context:
                break
            messages.append({"role": "user", "content": item["query"]})
            messages.append({"role": "assistant", "content": item["response"]})
        messages.append({"role": "user", "content": prompt})

        llm_data = {
            "model": model_name,
            "temperature": temperature,
            # "top_k": 5,
            # "top_p": top_p,
            "repetition_penalty": repetition_penalty,
            "do_sample": do_sample,
            "enable_thinking": enable_thinking,
            "stream": True,
            "messages": messages,
        }
        logger.info(f"llm_url:{llm_url},发送到大模型参数：{llm_data}")
        return send_request(llm_url, api_key, llm_data, valid_search_list)

    def curate_reference_text(text: str) -> str:
        """

        只会处理真正的“列表编号”
        不会破坏 URL / 端口号 / 浮点数 / 页码 / 图片链接
        """

        # ========= 保护 URL =========
        url_pattern = r'https?://[^\s)]+'
        urls = re.findall(url_pattern, text)

        protected = text
        for i, url in enumerate(urls):
            protected = protected.replace(url, f"__URL_PLACEHOLDER_{i}__")

        # ========= 保护 Markdown 图片 =========
        md_img_pattern = r'!\[[^\]]*\]\([^)]+\)'
        images = re.findall(md_img_pattern, protected)

        for i, img in enumerate(images):
            protected = protected.replace(img, f"__IMG_PLACEHOLDER_{i}__")

        # ========= 匹配真正的编号 =========
        pattern = r'(?x)(?<![a-zA-Z0-9])(\(?\d+(?:\.\d+)*\)?)([\.．、\)]+[\s]*)(?!分钟|秒|小时|个|只|次|℃|\d)'
        def replace_func(match):
            num = match.group(1)
            # 增加分割线和换行，强制切断语义连续性
            return f"\n【编号 {num}】**: "

        structured = re.sub(pattern, replace_func, protected, flags=re.VERBOSE)

        # ========= 还原图片 =========
        for i, img in enumerate(images):
            structured = structured.replace(f"__IMG_PLACEHOLDER_{i}__", img)

        # ========= 还原 URL =========
        for i, url in enumerate(urls):
            structured = structured.replace(f"__URL_PLACEHOLDER_{i}__", url)

        return structured


    async def multimodal_stream_generate(prompt, history, search_list, question, top_p, repetition_penalty, temperature,
                              custom_model_info, do_sample, score, msg_id, attachment_files, llm_config, enable_thinking):
        model_name = llm_config.model_name
        if isinstance(llm_config, LlmModelConfig):
            llm_url = llm_config.endpoint_url + "/chat/completions"
            api_key = llm_config.api_key
            context_size = llm_config.context_size
            max_tokens = llm_config.max_tokens
        else:
            raise ValueError(f"{model_name} is not llm model")

        if not llm_config.is_vision_support:
            logger.info(" llm is not support vision,multimodal_model_id:%s" % model_id)
            raise Exception(" llm is not support vision,multimodal_model_id:%s" % model_id)
        # ============== 开始组装 messages ==============
        num_tokens = 0
        prompt_content = []
        available_tokens_for_context = context_size - RESERVED_FOR_OUTPUT - 50  # 50 for buffer
        # === 多模态问答提示词构建
        citation = CITATION_INSTRUCTION if auto_citation else ""
        content_item = {"type": "text", "text": f"你是一个问答助手，主要任务是汇总参考信息回答用户问题, 请只根据参考信息中提供的上下文信息回答用户问题，**禁止**直接通过视觉形状猜测功能（必须严格执行）。**严禁**绕过编号仅凭视觉形状相似性进行主观推断（必须严格执行）。 {citation}"}
        prompt_content.append(content_item)
        content_item = {"type": "text", "text": f"用户问题：{question}"}
        prompt_content.append(content_item)
        num_tokens += calculate_multimodal_tokens([content_item])
        if attachment_files:
            content_items = [
                {"type": "text", "text": "用户问题上传的照片："},
                {"type": "image_url", "image_url": {"url": attachment_files[0]["image"]}},
            ]
            prompt_content.extend(content_items)
        prompt_content.append({"type": "text", "text": "\n参考信息：```\n"})
        num_tokens += calculate_multimodal_tokens(prompt_content)
        end_content_item = {"type": "text",
                            "text": "请根据参考信息回答用户问题，请严格按照以下要求输出：\n"
                                    "1. **参考信息中提及图片链接情况的输出要求**：若参考信息提及图片链接且链接格式符合markdown语法规范：“![图片标题](图片链接)” 。请按此链接格式将相关图像内容附加输出，注意确保图片链接格式完整不被截断。若参考信息未提及图片链接则忽略此规则并注意不要随意捏造图片链接，在答案输出中不要体现此条指令信息的任何内容。\n"
                                    "2. **输出语言要求**：必须使用与问题相同的语言回答用户的问题。\n"
                            }
        num_tokens += calculate_multimodal_tokens([end_content_item])
        valid_search_list = []
        for i, item in enumerate(search_list):
            processed_snippet = curate_reference_text(item['snippet'])
            content_items = [
                {"type": "text", "text": f"\n【{i + 1}^】\n"},
                {"type": "text", "text": f"{processed_snippet}\n"}
            ]
            for rerank_i in item["rerank_info"]:
                if rerank_i["type"] == "image":
                    file_url = rerank_i['file_url']
                    if not file_url.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in file_url:
                        suffix = file_url.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
                        file_url = f"http://{MINIO_ADDRESS}/{suffix}"
                    content_items.append({"type": "text", "text": f"\n此 {rerank_i['file_url']} 的图片是:"})
                    content_items.append({"type": "image_url", "image_url": {"url":file_url}})
            # 计算上下文取舍
            num_tokens += calculate_multimodal_tokens(content_items)
            if num_tokens > available_tokens_for_context:
                break
            prompt_content.extend(content_items)
            valid_search_list.append(item)
        prompt_content.append({"type": "text", "text": "\n```\n"})  # 添加好参考信息分界
        prompt_content.append(end_content_item)
        prompt_content.append({
            "type": "text",
            "text": (
                "### 视觉与文本映射标准 SOP（必须严格执行）：\n"
                "若用户有上传图片并询问图中特定位置的功能，请按以下步骤思考并回答：\n\n"
                "1. **物理定位**：描述该图标在整体布局中的位置（如：底部工具栏左起第 N 个）。\n"
                "2. **示意图对齐**：在参考图中寻找相同位置，若该位置存在明确【编号 X】，锁定对应的【编号 X】, 继续执行步骤 3 和 4，若该位置不存在编号，立即停止，不得进行检索或功能推导\n"
                "3. **原文检索与摘抄**：在参考文本中查找 **【编号 X】** 后的文字。\n"
                "4. **功能推导**：基于摘抄的文字得出结论。**警告：严禁根据图标形状自行猜测，必须以文字定义为准。**\n\n"
            )
        })
        # ===== 多模态提示词构建完成
        messages = []
        for item in history:
            content_items =[
                {"role": "user", "content": item["query"]},
                {"role": "assistant", "content": item["response"]}
            ]
            num_tokens += calculate_multimodal_tokens(content_items)
            if num_tokens > available_tokens_for_context:
                break
            messages.extend(content_items)
        messages.append({"role": "user", "content": prompt_content})

        llm_data = {
            "model": model_name,
            "temperature": temperature,
            "top_p": top_p,
            "repetition_penalty": repetition_penalty,
            "do_sample": do_sample,
            "enable_thinking": enable_thinking,
            "stream": True,
            "messages": messages,
        }

        logger.info(f"llm_url: {llm_url}, 发送到大模型参数：{llm_data}, temperature:{temperature}")
        return send_request(llm_url, api_key, llm_data, valid_search_list)

    async def no_search_list(return_answer, history, question, code, msg, score, msg_id):
        answer = ''
        for char in return_answer:
            answer = answer + char

            history_tmp = history.copy()
            subjson = {}
            subjson["query"] = question
            subjson["response"] = answer
            subjson["needHistory"] = True
            history_tmp.append(subjson)

            response_info = {
                'code': code,
                "message": msg,
                "msg_id": msg_id,
                "data": {"output": char,
                         "searchList": [],

                         },
                "history": history_tmp,
                "finish": 0
            }
            if score != -1:  # 如果允许返回得分，返回空
                response_info["data"]["score"] = []
            jsonarr = json.dumps(response_info, ensure_ascii=False)
            str_out = f'{jsonarr}'
            yield str_out
        # ======= 最后返回 ========
        response_info = {
            'code': code,
            "message": msg,
            "msg_id": msg_id,
            "data": {"output": "",
                     "searchList": [],

                     },
            "history": history_tmp,
            "finish": 1
        }
        if score != -1:  # 如果允许返回得分，返回空
            response_info["data"]["score"] = []
        jsonarr = json.dumps(response_info, ensure_ascii=False)
        str_out = f'{jsonarr}'
        yield str_out

    response_info = {
        'code': int(0),
        "message": "success",
        "data": {"output": "",
                 "searchList": [],
                },
        "history":[]

    }


    json_request = await request.json()
    # user_id = request.headers.get("X-uid")
    # kb_name = json_request["knowledgeBase"]
    knowledge_base_info = json_request.get("knowledge_base_info", {})
    enable_vision = json_request.get("enable_vision", False)
    attachment_files = json_request.get("attachment_files", [])
    question = json_request["question"]
    rate = float(json_request["threshold"])
    top_k = int(json_request["topK"])
    stream = json_request["stream"]
    history = json_request["history"]
    chichat = json_request.get("chichat", True)
    default_answer = json_request.get("default_answer", '根据已知信息，无法回答您的问题。')
    return_meta = json_request.get("return_meta", False)
    prompt_template = json_request.get("prompt_template", '')
    top_p = json_request.get("top_p", 0.85)
    repetition_penalty = json_request.get("repetition_penalty", 1.1)
    temperature = json_request.get("temperature", TEMPERATURE)
    if temperature <= 0.01:  # 强制到0.01以下
        temperature = 0.01
    max_history = json_request.get("max_history", 10)
    custom_model_info = json_request.get("custom_model_info", {})
    search_field = json_request.get('search_field', 'con')

    if "do_sample" not in json_request:  # 如果没有传参，则默认使用temperature决定是否开启采样
        if temperature > 0.1:
            do_sample = True
        else:
            do_sample = False
    else:
        do_sample = json_request.get('do_sample')
    # 是否开启自动引文，此参数与prompt_template互斥，当开启auto_citation时，prompt_template用户传参不生效
    auto_citation = json_request.get("auto_citation", False)
    # 是否开启数据飞轮
    data_flywheel = json_request.get("data_flywheel", False)
    # 是否返回得分
    return_score = json_request.get("return_score", False)
    # 是否query改写
    rewrite_query = json_request.get("rewrite_query", False)
    rerank_mod = json_request.get("rerank_mod", "rerank_model")
    rerank_model_id = json_request.get("rerank_model_id", '')
    weights = json_request.get("weights", None)
    retrieve_method = json_request.get("retrieve_method", "hybrid_search")
    use_graph = json_request.get("use_graph", False)
    enable_thinking = json_request.get("enable_thinking", None)

    # metadata filtering params
    metadata_filtering = json_request.get("metadata_filtering", False)
    metadata_filtering_conditions = json_request.get("metadata_filtering_conditions", [])
    if not metadata_filtering:
        metadata_filtering_conditions = []

    filter_attachment_files = []
    for item in attachment_files:
        file_type = item.get("file_type")
        if file_type == "image":
            file_url = item["file_url"]
            parsed_url = urllib.parse.urlparse(file_url)
            # 校验 URL 是否包含协议头和网络位置
            if not all([parsed_url.scheme, parsed_url.netloc]):
                raise ValueError(f"Invalid attachment file URL: {file_url}")
            filter_attachment_files.append({"image": file_url})
        else:
            raise ValueError(f"attachment_file type {file_type} not support")
    if len(filter_attachment_files) > 1:
        raise ValueError("Multiple attachment files are not supported.")
    attachment_files = filter_attachment_files
    if not question or len(str(question).strip()) <= 0:
        raise ValueError("question cannot be empty")

    logger.info('---------------流式查询---------------')
    logger.info('knowledge_base_info:'+repr(knowledge_base_info)+'\t'+repr(json_request))

    def params_check_failed(err_msg: str):
        response_info = {
            'code': 1,
            "message": err_msg,
            "data": {"output": "", "searchList": []},
            "history": []
        }
        logger.error(error_msg)
        if json_request.get("stream"):
            return EventSourceResponse(no_search_list(default_answer, history, question, 1, error_msg, -1, ''))
        else:
            return JSONResponse(content=response_info)

    # 检查 knowledge_base_info 是否为空
    if not knowledge_base_info:
        error_msg = "knowledge_base_info cannot be empty"
        return params_check_failed(error_msg)
    # 检查 custom_model_info['llm_model_id'] 是否为空
    if 'llm_model_id' not in custom_model_info or not custom_model_info.get('llm_model_id'):
        error_msg = "custom_model_info['llm_model_id'] 不能为空"
        return params_check_failed(error_msg)

    # 检查 rerank_model_id 是否为空
    if rerank_mod == "rerank_model" and not rerank_model_id:
        error_msg = "rerank_model_id cannot be empty when using model-based reranking."
        return params_check_failed(error_msg)

    if rerank_mod == "weighted_score" and weights is None:
        error_msg = "weights cannot be empty when using weighted score reranking."
        return params_check_failed(error_msg)
    if weights is not None and not isinstance(weights, dict):
        error_msg = "weights must be a dictionary or None."
        return params_check_failed(error_msg)

    if rerank_mod == "weighted_score" and retrieve_method != "hybrid_search":
        error_msg = "Weighted score reranking is only supported in hybrid search mode."
        return params_check_failed(error_msg)

    sRandom = str(uuid.uuid1()).replace("-", "")
    u_id = "{}_{}_{}".format(knowledge_base_info, question, sRandom)
    msg_id = hashlib.md5(u_id.encode("utf8")).hexdigest()
    chunk_conent=0
    chunk_size=CHUNK_SIZE
    use_cache_flag = False

    if max_history > 0:
        history = history[-max_history:]
    else:
        history = []
    for user_id, kb_info_list in knowledge_base_info.items():
        kb_names = [kb_info['kb_name'] for kb_info in kb_info_list]
        kb_ids = [kb_info['kb_id'] for kb_info in kb_info_list]
        if rewrite_query:
            query_dict_list = get_query_dict_cache(redis_client,user_id, kb_ids)
            if query_dict_list:
                rewritten_queries = query_rewrite(question, query_dict_list)
                logger.info("对query进行改写,原问题:%s 改写后问题:%s" % (question, ",".join(rewritten_queries)))
                if len(rewritten_queries) > 0:
                    question = rewritten_queries[0]
                    logger.info("按新问题:%s 进行召回" % question)
            else:
                logger.info("未启用或维护转名词表,query未改写,按原问题:%s 进行召回" % question)
    if top_k<=0:
        # top_k必须大于0
        return EventSourceResponse(no_search_list(default_answer,history,question,1,'top_k必须大于0'))
    else:
        prompt=question
        search_list=[]
        has_effective_cache = False
        try:
            temp_start_time = time.time()
            if data_flywheel:
                # 要存储的数据
                cache_key = "%s^%s^%s" % (knowledge_base_info, top_k, question)
                exists = redis_client.exists(cache_key)
                if exists:
                    use_cache_flag = True
                    logger.info("=========>命中缓存,cache_key=%s" % cache_key)
                    cache_result = redis_client.get(cache_key)
                    # 将字符串转换为 JSON 对象
                    cache_result_json = json.loads(cache_result)
                    if cache_result_json and 'data' in cache_result_json:
                        if 'searchList' in cache_result_json['data'] and 'prompt' in cache_result_json['data'] and 'score' in cache_result_json['data']:
                            if len(cache_result_json["data"]["searchList"]) > 0:
                                has_effective_cache = True
                if has_effective_cache:
                    rerank_result = cache_result_json
                else:
                    rerank_result = get_knowledge_based_answer(knowledge_base_info, question, rate, top_k, chunk_conent,
                                                               chunk_size, return_meta, prompt_template, search_field,
                                                               default_answer, auto_citation, retrieve_method,
                                                               rerank_model_id=rerank_model_id, rerank_mod=rerank_mod,
                                                               weights=weights,
                                                               metadata_filtering_conditions=metadata_filtering_conditions,
                                                               use_graph=use_graph, enable_vision=enable_vision,
                                                               attachment_files=attachment_files,
                                                               )
            else:
                rerank_result = get_knowledge_based_answer(knowledge_base_info, question, rate, top_k, chunk_conent,
                                                           chunk_size, return_meta, prompt_template, search_field,
                                                           default_answer, auto_citation, retrieve_method,
                                                           rerank_model_id=rerank_model_id, rerank_mod=rerank_mod,
                                                           weights=weights,
                                                           metadata_filtering_conditions=metadata_filtering_conditions,
                                                           use_graph=use_graph, enable_vision=enable_vision,
                                                           attachment_files=attachment_files,
                                                           )

            logger.info("===>data_flywheel=%s,has_effective_cache=%s,rerank_result=%s" % (data_flywheel,has_effective_cache,json.dumps(rerank_result, ensure_ascii=False)))
            if rerank_result['code'] != 0:
                raise RuntimeError(f"get_knowledge_based_answer error, err: {rerank_result['message']}")

            response_info['code'] = int(rerank_result['code'])
            response_info['message'] = str(rerank_result['message'])
            response_info['msg_id'] = str(msg_id)

            search_list = rerank_result['data']['searchList']
            prompt = rerank_result['data']['prompt']
            score = rerank_result['data'].get('score', [])
            logger.info('知识召回结果：'+json.dumps(repr(rerank_result), ensure_ascii=False))
            temp_end_time = time.time()
            logger.info(f"======知识召回使用时间：{temp_end_time - temp_start_time}秒")
        except Exception as e:
            # logger.info('知识召回异常：'+repr(e))
            import traceback
            logger.error("====> 知识召回异常 error %s" % e)
            logger.error(traceback.format_exc())
            response_info['code']=1
            response_info['message']=repr(e)
            response_info['msg_id'] = str(msg_id)

            prompt=question
            search_list=[]
            score = []
        if not return_score:  # 如果不返回分数
            score = -1
        if SSE_USE_MONGO:  # 如果使用mongo
            temp_start_time = time.time()
            message = {"id": msg_id}
            current_date = datetime.now().strftime("%Y%m%d")
            try:
                u_condition = {'id': msg_id}
                message = {
                    "id": msg_id,
                    "user_id": "",
                    "kb_name": "",
                    "knowledge_base_info": knowledge_base_info,
                    "question": question,
                    "rate": rate,
                    "top_k": top_k,
                    "top_p": top_p,
                    "repetition_penalty": repetition_penalty,
                    "temperature": temperature,
                    "max_history": max_history,
                    "do_sample": do_sample,
                    "return_meta": "true" if return_meta else "false",
                    "auto_citation": "true" if auto_citation else "false",
                    "data_flywheel": "true" if data_flywheel else "false",
                    "return_score": "true" if return_score else "false",
                    "use_cache": "true" if has_effective_cache else "false",
                    "prompt_template": prompt_template,
                    "default_answer": default_answer,
                    "model_name": custom_model_info['llm_model_id'],
                    "search_field": search_field,
                    "search_list": search_list,
                    "scores": score if return_score else [],
                    "status": 0,
                    "update_time": int(round(time.time() * 1000)),
                    "create_time": int(round(time.time() * 1000)),
                    "create_dt": int(current_date)
                }
                # collection.insert_one(message)
                # collection.update_one(u_condition, message, upsert=True)
                collection.update_one(u_condition, {'$set': message}, upsert=True)
                if "_id" in message:
                    del message["_id"]
                logger.info("=======>user log已存储至mongoDB,id=%s,data=%s" % (msg_id, json.dumps(message, ensure_ascii=False)))
            except Exception as err:
                # 存储mongodb异常的时候，接口msg_id返回-1
                msg_id = "-1"
                import traceback
                logger.error("====> stream search save mongoDB error %s" % err)
                logger.error(traceback.format_exc())
        # if not use_cache_flag and data_flywheel:
        #     # 判断在飞轮模式下且若未命中缓存，推送kafka触发飞轮策略构建
        #     try:
        #         kafka_utils.push_kafka_msg(message)
        #         logger.info("=======>user log已推送kakfa")
        #     except Exception as err:
        #         import traceback
        #         logger.error("====> stream search push kafka error %s" % err)
        #         logger.error(traceback.format_exc())
        # 大模型生成返回
            temp_end_time = time.time()
            logger.info(f"======save mongoDB 使用时间：{temp_end_time - temp_start_time}秒")
        if stream:
            if response_info['code'] !=0:
                return EventSourceResponse(no_search_list(default_answer,history,question,response_info['code'],response_info['message'], score, msg_id))
            # 需要大模型输出
            if len(search_list)>0 or chichat:
                model_id = custom_model_info["llm_model_id"]
                llm_config = get_model_configure(model_id)
                if llm_config.is_multimodal:
                    gen = await multimodal_stream_generate(prompt, history, search_list,question,top_p,repetition_penalty,temperature,custom_model_info,do_sample,score,msg_id,attachment_files,llm_config,enable_thinking)
                    return EventSourceResponse(gen)
                else:
                    gen = await stream_generate(prompt, history, search_list,question,top_p,repetition_penalty,temperature,custom_model_info,do_sample,score,msg_id,llm_config,enable_thinking)
                    return EventSourceResponse(gen)
             # 知识召回为空，并且使用兜底话术返回，不需要大模型输出
            else:
                return EventSourceResponse(no_search_list(default_answer,history,question,0,'success', score, msg_id))

        else:  # 非stream返回
            # if response_info['code'] != 0:
            #     response_info = {
            #         'code': response_info['code'],
            #         "message": response_info['message'],
            #         "msg_id": msg_id,
            #         "data": {"output": default_answer,
            #                  "searchList": [],
            #                  },
            #         "history": history
            #     }
            #     if return_score:  # 如果允许返回得分，返回空
            #         response_info["data"]["score"] = []
            #     return JSONResponse(content=response_info)
            # # 需要大模型输出
            # if len(search_list) > 0 or chichat:
            #     response_info = generate(prompt, history, search_list, question, top_p, repetition_penalty, temperature, model_name,do_sample, score,msg_id)
            #     logger.info(f"=======>response_info:{response_info}")
            #     return JSONResponse(content=response_info)
            # # 知识召回为空，并且使用兜底话术返回，不需要大模型输出
            # else:
            #     response_info = {
            #         'code': 0,
            #         "message": "success",
            #         "msg_id": msg_id,
            #         "data": {"output": default_answer,
            #                  "searchList": [],
            #                  },
            #         "history": history
            #     }
            #     if return_score:  # 如果允许返回得分，返回空
            #         response_info["data"]["score"] = []
            #     return JSONResponse(content=response_info)
            response_info = {
                'code': 1,
                "message": "fail",
                "data": {"output": "parameter stream need to be true"}
            }
            return JSONResponse(content=response_info)

