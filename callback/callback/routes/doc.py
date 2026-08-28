import json
import logging

from flask import Response, jsonify, request

from callback.services import doc as doc_service
from configs.config import config
from utils.response import BizError, response_ok

from . import callback_bp


@callback_bp.route("/doc_pra", methods=["POST"])
def req_chat_doc():
    """
    解析文档并生成 Prompt
    ---
    description: |
      接收用户 query 和文档 URL，对文档进行切块解析并根据 RAG 生成 Prompt。
    tags:
      - doc
    requestBody:
      description: 请求参数
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - query
              - upload_file_url
            properties:
              query:
                type: string
                description: 用户问题 / 主题内容
                example: "请对文档进行总结"
              upload_file_url:
                type: string
                description: 已上传文档的下载 URL
                example: "https://example.com/upload/file.pdf"
    responses:
      200:
        description: 生成并上传成功
        content:
          application/json:
            schema:
              type: object
              properties:
                prompt:
                  type: string
                  description: 根据文档内容生成的 Prompt
    """
    data = request.get_json()

    query = data.get("query")
    file_url = data.get("upload_file_url")

    prompt = doc_service.process_documents(query, file_url)

    return jsonify({"prompt": prompt})


@callback_bp.route("/doc_parse", methods=["POST"])
def req_parse_doc():
    """
    解析文档返回内容
    ---
    description: |
      接收文档 URL，对文档进行解析并返回完整文档内容，不进行切分。
    tags:
      - doc
    requestBody:
      description: 请求参数
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - upload_file_url
            properties:
              upload_file_url:
                type: string
                description: 已上传文档的下载 URL
                example: "https://example.com/upload/file.pdf"
              max_token:
                type: int
                description: 返回文档的最大token数
                example: 200000
    responses:
      200:
        description: 解析成功
        content:
          application/json:
            schema:
              type: object
              properties:
                code:
                  type: integer
                  description: 状态码
                msg:
                  type: string
                  description: 响应提示信息，例如 "解析成功"
                data:
                  type: string
                  description: 解析后的完整文档内容
    """
    data = request.get_json()

    file_url = data.get("upload_file_url")
    max_token = data.get("max_token", 0)

    if not file_url:
        raise BizError("upload_file_url is required")

    content = doc_service.parse_doc_only(file_url, max_token)

    return response_ok(content)


@callback_bp.route("/generate_file", methods=["POST"])
def generate_file_to_minio():
    """
    将 Markdown 内容生成为指定格式文件并获取下载链接
    ---
    tags:
      - doc
    requestBody:
      required: true
      content:
        multipart/form-data:
          schema:
            type: object
            required:
              - formatted_markdown
              - to_format
            properties:
              formatted_markdown:
                type: string
                description: 需要转换的 Markdown 文本内容
                example: "# Hello World\nThis is a test document."
              to_format:
                type: string
                enum:
                  - docx
                  - pdf
                  - txt
                  - md
                  - html
                description: 目标文件格式
              title:
                type: string
                description: 生成文件的标题(不包含后缀)
                default: "Untitled"
    responses:
      200:
        description: 生成并上传成功
        content:
          application/json:
            schema:
              type: object
              properties:
                download_link:
                  type: string
                  description: 生成文件的 MinIO 下载链接
                  example: "http://minio-url/bucket/my-document.docx"
      400:
        description: 业务逻辑错误
        content:
          application/json:
            schema:
              type: object
              properties:
                code:
                  type: integer
                  description: 错误码
                  example: 200000
                msg:
                  type: string
                  description: 错误描述信息
      500:
        description: 服务内部错误
        content:
          application/json:
            schema:
              type: object
              properties:
                code:
                  type: integer
                  description: 错误码
                  example: 200000
                msg:
                  type: string
                  description: 错误描述信息
    """

    formatted_markdown = request.form.get("formatted_markdown")
    to_format = request.form.get("to_format")
    filename = request.form.get("title", "Untitled")

    # 参数校验
    if not formatted_markdown:
        raise BizError("formatted_markdown cannot be empty")
    if not to_format:
        raise BizError("to_format cannot be empty")

    _, _, download_link = doc_service.generate_file_to_minio(
        formatted_markdown, filename, to_format
    )

    return jsonify({"download_link": download_link})
