# Python standard library imports
import base64
import configparser
import datetime
import io
import json
import logging
import os
import re
import ssl
import sys
import time
import uuid
import wave
from collections import Counter
from multiprocessing.dummy import Pool
from typing import Any, Dict, List

# Third party imports
import jwt
import requests
import urllib3
#import websocket

# Local application imports
from utils.auth import AccessTokenManager
#from utils.messages import req_trim_messages
#from utils.minio import (
   # extract_and_upload_first_frame,
    #upload_base64_data
#)
from utils.output_parser import extract_json
#from utils.redis_db import RedisClient

from utils.timing import advanced_timing_decorator
#from utils.oss_manager import OSSManager
#from utils.langchain_unicomai import ChatUnicomAI
from langchain_openai import ChatOpenAI 
from langchain_core.messages import (
    AIMessage,
    AIMessageChunk,
)
#from utils.ts_tencent import ChatTencentAI
from openai import RateLimitError  # New import method

from typing import List, Dict, Union, Any
from openai import RateLimitError
import json
import re


# Disable insecure request warnings
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

HISTORY_TURNS_NUM =10

logger = logging.getLogger(__name__)

# instantiate object
token_manager = AccessTokenManager()

#redis_client = RedisClient()


MODEL_NAME = config["MODELS"]["default_llm"]
bucket_type = config["DEPLOY"]["BUCKET_TPYE"]
#MODEL_URL = config["MODELS"]["model_url"]
'''
if bucket_type == "OSS":
    oss_manager = OSSManager()
    '''

OPENAI_BASE_URL = config["MODELS"]["openai_base_url"]
DEFAULT_TEMPERATURE = config["MODELS"]["default_llm_param_temperature"]


def parse_error_to_dict(error) -> Dict[str, Any]:
    """将错误信息转换为字典类型"""
    try:
        # Extract JSON part from error message
        error_str = str(error)
        # Use regular expressions to match JSON strings following '-'
        match = re.search(r'-\s*(\{.*\})', error_str)
        if match:
            json_str = match.group(1)
            return json.loads(json_str)
        # If no JSON format is matched, basic error information is returned.
        return {
            "error": {
                "message": str(error),
                "type": type(error).__name__,
                "code": getattr(error, 'code', 'unknown')
            }
        }
    except Exception as e:
        # Ensure a valid error dictionary is always returned
        return {
            "error": {
                "message": str(error),
                "parse_error": str(e),
                "type": "error_parse_failed"
            }
        }
    
@advanced_timing_decorator(task_name="req_unicom_llm_chat")
def req_unicom_llm_chat(messages: List, 
                       stream: bool = True,
                       model_name: str = MODEL_NAME, 
                       model_url = "",
                       do_sample: bool = False,
                       repetition_penalty: float = 1.1,
                       temperature: float = DEFAULT_TEMPERATURE
                    ) -> Union[Any, Dict[str, Any]]:
    
    logger.info(f"model_name:{model_name}  \nmodel_url:{model_url}")
    logger.info(f"req_unicom_llm_chat  model_name in model_url :{model_name.lower() in model_url}")
    if "unicom" in model_name and (model_name in model_url or not model_name):
        logger.info(f"-------unicom逻辑------")
        
        llm = ChatUnicomAI(
            model_name=model_name,
            temperature=temperature,
        )
   
    else:
        if model_url == MODEL_URL:
            base_url, _, model_name = model_url.rpartition('/')
        elif model_url and model_name.lower() not in model_url:
            base_url = model_url
        elif model_name.lower() in model_url:
            base_url, _, model_name = model_url.rpartition('/')
        else:
            base_url = OPENAI_BASE_URL

        logger.info(f"model_tools req_unicom_llm_chat base_url:{base_url} \nmodel_name:{model_name}")
            
        llm = ChatOpenAI(
            model=model_name,
            
            temperature=0.6,  #Deepseek series models are written to death
            base_url=base_url,
            api_key = token_manager.get_access_token()
        )

    # logger.info(f"llm:{llm}")

    
    try:

        if stream:
            result = llm.stream(messages)  
            # print(result)
            # for line in result:
            #     logger.info(f"line:{line.content}")
            
            return result 
        else:
            return llm.invoke(messages)
            
    except RateLimitError as e:
        error_dict = parse_error_to_dict(e)
        logger.info(f"\n限速错误: {json.dumps(error_dict, indent=2, ensure_ascii=False)}")
        return None
        
    except Exception as e:
        error_dict = parse_error_to_dict(e)
        logger.info(f"\n意外错误: {json.dumps(error_dict, indent=2, ensure_ascii=False)}")
        return None

    
@advanced_timing_decorator() 
def req_unicom_llm_chat_plus(messages:List, stop_words = ['REASON'], model_name=MODEL_NAME,model_url = "", temperature=0.1):    
    def check_words_in_string(input_string,stop_words):
        for word in stop_words:
            if word in input_string:
                return True
        return False
        
    if "unicom" in model_name and (model_name in model_url or not model_url):
        llm = ChatUnicomAI(
            model_name=model_name,
            temperature=temperature,
        )
   
    else:
        logger.info(f"req_unicom_llm_chat_plus model_name in model_url :{model_name.lower() in model_url}")
        if model_url == MODEL_URL:
            base_url, _, model_name = model_url.rpartition('/')
        elif model_url and model_name.lower() not in model_url:
            base_url = model_url
        elif model_name.lower() in model_url:
            base_url, _, model_name = model_url.rpartition('/')
        else:
            base_url = OPENAI_BASE_URL
        logger.info(f"model_tools req_unicom_llm_chat_plus  \nbase_url:{base_url}  \nmodel_name:{model_name}")
        llm = ChatOpenAI(
            model=model_name,
            temperature=0.6,  #Deepseek series models are written to death
            base_url=base_url,
            api_key = token_manager.get_access_token()
        )
        logger.info(f"llm:{llm}")

    try:
        # print(messages)
        response = llm.stream(messages)
        complete_content=""
        for chunk in response:
            incremental_content = chunk.content
            complete_content += incremental_content
            # logger.info(complete_content)
            # print(complete_content)
            if check_words_in_string(complete_content,stop_words):
                break
        
        logger.info(f"chat_plus result:{complete_content}")
        return complete_content
    except RateLimitError as e:
        error_dict = parse_error_to_dict(e)
        logger.info(f"\n限速错误: {json.dumps(error_dict, indent=2, ensure_ascii=False)}")
        return complete_content
        
    except Exception as e:
        error_dict = parse_error_to_dict(e)
        logger.info(f"\n意外错误: {json.dumps(error_dict, indent=2, ensure_ascii=False)}")
        return complete_content
       
     

# Controllable version of Wenshengtu
@advanced_timing_decorator() 
def upload_base64_to_url(image_bs64_list):
    img_url_list = []
    msg = 'success'
    try:
        for image_bs64 in image_bs64_list:       
            file_extension = '.jpg'  # Assuming we know the file should be a JPEG image
            object_prefix = 'images'
            
            if bucket_type == "MINIO":
                result = upload_base64_data(image_bs64,object_prefix,file_extension)
                logger.info(f"minio upload_image_base64_to_url result: {result}")
            else:
                result = oss_manager.upload_stream(image_bs64,object_prefix,file_extension)
                logger.info(f"oss upload_image_base64_to_url result: {result}")
            #result = upload_base64_data(image_bs64 ,object_prefix, file_extension)
            # logger.info(f"upload_base64_to_url result: {result}") 
            # result = json.loads(result)
            if result.get('code') == 0:
                img_url_list.append({"text2img_url":result['url']})
                # logger.info(f"img_url_list: {img_url_list}") 
                
        return img_url_list,msg
    
    except Exception as e:
        logger.info(f"Minio没有生成url: {str(e)}")  # Record error log
        msg = str(e)
        return img_url_list,msg
    


@advanced_timing_decorator()    
def req_txt2img_plus(query,images_per_prompt=1,height=1024,width=1024):
    
    filepath = None
    img_url_list = []
    msg = 'success'
    advanced_opt = {
            "height": height,
            "width": width,
            "num_images_per_prompt":images_per_prompt,
            "sampling_steps_prior":20,
            "cfg_scale_prior":4.0,        
            "sampling_steps_decoder": 10,
            "cfg_scale_decoder":0.0
        }
    advanced_opt_str = json.dumps(advanced_opt, ensure_ascii=False)
    data = {
        "algo": "txt2img",
        "model": "metaview_t2i",
        "prompt": query,
        "advanced_opt": advanced_opt_str
    }
    headers = { "Authorization": f"Bearer {token_manager.get_access_token()}"}

    # url = 'https://122.13.25.19:5001/openapi/v1/metaview_t2i'
    url  = config["MODELS"]["default_txt2img_url"]
    logger.info(f"req_txt2img_plus url: {url}")
    logger.info(f"req_txt2img_plus headers: {headers}")
    logger.info(f"req_txt2img_plus data: {data}")

    try:
        # Get start time
        start_time = time.perf_counter()
        response = requests.post(url,
                                 data=data, 
                                 headers=headers,
                                 verify=False)
        # Calculate execution time
        end_time = time.perf_counter()
        execution_time = end_time - start_time
        # Build log messages
        logger.info(f"文生图模型原始结果：{response.text}")
        logger.info(f"文生图模型耗时：{execution_time}")
        
        if response.status_code == 200:
            result = response.json()
          
            image_bs64_list = response.json().get('result')
            
            if image_bs64_list:
                img_url_list,msg = upload_base64_to_url(image_bs64_list)
            else:
                msg = "image_bs64_list 获取失败"
                
        else:
            msg = response.reason
    
    except Exception as e:
        logger.error("没有成功生成图片呢: %s", e)  # Record error log
        msg = str(e)
    
    return img_url_list, msg



