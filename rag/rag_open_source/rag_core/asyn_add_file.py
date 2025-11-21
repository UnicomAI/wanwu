import os
import nltk
# Set NLTK data path
# Get the absolute path of the current file
current_file_path = os.path.abspath(__file__)
# Get the directory where the current file is located
current_dir = os.path.dirname(current_file_path)
# Splice the path to the nltk_data folder
nltk_data_path = os.path.join(current_dir, 'nltk_data')
nltk.data.path.append(nltk_data_path)
nltk.data.path.append("/opt/nltk_data")
from utils import milvus_utils
from utils import minio_utils
from utils import es_utils
from utils import file_utils
from utils import mq_rel_utils
from utils import knowledge_base_utils
from utils.file_utils import SplitConfig
from utils import schema_utils
from utils import redis_utils
import subprocess
from kafka import KafkaConsumer, TopicPartition, OffsetAndMetadata
import json
import threading
from logging_config import setup_logging
from datetime import datetime
import re
from settings import *
from utils.constant import CONVERT_DIR, USER_DATA_PATH
graph_redis_client = redis_utils.get_redis_connection()

# define path
paths = ["./data", "./user_data"]
# Traverse a list of paths
for path in paths:
    # Check if the path exists
    if not os.path.exists(path):
        # Create the directory if it does not exist
        os.makedirs(path)
        print(f"目录 {path} 已创建。")
    else:
        print(f"目录 {path} 已存在。")


logger_name = 'rag_asyn_add_files_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

master_control_logger_name = 'mc_rag_asyn_add_files_utils'
master_control_app_name = os.getenv("LOG_FILE") + "_master_control"
master_control_logger = setup_logging(master_control_app_name, master_control_logger_name)
master_control_logger.info(logger_name + '---------LOG_FILE：' + repr(master_control_app_name))

CONVERT_OFFICE_FORMAT_MAP = {".doc": "docx", ".wps": "docx", ".xls": "xlsx", ".ppt": "pptx", ".ofd": "pdf"}

