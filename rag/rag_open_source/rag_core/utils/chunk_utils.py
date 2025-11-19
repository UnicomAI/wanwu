import os
import copy

from logging_config import setup_logging
from utils.tools import generate_md5
from utils import milvus_utils
from utils import es_utils
from utils import file_utils
from utils import redis_utils

logger_name = 'rag_chunk_utils'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name, logger_name)
logger.info(logger_name + '---------LOG_FILE：' + repr(app_name))

chunk_label_redis_client = redis_utils.get_redis_connection(redis_db=5)

def update_chunk_labels(user_id: str, kb_name: str, file_name: str, chunk_id: str, labels: list[str], kb_id=""):
    """
    根据file name和chunk id更新标签
    """
    logger.info(f"========= update_chunk_labels start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, labels: {labels}")
    response_info = milvus_utils.update_chunk_labels(user_id, kb_name, file_name, chunk_id, labels, kb_id=kb_id)
    logger.info(f"========= update_chunk_labels end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, labels: {labels}")

    return response_info


def save_chunks(user_id:str, kb_name:str, file_name:str, chunks:list, sub_chunks:list, kb_id:str=""):
    response_info = {
        "code": 1,
        "message": "",
        "data": {
            "success_count": 0
        }
    }
    # -------------insert vector
    logger.info('新增分段插入milvus开始' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
    insert_milvus_result = milvus_utils.add_milvus(user_id, kb_name, sub_chunks, file_name, "", kb_id=kb_id)
    logger.info(repr(file_name) + '新增分段添加milvus结果：' + repr(insert_milvus_result))
    if insert_milvus_result['code'] != 0:
        logger.error('新增分段插入milvus失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        response_info["message"] = insert_milvus_result["message"]
        return response_info
    else:
        logger.info('新增分段插入milvus完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    # --------------insert text
    logger.info('文档插入es开始')
    insert_es_result = es_utils.add_es(user_id, kb_name, chunks, file_name, kb_id=kb_id)
    logger.info(repr(file_name) + '添加es结果：' + repr(insert_es_result))
    if insert_es_result['code'] != 0:
        logger.error('文档插入es失败' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        response_info["message"] = insert_es_result["message"]
        return response_info
    else:
        logger.info('文档插入es完成' + "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    response_info["code"] = 0
    response_info["data"]["success_count"] = len(chunks)

    return response_info


def batch_add_chunks(user_id: str, kb_name: str, file_name: str, max_sentence_size: int, chunk_infos: list[dict],
                     split_type: str = "common", child_chunk_config: dict = None, kb_id: str = ""):
    """
    根据file name 新增chunks
    """
    logger.info(f"========= batch_add_chunks start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, max_sentence_size: {max_sentence_size}, chunks: {chunk_infos}")

    chunks = []
    for item in chunk_infos:
        chunks.append({
            "text": item["content"],
            "labels": item["labels"],
        })

    response_info = {
        "code": 1,
        "message": "",
        "data": {
            "success_count": 0
        }
    }

    allocate_chunk_result = es_utils.allocate_chunks(user_id, kb_name, file_name, len(chunks), kb_id=kb_id)
    logger.info(repr(file_name) + '新增分段分配chunk结果：' + repr(allocate_chunk_result))
    if allocate_chunk_result['code'] != 0:
        logger.error('新增分段分配chunk失败'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))
        response_info["message"] = allocate_chunk_result["message"]
        return response_info
    else:
        chunk_total_num = allocate_chunk_result["data"]["chunk_total_num"]
        meta_data = allocate_chunk_result["data"]["meta_data"]
        current_chunk_num = chunk_total_num - len(chunks) + 1
        if not kb_id:  # kb_id为空，则根据kb_name获取kb_id [EN] If kb_id is empty, get kb_id based on kb_name
            kb_id = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
        for chunk in chunks:
            chunk["meta_data"] = copy.deepcopy(meta_data)
            chunk["meta_data"]["chunk_current_num"] = current_chunk_num
            if chunk["labels"]:
                content_str = kb_id + chunk["text"] + file_name + str(current_chunk_num)
                content_id = generate_md5(content_str)
                redis_utils.update_chunk_labels(chunk_label_redis_client, kb_id, file_name, content_id, chunk["labels"])
            current_chunk_num += 1
        logger.info('新增分段分配chunk完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    if split_type == "parent_child":
        chunks, sub_chunks = file_utils.split_child_chunks(chunks, child_chunk_config)
    else:
        sub_chunks = file_utils.split_doc(chunks, max_sentence_size)

    response_info = save_chunks(user_id, kb_name, file_name,chunks, sub_chunks, kb_id= kb_id)
    logger.info(f"========= batch_add_chunks end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, max_sentence_size: {max_sentence_size}, chunks: {chunk_infos}")

    return response_info


def batch_add_child_chunks(user_id: str, kb_name: str, file_name: str, chunk_id: str, child_contents:list, kb_id: str = ""):
    """
    新增子chunks
    """
    logger.info(f"========= batch_add_child_chunks start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, child_contents: {child_contents}")

    response_info = {
        "code": 1,
        "message": "",
    }
    chunks = []
    sub_chunks = []

    allocate_child_chunk_result = es_utils.allocate_child_chunks(user_id, kb_name, file_name, chunk_id, len(child_contents), kb_id=kb_id)
    logger.info(repr(file_name) + '新增子分段分配chunk结果：' + repr(allocate_child_chunk_result))
    if allocate_child_chunk_result['code'] != 0:
        logger.error('新增子分段分配chunk失败'+ "user_id=%s,kb_name=%s,file_name=%s,chunk_id=%s" % (user_id, kb_name, file_name,chunk_id))
        response_info["message"] = allocate_child_chunk_result["message"]
        return response_info
    else:
        child_chunk_total_num = allocate_child_chunk_result["data"]["child_chunk_total_num"]
        parent_content = allocate_child_chunk_result["data"]["content"]
        meta_data = allocate_child_chunk_result["data"]["meta_data"]
        child_chunk_current_num = child_chunk_total_num - len(child_contents) + 1
        if not kb_id:  # kb_id为空，则根据kb_name获取kb_id [EN] If kb_id is empty, get kb_id based on kb_name
            kb_id = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
        for child_content in child_contents:
            copy_meta_data = copy.deepcopy(meta_data)
            copy_meta_data["child_chunk_current_num"] = child_chunk_current_num
            copy_meta_data["child_chunk_total_num"] = child_chunk_total_num

            sub_chunks.append({
                'content': parent_content,
                'embedding_content': child_content,
                'meta_data': copy_meta_data,
                "is_parent": False
            })

            chunks.append({
                "text": child_content,
                "parent_text": parent_content,
                'meta_data': copy_meta_data,
                "is_parent": False
            })

            child_chunk_current_num += 1
        logger.info('新增子分段分配chunk完成'+ "user_id=%s,kb_name=%s,file_name=%s" % (user_id, kb_name, file_name))

    response_info = save_chunks(user_id, kb_name, file_name,chunks, sub_chunks, kb_id= kb_id)
    logger.info(f"========= batch_add_child_chunks end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, child_contents: {child_contents}")

    return response_info

def update_chunk(user_id: str, kb_name: str, file_name: str, max_sentence_size: int, chunk_info: dict,
                 split_type: str = "common", child_chunk_config: dict = None, kb_id = ""):
    """
    根据file name和chunk信息更新分段
    """
    logger.info(f"========= update_chunk start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_info: {chunk_info}")

    response_info = {
        "code": 1,
        "message": "",
    }

    old_content_id = chunk_info["chunk_id"]
    chunk = {
        "text": chunk_info["content"],
    }

    content_response = milvus_utils.get_content_by_ids(user_id, kb_name, [old_content_id], kb_id=kb_id)
    logger.info(f"content_id: {old_content_id}, 分段信息结果: {content_response}")
    if content_response['code'] != 0:
        logger.error(f"获取分段信息失败， user_id: {user_id},kb_name: {kb_name}, file_name: {file_name}, content_id: {old_content_id}")
        response_info["message"] = content_response["message"]
        return response_info

    old_content = content_response["data"]["contents"][0]
    chunk_current_num = old_content["meta_data"]["chunk_current_num"]
    status = old_content["status"]

    chunk["meta_data"] = copy.deepcopy(old_content["meta_data"])
    if 'labels' in old_content:
        chunk['labels'] = old_content['labels']

    if not kb_id:  # kb_id为空，则根据kb_name获取kb_id [EN] If kb_id is empty, get kb_id based on kb_name
        kb_id = milvus_utils.get_milvus_kb_name_id(user_id, kb_name)  # 获取kb_id [EN] Get kb_id
    content_str = kb_id + chunk["text"] + file_name + str(chunk_current_num)
    new_content_id = generate_md5(content_str)
    if new_content_id != old_content_id:
        chunks = [chunk]

        if split_type == "parent_child":
            chunks, sub_chunks = file_utils.split_child_chunks(chunks, child_chunk_config)
        else:
            sub_chunks = file_utils.split_doc(chunks, max_sentence_size)

        save_resp = save_chunks(user_id, kb_name, file_name, chunks, sub_chunks, kb_id=kb_id)
        if save_resp["code"] != 0:
            response_info["message"] = save_resp["message"]
            #新增数据回滚 [EN] New data rollback
            milvus_utils.batch_delete_chunks(user_id, kb_name, file_name, [new_content_id], kb_id=kb_id)
            return response_info

        #----------------update status
        logger.info('更新分段status开始')
        update_status_result = milvus_utils.update_milvus_content_status(user_id, kb_name, file_name, new_content_id, status,
                                                                  on_off_switch=None, kb_id=kb_id)
        logger.info(f"file_name: {file_name}, content_id: {new_content_id}, 更新分段status: {update_status_result}")
        if update_status_result['code'] != 0:
            logger.error(f"更新分段status失败, user_id: {user_id}, kb_name={kb_name}, file_name: {file_name}, content_id: {new_content_id}")
            response_info["message"] = update_status_result["message"]
            # 新增数据回滚 [EN] New data rollback
            milvus_utils.batch_delete_chunks(user_id, kb_name, file_name, [new_content_id], kb_id=kb_id)
            return response_info
        else:
            logger.info(f"更新分段status完成, user_id: {user_id}, kb_name: {kb_name}, file_name: {file_name}, content_id: {new_content_id}")

        #清理旧数据 [EN] Clean old data
        milvus_utils.batch_delete_chunks(user_id, kb_name, file_name, [old_content_id], kb_id=kb_id)
        if "labels" in chunk and chunk["labels"]:
            redis_utils.update_chunk_labels(chunk_label_redis_client, kb_id, file_name, new_content_id, chunk["labels"])
    logger.info(f"========= update_chunk end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk: {chunk}")

    response_info["code"] = 0
    response_info["message"] = "success"
    return response_info



def update_child_chunk(user_id: str, kb_name: str, file_name: str, chunk_id: str, chunk_current_num: int,
                       child_chunk: dict, kb_id = ""):
    """
    根据file name和chunk信息更新分段
    """
    logger.info(f"========= update_child_chunk start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, chunk_current_num: {chunk_current_num}, child_chunk: {child_chunk}")
    response_info = milvus_utils.update_child_chunk(user_id, kb_name, chunk_id, chunk_current_num, child_chunk)
    logger.info(f"========= update_child_chunk end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}")

    return response_info


def batch_delete_chunks(user_id: str, kb_name: str, file_name: str, chunk_ids: list[str], kb_id=""):
    """
    根据file name和chunk ids删除分片chunk
    """
    logger.info(f"========= batch_delete_chunks start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_ids: {chunk_ids}")
    response_info = milvus_utils.batch_delete_chunks(user_id, kb_name, file_name, chunk_ids, kb_id=kb_id)
    logger.info(f"========= batch_delete_chunks end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_ids: {chunk_ids}")

    return response_info


def batch_delete_child_chunks(user_id: str, kb_name: str, file_name: str, chunk_id: str, chunk_current_num: int,
                       child_chunk_current_nums: list[int], kb_id=""):
    """
    根据file name和chunk id删除子分片chunk
    """
    logger.info(f"========= batch_delete_child_chunks start：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}， chunk_current_num： {chunk_current_num}, "
                f"child_chunk_current_nums: {child_chunk_current_nums}")
    response_info = milvus_utils.batch_delete_child_chunks(user_id, kb_name, file_name, chunk_id, chunk_current_num, child_chunk_current_nums, kb_id=kb_id)
    logger.info(f"========= batch_delete_child_chunks end：user_id: {user_id}, kb_name: {kb_name}, kb_id: {kb_id}, "
                f"file_name: {file_name}, chunk_id: {chunk_id}, response_info: {response_info}")

    return response_info