# Picture and text question and answer plus
@advanced_timing_decorator() 
def unicom_vqa_stream(image_url, prompt,messages):
    # url = "https://122.13.25.19:5001/openapi/v1/qwenvl_ft"
    url  = config["MODELS"]["vqa_url"]
    if bucket_type == "MINIO":
        try:
        
            image_response = requests.get(image_url, verify=False)
            image_response.raise_for_status()  # Make sure the request is successful
        except requests.RequestException as e:
            logger.error("Image download error: %s", e)  # Record error log
            return "No answer found due to image download error"
        # Create a temporary BytesIO object using the obtained image content
        image_file = io.BytesIO(image_response.content)
    else:
        BUCKET_NAME = config["OSS"]["BUCKET_NAME"]
        object_name = image_url.replace(f"/{BUCKET_NAME}/","")
        image_response = oss_manager.get_from_oss(object_name)
        data = image_response["data"]
        image_file = io.BytesIO(data)

    # Prepare request data
    files = {'img': ('image.jpg', image_file, 'image/jpeg')}
    messages = json.dumps(messages,ensure_ascii=False)

    data = {
        "prompt": prompt,
        "stream": True,
        "history":messages
    }
    logger.info(f"unicom_vqa_stream data :{data}")
    headers = {
            "Authorization": f"Bearer {token_manager.get_access_token()}"
            }
    # Send POST request
    logger.info(f"unicom_vqa_stream files:{files}")
    try:
        response = requests.post(url, 
                                 files=files,
                                 headers=headers,
                                 data=data, 
                                 stream=True,
                                 verify=False
                                )
        # logger.info(f"unicom_vqa_stream response:{response.text}")
        id=0
        for line in response.iter_lines(decode_unicode=True):
            # print(line)

            datajson = json.loads(line)
            increase_content = datajson.get('result', {}).get('text')

            finish_reason = datajson.get('result', {}).get('finish_reason')

            code = datajson.get('code')
            msg = datajson.get('message')

            choices = [{"index": 0, "message": {"role": "assistant", "content": increase_content}, "finish_reason": finish_reason}]
            usage = datajson.get('usage', {'prompt_tokens': 0, 'completion_tokens': 0, 'total_tokens': 0})
            
            input_tokens = usage.get('prompt_tokens')
            output_tokens = usage.get('completion_tokens')
            total_tokens = usage.get('total_tokens')


            response_metadata={'finish_reason': finish_reason, 'model_name': ''}

            usage_metadata={'input_tokens': input_tokens, 'output_tokens': output_tokens, 'total_tokens': total_tokens}
            
            ################'prompt_tokens'

            # content='smart' additional_kwargs={} response_metadata={} id='bece9cbc-e38f-11ef-a1b4-b64ea7f91903' usage_metadata={'input_tokens': 36, 'output_tokens': 7, 'total_tokens': 43}

            ai_message_chunk = AIMessageChunk(content=increase_content,usage_metadata=usage_metadata,response_metadata=response_metadata)
            
            yield ai_message_chunk

    except requests.RequestException as e:
        logger.info(f"API error: {str(e)}")
        logger.error("API error: %s", e)  # Record error log
        print(e)
        return "No answer found due to API error"

# Picture and text question and answer plus

@advanced_timing_decorator()    
def unicom_vqa_nonstream(image_url, prompt,messages):
    # url = "https://122.13.25.19:5001/openapi/v1/qwenvl_ft"
    url  = config["MODELS"]["vqa_url"]

    if bucket_type == "MINIO":
        try:
            image_response = requests.get(image_url, verify=False)
            image_response.raise_for_status()  # Make sure the request is successful
        except requests.RequestException as e:
            logger.error("Image download error: %s", e)  # Record error log
            return "No answer found due to image download error"
        # Create a temporary BytesIO object using the obtained image content
        image_file = io.BytesIO(image_response.content)
    else:
        object_name = image_url.replace("/assistant-obj/","")
        image_response = oss_manager.get_from_oss(object_name)
        data = image_response["data"]
        image_file = BytesIO(data)
        
    # Prepare request data
    files = {'img': ('image.jpg', image_file, 'image/jpeg')}
    messages = json.dumps(messages,ensure_ascii=False)

    data = {
        "prompt": prompt,
        "stream": False,
        "history":messages
    }
    logger.info(f"unicom_vqa_nonstream data :{data}")
    logger.info(f"unicom_vqa_nonstream files :{files}")
    headers = {
            "Authorization": f"Bearer {token_manager.get_access_token() }"
            }
    # Send POST request
    try:
        response = requests.post(url, 
                                 files=files,
                                 headers=headers,
                                 data=data,
                                 verify=False
                                )
      
        res_data = {"code": 200, "msg": "Success", "data": {"result": "Test result"}}
        datajson = response.json()
        full_content = datajson.get('result', {}).get('text')
        finish_reason = datajson.get('result', {}).get('finish_reason')

        choices = [{"index": 0, "message": {"role": "assistant", "content": full_content}, "finish_reason": finish_reason}]
        usage = datajson.get('usage', {'prompt_tokens': 0, 'completion_tokens': 0, 'total_tokens': 0})

        code = datajson.get('code')
        msg = datajson.get('message')

        res_data = {"code":code,"msg":msg,"data":{"choices":choices,"usage":usage}}

        return res_data
    except requests.RequestException as e:
        logger.info(f"API error: {str(e)}")
        logger.error("API error: %s", e)  # Record error log
        print(e)
        return "No answer found due to API error"
    
# Vincent Video
@advanced_timing_decorator()      
def request_text_to_video(prompt):
    # url = "https://122.13.25.19:5001/openapi/text-to-video/v1"
    url  = config["MODELS"]["txt2vid_url"]
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }
    data = {
        "prompt": prompt
    }
    data_json = json.dumps(data)
    
    try:
        response = requests.post(url, headers=headers, data=data_json, verify=False)      
        return response.json().get("resourceId")
    except Exception as e:
        logger.info(f"error:{str(e)}")
        return None

@advanced_timing_decorator()        
def get_video_base64(resource_id):
    # url = "https://122.13.25.19:5001/openapi/text-to-video/get"
    url  = config["MODELS"]["vid64_url"]
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }

    data = {
        "resourceId": resource_id
    }
    data_json = json.dumps(data)
    # print("SSSSSSSSS")

    try:
        response = requests.post(url, headers=headers, data=data_json, verify=False)
        return response.json().get("data",{}).get("video")
    except Exception as e:
        print(str(e))
        return None


@advanced_timing_decorator()    
def video_upload_to_MinIO(video_base64):
    video_url = None
    cover_pic_url = None
    result = {}
    
    try:
        object_prefix = "videos"
        file_extension = ".mp4"
        if bucket_type == "MINIO":
            upload_result = upload_base64_data(video_base64,object_prefix,file_extension)
            logger.info(f"minio upload_video_base64_to_url result: {upload_result}")
        else:
            upload_result = oss_manager.upload_stream(video_base64,object_prefix,file_extension)
            logger.info(f"oss upload_video_base64_to_url result: {upload_result}")

        video_url = upload_result.get("url")

        if video_url:
            if bucket_type == "MINIO":
                cover_pic_url = extract_and_upload_first_frame(video_url)
            #else:
            #    cover_pic_url = oss_manager.oss_extract_and_upload_first_frame(video_url)
            logger.info(f"cover_pic_url: {cover_pic_url}")
            
        result = {
                    "video_url":video_url,
                    "cover_pic_url":cover_pic_url
                 }
        return result
    except Exception as e:
        logger.info(f"error:{str(e)}")
        return result
    

# @advanced_timing_decorator()    
# def generate_and_upload_video(prompt):
#     # Step one: Generate video
#     resource_id = request_text_to_video(prompt)

#     # Wait 5 seconds: video generation takes 3-5 seconds
#     time.sleep(4)

#     # Step 2: Get video details
#     video_base64 = get_video_base64(resource_id)
#     # print("SSS:\n",video_base64,"\nHHH")

#     # if not video_details or 'video' not in video_details:
#     #     return json.dumps({"code": 1, "message": "Failed to get video details", "url": ""})
#     # Step 3: Upload video to MinIO
    
#     result = video_upload_to_MinIO(video_base64)
#     return result


@advanced_timing_decorator()    
def generate_and_upload_video(prompt):
    # Step 1: Generate video
    resource_id = request_text_to_video(prompt)
    # If the generation is not successful, failure information is returned.
    if not resource_id:
        return json.dumps({"code": 1, "message": "Failed to generate video", "url": ""})

    # Step 2: Poll to obtain video details, wait up to 30 seconds
    start_time = time.time()
    video_base64 = None
    while time.time() - start_time < 30:
        video_base64 = get_video_base64(resource_id)
        if video_base64:
            break
        time.sleep(1)  # Check every second

    # If the acquisition of video information fails, failure information will be returned.
    if not video_base64:
        return json.dumps({"code": 1, "message": "Failed to get video details within 10 seconds", "url": ""})

    # Step 3: Upload video to MinIO
    result = video_upload_to_MinIO(video_base64)
    return result



@advanced_timing_decorator()    
def get_intent_cls(query):
    
    url = config["MODELS"]["default_intent_cls_url"]
    result = ''
    try:
        result = req_bert_cls(query,url)
        return result
    except Exception as e:
        print("发生异常：", str(e))
        return 1


@advanced_timing_decorator() 
def get_query_rewrite_cls(query):
    url = config["MODELS"]["default_query_rewrite_cls"]
    try:
        result = req_bert_cls(query,url)
        return result
        # print('query_rewrite_cls: %d',result)
    except Exception as e:
        print("发生异常：", str(e))
        return 0
    

def req_bert_cls(query,url):  
    start_time = datetime.datetime.now()    
    headers =  {
        #"Authorization": f"Bearer {token_manager.get_access_token()}"
        "Authorization": f""
    }
    data = {
        "prompt": query
    }

    response = requests.post(url, headers=headers, data=data)
    logger.info(f"req_bert_cls  BERT-{url}: {response.text}")

    if response.status_code == 200:
        result_int = int(response.json().get("result",1))
        finish_time1 = datetime.datetime.now()
        time_difference1 = finish_time1 - start_time
        # print("BING  query rewirte time:", time_difference)
        # logger.info(f"BERT-{url}: {time_difference1}")
        return result_int

    else:
        # If not 200, throw a custom exception
        raise Exception("请求失败，错误信息：" + response.text)



@advanced_timing_decorator() 
def req_code_interpreter(query,need_file=False,upload_file_url='',history=[]):
    # url = "http://192.168.0.195:7257/api/cal"
    url = config["MODELS"]["default_intepreter_url"]
    # print(url)
    # print(query)
    if bucket_type == "OSS":
        P_ENDPOINT = config["OSS"]["P_ENDPOINT"]
        upload_file_url = f"{P_ENDPOINT}" + upload_file_url

    headers = {
        "Content-Type": "application/json"
    }

    # initial request
    payload = {
        "input": query,
        "history": [],
        "need_file": need_file,
        # "file_name": file_name,
        "upload_file_url": upload_file_url
    }

    # Make a streaming request
    response = requests.post(url, json=payload, headers=headers, stream=True)        
 
    return response




@advanced_timing_decorator()    
def file_to_base64(file_path):
    with open(file_path, 'rb') as file:
        file_content = file.read()
        base64_encoded = base64.b64encode(file_content)
        return base64_encoded.decode('utf-8')

@advanced_timing_decorator()       
def is_query_sen(query):
    # url = "http://192.168.0.217:1005/check"
    # URL configuration verification to increase compatibility with privatized versions
    url = config["MODELS"]["sensitive_url"]
  
    contents = []
    contents.append(query)
    headers = {
        "Content_type": "multipart/form-data",
        # "Authorization": f"Bearer {token_manager.get_access_token()}"
    }
    data = {
        "contents": contents
    }

    try:
        response = requests.post(
            url, 
            json=data, 
            # headers=headers, 
            verify=False
        )
        
        return response

    except Exception as e:
        logger.error("is_query_sen error: %s", str(e))  # Record error log
        return None


