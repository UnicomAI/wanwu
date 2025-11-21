from flask import Flask, request, jsonify
from flask import Response, stream_with_context
import asyncio
import configparser
import os
from openai import OpenAI
from datetime import datetime
import json
import requests
import concurrent.futures
import logging

URL_RAG = os.getenv("URL_RAG")
#from utils.build_prompt import build_docqa_prompt_from_search_list


config = configparser.ConfigParser()
config.read('/agent/agent_open_source/config.ini',encoding='utf-8')

SENTENCE_SIZE = 500
OVERLAP_SIZE = 0.0


app = Flask(__name__)




def parse_doc(file_url, sentence_size, overlap_size):
    """
    解析单个文档
    
    参数:
    file_url (str): 文件URL
    sentence_size (int): 句子大小
    overlap_size (float): 重叠比例
    user_token (str, optional): 用户token
    
    返回:
    list: 解析后的文档片段列表
    """
    
    #url = config["MODELS"]["default_doc_parser_url"]
    url = URL_RAG + ':8681/rag/doc_parser'
    payload = json.dumps({
        "url": file_url,
        "parser_choices":['text'],
        "sentence_size": sentence_size,
        "overlap_size": overlap_size,
        "separators":[
            "\n\n", "\n", " ", ",",
            "\u200b",  # zero width space
            "\uff0c",  # Full-width comma
            "\u3001",  # comma
            "\uff0e",  # full-width period
            "\u3002",  # period
            ".", "",
        ]
    })

    headers = {
        "Content_Type": "application/json;charset=utf-8"
    }

    try:
        response = requests.post(url, headers=headers, data=payload, verify=False)
        result_dict = json.loads(response.text)
        #result_dict.pop("docs")
        #print("req_chat_doc status :",{result_dict},flush=True)
        #response.raise_for_status()
        #docs = response.json().get("docs", [])
        docs = response.json().get("docs", [])
        print('docs',docs)
        return docs
    except Exception as e:
        print("parse_doc error:", {str(e)})
        return []



        
@app.route('/doc_pra', methods=['POST'])
def req_chat_doc():
    data = request.get_json()
    print("接收到请求数据：",data,flush=True)

    logging.info(f"接收到请求数据：{data}")
    file_url = data.get("upload_file_url")
    sentence_size = SENTENCE_SIZE
    overlap_size = OVERLAP_SIZE
    # unified processing into a list
    file_urls = [file_url] if isinstance(file_url, str) else file_url
    all_docs = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
        # Create a parsing task
        future_to_url = {
            executor.submit(parse_doc, url, sentence_size, overlap_size): url 
            for url in file_urls
        }
        # Collect analysis results
        for future in concurrent.futures.as_completed(future_to_url):
            url = future_to_url[future]
            try:
                docs = future.result()
                all_docs.extend(docs)
            except Exception as e:
                print("解析文档失败",{str(e)})
    
    if not all_docs:
        return jsonify([])
    
    # Build a document list
    doc_list = [
        {
            "snippet": doc.get("text"),
            "file_name": doc.get("metadata", {}).get("file_name"),
        } for doc in all_docs
    ]
    return all_docs




if __name__ == '__main__':
    app.run(host='0.0.0.0', port=15003)
