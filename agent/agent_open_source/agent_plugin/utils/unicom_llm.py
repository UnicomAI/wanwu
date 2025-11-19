import os
import logging
import requests
import json
import requests
from typing import List
import requests
import urllib3
import aiohttp
import asyncio
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

from datetime import datetime, timedelta
from utils.auth import AccessTokenManager

import configparser
config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

MODEL_NAME_CONFIG = config["MODELS"]["default_llm"]
MODEL_NAME = os.getenv('CUAI_DEFAULT_LLM_MODEL_ID', MODEL_NAME_CONFIG)

MODEL_URL_CONFIG = config["MODELS"]["model_url"]
MODEL_URL = os.getenv('CUAI_DEFAULT_LLM_MODEL_URL', MODEL_URL_CONFIG)

# 实例化对象 [EN] instantiate object
token_manager = AccessTokenManager()
logger = logging.getLogger(__name__)
# 元景语言大模型服务封装,支持流式和非流式 [EN] Yuanjing language large model service encapsulation, supporting streaming and non-streaming
# 可用模型：unicom-7b-chat, unicom-13b-chat, unicom-34b-chat,unicom-7b-math(数学计算),unicom-13b-special(多轮query改写、系统提示词、会议纪要/摘要、纪委约谈),unicom-72b-chat-ali,unicom-72b-chat-ali-v2（默认） [EN] Available models: unicom-7b-chat, unicom-13b-chat, unicom-34b-chat, unicom-7b-math (mathematical calculation), unicom-13b-special (multiple rounds of query rewriting, system prompt words, meeting minutes/summaries, disciplinary committee interviews), unicom-72b-chat-ali, unicom-72b-chat-ali-v2 (default)

def req_unicom_llm_chat(messages:List, stream=True, model_name='unicom-70b-chat',model_url =MODEL_URL, do_sample=True,temperature=0.6):
    if model_url == MODEL_URL:
        base_url, _, model_name = model_url.rpartition('/')

    elif model_url and model_name.lower() not in model_url:
        base_url = model_url
    elif model_name.lower() in model_url:
        base_url, _, model_name = model_url.rpartition('/')
    else:
        base_url = config["MODELS"]["unicom_base_url"]
        
        # print(access_token)
        # if model_name=="deepseek-r1":
        #     base_url = config["MODELS"]["unicom_base_url_hh"]
        #     access_token = token_manager.get_access_token("hh")  # 获取 token [EN] access_token = token_manager.get_access_token("hh") # Get token
    url = base_url +"/"+ model_name
    access_token = token_manager.get_access_token()  # 获取 token [EN] Get token
    payload ={
            "stream": stream,
            "model":model_name,
            "temperature": temperature,
            "do_sample": do_sample,
            "messages": messages
    } 
    # print(access_token)
    headers = {"Content-Type": "application/json","Authorization": f"Bearer {access_token}"}
    logger.info(f"model_llm req_unicom_llm_chat  base_url:{base_url}  model_name:{model_name}")
    try:
        response = requests.post(
            url, 
            json=payload, 
            headers = headers,
            # verify=False, 
            stream=True 
        )
        return response

    except requests.RequestException as e:
        print(str(e))
        return "No answer found due to LLM API error"

def req_unicom_llm(payload):
    model_name = payload.get("model","unicom-70b-chat")
    stream = payload.get("stream",False)   
    model_url = payload.get("model_url","")   
    if model_url == MODEL_URL:
        base_url, _, model_name = model_url.rpartition('/')
    elif model_url and model_name.lower() not in model_url:
        base_url = model_url
    elif model_name.lower() in model_url:
        base_url, _, model_name = model_url.rpartition('/')
    else:
        base_url = config["MODELS"]["unicom_base_url"]
    url = base_url +"/"+ model_name
    access_token = token_manager.get_access_token()  # 获取 token [EN] Get token
    
    headers = {"Content-Type": "application/json","Authorization": f"Bearer {access_token}"}
    logger.info(f"model_llm req_unicom_llm  base_url:{base_url}  model_name:{model_name}")
    payload.setdefault("temperature", 0.6)    
    payload.setdefault("do_sample", True)

    try:
        response = requests.post(url, json=payload, headers = headers, verify=False, stream=stream)
        return response

    except requests.RequestException as e:
        print(str(e))
        return "No answer found due to LLM API error"


