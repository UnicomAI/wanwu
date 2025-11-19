import os
import json
import time
import copy

import requests
import pandas as pd
from datetime import datetime

from utils import milvus_utils
from utils import es_utils
from utils import redis_utils
from utils import timing

from logging_config import setup_logging
from settings import GRAPH_SERVER_URL
from model_manager import get_model_configure, LlmModelConfig
from utils.tools import generate_md5

logger_name = 'rag_graph_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))


def parse_excel_to_schema_json(file_path):
    """
    解析 Excel 文件中的 '类目表' 和 '类目属性表'，输出指定 JSON 结构
    """
    schema = {}
    try:
        # Automatic inference engine using pd.read_excel (supports .xls and .xlsx)
        df_category = pd.read_excel(file_path, sheet_name='类目表')
        df_attribute = pd.read_excel(file_path, sheet_name='类目属性表')
        # Clean column names: remove spaces and newlines
        df_category.columns = df_category.columns.str.strip()
        df_attribute.columns = df_attribute.columns.str.strip()

        # === Analysis Category Table ===
        category_list = []
        for _, row in df_category.iterrows():
            item = {
                "类名": str(row["类名"]).strip() if pd.notna(row["类名"]) else "",
                "类描述": str(row["类描述"]).strip() if pd.notna(row["类描述"]) else ""
            }
            category_list.append(item)

        # === Parse category attribute table ===
        attribute_list = []

        for _, row in df_attribute.iterrows():
            class_name = str(row["类名"]).strip() if pd.notna(row["类名"]) else ""
            attr_name = str(row["属性/关系名"]).strip() if pd.notna(row["属性/关系名"]) else ""

            # Fix description field
            key = (class_name, attr_name)

            desc = str(row["属性/关系说明"]).strip() if pd.notna(row["属性/关系说明"]) else ""

            # Processing alias fields (supports multiple aliases separated by |)
            alias = row["别名(多别名以|隔开)"]
            if pd.isna(alias) or str(alias).strip() == "" or str(alias).lower() == "nan":
                alias_str = ""
            else:
                alias_str = str(alias).strip()

            value_type = str(row["值类型"]).strip() if pd.notna(row["值类型"]) else ""

            attribute_list.append({
                "类名": class_name,
                "属性/关系名": attr_name,
                "属性/关系说明": desc,
                "属性别名(多别名以|隔开)": alias_str,
                "值类型": value_type
            })

        # Build the final JSON structure
        schema = {
            "schema定义": {
                "类目表": category_list,
                "类目属性表": attribute_list
            }
        }
        logger.info("schema:%s" % json.dumps(schema, ensure_ascii=False))
    except Exception as e:
        import traceback
        logger.error(traceback.format_exc())
        logger.error(f"无法读取Excel文件或工作表不存在: {e}")
    return schema


@timing.timing_decorator(logger, include_args=False)
def get_extrac_graph_data(user_id, kb_name, chunks, file_name, graph_model_id, schema=None):
    """获取知识图谱数据"""
    try:
        llm_config = get_model_configure(graph_model_id)
        llm_model = llm_config.model_name
        llm_base_url = ""
        llm_api_key = ""
        if isinstance(llm_config, LlmModelConfig):
            llm_base_url = llm_config.endpoint_url + "/chat/completions"
            llm_api_key = llm_config.api_key
        start_time = datetime.now()
        headers = {
            "Content-Type": "application/json",
        }
        data = {
            "user_id": user_id,
            "kb_name": kb_name,
            "chunks": chunks,
            "schema": schema,
            "file_name": file_name,
            "llm_model": llm_model,
            "llm_base_url": llm_base_url,
            "llm_api_key": llm_api_key
        }
        # Convert JSON data into good format
        json_data = json.dumps(data)
        extract_graph_url = GRAPH_SERVER_URL + "/extrac_graph_data"
        response = requests.post(extract_graph_url, headers=headers, data=json_data, timeout=600)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            finish_time1 = datetime.now()
            time_difference1 = finish_time1 - start_time
            logger.info(f"extrac_graph_data -{extract_graph_url}: 请求成功 耗时：{time_difference1}")
            return result_data
        else:
            # If not 200, throw a custom exception
            raise Exception(f"{extract_graph_url} 请求失败，错误信息：" + response.text)
    except Exception as e:
        raise Exception("get_extrac_graph_data 发生异常：" + str(e))


