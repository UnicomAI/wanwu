from rank_bm25 import BM25Okapi
import datetime
import json
import logging
from utils.model_loader import ModelLoader
from utils.timing import advanced_timing_decorator

logger = logging.getLogger(__name__)

# Get a preloaded tokenizer instance
model_loader = ModelLoader.get_instance()
ANALYZER = model_loader.build_default_analyzer(language='zh')

@advanced_timing_decorator(task_name="rerank_by_bm25")
def rerank_by_bm25(query, raw_search_list, top_k, threshold, sort_enable=True):
    """
    根据BM25分数对文档列表进行排序和过滤。
    
    参数:
    - query (str): 查询字符串。
    - raw_search_list (list of dict): 包含 'snippet' 的字典列表。
    - top_k (int): 返回的最高分结果数量。
    - threshold (float): 分数阈值，只返回高于此阈值的结果。
    
    返回:
    - 字典，包含分数sorted_scores和对应文本列表sorted_search_list。
    """
    
    if len(raw_search_list) == 0:
        print("没有文档可用于重新排序。")
        return {'sorted_scores': [], 'sorted_search_list': []}
    
    # Word segmentation query
    tokenized_query = ANALYZER(query)
    # print(tokenized_query)
    # tokenized_query = query.split()
    # print(tokenized_query)

    
    # Use jieba for word segmentation processing
    # tokenized_corpus = [list(jieba.cut(doc['snippet'])) for doc in raw_search_list]
    tokenized_corpus = [ ANALYZER(doc['snippet']) for doc in raw_search_list]
    # print(tokenized_corpus)
    bm25 = BM25Okapi(tokenized_corpus)
    

    # Get BM25 Score
    scores = bm25.get_scores(tokenized_query)
    

    raw_rerank_result = list(zip(scores, raw_search_list)) 
    # logger.info('raw_search_list_with_score: '+ json.dumps(raw_rerank_result, ensure_ascii=False,indent=4))
    scores_dicts = [(score, item) for score, item in zip(scores, raw_search_list) if score >= threshold]
    # Filter and sort results

    if not scores_dicts:
        return {'sorted_scores': [], 'sorted_search_list': []}
    if sort_enable:  
        scores_dicts = sorted(scores_dicts, key=lambda x: x[0], reverse=True)[:top_k]
        top_scores, top_search_list = zip(*scores_dicts) if scores_dicts else ([], [])            
        return {'sorted_scores': top_scores, 'sorted_search_list': top_search_list}
    else:
        top_scores, top_search_list = zip(*scores_dicts[:top_k]) if scores_dicts else ([], [])            
        return {'sorted_scores': top_scores, 'sorted_search_list': top_search_list}








if __name__ == "__main__":
        


    # Sample data and query parameters
    query = "刘姓 历史 称帝王 数量"
# query = "The Liu Song regime saw the emergence of Liu Yilong, Liu Jun and other emperors who worked hard to govern"

    search_list = [
        {
            "title": "我国古代的400多位帝王中，为何刘姓皇帝的数量最多 - 百度百科",
            "snippet": "在中国古代开创和延续众多朝代的400多位帝王中，刘姓皇帝的数量是最多的。那么，刘姓皇族的数量为何最多呢？上图_ 汉高��� 刘邦（公元前256年—前195年） 第一，汉帝国的超长待机是刘姓皇帝数量最多的直接原因。",
            "link": "https://baike.baidu.com/tashuo/browse/content?id=2d01677692dfa0b9e9267e9d",
            "datePublished": "2021-08-24",
            "dateLastCrawled": "2024-08-02"
        },
        {
            "title": "历史上有多少位姓刘的皇帝？_百度知道",
            "snippet": "历史上有多少位姓刘的皇帝？自西汉高祖刘邦公元前206年登王位（公元前201年登皇帝位）至南宋、金、齐帝刘豫公元1137年退位止，期间1343年中，刘氏共有59位皇帝，在帝王位共计676年，其中西汉一一东汉一统中华天下426",
            "link": "https://zhidao.baidu.com/question/198466182290908965.html",
            "datePublished": "2015-07-25",
            "dateLastCrawled": "2024-08-23"
        },
        {
            "title": "中国历史上皇帝最多姓氏——刘姓，其对中华文化产生了 ...",
            "snippet": "在中国历史上，刘姓先后产生正统皇帝多达92人！另外，历史上的刘氏诸侯国数量也是排名第一的，刘姓称诸侯王者多���1000多人。自刘累、刘康公开始，刘氏政权4000年连绵不断，先后建立朝代包括韦国、西汉、东汉��蜀汉、前赵、南朝宋、后汉、南汉、北汉",
            "link": "https://www.sohu.com/a/234046455_100185916",
            "datePublished": "2018-06-04",
            "dateLastCrawled": "2023-05-10"
        }
    ]

    # query1 = " ".join(['Tengwangge', 'Preface', 'Full text'])
    # query2 = " ".join(['Tengwang Pavilion Preface', 'Full text'])
    # query3 = " ".join(['Tengwang Pavilion Preface', 'Tengwang Pavilion', 'Preface', 'Full text'])
    query = "巴黎奥运会 金牌数量排名"
    top_k = 10
    threshold = 0  # Adjust the threshold according to actual conditions

    # call function
    start_time = datetime.datetime.now()
    results = rerank_by_bm25(query, search_list4, top_k, threshold)
    finish_time1 = datetime.datetime.now()
    time_difference = finish_time1 - start_time
    print("排序用时:", time_difference)
    print(json.dumps(results,ensure_ascii=False,indent=4))

    
#         # Call function
#     start_time = datetime.datetime.now()
#     results = rerank_by_bm25(query2, search_list3, top_k, threshold)
#     finish_time1 = datetime.datetime.now()
#     time_difference = finish_time1 - start_time
#     print("Sorting time:", time_difference)
#     print(json.dumps(results,ensure_ascii=False,indent=4))

#     # Call function
#     start_time = datetime.datetime.now()
#     results = rerank_by_bm25(query3, search_list3, top_k, threshold)
#     finish_time1 = datetime.datetime.now()
#     time_difference = finish_time1 - start_time
#     print("Sorting time:", time_difference)
#     print(json.dumps(results,ensure_ascii=False,indent=4))
#     # Output results
    # print("Sorted Scores:", results['sorted_scores'])
    # print("Sorted Documents:")
    # for doc in results['sorted_search_list']:
    #     print(doc)