# 定义异步请求函数-流式 [EN] Define asynchronous request function-streaming
async def req_unicom_llm_stream_async(payload):
    model_name = payload.get("model","unicom-70b-chat")
    model_url = payload.get("model_url","")   
    if model_url == MODEL_URL:
        base_url, _, model_name = model_url.rpartition('/')
    elif model_url and model_name.lower() not in model_url:
        base_url = model_url
    elif model_name.lower() in model_url:
        base_url, _, model_name = model_url.rpartition('/')
    else:
        base_url = config["MODELS"]["unicom_base_url"]
        
        # if model_name=="deepseek-r1":
        #     base_url = config["MODELS"]["unicom_base_url_hh"]
        #     access_token = token_manager.get_access_token("hh")  # 获取 token [EN] access_token = token_manager.get_access_token("hh") # Get token
            
    url = base_url +"/"+ model_name   
    access_token = token_manager.get_access_token()  # 获取 token [EN] Get token
    headers = {"Content-Type": "application/json", "Authorization": f"Bearer {access_token}"}
    
    logger.info(f"model_llm req_unicom_llm_stream_async  base_url:{base_url}  model_name:{model_name}")
    payload.setdefault("temperature", 0.6)    
    payload.setdefault("do_sample", True)
    payload["stream"]= True

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=payload, headers=headers, ssl=False, timeout=aiohttp.ClientTimeout(total=300)) as response:
                async for line in response.content:
                    line = line.decode('utf-8')  # 将字节流解码为字符串 [EN] Decode byte stream into string
                    if line.startswith("data:"):
                        line = line[5:]  # 移除 "data:" 前缀 [EN] Remove "data:" prefix
                        line_dict = json.loads(line)
                        yield line_dict  # 生成每一行数据 [EN] Generate each row of data
    
    except aiohttp.ClientError as e:         
        yield json.dumps({"code": 1, "msg": f"FAILED:{str(e)}"})

# 定义异步请求函数-非流式 [EN] Define asynchronous request function - non-streaming
async def req_unicom_llm_nonstream_async(payload):
    model_name = payload.get("model","unicom-70b-chat")
    model_url = payload.get("model_url","")   
    if model_url == MODEL_URL:
        base_url, _, model_name = model_url.rpartition('/')
    elif model_url and model_name.lower() not in model_url:
        base_url = model_url
    elif model_name.lower() in model_url:
        base_url, _, model_name = model_url.rpartition('/')
    else:
        base_url = config["MODELS"]["unicom_base_url"]
        # if model_name=="deepseek-r1":
        #     base_url = config["MODELS"]["unicom_base_url_hh"]
        #     access_token = token_manager.get_access_token("hh")  # 获取 token [EN] access_token = token_manager.get_access_token("hh") # Get token
            
    url = base_url +"/"+ model_name   
    access_token = token_manager.get_access_token()  # 获取 token [EN] Get token

    headers = {"Content-Type": "application/json", "Authorization": f"Bearer {access_token}"}
    logger.info(f"model_llm req_unicom_llm_nonstream_async  base_url:{base_url}  model_name:{model_name}")

    payload.setdefault("temperature", 0.6)    
    payload.setdefault("do_sample", True)
    payload["stream"]= False

    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=payload, headers=headers, ssl=False, timeout=aiohttp.ClientTimeout(total=300)) as response:
                
                return  await response.json()

    except aiohttp.ClientError as e:
        return {"code": 1, "msg": f"No answer found due to LLM API error:{str(e)}"}

