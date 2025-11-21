import settings
from enum import Enum

class IndexType(Enum):
    """索引类型枚举"""
    MAIN = "main"
    SNIPPET = "snippet"
    CONTENT_CONTROL = "content_control"
    FILE_CONTROL = "file_control"

#GetMainIndex名
def get_main_index_name(user_id:str) -> str:
    return settings.INDEX_NAME_PREFIX + user_id

def get_snippet_index_name(user_id:str) -> str:
    return settings.SNIPPET_INDEX_NAME_PREFIX + user_id

def get_content_control_index_name(user_id:str) -> str:
    return 'content_control_' + get_main_index_name(user_id)

def get_file_index_name(user_id:str) -> str:
    return 'file_control_' + get_main_index_name(user_id)