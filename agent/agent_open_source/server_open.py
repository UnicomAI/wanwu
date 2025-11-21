import requests
import json
import os
import anyio
from openai import OpenAI
import asyncio
from bing_plus import *

from langchain_community.tools.tavily_search import TavilySearchResults
from langgraph.graph import StateGraph,END,START
from typing import List, Dict, Optional,Annotated
from typing_extensions import TypedDict
from langgraph.graph.message import add_messages
from langchain_core.messages import ToolMessage, HumanMessage
from langgraph.checkpoint.memory import MemorySaver
import io
from datetime import datetime
from flask import  stream_with_context, Response, request,jsonify
from flask import Flask
from flask_cors import CORS
from logging_config import setup_logging
from bing_plus import *
import configparser
from langchain.requests import RequestsWrapper
from mcp_client import *
from urllib.parse import urlparse


import logging

URL_RAG_STREAM = os.getenv("URL_RAG_STREAM")
URL_MODEL = os.getenv("URL_MODEL")
URL_RAG = os.getenv("URL_RAG")


log_dir = "./logs"
os.makedirs(log_dir, exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] [%(name)s] %(message)s',
    handlers=[
        logging.FileHandler(f"{log_dir}/server.log", encoding='utf-8'),
        logging.StreamHandler()  # output to console
    ]
)

logger = logging.getLogger(__name__)
logger.info("主服务启动")


app = Flask(__name__)
CORS(app, supports_credentials=True)



def img2base64(img_path):
    img_format = img_path.split('.')[-1]
    with open(img_path, "rb") as f:
        encoded_image = base64.b64encode(f.read())
        encoded_image_text = encoded_image.decode("utf-8")
        img_base64_str = f"data:image/{img_format};base64,{encoded_image_text}"
        return img_base64_str


os.environ["ARK_API_KEY"] = "eyJh"