@advanced_timing_decorator()    
def req_action(query,plugin_list = None,function_calls_list = None,action_type = "qwen_agent",history = []):
    data = {"input":query,
            "plugin_list":plugin_list,
            "function_calls_list":function_calls_list,
            "action_type":action_type,
            "history":history
           }
    headers = {
        "Content-Type": "application/json"
    }
    url = config["ACTION"]["default_server_url"]
    
    try:
        response = requests.post(url, json=data, verify=False)
        logger.info(f"req_action response:{response.text}")
        # if response.status_code == "200":
        #     response = response.json()
        #     return response
        return response

    except Exception as e:
        logger.error("req_action error: %s", str(e))  # Record error log
        
        return e

@advanced_timing_decorator()  
def req_choose_cls(query,userId ,kn_list,kn_topK,auto_citation = False):
    data = {"query":query,
            "userId":userId,
            "kn_list":kn_list,
            "kn_topK":kn_topK,
            "auto_citation":auto_citation
           }
    headers = {
        "Content-Type": "application/json"
    }
    url = config["MODELS"]["default_query_choose_url"]
    
    try:
        response = requests.post(url, json=data, verify=False)        
        response = response.json()
        if response['code'] == 0:
            data = response['data']
            logger.info(f"req_choose_cls data :{data}")
            return data
            

    except Exception as e:
        logger.error("is_query_sen error: %s", str(e))  # Record error log
        return e    

@advanced_timing_decorator()  
def req_query_rewrite(query,history=[]):
    has_rewrite_query = any("rewrite_query" in item.keys() for item in history) 
    hitory_temp = []
    if has_rewrite_query:
        hitory_temp = [{"query": item.get("rewrite_query"), "response": item.get("response")} for item in history]
    else:
        for i in range(0,len(history),2):
            hitory_temp.append({"query": history[i]['content'], "response": history[i+1]['content']}) 
    
    history1=[
    {
      "query": "李安是谁",
      "response": "李安是一位杰出的华语电影导演，他以其深刻的艺术造诣和精湛的电影技艺在国际影坛享有极高的声誉。李安出生于台湾，后移民至美国，在那里接受电影教育并开始了他的电影生涯。"
    },
    {
      "query": "他有哪些作品",
      "response": "李安的一些著名作品包括《喜宴》、《卧虎藏龙》、《断背山》、《少年派的奇幻漂流》等，这些作品不仅商业上取得成功，也在国际影坛上获得了广泛认可，赢得了包括奥斯卡奖在内的多个电影奖项。特别是《卧虎藏龙》和《少年派的奇幻漂流》，它们的成功使得李安成为了连接东西方文化的桥梁，展现了他在跨文化交流中的独特魅力。"
    }
  ]



    history2=[
    {
      "query": "繁花是什么",
      "response": "电视剧《繁花》是一部在2024年引起广泛讨论的作品，它以1990年代的上海为背景，讲述了主人公阿宝等人在改革开放初期经历的故事，涉及到商战、情感以及城市生活的多重维度。该剧不仅因为其浓郁的海派风格和王家卫式的表达而深受欢迎，还因其在细节上对时代特色的精准捕捉，如剧中频繁出现的上海地标、食物和文化习俗等，使得这部电视剧成为展现上海城市文化的重要窗口。电视剧的热播也带动了相关旅游和文化产品的热销，例如相关的美食攻略、打卡地点等。"
    },
    {
      "query": "北京今天天气怎么样",
      "response": "今天是个大晴天，适合外出"
    }
  ]


    history3=[
    {
        "query": "联通董事长是谁？",
        "response": "联通董事长是陈忠岳。"
    },
    {
        "query": "联通总经理是谁？",   
        "response": "联通总经理是简勤。"
    },
    {
        "query": "联通数科的全称。",
        "response": "联通数科的全称是联通数字科技有限公司。"
    },
    {
        "query": "联通数科的董事长是谁？",
        "response": "联通数科的董事长是李广聚。"
    }
]


    history4=[
    {
      "query": "中国哪些省份在南方",
      "response": "中国南方主要包括广东、福建、广西、湖南、江西、湖北、重庆、四川、贵州、安徽、云南等11个省、自治区和直辖市。这些省份位于中国的南部，气候温暖湿润，地形复杂多样，包括了山地、丘陵、平原等多种地貌。"
    },
    {
      "query": "这些省份哪几个在一个经度上",
      "response": "在中国的地图上，省份并不是严格按经度划分的，因为中国东西跨度非常广，所以大部分省份都跨越了多个经度。但是，如果我们要找出在大致相同经度上的省份，可以考虑中国的地理位置和省份的地理位置。例如，广东、福建、广西、湖南、江西、湖北这几个省份主要位于东经110°至120°之间，这个范围内的经度跨度。重庆、四川、贵州、安徽、云南则主要位于东经105°至120°之间。虽然每个省份内部跨越多个经度，但它们的主要部分都位于这个大致的经度区间内。需要注意的是，这里提到的经度是基于中国标准时间（CST）即北京时间，也就是东八区（UTC+8）。因此，上述省份的主要部分都位于东经110°至120°之间，而重庆、四川、贵州、安徽、云南则主要位于东经105°至120°之间。"
    },
    {
      "query": "这些省份中哪些省说粤语",
      "response": "在中国南方省份中，广东省是主要的粤语使用省份。粤语，也称为广东话或广州话，主要流行于广东省中部和南部，尤其是广州、深圳、佛山、东莞、中山、珠海等珠江三角洲地区城市。此外，香港和澳门两个特别行政区也以粤语为官方语言之一。广西壮族自治区的部分地区，尤其是南宁市周边地区，也有使用粤语的情况，但粤语在广西并不是主流方言。广东省以外的其他省份，如福建、湖南、江西、湖北等，主要使用的是其他类型的汉语方言，如闽南语、湘语、赣语等，而不是粤语。"
    }
  ]


    history5=[
    {
        "query": "子夜的作者是谁？",
        "response": "茅盾"
    },
    {
        "query": "联通远景大模型简介",
        "response": "联通元景大模型体系及系列人工智能创新成果发布。该大模型的设计理念不再一味追求大模型的参数量，而是将参数量和效果放在同等重要位置。它针对不同的功能需求提供了多个参数版本，以提供最具性价比的模型。例如，对于涉诈识别等短文本识别需求，使用10亿参数版本即可满足服务需求；而对于工单分类识别等长文本识别需求，则采用70亿参数版本；对于客服等垂直行业问答，采用130亿版本；对于需要更多知识量的通用问答，则采用340亿、700亿版本。此外，为了在不同细分场景中提供更合适的功能，联通元景采用了“模型+工具”的模式，允许用户根据不同场景选择使用大模型底座内置的或自己定义的插件工具，使模型通用能力和工具专业能力相互补充，解决了大模型在实体经济发展中的“最后一公里”问题。目前，中国联通的大模型已经在网络、客服、反诈、工业、政务等多个领域得到应用。"
    },
    {
        "query": "今天北京天气",
        "response": "今天的天气是晴朗，最高温度为30摄氏度，最低温度为14摄氏度。风向为西南，风力等级为4级，湿度为56%。\n"        
    },
    {
        "query": "今天武汉的天气怎么样？",
        "response": "今天的天气是阴天，最高温度为30摄氏度，最低温度为18摄氏度。\n"       
    }
  ]

    history6= [
    {
        "query": "联通董事长是谁？",
        "response": "联通董事长是陈忠岳。",
        "qa_type": 3
    },
    {
        "query": "联通总经理是谁？",
        "response": "联通总经理是简勤。",
    },
    {
        "query": "今天北京天气",
        "response": "今天的天气是晴朗，最高温度为30摄氏度，最低温度为14摄氏度。风向为西南，风力等级为4级，湿度为56%。\n",
        
    },
    {
        "query": "上海明天的天气怎么样",
        "response": "上海明天的天气预计是晴朗的，最低气温将达到1℃，而最高气温则会上升到7℃。风向方面，预计不会有持续的风向，风速小于3级。此外，紫外线指数较低，属于最弱级别，因此只需涂抹SPF8-12的防晒护肤品即可。总体来看，明天上海的天气状况良好，适合户外活动。\n",
        
    }
  ]

    history7 = [
      {
        "query": "子夜的作者是谁",
        "response": "茅盾"
      },
      {
        "query": "茅盾还有哪些代表作？",
        "response": "《子夜》是中国现代著名作家茅盾的代表作之一。茅盾的其他代表作品还包括《蚀》三部曲（《幻灭》、《动摇》、《追求》）、《林家铺子》、《春蚕》、《秋收》、《残冬》等。这些作品都展示了茅盾对社会现实的深刻关注和对人性的深入剖析。"
      },
      {
        "query": "北京今天天气怎么样？",
        "response": "今天白天晴，最高气温27℃，东风三级，相对湿度40%。预计今天夜间晴，最低气温15℃。请注意防暑降温和补充水分。"
      },
      {
        "query": "明天北京天气怎么样？",
        "response": "明天白天晴，最高气温27℃，东风三级，相对湿度40%。预计明天夜间晴，最低气温15℃。请注意防暑降温和补充水分。"
      },
      {
        "query": "今天长沙天气怎么样？",
        "response": "根据我了解到的信息，长沙今天的天气是阴，气温为19℃，风力小于三级。"
      }
    ]
    
    history8 =  [{"query": "这是什么？", "response": "这是一张显示砖墙的照片，上面安装着一个电表箱，周围缠绕着许多电线。电表箱是灰色的，上面有多个电表，用于测量和显示电力使用情况。电表箱上方有一个金属支架，上面有绝缘子，用于支撑和固定电线。照片中没有人物或动物。",}, {"query": "能看出电线有多少年了吗","response": "从图片中无法确定电线的确切年龄。电线看起来有些老化，但如果没有关于电线安装日期或维护记录的额外信息，就无法准确判断其年龄。电线老化可能导致安全问题，因此定期检查和维护对于确保电气系统的安全至关重要。"}, {"query": "这只猫现在什么心情？",  "response": "这张照片中确实有一只猫。它看起来心情平静，可能有点困倦，因为它正躺在木制表面上，头靠在爪子上，眼睛半闭着。猫通常会表现出这种放松的姿势，当它们感到舒适和满足时。"}]
    

    history9 =  [{'query': '介绍下刘亦菲近况', 'response': '刘亦菲近期参与了电视剧《玫瑰的故事》的拍摄，她在剧中饰演女主角黄亦玫。在这部剧中，刘亦菲以其精湛的演技和出众的颜值受到了观众的好评。尽管已经36岁，但她保养得非常好，几乎没有岁月的痕迹，饰演大学生毫无违和感。她的表演赋予了角色在不同年龄阶段的独特状态，无论是少女时期的天真无邪还是成熟后的沉稳干练，都被她演绎得淋漓尽致。\n\n此外，刘亦菲还参加了某品牌的晚宴活动，与国际知名女星安妮·海瑟薇同台亮相。在活动中，刘亦菲身穿华丽礼服，展现了其独特的东方韵味和优雅气质。她的造型和妆容都备受赞誉，成为了当晚的焦点之一。\n\n总的来说，刘亦菲近期在影视和时尚领域都有着出色的表现，她的才华和美丽继续吸引着广大粉丝的关注和支持。'}, {'query': '刘亦菲参加过哪些综艺节目', 'response': '刘亦菲参加过的综艺节目包括《大片起来嗨》第一季、《抱走吧！爱豆》、《天天向上》、《快乐大本营》、《非常静距离》、《可凡倾听》、《艺术人生》等。在这些节目中，她在《快乐大本营》中与《露水红颜》剧组的唐嫣一同登台，不仅参与游戏环节，还演唱了歌曲《因为爱情》，展示了她的多才多艺。此外，她还参加了《开门大吉》等节目。这些综艺节目让观众更加了解了刘亦菲的魅力和才华。'}]
    messages = []
    
    if history:
        system_prompt = "你唯一的任务是根据对话历史完成补全用户当前问题，注意不要直接回答用户当前问题，这不是你的任务。"
        prompt = (
            '你是一个多轮对话query改写助手，根据对话历史，完成指代消岐，补全用户问题。我将举多个改写示例，每个示例包括对话历史，用户问题，重述问题3个部分。' +
            '\n```' +
            '##Example1：' +
            '对话历史：' + str(history1) + '\n' +
            '用户问题:\n他的这四部电影上映年份是什么时候' + '\n' +
            '重述问题:\n李安的《喜宴》、《卧虎藏龙》、《断背山》、《少年派的奇幻漂流》这四部电影上映年份是什么时候' +
            '\n' +
            '##Example2：' +
            '对话历史：' + str(history2) + '\n' +
            '用户问题:\n这部剧主角是谁' + '\n' +
            '重述问题:\n繁花这部剧的主角是谁' +
            '\n' +
            '##Example3：' +
            '对话历史：' + str(history3) + '\n' +
            '用户问题:\总裁是谁呢' + '\n' +
            '重述问题:\n联通数科的总裁是谁呢' +
            '\n' +
            '##Example4：' +
            '对话历史：' + str(history4) + '\n' +
            '用户问题:\n你提到的省份它的当地人饮食习惯是什么' + '\n' +
            '重述问题:\n广东省的当地人的饮食习惯是什么' +
            '##Example5：' +
            '对话历史：' + str(history5) + '\n' +
            '用户问题:\n他还有哪些代表作' + '\n' +
            '重述问题:\n茅盾还有哪些代表着' +
            '\n' +       
            '##Example6：' +
            '对话历史：' + str(history6) + '\n' +
            '用户问题:\n后天呢？' + '\n' +
            '重述问题:\n后天上海天气' +
            '\n' +       
            '##Example7：' +
            '对话历史：' + str(history7) + '\n' +
            '用户问题:\他的夫人是谁？' + '\n' +
            '重述问题:\n茅盾的夫人是谁' +
            '\n' +       
            '##Example8：'+
            '对话历史：' + str(history8) + '\n' +
            '用户问题:他能卖多少钱？' + '\n' +
            '重述问题:\n这只猫能卖多少钱' +
            '\n' +       
            '##Example9：'+
            '对话历史：' + str(history9) + '\n' +
            '用户问题:生成一张他的图片' + '\n' +
            '重述问题:\n生成一张刘亦菲的图片' +
            '\n```' +
            '\n' + 
            '请根据以上样例，补全问题。' +
            '\n' +        
            '当前对话历史:' + str(hitory_temp) + '\n' +
            '当前用户问题:' + query+ '\n' +
            '注意你唯一的任务是根据对话历史完成补全用户问题，不要直接回答用户当前问题，这不是你的任务!'
        )
        messages = [ 
                {  "role":"system", "content":system_prompt},
                {   "role":"user",  "content":prompt }
               ]
        
    else:
        prompt = (
            '你是一个多轮对话query澄清助手，对于依赖上文背景但是又没有提供上文背景的query生成一个追问的话术,注意不要原样复述用户的query。' +
            '\n```' +
            '##Example1：' +
            '用户问题:\n他的这四部电影上映年份是什么时候' + '\n' +
            '重述问题:\n请问您指的是哪位导演或演员的哪四部电影？请提供更多信息，以便我能准确回答您的问题。' +
            '\n' +
            '##Example2：' +
            '用户问题:\n这部剧主角是谁' + '\n' +
            '重述问题:\n请您告诉我您指的是哪部剧，这样我才能准确地回答您的问题。' +
            '\n' +
            '##Example3：' +
            '用户问题:\总裁是谁呢' + '\n' +
            '重述问题:\n抱歉，我需要更多信息才能回答您的问题。请您告诉我您指的是哪位总裁，或者哪部剧的总裁角色，这样我才能更好地帮助您。' +
            '\n' +           
            '##Example4：' +
            '用户问题:\n他还有哪些代表作' + '\n' +
            '重述问题:\n抱歉，您指的是哪一位？请您提供更多信息，以便我能更好地回答您的问题' +
            '\n' +       
            '##Example5：' +
            '用户问题:\n后天呢？' + '\n' +
            '重述问题:\n抱歉，请您提供更多信息，以便我能更好地回答您的问题' +
            '\n' +       
            '##Example6：' +
            '用户问题:\他的夫人是谁？' + '\n' +
            '重述问题:\n您指的是哪一位的夫人？请您提供更多信息，以便我能准确回答您的问题。' +
            '\n```' +            
            '\n' + 
            '请根据以上样例，请补全当前问题缺失的上下文信息，输出重述问题' +
            '\n' +        
            '当前用户问题:' + query
        )
        messages = [    {   "role":"user",  "content":prompt }   ]

    # print(messages)
    
    response = req_unicom_llm_chat(messages, stream=False)
    
    result = response.content
    result = result.replace("重述问题:", "").replace("重述问题：", "")
    logger.info(f"req_query_rewrite:  query:{query}  +histroy:{hitory_temp}  --->{result}")

    return result


