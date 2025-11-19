import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from utils import mq_rel_utils
from kafka import KafkaConsumer
import json
import requests
import threading
from logging_config import setup_logging
from settings import KAFKA_BOOTSTRAP_SERVERS, KAFKA_SASL_PLAIN_USERNAME, KAFKA_SASL_PLAIN_PASSWORD
from minio import Minio

TEMP_URL_FILES_DIR = os.path.join(os.path.dirname(__name__), 'temp_url_files')
os.makedirs(TEMP_URL_FILES_DIR, exist_ok=True)

logger_name='url_batch_insert'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))

# Kafka配置 [EN] Kafka configuration
KAFKA_TOPIC = 'url-batch-i-prod'
GROUP_ID = 'my-group-dev2'

# Minio配置 [EN] Minio configuration
MINIO_URL = 'http://localhost:15000/upload'
MINIO_BUCKET_NAME = 'rag-doc'

    
def kafkal():
    while True:
        print('开始消费消息')
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
            

            task_id =  message_value["doc"]["task_id"]
            file_name =  message_value["doc"]["file_name"]
            

            kb_name =  message_value["doc"]["knowledgeBase"]
            user_id =  message_value["doc"]["user_id"]
            overlap_size =  message_value["doc"]["overlap_size"]
            sentence_size =  message_value["doc"]["sentence_size"]
            #chunk_type =  message_value["doc"]["chunk_type"]
            separators =  message_value["doc"]["separators"]
            is_enhanced = message_value["doc"]["is_enhanced"]
            try:
                lock = threading.Lock()
                # thread = threading.Thread(target=url_insert, args=(task_id,file_name,kb_name,user_id,overlap_size,sentence_size,separators,is_enhanced))
                thread = threading.Thread(target=url_insert, args=(task_id, file_name))
                lock.acquire()
                thread.start()
                lock.release()
                logger.info('----->kafka异步消费完成：user_id=%s,task_id=%s,process finished' % (user_id, task_id))
                # add_files(user_id,kb_name,file_name,sentence_size,overlap_size,object_name,task_id,separators,chunk_type,is_enhanced)
            except Exception as e:
                import traceback
                logger.error("====> kafkal error %s" % e)
                logger.error(traceback.format_exc())
                logger.error(repr(e))
                logger.error("kafka处理异常："+repr(e))
                continue
                

#解析出内容入库 [EN] Parse out the content and store it in the database

# def url_insert(task_id,kb_name,user_id,overlap_size,sentence_size,file_name,separators,is_enhanced):
def url_insert(task_id, file_name, **kwargs):
    try:
        name = task_id+'.txt'
        file_path = os.path.join(TEMP_URL_FILES_DIR, name)

        link = ''
        with open(file_path, "rb") as file:
            files_minio = {"file": file}
            data = {
                "file_name":task_id,
                "bucket_name":MINIO_BUCKET_NAME
            }
            response = requests.post(MINIO_URL, files=files_minio,data=data)
            if response.status_code == 200:
                
                link = response.json()["download_link"]
                logger.info('link：'+repr(link))
            else:
                logger.info(f"minio wrong") 

        response_data = {  
            "file_name": file_name,
            "download_link":link
        }
        logger.info('response_data:'+repr(response_data))
        json_str = json.dumps(response_data, ensure_ascii=False)
        mq_rel_utils.update_urlinsert_status(task_id, status=10 ) 
        return
    except Exception as e:
        logger.info(f"error: {str(e)}")
        mq_rel_utils.update_urlinsert_status(task_id, status=55 )
        return
    return 



if __name__ == "__main__":
    kafkal()