@app.route("/agent",methods=['POST'])
def agent_start():
    @stream_with_context
    def generate():
        try:

            data = request.get_json()
            logger.info('入参是request_params: '+ json.dumps(data, ensure_ascii=False))
            #Basic parameters
            question = data.get("input")
            stream = data.get("stream",True)
            history = data.get("history",[])
            userId = request.headers.get('X-Uid')
            #function_call = data.get("function_call",False)
            logger.info('user_id是:'+userId)


            mcp_tools = data.get("mcp_tools", {})
            tools_name = data.get("tools_name",[]) 


            #Large model parameters
            model = data.get("model")
            model_url = data.get("model_url")
            system_role = data.get("system_role",'')
            model_id = data.get("model_id")
            
            do_sample = data.get("do_sample",False)
            temperature = data.get("temperature",0.7)
            repetition_penalty = data.get("repetition_penalty",1)
            frequency_penalty = data.get("frequency_penalty",0)
            presence_penalty = data.get("presence_penalty",0)
            top_p = data.get("top_p",0.7)
            top_k = data.get("top_k",50)
            max_tokens = data.get("max_tokens",1024)
            enable_thinking = data.get("enable_thinking",False)

            #Search parameters
            auto_citation = data.get("auto_citation",False)
            use_search = data.get("use_search",False)
            need_search_list = data.get("need_search_list",True)
            search_url = data.get("search_url",'')
            search_key = data.get("search_key",'')
            search_rerank_id = data.get("search_rerank_id",'11')



            #code interpreter parameters
            use_code = data.get("use_code",False)
            file_name = data.get("file_name")
            upload_file_url = data.get("upload_file_url",[])


            #rag parameters
            chitchat = data.get("chitchat",False)
            kn_params = data.get("kn_params",{})
            use_know = data.get("use_know",False)

            #Other plug-in parameters
            plugin_list = data.get("plugin_list",[])
            
            
            #url = f"http://bff-service:6668/callback/v1/model/{model_id}"
            url = URL_MODEL+f":6668/callback/v1/model/{model_id}"
            response = requests.get(url)
            function_call = False
            is_vision = False
            if response.status_code == 200:
                data = response.json()
                result = data.get("data", {}).get("config").get("functionCalling")
                is_visionSupport = data.get("data", {}).get("config").get("visionSupport")
                logger.info(f"function_call support result{result}")
                logger.info(f"vision support result{is_visionSupport}")
                if result == "noSupport":
                    function_call = False
                    logger.info(f"不支持function_call {function_call}")
                else:
                    function_call = False
                    logger.info(f"支持function_call {function_call}")

                if is_visionSupport == "support":
                    is_vision = True






            messages = []
            for item in history:
                query = item.get("query")
                response = item.get("response")
                if query:
                    messages.append({"role": "user", "content": query})
                if response:
                    messages.append({"role": "assistant", "content": response})

            # Limit to keep only the last 5 rounds (i.e. 10 messages)
            history = history[-10:]
            messages = messages[-10:]

            # Add user input for this round
            
            
            
            
            chatdoc_schema =         {
    "api_schema": {
        "info": {
            "description": "用于解析并回答用户上传的docx、txt、xlsx各种类型文件内容的问题",
            "title": "chatdoc",
            "version": "1.0.0"
        },
        "openapi": "3.0.0",
        "paths": {
            "/doc_pra": {
                "post": {
                    "description": "用于解析并回答用户上传的docx、txt、xlsx各种类型文件内容的问题",
                    "summary":"chatdoc",
                    "operationId": "chatdoc",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "properties": {
                                        "query": {
                                            "description": "用户提出的问题",
                                            "type": "string"
                                        },
                                        "upload_file_url":{
                                            "description":"文件下载链接",
                                            "type":"string"
                                            }                                               
                                    },
                                    "required": [
                                        "query",
                                        "upload_file_url"
                                    ],
                                    "type": "object"
                                }
                            }
                        }
                    }
                }
            }
        },
        "servers": [
            {
                "url": "http://localhost:1991"
            }
        ]
    }
}
            if upload_file_url:
                plugin_list.append(chatdoc_schema)
                
                

            if kn_params:
                knowledgebase_name = kn_params.get('knowledgeBase')
                threshold = kn_params.get('threshold',0.4)
                topk = kn_params.get('topK',5)
                rerank_id = kn_params.get('rerank_id')
                rerank_mod = kn_params.get('rerank_mod','rerank_model')
                retrieve_method = kn_params.get('retrieve_method','hybrid_search') 
                weights = kn_params.get('weights',None)
                max_history = kn_params.get('max_history')
                rewrite_query = kn_params.get('rewrite_query')
                term_weight_coefficient = kn_params.get('term_weight_coefficient',1.0)
                metadata_filtering = kn_params.get('metadata_filtering',True)
                knowledgeIdList = kn_params.get('knowledgeIdList',[])
                metadata_filtering_conditions = kn_params.get('metadata_filtering_conditions',[])
                use_graph = kn_params.get('use_graph',False)


