import time
import logging

from logging_config import init_logging
from utils import minio_utils
from utils import es_utils
from utils import milvus_utils
from utils import mq_rel_utils
from utils import redis_utils
from utils import graph_utils
from utils import knowledge_base_utils
from utils.otel import init_tracer

from concurrent.futures import ThreadPoolExecutor
from kafka import KafkaConsumer, TopicPartition, OffsetAndMetadata
import json
import threading

from settings import *

# 初始化 OpenTelemetry，并自动 instrument requests / kafka 客户端
init_tracer("rag-wanwu-graph-extract")
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

logger = logging.getLogger(__name__)

graph_redis_client = redis_utils.get_redis_connection()


def kafkal():
    executor = ThreadPoolExecutor(max_workers=5) if KAFKA_ENABLE_AUTO_COMMIT else None
    while True:
        print('开始消费消息')
        if KAFKA_SASL_USE:
            consumer = KafkaConsumer(KAFKA_GRAPH_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     security_protocol='SASL_PLAINTEXT',
                                     sasl_mechanism='PLAIN',
                                     sasl_plain_username=KAFKA_SASL_PLAIN_USERNAME,
                                     sasl_plain_password=KAFKA_SASL_PLAIN_PASSWORD,
                                     group_id=KAFKA_GRAPH_GROUP_ID,
                                     enable_auto_commit=KAFKA_ENABLE_AUTO_COMMIT,
                                     max_poll_records=1,  # 设置每次最多拉取1条消息
                                     # max_poll_interval_ms=8000000,  # 设置最大轮询间隔为120分钟
                                     value_deserializer=lambda x: x.decode('utf-8'))

        else:
            consumer = KafkaConsumer(KAFKA_GRAPH_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     group_id=KAFKA_GRAPH_GROUP_ID,
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
                print('收到kafka消息：' + repr(message.value))
                logger.info('收到kafka消息：' + repr(message.value))
                message_value = json.loads(message.value)

                kb_name = message_value["doc"]["categoryId"]
                user_id = message_value["doc"]["userId"]
                kb_id = message_value["doc"].get("kb_id")
                if not kb_id:
                    logger.error(f"kafka 消息缺少 kb_id, doc: {message_value.get('doc')}")
                    continue
                filename = message_value["doc"].get("originalName", "")
                file_id = message_value["doc"].get("id", "")
                graph_schema_objectname = message_value["doc"].get("graph_schema_objectname", "")
                graph_schema_filename = message_value["doc"].get("graph_schema_filename", "")
                enable_knowledge_graph = message_value["doc"].get("enable_knowledge_graph", False)
                message_type = message_value["doc"].get("message_type", "graph")
                graph_model_id = message_value["doc"].get("graph_model_id", "")

                try:
                    if not KAFKA_ENABLE_AUTO_COMMIT:
                        # 提交当前消息的偏移量
                        tp = TopicPartition(KAFKA_GRAPH_TOPICS, message.partition)
                        offset_and_metadata = OffsetAndMetadata(offset=message.offset + 1, metadata="")
                        offsets = {tp: offset_and_metadata}
                        consumer.commit()
                        logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                        logger.info('consumer.commit offset：' + repr(offsets))

                    kb_info = {"kb_name": kb_name, "kb_id": kb_id}
                    if KAFKA_USE_GRAPH_ASYN_ADD:
                        # ============ 异步添加 =============
                        if message_type == "graph":
                            executor.submit(extrac_graph_data,
                                            user_id, kb_info, filename, file_id, enable_knowledge_graph,
                                            graph_schema_objectname, graph_schema_filename, graph_model_id)
                        elif message_type == "community_report":
                            executor.submit(generate_community_report,
                                            user_id, kb_info, enable_knowledge_graph, graph_model_id)
                        else:
                            logger.warning(f"未知的message_type: {message_type}")
                            continue
                    else:
                        # ============ 顺序添加 =============
                        if message_type == "graph":
                            extrac_graph_data(user_id, kb_info, filename, file_id, enable_knowledge_graph,
                                            graph_schema_objectname, graph_schema_filename, graph_model_id)
                        elif message_type == "community_report":
                            generate_community_report(user_id, kb_info, enable_knowledge_graph, graph_model_id)
                        else:
                            logger.warning(f"未知的message_type: {message_type}")
                            continue
                    logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name,filename,file_id))

                except Exception as e:
                    logger.error("kafka处理异常：" + repr(e))
                    continue


