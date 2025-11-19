import time

from utils import minio_utils
from utils import es_utils
from utils import milvus_utils
from utils import mq_rel_utils
from utils import redis_utils
from utils import graph_utils
from utils import knowledge_base_utils
from model_manager import get_model_configure, LlmModelConfig

from concurrent.futures import ThreadPoolExecutor
from kafka import KafkaConsumer, TopicPartition, OffsetAndMetadata
import json
import threading
from logging_config import setup_logging

from settings import *

logger_name = 'rag_graph_asyn_add_files_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

master_control_logger_name = 'mc_rag_graph_asyn_add_files_utils'
master_control_app_name = os.getenv("LOG_FILE") + "_master_control"
master_control_logger = setup_logging(master_control_app_name, master_control_logger_name)
master_control_logger.info(logger_name + '---------LOG_FILE：' + repr(master_control_app_name))

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
                                     max_poll_records=1,  # Set to pull at most 1 message at a time
                                     # max_poll_interval_ms=8000000, # Set the maximum polling interval to 120 minutes
                                     value_deserializer=lambda x: x.decode('utf-8'))

        else:
            consumer = KafkaConsumer(KAFKA_GRAPH_TOPICS,
                                     bootstrap_servers=KAFKA_BOOTSTRAP_SERVERS,
                                     group_id=KAFKA_GRAPH_GROUP_ID,
                                     enable_auto_commit=KAFKA_ENABLE_AUTO_COMMIT,
                                     max_poll_records=1,  # Set to pull at most 1 message at a time
                                     # max_poll_interval_ms=8000000, # Set the maximum polling interval to 120 minutes
                                     value_deserializer=lambda x: x.decode('utf-8'))
        for message in consumer:
            print('收到kafka消息：' + repr(message.value))
            logger.info('收到kafka消息：' + repr(message.value))
            master_control_logger.info('收到kafka消息：' + repr(message.value))
            message_value = json.loads(message.value)

            kb_name = message_value["doc"]["categoryId"]
            user_id = message_value["doc"]["userId"]
            kb_id = message_value["doc"].get("kb_id", "")
            filename = message_value["doc"].get("originalName", "")
            file_id = message_value["doc"].get("id", "")
            graph_schema_objectname = message_value["doc"].get("graph_schema_objectname", "")
            graph_schema_filename = message_value["doc"].get("graph_schema_filename", "")
            enable_knowledge_graph = message_value["doc"].get("enable_knowledge_graph", False)
            message_type = message_value["doc"].get("message_type", "graph")
            graph_model_id = message_value["doc"].get("graph_model_id", "")

            try:
                if not KAFKA_ENABLE_AUTO_COMMIT:
                    # Commit the offset of the current message
                    tp = TopicPartition(KAFKA_GRAPH_TOPICS, message.partition)
                    offset_and_metadata = OffsetAndMetadata(offset=message.offset + 1, metadata="")
                    offsets = {tp: offset_and_metadata}
                    consumer.commit()
                    logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                    master_control_logger.info('kafka异步消费完成 ===== 已提交 offset：' + str(message.offset) + '===== kafka消息：' + repr(message.value))
                    logger.info('consumer.commit offset：' + repr(offsets))
                    master_control_logger.info('consumer.commit offset：' + repr(offsets))

                if KAFKA_USE_GRAPH_ASYN_ADD:
                    # ============ Asynchronous addition =============
                    if message_type == "graph":
                        executor.submit(extrac_graph_data,
                                        user_id, kb_name, filename, file_id, enable_knowledge_graph,
                                        graph_schema_objectname, graph_schema_filename, graph_model_id, kb_id)
                    elif message_type == "community_report":
                        executor.submit(generate_community_report,
                                        user_id, kb_name, enable_knowledge_graph, graph_model_id, kb_id)
                    else:
                        logger.warning(f"未知的message_type: {message_type}")
                        master_control_logger.warning(f"未知的message_type: {message_type}")
                        continue
                else:
                    # ============ Sequential addition =============
                    if message_type == "graph":
                        extrac_graph_data(user_id, kb_name, filename, file_id, enable_knowledge_graph,
                                        graph_schema_objectname, graph_schema_filename, graph_model_id, kb_id=kb_id)
                    elif message_type == "community_report":
                        generate_community_report(user_id, kb_name, enable_knowledge_graph, graph_model_id,
                                                kb_id=kb_id)
                    else:
                        logger.warning(f"未知的message_type: {message_type}")
                        master_control_logger.warning(f"未知的message_type: {message_type}")
                        continue
                logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name,filename,file_id))
                master_control_logger.info('----->kafka异步消费完成：user_id=%s,kb_name=%s,filename=%s,file_id=%s,process finished' % (user_id, kb_name, filename, file_id))

            except Exception as e:
                logger.error("kafka处理异常：" + repr(e))
                master_control_logger.error("kafka处理异常：" + repr(e))
                continue


