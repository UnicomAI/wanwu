import logging
from pathlib import Path
from typing import Dict, Any
from langchain_openai import ChatOpenAI  # type: ignore
from langchain_core.prompts import PromptTemplate # type: ignore
from utils.output_parser import extract_json,extract_json_plus
from langchain_core.runnables import RunnableSequence
from langchain_core.messages import AIMessage
from utils.auth import  AccessTokenManager
from utils.timing import advanced_timing_decorator

import configparser
from dotenv import load_dotenv

load_dotenv()

logger = logging.getLogger(__name__)


config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

token_manager = AccessTokenManager()
prompts_base_path = "/agent/agent_open_source/utils/prompts"
OPENAI_BASE_URL = config["MODELS"]["openai_base_url"]


def load_prompt_template(template_file: Path, template_file_base:str=prompts_base_path,encoding: str = "utf-8") -> PromptTemplate:
    """Load and validate a prompt template from file.
    
    Args:
        template_path: Path to the template file
        encoding: File encoding (default: utf-8)
        
    Returns:
        Loaded PromptTemplate instance
        
    Raises:
        FileNotFoundError: If template file doesn't exist
        ValueError: If template format is invalid
    """
    try:
        prompt_template_path = Path(prompts_base_path, template_file) 
        return PromptTemplate.from_file(prompt_template_path, encoding=encoding)
    except FileNotFoundError as e:
        logger.error(f"Template file not found: {prompt_template_path}")
        raise
    except ValueError as e:
        logger.error(f"Invalid template format in {prompt_template_path}: {e}")
        raise
    except Exception as e:
        logger.error(f"Unexpected error loading template: {e}")
        raise

def format_prompt_template(template_name, **kwargs):
    """
    格式化提示词模板
    
    参数:
    template_name (str): 模板文件名称
    kwargs: 模板所需的参数字典
    
    返回:
    str: 格式化后的提示词
    """
    prompt_template = load_prompt_template(template_name)
    return prompt_template.format(**kwargs)

def generate(llm_chain: RunnableSequence, stream:bool=False, **kwargs: Any) -> Dict[str, Any]:
    """Generate a research plan using the provided LLM chain."""
    try:
        if stream:
            return llm_chain.stream(input=kwargs)  # 使用 input 参数 [EN] Use input parameters
        else:
            # print(kwargs)
            response = llm_chain.invoke(input=kwargs)  # 使用 input 参数 [EN] Use input parameters
            return response
    except Exception as e:
        logger.error(f"lc_generate failed: {e}")
        raise RuntimeError("lc_generate failed") from e
    

@advanced_timing_decorator(task_name="req_llm_structed_output")
def req_llm_structed_output(promt_file, model="yuanjing-70b-chat", **kwargs: Any):
    
    prompt_template = load_prompt_template(promt_file)
    
    
    token = token_manager.get_access_token()
    # print(token)
    # prompt = prompt_template.format(**kwargs)
    # print(prompt)
    llm = ChatOpenAI(model=model, api_key=token,base_url=OPENAI_BASE_URL)
    llm_chain =  prompt_template | llm | extract_json_plus
    
    # print("-" * 50)
    # 获取生成器的结果并转换为列表 [EN] Get the result of the generator and convert it to a list
    result = list(generate(llm_chain, stream=True, **kwargs))
    
    # 现在可以安全地序列化为JSON [EN] Can now be safely serialized to JSON
    # print(json.dumps(result[0], ensure_ascii=False, indent=2))
    # print("-" * 50)
    return result[0]

if __name__ == "__main__":
    query = "编写一个GUI Agent最新进展报告"
    query = "杭州明天天气"
    query = "将此word文件转换成html,并输出转换成功后的html给我，其中上传文档的url为：https://obs-nmhhht6.cucloud.cn/assistant-obj-prod/6c9832df-7eb6-4bd9-8a45-04be42d053a9.docx"
    search_plan =req_llm_structed_output(
        "bing_query_rewrite.txt",
        # model="yuanjing-70b-chat",
        model="deepseek_r1",
        query=query,
        cur_date = "2025年3月22日",
        cur_year = "2025")
    print(search_plan)
    # print(search_plan["sub_queries"][:2]+search_plan["sub_queries_en"][:2])