def extrac_graph_data(user_id, kb_info, file_name, file_id, enable_knowledge_graph, graph_schema_objectname, graph_schema_filename, graph_model_id=""):
    kb_name = kb_info["kb_name"]
    kb_id = kb_info["kb_id"]
    # 图谱解析开始执行
    mq_rel_utils.update_doc_status(file_id, status=110)

    user_data_path = './user_data'
    filepath = os.path.join(user_data_path, user_id, kb_name)
    logger.info('add_files_filepath=%s' % filepath)
    if not os.path.exists(filepath):
        os.makedirs(filepath)
    else:
        logger.info('filepath=%s 已存在' % filepath)

    # -------------- 先解析 graph schema --------------
    schema = {}
    # 当graph_schema_filename,graph_schema_objectname有值则说明用户自己上传excel，否则schema为空后续会用内置schema抽取
    if enable_knowledge_graph and graph_schema_filename and graph_schema_objectname:
        try:
            schema_file_path = os.path.join(filepath, graph_schema_filename)
            graph_download_status, graph_download_link = minio_utils.get_file_from_minio(graph_schema_objectname,
                                                                                         schema_file_path)
            logger.info("graph_download_status=%s,graph_download_link=%s" %
                        (graph_download_status, graph_download_link))
            schema = graph_utils.parse_excel_to_schema_json(schema_file_path)
            logger.info(f'提取graph schema成功'
                         + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + str(schema))
            # graph schema 解析成功
            mq_rel_utils.update_doc_status(file_id, status=111)
        except Exception as e:
            logger.error(repr(e))
            logger.error(f'提取graph schema失败'
                         + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            mq_rel_utils.update_doc_status(file_id, status=104)
            return

    # -------------- 再从数据库中获取 all_extrac_graph_chunks --------------
    try:
        all_wait_extrac_chunks = graph_utils.get_all_extrac_graph_chunks(user_id, kb_info, file_name)
        logger.info(repr(file_name) + 'all_wait_extrac_chunks长度：' + repr(len(all_wait_extrac_chunks)))
        logger.info('all_wait_extrac_chunks 获取完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        # 生成图谱获取 chunk 文本成功
        mq_rel_utils.update_doc_status(file_id, status=112)
    except Exception as e:
        import traceback
        logger.error(traceback.format_exc())
        logger.error(repr(e))
        logger.error('all_wait_extrac_chunks 获取失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        mq_rel_utils.update_doc_status(file_id, status=101)
        return


    # -------------- 将切分好的chunks 进行图谱数据提取构建 --------------
    # 将 chunks 按 batch_size 分组提取
    all_graph_chunks = []
    all_graph_vocabulary_set = set()
    batch_size = 10
    if enable_knowledge_graph:
        for i in range(0, len(all_wait_extrac_chunks), batch_size):
            batch_num = int(i/batch_size) + 1
            temp_chunks = all_wait_extrac_chunks[i:i + batch_size]
            try:
                result_data = graph_utils.get_extrac_graph_data(user_id, kb_name, temp_chunks, file_name, graph_model_id,
                                                                schema=schema)
                graph_chunks = result_data['graph_chunks']
                all_graph_chunks.extend(graph_chunks)
                all_graph_vocabulary_set.update(result_data['graph_vocabulary_set'])
                logger.info(f'第{batch_num}批文档提取graph数据成功'
                             + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + str(result_data))
            except Exception as e:
                logger.error(repr(e))
                logger.error(f'第{batch_num}批文档提取graph数据失败'
                             + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                mq_rel_utils.update_doc_status(file_id, status=102)
                return

        # 提取图谱成功
        mq_rel_utils.update_doc_status(file_id, status=113)

    # --------------  insert es graph_data ----------------
    try:
        if enable_knowledge_graph and len(all_graph_chunks) > 0:
            logger.info(f'graph_data 插入es开始,all_graph_chunks len:{len(all_graph_chunks)}')
            insert_es_result = es_utils.add_es(user_id, kb_info, all_graph_chunks, file_name)
            logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
            if insert_es_result['code'] != 0:
                # 回调
                logger.error('graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                mq_rel_utils.update_doc_status(file_id, status=103)
                return
            else:
                # 插入成功后，更新update_graph_vocabulary_set 数据
                redis_utils.update_graph_vocabulary_set(graph_redis_client, kb_id,
                                                        elements_to_add=all_graph_vocabulary_set)
                # 回调
                logger.info('graph_data插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                # 图谱持久化存储成功
                mq_rel_utils.update_doc_status(file_id, status=114)
    except Exception as e:
        logger.error(repr(e))
        logger.error('graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        mq_rel_utils.update_doc_status(file_id, status=103)
        return

    # --------------7、最终完成
    # 回调
    logger.info("user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + '===== 文档grahp解析成功且完成')
    mq_rel_utils.update_doc_status(file_id, status=100)


def generate_community_report(user_id, kb_info, enable_knowledge_graph, graph_model_id=""):
    kb_name = kb_info["kb_name"]
    kb_id = kb_info["kb_id"]
    # 社区报告开始生成
    mq_rel_utils.update_kb_status(kb_id, status=130)

    # 清理旧的社区报告
    try:
        clear_result = milvus_utils.del_community_reports(user_id, kb_info, clear_reports=True)
        if clear_result['code'] != 0:
            raise RuntimeError(clear_result["message"])
        logger.info(f'清理社区报告成功'
                    + "user_id=%s,kb_name=%s" % (user_id, kb_name) + str(clear_result))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"清理社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        mq_rel_utils.update_kb_status(kb_id, status=124)
        return

    # 提取社区报告
    try:
        reports_result = graph_utils.generate_community_reports(user_id, kb_name, graph_model_id)
        reports = reports_result['community_reports']
        if len(reports) == 0:
            raise ValueError("社区报告数量为0，生成失败")
        logger.info(f"生成社区报告成功, user_id=%s,kb_name=%s, size=%s" % (user_id, kb_name, len(reports)))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"生成社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        mq_rel_utils.update_kb_status(kb_id, status=122)
        return

    # 存储社区报告
    try:
        chunk_current_num = 0
        sub_chunks = []
        chunk_total_num = len(reports)
        file_name = "社区报告"
        for report_data in reports:
            embedding_content = f"# {report_data['report_title']} \n\n {report_data['report_summary']}"
            sub_chunks.append({
                "content": report_data["report"],
                "title": report_data["report_title"],
                "embedding_content": embedding_content[:200],
                "meta_data": {
                    "file_name": file_name,
                    "entities": report_data["entities"],
                    "chunk_total_num": chunk_total_num,
                    "chunk_current_num": chunk_current_num
                },
                "create_time":str(int(time.time() * 1000))
            })
            chunk_current_num += 1
        logger.info('社区报告插入milvus开始' + "user_id=%s,kb_name=%s" % (user_id, kb_name))
        insert_milvus_result = milvus_utils.add_milvus(user_id, kb_info, sub_chunks, file_name, "",
                                                       milvus_url=milvus_utils.ADD_COMMUNItY_REPORT_URL)
        if insert_milvus_result['code'] != 0:
            raise RuntimeError(insert_milvus_result["message"])
        logger.info(f'社区报告插入milvus成功'
                    + "user_id=%s,kb_name=%s" % (user_id, kb_name) + str(insert_milvus_result))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"社区报告插入milvus失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        mq_rel_utils.update_kb_status(kb_id, status=123)
        return

    # 最终完成
    logger.info("user_id=%s,kb_name=%s" % (user_id, kb_name) + '===== 社区报告生成且存储完成')
    mq_rel_utils.update_kb_status(kb_id, status=120)

if __name__ == "__main__":
    init_logging()
    kafkal()