def kafkal():
    while True:
        print('开始消费消息')
        if KAFKA_SASL_USE:
            consumer = KafkaConsumer(KAFKA_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     security_protocol='SASL_PLAINTEXT',
                                     sasl_mechanism='PLAIN',
                                     sasl_plain_username=KAFKA_SASL_PLAIN_USERNAME,
                                     sasl_plain_password=KAFKA_SASL_PLAIN_PASSWORD,
                                     group_id=KAFKA_GROUP_ID,
                                     enable_auto_commit=KAFKA_ENABLE_AUTO_COMMIT,
                                     max_poll_records=1,  # Set to pull at most 1 message at a time
                                     # max_poll_interval_ms=8000000, # Set the maximum polling interval to 120 minutes
                                     value_deserializer=lambda x: x.decode('utf-8'))

        else:
            consumer = KafkaConsumer(KAFKA_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     group_id=KAFKA_GROUP_ID,
                                     enable_auto_commit=KAFKA_ENABLE_AUTO_COMMIT,
                                     max_poll_records=1,  # Set to pull at most 1 message at a time
                                     # max_poll_interval_ms=8000000, # Set the maximum polling interval to 120 minutes
                                     value_deserializer=lambda x: x.decode('utf-8'))
        for message in consumer:
            # Initialize user knowledge base path
            print('收到新kafka消息：' + repr(message.value))
            logger.info('收到新kafka消息：' + repr(message.value))
            master_control_logger.info('收到新kafka消息：' + repr(message.value))
            message_value = json.loads(message.value)
            if "ocr_model_id" not in message_value["doc"]:
                logger.error("no ocr_model_id")
                continue
            file_id = message_value["doc"]["id"]
            kb_name = message_value["doc"]["categoryId"]
            user_id = message_value["doc"]["userId"]
            kb_id = message_value["doc"].get("kb_id", "")
            overlap_size = message_value["doc"]["overlap"]
            object_name = message_value["doc"]["objectName"]
            sentence_size = message_value["doc"]["chunk_size"]
            filename = message_value["doc"]["originalName"]
            split_type = message_value["doc"].get("split_type", "common")
            chunk_type = message_value["doc"].get("chunk_type", "default")
            separators = message_value["doc"].get("separators", ['。'])
            is_enhanced = message_value["doc"].get("is_enhanced", 'false')
            enable_knowledge_graph = message_value["doc"].get("enable_knowledge_graph", "false")

            # When importing files, select the parsing method. Text extraction is checked by default. Optical recognition OCR is optional. When multiple selections are made, this parameter defaults to ["text"]. When OCR is checked, pass: ["text", "ocr"]
            parser_choices = message_value["doc"]["parser_choices"] if "parser_choices" in message_value["doc"] else [
                "text"]
            ocr_model_id = message_value["doc"]["ocr_model_id"] if "ocr_model_id" in message_value["doc"] else [
                ""]
            pre_process = message_value["doc"]["pre_process"] if "pre_process" in message_value["doc"] else []
            meta_data_rules = message_value["doc"]["meta_data"] if "meta_data" in message_value["doc"] else []
            child_chunk_config = message_value["doc"]["child_chunk_config"] if "child_chunk_config" in message_value["doc"] else None

            split_config = SplitConfig(
                sentence_size=sentence_size,
                overlap_size=overlap_size,
                chunk_type=chunk_type,
                separators=separators,
                parser_choices=parser_choices,
                ocr_model_id=ocr_model_id,
                split_type=split_type,
                child_chunk_config=child_chunk_config
            )

            try:
                if not KAFKA_ENABLE_AUTO_COMMIT:
                    # Commit the offset of the current message
                    tp = TopicPartition(KAFKA_TOPICS, message.partition)
                    offset_and_metadata = OffsetAndMetadata(offset=message.offset + 1, metadata="")
                    offsets = {tp: offset_and_metadata}
                    # consumer.commit(offsets=offsets) # No need to use this commit
                    consumer.commit()
                    logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                    master_control_logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                    logger.info('consumer.commit offset：' + repr(offsets))
                    master_control_logger.info('consumer.commit offset：' + repr(offsets))

                if KAFKA_USE_ASYN_ADD:
                    # ============ Asynchronous addition =============
                    lock = threading.Lock()
                    thread = threading.Thread(target=add_files, args=(
                    user_id, kb_name, filename, object_name, file_id, is_enhanced, enable_knowledge_graph, pre_process, meta_data_rules, split_config))
                    lock.acquire()
                    thread.start()
                    lock.release()
                    # ============ Asynchronous addition =============
                else:
                    # ============ Sequential addition =============
                    add_files(user_id, kb_name, filename, object_name, file_id, is_enhanced, enable_knowledge_graph, pre_process, meta_data_rules, split_config, kb_id=kb_id)
                    # ============ Sequential addition =============
                logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name,filename,file_id))
                master_control_logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name, filename, file_id))

            except Exception as e:
                logger.error("kafka处理异常：" + repr(e))
                master_control_logger.error("kafka处理异常：" + repr(e))
                continue



def pre_process_text(text: str, pre_processing_rules: list[str]) -> str:
    for pre_processing_rule in pre_processing_rules:
        if pre_processing_rule == "replace_symbols":
            pattern = r"\n{3,}"
            text = re.sub(pattern, "\n\n", text)
            pattern = r"[\t\f\r\x20\u00a0\u1680\u180e\u2000-\u200a\u202f\u205f\u3000]{2,}"
            text = re.sub(pattern, " ", text)
        elif pre_processing_rule == "delete_links":
            pattern = r"([a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)"
            text = re.sub(pattern, "", text)

            # Remove URL but keep Markdown image URLs
            # First, temporarily replace Markdown image URLs with a placeholder
            markdown_image_pattern = r"!\[.*?\]\((https?://[^\s)]+)\)"
            placeholders: list[str] = []

            def replace_with_placeholder(match, placeholders=placeholders):
                url = match.group(1)
                placeholder = f"__MARKDOWN_IMAGE_URL_{len(placeholders)}__"
                placeholders.append(url)
                return f"![image]({placeholder})"

            text = re.sub(markdown_image_pattern, replace_with_placeholder, text)

            # Now remove all remaining URLs
            url_pattern = r"https?://[^\s)]+"
            text = re.sub(url_pattern, "", text)

            # Finally, restore the Markdown image URLs
            for i, url in enumerate(placeholders):
                text = text.replace(f"__MARKDOWN_IMAGE_URL_{i}__", url)
    return text