@timing.timing_decorator(logger, include_args=False)
def generate_community_reports(user_id, kb_name, graph_model_id):
    """获取知识图谱社区报告"""
    try:
        llm_config = get_model_configure(graph_model_id)
        llm_model = llm_config.model_name
        llm_base_url = ""
        llm_api_key = ""
        if isinstance(llm_config, LlmModelConfig):
            llm_base_url = llm_config.endpoint_url + "/chat/completions"
            llm_api_key = llm_config.api_key
        start_time = datetime.now()
        headers = {
            "Content-Type": "application/json",
        }
        data = {
            "user_id": user_id,
            "kb_name": kb_name,
            "llm_model": llm_model,
            "llm_base_url": llm_base_url,
            "llm_api_key": llm_api_key
        }
        # Convert JSON data into good format
        json_data = json.dumps(data)
        community_url = GRAPH_SERVER_URL + "/generate_community_reports"
        response = requests.post(community_url, headers=headers, data=json_data, timeout=600)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            finish_time1 = datetime.now()
            time_difference1 = finish_time1 - start_time
            logger.info(f"generate_community_reports -{community_url}: 请求成功 耗时：{time_difference1}")
            return result_data
        else:
            # If not 200, throw a custom exception
            raise Exception(f"{community_url} 请求失败，错误信息：" + response.text)
    except Exception as e:
        raise Exception("generate_community_reports 发生异常：" + str(e))


@timing.timing_decorator(logger, include_args=False)
def delete_file_from_graph(user_id, kb_name, file_name):
    """知识图谱删除文件"""
    try:
        start_time = datetime.now()
        headers = {
            "Content-Type": "application/json",
        }
        data = {
            "user_id": user_id,
            "kb_name": kb_name,
            "file_name": file_name
        }
        # Convert JSON data into good format
        json_data = json.dumps(data)
        delete_file_url = GRAPH_SERVER_URL + "/delete_file"
        response = requests.post(delete_file_url, headers=headers, data=json_data, timeout=600)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            finish_time1 = datetime.now()
            time_difference1 = finish_time1 - start_time
            logger.info(f"graph delete_file -{delete_file_url}: 请求成功 耗时：{time_difference1}")
            return result_data
        else:
            # If not 200, throw a custom exception
            raise Exception(f"{delete_file_url} 请求失败，错误信息：" + response.text)
    except Exception as e:
        raise Exception("graph delete_file 发生异常：" + str(e))


@timing.timing_decorator(logger, include_args=False)
def delete_kb_graph(user_id, kb_name):
    """知识图谱删除"""
    try:
        start_time = datetime.now()
        headers = {
            "Content-Type": "application/json",
        }
        data = {
            "user_id": user_id,
            "kb_name": kb_name,
        }
        # Convert JSON data into good format
        json_data = json.dumps(data)
        delete_kb_url = GRAPH_SERVER_URL + "/delete_kb"
        response = requests.post(delete_kb_url, headers=headers, data=json_data, timeout=600)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            finish_time1 = datetime.now()
            time_difference1 = finish_time1 - start_time
            logger.info(f"graph delete_kb_graph -{delete_kb_url}: 请求成功 耗时：{time_difference1}")
            return result_data
        else:
            # If not 200, throw a custom exception
            raise Exception(f"{delete_kb_url} 请求失败，错误信息：" + response.text)
    except Exception as e:
        raise Exception("graph delete_kb 发生异常：" + str(e))


@timing.timing_decorator(logger, include_args=False)
def get_kb_graph_data(user_id, kb_name, kb_id):
    """获取知识图谱数据"""
    try:
        start_time = datetime.now()
        headers = {
            "Content-Type": "application/json",
        }
        data = {
            "user_id": user_id,
            "kb_name": kb_name,
            "kb_id": kb_id
        }
        # Convert JSON data into good format
        json_data = json.dumps(data)
        graph_url = GRAPH_SERVER_URL + "/get_kb_graph_data"
        response = requests.post(graph_url, headers=headers, data=json_data, timeout=600)
        if response.status_code == 200:
            result_data = json.loads(response.text)
            finish_time1 = datetime.now()
            time_difference1 = finish_time1 - start_time
            if not result_data["success"]:
                raise RuntimeError(result_data["message"])
            logger.info(f"extrac_graph_data -{graph_url}: 请求成功 耗时：{time_difference1}")
            return result_data["graph_data"]
        else:
            # If not 200, throw a custom exception
            raise Exception(f"{graph_url} 请求失败，错误信息：" + response.text)
    except Exception as e:
        raise Exception("get_kb_graph_data 发生异常：" + str(e))


