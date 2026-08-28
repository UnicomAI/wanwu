import os
import json
import logging
from typing import Optional
from utils import es_utils, rerank_utils
from model_manager.model_config import get_model_configure

logger = logging.getLogger(__name__)

def search_qa_base(question, top_k, threshold=0.0, return_meta=False, retrieve_method="hybrid_search",
                   rerank_model_id='', rerank_mod="rerank_model",weights: Optional[dict] | None = None,
                   metadata_filtering_conditions=[], qa_base_info={}):
    """ qa_base_info: {"user_id1": [{ "qa_base_id": "","qa_base_name": ""}]}"""
    response_info = {'code': 0, "message": "成功", "data": {"searchList": [], "score": []}}

    try:
        if top_k == 0:
            return response_info

        duplicate_set = set()
        qa_result_list = []
        search_list_infos = {}

        for user_id, qa_base_name_id_list in qa_base_info.items():
            search_result = None
            if retrieve_method in {"semantic_search", "hybrid_search"}:
                search_result = es_utils.vector_search(user_id, qa_base_name_id_list, question, top_k, threshold=threshold,
                                                       metadata_filtering_conditions = metadata_filtering_conditions)

            if retrieve_method in {"full_text_search", "hybrid_search"}:
                search_result = es_utils.full_text_search(user_id, qa_base_name_id_list, question, top_k,threshold=threshold, metadata_filtering_conditions=metadata_filtering_conditions)

            search_result_str = json.dumps(search_result, ensure_ascii=False)
            logger.info(f"问题问答库查询结果：查询类型：{retrieve_method}, user_id: {user_id}, qa_base_name_id_list: {qa_base_name_id_list}, question: {question}, search_result: {search_result_str}")
            if search_result['code'] != 0:
                raise RuntimeError(search_result['message'])

            search_list = search_result['data']["search_list"]
            for qa_info in search_list:
                question_text = qa_info["question"]
                if question_text in duplicate_set:
                    continue
                duplicate_set.add(question_text)
                qa_info["title"] = "问答库"
                qa_info["content_type"] = "qa"
                qa_result_list.append(qa_info)

            search_list_infos[user_id] = {
                "base_infos": qa_base_name_id_list,
                "search_list": search_list
            }

        # reank重排
        if not qa_result_list:
            logger.info('qa_result_list is None 重排结果：' + json.dumps(repr(response_info),ensure_ascii=False))
            return response_info
        if rerank_mod == "rerank_model":
            model_config = get_model_configure(rerank_model_id)
            is_support_multimodal = model_config.is_multimodal
            query = question
            if is_support_multimodal:
                documents = [{"text": qa_info["question"]} for qa_info in qa_result_list]
                query = {"text": question}
            else:
                documents = [qa_info["question"] for qa_info in qa_result_list]
            rerank_result = rerank_utils.model_rerank(query,
                                                      top_k,
                                                      documents,
                                                      qa_result_list,
                                                      rerank_model_id,
                                                      model_config=model_config) #type: ignore
        elif rerank_mod == "weighted_score":
            rerank_result = rerank_utils.qa_weighted_rerank(question, weights, top_k, search_list_infos)
        else:
            raise Exception("rerank_mod is not valid")
        if rerank_result["code"] != 0:
            logger.warning(f"rerank failed, rerank method: {rerank_mod}, rerank result: {rerank_result}")
            raise RuntimeError(rerank_result["message"])
        sorted_scores = rerank_result['data']["sorted_scores"]
        sorted_search_list = rerank_result['data']["sorted_search_list"]

        if not return_meta:
            for x in sorted_search_list:
                if 'meta_data' in x: x['meta_data'] = {}
        res_score = []
        res_search_list = []
        for score, doc_item in zip(sorted_scores, sorted_search_list):
            if score >= threshold:
                res_score.append(score)
                doc_item["title"] = "问答库"
                doc_item["content_type"] = "qa"
                res_search_list.append(doc_item)
        response_info['data']['searchList'] = res_search_list
        response_info['data']['score'] = res_score

        logger.info('问答库重排结果：' + repr(res_search_list))
        return response_info
    except Exception as e:
        logger.warn(f"问答库查询失败, exception: {repr(e)}")
        response_info["code"] = 1
        response_info["message"] = str(e)
        return response_info