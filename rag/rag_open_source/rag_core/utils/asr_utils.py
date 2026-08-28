#!/usr/bin/env python
# coding=utf-8


import json
import os
import uuid
import requests
import subprocess
import multiprocessing
import shutil
import logging

from pathlib import Path
from configs.config_parser import Config
from model_manager.model_config import get_model_configure, LlmModelConfig
from langchain.text_splitter import CharacterTextSplitter
from utils.constant import MAX_SENTENCE_SIZE
from utils import minio_utils
import time

from typing import List, Union, Any, Dict
import re
logger = logging.getLogger(__name__)
from settings import MINIO_ADDRESS,REPLACE_MINIO_DOWNLOAD_URL



def parse_error_to_dict(error) -> Dict[str, Any]:
    """将错误信息转换为字典类型"""
    try:
        # 从错误信息中提取 JSON 部分
        error_str = str(error)
        # 使用正则表达式匹配 '-' 后面的 JSON 字符串
        match = re.search(r'-\s*(\{.*\})', error_str)
        if match:
            json_str = match.group(1)
            return json.loads(json_str)
        # 如果没有匹配到 JSON 格式，返回基本错误信息
        return {
            "error": {
                "message": str(error),
                "type": type(error).__name__,
                "code": getattr(error, 'code', 'unknown')
            }
        }
    except Exception as e:
        # 确保总是返回一个有效的错误字典
        return {
            "error": {
                "message": str(error),
                "parse_error": str(e),
                "type": "error_parse_failed"
            }
        }


