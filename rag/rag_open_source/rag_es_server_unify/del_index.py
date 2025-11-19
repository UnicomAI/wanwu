import warnings
from utils.config_util import es
from settings import INDEX_NAME_PREFIX, SNIPPET_INDEX_NAME_PREFIX

warnings.filterwarnings("ignore")

def delete_index(index_name):
    """根据索引名删除整个索引，并返回操作的状态"""
    try:
        response = es.indices.delete(index=index_name)
        # 如果索引成功删除，通常响应中会包含 acknowledged = True [EN] If the index is successfully deleted, the response usually contains acknowledged = True
        delete_status = {
            "success": response.get('acknowledged', False),
            "error": None
        }
    except Exception as e:
        # 捕获异常，如索引不存在或其他Elasticsearch错误 [EN] Catch exceptions like index not existing or other Elasticsearch errors
        delete_status = {
            "success": False,
            "error": str(e)
        }

    return delete_status


# 获取索引统计信息 [EN] Get index statistics
def get_index_stats(es, index_name):
    stats = es.indices.stats(index=index_name)
    return stats


# 你可以根据需要获取特定的统计信息，例如文档数量、已删除文档数等 [EN] You can get specific statistical information as needed, such as the number of documents, the number of deleted documents, etc.
# 以下是一个获取文档数量和已删除文档数的例子 [EN] The following is an example of getting the number of documents and the number of deleted documents
def get_doc_count_and_deleted_docs_count(index_stats):
    total = index_stats['_all']['total']
    return {
        'docs_count': total['docs']['count'],
        'deleted_docs_count': total['docs']['deleted']
    }


def get_distribution_index_name(es):
    """
    根据 索引里的数据量条数，返回判定可用的 index_name
    """
    index_prefix = "rag_dev_basic_index"



if __name__ == '__main__':
    # ============= 删除 索引 ================= [EN] ============= Delete index =================
    # index_name = "rag_new_unify_dev_userid_kbname_mapping"
    # print(delete_index(KBNAME_MAPPING_INDEX))
    # index_name = 'rag_new_unify_dev_hhh20240815'
    # print(delete_index(index_name))
    # 查看所有索引，且展示每个索引的详细结构。 最新版本注意区别 [EN] View all indexes and display the detailed structure of each index. Note the differences in the latest version
    indexs = es.indices.get_alias(index="*")
    # 查看es中的所有索引的名称 [EN] View the names of all indexes in es
    index_names = indexs.keys()
    for name in index_names:
        if INDEX_NAME_PREFIX in name or SNIPPET_INDEX_NAME_PREFIX in name:
            # 打印索引统计信息 [EN] Print index statistics
            index_stats = get_index_stats(es, name)
            # 打印文档数量和已删除文档数 [EN] Number of printed documents and number of deleted documents
            doc_counts = get_doc_count_and_deleted_docs_count(index_stats)
            print("文档数量：", doc_counts['docs_count'])
            print("已删除文档数量：", doc_counts['deleted_docs_count'])
            print(name)
            print("============= ==============")