async def handle_response():
    query = '你好'
    
    payload = {
        "model": "unicom-70b-chat",
        "stream": True,  # 如果服务器不支持 stream，可以尝试去掉这行 [EN] If the server does not support streams, you can try removing this line
        "temperature": 0.7,
        "do_sample": True,
        "messages": [{"role": "user", "content": query}]
    }
    
    # 打印返回的异步生成器类型 [EN] Print the returned asynchronous generator type
    response = req_unicom_llm_stream_async(payload)
       

    # 使用异步迭代器逐行处理流式响应 [EN] Process streaming responses line by line using an asynchronous iterator
    async for line in response:
        print(f"Received line: {line}")  # 打印每行流数据 [EN] Print each line of stream data

    # response = await req_unicom_llm_nonstream_async(payload)
    # print(response)
    

    
# async def main():
#     print("Starting async task...")
#     await asyncio.sleep(5)  # 模拟耗时操作 [EN] await asyncio.sleep(5) # Simulate time-consuming operations
#     print("Async task finished.")
#     print("the end")

if __name__ == "__main__":
    # asyncio.run(handle_response())
    # print(token_manager.get_access_token())
    # model_name = 'unicom-34b-chat'
    # model_name = 'unicom-16b-math'

    # query = '''假设您计划在 7 天内游览欧洲的三个国家：法国、意大利和德国。请详细阐述您如何制定旅游行程策略，包括交通选择、景点安排、住宿预订和时间分配。同时说明您在制定策略时考虑的关键因素，例如预算、个人兴趣和当地气候。''' [EN] query = '''Suppose you plan to visit three countries in Europe in 7 days: France, Italy, and Germany. Please elaborate on how you plan your travel itinerary, including transportation options, attraction arrangements, accommodation reservations, and time allocation. Also describe the key factors you considered when developing your strategy, such as budget, personal interests and local climate. '''
    # messages = [
    # {
    #     "role": "user",
    #     "content": query
    # }
    # ]
    
    # 非流式示例 [EN] Non-streaming example
    # response = req_unicom_llm_chat(messages,stream=False,model_name=model_name)   
    # print(response)
    # print(response.text)
    # print(response.json()["data"]["choices"][0]["message"]["content"])
    
    
    
    
    
    # 流式示例 [EN] Streaming example
    # model_name = 'unicom-70b-chat'
    # response = req_unicom_llm_chat(messages,stream=True,model_name=model_name)   
    
    
    # print(model_name)
    # for line in response.iter_lines(decode_unicode=True):
        
    #     if line.startswith("data:") :
    #         # print(line)
    #         line = line[5:]
    #         # print(line)
    #         line_dict = json.loads(line)
    #         incremental_content = line_dict["data"]["choices"][0]["message"]["content"]
            # print(incremental_content,end="")

    # 示例调用异步函数 [EN] Example of calling an asynchronous function
    payload = {
        "model": "unicom-70b-chat",
        "stream": True,
        "temperature": 0.5,
        "do_sample": True,
        "messages": [{"role": "user", "content": "你是谁"}]
    }
    # response = asyncio.run(req_unicom_llm_stream_async(payload))  # 在需要的地方调用 [EN] response = asyncio.run(req_unicom_llm_stream_async(payload)) # Call where needed
    # print(response)
    
    messages = [{"role": "user", "content": "你是谁"}]
    response = req_unicom_llm_chat(messages,stream=True,model_name="unicom-70b-chat")
    # response  = req_unicom_llm(payload)

    print(response.text)
   
    for line in response.iter_lines(decode_unicode=True):
        
        if line.startswith("data:") :
            # print(line)
            line = line[5:]
            # print(line)
            line_dict = json.loads(line)
            incremental_content = line_dict["data"]["choices"][0]["message"]["content"]
            print(incremental_content,end="")
    

    
    