def extract_pcm_audio_from_video(video_path, audio_path):
    """
    使用 FFmpeg 提取音频为 16kHz 单声道 PCM 格式。
    """

    # 检查视频文件是否存在
    # if not video_path.exists():
    #     raise FileNotFoundError(f"视频文件不存在: {video_path}")
    logger.info(f"video_path:{video_path}")
    # 检查并获取FFmpeg路径
    ffmpeg_path = shutil.which("ffmpeg")
    logger.info(f"ffmpeg_path:{ffmpeg_path}")

    # 创建输出目录
    # audio_path.parent.mkdir(parents=True, exist_ok=True)
    # 构建命令
    command = [
        ffmpeg_path,
        "-y",  # 覆盖已存在文件
        "-i", str(video_path),
        "-vn",  # 无视频
        "-acodec", "pcm_s16le",  # 音频编码
        "-ar", "16000",  # 采样率
        "-ac", "1",  # 单声道
        str(audio_path)
    ]

    logger.info(f"执行命令: {' '.join(command)}")

    try:
        # 执行命令并捕获错误输出
        result = subprocess.run(
            command,
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        # logger.info(f"音频提取完成: {result}")
        return True
    except subprocess.CalledProcessError as e:
        stderr = e.stderr or ""
        if "does not contain any stream" in stderr:
            # 视频无音轨，属于正常情况，不打印整段 ffmpeg stderr
            logger.info("视频无音轨，跳过 ASR 音频提取")
        else:
            logger.info(f"FFmpeg执行失败: {stderr}")
        return False
    except Exception as e:
        logger.info(f"提取音频时发生未知错误: {e}")
        raise
        return False


def req_unicom_asr(wav_url: str,
                   asr_model_id: str):
    retries = 0
    max_retries = 3

    llm_config = get_model_configure(asr_model_id)
    wanwu_asr_url = llm_config.endpoint_url + "/asr"
    logger.info("=========>wanwu_asr_url:%s,model_name:%s,provider:%s,model_type:%s" %
                (wanwu_asr_url, llm_config.model_name, llm_config.provider, llm_config.model_type))
    headers = {
        "Content-type": "application/json;charset=utf-8"
    }
    # 本版本支持同步接口调用模式
    while retries < max_retries:
        try:

            payload = {
                "model": llm_config.model_name,
                "messages": [
                    {"content": [{"type": "audio", "audio": {"data": wav_url}}
                                 ],
                     "role": "user",
                     "extra": {}

                     }
                ]
            }
            response = requests.post(wanwu_asr_url, headers=headers, json=payload, timeout=60)
            # response = requests.post(wanwu_asr_url, headers=headers, data=json.dumps(payload, ensure_ascii=False).encode('utf-8'), timeout=60)
            logger.info("====>params:%s" % json.dumps(payload, ensure_ascii=False))
            response.raise_for_status()
            ret_json = response.json()
            logger.debug("ASR raw response: %s", json.dumps(ret_json, ensure_ascii=False))
            if "choices" in ret_json and ret_json["choices"]:
                content = ret_json["choices"][0]["message"].get("content", [])
                if isinstance(content, list):
                    return content
                else:
                    # 如果 content 是字符串，包装成标准格式
                    return [{"text": content}]

        except Exception as e:
            import traceback
            logger.error("=======>req_unicom_asr error %s" % e)
            logger.error(traceback.format_exc())
            error_dict = parse_error_to_dict(e)
            logger.error(f"\n意外错误: {json.dumps(error_dict, ensure_ascii=False)}")
            retries += 1
            time.sleep(1)
    return []


def asr_parser_text(file_path, asr_model_id):
    """
    调用音频撰写服务,解析生成文本
    :param file_path:
    """
    text = ""
    llm_config = get_model_configure(asr_model_id)
    session_id = str(uuid.uuid4())
    file_suffix = Path(file_path).suffix.lower()
    #####判断是否为音频文件
    audio_exts = {".wav", ".mp3", ".aac", ".flac", ".m4a", ".ogg", ".wma"}
    is_audio = file_suffix in audio_exts
    ####判断是否为视频文件
    video_exts = {".mp4", ".mov", ".avi"}
    is_video = file_suffix in video_exts
    # 获取文件所在的目录路径
    directory = os.path.dirname(file_path)
    temp_audio_path = os.path.join(directory, f"audio_{session_id}.wav")

    if is_video:
        # 视频文件先提取音轨
        if extract_pcm_audio_from_video(file_path, temp_audio_path):
            # 若能提取出音频流则仅对音频流调用asr解析
            new_file_path = temp_audio_path
        else:
            # 若视频文件未提取到任何音频，直接返回
            return text
    elif is_audio:
        # 音频文件默认还是按原文件调用asr解析
        new_file_path = file_path
    else:
        # 非音视频文件按原文件调用asr解析
        new_file_path = file_path
    minio_result = minio_utils.upload_local_file(new_file_path)
    if minio_result['code'] == 0:
        wav_url = minio_result['download_link']
        if not wav_url.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in wav_url:
            suffix = wav_url.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
            wav_url = f"http://{MINIO_ADDRESS}/{suffix}"

    results = req_unicom_asr(wav_url, asr_model_id)

    texts = [
        item["text"]
        for item in results
        if isinstance(item, dict) and item.get("text")
    ]
    text = "\n".join(texts) + ("\n" if texts else "")
    logger.info("=======>asr_parser_text,text=%s" % text)
    return text


def asr_parser_chunk(file_path, asr_model_id):
    """
    调用音频撰写服务，解析生成切分chunks
    :param file_path:
    """

    page_chunks = []
    text = ""
    session_id = str(uuid.uuid4())
    file_suffix = Path(file_path).suffix.lower()
    #####判断是否为音频文件
    audio_exts = {".wav", ".mp3", ".aac", ".flac", ".m4a", ".ogg", ".wma"}

    is_audio = file_suffix in audio_exts
    ####判断是否为视频文件
    video_exts = {".mp4", ".mov", ".avi"}
    is_video = file_suffix in video_exts
    # 获取文件所在的目录路径
    directory = os.path.dirname(file_path)
    temp_audio_path = os.path.join(directory, f"audio_{session_id}.wav")

    if is_video:
        # 视频文件先提取音轨
        if extract_pcm_audio_from_video(file_path, temp_audio_path):
            # 若能提取出音频流则仅对音频流调用asr解析
            new_file_path = temp_audio_path
        else:
            # 若视频文件未提取到任何音频，直接返回
            return page_chunks
    elif is_audio:
        # 音频文件默认还是按原文件调用asr解析
        new_file_path = file_path
    else:
        # 非音视频文件按原文件调用asr解析
        new_file_path = file_path

    minio_result = minio_utils.upload_local_file(new_file_path)

    if minio_result['code'] == 0:
        wav_url = minio_result['download_link']
        if not wav_url.startswith(f"http://{MINIO_ADDRESS}") and REPLACE_MINIO_DOWNLOAD_URL in wav_url:
            suffix = wav_url.replace(REPLACE_MINIO_DOWNLOAD_URL, "").lstrip("/")
            wav_url = f"http://{MINIO_ADDRESS}/{suffix}"

    results = req_unicom_asr(wav_url, asr_model_id)

    texts = [
        item["text"]
        for item in results
        if isinstance(item, dict) and item.get("text")
    ]
    text = "\n".join(texts) + ("\n" if texts else "")
    logger.info("=======>asr_parser_chunk,text=%s" % text)

    if len(text) < MAX_SENTENCE_SIZE:
        page_chunk = {}
        page_chunk["text"] = text
        # page_chunk["page_num"] = [-1]
        page_chunk["file_path"] = file_path
        page_chunk["length"] = len(text)
        page_chunk["type"] = "text"
        page_chunks.append(page_chunk)
    else:
        # 使用langchain.text_splitter进行文本切分
        chunk_size = MAX_SENTENCE_SIZE - 1
        chunk_overlap = int(chunk_size * 0.05)

        text_splitter = CharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            separator="\n"
        )
        # 分割文本
        chunks = text_splitter.split_text(text)
        # 为每个chunk创建page_chunk，并加入page_chunks
        for chunk in chunks:
            page_chunk = {}
            page_chunk["text"] = chunk
            # page_chunk["page_num"] = [-1]
            page_chunk["file_path"] = file_path
            page_chunk["type"] = "text"
            page_chunk["length"] = len(chunk)
            page_chunks.append(page_chunk)
    return page_chunks

if __name__ == "__main__":
    add_file_path = "./test.mp4"  # 替换为你的 PDF 文件路径
    chunks = asr_parser_chunk(add_file_path)
    print("chunks=%s" % json.dumps(chunks, ensure_ascii=False))