def extrac_graph_data(user_id, kb_name, file_name, file_id, enable_knowledge_graph, graph_schema_objectname, graph_schema_filename, graph_model_id="", kb_id=""):
    # Graph analysis begins
    mq_rel_utils.update_doc_status(file_id, status=110)

    # --------------- We will first get all_extrac_graph_chunks from the database--------------
    try:
        user_data_path = './user_data'
        filepath = os.path.join(user_data_path, user_id, kb_name)
        logger.info('add_files_filepath=%s' % filepath)
        master_control_logger.info('add_files_filepath=%s' % filepath)
        if not os.path.exists(filepath):
            os.makedirs(filepath)
        else:
            logger.info('filepath=%s 已存在' % filepath)
            master_control_logger.info('filepath=%s 已存在' % filepath)
        all_wait_extrac_chunks = graph_utils.get_all_extrac_graph_chunks(user_id, kb_name, file_name)
        logger.info(repr(file_name) + 'all_wait_extrac_chunks长度：' + repr(len(all_wait_extrac_chunks)))
        master_control_logger.info(repr(file_name) + 'all_wait_extrac_chunks长度：' + repr(len(all_wait_extrac_chunks)))
        logger.info('all_wait_extrac_chunks 获取完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.info('all_wait_extrac_chunks 获取完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    except Exception as e:
        import traceback
        logger.error(traceback.format_exc())
        logger.error(repr(e))
        logger.error('all_wait_extrac_chunks 获取失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.error(
            'all_wait_extrac_chunks 获取失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=101)
        return


    # -------------- Extract and construct the map data from the segmented chunks --------------
    # Group and extract chunks by batch_size
    all_graph_chunks = []
    all_graph_vocabulary_set = set()
    batch_size = 10
    if enable_knowledge_graph:
        schema = {}
        # When graph_schema_filename and graph_schema_objectname have values, it means that the user has uploaded excel himself. Otherwise, if the schema is empty, the built-in schema will be used to extract it later.
        if graph_schema_filename and graph_schema_objectname:
            try:
                schema_file_path = os.path.join(filepath, graph_schema_filename)
                graph_download_status, graph_download_link = minio_utils.get_file_from_minio(graph_schema_objectname,
                                                                                             schema_file_path)
                logger.info("graph_download_status=%s,graph_download_link=%s" %
                            (graph_download_status, graph_download_link))
                schema = graph_utils.parse_excel_to_schema_json(schema_file_path)
                logger.info(f'提取graph schema成功'
                             + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + str(schema))
                master_control_logger.info(f'提取graph schema成功'
                    + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            except Exception as e:
                logger.error(repr(e))
                logger.error(f'提取graph schema失败'
                             + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                master_control_logger.error(f'提取graph schema失败'
                                            + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
                mq_rel_utils.update_doc_status(file_id, status=104)
                return

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
                master_control_logger.info(f'第{batch_num}批文档提取graph数据成功'
                    + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            except Exception as e:
                logger.error(repr(e))
                logger.error(f'第{batch_num}批文档提取graph数据失败'
                             + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                master_control_logger.error(f'第{batch_num}批文档提取graph数据失败'
                    + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
                mq_rel_utils.update_doc_status(file_id, status=102)
                return

    # --------------  insert es graph_data ----------------
    try:
        if enable_knowledge_graph and len(all_graph_chunks) > 0:
            logger.info(f'graph_data 插入es开始,all_graph_chunks len:{len(all_graph_chunks)}')
            master_control_logger.info(f'graph_data 插入es开始,all_graph_chunks len:{len(all_graph_chunks)}')
            insert_es_result = es_utils.add_es(user_id, kb_name, all_graph_chunks, file_name, kb_id=kb_id)
            logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
            master_control_logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
            if insert_es_result['code'] != 0:
                # callback
                logger.error('graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                master_control_logger.error(
                    'graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                mq_rel_utils.update_doc_status(file_id, status=103)
                return
            else:
                # After the insertion is successful, update the update_graph_vocabulary_set data
                kb_id = knowledge_base_utils.get_kb_name_id(user_id, kb_name)
                redis_utils.update_graph_vocabulary_set(graph_redis_client, kb_id,
                                                        elements_to_add=all_graph_vocabulary_set)
                # callback
                logger.info('graph_data插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
                master_control_logger.info(
                    'graph_data插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    except Exception as e:
        logger.error(repr(e))
        logger.error('graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        master_control_logger.error(
            'graph_data插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + repr(e))
        mq_rel_utils.update_doc_status(file_id, status=103)
        return

    # ---------------7. Final completion
    # callback
    logger.info("user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name) + '===== 文档grahp解析成功且完成')
    master_control_logger.info("user_id=%s,kb_name=%s,file_name=%s,kb_id=%s" % (user_id, kb_name, file_name, kb_id) + '===== 文档grahp解析成功且完成')
    mq_rel_utils.update_doc_status(file_id, status=100)


def generate_community_report(user_id, kb_name, enable_knowledge_graph, graph_model_id="", kb_id=""):
    # Community report starts to be generated
    mq_rel_utils.update_kb_status(kb_id, status=130)

    # Clean up old community reports
    try:
        clear_result = milvus_utils.del_community_reports(user_id, kb_name, clear_reports=True, kb_id=kb_id)
        if clear_result['code'] != 0:
            raise RuntimeError(clear_result["message"])
        logger.info(f'清理社区报告成功'
                    + "user_id=%s,kb_name=%s" % (user_id, kb_name) + str(clear_result))
        master_control_logger.info(f"社区报告插入milvus成功, user_id=%s,kb_name=%s" % (user_id, kb_name))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"清理社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        master_control_logger.error(f"清理社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name) + repr(e))
        mq_rel_utils.update_kb_status(kb_id, status=124)
        return

    # Pull community reports
    try:
        reports_result = graph_utils.generate_community_reports(user_id, kb_name, graph_model_id)
        reports = reports_result['community_reports']
        if len(reports) == 0:
            raise ValueError("社区报告数量为0，生成失败")
        logger.info(f"生成社区报告成功, user_id=%s,kb_name=%s, size=%s" % (user_id, kb_name, len(reports)))
        master_control_logger.info(f"生成社区报告成功, user_id=%s,kb_name=%s, size=%s" % (user_id, kb_name, len(reports)))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"生成社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        master_control_logger.error(f"生成社区报告失败, user_id=%s,kb_name=%s" % (user_id, kb_name) + repr(e))
        mq_rel_utils.update_kb_status(kb_id, status=122)
        return

    # Store community reports
    try:
        chunk_current_num = 0
        sub_chunks = []
        chunk_total_num = len(reports)
        file_name = "社区报告"
        for report_data in reports:
            sub_chunks.append({
                "content": report_data["report"],
                "embedding_content": report_data["report_title"],
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
        master_control_logger.info(f"社区报告插入milvus开始, user_id=%s,kb_name=%s" % (user_id, kb_name))
        insert_milvus_result = milvus_utils.add_milvus(user_id, kb_name, sub_chunks, file_name, "",
                                                       milvus_url=milvus_utils.ADD_COMMUNItY_REPORT_URL)
        if insert_milvus_result['code'] != 0:
            raise RuntimeError(insert_milvus_result["message"])
        logger.info(f'社区报告插入milvus成功'
                    + "user_id=%s,kb_name=%s" % (user_id, kb_name) + str(insert_milvus_result))
        master_control_logger.info(f"社区报告插入milvus成功, user_id=%s,kb_name=%s" % (user_id, kb_name))
    except Exception as e:
        logger.error(repr(e))
        logger.error(f"社区报告插入milvus失败, user_id=%s,kb_name=%s" % (user_id, kb_name))
        master_control_logger.error(f"社区报告插入milvus失败, user_id=%s,kb_name=%s" % (user_id, kb_name) + repr(e))
        mq_rel_utils.update_kb_status(kb_id, status=123)
        return

    # finally completed
    logger.info("user_id=%s,kb_name=%s" % (user_id, kb_name) + '===== 社区报告生成且存储完成')
    master_control_logger.info("user_id=%s,kb_name=%s,kb_id=%s" % (user_id, kb_name, kb_id) + '===== 社区报告生成且存储完成')
    mq_rel_utils.update_kb_status(kb_id, status=120)

if __name__ == "__main__":
    kafkal()
