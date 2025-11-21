
import warnings
from settings import INDEX_NAME_PREFIX, SNIPPET_INDEX_NAME_PREFIX, KBNAME_MAPPING_INDEX
from utils.config_util import es

warnings.filterwarnings("ignore")

if __name__ == '__main__':
    # View cluster health status
    # View all indexes and display the detailed structure of each index. Note the differences in the latest version
    indexs = es.indices.get_alias(index="*")
    # View the names of all indexes in es
    index_names = indexs.keys()
    # Problematic index list
    problem_indexs = []
    for name in index_names:
        if INDEX_NAME_PREFIX in name or SNIPPET_INDEX_NAME_PREFIX in name:
            index_settings = es.indices.stats(index=name)
            health = index_settings["indices"][name]["health"]
            status = index_settings["indices"][name]["status"]
            print(f"索引名称: {name}, 状态: {status}, 健康状态: {health}")
            if status != "open" or health != "green":
                problem_indexs.append((name, status, health))
    # Print problematic indexes
    print("索引转态异常的索引有：")
    print(problem_indexs)


