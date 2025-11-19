import logging
import datetime
import pytz
import configparser
from pathlib import Path
from re import sub

from utils.tokenizers import CustomTokenizer
from utils.llm_tools import format_prompt_template, prompts_base_path

# 日志对象初始化 [EN] Log object initialization
logger = logging.getLogger(__name__)

# 加载配置文件 [EN] Load configuration file
config = configparser.ConfigParser()
config.read("config.ini", encoding="utf-8")

# 常量定义 [EN] constant definition
MAX_INPUT_TOKENS = int(config["AGENTS"]["MAX_INPUT_TOKENS"])  # 从配置读取的最大模型输入 token 数 [EN] Maximum number of model input tokens read from configuration
DEFAULT_MODEL = "deepseek-v3"                                   # 默认使用的模型名称 [EN] The model name used by default
DEFAULT_TIMEZONE = "Asia/Shanghai"                             # 默认的时区设置 [EN] Default time zone setting

# 拼接文本时的截断阈值（单位：token） [EN] Truncation threshold when splicing text (unit: token)
# 搜索内容拼接时如果剩余token数量低于该值，则舍弃该chunk，否则将chunk截断后再拼接 [EN] When the search content is spliced, if the number of remaining tokens is lower than this value, the chunk will be discarded, otherwise the chunk will be truncated and spliced ​​again.
TRUNCATION_THRESHOLD = 20                                   


def assemble_search_context(
    search_list: list,
    citation_label: str,
    tokenizer: CustomTokenizer,
    max_token_length: int
) -> str:
    """
    根据搜索结果构建受限长度的上下文内容字符串，拼接内容不超过指定 token 上限。

    参数：
        search_list (list): 搜索结果组成的列表，每项为字典，包含以下字段：
            - file_name (str): 文件名或标题。
            - sub_query (str): 子查询内容。
            - snippet (str): 匹配片段。
        citation_label (str): 每段拼接内容前缀标签，例如 "citation" 或 "参考信息"。
        tokenizer (CustomTokenizer): 用于统计和截断 token 的 tokenizer 实例。
        max_token_length (int): 限定的搜索内容的最大 token 数，超过后停止拼接。

    返回：
        str: 构建完成的上下文字符串，用于填充提示词模板。
    """
    search_context = ""
    current_tokens = 0

    for idx, item in enumerate(search_list):
        file_name = f"参考文件：{item.get('file_name', '')}\n" if item.get("file_name") else ""
        sub_query = f"{item.get('sub_query', '')}\n" if item.get("sub_query") else ""
        snippet = item.get("snippet", "")

        segment = f"{citation_label}-{idx + 1:02d}:\n{sub_query}{file_name}{snippet}" if citation_label else \
                  f"{idx + 1:02d}:\n{sub_query}{file_name}{snippet}"

        segment_tokens = tokenizer.count_tokens(segment)
        new_total = current_tokens + segment_tokens

        if new_total > max_token_length:
            leftover = max_token_length - current_tokens
            if leftover > TRUNCATION_THRESHOLD:
                truncated = tokenizer.truncate_text(segment, leftover)
                search_context += "\n\n" + truncated
                logger.info(f"超出token限制，终止拼接，使用chunk数量: {idx+1}（最后一项已截断）")
            else:
                logger.info(f"超出token限制，终止拼接，使用chunk数量: {idx}")
            break

        search_context += "\n\n" + segment
        current_tokens = new_total

    return search_context


def build_prompt_from_search_list(
    query: str,
    search_list: list,
    template_prefix: str,
    auto_citation: bool = False,
    model: str = DEFAULT_MODEL,
    max_length: int = MAX_INPUT_TOKENS,
    **kwargs
) -> str:
    """
    构建完整提示词（Prompt），将模板与搜索上下文拼接后填充参数。

    参数：
        query (str): 用户原始提问内容。
        search_list (list): 搜索结果组成的列表，见 build_limited_search_contents 中字段要求。
        template_prefix (str): 所用模板的前缀（例如 "bing"、"docqa"、"rag"）。
        auto_citation (bool): 是否启用 citation 样式引用（影响模板选择及编号格式）。
        model (str): 当前使用的语言模型名（用于 tokenizer 选择）。
        max_length (int): 总 token 限制上限，包括模板 + 搜索内容。
        **kwargs: 其他模板所需参数，例如 cur_date 等。

    返回：
        str: 构造好的提示词，可直接送入 LLM 使用。
    """
    tokenizer = CustomTokenizer(model_name=model)
    citation_label = "citation" if auto_citation else "参考信息"

    template_name = f"{template_prefix}_prompt{'_citation' if auto_citation else ''}.txt"
    template_path = Path(prompts_base_path, template_name)  # 确保 prompts_base_path 为 Path 类型 [EN] Make sure prompts_base_path is of type Path
    template_content = template_path.read_text(encoding='utf-8')
    template_token_num = tokenizer.count_tokens(template_content)

    search_content_max_length = max_length - template_token_num
    search_contents = assemble_search_context(
        search_list, citation_label, tokenizer, search_content_max_length
    )

    prompt = format_prompt_template(
        template_name,
        question=query,
        context=search_contents,
        **kwargs
    )

    logger.info(f"{template_prefix}_prompt token数: {tokenizer.count_tokens(prompt)}")
    return prompt


def build_bing_prompt_from_search_list(
    query: str,
    search_list: list,
    auto_citation: bool = False,
    model: str = DEFAULT_MODEL,
    max_length: int = MAX_INPUT_TOKENS
) -> str:
    """
    构建适用于 Bing 场景的提示词，附带当前日期（用于提示词中的上下文提示）。

    参数：
        query (str): 用户问题。
        search_list (list): 搜索结果。
        auto_citation (bool): 是否使用 citation 风格格式。
        model (str): 使用的模型名称。
        max_length (int): token 限制。

    返回：
        str: 拼接后的完整提示词。
    """
    now = datetime.datetime.now(pytz.utc).astimezone(pytz.timezone(DEFAULT_TIMEZONE))
    cur_date = now.strftime("%Y年%m月%d日")

    return build_prompt_from_search_list(
        query=query,
        search_list=search_list,
        template_prefix="bing",
        auto_citation=auto_citation,
        model=model,
        max_length=max_length,
        cur_date=cur_date
    )


def build_docqa_prompt_from_search_list(
    query: str,
    search_list: list,
    auto_citation: bool = False,
    model: str = DEFAULT_MODEL,
    max_length: int = MAX_INPUT_TOKENS
) -> str:
    """
    构建适用于 DocQA（基于文档问答）场景的提示词。
    与 Bing 相比不包含日期。
    """
    return build_prompt_from_search_list(
        query=query,
        search_list=search_list,
        template_prefix="docqa",
        auto_citation=auto_citation,
        model=model,
        max_length=max_length
    )


def build_rag_prompt_from_search_list(
    query: str,
    search_list: list,
    auto_citation: bool = False,
    model: str = DEFAULT_MODEL,
    max_length: int = MAX_INPUT_TOKENS - 500
) -> str:
    """
    构建适用于 RAG（检索增强生成）场景的提示词，
    并将搜索项中的 title 字段映射为 file_name 供模板使用。

    参数与返回：同上。
    """
    updated_list = [dict(item, file_name=item.get('title', '')) for item in search_list]
    return build_prompt_from_search_list(
        query=query,
        search_list=updated_list,
        template_prefix="rag",
        auto_citation=auto_citation,
        model=model,
        max_length=max_length
    )