def get_graph_vocabulary_set(kb_ids: list):
    """处理获取知识图谱实体词表数据"""
    graph_redis_client = redis_utils.get_redis_connection()
    kb_graph_vocabulary_list = []
    for kb_id in kb_ids:
        kb_graph_vocabulary_set = redis_utils.query_graph_vocabulary_set(graph_redis_client, kb_id)
        res_vocabulary_list = [v.split("|||schema_type:")[0] for v in kb_graph_vocabulary_set]
        res_vocabulary_type_list = [v.split("|||schema_type:")[1] for v in kb_graph_vocabulary_set]
        kb_graph_vocabulary_list.append((res_vocabulary_list, res_vocabulary_type_list))
    # Process the returned results
    return kb_graph_vocabulary_list


def get_all_extrac_graph_chunks(user_id, kb_name, file_name, kb_id=""):
    """
    获取用户知识库中对应文件的所有chunk
    """

    try:
        time.sleep(10)  # Wait 10s first
        retry_num = 0
        all_chunks = []
        chunk_total_num = 1
        page_size = 100
        search_after = 0
        complete_flag = True
        while len(all_chunks) < chunk_total_num and complete_flag:
            response_info = milvus_utils.get_milvus_file_content_list(user_id, kb_name, file_name, page_size,
                                                                      search_after, kb_id=kb_id)
            temp_content_list = response_info["data"]["content_list"]
            if not temp_content_list:  # If you can't get it, try again or set it to complete.
                retry_num += 1
                time.sleep(10)
                if retry_num >= 5 or len(all_chunks) >= chunk_total_num:
                    complete_flag = False
            else:
                if chunk_total_num == 1:
                    chunk_total_num = temp_content_list[0]["meta_data"]["chunk_total_num"]
                for doc in temp_content_list:  # Organize the format and add it
                    all_chunks.append({
                        "title": doc["file_name"],
                        "snippet": doc["content"],
                        "source_type": "RAG_KB",
                        "meta_data": doc["meta_data"]
                    })
                search_after += 100
        # ======== Return directly after taking out =========
        return all_chunks
    except Exception as e:
        logger.info(f"get_all_extrac_graph_chunks Error: {e}")
        return []


