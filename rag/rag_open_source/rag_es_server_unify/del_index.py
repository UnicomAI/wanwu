import warnings
from utils.config_util import es
from settings import INDEX_NAME_PREFIX, SNIPPET_INDEX_NAME_PREFIX

warnings.filterwarnings("ignore")

def delete_index(index_name):
    """根据索引名删除整个索引，并返回操作的状态"""
    try:
        response = es.indices.delete(index=index_name)
        # If the index is successfully deleted, the response usually contains acknowledged = True
        delete_status = {
            "success": response.get('acknowledged', False),
            "error": None
        }
    except Exception as e:
        # Catch exceptions like index not existing or other Elasticsearch errors
        delete_status = {
            "success": False,
            "error": str(e)
        }

    return delete_status


# Get index statistics
def get_index_stats(es, index_name):
    stats = es.indices.stats(index=index_name)
    return stats


# You can get specific statistical information as needed, such as the number of documents, the number of deleted documents, etc.
# The following is an example of getting the number of documents and the number of deleted documents
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
    # ============= Delete index =================
    # index_name = "rag_new_unify_dev_userid_kbname_mapping"
    # print(delete_index(KBNAME_MAPPING_INDEX))
    # index_name = 'rag_new_unify_dev_hhh20240815'
    # print(delete_index(index_name))
    # View all indexes and display the detailed structure of each index. Note the differences in the latest version
    indexs = es.indices.get_alias(index="*")
    # View the names of all indexes in es
    index_names = indexs.keys()
    for name in index_names:
        if INDEX_NAME_PREFIX in name or SNIPPET_INDEX_NAME_PREFIX in name:
            # Print index statistics
            index_stats = get_index_stats(es, name)
            # Number of printed documents and number of deleted documents
            doc_counts = get_doc_count_and_deleted_docs_count(index_stats)
            print("文档数量：", doc_counts['docs_count'])
            print("已删除文档数量：", doc_counts['deleted_docs_count'])
            print(name)
            print("============= ==============")
