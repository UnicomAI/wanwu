import copy
import logging

from logging_config import init_logging

from utils import milvus_utils
from utils import minio_utils
from utils import es_utils
from utils import file_utils
from utils import mq_rel_utils
from utils import knowledge_base_utils
from utils.file_utils import SplitConfig
from utils import schema_utils
from utils import redis_utils
from utils.otel import init_tracer
import subprocess
from kafka import KafkaConsumer, TopicPartition, OffsetAndMetadata
import json
import threading
from datetime import datetime
import re
from settings import *
from utils.constant import CONVERT_DIR, USER_DATA_PATH

# 初始化 OpenTelemetry，并自动 instrument requests / kafka 客户端
init_tracer("rag-wanwu-add-file")
from opentelemetry import trace as otel_trace
from opentelemetry.propagate import extract
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.kafka import KafkaInstrumentor

RequestsInstrumentor().instrument()
KafkaInstrumentor().instrument()

# KafkaInstrumentor 对 kafka-python 的 consumer 自动 span 只在 __next__ 那一瞬有效，
# for 循环体内业务代码会失去 active span。手动用 extract + start_as_current_span
# 重建 consumer span，让整段处理（含业务日志和出向 HTTP）都能挂在同一条 trace 上。
tracer = otel_trace.get_tracer(__name__)

graph_redis_client = redis_utils.get_redis_connection()

# 定义路径
paths = ["./data", "./user_data"]
# 遍历路径列表
for path in paths:
    # 检查路径是否存在
    if not os.path.exists(path):
        # 如果不存在，则创建目录
        os.makedirs(path)
        print(f"目录 {path} 已创建。")
    else:
        print(f"目录 {path} 已存在。")


logger = logging.getLogger(__name__)

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
                                     max_poll_records=1,  # 设置每次最多拉取1条消息
                                     # max_poll_interval_ms=8000000,  # 设置最大轮询间隔为120分钟
                                     value_deserializer=lambda x: x.decode('utf-8'))

        else:
            consumer = KafkaConsumer(KAFKA_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     group_id=KAFKA_GROUP_ID,
                                     enable_auto_commit=KAFKA_ENABLE_AUTO_COMMIT,
                                     max_poll_records=1,  # 设置每次最多拉取1条消息
                                     # max_poll_interval_ms=8000000,  # 设置最大轮询间隔为120分钟
                                     value_deserializer=lambda x: x.decode('utf-8'))
        for message in consumer:
            # 从 Kafka 消息 headers 抽出上游 producer 注入的 traceparent
            headers_dict = {}
            if message.headers:
                for hk, hv in message.headers:
                    key = hk.decode() if isinstance(hk, (bytes, bytearray)) else hk
                    val = hv.decode() if isinstance(hv, (bytes, bytearray)) else hv
                    headers_dict[key] = val
            parent_ctx = extract(headers_dict)

            with tracer.start_as_current_span(
                "kafka.consume",
                context=parent_ctx,
                kind=otel_trace.SpanKind.CONSUMER,
            ):
                # 初始化用户知识库路径
                print('收到新kafka消息：' + repr(message.value))
                logger.info('收到新kafka消息：' + repr(message.value))
                try:
                    message_value = json.loads(message.value)
                    doc = message_value.get("doc", {})
                except (json.JSONDecodeError, AttributeError) as e:
                    logger.error(f"Failed to parse message: {e}")
                    continue


                file_id = doc.get("id")
                user_id = doc.get("userId")
                kb_id = doc.get("kb_id")
                if not kb_id:
                    logger.error(f"kafka 消息缺少 kb_id, doc: {doc}")
                    continue
                kb_name = doc.get("categoryId")
                filename = doc.get("originalName")
                object_name = doc.get("objectName")

                # 分块与解析配置
                sentence_size = doc.get("chunk_size")
                overlap_size = doc.get("overlap")
                split_type = doc.get("split_type", "common")
                chunk_type = doc.get("chunk_type", "default")
                separators = doc.get("separators", ['。'])

                # 状态开关（建议处理成布尔值或统一格式）
                is_enhanced = doc.get("is_enhanced", "false")
                enable_knowledge_graph = doc.get("enable_knowledge_graph", "false")

                # 文件导入时选择解析方式，默认勾选文字提取。可选模型解析(model)，当多选时此参数默认为["text"]，当勾选model时传：["text","model"]
                parser_choices = doc.get("parser_choices", ["text"])
                asr_model_id = doc.get("asr_model_id", "")
                multimodal_model_id = doc.get("multimodal_model_id", "")
                ocr_model_id = doc.get("ocr_model_id", "")
                pre_process = doc.get("pre_process", [])
                meta_data_rules = doc.get("meta_data", [])
                child_chunk_config = doc.get("child_chunk_config")

                split_config = SplitConfig(
                    sentence_size=sentence_size,
                    overlap_size=overlap_size,
                    chunk_type=chunk_type,
                    separators=separators,
                    parser_choices=parser_choices,
                    ocr_model_id=ocr_model_id,
                    asr_model_id=asr_model_id,
                    multimodal_model_id=multimodal_model_id,
                    split_type=split_type,
                    child_chunk_config=child_chunk_config
                )

                try:
                    if not KAFKA_ENABLE_AUTO_COMMIT:
                        # 提交当前消息的偏移量
                        tp = TopicPartition(KAFKA_TOPICS, message.partition)
                        offset_and_metadata = OffsetAndMetadata(offset=message.offset + 1, metadata="")
                        offsets = {tp: offset_and_metadata}
                        # consumer.commit(offsets=offsets)  # 不需要使用这个提交
                        consumer.commit()
                        logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                        logger.info('consumer.commit offset：' + repr(offsets))

                    kb_info = {"kb_name": kb_name, "kb_id": kb_id}
                    if KAFKA_USE_ASYN_ADD:
                        # ============ 异步添加 =============
                        lock = threading.Lock()
                        thread = threading.Thread(target=add_files, args=(
                        user_id, kb_info, filename, object_name, file_id, is_enhanced, enable_knowledge_graph, pre_process, meta_data_rules, split_config))
                        lock.acquire()
                        thread.start()
                        lock.release()
                        # ============ 异步添加 =============
                    else:
                        # ============ 顺序添加 =============
                        add_files(user_id, kb_info, filename, object_name, file_id, is_enhanced, enable_knowledge_graph, pre_process, meta_data_rules, split_config)
                        # ============ 顺序添加 =============
                    logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name,filename,file_id))

                except Exception as e:
                    logger.error("kafka处理异常：" + repr(e))
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