#If it is a multimodal model, enter multimodal mode
            if is_vision:
                logger.info(f"to_vision model answer")
                client = OpenAI(api_key='aaaaa',
                         base_url = model_url,
                )
                if upload_file_url:
                    urls = upload_file_url if isinstance(upload_file_url, list) else [upload_file_url]
                    image_contents = []
                    for url in urls:
                        logger.info(f"minio_file_is {upload_file_url}")
                        #Download image files to local
                        path = urlparse(url).path  # "/bucket/yourfile.jpg"
                        filename = os.path.basename(path)  # "yourfile.jpg"
                        local_path = "/agent/agent_open_source/file/"+filename
                        logger.info(f"minio_file_path {local_path}")

                        with requests.get(url, stream=True) as r:
                            r.raise_for_status()
                            with open(local_path, "wb") as f:
                                for chunk in r.iter_content(chunk_size=8192):
                                    if chunk:
                                        f.write(chunk)

                        print(f"jpg_file save: {local_path}")
                        image_contents.append({
                            "type": "image_url",
                            "image_url": {"url": img2base64(local_path)}
                        })

                    messages = [{
                        "role": "user",
                        "content": [
                            {"type": "text", "text": question},
                            *image_contents
                        ]
                    }]
                else:
                    messages = [{
                        "role": "user",
                        "content": [
                            {"type": "text", "text": question}
                        ]
                    }]

                completion = client.chat.completions.create(
                    model="YuanjingVL",
                    messages=messages,
                    stream=True,
                    extra_body={  # Other parameters of the model, not required
                        "api_option": "general",  # general: general; ocr: multi-modal ocr; math: taking pictures to answer questions
                    }
                )
                answer = {
                    "code": 0,
                    "message": "success",
                    "response": "",
                    "gen_file_url_list": [],
                    "history": [],
                    "finish": 0,
                    "usage": {
                        "prompt_tokens": 0,
                        "completion_tokens": 0,
                        "total_tokens": 0
                    },
                    "search_list": [],
                    "qa_type": 0
                }
                for event in completion:
                    logger.info(f"event_is {event}")
                    if event.choices:
                       answer['response'] = event.choices[0].delta.content
                       logger.info(f"finish_reason_receive_is:{event.choices[0].finish_reason}")
                       if event.choices[0].finish_reason == 'stop':
                           logger.info(f"finish_stop")
                           answer['finish'] = 1
                    yield f"data:{json.dumps(answer, ensure_ascii=False)}\n"
                return

            if mcp_tools:
                mcp_server_response = mcp_server_client(question, mcp_tools,tools_name, temperature=temperature,model_name=model,model_url=model_url,
                                                        stream=True,
                                                        history=history)
                if mcp_server_response:
                    for item in mcp_server_response:
                        logger.info('result: %s',item)
                        answer = {
                            "code": 0,
                            "message": "success",
                            "response": "",
                            "gen_file_url_list": [],
                            "history": [],
                            "finish": 0,
                            "usage": {
                                "prompt_tokens": 0,
                                "completion_tokens": 0,
                                "total_tokens": 0
                            },
                            "search_list": [],
                            "qa_type": 20
                        }
                            # Processing AIMessage content
                        if isinstance(item, AIMessage) and item.content:
                            token_usage = getattr(item, "response_metadata", {}).get("token_usage", {})

                            answer["response"] = item.content
                            answer["usage"] = {
                                "prompt_tokens": token_usage.get("prompt_tokens", 0),
                                "completion_tokens": token_usage.get("completion_tokens", 0),
                                "total_tokens": token_usage.get("total_tokens", 0)
                            }
                            #yield f"data:{json.dumps(answer, ensure_ascii=False)}\n\n"


                            tool_calls = getattr(item, "tool_calls", [])
                            if tool_calls:
                                logger.info('tool_calls is: %s', tool_calls)
                                tool_name = tool_calls[0]['name']
                                args = tool_calls[0]['args']
                                logger.info('tool_name is: %s',tool_name)
                                logger.info('args_str is: %s',args)
                                text = f"<tool>mcp-工具名：{tool_name}\n\n\n```请求参数：\n{args}\n```\n\n"
                                answer["response"] = text
                            yield f"data:{json.dumps(answer, ensure_ascii=False)}\n\n"


                        # isToolMessage
                        elif isinstance(item, ToolMessage) and item.content:
                            text1 = f"```请求结果：\n{item.content}\n```\n\n</tool>"
                            answer["response"] = text1
                            yield f"data:{json.dumps(answer, ensure_ascii=False)}\n\n"


                    # The last one is marked complete
                    answer["finish"] = 1
                    answer["usage"] = {
                        "prompt_tokens": 0,
                        "completion_tokens": 0,
                        "total_tokens": 0
                    }
                    yield f"data:{json.dumps(answer, ensure_ascii=False)}\n"
                    return
                else:
                    logger.info('mcp无法回答 大模型兜底回答')
                    llm = ChatOpenAI(
                        model_name=model,
                        streaming=True,
                        base_url=model_url,
                        openai_api_key=os.environ["ARK_API_KEY"],
                    )
                    answer = {
                        "code": 0,
                        "message": "success",
                        "response": "",
                        "gen_file_url_list": [],
                        "history": [],
                        "finish": 0,
                        "usage": {
                            "prompt_tokens": 0,
                            "completion_tokens": 0,
                            "total_tokens": 0
                        },
                        "search_list": [],
                        "qa_type": 0
                    }

                    assistant_reply = ""
                    messages.append({"role": "user", "content": question})
                    if system_role:
                        messages.append({"role": "system", "content": system_role})
                    for chunk in llm.stream(messages):
                        if hasattr(chunk, "content"):
                            print('大模型输出是:', chunk)
                            assistant_reply += chunk.content
                            answer['response'] = chunk.content

                            if hasattr(chunk, "response_metadata"):
                                if 'finish_reason' in chunk.response_metadata and chunk.response_metadata[
                                    'finish_reason'] == 'stop':
                                    updated_history = history[-4:] if len(history) > 4 else history
                                    updated_history.append({
                                        "query": question,
                                        "response": assistant_reply
                                    })
                                    answer['finish'] = 1
                                    answer["history"] = updated_history
                            if hasattr(chunk, "usage_metadata") and chunk.usage_metadata is not None:
                                answer['usage']['prompt_tokens'] = chunk.usage_metadata[
                                    'input_tokens'] if 'input_tokens' in chunk.usage_metadata else 0
                                answer['usage']['completion_tokens'] = chunk.usage_metadata[
                                    'output_tokens'] if 'output_tokens' in chunk.usage_metadata else 0
                                answer['usage']['total_tokens'] = chunk.usage_metadata[
                                    'total_tokens'] if 'total_tokens' in chunk.usage_metadata else 0
                            yield f"data:{json.dumps(answer, ensure_ascii=False)}\n"
                    return



            used_rag = False
            #If there is a knowledge base for passing parameters, go to rag first
            if use_know:
                print('进入rag问题是:',question)
                
                #url = "http://172.17.0.1:10891/rag/knowledge/stream/search"
                url = URL_RAG_STREAM
                logger.info(f"rag_url是:{url}")

                payload = {
                    "knowledgeBase": knowledgebase_name,
                    "question": question,
                    "threshold": threshold,
                    "topK": topk,
                    "temperature":temperature,
                    "do_sample":do_sample,
                    "repetition_penalty":repetition_penalty,
                    "stream": True,
                    "chitchat": False,
                    "history": history,
                    "auto_citation":auto_citation,
                    "rerank_model_id":rerank_id,
                    "custom_model_info":{"llm_model_id":model_id},
                    "rerank_mod":rerank_mod,
                    "retrieve_method":retrieve_method,
                    "weights":weights,
                    "max_history":max_history,
                    "rewrite_query":rewrite_query,
                    "term_weight_coefficient":term_weight_coefficient,
                    "metadata_filtering":metadata_filtering,
                    "metadata_filtering_conditions":metadata_filtering_conditions,
                    "knowledgeIdList":knowledgeIdList,
                    "use_graph":use_graph

                }


                headers = {
                    "Content-Type": "application/json",
                    "X-uid": userId
                }
                logger.info('rag:'+json.dumps(payload))
                # Send POST request
                response = requests.post(url, headers=headers, data=json.dumps(payload),stream=True,verify=False)
                if response.status_code == 200:
                    first_line_checked = False
                    for line in response.iter_lines():
                        if line:
                            try:
                                decoded_line = line.decode("utf-8").strip()
                                if decoded_line.startswith("data:"):
                                    decoded_line = decoded_line[len("data:"):].strip()
                                data = json.loads(decoded_line)
                                if not first_line_checked:
                                    first_line_checked = True
                                    if data["data"]["searchList"]:
                                        logger.info(f"知识库有召回")
                                        used_rag = True
                                    else:
                                        logger.info(f"知识库无召回")
                                if used_rag:
                                    answer = {
                "code": 0,
                "message": "success",
                "response": "",
                "gen_file_url_list": [

                ],
                "history": [],
                "finish": 0,
                "usage": {
                    "prompt_tokens": 0,
                    "completion_tokens": 0,
                    "total_tokens": 0
                },
                "search_list": [],
                "qa_type":1
                }
                                    answer['response']=data["data"]["output"]
                                    answer['code']=data["code"]
                                    answer['finish']=data["finish"]
                                    answer['message']=data["message"]
                                    answer['search_list']=data["data"]["searchList"]
                                    answer['history']=data["history"]
                                    
                                    yield f"data: {json.dumps(answer, ensure_ascii=False)}\n\n"
                            except Exception as e:
                                yield f"data: {json.dumps({'error': str(e)}, ensure_ascii=False)}\n\n"
            if not used_rag:               
                if use_search == True:
                    logger.info(f"走搜索")
                    #Call network search, transparently transmit the searched search_list and answers, and return the results directly
                    loop = asyncio.new_event_loop()
                    asyncio.set_event_loop(loop)

                    try:
                        llm = ChatOpenAI(
                            model_name=model,
                            base_url=model_url,
                            openai_api_key=os.environ["ARK_API_KEY"],
                        )
                        rewrite_prompt = '请根据历史信息针对本次问题进行改写，如果你认为没必要改写的就直接输出原问题即可'+'\n'+'历史信息是'+str(history)+'\n'+'本次问题是:'+question
                        logger.info(f"改写模板是:{rewrite_prompt}")
                        response = llm.invoke(rewrite_prompt)
                        rewrite_query = response.content
                        logger.info(f"改写后的问题是:{rewrite_query}")
                        
                        bing_top_k = 5
                        bing_time_out = 3
                        auto_citation = False
                        days_limit = -1
                        bing_result_len = 15
                        bing_target_success = 10
                        task = start_async_search(
                            loop, rewrite_query, bing_top_k, bing_time_out,
                            bing_target_success, bing_result_len,
                            model, days_limit, auto_citation,search_url,search_key,search_rerank_id
                        )
                        result = loop.run_until_complete(task)        
                        bing_prompt, bing_search_list = result

                        llm = ChatOpenAI(
                            model_name=model,
                            streaming=True,
                            base_url=model_url,
                            openai_api_key=os.environ["ARK_API_KEY"],
                        )
                        
                        
                        
                        
                        
                        first_chunk = True
                        answer = {
                "code": 0,
                "message": "success",
                "response": "",
                "gen_file_url_list": [],
                "history": [],
                "finish": 0,
                "usage": {
                "prompt_tokens": 0,
                "completion_tokens": 0,
                "total_tokens": 0
                },
                "search_list": [],
                "qa_type":0
                }
                        
                        



                        assistant_reply = ""
                        for chunk in llm.stream(bing_prompt):
                            if hasattr(chunk, "content"):
                                print('大模型输出是:',chunk)
                                answer['response'] = chunk.content
                                assistant_reply += chunk.content
                                answer["search_list"] = bing_search_list  # Only the first output


                                if hasattr(chunk, "response_metadata"):
                                    if 'finish_reason' in chunk.response_metadata and chunk.response_metadata['finish_reason']=='stop':
                                        updated_history = history[-4:] if len(history) > 4 else history
                                        updated_history.append({
                                            "query": question,
                                            "response": assistant_reply
                                        })
                                        answer['finish']=1
                                        answer["history"] = updated_history
                                        
                                if hasattr(chunk, "usage_metadata") and chunk.usage_metadata is not None:
                                    answer['usage']['prompt_tokens'] = chunk.usage_metadata['input_tokens'] if 'input_tokens' in chunk.usage_metadata else 0
                                    answer['usage']['completion_tokens'] = chunk.usage_metadata['output_tokens'] if 'output_tokens' in chunk.usage_metadata else 0
                                    answer['usage']['total_tokens'] = chunk.usage_metadata['total_tokens'] if 'total_tokens' in chunk.usage_metadata else 0
                                yield f"data:{json.dumps(answer,ensure_ascii=False)}\n"


                    except Exception as e:
                        print("错误:", str(e), flush=True)
                        error_response = {
                            "code": -1,
                            "message": f"error: {str(e)}",
                            "response": "",
                            "gen_file_url_list": [],
                            "history": [],
                            "finish": 1,
                            "usage": {
                                "prompt_tokens": 0,
                                "completion_tokens": 0,
                                "total_tokens": 0
                            },
                            "search_list": [],
                            "qa_type": 0
                        }
                        yield f"data:{json.dumps(error_response, ensure_ascii=False)}\n"
                        
                    finally:
                        loop.close()
                    return

                #If the tool is configured, the action will answer directly
                if plugin_list:
                    action_url = "http://localhost:1992/agent/action"
                    headers = {
                        "Content-Type": "application/json"
                    }
                    if upload_file_url:
                        question = '问题是:'+question+'\n'+'以下是调用工具可能用到的参数：'+'upload_file_url:'+upload_file_url[0]
                    else:
                        question = '问题是:' + question
                    logger.info("送入action问题是:{question}")
                    print('plugin_list是什么:',plugin_list)
                    if function_call:
                        payload = {
                            "input":question,
                            "plugin_list":plugin_list,
                            "action_type": "function_call",
                            "model_name":model,
                            "model_url":model_url,
                            "history":history
                        }
                    else:
                        payload = {
                            "input":question,
                            "plugin_list":plugin_list,
                            "action_type": "action_agent",
                            "model_name":model,
                            "model_url":model_url,
                            "history":history
                        }
                    logger.info(f"is_function_call?{function_call}")
                    response = requests.post(action_url, headers=headers, data=json.dumps(payload),stream=True,verify=False)
                    if response:
                        answer = {
    "code": 0,
    "message": "success",
    "response": "",
    "gen_file_url_list": [

    ],
    "history": [],
    "finish": 0,
    "usage": {
        "prompt_tokens": 0,
        "completion_tokens": 0,
        "total_tokens": 0
    },
    "search_list": [],
    "qa_type":20
}
                        
                        
                        assistant_reply = ""

                        for line in response.iter_lines(decode_unicode=True):
                            logger.info(f"action输出是什么:{line}")
                            if line.startswith("data:"):
                                line = line[5:]
                                datajson = json.loads(line)
                                answer['code'] = datajson['code']
                                answer['message'] = datajson['msg']
                                content_str = datajson['data']['choices'][0]['message']['content']
                                if isinstance(content_str, dict):
                                    if "search_list" in content_str:
                                        answer['search_list'] = content_str.get('search_list')
                                    if "gen_file_url_list" in content_str:
                                        answer['gen_file_url_list'] = content_str.get('gen_file_url_list')

                                    answer['response'] = content_str.get('response')
                                    assistant_reply += content_str.get('response')
                                    
                                    
                                else:
                                    answer['response'] = datajson['data']['choices'][0]['message']['content']
                                    assistant_reply += answer['response']
                                    
                                    
                                    
                                    
                                if datajson["data"]["choices"][0]["finish_reason"] == '':
                                    answer['finish']=0
                                else:
                                    answer['finish']=1
                                    updated_history = history[-4:] if len(history) > 4 else history
                                    updated_history.append({
                                        "query": question,
                                        "response": assistant_reply
                                    })
                                    answer["history"] = updated_history
                                    



                                answer['usage']['completion_tokens'] = datajson["data"]["usage"]['completion_tokens']
                                answer['usage']['prompt_tokens'] = datajson["data"]["usage"]['prompt_tokens']
                                answer['usage']['total_tokens'] = datajson["data"]["usage"]['total_tokens']


                                yield f"data:{json.dumps(answer,ensure_ascii=False)}\n"

                    else:

                        print('------没命中任何工具纯大模型回答')
                        llm = ChatOpenAI(
                            model_name=model,
                            streaming=True,
                            base_url=model_url,
                            openai_api_key=os.environ["ARK_API_KEY"],
                        )
                        answer = {
    "code": 0,
    "message": "success",
    "response": "",
    "gen_file_url_list": [],
    "history": [],
    "finish": 0,
    "usage": {
        "prompt_tokens": 0,
        "completion_tokens": 0,
        "total_tokens": 0
    },
    "search_list": [],
    "qa_type":0
}
                        assistant_reply = ""
                        messages.append({"role": "user", "content": question})
                        if system_role:
                            messages.append({"role": "system", "content": system_role})
                        for chunk in llm.stream(messages):
                            if hasattr(chunk, "content"):
                                print('大模型输出是:',chunk)
                                answer['response'] = chunk.content
                                assistant_reply += chunk.content
                            if hasattr(chunk, "response_metadata"):
                                if 'finish_reason' in chunk.response_metadata and chunk.response_metadata['finish_reason']=='stop':
                                    updated_history = history[-4:] if len(history) > 4 else history
                                    updated_history.append({
                                        "query": question,
                                        "response": assistant_reply
                                    })
                                    answer['finish']=1
                                    answer["history"] = updated_history
                            if hasattr(chunk, "usage_metadata") and chunk.usage_metadata is not None:
                                answer['usage']['prompt_tokens'] = chunk.usage_metadata['input_tokens'] if 'input_tokens' in chunk.usage_metadata else 0
                                answer['usage']['completion_tokens'] = chunk.usage_metadata['output_tokens'] if 'output_tokens' in chunk.usage_metadata else 0
                                answer['usage']['total_tokens'] = chunk.usage_metadata['total_tokens'] if 'total_tokens' in chunk.usage_metadata else 0
                            yield f"data:{json.dumps(answer,ensure_ascii=False)}\n"




                else:
                    print('------未配置任何工具纯大模型回答')
                    llm = ChatOpenAI(
                        model_name=model,
                        streaming=True,
                        base_url=model_url,
                        openai_api_key=os.environ["ARK_API_KEY"],
                    )
                    answer = {
"code": 0,
"message": "success",
"response": "",
"gen_file_url_list": [],
"history": [],
"finish": 0,
"usage": {
    "prompt_tokens": 0,
    "completion_tokens": 0,
    "total_tokens": 0
},
"search_list": [],
"qa_type":0
}
                    
                    
                    assistant_reply = ""
                    messages.append({"role": "user", "content": question})
                    if system_role:
                        messages.append({"role": "system", "content": system_role})
                    print('送给大模型的输入:messages')
                    for chunk in llm.stream(messages):
                        if hasattr(chunk, "content"):
                            print('大模型输出是:',chunk)
                            assistant_reply += chunk.content
                            answer['response'] = chunk.content

                            if hasattr(chunk, "response_metadata"):
                                if 'finish_reason' in chunk.response_metadata and chunk.response_metadata['finish_reason']=='stop':
                                    updated_history = history[-4:] if len(history) > 4 else history
                                    updated_history.append({
                                        "query": question,
                                        "response": assistant_reply
                                    })
                                    answer['finish']=1
                                    answer["history"] = updated_history
                            if hasattr(chunk, "usage_metadata") and chunk.usage_metadata is not None:
                                answer['usage']['prompt_tokens'] = chunk.usage_metadata['input_tokens'] if 'input_tokens' in chunk.usage_metadata else 0
                                answer['usage']['completion_tokens'] = chunk.usage_metadata['output_tokens'] if 'output_tokens' in chunk.usage_metadata else 0
                                answer['usage']['total_tokens'] = chunk.usage_metadata['total_tokens'] if 'total_tokens' in chunk.usage_metadata else 0
                            yield f"data:{json.dumps(answer,ensure_ascii=False)}\n"



                    
        except Exception as e:
            logger.exception("❌ 处理请求失败：")
            error_data = {
                "code": 1,
                "message": str(e),
                "response": "",
                "finish": 1
            }
            yield f"data:{json.dumps(error_data, ensure_ascii=False)}\n"

    return Response(generate(), mimetype="text/event-stream")
    
    
    
if __name__ == '__main__':
    logger.info("agent_server start")
    app.run(host='0.0.0.0', port=7258, threaded=False,debug=False)
