import asyncio
import json
import math
import os
import json
import configparser
import urllib3
import logging
import copy
from pyexpat import model
from dotenv import load_dotenv
from langchain_mcp_adapters.client import MultiServerMCPClient
from langgraph.prebuilt import create_react_agent
from langchain_openai import ChatOpenAI
#from llm_agent_server_app.utils.llm_tools import load_prompt_template
#from llm_agent_server_app.utils.model_tools import req_unicom_llm_chat
from langchain_core.messages import (
    BaseMessage,
    AIMessage,
    HumanMessage,
    SystemMessage,
    ToolMessage,
    trim_messages,
    messages_from_dict,
    messages_to_dict,
    convert_to_messages,
    convert_to_openai_messages

)
# from llm_agent_server_app.utils.generator import generate_stream_response,generate_non_stream_response,create_error_response
#from llm_agent_server_app.utils.auth import AccessTokenManager


urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

###加载配置文件 [EN] ##Load configuration file
config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')
OPENAI_BASE_URL = config["MODELS"]["openai_base_url"]
DEFAULT_FUNCTION_MODEL = 'deepseek-v3'

#token_manager = AccessTokenManager()
logger = logging.getLogger(__name__)

def tool_to_dict(tool):
    return {
        "name": tool.name,
        "description": tool.description,
        "args_schema": tool.args_schema,
        "response_format": tool.response_format
        # 如果你不需要 coroutine，可以不加，因为它是函数对象，不能序列化 [EN] If you don’t need coroutine, you don’t need to add it because it is a function object and cannot be serialized.
    }


###判断是否命中mcp server [EN] ##Determine whether the mcp server is hit
def generator_empty_check(gen):
    try:
        next(gen)
        # logger.info(f"generator_empty_check a:{a}")
        return gen
    except StopIteration:
        return None

async def mcp_client(query,mcp_tools,tools_name,temperature= 0.01,model_name = DEFAULT_FUNCTION_MODEL,model_url= "",stream = True,history = []):
    '''
    query：用户请求问题
    mcp_list: key为server_name，value为对应的mcp server详细信息，可以是多个key与value组合的json数组
    model_name：mcp调用的大模型名称
    model_url：mcp调用的大模型base_url
    temperature：语言模型参数，控制输出的随机性，取值范围 [0.01,1.00]，小数位最多保留两位 。值越大（例如0.8），会使输出更随机，更具创造性；值越小（例如0.2），输出会更加稳定或确定。系统默认值为0.5,
    stream：输出方式，流式为true，非流式为False，默认为流式输出
    history：用户对话历史，可基于对话历史实现多轮对话效果
    '''
    try:
        if model_name == "qwen2.5-7b-instruct":
            llm = ChatOpenAI(
                model="qwen2.5-7b-instruct",
                api_key= "XXXXXXXXXX",
                base_url="https://dashscope.aliyuncs.com/compatible-mode/v1",
                temperature=0,
                streaming=stream  # 启用流式输出 [EN] Enable streaming output
            )
        else:
            llm = ChatOpenAI(
                model=model_name,
                api_key='Ojsls',
                base_url= model_url or OPENAI_BASE_URL,
                temperature = temperature
            )

        async with MultiServerMCPClient(mcp_tools) as client:
            # 转换成可序列化的列表以便查看传入的tool列表 [EN] Convert to a serializable list to view the incoming tool list
            logger.info(f"mcp_server_is:{mcp_tools}")
            tools = client.get_tools()
            logger.info(f"tools_name is:{tools_name}")
           # wanted_tool_names = ["maps_direction_bicycling","mcp_howtocook_getAllRecipes"]
            tools = [tool for tool in tools if tool.name in tools_name]
            logger.info(f"tools_is:{tools}")
            serializable_tools = [tool_to_dict(t) for t in tools]
            # logger.info(f"Tools:{json.dumps(serializable_tools, indent=2, ensure_ascii=False)}")
            logger.info(f"MCP Tools:共计{len(serializable_tools)}个")
            logger.info(f"MCP Tools有哪些:{serializable_tools}")
            ###创建mcp智能体 [EN] ##Create mcp agent
            agent = create_react_agent(llm, tools)
            messages_input = []
            
            ###基于历史对话实现多轮对话效果 [EN] ##Achieve multiple rounds of dialogue effects based on historical dialogues
            print(f"---history长度为：{len(history)}---")
            logger.info(f"---history长度为：{len(history)}---")
            if history:
                has_rewrite_query = any("rewrite_query" in item.keys() for item in history)
                if has_rewrite_query:
                    for item in history:
                        messages_input.append({"role": "user", "content": item["rewrite_query"]})
                        messages_input.append({"role": "assistant", "content": ""})
                        # messages_input.append({"role": "assistant", "content": item["response"]})
                else:
                    history = [item for item in history if item['role'] != "system"]
                    for i in range(0, len(history), 2):
                        messages_input.append({"role": "user", "content": history[i]['content']})
                        messages_input.append({"role": "assistant", "content": ""})
                        # messages_input.append({"role": "assistant", "content": history[i + 1]["content"]})
            messages_input.append({'role': 'user', 'content': query})
            print('messages是什么:',messages_input)
            if stream:
                result = agent.astream({"messages": messages_input})
                i = 0
                text = ""
                async for line in result:
                    logger.info(f"流式mcp line:{line}")
                    print('流式mcp line:',line)
                    # await asyncio.sleep(0.5)
                    if line and "agent" in line.keys():
                        line = line["agent"].get("messages")[0]
                        if line.additional_kwargs.get("tool_calls") or text:
                            if i == 0:  ###打印mcp起始标志，用以判断是否命中mcp server [EN] ##Print the mcp start flag to determine whether the mcp server is hit.
                                yield "命中mcp server"
                                yield line
                            ####最后一次流式拆分输出 [EN] ###Last streaming split output
                            else:
                                answer = line.content
                                # print(f"line:{line}")
                                for t in range(0, len(answer), 3):
                                    split_text = line.copy()
                                    split_text.content = answer[t:t + 3]
                                    split_text.response_metadata = {}
                                    split_text.usage_metadata = {}
                                    yield split_text
                                line.content = ""  ###流式输出结束标识符 [EN] ##Streaming output end identifier
                                yield line
                    elif line and "tools" in line.keys():
                        line = line["tools"].get("messages")[0]
                        text += line.content
                        yield line
                    i += 1
            else:
                result = await agent.ainvoke({"messages": messages_input})
                # logger.info(f"result:{result}")
                yield "命中mcp server"
                yield result

    except Exception as e:
        # print(f"mcp_client 发生错误：{str(e)}") [EN] print(f"mcp_client error occurred: {str(e)}")
        logger.info(f"mcp_client 发生错误：{str(e)}")
        pass