def update_community_reports(user_id: str, kb_name: str, report:dict, kb_id = ""):
    """
    更新community report
    """
    logger.info(f"========= update_community_reports start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, report: {report}")

    response_info = {
        "code": 1,
        "message": "",
    }

    old_content_id = report["report_id"]
    chunk = {
        "content": report["content"],
        "embedding_content": report["title"],
    }

    report_response = milvus_utils.get_content_by_ids(user_id, kb_name, [old_content_id],
                                                       content_type="community_report", kb_id=kb_id)
    logger.info(f"content_id: {old_content_id}, 社区报告结果: {report_response}")
    if report_response['code'] != 0:
        logger.error(f"获取community report信息失败， user_id: {user_id},kb_name: {kb_name}, content_id: {old_content_id}")
        response_info["message"] = report_response["message"]
        return response_info

    old_report = report_response["data"]["contents"][0]
    chunk_current_num = old_report["meta_data"]["chunk_current_num"]
    chunk["meta_data"] = copy.deepcopy(old_report["meta_data"])
    chunk["create_time"] = old_report["create_time"]
    if not kb_id:  # If kb_id is empty, get kb_id based on kb_name
        kb_id = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)  # Get kb_id

    file_name = "社区报告"
    content_str = kb_id + chunk["content"] + file_name + str(chunk_current_num)
    new_content_id = generate_md5(content_str)
    if new_content_id != old_content_id:
        chunks = [chunk]
        logger.info('新增report插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        insert_report_result = milvus_utils.add_milvus(user_id, kb_name, chunks, file_name, "",
                                                       kb_id=kb_id, milvus_url=milvus_utils.ADD_COMMUNItY_REPORT_URL)
        if insert_report_result['code'] != 0:
            logger.error(
                '新增report插入milvus失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
            response_info["message"] = insert_report_result["message"]
            #New data rollback
            milvus_utils.del_community_reports(user_id, kb_name, content_ids=[new_content_id], kb_id=kb_id)
            return response_info
        else:
            logger.info('新增report插入milvus完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

        #Clean old data
        milvus_utils.del_community_reports(user_id, kb_name, content_ids=[old_content_id], kb_id=kb_id)
    logger.info(f"========= update_community_reports end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk: {chunk}")

    response_info["code"] = 0
    response_info["message"] = "success"
    return response_info


def batch_add_community_reports(user_id: str, kb_name: str, reports:list, kb_id: str = ""):
    """
    根据file name 新增reports
    """
    logger.info(f"========= batch_add_community_reports start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, reports: {reports}")

    chunks = []
    for item in reports:
        chunks.append({
            "content": f"# {item['title']} \n\n {item['content']}",
            "embedding_content": item["title"],
            "create_time": str(int(time.time() * 1000))
        })

    response_info = {
        "code": 1,
        "message": "",
        "data": {
            "success_count": 0
        }
    }
    file_name = "社区报告"
    allocate_report_result = es_utils.allocate_chunks(user_id, kb_name, file_name, len(chunks), chunk_type="community_report", kb_id=kb_id)
    logger.info(repr(file_name) + '新增reports分配结果：' + repr(allocate_report_result))
    if allocate_report_result['code'] != 0:
        logger.error('新增reports分配chunk失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        response_info["message"] = allocate_report_result["message"]
        return response_info
    else:
        chunk_total_num = allocate_report_result["data"]["chunk_total_num"]
        meta_data = allocate_report_result["data"]["meta_data"]
        current_chunk_num = chunk_total_num - len(chunks) + 1
        if not kb_id:  # If kb_id is empty, get kb_id based on kb_name
            kb_id = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)  # Get kb_id
        for chunk in chunks:
            chunk["meta_data"] = copy.deepcopy(meta_data)
            chunk["meta_data"]["chunk_current_num"] = current_chunk_num
            chunk["meta_data"]["entities"] = []
            current_chunk_num += 1
        logger.info('新增reports分配chunk完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    logger.info('新增reports插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    insert_reports_result = milvus_utils.add_milvus(user_id, kb_name, chunks, file_name, "",
                                                   kb_id=kb_id, milvus_url=milvus_utils.ADD_COMMUNItY_REPORT_URL)
    if insert_reports_result['code'] != 0:
        logger.error(
            '新增reports插入milvus失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        response_info["message"] = insert_reports_result["message"]
        return response_info
    else:
        logger.info('新增reports插入milvus完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    logger.info(f"========= batch_add_community_reports end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, reports: {reports}")

    response_info["code"] = 0
    response_info["data"]["success_count"] = len(chunks)
    return response_info

def batch_delete_community_reports(user_id: str, kb_name: str, report_ids: list[str], kb_id=""):
    """
    根据report_ids 删除分片reports
    """
    logger.info(f"========= batch_delete_community_reports start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, report_ids: {report_ids}")
    response_info = milvus_utils.del_community_reports(user_id, kb_name, content_ids=report_ids, kb_id=kb_id)
    logger.info(f"========= batch_delete_community_reports end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, report_ids: {report_ids}")

    return response_info

def get_community_report_list(user_id: str, kb_name: str, page_size: int, search_after: int, kb_id=""):
    """
    获取知识库reports片段列表,用于分页展示
    """
    logger.info(f"get_community_report_list start: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"page_size:{page_size}, search_after:{search_after}")
    file_name = "社区报告"
    response_info = milvus_utils.get_milvus_file_content_list(user_id, kb_name, file_name, page_size, search_after,
                                                              kb_id=kb_id, content_type="community_report")
    if response_info["code"] == 0:
        content_list = response_info["data"]["content_list"]
        for content_info in content_list:
            content_info["report_title"] = content_info["embedding_content"]
            content_info.pop("embedding_content")
    logger.info(f"get_community_report_list end: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"page_size:{page_size}, search_after:{search_after}, response: {response_info}")
    return response_info

@timing.timing_decorator(logger, include_args=True)
def get_graph_search_list(user_id, kb_names, question, top_k, kb_ids=[], filter_file_name_list=[]):
    """ 根据问题召回知识图谱的 search列表"""
    # Use query to recall graph SPO information
    try:
        if not kb_ids:
            for kb_n in kb_names:
                kb_ids.append(milvus_utils.get_milvus_kb_name_id(user_id, kb_n))  # Get kb_id
        kb_graph_vocabulary_list = get_graph_vocabulary_set(kb_ids)
        graph_node_query = ""
        entities = []
        for kb_vocabulary_list, kb_vocabulary_type_list in kb_graph_vocabulary_list:
            kb_entities = []
            for vocabulary in kb_vocabulary_list:
                if vocabulary in question:
                    if len(vocabulary) > 3:
                        kb_entities.append(vocabulary)
                    graph_node_query += vocabulary
            entities.append(kb_entities)
        if not graph_node_query:
            graph_node_query = question
        search_top_k = 100
        es_graph_search_list = es_utils.search_graph_es(user_id, kb_names, graph_node_query, search_top_k, kb_ids=kb_ids,
                                            filter_file_name_list=filter_file_name_list)
        graph_list = []
        report_topk = min(2, int(top_k*0.4))
        community_report_result = milvus_utils.search_milvus(user_id, kb_names, report_topk, question, threshold=0,
                                                   search_field="content", kb_ids=kb_ids, milvus_url=milvus_utils.KNN_COMMUNITY_SEARCH_URL)
        logger.info(f"search report done, user_id:{user_id}, kb_names: {kb_names}, report_topk: {report_topk}, "
                    f"entities: {entities}, reports: {community_report_result}")
        if community_report_result["code"] == 0:
            search_list = community_report_result['data']['search_list']
            contents = []
            for s in search_list:
                contents.append(s["content"])
            if contents:
                newline = '\n'
                report_texts = f"社区报告信息:{newline}{newline.join(contents)} "
                graph_list.append({"snippet": report_texts, "meta_data": {},
                                          "title": "知识图谱-社区报告", "content_type": "community_report"})
        if not all([not(ent) for ent in entities]):  # If there are image keywords, perform priority community report retrieval.
            # ======= Build triple_text to generate a chunk and insert it into the beginning of the community report =======
            triple_text_list = []
            for s in es_graph_search_list:
                # ====== SPO triple preprocessing =======
                text = s["graph_data_text"]
                # Check whether the string contains Chinese and contains has_attribute
                if "has_attribute" in text and any('\u4e00' <= char <= '\u9fff' for char in text):
                    s["graph_data_text"] = text.replace("has_attribute", "其具有属性")
                for kb_entities in entities:  # If there are image keywords, perform SPO pull
                    for kb_entity in kb_entities:
                        if kb_entity in s["graph_data_text"] and s["graph_data_text"] not in triple_text_list:
                            triple_text_list.append(s["graph_data_text"])
            if triple_text_list:
                triple_text = f"知识图谱信息:({'|'.join(triple_text_list)}) "
                graph_list.append({"snippet": triple_text, "meta_data": {},
                                          "title": "知识图谱-实体属性关系", "content_type": "graph"})

        logger.info(repr(user_id) + repr(kb_names) + repr(question)
                    + f'问题 graph 查询结果 es_graph_search_list len：{len(es_graph_search_list)},graph_list len：{len(graph_list)}')
        # ====== Remove duplicates =======
        tmp_content = []
        graph_search_list = []
        for i in es_graph_search_list:  # Remove duplicates
            i["snippet"] = i["meta_data"]["reference_snippet"]
            if i["snippet"] in tmp_content:
                continue
            graph_search_list.append(i)
            tmp_content.append(i["snippet"])
    except Exception as err:
        import traceback
        logger.error("====> get_graph_search_list error %s" % err)
        logger.error(traceback.format_exc())
        graph_search_list = []
        graph_list = []
    res_graph_search_list = graph_search_list[:top_k*2]
    logger.info(repr(user_id) + repr(kb_names) + repr(question) + f'问题 graph 最终查询分段结果：'
                + repr(res_graph_search_list) + f'graph_list:'+ repr(graph_list))
    return res_graph_search_list, graph_list