@advanced_timing_decorator(task_name="is_kb_belong_to_user_v2")
def is_kb_belong_to_user_v2(kn_list, userId):

    # url = "http://192.168.0.214:7801/rag/list-knowledge-base" #Load address
    url = config["MODELS"]["default_list_kb_url"]
    
    # initialization return value
    code = 0
    msg = f"当前用户下不存在知识库：{kn_list}，请检查知识库名字是否正确"
    
    data = {"userId": userId}
    
    data_json = json.dumps(data)
    try:
        response = requests.post(url, data=data_json, verify=False)
        # print(response.text)
        
        if response.status_code == 200:
            response_json = response.json()   
            user_kb_list = response_json.get("data", {}).get("knowledge_base_names", [])
            
            # Use a set to determine whether kn_list is a subset of user_kb_list
            kn_set = set(kn_list)
            user_kb_set = set(user_kb_list)
            # print(user_kb_set)
            # print(kn_set)
   
            if kn_set.issubset(user_kb_set):
                code = 1
                msg = "success"
            else:
                difference_set = kn_set - user_kb_set
                code = 0 
                not_in_kb_str = "、".join(difference_set)
                
                msg = f"以下知识库在当前用户下未找到：{not_in_kb_str}。请检查上述知识库名字是否正确。"
        
        return code, msg
    except Exception as e:
        code = -1
        logger.error("is_kb_belong_to_user_v2 error: %s", str(e))  # Record error log
        print(f"An error occurred: {e}")  # Used for debugging and can be deleted
        msg = str(e)
        return code, msg



