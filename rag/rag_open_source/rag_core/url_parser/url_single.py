import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import json
import requests
import chardet
from flask import Flask, jsonify, request, make_response
# from knowledge_base_utils import *
from flask_cors import CORS
from bs4 import BeautifulSoup
from readability import Document
from playwright.sync_api import sync_playwright
from urllib.parse import unquote_plus 
import argparse
import re

from logging_config import setup_logging


TEMP_URL_FILES_DIR = os.path.join(os.path.dirname(__name__), 'temp_url_files')
os.makedirs(TEMP_URL_FILES_DIR, exist_ok=True)


logger_name='url_single'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))

app = Flask(__name__)
CORS(app, resources={r"/*": {"origins": "*"}})

app.config['JSON_AS_ASCII'] = False
app.config['JSONIFY_MIMETYPE'] ='application/json;charset=utf-8'

MINIO_URL = 'http://localhost:15000/upload'
MINIO_BUCKET_NAME = 'rag-doc'
CHROME_PATH = "/opt/chrome-linux/chrome"

def clean_text(text):
    """清除文本中的特殊字符和多余的空白，以及HTML标签。"""
    patterns = [
        r'\xa0+', r'\u3000', r'\t+', r'\r+', r'\n+',   # Clear special whitespace characters and multiline newlines
        r'<[/]?b>|&gt;|&lt;'                        # Clear HTML tags
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

def is_text_garbled(text):
    non_displayable_char_ratio = len(re.findall(r'[^\x20-\x7E\u4e00-\u9fff]', text)) / len(text)
    return non_displayable_char_ratio > 0.5    


def fetch_html_with_chromium(url: str, wait_until="networkidle"):
    with sync_playwright() as p:
        browser = p.chromium.launch(
            executable_path=CHROME_PATH,
            headless=True,
            args=["--no-sandbox", "--disable-dev-shm-usage"]
        )  # Start a headless browser
        page = browser.new_page()
        page.goto(url, wait_until=wait_until)
        html_text = page.content()
        browser.close()
        doc = Document(html_text)
        title = doc.title()
        summary_html = doc.summary()
        text = BeautifulSoup(summary_html, "lxml").get_text()
        return text.strip(), title


#parsing service
@app.route('/url_pra', methods=["POST","GET"])
def url_add():
    data = request.json
    url = data.get('url')
    logger.info(f"入参是: {data}")
    url = unquote_plus(url) 
    
    logger.info(f"The value of url is: {url}")
    if not url:
        return jsonify({'error': 'URL is required'}), 400
    try:
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.76 Safari/537.36'
        }

        text= ""
        title=""
        try:
            response = requests.get(url, headers=headers, timeout=10)
            response.raise_for_status()
            encoding = chardet.detect(response.content)['encoding']
            response.encoding = encoding if encoding else 'utf-8'# Set the encoding to ensure that Chinese display is normal
            soup = BeautifulSoup(response.content, 'html.parser')
            text = clean_text(soup.get_text())
            title = soup.find('title').get_text()
            logger.info(f"解析出的内容是: {text}")
        except Exception as e:
            logger.info(f"error: {str(e)}")
            logger.info(f"retry fetch url with chromium, url: {url}")

        if len(text) < 30 or is_text_garbled(text):
            text, title = fetch_html_with_chromium(url)
            logger.info(f"解析出的内容是: {text}")

        #There is a problem with parsing, please return here
        if len(text) < 30:
            response_data = {  
                "file_name": '',
                "old_name":url,# Add original name and file name to response data
                "response_info": {
                    "code": 1,
                    "message": "该网站不支持抓取解析"                
                }   
            }
            logger.info(f"short: {url}")
            json_str = json.dumps(response_data, ensure_ascii=False)
            response = make_response(json_str) 
            return response
        if is_text_garbled(text):
            response_data = {
                "file_name": '',
                "old_name":url,# Add original name and file name to response data
                "response_info": {
                    "code": 1,
                    "message": "该网站不支持抓取解析"
                }
            }
            logger.info(f"content_garbled: {url}")
            json_str = json.dumps(response_data, ensure_ascii=False)
            response = make_response(json_str)
            return response

        title = title.replace('|', '_').replace(':', '_').replace(' ', '_')
        if len(title) >= 50:
            title = title[:30]
        title = title if title else '无标题'
        logger.info(f"标题是: {title}")

        title_name = title+'.txt'
        logger.info(f"解析文件名是: {title_name}")
        file_path = os.path.join(TEMP_URL_FILES_DIR, title_name)
        with open(file_path, 'w', encoding='utf-8') as file:
            file.write(text)
        file_size = os.path.getsize(file_path)
        file_size_kb = file_size / 1024
        response_data = {  
            "file_name": title,
            "old_name":url,# Add original name and file name to response data
            "file_size":file_size_kb,
            "response_info": {
                "code": 0,
                "message": "解析成功"                
            }  
        }
        
        json_str = json.dumps(response_data, ensure_ascii=False)
        response = make_response(json_str)       
    except Exception as e:
        logger.info(f"error: {str(e)}")
        if "HTTPSConnectionPoolstr" in str(e):
            response_data = {  
                "file_name": '',
                "old_name":url,# Add original name and file name to response data
                "response_info": {
                    "code": 1,
                    "message": "该网站不支持抓取解析"                
                }  
            }
        
        json_str = json.dumps(response_data, ensure_ascii=False)
        response = make_response(json_str)
    return response,200


#Parse out the content and store it in the database
@app.route('/url_insert', methods=["POST","GET"])
def url_insert_data():
    logger.info('进入入库')
    data = request.json
    file_name = data.get('file_name')
    # overlap_size = data.get('overlap_size',0.0)
    # sentence_size = data.get('sentence_size', 300)    
    # chunk_type = data.get('chunk_type', 'split_by_default') 
    # user_id = data.get("userId")
    # kb_name = data.get("knowledgeBase")
    # is_enhanced = data.get("is_enhanced", 'false')
    # separators = data.get("separators", ['。'])
    task_id = data.get("task_id")
    try:
        name = file_name+'.txt'
        old_file_path = os.path.join(TEMP_URL_FILES_DIR, name)
        new_file_path = os.path.join(TEMP_URL_FILES_DIR, task_id+'.txt')
        os.rename(old_file_path, new_file_path)       
        link = ''
        try:
            with open(new_file_path, "rb") as file:
                files_minio = {"file": file}
                data = {
                    "file_name":task_id,
                    "bucket_name":MINIO_BUCKET_NAME,
                }
                response = requests.post(MINIO_URL, files=files_minio,data=data)
                if response.status_code == 200:
                    
                    link = response.json()["download_link"]
                    logger.info(f"minio sucess: {link}")
                else:
                    logger.info(f"minio wrong")


            response_data = {  
                "file_name": task_id + '.txt',
                "download_link":link,
                "response_info":{
                "code":0,
                "message":"入库成功"}
            }
            logger.info(f"response_data: {response_data}")
            json_str = json.dumps(response_data, ensure_ascii=False)
            response = make_response(json_str)
        except Exception as e:
            import traceback
            logger.error("====> split_chunks error %s" % e)
            logger.error(traceback.format_exc())
            logger.error(repr(e))
            
            
    except Exception as e:
        logger.info(f"error: {str(e)}")
        return jsonify({'error': str(e)}), 500
    logger.info(f"insert sucess: {response}")
    return response,200




if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int)
    args = parser.parse_args()
    app.run(host='0.0.0.0', port=args.port)

