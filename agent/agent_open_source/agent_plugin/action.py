
import json
import logging  
from utils.model_tools import req_unicom_llm_chat  #Import third-party tool models
from utils.actions.agents import Assistant,ReActChat
from utils.actions.tools.base import BaseTool, register_tool
from utils.actions.llm import get_chat_model
from utils.actions.tools.openapi_plugin import openapi_schema_convert,add_openapi_plugin_to_additional_tool
from utils.generator import *
import yaml
import urllib3
import os
import configparser
import time

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
logger = logging.getLogger(__name__)


config = configparser.ConfigParser()
config.read('config.ini')

APP_ID = config["ACTION"]['APP_ID']
API_KEY = config["ACTION"]['API_KEY']
SECRET_KEY = config["ACTION"]['SECRET_KEY']




def action_infer(query,plugin_list,function_calls_list,action_type = "qwen_agent",history = []):
    # Get query from request
    llm = get_chat_model({
        # Use the model service provided by DashScope:
        'model': 'unicom-72b-chat-ali-v2',
        'model_server': 'http://1654416620085582.cn-wulanchabu.pai-eas.aliyuncs.com/api/predict/llm_70b_2/v1/chat/completions',
        'api_key': API_KEY,

        # Use the OpenAI-compatible model service provided by DashScope:
        # 'model': 'qwen1.5-14b-chat',
        # 'model_server': 'https://dashscope.aliyuncs.com/compatible-mode/v1',
        # 'api_key': os.getenv('DASHSCOPE_API_KEY'),

        # Use the model service provided by Together.AI:
        # 'model': 'Qwen/Qwen1.5-14B-Chat',
        # 'model_server': 'https://api.together.xyz',  # api_base
        # 'api_key': os.getenv('TOGETHER_API_KEY'),

        # Use your own model service compatible with OpenAI API:
        # 'model': 'Qwen/Qwen1.5-72B-Chat',
        # 'model_server': 'http://localhost:8000/v1',  # api_base
        # 'api_key': 'EMPTY',
    })

    try:
        print(f"query is {query}")
        messages_input = []
        if history:
            messages_input.extend(history)        
        messages_input.append({'role': 'user', 'content': query})
        name = "unicomllm"

        print("\n---------识别action插件并进行配置-------------")
        function_list = []
        
        print(f"plugin_list:{len(plugin_list)}")
        for i in range(0,len(plugin_list)):
            # print(f"The current plugin_list is: {plugin_list[i]}")
            api_schema = plugin_list[i]['api_schema']
            # print(f"api_schema:{api_schema}")
            if 'api_auth' in plugin_list[i]:
                api_auth = plugin_list[i]['api_auth']
            else:
                api_auth = {
                'type': None,  
                'in': 'header',  
                'name': 'Authorization', 
                'value': None
            }
  
            api_cfg = openapi_schema_convert(api_schema,api_auth)
            if api_cfg:
                plugin_cfg = api_cfg
                # Register function
                fn_list = add_openapi_plugin_to_additional_tool(plugin_cfg, [])
                function_list.extend(fn_list)
        
        print(f"---完成action插件并进行配置，function_list为：{function_list}---")
        
        if action_type == "modelscope_agent":
            system_message = "你是一个API接口查询助手,你需要根据用户的问题选择最优的API接口,并返回API参数"
        elif action_type == "yanjing_function_call":
            system_message = "你是一个function查询助手,你需要根据用户的问题返回最优的function及入参"
        else:
            system_message = "你是一个任务规划助手，你需要基于已有的工具对用户问题进行拆分成多个任务task，并对每个任务进行分步推理,请注意，后一步的推理务必用到前一步的结果，最后，请逐步输出结果"
        
        print("-------------启动llm_actions函数------------")
        llm_actions = ReActChat(function_list = function_list,llm=llm,system_message = system_message,name = name,function_calls_list=function_calls_list,action_type = action_type)
        
        ans = ""
        response = llm_actions.run(messages_input)
        
        for line in response:
            # print(f"line:{line}")
            line = line[0]['content']
            line = json.loads(line)
        
        result = {"code":0,"message":"success","result":line}
        jsonarr = json.dumps(result, ensure_ascii=False)
        return jsonarr
        # return response

    except Exception as e:
        error_msg = str(e)
        print("params error: %s", error_msg)
        arr = {}
        arr['code'] = 1
        arr['message'] = error_msg
        arr['result'] = ''
        jsonarr = json.dumps(arr, ensure_ascii=False)
        return jsonarr