def parse_meta_data(docs, parse_rules):
    result = []

    if not parse_rules:
        return result

    def parse_date_to_timestamp(date_str):
        # List of common date formats
        date_formats = [
            # English format
            "%Y-%m-%d %H:%M:%S",
            "%Y-%m-%d %H:%M",
            "%Y-%m-%d",
            "%m/%d/%Y",
            "%m/%d/%Y %H:%M:%S",
            "%d/%m/%Y",
            "%d/%m/%Y %H:%M:%S",
            "%Y/%m/%d",
            "%Y/%m/%d %H:%M:%S",

            # Chinese format
            "%Y年%m月%d日 %H时%M分%S秒",
            "%Y年%m月%d日 %H:%M:%S",
            "%Y年%m月%d日 %H时%M分",
            "%Y年%m月%d日",
            "%Y年%m月%d日 %H:%M",

            # Other common formats
            "%Y.%m.%d",
            "%Y.%m.%d %H:%M:%S",
        ]

        # Preprocessing: Handling some special cases
        processed_str = date_str.strip()
        processed_str = processed_str.replace(" ", "")

        # try every format
        for fmt in date_formats:
            try:
                dt = datetime.strptime(processed_str, fmt)
                return int(dt.timestamp() - 8 * 3600) * 1000
            except ValueError:
                continue

        raise ValueError(f"无法解析日期格式: {date_str}")

    # for rule_name, rules in parse_rules.items():
    for item in parse_rules:
        pattern = item["rule"]
        if pattern:
            # Extract metadata through rules
            meta_datas = []
            #Anti-escaping
            tmp = pattern.encode().decode('unicode_escape')
            pattern = tmp.encode('latin-1').decode('utf-8')
            for doc in docs:
                text = doc["text"]
                matches = re.findall(pattern, text)
                for match in matches:
                    meta_datas.append(str(match.strip()))
            if meta_datas:
                item["value"] = meta_datas[-1]
                if item["value_type"] == "time":
                    try:
                        item["value"] = str(parse_date_to_timestamp(item["value"]))
                    except ValueError:
                        # The extracted time type metadata does not support conversion to unix timestamp.
                        continue
            else:
                #Unable to extract metadata
                continue
        else:
            item["value"] = str(item["value"])

        # If item["rule"] is empty, item["value"] is a fixed value
        result.append(item)

    return result