def retype_meta_datas(meta_datas: list):
    result = []
    for item in meta_datas:
        new_item = copy.deepcopy(item)
        if item["value_type"] == "string":
            new_item["string_value"] = str(item["value"])
        elif item["value_type"] == "number" or item["value_type"] == "time":
            new_item["int_value"] = int(item["value"])
        result.append(new_item)
    return result


def parse_meta_data(docs, parse_rules):
    result = []

    if not parse_rules:
        return result

    def parse_date_to_timestamp(date_str):
        # 常见的日期格式列表
        date_formats = [
            # 英文格式
            "%Y-%m-%d %H:%M:%S",
            "%Y-%m-%d %H:%M",
            "%Y-%m-%d",
            "%m/%d/%Y",
            "%m/%d/%Y %H:%M:%S",
            "%d/%m/%Y",
            "%d/%m/%Y %H:%M:%S",
            "%Y/%m/%d",
            "%Y/%m/%d %H:%M:%S",

            # 中文格式
            "%Y年%m月%d日 %H时%M分%S秒",
            "%Y年%m月%d日 %H:%M:%S",
            "%Y年%m月%d日 %H时%M分",
            "%Y年%m月%d日",
            "%Y年%m月%d日 %H:%M",

            # 其他常见格式
            "%Y.%m.%d",
            "%Y.%m.%d %H:%M:%S",
        ]

        # 预处理：处理一些特殊情况
        processed_str = date_str.strip()
        processed_str = processed_str.replace(" ", "")
        for suffix in ["发布", "实施"]:
            if processed_str.endswith(suffix):
                processed_str = processed_str[:-len(suffix)]
                break

        # 尝试每种格式
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
            # 通过rule提取元数据
            meta_datas = []
            #反转义
            tmp = pattern.encode().decode('unicode_escape')
            pattern = tmp.encode('latin-1').decode('utf-8')
            for doc in docs:
                text = doc["text"]
                matches = re.findall(pattern, text)
                for match in matches:
                    if isinstance(match, tuple):
                        raise ValueError(f"Regex rule returned tuple (capture group detected): {match}")
                    meta_datas.append(str(match.strip()))
            if meta_datas:
                item["value"] = meta_datas[-1]
                if item["value_type"] == "time":
                    try:
                        item["value"] = str(parse_date_to_timestamp(item["value"]))
                    except ValueError:
                        # 提取的时间类型的元数据不支持转成unix timestamp
                        continue
            else:
                #提取不到元数据
                continue
        else:
            item["value"] = str(item["value"])

        # 如果item["rule"] 为空，item["value"]是固定值
        result.append(item)

    return retype_meta_datas(result)


