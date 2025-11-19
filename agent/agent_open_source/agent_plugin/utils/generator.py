import json
import requests
import logging
import datetime
import re
from flask import  stream_with_context, Response, request


logger = logging.getLogger(__name__)



def create_error_response(msg):
    """构造错误响应的函数，并设置状态码"""
    return Response(json.dumps({'code': 2, 'msg': msg, 'response': ''},ensure_ascii=False), status=400, mimetype='application/json')

def generate_stream_response(query, rewrite_query, prompt, response, qa_type, history,is_query_sens = False, need_search_list=False,search_list=[],gen_file_url_list=[],upload_file_url='',non_llm_response=None,func_name = '',func_params = '',thought_inference = ''):
    """
    模拟流式返回API响应的生成器函数，根据提供的响应内容分步骤生成流式响应。
    :param question: 提问内容。
    :param prompt: 提示信息。
    :param full_response: 完整的响应文本。
    :param history: 历史交互记录列表。
    :param need_search_list: 是否需要搜索列表。
    :param search_list: 可选的搜索列表。
    :return: 生成流式响应的单个部分。
    """
    # now = datetime.datetime.now()
       
    # cur_date = now.strftime(f"%Y-%m-%d-%H:%M:%S")
    
    logger.info("stream response start")
    # print("【Start response】:",cur_date)
    
    id = 0
    if history:
        history_tmp = history.copy()
    else:
        history_tmp = []
    
    # 定义全局变量，流式结束原因：0（生成中），1（正常结束），2（非正常结束，输出超过最大长度） [EN] Define global variables, reasons for streaming end: 0 (generating), 1 (normal end), 2 (abnormal end, output exceeds the maximum length)
    finish_reason = 0
    
    if response:
        incremental_content  = ""
        complete_content = ""
        infer_content = ""
        thought = ""
        
        # print(f"thought_inference:{thought_inference[0:10]}")
        
        if qa_type in [7,9]  and thought_inference:
            print("------start thought_inference response-----")
            
            for i in range(0, len(thought_inference), 10):
                infer_content = thought_inference[i:i + 10]
                thought += thought_inference[i:i + 10]
                finish_reason = 0

                usage = {
                    "prompt_tokens": len(infer_content),
                    "completion_tokens": len(infer_content),
                    "total_tokens": len(thought_inference)
                }

                result = {
                    "code": 0,
                    "message": "success",
                    "response": "",
                    "gen_file_url_list": gen_file_url_list,
                    "qa_type": qa_type,
                    "func_name":func_name,
                    "func_params":func_params,
                    "thought_inference":infer_content,
                    "history":history_tmp,                
                    "finish": finish_reason,
                    "usage": usage
                }

                if need_search_list:
                    result["search_list"] = search_list

                jsonarr = json.dumps(result, ensure_ascii=False)
                id += 1
                str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'          
                yield str_out


        if isinstance(response, requests.models.Response):
            # http服务方式的 response，来自语言或编码模型 [EN] HTTP service response, from language or encoding model
            for line in response.iter_lines(decode_unicode=True):
                if isinstance(line,list) and isinstance(line[0]['content'],str):
                    line = line[0]['content']
                if isinstance(line,list) and isinstance(line[0]['content'][0]['text'],str):
                    line = line[0]['content'][0]['text']
                if line.startswith("data:") and not line.startswith('data:""'):
                    # print(line)
                    line = line[5:]
                    # print(line)
                    datajson = json.loads(line)
                    
                    # print(datajson)

                    incremental_content = datajson["data"]["choices"][0]["message"]["content"]
                    
                    complete_content += datajson["data"]["choices"][0]["message"]["content"]
                    if qa_type in [7,9]:
                        
                        infer_content = datajson["data"]["choices"][0]["message"]["content"]
                        thought += datajson["data"]["choices"][0]["message"]["content"]


                    if not datajson['data']['choices'][0]['finish_reason'] :
                        finish_reason = 0

                        # subjson["needHistory"] = False
                        # subjson["response"] = incremental_content
                        # history_tmp.append(subjson)
                        result = {
                            "code": 0,
                            "message": "success",
                            "response": incremental_content,
                            "gen_file_url_list": gen_file_url_list,
                            "qa_type": qa_type,
                            "func_name":func_name,
                            "func_params":func_params,
                            "thought_inference":infer_content,
                            "history": [],
                            "finish": finish_reason,
                            "usage": datajson['data']['usage']
                        }

                        jsonarr = json.dumps(result, ensure_ascii=False)
                        id += 1
                        str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'

                        yield str_out

                    else:
                        if datajson['data']['choices'][0]['finish_reason']=="stop":  
                            # 正常结束 [EN] End normally
                            finish_reason = 1
                        if datajson['data']['choices'][0]['finish_reason']=="tool_calls": 
                            # 正常结束 [EN] End normally
                            finish_reason = 1
                            
                        if datajson['data']['choices'][0]['finish_reason']=="length":
                            # 长度被截断 [EN] length truncated
                            finish_reason = 2
                            

                        if not is_query_sens: 
                            # 将当前对话存入对话历史，但有条件，如下 [EN] Save the current conversation into the conversation history, but there are conditions, as follows
                            # only if user query is not sensitive, push the current turn into dialog history
                            subjson = {}
                            subjson["query"] = query
                            subjson["rewrite_query"] = rewrite_query
                            subjson["upload_file_url"] = upload_file_url
                            # subjson["prompt"] = prompt
                            # print(datajson['data']['choices'][0]['finish_reason'])
                            # print(gen_file_url_list)
                            if not gen_file_url_list:
                                gen_file_url_list = datajson.get("gen_file_url_list",[])
                                # print(gen_file_url_list)
                                if qa_type == 4 and len(gen_file_url_list)>0:                                    
                                    complete_content = complete_content + "\n 已处理好的文件下载地址是：" + gen_file_url_list[0]["output_file_url"]
                                    incremental_content = incremental_content  + "\n 已处理好的文件下载地址是：" + gen_file_url_list[0]["output_file_url"]
                                
                            subjson["response"] = complete_content
                            subjson["gen_file_url_list"] = gen_file_url_list
                            subjson["qa_type"] = qa_type
                            subjson["func_name"] = func_name
                            subjson["func_params"] = func_params
                            subjson["thought_inference"] = thought
                            history_tmp.append(subjson)
                            
                        result = {
                            "code": 0,
                            "message": "success",
                            "response": incremental_content,
                            "gen_file_url_list": gen_file_url_list,
                            "qa_type": qa_type,
                            "func_name":func_name,
                            "func_params":func_params,
                            "thought_inference":infer_content,
                            "history": history_tmp,
                            "finish": finish_reason,
                            "usage": datajson['data']['usage']
                        }

                        if need_search_list:
                            result["search_list"] = search_list

                        jsonarr = json.dumps(result, ensure_ascii=False)
                        id += 1
                        str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'
                        logger.info(f'generate_stream_response:{complete_content}')
                        logger.info("stream response end")

                        yield str_out

        elif qa_type==6:
             for line in response:
                
                if isinstance(line,list) and isinstance(line[0]['content'],str):
                    line = line[0]['content']
                if isinstance(line,list) and isinstance(line[0]['content'][0]['text'],str):
                    line = line[0]['content'][0]['text']
                if line.startswith("data:") and not line.startswith('data:""'):
                    # print(line)
                    line = line[5:]
                    # print(line)
                    datajson = json.loads(line)
                    
                    # print(datajson)

                    incremental_content = datajson["data"]["choices"][0]["message"]["content"]
                    
                    complete_content += datajson["data"]["choices"][0]["message"]["content"]
                    # if not gen_file_url_list:
                    #     gen_file_url_list = datajson.get("gen_file_url_list",[])

                    if not datajson['data']['choices'][0]['finish_reason'] :
                        finish_reason = 0

                        # subjson["needHistory"] = False
                        # subjson["response"] = incremental_content
                        # history_tmp.append(subjson)
                        result = {
                            "code": 0,
                            "message": "success",
                            "response": incremental_content,
                            "gen_file_url_list": gen_file_url_list,
                            "qa_type": qa_type,
                            "func_name":func_name,
                            "func_params":func_params,
                            "thought_inference":"",
                            "history": [],
                            "finish": finish_reason,
                            "usage": datajson['data']['usage']
                        }

                        jsonarr = json.dumps(result, ensure_ascii=False)
                        id += 1
                        str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'

                        yield str_out

                    else:
                        if datajson['data']['choices'][0]['finish_reason']=="stop":  
                            # 正常结束 [EN] End normally
                            finish_reason = 1
                        if datajson['data']['choices'][0]['finish_reason']=="length":
                            # 长度被截断 [EN] length truncated
                            finish_reason = 2
                            

                        if not is_query_sens: 
                            # 将当前对话存入对话历史，但有条件，如下 [EN] Save the current conversation into the conversation history, but there are conditions, as follows
                            # only if user query is not sensitive, push the current turn into dialog history
                            subjson = {}
                            subjson["query"] = query
                            subjson["rewrite_query"] = rewrite_query
                            subjson["upload_file_url"] = upload_file_url
                            # subjson["prompt"] = prompt
                            # print(datajson['data']['choices'][0]['finish_reason'])
                            # print(gen_file_url_list)
                            if not gen_file_url_list:
                                gen_file_url_list = datajson.get("gen_file_url_list",[])
                                # print(gen_file_url_list)
                                if qa_type == 4 and len(gen_file_url_list)>0:                                    
                                    complete_content = complete_content + "\n 已处理好的文件下载地址是：" + gen_file_url_list[0]["output_file_url"]
                                    incremental_content = incremental_content  + "\n 已处理好的文件下载地址是：" + gen_file_url_list[0]["output_file_url"]
                                
                            subjson["response"] = complete_content
                            subjson["gen_file_url_list"] = gen_file_url_list
                            subjson["qa_type"] = qa_type
                            subjson["func_name"] = func_name
                            subjson["func_params"] = func_params
                            subjson["thought_inference"] = ""
                            history_tmp.append(subjson)
                            
                        result = {
                            "code": 0,
                            "message": "success",
                            "response": incremental_content,
                            "gen_file_url_list": gen_file_url_list,
                            "qa_type": qa_type,
                            "func_name":func_name,
                            "func_params":func_params,
                            "thought_inference":"",
                            "history": history_tmp,
                            "finish": finish_reason,
                            "usage": datajson['data']['usage']
                        }

                        if need_search_list:
                            result["search_list"] = search_list

                        jsonarr = json.dumps(result, ensure_ascii=False)
                        id += 1
                        str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'
                        logger.info(f'generate_stream_response:{complete_content}')
                        logger.info("stream response end")

                        yield str_out           
                    
    elif non_llm_response:
        #非语言模型的消息，模拟流式结果 [EN] Messages from non-linguistic models, simulating streaming results
        finish_reason = 1
        if qa_type in [7,9]:
            thought_inference = thought_inference+"\n"+non_llm_response
        if not is_query_sens:
            # only if user query is not sensitive, push the current turn into dialog history
            subjson = {}
            subjson["query"] = query
            subjson["rewrite_query"] = rewrite_query
            subjson["upload_file_url"] = upload_file_url
            # subjson["prompt"] = prompt                
            subjson["response"] = non_llm_response
            subjson["gen_file_url_list"] = gen_file_url_list  
            subjson["qa_type"] = qa_type   
            subjson["func_name"] = func_name
            subjson["func_params"] = func_params
            subjson["thought_inference"] = thought_inference
            history_tmp.append(subjson)
        
        usage = {
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0
        }
    
        result = {
            "code": 0,
            "message": "success",
            "response": non_llm_response,
            "gen_file_url_list": gen_file_url_list,
            "qa_type": qa_type,
            "func_name":func_name,
            "func_params":func_params,
            "thought_inference":thought_inference,
            "history":history_tmp,                
            "finish": finish_reason,
            "usage": usage
        }
         
        if need_search_list:
            result["search_list"] = search_list


        jsonarr = json.dumps(result, ensure_ascii=False)
        id += 1
        str_out = f'id:{id}\nevent:result\ndata:{jsonarr}\n\n'
        logger.info(f'generate_stream_non_llm_response:{non_llm_response}')
        logger.info("stream response end")                

        yield str_out
    else:
        msg_error = "参数错误: 缺少答案正文，语言模型输出（response）和非语言模型输出（non_llm_response）不能同时为空!"
        logger.info(msg_error)
        return create_error_response(msg_error)
    