@advanced_timing_decorator()    
def req_kb_search_v2(question, kn_userId, kn_name, threshold, topK,auto_citation=False):
    
   
    print("start req_kb_search...")
    prompt = ''
    search_list = []
    # url = "http://192.168.0.214:7801/rag/search-knowledge-base" # Payload address
    url = config["MODELS"]["default_search_kb_url"]
    headers = {
        "Content_type": "application/json;charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }

    if auto_citation:
        rag_prompt = config["RAG"]["RAG_PROMPT_AUTO"]
    else:
         rag_prompt = config["RAG"]["RAG_PROMPT"]

    data = {
        "userId":kn_userId,
        "knowledgeBase": kn_name,
        "prompt_template":rag_prompt,
        "question": question,
        "threshold": threshold,
        "topK": topK,
        "stream": False,
        "search_field":"emc",
        "auto_citation":auto_citation
    }
    logger.info(f"请求入参为: {data}")

    response = requests.post(url, headers=headers, data=json.dumps(data), verify=False, stream=False)
    
    
    if response.status_code ==200:
        # print( response.json())
        search_list = response.json().get("data",{}).get("searchList",[])
        prompt = response.json().get("data",{}).get("prompt","")
        # print(response.text)
        if search_list:
            for i in range(len(search_list)):
                search_list[i]["type"]="RAG"

    return prompt, search_list





KN_PROMPT_TEMPLATE = """文本内容：
{context}

根据上述文本内容，简洁和专业的来回答用户的问题。如果无法从中得到答案，请说 "根据已知信息无法回答该问题"。回答必须是原文提及内容，答案请使用中文。 用户问题是：{question}"""


@advanced_timing_decorator()    
def can_answer_question(question, search_list,model="deepseek-v3"):  
   
    kb_result_str =  "\n\n".join([f"## {index + 1}：{item['snippet']}" for index, item in enumerate(search_list)])
    KN_USED_PROMPT_TEMPLATE = '''根据下面参考信息，判断是否可以回答用户问题
    
    【参考信息】：
    ```
     {context} 
    ```   

     【用户问题】
     ```
     {question}
     ```
     
     输出要求：
     ## If you cannot get an answer from it, please answer "0" and explain the reason. If you can get an answer from it, please answer "1".
     ## Please answer in json format, the keys are TYPE and REASON. '''

    prompt = KN_USED_PROMPT_TEMPLATE.format(context=kb_result_str, question=question)
    logger.info(f"can_answer_question prompt:{prompt}")
    messages = []
    subjson = {}
    subjson["role"] = "user"
    subjson["content"] = prompt
    messages.append(subjson)
    
    response = req_unicom_llm_chat_plus(messages, model_name = model, stop_words=["REASON"])
    logger.info(f"can_answer_question response:{response}")

    pattern = r',\s*"?REASON.*'
    result = re.sub(pattern, '}', response)
    
    result_json = extract_json(result)
    logger.info(f"can_answer_question  response: {result_json}")

    flag = result_json.get("TYPE",0)
    
    
    
  
    return flag


@advanced_timing_decorator()    
def remove_intent(query,history):
    history_context = ""
    if history:
        history_context = "\n".join(item['content'] for item in history)
    query = history_context + query 
    logger.info(f"remove_intent query :{query}")
    
    prompt = (
        """
        你是一个query改写助手，负责移除用户query中的意图部分，只保留要纯内容部分。

        ##Example 1:
        用户问题:
        将以下这句话转为语音：2025年亚洲冬季运动会的口号是"激情在这里燃烧，梦想在这里起航"。这个口号简洁而有力地表达了赛事的精神和愿景，旨在激发运动员的热情和斗志，同时也向全世界展示亚洲冰雪运动的魅力和发展潜力。通过这一口号，组委会希望传达出对体育精神的追求、对卓越成绩的渴望以及对团结合作的重视。
        重述问题:
        2025年亚洲冬季运动会的口号是"激情在这里燃烧，梦想在这里起航"。这个口号简洁而有力地表达了赛事的精神和愿景，旨在激发运动员的热情和斗志，同时也向全世界展示亚洲冰雪运动的魅力和发展潜力。通过这一口号，组委会希望传达出对体育精神的追求、对卓越成绩的渴望以及对团结合作的重视。

        ##Example 2:
        用户问题:
        合成一段语音，玄武湖景区是中国最大的皇家园林湖泊，被誉为"金陵明珠"。
        重述问题:
        玄武湖景区是中国最大的皇家园林湖泊，被誉为"金陵明珠"。

        ##Example 3:
        用户问题:
        九天海算政务大模型是由中国移动发布的九天人工智能行业大模型的组成部分，它旨在构建一个全面、开放、高效的智能决策支持系统，以向政府决策者提供精确的策略建议。将上面这句话合成为语音。
        重述问题:
        九天海算政务大模型是由中国移动发布的九天人工智能行业大模型的组成部分，它旨在构建一个全面、开放、高效的智能决策支持系统，以向政府决策者提供精确的策略建议。

        请根据以上样例，请补全当前问题缺失的上下文信息，输出重述问题。

        用户问题: """ + query
    )
    messages = [ {
                "role":"user",
                "content":prompt }
            ]
    response = req_unicom_llm_chat(messages, stream=False)
    result = response.content
    result = result.replace("重述问题:", "").replace("重述问题：", "")

    return result




@advanced_timing_decorator()    
def get_knowledge_prompt_with_llm(question, kn_userId, kn_name, threshold, topK=5, extend=0, extendedLength=400):
    # print("start get_knowledge_prompt_with_llm...")
    prompt = ''
    search_list = []
    context, kn_search_list = req_kb_search(question, kn_userId, kn_name, threshold, topK, extend, extendedLength)
    # logger.info(f'kb_context: {context}')
    logger.info(f'kn_search_list: {kn_search_list}')
    
 
    flag = False
    if len(kn_search_list) > 0:
        flag = is_used_knowledge_prompt(question, context)
    
    logger.info(f"is_kn_hit: %d",flag)
    
    
    if flag:
        
        prompt = KN_PROMPT_TEMPLATE.format(context=context, question=question)    
        search_list = kn_search_list

        # print(f"KN_PROMPT:{context}")
   
    
    # print("get_knowledge_prompt_with_llm end")

    return prompt, search_list

@advanced_timing_decorator()    
def pcm2wav(pcm_file, wav_file, channels=1, bits=16, sample_rate=22050):
    """将 PCM 文件转换为 WAV 文件。"""
    with open(pcm_file, 'rb') as pcmf:
        pcmdata = pcmf.read()
    if bits % 8 != 0:
        raise ValueError(f"位数（bits）必须是 8 的倍数，当前位数：{bits}")
    with wave.open(wav_file, 'wb') as wavfile:
        wavfile.setnchannels(channels)
        wavfile.setsampwidth(bits // 8)
        wavfile.setframerate(sample_rate)
        wavfile.writeframes(pcmdata)
    print(f"已完成 WAV 文件生成：{wav_file}")


@advanced_timing_decorator()    
def req_tts(text, name_save, speakerID='baker', audioFormat='pcm', sampleRate=22050, energy=1.0):
    """请求文本转语音服务。"""
    url = config["MODELS"]["tts_url"]
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }
    data = {
        "speakerID": speakerID,
        "text": text
    }
    sslopt = {
        "cert_reqs": ssl.CERT_NONE,
        "check_hostname": False
    }
    ws = websocket.create_connection(url, header=headers, sslopt=sslopt)
    ws.send(json.dumps(data))
    with open(name_save, 'wb') as fw:
        while True:
            res = ws.recv()
            if isinstance(res, str):
                res = json.loads(res)
                if res.get('finish') == 1:
                    break
            else:
                fw.write(res)
    ws.close()

def clean_up(*file_paths):
    """删除临时文件。"""
    for file_path in file_paths:
        if os.path.exists(file_path):
            os.remove(file_path)
            logger.info(f"已删除临时文件：{file_path}")


@advanced_timing_decorator()    
def upload_file_to_minio(file_path):
    """上传文件到 MinIO，并返回下载链接。"""
    
    url = config["MINIO"]["minio_upload_url"] if bucket_type == "MINIO" else config["OSS"]["oss_upload_url"]
    headers = {
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }
    try:
        with open(file_path, "rb") as f:
            response = requests.post(
                url,
                headers=headers,
                files={"file": f},
                verify=False  # It is recommended to configure SSL verification according to needs
            )
        if response.status_code == 200:
            data = response.json()
            download_link = data.get("download_link")
            return download_link
        else:
            # print(f"Upload failed, status code: {response.status_code}")
            logger.info(f"响应内容：{response.text}")
            return None
    except Exception as e:
        logger.info(f"发生错误：{e}")
        return None


@advanced_timing_decorator()    
def req_tts_minio(text, speakerID='baker'):
    """请求 TTS 服务并上传结果到 MinIO，返回下载链接。"""

    # Make sure the log directory exists
    temp_path = "./temp"
    if not os.path.exists(temp_path):
        os.makedirs(temp_path)
    temp_file_id = uuid.uuid4().hex
    file_path_pcm = f"{temp_path}/{temp_file_id}.pcm"
    file_path_wav = f"{temp_path}/{temp_file_id}.wav"
    req_tts(text, file_path_pcm, speakerID=speakerID)
    pcm2wav(file_path_pcm, file_path_wav)

    if bucket_type == "OSS":
        BUCKET_NAME = config["OSS"]["BUCKET_NAME"]
        oss_result = oss_manager.upload_local_file(file_path_wav,temp_file_id,BUCKET_NAME)
        download_link = oss_result["download_url"]
    else:
        download_link = upload_file_to_minio(file_path_wav)
    logger.info(f"req_tts_minio download_link : {download_link}")
    # Clean temporary files
    clean_up(file_path_pcm, file_path_wav)
    
    speech_url_list=[]    
    
    if download_link:
        logger.info(f"TTS语音文件下载链接为：{download_link}")
        speech_url_list.append({"text2spe_url":download_link})        
    else:
        logger.info("TTS语音文件上传失败。")
        
    return speech_url_list



@advanced_timing_decorator()    
def req_asr(wav_url):
   
    url = config["MODELS"]["asr_url"] 
    session_id = str(uuid.uuid4())

    asr_config = {"config":{"url":wav_url, "diarization": 1, "translate": 0, "session_id":session_id, "add_punc":1}}

    payload = json.dumps({"config": asr_config}, ensure_ascii=False)
    
    logger.info(f"req_asr asr_config：{asr_config}")
    headers = {
        "Content_type": "application/json;charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }


    response = requests.post(url, headers=headers, data=payload, verify=False)

    results = response.json().get("result")
    if not results:
        print("fail")
        print(response.json())
        return None

    full_result = ''


    for seg in results["diarization"]:
        if seg["text"].strip() == "": continue
        full_result += seg["text"] + "\n"

    return full_result.strip()


@advanced_timing_decorator()    
def req_chat_doc(query,file_url):  
    
    DOC_PROMPT_TEMPLATE = '''
        【任务描述】
        请根据用户输入的文档内容回答问题，并遵守回答要求。

        【文档内容】
        {context}

        【回答要求】
        - 你需要严格根据背景知识的内容回答，禁止根据常识和已知信息回答问题。
        - 答案要完整详细。
        - 对于不知道的信息，直接回答"未找到相关答案"
        -----------
        【用户问题】{question}
        '''
          # - If the user's question is "what is this", "document content", "what is the content of this file", etc., output the document content;
 
    if bucket_type == "OSS":
        P_ENDPOINT = config["OSS"]["P_ENDPOINT"]
        file_url = f"{P_ENDPOINT}" + file_url
    
    url = config["MODELS"]["default_doc_parser_url"]
    sentence_size = int (config["MODELS"]["DOC_CHUNK_SIZE"])
    overlap_size = float (config["MODELS"]["DOC_OVERLAP_RATIO"])
 
    access_token = token_manager.get_access_token()  # Get token
  
    payload =  json.dumps({
        "url": file_url,
        "sentence_size":sentence_size,
        "overlap_size":overlap_size,
        "separators":[
            "\n\n",
            "\n",
            " ",
            ",",
            "\u200b",  # zero width space
            "\uff0c",  # Full-width comma
            "\u3001",  # comma
            "\uff0e",  # full-width period
            "\u3002",  # period
            ".",
            "",
        ]
    })

    logger.info(f"req_chat_doc  payload:{payload}")
    
    headers = {
        "Content_type": "application/json;charset=utf-8",
        "Authorization": f"Bearer {token_manager.get_access_token()}"
    }


    response = requests.post(url, headers=headers, data=payload, verify=False)
    # result_dict = json.loads(response.text())
    docs = response.json().get("docs",[])
    doc_list=[]
    full_text = ""

    if docs:
        full_text =  docs[0].get("text")
   

    
    
    prompt = DOC_PROMPT_TEMPLATE.format(context=full_text, question=query)
    logger.info(f"req_chat_doc prompt:{prompt}")

    # redis_client.set_value(file_url,doc_list)
   
    return prompt


@advanced_timing_decorator()    
def req_chat_speech(query,file_url): 
    
    SPE_PROMPT_TEMPLATE = '''
        【任务描述】
        你是一个基于语音转写记录的智能问答助手，根据语音转写内容回答问题，并遵守回答要求。

        【语音转写内容】
        {context}
        
        【用户问题】{question}
        
        【回答要求】
        - 你需要严格根据语音转写内容回答，禁止根据常识和已知信息回答问题。
        - 如果用户问题是"语音转写"、"将这段语音转为文字"等语音转写意图的问题，直接输出语音转写内容；
        - 如果用户问题是"这是什么"、"描述一下这段语音"、"这段语音文件的内容是什么"等之类问题时，直接输出语音转写内容；
        - 直接输出语音转写内容时，需要你在不更改原始语种的前提下转写中的错别字做矫正，例如原始是英语，矫正之后的内容仍然是英语。
        -----------
        '''
   
    full_text = req_asr(file_url)   
    
    prompt = SPE_PROMPT_TEMPLATE.format(context=full_text, question=query)
   
    return prompt

####Classification of statistical problems and non-statistical problems
@advanced_timing_decorator(task_name="qa_stats_cls_by_llm")
def qa_stats_cls_by_llm(query,model= "unicom-70b-chat"):
    begin_time = datetime.datetime.now()
    prompt = (" 你是一个用户query分类器，你的任务是判定用户问题是统计类问题还是非统计类问题。以JSON格式返回，key包括type(0表示非统计类问题，1表示统计类问题)和reason(判定理由)。 以下给你提供了一些样例，以便让你做出更准确的判断。"+
        "\n\n"+"## Example：" +
        "\n用户问题:\n月份为2016-01出现了几次故障ATA？" +
        '\n判定结果:{"TYPE":1,"REASON":"该问题询问特定时间段内（2016年1月）的故障次数，涉及数据的数量或频率分析，因此属于统计类问题。"}\n' +
        "\n用户问题:\nAVIONICS VENT DEGRADE 的告警等级是？" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题询问的是关于AVIONICS VENT DEGRADE告警的具体信息，并不涉及数据的收集、分析或解释等统计活动。"}\n' +
        "\n用户问题:\n航空公司OTT-一二三航空，机号B-123A共出现了几次故障？" +
        '\n判定结果:{"TYPE":1,"REASON":"该问题询问的是特定航空公司、特定飞机的故障次数，这涉及到对数据的量化分析，因此属于统计类问题。"}\n' +
        "\n用户问题:\n月份为2016-07对应发生故障的机号有几个？" +
        '\n判定结果:{"TYPE":1,"REASON":"该问题涉及对特定时间段内（即2016年7月）发生的事件进行计数，符合统计类问题的特点。"}\n' +
        "\n用户问题:\n左配平空气活门交联的是IASC？" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题与具体的统计数据无关，而是询问关于左配平空气活门的连接对象"}\n' +
        "\n用户问题:\n数据集组2023年指标数值?" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题属于细节知识的问答，不应属于统计"}\n' +
        "\n用户问题:\n失速保护计算机怎么和IASC交联？" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题涉及的是技术实现或系统架构方面的内容，与统计数据的收集、分析无关。"}\n' +
        "\n用户问题:\n三级故障（严重），影响的业务范围是什么？" +
        '\n判定结果:{"TYPE":0,"REASON":"用户询问的是关于业务范围的问题，并没有涉及具体的统计数据或对数据的要求"}\n' +
        "\n用户问题:\n空气分配系统有什么功能？" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题询问的是空气分配系统的具体功能，不涉及数据的收集、分析或解释，因此不是统计类问题。"}\n' +   
        "\n用户问题:\nCMS信息为ECS:AVIONICS FAN2 LOW SPEED对应的AMM参考有几种？" +
        '\n判定结果:{"TYPE":1,"REASON":"该问题询问的是关于特定CMS信息（ECS:AVIONICS FAN2 LOW SPEED）的AMM参考数量,涉及数据的汇总、分析或推断等统计过程"}\n' +
        "\n用户问题:\n类型为技术出版物对应有多少种机型？" +
        '\n判定结果:{"TYPE":1,"REASON":"用户询问的是关于特定类型（技术出版物）对应的机型数量，这是一个需要统计数据来回答的问题。"}\n' +
        "\n用户问题:\n型号为ARJ21的飞机AWM参考ARJ21-SVV19-20003-00出现了多少次" +
        '\n判定结果:{"TYPE":1,"REASON":"该问题询问的是特定型号飞机（ARJ21）的某一具体参考编号（ARJ21-SVV19-20003-00）出现的次数，这需要对数据进行计数或统计分析来得出答案。"}\n' +
         "\n用户问题:\n故障等级为STATUS的故障发生了几次？" +
        '\n判定结果:{"TYPE":1,"REASON":"用户询问的是特定类型（STATUS）的故障发生的次数，这是一个明确的统计需求。"}\n' +
        "\n用户问题:\n电话响应时限5分钟的是哪一级别的故障" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题询问关于电话响应时限的具体规定，不涉及数据的收集、分析或解释，因此不是统计类问题。"}\n' +
        "\n用户问题:\n二级故障，在责任部门内，通报批评的范围是？" +
        '\n判定结果:{"TYPE":0,"REASON":"该问题询问的是关于二级故障在责任部门内的通报批评范围，这是一个具体的政策或规定查询，并不涉及数据的收集、分析或解释。因此，这不是一个统计类问题。"}\n' +
        "\n用户问题:\n调节空气分配系统主要功能？" +
        '\n判定结果:{"TYPE":0,"REASON":"用户询问的是关于调节空气分配系统的功能，这是一个描述性的问题，并不涉及数据的收集、分析或解释。因此，这不是一个统计类问题"}\n'  +     
        "\n\n请根据以上样例，对用户query进行分类。输出要求JSON格式" +              
        "\n\n当前用户query: " + query  +
        "\n\n输出要求:\n" +
        "\n## 分类标签不要混淆了，0Table示Non统计类问题，1Table示统计类问题。"    +
        "\n## 用JSONFormat回答，包括TYPE和REASON两个key。"      
             )
    
    messages = [ {"role":"user",
                "content":prompt }   ]
    response = req_unicom_llm_chat_plus(messages, model_name = model,stop_words=["REASON"])
    
    
    pattern = r',\s*"?REASON.*'         # Matches everything after , or "REASON
    result = re.sub(pattern, '}', response)
    result_json = extract_json(result)
    end_time = datetime.datetime.now()
    delay_time = end_time - begin_time
    logger.info(f"qa_cls: {delay_time}")
    return result_json

###chat_excel service

@advanced_timing_decorator()    
def req_chat_excel(data):
    url =  config["ACTION"]["default_chatexcel_server_url"]
    payload = data
    headers = {"Content-Type": "application/json"}
    try:
        response = requests.post(url, data=json.dumps(payload), headers=headers,stream=True)
        # logger.info(f"req_chat_excel response: {response.text}")
        return  response
    
    except requests.RequestException as e:
        logger.error("chatexcel_server API error: %s", e)  # Record error log
        return None
    
def req_chat_gui(data):
    url =  config["ACTION"]["default_gui_server_url"]
    payload = data
    headers = {"Content-Type": "application/json"}
    try:
        response = requests.post(url, data=json.dumps(payload), headers=headers,stream=True)
        # logger.info(f"req_chat_gui response: {response.text}")
        return  response
    
    except requests.RequestException as e:
        logger.error("req_chat_gui API error: %s", e)  # Record error log
        return None

@advanced_timing_decorator()    
def req_chat_docs_excel(data):
    url =  config["ACTION"]["default_docs_chatexcel_server_url"]
    payload = data
    headers = {"Content-Type": "application/json"}
    try:
        response = requests.post(url, data=json.dumps(payload), headers=headers,stream=True)
        logger.info(f"req_chat_excel response: {response.text}")
        return  response
    
    except requests.RequestException as e:
        logger.error("chatexcel_server API error: %s", e)  # Record error log
        return None
    


@advanced_timing_decorator()    
def intent_cls_by_llm(query,model="unicom-70b-chat"):  
    begin_time = datetime.datetime.now()
    
    prompt = (" 你本身是一个大语言模型，你的训练数据是有时限性的，在训练数据之后的新知识你是没有的。现在让你充当用户query分类器，你的任务是判定是根据你自有知识回答用户的问题（相对固定的通用常识）还是需要联网查搜索引擎回答用户问题（可能随时间动态变化的知识），主要的判断的依据是用户问题是否超过了你的训练知识，如果超过了，就要走联网搜索，以免给出错误或过期的的答案。以JSON格式返回，key包括type(0表示无需联网搜索的通用常识，1表示需要走联网搜索的相对动态知识)和reason(判定理由)。 以下给你提供了一些样例，以便让你做出更准确的判断。"+
        "\n\n"+"## Example：" +
        "\n用户问题:\n中国联通元景大模型的落地案例" +
        '\n判定结果:{"TYPE":1,"REASON":"中国联通元景大模型的落地案例是一个时时更新的情况。"}\n' +
        "\n用户问题:\n元景、星辰、九天，哪个大模型更好" +
        '\n判定结果:{"TYPE":1,"REASON":"元景、星辰、九天 三个大模型是一个新东西，而且是动态变化的。"}\n' +
        "\n用户问题:\n李时珍是谁" +
        '\n判定结果:{"TYPE":0,"REASON":"李时珍是中国古代的著名医生，是一个通用的常识知识，不会动态变化。"}\n' +
        "\n用户问题:\n李广聚是谁" +
        '\n判定结果:{"TYPE":1,"REASON":"李广聚不是一个著名的人物，需要联网搜索。"}\n' +
        "\n用户问题:\火星上是有生命的吗" +
        '\n判定结果:{"TYPE":0,"REASON":"火星上是否有生命存在虽然尚无定论，并且随着科技的进步和新的发现可能会发生变化，但是这个问题相当长时间内不会有新的突破，因此不需要联网搜索。"}\n' +
        "\n用户问题:\n北京的美食有哪些？" +
        '\n判定结果:{"TYPE":0,"REASON":"属于比较宽泛的信息咨询，而且短期内相对固定，不需要联网搜索。"}\n' +
        "\n用户问题:\甘肃省武威市有什么好玩的？" +
        '\n判定结果:{"TYPE":0,"REASON":"属于比较宽泛的信息咨询，而且短期内相对固定，因此不需要联网搜索。"}\n' +
        "\n用户问题:\介绍一下抖音" +
        '\n判定结果:{"TYPE":0,"REASON":"抖音是一个非常知名的社交媒体平台，用户咨询的是非常宽泛的信息，因此不需要联网搜索。"}\n' +   
        "\n用户问题:\n雍和宫附近哪里可以吃到羊肉泡馍？" +
        '\n判定结果:{"TYPE":1,"REASON":"属于衣食住行等方面很具体的生活服务类信息查询，用户的实际意图是要查询具体的可以吃到羊肉泡馍的饭馆，所以回答要精准，因此需要联网搜索。"}\n' +
        "\n用户问题:\n故宫附近有星巴克没？" +
        '\n判定结果:{"TYPE":1,"REASON":"属于衣食住行等方面很具体的生活服务类信息查询，用户的实际意图是要查询具体的可以吃到羊肉泡馍的饭馆，所以回答要精准，因此需要联网搜索。"}\n' +
        "\n用户问题:\n中国有多条河流" +
        '\n判定结果:{"TYPE":1,"REASON":"中国的河流数量可能会因自然环境的变化或人为因素而有所变动，因此需要通过联网搜索来获取最新的数据。"}\n' +
         "\n用户问题:\n九天、星辰、元景，哪个大模型比较好" +
        '\n判定结果:{"TYPE":1,"REASON":"九天、星辰、元景三个大模型之类的产品，发展很快，因此需要通过联网搜索来获取最新的数据。"}\n' +
        "\n用户问题:\n陶鹰鼎具体尺寸" +
        '\n判定结果:{"TYPE":0,"REASON":"陶鹰鼎是知名文物，属于通用常识，不需要联网搜索。"}\n' +
        "\n用户问题:\唐宋八大家诗歌创作风格" +
        '\n判定结果:{"TYPE":0,"REASON":"唐宋八大家是经典历史人物，诗词创作风格是通用常识，不需要联网搜索。"}\n' +
        "\n用户问题:\n北京香山红叶的最佳观赏期是什么时候" +
        '\n判定结果:{"TYPE":0,"REASON":"北京香山红叶的最佳观赏期虽然会受到每年气候的影响，可能存在变化，但是每年变化不大，相对比较固定和规律性的，因此可以认为属于通用常识，不需要联网搜索。"}\n' +
        "\n用户问题:\陕西省GDP最大的城市是哪里" +
        '\n判定结果:{"TYPE":1,"REASON":"这是查询实时性信息问题，你作为语言模型是不知道当前日期的，因此需要联网搜索。"}\n' +
        "\n用户问题:\中国联通人工智能大模型叫什么名字好？" +
        '\n判定结果:{"TYPE":0,"REASON":"这是一个文案生成问题，是你作为语言模型的强项，因此不需要联网搜索。"}\n' +               
        "\n用户问题:\n今天日期" +
        '\n判定结果:{"TYPE":1,"REASON":"这是查询实时性信息问题，你作为语言模型是不知道当前日期的，因此需要联网搜索。"}\n' +
        "\n用户问题:\n今天是什么节气？" +
        '\n判定结果:{"TYPE":1,"REASON":"这是一个关于特定作品的信息查询，虽然通常这类信息在作品发布后本身就会确定下来，但是这个作品可能是在你训练截至时间之后出现的，因此需要联网搜索。如果你确定自己知道这个作品的信息，可以判定不走联网搜索。"}\n' +
        "\n用户问题:\n电影阿甘正传的主演是谁？" +
        '\n判定结果:{"TYPE":0,"REASON":"电影阿甘正传，发布时间相对比较久了，可以十分确定是你已训练的知识，因此不走联网搜索。"}\n' +
        "\n用户问题:\n联通数科在安全方面主要有哪些资质？" +
        '\n判定结果:{"TYPE":1,"REASON":"查询某个企业的内部知识，不属于通用知识，因此需要联网搜索。"}\n' +
        "\n用户问题:\n理想汽车L9有几款颜色？" +
        '\n判定结果:{"TYPE":1,"REASON":"查询某个企业的公开产品知识，但是可能是在你训练截至时间之后出现的，因此需要联网搜索。如果你确定自己知道这个作品的信息，可以判定不走联网搜索。"}\n' +
        "\n用户问题:\n解释股足永弃的含义" +
        '\n判定结果:{"TYPE":1,"REASON":"“股足永弃”并不是一个常见的成语或词汇，实际上是网友对国足持续败绩的谐音，属于动态变化的知识，因此需要联网搜索。"}\n' +
        "\n用户问题:\n解释凿壁偷光的意思" +
        '\n判定结果:{"TYPE":0,"REASON":"“凿壁偷光”是一个常见的成语或词汇，属于通用知识，因此不需要联网搜索。"}\n' +
        "\n用户问题:\n北京前三的大学" +
        '\n判定结果:{"TYPE":1,"REASON":"北京的大学排名由综合能力决定，会经常性的变化，因此需要联网搜索。"}\n' +
        "\n用户问题:\n中国联通“1+N+X”智算能力体系是什么" +
        '\n判定结果:{"TYPE":1,"REASON":"查询某个企业的公开产品知识，但是可能是在你训练截至时间之后出现的，因此需要联网搜索。如果你确定自己知道这个作品的信息，可以判定不走联网搜索。"}\n' +
        "\n用户问题:\n在西安，每个车一周限号几天" +
        '\n判定结果:{"TYPE":1,"REASON":"查询某个企业的公开产品知识，但是可能是在你训练截至时间之后出现的，因此需要联网搜索。如果你确定自己知道这个作品的信息，可以判定不走联网搜索。"}\n' +
        "\n用户问题:\n请帮我查找并搜索【影入平羌江水流出自哪里?】" +
        '\n判定结果:{"TYPE":1,"REASON":"鉴于用户问题明确指示需要进行搜索操作，因此需优先选择联网搜索的方式来寻找所需的信息"}\n' +
        "\n用户问题:\n请帮我查找并搜素“两水夹明镜，双桥落彩虹”这句诗出自哪里?" +
        '\n判定结果:{"TYPE":1,"REASON":"鉴于用户问题明确指示需要进行搜索操作，因此需优先选择联网搜索的方式来寻找所需的信息"}\n' +
        "\n用户问题:\n中国联通“1+N+X”智算能力体系是什么" +
        '\n判定结果:{"TYPE":1,"REASON":"查询某个企业的公开产品知识，但是可能是在你训练截至时间之后出现的，因此需要联网搜索。如果你确定自己知道这个作品的信息，可以判定不走联网搜索。"}\n' +     
        "\n\n请根据以上样例，对用户query进行分类。输出要求JSON格式" +              
        "\n\n当前用户query: " + query  +
        "\n\n输出要求:\n" +
        "\n## 分类标签不要混淆了，0Table示不Need联网Search，1Table示Need联网Search。"    +
        # "\n## Reason description should be brief." +
        "\n## 用JSONFormat回答，包括TYPE和REASON两个key。"      
             )
   
    messages = [ {"role":"user",
                "content":prompt }   ]
    # response = req_unicom_llm_chat_plus(messages, ["111"])
    # print(response)
    response = req_unicom_llm_chat_plus(messages, model_name = model,stop_words=["REASON"])
    # print(response)
    
    pattern = r',\s*"?REASON.*'         # Matches everything after , or "REASON
    result = re.sub(pattern, '}', response)
    result_json = extract_json(result)
    
    result_json = extract_json(result)    
        
    end_time = datetime.datetime.now()
    delay_time = end_time - begin_time
    logger.info(f"intent_cls_by_LLM_delay: {delay_time}")
     
    # merged_dict = {"query":query, **result_json}
    

    return result_json

@advanced_timing_decorator()
def intent_cls_by_memory(query,history_messages,model="unicom-70b-chat"):
    begin_time = datetime.datetime.now()
    
    memory_context = ""
    for item in history_messages:
        memory_context += item["role"]+" : "+item["content"]+"\n"
    
        
    

    MEM_USED_PROMPT_TEMPLATE = '''根据下面对话历史，判断是否可以回答用户本轮问题
    
    【参考信息】：
    ```
     {context} 
    ```   

     【用户本轮问题】
     ```
     {question}
     ```
     
     输出要求：
     ## If you cannot get an answer from it at all, please answer "0" and explain the reason. If you can get an answer from it, please answer "1".
     ## Please answer in json format, the keys are TYPE and REASON. '''

    prompt = MEM_USED_PROMPT_TEMPLATE.format(context=memory_context, question=query)
    # logger.info(f"can_answer_question  prompt: {prompt}")
    
    messages = []
    subjson = {}
    subjson["role"] = "user"
    subjson["content"] = prompt
    messages.append(subjson)
    
    response = req_unicom_llm_chat_plus(messages, model_name = model, stop_words=["REASON"])
    
    pattern = r',\s*"?REASON.*'        # Matches everything after , or "REASON
    result = re.sub(pattern, '}', response)
    result_json = extract_json(result)
    
    
    flag = result_json.get("TYPE",0)
    # begin_time = datetime.datetime.now()
    end_time = datetime.datetime.now()
    delay_time = end_time - begin_time
    logger.info(f"intent_cls_by_MEM_delay: {delay_time}")
  
    return flag

@advanced_timing_decorator()    
def answer_by_memory(query,history_messages,stream):
    
    memory_context = ""
    for item in history_messages:
        memory_context += item["role"]+" : "+item["content"]+"\n"
    
        
    

    MEM_USED_PROMPT_TEMPLATE = '''根据下面对话历史，回答用户本轮问题
    
    【对话历史】：
    ```
     {context} 
    ```   

     【用户本轮问题】
     ```
     {question}
     ```
     
     输出要求：
     ## Do not include the words "based on conversation history" in the answer.
  '''

    prompt = MEM_USED_PROMPT_TEMPLATE.format(context=memory_context, question=query,stream=stream)
    # logger.info(f"can_answer_question  prompt: {prompt}")
    
    messages = []
    subjson = {}
    subjson["role"] = "user"
    subjson["content"] = prompt
    messages.append(subjson)
    messages_truncate =req_trim_messages(messages)
    
    response = req_unicom_llm_chat(messages_truncate)
  
    return response
    
    
if __name__ == "__main__":
    pass
    query = "法沦功"
    result = is_query_sen(query)
    print(result.json())
    print(result.json()['data']['details'][0]['containsSensitiveWords'])
    
    # print(get_intent_cls(query))
    # messages = [{"role":"user","content":query}]
    # print(intent_cls_by_llm(query))
    # print(req_unicom_llm_chat2(messages,["reason"]))

    # print(get_intent_cls("What is this"))
    # url = "https://maas-gz.ai-yuanjing.com/minio/download/api/public/abcc1e25-751f-49b7-9955-b8a3cbf57e78-2fe925ba-4341-438f-912e-8d3d3c692b9f@civillaw.pdf"
    # query = "Which department is directly responsible for the company's IT system data security work?"
    # req_chat_doc(query,url)
    # print(redis_client.get_value(url))
    
    # query = "Help me synthesize the intended speech. The content is: There are many Sichuan delicacies, including spicy hot pot, Kung Pao chicken, Mapo tofu, dandan noodles, twice-cooked pork, couple's lung slices, etc., each one is mouth-watering!"
    # print(remove_intent(query))
    
    # audio_file = "https://maas-gz.ai-yuanjing.com/minio/download/api/public/tmpuahrygol.wav"
    # print(req_asr(audio_file))
    
#     file_path = "dyEnglish video.MP4"
    
#     print(upload_file_to_minio(file_path))
              
    # query = "Who is the Poet of the Tang Dynasty?"
    # result = intent_cls_by_llm(query)
    # print(result)
              
    
    
    


#     download_link = req_tts_minio(texts[1])
#     if download_link:
#         print(f"The file has been uploaded, the download link is: {download_link}")
#     else:
#         print("File upload failed.")
  
    #-----vqa-----
    # pic_url = "https://maas-gz.ai-yuanjing.com/minio/download/api/public/tmpoy4ovbqi.jpg"
    # prompt = "What is this"

    # response = unicom_vqa_stream(pic_url,prompt,[])
    # for chunk in response:
    #     print(chunk)

    #-----vqa-----


    #-----rag-------
    # prompt,search_list= req_kb_search_v2("2025 Asian Winter Games slogan", "b5d139b5-901b-4ae0-8fd5-8c57df77476e", "AsianWinterGame", threshold=0.0, topK=5)
    # print(prompt)
    # print(search_list)
    # code,msg = is_kb_belong_to_user_v2(["susu","asan"],"731a5fee-0ab7-4431-b0d3-f6807fba5s999")
    # print(code)
    # print(msg)

    # ----langchain LLM ----
    # query = "What are her masterpieces"
    # query = 'When is Zhao Liying's birthday? '
    
    # messages =[
    #     {"role": "user", "content": "Who are you"},
    #     {"role": "assistant", "content": "I am Yuan Jing, an artificial intelligence chatbot. I was developed by China Unicom and specially designed to help people answer questions, provide information and have conversations. If you have any questions or need help, you can always ask me."},
    #     {"role": "user", "content": "Who is Zhao Liying"},
    #     {"role": "assistant", "content": "Zhao Liying is a famous Chinese actress. She was born in Langfang City, Hebei Province on October 16, 1987. She graduated from Langfang Electronic Information Engineering School.\n\nShe entered the entertainment industry in 2006 when she won the championship of Feng Xiaogang's group in the Yahoo Star Search Competition, and subsequently appeared in many well-known TV series and movies, such as Qing'er in "The My Fair Princess", "Lu Zhen" "Legend" and other roles, these works have made her gradually popular and accumulated a large number of fans. \n\nZhao Liying has won the love of the audience with her superb acting skills and sweet image, and has become one of the high-profile stars in the Chinese entertainment industry. In addition to her acting career, she also actively participates in charity activities, demonstrating her sense of social responsibility and public welfare spirit.
    #     # {"role":"user","content":query},
    # ]

    # messages = [{'role': 'user', 'content': 'Who are you'}]

    
    # result  = intent_cls_by_memory(query,messages)
    # print(result)
    # model_name="deepseek-r1"

    # result = req_unicom_llm_chat(messages,stream=True,model_name=model_name)
    # if result:
    #     for chunk in result:
    #         print(chunk)
    # print(result)



    #------code------

    # query = "Generate a histogram with the vertical axis data being 100, 200, 300 and the horizontal axis being 2022, 2023, 2024?"
    # query = "Programming to generate a histogram, the vertical axis data are 100, 200, 300, the horizontal axis is 2022, 2023, 2024, save to a new file"
    # response = is_query_sen(query)
    # print(response.text)

# #     file_name = "interpreter_test.txt"
    # upload_file_url = "https://maa÷s-gz.ai-yuanjing.com/minio/download/api/public/tmp69yhsw4s.jpg",
    
   
    
    # query = "Write a program to convert pictures to black and white."
    # query = "Write a bubble sort presentation"
    # datajson = {}
    # response = req_code_interpreter(query,need_file=False,upload_file_url=upload_file_url)
    # for line in response.iter_lines(decode_unicode=True):
    #     if  line.startswith("data:") and not line.startswith('data:""'):
    #         line = line[5:]
    #         print(line)
    #         datajson = json.loads(line)
    #         incremental_content = datajson["data"]["choices"][0]["message"]["content"]
            # print(incremental_content,end="",flush=True)
            # if datajson["data"]["choices"][0]["finish_reason"]:
            #     print(datajson["gen_file_url_list"][0]["output_file_url"])
    # print(datajson)
    


    
    # query = 'Is this how much it charges? '
    # can_answer_by_messge(query,messages)

    # response = req_unicom_llm_chat(messages,stream=True)

#     # print(response)
      
    # for line in response.iter_lines(decode_unicode=True):
    #     # print(line)
    #     if line.startswith("data:") and not line.startswith('data:""'):
    #         line = line[5:]
    #         # print(line)
    #         datajson = json.loads(line)
    #         if not datajson['data']['choices'][0]['finish_reason']:
    #             incremental_content = datajson["data"]["choices"][0]["message"]["content"]
    #             print(incremental_content,flush=True,end="")

    # model_name = config["MODELS"]["unicom-72b-chat-ali"].strip()
    # print(model_name)
    # print(model_name == "unicom-72b-chat-ali")
    # model_name = config["MODELS"]["unicom_72b_chat_ali"]
    # print(repr(model_name))
    # print(repr(model_name.strip()))
    # print(model_name.strip() == "unicom-72b-chat-ali")

    # model_name = config["MODELS"]["unicom_72b_chat_ali"]
    # print(repr(model_name))
    # print(model_name == "unicom-72b-chat-ali")
    
   

    # print(url)
    # print(model_name)
#     history= [ 
#         {
#             "query": "What are the famous tourist attractions in Nanjing?",
#             "rewrite_query": "What are the famous tourist attractions in Nanjing?",
#             "upload_file_url": "",
#             "qa_type": 2,
#             "response": "Nanjing has many famous tourist attractions. Here are some places that are popular among tourists:\n\n1. **Zhongshan Scenic Area**: This is a concentrated expression of Nanjing's urban characteristics and a famous attraction in Jiangsu and Nanjing. It is famous for its unique geographical location and rich natural landscape, and it also has a profound cultural heritage. Zhongshan Scenic Area includes many core scenic spots such as Sun Yat-sen Mausoleum Scenic Area, Ming Xiaoling Scenic Area and Linggu Scenic Area, which are home to more than 200 historical sites and commemorative buildings. \n\n2. **Confucius Temple-Qinhuai Scenic Belt**: This is a tourist attraction integrating natural scenery, landscape gardens, temples, schools, residents and local customs. With the ancient buildings of Confucius Temple as the center and the Qinhuai River within ten miles as the axis, it shows the historical and cultural accumulation of Nanjing for more than two thousand years. Here, tourists can not only visit historical buildings, but also taste authentic Qinhuai snacks. **Niushou Mountain Cultural Tourism Area**: As a scenic spot in Jinling, Niushou Mountain is a cultural Buddhist resort with thousands of years of history. The Foding Palace in the scenic area is majestic and majestic, and is one of the most important attractions in Niushou Mountain. In addition, the surrounding natural scenery is also very beautiful, and you can overlook the charming scenery when you climb to the top of the mountain \n\n4. **Ming Xiaoling Mausoleum**: This is the joint burial mausoleum of Ming Taizu Zhu Yuanzhang and Queen Ma. It is known as the "First Royal Mausoleum of the Ming and Qing Dynasties". Although it has experienced more than 600 years of vicissitudes, it still retains the original layout of the mausoleum, allowing visitors to deeply experience the magnificent landscape of the ancient royal mausoleum and the history and culture of the Ming Dynasty. \n\n5. **Nanjing Presidential Palace**: This building complex covers a vast area and is one of the largest and most well-preserved architectural complexes in modern China. It has witnessed many important events and activities in modern Chinese history. \n\n6. **Qixia Mountain Scenic Area**: This place is known as the "No. 1 Mingxiu Mountain in Jinling" and was listed as one of the "Forty-Eight Scenic Spots in Jinling" during the Ming Dynasty. During Qianlong's six visits to Jiangnan, the main attractions include Mingjing Lake, Qixia Temple, Thousand Buddha Rock, Relic Pagoda and Biyun Pavilion. \n\n8. **Nanjing Museum**: This is one of the three major museums in China. It covers a vast area and adopts the "one courtyard and six museums" layout. It has a large number of precious cultural relics, second only to the Palace Museum. It is an important place to understand Nanjing and even Chinese culture. \n\nThese attractions have their own characteristics, and they can leave a deep impression on tourists, whether it is natural scenery or human history. ",
#             "gen_file_url_list": [

#             ]
#         },
#     ]
#     query = 'If I want to go to this place, what are the local airports and train stations? '

#     result = req_query_rewrite(query,history)
#     print(result)

    # pic_list,_ = req_txt2img_plus("Generate a histogram, the vertical axis data are 100, 200, 300, and the horizontal axis is 2022, 2023, 2024?", 4)
    # print(pic_list)


    # _, search_list =search_knowledge_base("Who is the chairman of China Unicom", "731a5fee-0ab7-4431-b0d3-f6807fba5ae5",["Unicom_KB_APP"], threshold=0.4, topK=5, extend=0, extendedLength=400)
    # print(search_list)

  



    # Usage example:
    # resource_id = "20240515-18.46.38-05e3db06-a7cb-4add-8ad9-7bdf6d1f29e9.mp4"
    # video_base64 = get_video_base64(resource_id)
    # print("SSS:\n",video_base64,"\nHHH")
    # prompt = "Draw a video, with the wind. Swinging willow tree"
    # result = generate_and_upload_video(prompt)
    # print(result)
    # query = 'Corgi'
    # result = req_txt2img_plus(query)
    # print(result)
    
    # req_unicom_llm_chat("")
    # pic_url = "https://39.101.74.2:7776/minio/download/api/public/tmpanin39zs.png"
    # pic_url = "https://39.101.74.2:7776/minio/download/api/public/tmpsx7w09dz.png"
    # result = generate_and_upload_video_img2vid(pic_url,"Generate a video")
    # print(result)

    
#     The prompt word for query = """ construction parameters is: the selected tool is action_getWeatherNow, which requires the following parameters:
# Parameter location: The location of the query, which can be a city name, zip code, etc.
# Please generate the required parameters based on the following user questions. Please output the parameters in json format: How is the weather in Beijing today"""

    # print("Messages：", messages)
    # query  = "who are you"
    # query  = "hi"
    # messages = [{"role":"user","content":query}]
    
   
     # response = req_unicom_llm_chat(messages,stream=False)

#     print(response.headers["Content-Type"])
    # print(response.text)
    
    # print(get_intent_cls("Tomorrow's weather"))
    # try:
    #     response_data = response.json()
    #     print("\n【unicom-34b-chat】: ", response_data['data']['choices'][0]['message']['content'])
    # except requests.exceptions.JSONDecodeError as e:
    #     print(f"JSONDecodeError: {e.msg}")
    #     print("Response text was:", response.text)
    # print(response)
    # print("\n【qwen-14b-chat】: " , response.json()['data']['choices'][0]['message']['content'])
    # response = req_unicom_llm_chat(messages,stream=False,model_name="unicom-13b-chat")
    # print("\n【unicom-13b-chat】: ", response.json()['data']['choices'][0]['message']['content'])
    # response = req_unicom_llm_chat(messages,stream=False,model_name="unicom-34b-chat")
    # print(response.text)
    # print("\n【unicom-34b-chat】: ", response.json()['data']['choices'][0]['message']['content'])
   
        
        