def add_files(user_id, kb_info, file_name, object_name, file_id,
              is_enhanced, enable_knowledge_graph, pre_process_rules, meta_data_rules, split_config: SplitConfig):
    response_info = {'code': 0, "message": "成功"}
    user_data_path = USER_DATA_PATH
    convert_dir = CONVERT_DIR
    res_filename = ""
    kb_name = kb_info["kb_name"]
    # 收到 kafka 消息，开始解析
    mq_rel_utils.update_doc_status(file_id, status=20)
    if not is_safe_filename(file_name):
        logger.error('文件名不合法')
        mq_rel_utils.update_doc_status(file_id, status=53)
        return
    try:
        filepath = os.path.join(user_data_path, user_id, kb_name)
        logger.info('add_files_filepath=%s' % filepath)
        if not os.path.exists(filepath):
            os.makedirs(filepath)
        else:
            logger.info('filepath=%s 已存在' % filepath)
        logger.info('文档查重开始')
        files_in_milvus = milvus_utils.list_knowledge_file(user_id, kb_info)
        logger.info('向量库已有文档查询结果：' + repr(files_in_milvus))

        if files_in_milvus['code'] != 0:
            logger.error('文档向量库重复查询校验失败')
            mq_rel_utils.update_doc_status(file_id, status=51)
            return
        filenames_in_milvus = files_in_milvus['data']['knowledge_file_names']
        if file_name in filenames_in_milvus:
            logger.error('文档已存在该知识库')
            mq_rel_utils.update_doc_status(file_id, status=52)
            return
        else:
            logger.info('文档查重完成')
            mq_rel_utils.update_doc_status(file_id, status=31)
    except Exception as e:
        logger.error(repr(e))
        logger.error('文档向量库重复查询校验失败')
        mq_rel_utils.update_doc_status(file_id, status=51)
        return

    try:
        logger.info('文档下载开始')
        download_path = os.path.join(filepath, file_name)
        download_status, download_link = minio_utils.get_file_from_minio(object_name, download_path)
        logger.info("------>download_link:%s" % download_link)
        if not download_status:
            logger.error('文档下载失败')
            mq_rel_utils.update_doc_status(file_id, status=53)
            return
        else:
            logger.info('文档下载完成')
            mq_rel_utils.update_doc_status(file_id, status=32)
            # 转换文件格式
            base_filename, file_extension = os.path.splitext(file_name)  # 分离文件名和后缀
            logger.info(f"base_filename={base_filename} file_extension={file_extension}")
            if "model" in split_config.parser_choices and file_extension in [".doc", ".docx", ".pptx"]:  # 先判断是否模型解析
                convert_office_format_map = {".doc": "pdf", ".docx": "pdf", ".pptx": "pdf"}
                target_format = convert_office_format_map[file_extension]  # 获取目标格式
                res_filename = knowledge_base_utils.convert_office_file(download_path, convert_dir, target_format)
                if res_filename:
                    logger.info(f"{download_path} convert_office_file successfully => {res_filename}")
                else:
                    logger.error(f"{download_path} convert_office_file failed")
                    mq_rel_utils.update_doc_status(file_id, status=53)
                    return
            elif file_extension in CONVERT_OFFICE_FORMAT_MAP:
                target_format = CONVERT_OFFICE_FORMAT_MAP[file_extension]  # 获取目标格式
                res_filename = knowledge_base_utils.convert_office_file(download_path, convert_dir, target_format)
                if res_filename:
                    logger.info(f"{download_path} convert_office_file successfully => {res_filename}")
                else:
                    logger.error(f"{download_path} convert_office_file failed")
                    mq_rel_utils.update_doc_status(file_id, status=53)
                    return

    except Exception as e:
        logger.error(repr(e))
        logger.error('文档下载失败')
        mq_rel_utils.update_doc_status(file_id, status=53)
        return

    meta_parsed = {}
    try:
        logger.info('文档切分开始')
        if res_filename:  # 需要传递转换后的 文件路径
            add_file_path = res_filename
        else:
            add_file_path = download_path
        logger.info('------>add_file_path=%s' % add_file_path)
        # 检查文件是否存在
        if os.path.exists(add_file_path):
            logger.info(f'{user_id}-{kb_name}' + '文件已成功保存存在本地, 文件路径是：' + add_file_path)
        else:
            logger.info(f'{user_id}-{kb_name}' + add_file_path + ",文件在本地不存在，未保存成功")
            logger.error('文档下载完成，但文件不存在本地')
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

        file_meta = {}
        with open("./data/%s_chunk.txt" % file_name, 'w', encoding='utf-8') as chunks_file:
            for item in chunks:
                if "download_link" not in item["meta_data"]:
                    item["meta_data"]["download_link"] = download_link  # 添加file下载链接
                if res_filename and "file_name" in item["meta_data"]:  # 如果有转换后的文件，则替换回原来文件名
                    item["meta_data"]["file_name"] = file_name
                # 存储 BUCKET 和 object_name
                item["meta_data"]["bucket_name"] = BUCKET_NAME  # 添加文件桶名
                item["meta_data"]["object_name"] = object_name  # 添加文件下载对象名

                if pre_process_rules:
                    item["text"] = pre_process_text(item["text"], pre_process_rules)
                    # parent_text 作为父段存储在snippet index中，该字段用户生成content_id, 需要和subchunk中的content字段内容保持一致
                    if "parent_text" in item:
                        item["parent_text"] = pre_process_text(item["parent_text"], pre_process_rules)
                item["meta_data"]["doc_meta"] = meta_parsed
                if not file_meta:
                    file_meta = item["meta_data"]
                chunks_file.write(json.dumps(item, ensure_ascii=False))
                chunks_file.write("\n")
        with open("./data/%s_subchunk.txt" % file_name, 'w', encoding='utf-8') as sub_chunk_file:
            for item in sub_chunk:
                if "download_link" not in item["meta_data"]:
                    item["meta_data"]["download_link"] = download_link  # 添加file下载链接
                if res_filename and "file_name" in item["meta_data"]:  # 如果有转换后的文件，则替换回原来文件名
                    item["meta_data"]["file_name"] = file_name
                # 存储 BUCKET 和 object_name
                item["meta_data"]["bucket_name"] = BUCKET_NAME  # 添加文件桶名
                item["meta_data"]["object_name"] = object_name  # 添加文件下载对象名

                if pre_process_rules:
                    item["content"] = pre_process_text(item["content"], pre_process_rules)
                item["meta_data"]["doc_meta"] = meta_parsed
                sub_chunk_file.write(json.dumps(item, ensure_ascii=False))
                sub_chunk_file.write("\n")

        if len(chunks) == 0 or len(sub_chunk) == 0:
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=61)
            return
        else:
            logger.info('文档切分完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=33)
    except Exception as e:
        import traceback
        logger.error("add_konwledge error %s" % e)
        logger.error(traceback.format_exc())
        if "Error loading" in repr(e):  # 文件不可用
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=62)
            return
        else:
            logger.error(repr(e))
            logger.error('文档切分失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=54)
            return
    try:
        logger.info('添加文档meta开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        add_file_result = es_utils.add_file(user_id, kb_info, file_name, file_meta)
        logger.info(repr(file_name) + '添加文档meta结果：' + repr(add_file_result))
        if add_file_result['code'] != 0:
            # 回调
            logger.error('添加文档meta失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=55)
            return
        else:
            logger.info('添加文档meta完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    except Exception as e:
        logger.error(repr(e))
        logger.error('添加文档meta失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        mq_rel_utils.update_doc_status(file_id, status=55)
        return

    try:
        logger.info('文档插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        insert_milvus_result = milvus_utils.add_milvus(user_id, kb_info, sub_chunk, file_name, add_file_path)
        logger.info(repr(file_name) + '添加milvus结果：' + repr(insert_milvus_result))
        if insert_milvus_result['code'] != 0:
            logger.error('文档插入milvus失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=55)
            return
        else:
            logger.info('文档插入milvus完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=34)
    except Exception as e:
        logger.error(f"文档插入milvus失败, user_id={user_id},kb_name={kb_name},file_name={file_name}, exception={repr(e)}")
        mq_rel_utils.update_doc_status(file_id, status=55)
        return

    try:
        logger.info('文档插入es开始')
        insert_es_result = es_utils.add_es(user_id, kb_info, chunks, file_name)
        logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
        if insert_es_result['code'] != 0:
            logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=56)
            return
        else:
            logger.info('文档插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=35)
    except Exception as e:
        logger.error(repr(e))
        logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        mq_rel_utils.update_doc_status(file_id, status=56)
        return

    # --------------7、最终完成

    logger.info("user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + '===== 文档上传成功且完成')
    mq_rel_utils.update_doc_status(file_id, status=10, meta_datas=meta_parsed)


def is_safe_filename(name: str) -> bool:
    if "/" in name or "\\" in name:
        return False
    if ".." in name:
        return False
    return True

if __name__ == "__main__":
    init_logging()
    kafkal()
