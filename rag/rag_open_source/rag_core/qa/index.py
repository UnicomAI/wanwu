# -*- coding: utf-8 -*-

from utils import es_utils
from typing import List, Dict, Any


# ---------------------- 1. 问答库生命周期 ----------------------
def init_qa_base(user_id: str, qa_base: str, qa_id: str, embedding_model_id: str) -> Dict[str, Any]:
    """
    创建问答库
    """
    if not user_id or not qa_base or not qa_id or not embedding_model_id:
        return {"code": 1, "message": "缺失必填参数"}

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.init_qa_base(user_id, qa_info, embedding_model_id)


def delete_qa_base(user_id: str, qa_base: str, qa_id: str) -> Dict[str, Any]:
    """
    删除问答库
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.del_qa_base(user_id, qa_info)


# ---------------------- 2. 问答对 CRUD ----------------------

def batch_add_qas(user_id: str, qa_base: str, qa_id: str, qa_pairs: List[Dict[str, str]]) -> Dict[str, Any]:
    """
    批量新增问答对
    qa_pairs: [{"qa_pair_id":"123","question":"q","answer":"a"}, ...]
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}
    if not qa_pairs or not isinstance(qa_pairs, list):
        raise ValueError("qa_pairs must be a list and not empty")

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.add_qas(user_id, qa_info, qa_pairs)


def get_qa_list(user_id: str, qa_base: str, qa_id: str, page_size: int, search_after: int):
    """
    分页获取问答对（冗余列表）
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.get_qa_list(user_id, qa_info, page_size, search_after)


def update_qa(user_id: str, qa_base: str, qa_id: str, qa_pair):
    """
    批量更新问答对（全量覆盖）
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    if not qa_pair or not isinstance(qa_pair, dict):
        raise ValueError("qa_pair must be a dict and not empty")

    if "qa_pair_id" not in qa_pair:
        raise ValueError("qa_pair_id must exist in qa_pair")

    if "question" not in qa_pair and "answer" not in qa_pair:
        raise ValueError("question or answer should be in qa_pair")

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.update_qa_data(user_id, qa_info, qa_pair["qa_pair_id"], qa_pair)


def batch_delete_qas(user_id: str, qa_base: str, qa_id: str, qa_pair_ids: List[str]) -> Dict[str, Any]:
    """
    批量删除问答对
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    if not qa_pair_ids or not isinstance(qa_pair_ids, list):
        raise ValueError("qa_pair_ids must be a list and not empty")

    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.del_qas(user_id, qa_info, qa_pair_ids)


def update_qa_status(user_id: str, qa_base: str, qa_id: str, qa_pair_id: str, status: bool) -> Dict[str, Any]:
    """
    启停单个问答对
    """
    if not user_id or not qa_base or not qa_id or not qa_pair_id:
        return {"code": 1, "message": "缺失必填参数"}

    update_data = {
        "qa_pair_id": qa_pair_id,
        "status": status
    }
    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.update_qa_data(user_id, qa_info, qa_pair_id, update_data)


# ---------------------- 3. 元数据管理 ----------------------

def update_qa_metas(user_id: str, qa_base: str, qa_id: str, metas: List[Dict[str, Any]]) -> Dict[str, Any]:
    """
    全量覆盖更新元数据
    metas: [{"qa_pair_id":"123","metadata_list":[...]}, ...]
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    if not metas or not isinstance(metas, list):
        raise ValueError("metas must be a list and not empty")
    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.update_qa_metas(user_id, qa_info, metas, "update_metas")


def delete_meta_by_keys(user_id: str, qa_base: str, qa_id: str, keys: List[str]) -> Dict[str, Any]:
    """
    批量删除指定 key 的元数据（跨所有问答对）
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    if not keys or not isinstance(keys, list):
        raise ValueError("keys must be a list and not empty")
    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.update_qa_metas(user_id, qa_info, keys, "delete_keys")


def rename_meta_keys(user_id: str, qa_base: str, qa_id: str, mappings: List[Dict[str, str]]) -> Dict[str, Any]:
    """
    批量重命名元数据 key
    mappings: [{"old_key":"company","new_key":"organization"}, ...]
    """
    if not user_id or not qa_base or not qa_id:
        return {"code": 1, "message": "缺失必填参数"}

    if not mappings or not isinstance(mappings, list):
        raise ValueError("mappings must be a list and not empty")
    qa_info = {"QABase": qa_base, "QAId": qa_id}
    return es_utils.update_qa_metas(user_id, qa_info, mappings, "rename_keys")