def generate_non_stream_response(query, rewrite_query, prompt, llm_response,qa_type, history,is_query_sens = False, need_search_list=False,search_list=[],gen_file_url_list=[],upload_file_url='',non_llm_response=None,func_name = '',func_params = '',thought_inference = ''):
    """
    非流式输出。
    :param llm_response: 来自大模型模型的response。    
    :param non_llm_repsonse: 非来自大模型模型的response，目前仅用于文生图场景中。
    以上两个参数二选一，不能同时为空
   
    """

    # 直接在历史记录中添加新的项 [EN] Add new items directly to history
    if history:
        history_tmp = history.copy()
    else:
        history_tmp = []
    subjson={}
    

    if llm_response:
        if qa_type in [7,9]:
            thought_inference = thought_inference+"\n"+llm_response.json()["data"]["choices"][0]["message"]["content"]
        if not is_query_sens:
            subjson["query"] = query
            subjson["rewrite_query"] = rewrite_query
            # subjson["prompt"] = prompt 
            subjson["upload_file_url"] = upload_file_url
            subjson["response"] = llm_response.json()["data"]["choices"][0]["message"]["content"]
            subjson["gen_file_url_list"]=gen_file_url_list
            subjson["qa_type"] = qa_type  
            subjson["func_name"] = func_name
            subjson["func_params"] = func_params
            subjson["thought_inference"] = thought_inference
            history_tmp.append(subjson)
        
        result = {
            "code": 0,
            "message": "success",
            "response": llm_response.json()["data"]["choices"][0]["message"]["content"],    
            "gen_file_url_list":gen_file_url_list,
            "qa_type": qa_type,   
            "func_name":func_name,
            "func_params":func_params,
            "thought_inference":thought_inference,
            "history": history_tmp,            
            "usage": llm_response.json()['data']['usage']
            }
        
        if need_search_list:
            result["search_list"] = search_list
            
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f'generate_non_stream_response: {jsonarr}')
        # 返回响应 [EN] Return response
        return Response(jsonarr, mimetype='application/json')
        
    elif non_llm_response:
        if qa_type in [7,9]:
            print(f"----非流式action输出---")
            thought_inference = thought_inference+"\n"+non_llm_response
        if not is_query_sens:
            subjson["query"] = query
            subjson["rewrite_query"] = rewrite_query
            # subjson["prompt"] = prompt 
            subjson["upload_file_url"] = upload_file_url
            subjson["response"] = non_llm_response
            subjson["gen_file_url_list"]=gen_file_url_list
            subjson["qa_type"] = qa_type  
            subjson["func_name"] = func_name
            subjson["func_params"] = func_params
            subjson["thought_inference"] = thought_inference
            history_tmp.append(subjson)
        
        usage = {
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0
        }
        
        result = {
            "code": 0,
            "message": "success",
            "response": non_llm_response,
            "gen_file_url_list":gen_file_url_list,
            "qa_type": qa_type,
            "func_name":func_name,
            "func_params":func_params,
            "thought_inference":thought_inference,
            "history": history_tmp,
            "usage": usage
        }
       
        if need_search_list:
            result["search_list"] = search_list
            
        jsonarr = json.dumps(result, ensure_ascii=False)
        logger.info(f'generate_non_stream_non_llm_response: {jsonarr}')

        # 直接返回响应 [EN] Return response directly
        return Response(jsonarr, mimetype='application/json')
    else:
        msg_error = "参数错误: 缺少答案正文，语言模型输出（response）和非语言模型输出（non_llm_response）不能同时为空!"
        logger.info(msg_error)
        return create_error_response(msg_error)