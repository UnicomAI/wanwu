import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from utils import mq_rel_utils
from kafka import KafkaConsumer
import json
import chardet
import requests
import threading
from logging_config import setup_logging
from settings import KAFKA_BOOTSTRAP_SERVERS, KAFKA_SASL_PLAIN_USERNAME, KAFKA_SASL_PLAIN_PASSWORD

from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
import time
from urllib.parse import unquote_plus 
import re
from selenium.webdriver.chrome.options import Options
from bs4 import BeautifulSoup

CHROME_DIR = os.path.abspath('/opt')

TEMP_URL_FILES_DIR = os.path.join(os.path.dirname(__name__), 'temp_url_files')
os.makedirs(TEMP_URL_FILES_DIR, exist_ok=True)


logger_name='url_batch_parse'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))

# Kafka配置 [EN] Kafka configuration
KAFKA_TOPIC = 'url-batch-a-prod'
GROUP_ID = 'my-group-dev2'


def md5(string):
    """
    返回md5签名
    """
    import hashlib
    string_bytes = string.encode('utf-8') 
    return hashlib.md5(string_bytes).hexdigest()


def clean_text(text):
    """清除文本中的特殊字符和多余的空白，以及HTML标签。"""
    patterns = [
        r'\xa0+', r'\u3000', r'\t+', r'\r+', r'\n+',   # 清除特殊空白字符和多行换行符 [EN] Clear special whitespace characters and multiline newlines
        r'<[/]?b>|&gt;|&lt;'                        # 清除HTML标签 [EN] Clear HTML tags
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

def is_text_garbled(text):
    non_displayable_char_ratio = len(re.findall(r'[^\x20-\x7E\u4e00-\u9fff]', text)) / len(text)
    return non_displayable_char_ratio > 0.5



def kafkal():
    while True:

        consumer = KafkaConsumer(KAFKA_TOPIC,
                                 bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                 security_protocol='SASL_PLAINTEXT',
                                 sasl_mechanism='PLAIN',
                                 sasl_plain_username=KAFKA_SASL_PLAIN_USERNAME,
                                 sasl_plain_password=KAFKA_SASL_PLAIN_PASSWORD,
                                 group_id=GROUP_ID,
                                 enable_auto_commit=False,
                                 max_poll_records=1,
                                 value_deserializer=lambda x:x.decode('utf-8'))

        for message in consumer:
            #初始化用户知识库路径 [EN] Initialize user knowledge base path
            logger.info('收到kafka消息：'+repr(message.value))
            message_value = json.loads(message.value)
            
            url = message_value["doc"]["url"]
            task_id = message_value["doc"]["task_id"]

            url = unquote_plus(url)

            try:
                lock = threading.Lock()
                #mq_rel_utils.update_url_status(task_id, status=10 ) 
                thread = threading.Thread(target=url_ana, args=(url,task_id))
                lock.acquire()
                thread.start()
                lock.release()
                logger.info('----->kafka异步消费完成')
                # add_files(user_id,kb_name,filename,sentence_size,overlap_size,object_name,task_id,separators,chunk_type,is_enhanced)
            except Exception as e:
                import traceback
                logger.error(traceback.format_exc())
                logger.error("kafka处理异常："+repr(e))
                continue
                

def url_ana(url, task_id):
    a = ''
    url = unquote_plus(url) 
    
    logger.info(f"The value of url is: {url}")

    try:
        headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.76 Safari/537.36'
        }
        response = requests.get(url, headers=headers, timeout=10)
        response.raise_for_status()
        encoding = chardet.detect(response.content)['encoding']
        response.encoding = encoding if encoding else 'utf-8'# 设置编码，确保中文显示正常 [EN] Set the encoding to ensure that Chinese display is normal
        soup = BeautifulSoup(response.content, 'html.parser')
        a = clean_text(soup.get_text())
        logger.info(f"正常解析出的内容是: {a}")
        b = ''
        title_tag = soup.find('title')
        logger.info("title is:"+title_tag.text)
        #logger.info(f"原来标题是: {title_tag.text}") [EN] logger.info(f"The original title is: {title_tag.text}")
        c = title_tag.text
        b = c.replace('|', '_').replace(':', '_').replace(' ', '_')
        
        if len(b) >= 50:
            b = b[:30]
        else:
            b = b
        title_text = b if title_tag else '无标题'
        logger.info("处理完的标题是："+title_text)

        name = task_id+'.txt'
        file_path = os.path.join(TEMP_URL_FILES_DIR, name)
        with open(file_path, 'w', encoding='utf-8') as file:
            file.write(a)
        with open(file_path, 'r', encoding='utf-8') as file:  
            content = file.read() 
        logger.info(f"写入结束: {content}")
        file_size = os.path.getsize(file_path)
        file_size_kb = file_size / 1024
        #json_str = json.dumps(response_data, ensure_ascii=False)
        mq_rel_utils.update_url_status(task_id, status=10,fileSize=file_size_kb,fileName=title_text)
        return
    except Exception as e:
        logger.info(f"error: {str(e)}")
        logger.info(f"未解析出的url: {url}")
        
        response_data = {  
            "file_name": '',
            "old_name":url,# 添加原始name和文件名到响应数据中 [EN] Add original name and file name to response data  
            "response_info": {
                "code": 1,
                "message": "该网站不支持抓取解析"                
            }  
        }
        
        json_str = json.dumps(response_data, ensure_ascii=False)
        #response = make_response(json_str)
        mq_rel_utils.update_url_status(task_id, status=57,fileSize=0,fileName='')
        return
    #return response,200
    return


if __name__ == "__main__":
    kafkal()