def list_all_files(directory):
    """
    递归地列出指定目录及其子目录下的所有文件
    """
    all_files = []
    for root, dirs, files in os.walk(directory):
        for file in files:
            all_files.append(os.path.join(root, file))
    return all_files


###Batch configuration action
def actions_config(act_filepaths):
    configurations = []
    for i in range(0,len(act_filepaths)):
        if ".yaml" in act_filepaths[i] or ".yml" in act_filepaths[i]:
            with open(act_filepaths[i], 'r', encoding='utf-8') as yaml_file:
                api_schema = yaml.safe_load(yaml_file)
                ##Save as json string format
                api_schema = json.dumps(api_schema)
                ##json parsed into dictionary format
                api_schema = json.loads(api_schema)
                #print(api_schema)
                #result = LLMActions(api_schema, api_auth)
                configurations.append(api_schema)

        if ".json" in act_filepaths[i]:
            with open(act_filepaths[i], 'r', encoding='utf-8') as file:
                api_schema = json.load(file)
                configurations.append(api_schema)
    return configurations

if __name__ == '__main__':
    tool_list = list_all_files('action_files')
    tools = actions_config(tool_list)
    func_list = list_all_files('function_files')
    func_tools = actions_config(func_list)
    function_names = [func["function_name"] for func in func_tools]
    # query = "What bookstores are near the Lama Temple in Beijing and what is the weather there today?"
    # query = "How is the weather in Beijing today?"
    # query = "Coffee shops near Qingyouyuan in Beijing"
    # query = "Ningxia University Admissions Major"
    # query = "Bookstores near Ningxia University"
    # query = "Please recommend a hot pot restaurant in Xicheng District, Beijing"
    # query = "Please recommend stinky tofu shops in Yuhua District, Changsha and where they are"
    # query = "The location of the Summer Palace in Beijing"
    query = "帮我查询2024年5月份的销量"
    # query = "A plane ticket on August 15th"
    # query = "Help me generate a mind map about the history of artificial intelligence development"
    
    # action_types = ['qwen_agent','modelscope_agent','yanjing_function_call']
    # action_types = ['yuanjing_function_call']
    action_types = ['qwen_agent']
    # tools = [{"api_schema": {"info": {"description": "Recommend nearby food based on the specified location POI.", "title": "recommendation_v2_1 API", "version": "1.0.0"}, "openapi": "3.0.0", "paths": {"/run_for_bigmodel/07ee0365-3aa3-42c1-8ee3-f906e691d584/recommendation_v2_1": {"post": {"description": "Recommend nearby food based on the specified location POI., ", "operationId": "action_recommendation_v2_1", "parameters": [{"in": "header", "name": "content-type", "required": True, "schema": {"example": "application/json", "type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"properties": {"food": {"description": "Name of food, such as roast duck, cafe, salted duck, etc.", "type": "string"}, "poi": {"description": "Place name, such as China, Guangzhou City, Haidian District, Heping Street, etc.", "type": "string"}}, "required": ["poi", "food"], "type": "object"}}}}, "responses": {"200": {"content": {"application/json": {"schema": {"type": "object"}}}, "description": "Successfully obtained query results"}, "default": {"content": {"application/json": {"schema": {"type": "object"}}}, "description": "Error message when the request fails"}}, "summary": "recommendation"}}}, "servers": [{"url": "https://maas.ai-yuanjing.com/plugin/api"}]}}]
    
    
    '''
    tools = [
        {
          "api_schema": {
            "openapi": "3.0.0",
            "info": {
              "title": "心知天气API",
              "version": "1.0.0",
              "description": "提供当前天气信息的API，包括温度、天气状况等。"
            },
            "servers": [
              {
                "url": "https://api.seniverse.com/v3"
              }
            ],
            "paths": {
              "/weather/now.json": {
                "get": {
                  "summary": "天气查询工具",
                  "operationId": "getWeatherNow",
                  "description": "根据地点获取当前的天气情况，包括温度和天气状况描述。",
                  "parameters": [
                    {
                      "name": "location",
                      "description": "查询的地点，可以是城市名、邮编等。",
                      "in": "query",
                      "required": True,
                      "schema": {
                        "type": "string"
                      }
                    }
                  ],
                  "responses": {
                    "200": {
                      "description": "成功获取天气信息",
                      "content": {
                        "application/json": {
                          "schema": {
                            "type": "object",
                            "properties": {
                              "location": {
                                "type": "string"
                              },
                              "text": {
                                "type": "string"
                              },
                              "code": {
                                "type": "string"
                              },
                              "temperature": {
                                "type": "string"
                              }
                            }
                          }
                        }
                      }
                    },
                    "default": {
                      "description": "请求失败时的错误信息",
                      "content": {
                        "application/json": {
                          "schema": {
                            "type": "object",
                            "properties": {
                              "error": {
                                "type": "string"
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "api_auth": {
            "type": "apiKey",
            "in": "query",
            "name": "key",
            "value": "S2ph5B1WETiu7imD6"
          }
        },
        {
          "api_schema": {
            "openapi": "3.1.0",
            "info": {
              "title": "高德地图",
              "description": "获取 POI 的相关信息",
              "version": "v1.0.0"
            },
            "servers": [
              {
                "url": "https://restapi.amap.com/v5/place/"
              }
            ],
            "paths": {
              "/text": {
                "get": {
                  "summary": "action_get_location_coordinate",
                  "description": "根据POI名称，获得POI的经纬度坐标",
                  "operationId": "get_location_coordinate",
                  "parameters": [
                    {
                      "name": "keywords",
                      "in": "query",
                      "description": "POI名称，必须是中文",
                      "required": True,
                      "schema": {
                        "type": "string"
                      }
                    },
                    {
                      "name": "region",
                      "in": "query",
                      "description": "POI所在的区域名，必须是中文",
                      "required": False,
                      "schema": {
                        "type": "string"
                      }
                    }
                  ],
                  "deprecated": False
                }
              },
              "/around": {
                "get": {
                  "summary": "action_search_nearby_pois",
                  "description": "搜索给定坐标附近的POI",
                  "operationId": "search_nearby_pois",
                  "parameters": [
                    {
                      "name": "keywords",
                      "in": "query",
                      "description": "目标POI的关键字",
                      "required": True,
                      "schema": {
                        "type": "string"
                      }
                    },
                    {
                      "name": "location",
                      "in": "query",
                      "description": "中心点的经度和纬度，用逗号分隔",
                      "required": False,
                      "schema": {
                        "type": "string"
                      }
                    }
                  ],
                  "deprecated": False
                }
              }
            },
            "components": {
              "schemas": {}
            }
          },
          "api_auth": {
            "type": "apiKey",
            "in": "query",
            "name": "key",
            "value": "77b5f0d102c848d443b791fd469b732d"
          }
        }
      ]
    func_tools = [
		{
    "function_name": "play_music",
    "description": "播放指定歌手的指定歌曲",
    "parameters": {
        "type": "object",
        "properties": {
            "singer": {
                "type": "string",
                "description": "歌手名字，例如周杰伦、孙燕姿、毛不易"
            },
            "song": {
                "type": "string",
                "description": "歌曲名字，例如双节棍、"
            }
        },
        "required": [
            "singer"
        ]
    }
}
	]
    '''
    
    
    times_count = []
    for action_type in action_types:
        start_time = time.time()
        results = action_infer(query,tools,func_tools,action_type = action_type)
       
        print(f"results:{results}")
        end_time = time.time()
        use_time = end_time - start_time
        times_count.append([action_type,use_time])
        
    print(f"action耗时：{times_count}")
    

    # func_name = results['func_names']
    # func_params = results['func_params']
    # prompt = results['prompt']
    # thought_inference = results['thought_inference']
    # qa_type = results['qa_type']
   
    
    # content = ""
    # for line in results:
    #     if isinstance(line,bytes):
    #         line=line.decode()
    #     if isinstance(line,str):
    #         content += line
    #     if isinstance(line[0]['content'],str):
    #         content += line[0]['content']
    #     else :
    #         line = line[0]['content'][0]['text']
    #         if line.startswith("data:"):
    #             line = line[5:]
    #             datajson = json.loads(line)
    #             # print(datajson)
    #             incremental_content = datajson["data"]["choices"][0]["message"]["content"]
    #             content += incremental_content
    # print(f"The model output result is:\n{content}")



