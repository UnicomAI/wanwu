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


from utils.build_prompt import build_docqa_prompt_from_search_list

URL_RAG = os.getenv("URL_RAG")

config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

SENTENCE_SIZE = int(config["MODELS"]["DOC_CHUNK_SIZE"])
OVERLAP_SIZE = float(config["MODELS"]["DOC_OVERLAP_RATIO"])


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
    url = URL_RAG+':8681/rag/doc_parser'
    payload = json.dumps({
        "url": file_url,
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
        print('res:',response)
        result_dict = json.loads(response.text)
        print('result_dict:',result_dict)
        #result_dict.pop("docs")
        #print("result_dict :",{result_dict})
        #response.raise_for_status()
        docs = response.json().get("docs", [])
        print('docs',docs)
        return docs
    except Exception as e:
        print("parse_doc error:", {str(e)})
        return []




@app.route('/doc_pra', methods=['POST'])
def req_chat_doc():
    data = request.get_json()
    print("接收到请求数据：", data,flush=True)

    query = data.get("query")
    file_url = data.get("upload_file_url")
    sentence_size = SENTENCE_SIZE
    overlap_size = OVERLAP_SIZE
    # unified processing into a list
    file_urls = [file_url] if isinstance(file_url, str) else file_url
    all_docs = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
        # Create a parsing task
        print('走到解析任务这里',flush=True)
        future_to_url = {
            executor.submit(parse_doc, url, sentence_size, overlap_size): url
            for url in file_urls
        }
        # Collect analysis results
        for future in concurrent.futures.as_completed(future_to_url):
            print('开始解析',flush=True)
            url = future_to_url[future]
            try:
                docs = future.result()
                all_docs.extend(docs)
                print('all_docs:',all_docs,flush=True)
            except Exception as e:
                print("解析文档失败",{str(e)})

    if not all_docs:
        return jsonify({"error": "No document content parsed."}), 400


    # Build a document list
    doc_list = [
        {
            "snippet": doc.get("text"),
            "file_name": doc.get("metadata", {}).get("file_name"),
        } for doc in all_docs
    ]

    # Build prompt words
    prompt = build_docqa_prompt_from_search_list(query, doc_list)
    json_str = json.dumps({"prompt":prompt},ensure_ascii=False)

    return Response(json_str,content_type='application/json;charset=utf-8')





if __name__ == '__main__':
    app.run(host='0.0.0.0', port=1991)