import json
import re
from typing import List
import logging
from langchain_core.messages import AIMessage

logger = logging.getLogger(__name__)

def extract_json(text):
    """
    从字符串中提取 JSON 数据，支持嵌套大括号匹配。
    """
    stack = []
    start = -1
    extracted_json = None

    # 遍历文本，查找大括号 [EN] Traverse the text, looking for braces
    for i, char in enumerate(text):
        if char == '{':
            if not stack:
                start = i  # 记录最外层 '{' 的起始位置 [EN] Record the starting position of the outermost '{'
            stack.append(char)
        elif char == '}':
            if stack:
                stack.pop()
                if not stack:  # 栈为空，表示匹配到完整 JSON [EN] The stack is empty, indicating that the complete JSON is matched.
                    extracted_json = text[start:i+1]
                    break

    # 如果找到了 JSON 结构，尝试解析 [EN] If a JSON structure is found, try to parse
    if extracted_json:
        try:
            json_content = extracted_json.replace("'", '"')  # 转换单引号为双引号 [EN] Convert single quotes to double quotes
            json_data = json.loads(json_content)  # 解析 JSON [EN] Parse JSON
            return json_data
        except json.JSONDecodeError:
            return {}

    return {}

def extract_json_plus(message: AIMessage) -> List[dict]:
    """
    从 AIMessage 中提取一个或多个 JSON 块，并以列表形式返回解析后的字典对象。
    """
    text = message.content

    # 匹配三重反引号（可能带有 json 标记），或直接的花括号 [EN] Matches triple backticks (possibly with json tags), or straight curly braces
    pattern = r"```(?:json)?\s*([\s\S]*?)```|({[\s\S]*})"
    matches = re.findall(pattern, text, re.DOTALL)

    if not matches:
        raise ValueError(f"在消息中未发现可匹配的 JSON 块: {text}")

    results = []
    for group1, group2 in matches:
        # 取到三重反引号包裹的文本或花括号文本 [EN] Get text wrapped in triple backticks or curly braced text
        json_str = group1.strip() if group1.strip() else group2.strip()
        if not json_str:
            continue
        
        try:
            # 1. 替换转义引号 [EN] 1. Replace escaped quotes
            cleaned_str = json_str.replace('\\"', '"')
            # 2. 替换单引号为双引号（注意要先处理嵌套的情况） [EN] 2. Replace single quotes with double quotes (note that nesting needs to be dealt with first)
            cleaned_str = cleaned_str.replace("'", '"')
            # 3. 解析 JSON [EN] 3. Parse JSON
            parsed_data = json.loads(cleaned_str)
            results.append(parsed_data)
        except json.JSONDecodeError as e:
            logger.warning(f"无法解析 JSON: {e}\n原始字符串: {json_str}")
            continue  # 继续处理下一个匹配项，而不是直接失败 [EN] Continue processing the next match instead of failing directly
    
    if not results:
        raise ValueError(f"未能解析出任何有效的 JSON: {text}")

    return results

if __name__ == "__main__":
    text = "Error code: 422 - {'code': 14, 'msg': 'Unavailable', 'lid': '387f9907-42dc-47fc-862c-9c68f67ad1d2', 'data': {}}"
    text = "分析结果是:{'type':2,'REASON':'联网搜索'}de "
    print(extract_json(text))