def sync_generator(query,mcp_tools,tools_name,temperature= 0.01,model_name = DEFAULT_FUNCTION_MODEL,model_url= "",stream = True,history = []):
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    async def consume_async_gen(mcp_client_response):
        async for item in mcp_client_response:
            yield item
    try:
        mcp_client_response = mcp_client(query,mcp_tools,tools_name,temperature,model_name ,model_url,stream ,history)
        gen = consume_async_gen(mcp_client_response)
        
        while True:
            try:
                item = loop.run_until_complete(gen.__anext__())
                yield item
            except StopAsyncIteration as e:
                break
            except Exception as e:
                logger.critical(f"sync_generator Exception: {str(e)}")
                print(f"sync_generator Exception: {str(e)}")
                break    
    except Exception as e:
        logger.info(f"sync_generator 发生错误：{str(e)}")
        print(f"sync_generator 发生错误：{str(e)}")
    finally:
        # 确保在函数结束时关闭事件循环 [EN] Make sure you close the event loop when the function ends
        loop.close()
        
###model：yuanjing-70b-functioncall、qwq-32b、deepseek-v3-functioncall
        
def mcp_server_client(query,mcp_tools,tools_name,temperature= 0.01,model_name = DEFAULT_FUNCTION_MODEL,model_url= "",stream = True,history = []):
    try:
        result = sync_generator(query,mcp_tools,tools_name,temperature,model_name ,model_url,stream,history)
        result = generator_empty_check(result)
        logger.info(f"mcp_server_client result:{result}")
        if result:
            if stream:
                return result   
            else:
                non_llm_response = ""
                for line in result:
                    line = line["messages"]
                    # logger.info(f"非流式line：{line}\n长度为{len(line)}") [EN] logger.info(f"Non-streaming line: {line}\nThe length is {len(line)}")
                    if len(line) == 2:  ####仅包含输入和答案，不包含tool_calls相关返回 [EN] ###Contains only input and answers, does not include tool_calls related returns
                        logger.info(f"-----mcp_server_client 非流式未命中mcp-----")
                        return None
                    for text in line:
                        logger.info(f"非流式text:{text}\n")
                        if isinstance(text, HumanMessage):
                            # non_llm_response += f"query为：" + text.content  + "\n" [EN] non_llm_response += f"query is: " + text.content + "\n"
                            non_llm_response += ""
                        elif isinstance(text, AIMessage) and text.response_metadata.get("finish_reason", "") == "tool_calls":
                            mcp_name = text.tool_calls[0]["name"]
                            mcp_args = text.tool_calls[0]["args"]
                            non_llm_response += f"<tool>mcp-工具名：{mcp_name}\n\n\n```mcp-请求参数：\n{mcp_args}\n```\n\n"
                            # non_llm_response += f"<tool>\n\n\n```mcp 工具名：\n" + json.dumps(text.tool_calls,ensure_ascii= False) + "\n```\n\n" [EN] non_llm_response += f"<tool>\n\n\n```mcp tool name:\n" + json.dumps(text.tool_calls,ensure_ascii= False) + "\n```\n\n"
                        elif isinstance(text, ToolMessage) and text.type == "tool":
                            non_llm_response += f"\n\n\n```mcp-调用结果：\n" + text.content + "\n```\n\n" + "</tool>\n\n"
                        else:
                            non_llm_response +=  text.content  + "\n"
                    # logger.info(f"non_llm_response:{non_llm_response}")
                        
                return non_llm_response
        else:
            return None
    except Exception as e:
        logger.info(f"mcp_server_client 发生错误：{str(e)}")
        # print(f"mcp_server_client 发生错误：{str(e)}") [EN] print(f"mcp_server_client error occurred: {str(e)}") 
        pass
        

if __name__ == "__main__":
    # query = "3*5等于多少" [EN] query = "What is 3*5 equal to"
    query = "北京今天天气如何"
    query = "北京天安门经纬度"

    mcp_list = {
            "a_map": {
                "url": "https://mcp.amap.com/sse?key=77b5f0d102c848d443b791fd469b732d",
                "transport": "sse",
            },
            "Bing CN MCP": {
                        "url": "https://mcp.api-inference.modelscope.net/428c5410d4c64b/sse",
                        "transport": "sse",
                    }
    }
    
    stream = True
    print(f"-----start----")
    result = mcp_server_client(query,mcp_list,temperature= 0.01,model_name = 'deepseek-v3-functioncall',model_url= "http://172.17.0.1:6668/callback/v1/model/3",stream=stream)
    print(f"test result:{result}")
    if result:
        if  stream:
            for line in result:
                print(f"\nfinal_result :{line}")
    
        else:
            print(result)
    else:
        print(f"未命中mcp")
