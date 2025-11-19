
import warnings
from settings import INDEX_NAME_PREFIX, SNIPPET_INDEX_NAME_PREFIX, KBNAME_MAPPING_INDEX
from utils.config_util import es

warnings.filterwarnings("ignore")

if __name__ == '__main__':
    # 查看集群健康状态 [EN] View cluster health status
    # 查看所有索引，且展示每个索引的详细结构。 最新版本注意区别 [EN] View all indexes and display the detailed structure of each index. Note the differences in the latest version
    indexs = es.indices.get_alias(index="*")
    # 查看es中的所有索引的名称 [EN] View the names of all indexes in es
    index_names = indexs.keys()
    # 有问题的索引list [EN] Problematic index list
    problem_indexs = []
    for name in index_names:
        if INDEX_NAME_PREFIX in name or SNIPPET_INDEX_NAME_PREFIX in name:
            index_settings = es.indices.stats(index=name)
            health = index_settings["indices"][name]["health"]
            status = index_settings["indices"][name]["status"]
            print(f"索引名称: {name}, 状态: {status}, 健康状态: {health}")
            if status != "open" or health != "green":
                problem_indexs.append((name, status, health))
    # 打印有问题的索引 [EN] Print problematic indexes
    print("索引转态异常的索引有：")
    print(problem_indexs)