def add_files(user_id, kb_name, file_name, object_name, file_id,
              is_enhanced, enable_knowledge_graph, pre_process_rules, meta_data_rules, split_config: SplitConfig, kb_id=""):
    response_info = {'code': 0, "message": "成功"}
    user_data_path = USER_DATA_PATH
    convert_dir = CONVERT_DIR
    res_filename = ""

    try:
        filepath = os.path.join(user_data_path, user_id, kb_name)
        logger.info('add_files_filepath=%s' % filepath)
        master_control_logger.info('add_files_filepath=%s' % filepath)
        if not os.path.exists(filepath):
            os.makedirs(filepath)
        else:
            logger.info('filepath=%s 已存在' % filepath)
            master_control_logger.info('filepath=%s 已存在' % filepath)
        logger.info('文档查重开始')
        master_control_logger.info('文档查重开始')
        files_in_milvus = milvus_utils.list_knowledge_file(user_id, kb_name, kb_id=kb_id)
        logger.info('向量库已有文档查询结果：' + repr(files_in_milvus))
        master_control_logger.info('向量库已有文档查询结果：' + repr(files_in_milvus))

        if files_in_milvus['code'] != 0:
            logger.error('文档向量库重复查询校验失败')
            master_control_logger.error('文档向量库重复查询校验失败')
            mq_rel_utils.update_doc_status(file_id, status=51)
            return
        filenames_in_milvus = files_in_milvus['data']['knowledge_file_names']
        if file_name in filenames_in_milvus:
            logger.error('文档已存在该知识库')
            master_control_logger.error('文档已存在该知识库')
            mq_rel_utils.update_doc_status(file_id, status=52)
            return
        else:
            logger.info('文档查重完成')
            master_control_logger.info('文档查重完成')
            mq_rel_utils.update_doc_status(file_id, status=31)
    except Exception as e:
        logger.error(repr(e))
        logger.error('文档向量库重复查询校验失败')
        master_control_logger.error('文档向量库重复查询校验失败' + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=51)
        return

    try:
        logger.info('文档下载开始')
        master_control_logger.info('文档下载开始')
        download_path = os.path.join(filepath, file_name)
        download_status, download_link = minio_utils.get_file_from_minio(object_name, download_path)
        logger.info("------>download_link:%s" % download_link)
        master_control_logger.info("------>download_link:%s" % download_link)
        if not download_status:
            logger.error('文档下载失败')
            master_control_logger.error('文档下载失败')
            mq_rel_utils.update_doc_status(file_id, status=53)
            return
        else:
            logger.info('文档下载完成')
            master_control_logger.info('文档下载完成')
            mq_rel_utils.update_doc_status(file_id, status=32)
            # Convert file format
            base_filename, file_extension = os.path.splitext(file_name)  # Separate filename and suffix
            master_control_logger.info(f"base_filename={base_filename} file_extension={file_extension}")
            if "model" in split_config.parser_choices and file_extension in [".doc", ".docx", ".pptx"]:  # First determine whether the model is analyzed
                convert_office_format_map = {".doc": "pdf", ".docx": "pdf", ".pptx": "pdf"}
                target_format = convert_office_format_map[file_extension]  # Get target format
                res_filename = knowledge_base_utils.convert_office_file(download_path, convert_dir, target_format)
                if res_filename:
                    master_control_logger.error(f"{download_path} convert_office_file successfully => {res_filename}")
                else:
                    master_control_logger.error(f"{download_path} convert_office_file failed")
                    mq_rel_utils.update_doc_status(file_id, status=53)
                    return
            elif file_extension in CONVERT_OFFICE_FORMAT_MAP:
                target_format = CONVERT_OFFICE_FORMAT_MAP[file_extension]  # Get target format
                res_filename = knowledge_base_utils.convert_office_file(download_path, convert_dir, target_format)
                if res_filename:
                    master_control_logger.error(f"{download_path} convert_office_file successfully => {res_filename}")
                else:
                    master_control_logger.error(f"{download_path} convert_office_file failed")
                    mq_rel_utils.update_doc_status(file_id, status=53)
                    return

    except Exception as e:
        logger.error(repr(e))
        logger.error('文档下载失败')
        master_control_logger.error('文档下载失败' + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=53)
        return

    meta_parsed = {}
    try:
        logger.info('文档切分开始')
        master_control_logger.info('文档切分开始')
        if res_filename:  # Need to pass the converted file path
            add_file_path = res_filename
        else:
            add_file_path = download_path
        logger.info('------>add_file_path=%s' % add_file_path)
        master_control_logger.info('------>add_file_path=%s' % add_file_path)
        # Check if the file exists
        if os.path.exists(add_file_path):
            logger.info(f'{user_id}-{kb_name}' + '文件已成功保存存在本地, 文件路径是：' + add_file_path)
            master_control_logger.info(f'{user_id}-{kb_name}' + '文件已成功保存存在本地, 文件路径是：' + add_file_path)
        else:
            logger.info(f'{user_id}-{kb_name}' + add_file_path + ",文件在本地不存在，未保存成功")
            master_control_logger.info(f'{user_id}-{kb_name}' + add_file_path + ",文件在本地不存在，未保存成功")
            logger.error('文档下载完成，但文件不存在本地')
            master_control_logger.error('文档下载完成，但文件不存在本地')
            mq_rel_utils.update_doc_status(file_id, status=53)
            return

        sub_chunk, chunks = file_utils.split_text_file(add_file_path, download_link, split_config)

        if is_enhanced == 'true' and len(chunks) > 0:
            logger.info(f'is_enhanced:{is_enhanced}')
            # enhance_subchunk = add_enhance_subchunk(chunks)
            # if len(enhance_subchunk) > 0:
            #     sub_chunk.extend(enhance_subchunk)

        meta_parsed = parse_meta_data(chunks, meta_data_rules)
        logger.info(f"file_name: {file_name}, meta提取规则: {meta_data_rules}, 提取元数据: {meta_parsed}")
        logger.info(f"file_name: {file_name}, 文本预处理规则: {pre_process_rules}")
        logger.info(repr(file_name) + '文档切分长度：' + repr(len(chunks)))
        logger.info(repr(file_name) + '文档递归切分长度：' + repr(len(sub_chunk)))
        master_control_logger.info(repr(file_name) + '文档切分长度：' + repr(len(chunks)))
        master_control_logger.info(repr(file_name) + '文档递归切分长度：' + repr(len(sub_chunk)))

        file_meta = {}
        with open("./data/%s_chunk.txt" % file_name, 'w', encoding='utf-8') as chunks_file:
            for item in chunks:
                if "download_link" not in item["meta_data"]:
                    item["meta_data"]["download_link"] = download_link  # Add file download link
                if res_filename and "file_name" in item["meta_data"]:  # If there is a converted file, replace it back with the original file name.
                    item["meta_data"]["file_name"] = file_name
                # Store BUCKET and object_name
                item["meta_data"]["bucket_name"] = BUCKET_NAME  # Add file bucket name
                item["meta_data"]["object_name"] = object_name  # Add file download object name

                if pre_process_rules:
                    item["text"] = pre_process_text(item["text"], pre_process_rules)
                item["meta_data"]["doc_meta"] = meta_parsed
                if not file_meta:
                    file_meta = item["meta_data"]
                chunks_file.write(json.dumps(item, ensure_ascii=False))
                chunks_file.write("\n")
        with open("./data/%s_subchunk.txt" % file_name, 'w', encoding='utf-8') as sub_chunk_file:
            for item in sub_chunk:
                if "download_link" not in item["meta_data"]:
                    item["meta_data"]["download_link"] = download_link  # Add file download link
                if res_filename and "file_name" in item["meta_data"]:  # If there is a converted file, replace it back with the original file name.
                    item["meta_data"]["file_name"] = file_name
                # Store BUCKET and object_name
                item["meta_data"]["bucket_name"] = BUCKET_NAME  # Add file bucket name
                item["meta_data"]["object_name"] = object_name  # Add file download object name

                if pre_process_rules:
                    item["content"] = pre_process_text(item["content"], pre_process_rules)
                item["meta_data"]["doc_meta"] = meta_parsed
                sub_chunk_file.write(json.dumps(item, ensure_ascii=False))
                sub_chunk_file.write("\n")

        if len(chunks) == 0 or len(sub_chunk) == 0:
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=61)
            return
        else:
            logger.info('文档切分完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.info('文档切分完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=33)
    except Exception as e:
        import traceback
        logger.error("add_konwledge error %s" % e)
        logger.error(traceback.format_exc())
        if "Error loading" in repr(e):  # File not available
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error(
                '文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
            mq_rel_utils.update_doc_status(file_id, status=62)
            return
        else:
            logger.error(repr(e))
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
            mq_rel_utils.update_doc_status(file_id, status=54)
            return
    try:
        logger.info('添加文档meta开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.info('添加文档meta开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        add_file_result = es_utils.add_file(user_id, kb_name, file_name, file_meta, kb_id=kb_id)
        logger.info(repr(file_name) + '添加文档meta结果：' + repr(add_file_result))
        master_control_logger.info(repr(file_name) + '添加文档meta结果：' + repr(add_file_result))
        if add_file_result['code'] != 0:
            # callback
            logger.error('添加文档meta失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error('添加文档meta失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=55)
            return
        else:
            logger.info('添加文档meta完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    except Exception as e:
        logger.error(repr(e))
        logger.error('添加文档meta失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.error(
            '添加文档meta失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=55)
        return

    try:
        logger.info('文档插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.info('文档插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        insert_milvus_result = milvus_utils.add_milvus(user_id, kb_name, sub_chunk, file_name, add_file_path, kb_id=kb_id)
        logger.info(repr(file_name) + '添加milvus结果：' + repr(insert_milvus_result))
        master_control_logger.info(repr(file_name) + '添加milvus结果：' + repr(insert_milvus_result))
        if insert_milvus_result['code'] != 0:
            logger.error('文档插入milvus失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error('文档插入milvus失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=55)
            return
        else:
            logger.info('文档插入milvus完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.info('文档插入milvus完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=34)
    except Exception as e:
        logger.error(repr(e))
        logger.error('文档插入milvus失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.error('文档插入milvus失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=55)
        return

    try:
        logger.info('文档插入es开始')
        master_control_logger.info('文档插入es开始')
        insert_es_result = es_utils.add_es(user_id, kb_name, chunks, file_name, kb_id=kb_id)
        logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
        master_control_logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
        if insert_es_result['code'] != 0:
            logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=56)
            return
        else:
            logger.info('文档插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            master_control_logger.info('文档插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=35)
    except Exception as e:
        logger.error(repr(e))
        logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=56)
        return

    # ---------------7. Final completion

    logger.info("user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + '===== 文档上传成功且完成')
    master_control_logger.info("user_id=%s,kb_name=%s,file_name=%s,kb_id=%s" % (user_id, kb_name, file_name, kb_id) + '===== 文档上传成功且完成')
    mq_rel_utils.update_doc_status(file_id, status=10)


if __name__ == "__main__":
    kafkal()
