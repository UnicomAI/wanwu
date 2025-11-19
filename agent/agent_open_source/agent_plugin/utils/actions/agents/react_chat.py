import json
import logging  
import re
import time
from typing import Dict, Iterator, List, Literal, Optional, Tuple, Union
from utils.actions.agents.fncall_agent import FnCallAgent
from utils.actions.llm import BaseChatModel
from utils.actions.llm.schema import ASSISTANT, DEFAULT_SYSTEM_MESSAGE, Message,FUNCTION
from utils.actions.tools.openapi_plugin import OpenAPIPluginTool
from utils.actions.settings import MAX_LLM_CALL_PER_RUN
from utils.actions.tools import BaseTool
from utils.actions.utils.utils import format_as_text_message, merge_generate_cfgs
from utils.model_tools import *


logger = logging.getLogger(__name__) 
TOOL_DESC = (
    '{name_for_model}: Call this tool to interact with the {name_for_human} API. '
    'What is the {name_for_human} API useful for? {description_for_model} Parameters: {parameters} {args_format}')

'''



PROMPT_REACT = """尽可能回答以下问题。您可以使用以下工具：
{tool_descs}

请使用以下格式回答：

Question: 需要你回答的问题
Thought: 你应该思考该做什么
Action: 选择要调用的工具名称，必须是 [{tool_names}] 中的一个，请直接输出工具名称，不要加句号
Action Input: 调用这个工具所需的全部参数，如果无法识别，请不要随意编造。格式应为标准 JSON。
Observation: 动作的结果
... （此 Thought/Action/Action Input/Observation 可以重复零次或多次）
Thought: 输出最终答案
Final Answer: 用 Markdown 格式回答原始问题。

开始！

Question: {query}

必须严格按照上述格式输出答案。
请在最终输出时用简体中文作答和分析，禁止输出繁体中文。"""


'''



PROMPT_REACT = """Answer the following questions as best you can. You have access to the following tools:
{tool_descs}

Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, Must be one of [{tool_names}],  If the action is not recognized or not highly relevant to the question, return None, return None.
Action Input: the all parameters required for the chosen action, If it cannot be recognized, do not improvise. formatted in standard JSON Observation.
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can be repeated zero or more times)
Thought: I now know the final answer
Final Answer: the final answer to the all original input question. Must use standard  Markdown formatting in the final answer,Please use <br> as the line break instead of \n

Begin!

Question: {query}

Must output strictly in the above format to answer the Question.
Please answer and analyze in Simplified Chinese during the final output, and prohibit outputting Traditional Chinese
"""


TOOL_DESC_MODELSCOPE = (
    '{name_for_model}: Call this tool to interact with the {name_for_human} API. '
    'What is the {name_for_human} API useful for? {description_for_model} ')

PROMPT_MODELSCOPE = """ 

You are an API interface query assistant; you need to select the optimal API interface based on Question: {query} and return the API parameters.
You have access to the following functions: {tool_descs}.

based on the above information,please must be  Use the following format:

Action: please return the appropriate toolname from [{tool_names}] to answer the question  or  Return None if no tool applies directly. Do not include any additional information.

"""

ANSWER_PROMPT = "Based on the above information, think about what to do next. "

class ReActChat(FnCallAgent):
    """This agent use ReAct format to call tools"""
    def __init__(self,
                 function_list: Optional[List[Union[str, Dict, BaseTool,OpenAPIPluginTool]]] = None,
                 function_calls_list=None,
                 llm: Optional[Union[Dict, BaseChatModel]] = None,
                 system_message: Optional[str] = DEFAULT_SYSTEM_MESSAGE,
                 name: Optional[str] = None,
                 description: Optional[str] = None,
                 files: Optional[List[str]] = None,
                 action_type = "action_agent",
                 model_name = "",
                 model_url = "",
                 **kwargs):
        super().__init__(
                         function_list=function_list,
                         function_calls_list=function_calls_list,
                         llm=llm,
                         system_message=system_message,
                         name=name,
                         description=description,
                         files=files,
                         action_type = action_type,
                         model_name = model_name,
                         model_url = model_url,
                         **kwargs)
        self.extra_generate_cfg = merge_generate_cfgs(
            base_generate_cfg=self.extra_generate_cfg,
            new_generate_cfg={'stop': ['Observation:', 'Observation:\n','Observation: ','Observation: \n']},
        )

    def _run(self, messages: List[Message],lang: Literal['en', 'zh'] = 'en', **kwargs) -> Iterator[List[Message]]:
        func_names = []
        func_params = []
        action_code = 0
        qa_type = 20
        thought_inference = ""
        param_des = []  ###保存不同调用模型对应的参数描述，便于识别哪些参数缺失，从而调用大模型进行对话跟踪 [EN] ##Save the parameter descriptions corresponding to different calling models to facilitate identifying which parameters are missing, so as to call the large model for dialogue tracking
        function_items =  []
        functions_total_lists = []
        param_pattern = r"参数 (\w+):"
        keys_identify = []
        Invalid_character = ["无","{}","","NONE","[]","未知"]
        count_qs = 0
        first_index = 0
        
        try:
            ####1、yuanjing_function_call模型输出格式与其他架构不一致，需单独处理------直接调用 [EN] ###1. The output format of the yuanjing_function_call model is inconsistent with other architectures and needs to be processed separately------directly called
            logger.info(f"self.action_type:{self.action_type}")
            # logger.info(f"self.system_message:{self.system_message}")
            query = messages[-1]['content']
            logger.info(f"ReActChat的query:{query}")
            total_result = ""

            if self.action_type == "function_call":
                function_calls_names = [item['function']['name'] for item in self.function_calls_list]
                qa_type = 21
                tools = self.function_calls_list
                messages_prompt = [msg.model_dump() for msg in messages]
                system_messages = [{'role': 'system', 'content': self.system_message}]
                system_messages.extend(messages_prompt)

                logger.info(f"调用req_unicom_llm_chat_function_call的prompt为{system_messages}")
                if tools:
                    time1 = time.time()
                    response = req_unicom_llm_chat_function_call(messages=system_messages, tools = tools,stream=False)
                    time2 = time.time()
                    time_use = time2 - time1
                    logger.info(f"调用req_unicom_llm_chat_function_call共耗时{time_use}")
                    if response.status_code == 200 :
                        res = response.json()
                        logger.info(f"res:{res}")
                        if res["data"]["choices"][0]["finish_reason"] == "tool_calls":
                            action = res["data"]["choices"][0]["message"]["tool_calls"][0]["function"]["name"]
                            if action.upper() != "NONE":
                                thought = res["data"]["choices"][0]["message"]["content"]
                                action_input = res["data"]["choices"][0]["message"]["tool_calls"][0]["function"]["arguments"]
                                action_input = json.loads(action_input)
                                ###获取详细的参数描述 [EN] ##Get detailed parameter description
                                index_function = function_calls_names.index(action)
                                param_descriptions = "\n".join([f"参数 {key}: {value.get('description', '无描述')}" for key,value in self.function_calls_list[index_function].get('function', {}).get('parameters', {}).get('properties', {}).items() if key in self.function_calls_list[index_function].get('function', {}).get('required')])
                                matches = re.findall(param_pattern, param_descriptions)
                                param_des.append(param_descriptions)
                                logger.info(f"yuanjing_function_call：【{action}】的必填参数描述为：{param_des}")
                                ###判断function_call参数识别是否缺失 [EN] ##Determine whether function_call parameter recognition is missing
                                keys_identify = [key for key in action_input.keys()]
                                
                                if "<tool>" not in total_result and action:
                                    total_result += "<tool>"
                                    if t == 1:
                                        yield [Message(role=ASSISTANT, content=f"\n--start--\n")]
                                    yield [Message(role=ASSISTANT, content=f"<tool>工具名：{action}")]
                                    yield [Message(role=ASSISTANT, content=f"\n\n```请求参数：\n{action_input}\n```\n\n")]
                                    
                                elif "<tool>" in total_result and action:
                                    if t == 1:
                                        yield [Message(role=ASSISTANT, content=f"\n--start--\n")]
                                    yield [Message(role=ASSISTANT, content=f"工具名：{action}")]
                                    
                                    yield [Message(role=ASSISTANT, content=f"\n\n```请求参数：\n{action_input}\n```\n\n")]
                                    
                                    
                                if action_input and (not any(v is None  or  (v and v.strip() == "") or (not v) for v in action_input.values())) and  all(element in keys_identify for element in matches):
                                    func_names.append(action)
                                    func_params.append(action_input)
                                    action_result = thought + f"\nAction: {action}\nAction Input: {action_input}\nObservation:\n"
                                    thought_inference += action_result
                                    action_code = 0
                                    action_output = f"已基于{[action]}完成参数提取"    
                                    yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                else:
                                    prompt = f'''

                                    你是一个问题提炼助手，你的主要任务是通过以下信息，简洁、准确地告诉用户缺失哪些参数以及需要输入的参数要求。

                                    已知完整的参数信息为：{param_des}，

                                    你提取出来的参数为：{[action_input]}

                                    输出要求

                                    ## 答案中不要出现"根据您提供的信息"、"根据提供的信息"、"根据参考信息"等之类的话术。 [EN] # Do not use words such as "based on the information you provided", "based on the information provided", "based on reference information", etc. in the answer.
                                    ## 请不要使用markdown格式。 [EN] # Please do not use markdown format.
                                    ## 回答请简洁、准确地告诉用户缺失哪些参数以及需要输入的参数要求 [EN] # Please answer concisely and accurately tell the user which parameters are missing and the parameter requirements that need to be entered.
                                    '''   
                                    ###调用大模型进行回答 [EN] ##Call the large model to answer
                                    messages = []
                                    subjson = {}
                                    subjson["role"] = "user"
                                    subjson["content"] = prompt
                                    messages.append(subjson)
                                    response = req_unicom_llm_chat_plus(messages = messages,model_name = self.model_name)
                                    action_code = 2
                                    action_output = response
                                    yield [Message(role=ASSISTANT, content=f"</tool>\n")]

                            else:
                                action_code = 1
                                action_output = "function识别为NONE"

                        else:
                            ###判断是否是对问题的进一步追问 [EN] ##Determine whether it is a further inquiry into the question
                            content =  res["data"]["choices"][0]["message"]["content"]
                            time_add1  = time.time()
                            is_add = is_answer_additional_remarks(query,content)
                            time_add2  = time.time()
                            time_add = time_add2 - time_add1
                            logger.info(f"is_answer_additional_remarks的响应时间:{time_add}")
                            logger.info(f"判断是否是对问题的进一步追问、补充:{is_add}")
                            if int(is_add['flag']) :
                                action_code = 0
                                action_output = content
                                yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                            else:
                                action_code = 1
                                action_output = "未识别出合适的function"
                    else:
                        action_code = 1
                        action_output = f"服务请求失败:{response.text}"
                else:
                    action_code = 1
                    action_output = "function缺失，请补充"  
                result = {"action_code": action_code,"description":"function缺失，请补充","func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":action_output}
                result = json.dumps(result, ensure_ascii=False)
                logger.info(f"action最终的输出结果为：{result}")
                yield [Message(role=ASSISTANT, content=result)]

            else:
                text_messages = self._prepend_react_prompt(messages, lang=lang)
                # logger.info(f"text_messages：{text_messages}")
                ###定义插件名称 [EN] ##Define plugin name
                function_calls_names = [item['function_name'] for item in self.function_calls_list]
                function_items.extend(function_calls_names)
                function_items.extend(self.function_list)
                logger.info(f"function_items:{function_items}")

                ###定义插件详细列表 [EN] ##Define detailed list of plugins
                functions = [f for f in self.function_map.values()]
                functions_total_lists.extend(functions)
                functions_total_lists.extend(self.function_calls_list)

                ###定义模型循环次数 [EN] ##Define model cycle times
                self.extra_generate_cfg['lang'] = lang
                num_llm_calls_available = MAX_LLM_CALL_PER_RUN  
                
                
                while num_llm_calls_available > 0:
                    num_llm_calls_available -= 1
                    t = MAX_LLM_CALL_PER_RUN - num_llm_calls_available
                    logger.info(f"第{t}次start _call_llm的messages为：\n{text_messages}")
                    time_output_stream1 = time.time()
                    output_stream = self._call_llm(messages=text_messages,functions=[func.function for func in self.function_map.values()],function_calls_list=self.function_calls_list,model_name = self.model_name,model_url = self.model_url,extra_generate_cfg=self.extra_generate_cfg)
                    time_output_stream2 = time.time()
                    time_output_stream = time_output_stream2 - time_output_stream1
                    logger.info(f"第{t}次调用_call_llm的响应时间为：{time_output_stream}")
                    # Accumulate the current response
                    response = ""
                    output = []
                    time_output1 = time.time()
                    for output in output_stream:
                        if output and output[-1].content:
                            res = output[-1].content[0]['text']
                            if isinstance(res,bytes):
                                res=res.decode()
                            response += res
                            if "<TOOL>" in response.upper():
                                # logger.info(f"-----deepseek截断提取：stop_token效果------") [EN] logger.info(f"-----deepseek truncation extraction: stop_token effect------")
                                response_new = re.sub('<TOOL>.*?</TOOL>', '', response, flags=re.DOTALL)
                                if "</TOOL>" in response and ("ACTION INPUT" in response_new.upper() and "}" in response_new.upper()):
                                    break
                            # ####如果直接是答案行，则无需校验，直接输出结果 [EN] ####If it is directly the answer line, no verification is needed and the result is output directly.
                            # elif "Final Answer" in response:
                            #     if first_index == 0 and "Final Answer:" in response:
                            #         final_answer_pos = response.upper().index("Final Answer:")
                            #         res = response[final_answer_pos:]
                            #     elif first_index == 0 and "Final Answer" in response:
                            #         final_answer_pos = response.upper().index("Final Answer")
                            #         res = response[final_answer_pos:]
                            #     result = {"action_code": 0, "description": "", "func_names": func_names,
                            #               "func_params": func_params, "thought_inference": thought_inference,
                            #               "qa_type": qa_type, "action_output": f"{res}"}
                            #     result = json.dumps(result, ensure_ascii=False)
                            #     yield [Message(role=ASSISTANT, content=res)]
                            #     # logger.info(f"action最终的输出结果为：{result}") [EN] # logger.info(f"The final output result of action is: {result}")
                            #     first_index += 1

                            else:
                                # logger.info(f"-----非 deepseek截断提取：stop_token效果------") [EN] logger.info(f"-----Non-deepseek truncation extraction: stop_token effect------")
                                if "OBSERVATION" in response.upper() or ("ACTION INPUT" in response.upper() and "}" in response.upper()):
                                    break
                        
                    time_output2 = time.time()
                    time_output = time_output2 - time_output1
                    logger.info(f"第{t}次提取response的响应时间为:{time_output}")

                    ####大模型报错，直接退出 [EN] ###Large model reports an error, exit directly
                    if not response:
                        result = {"action_code": 1,"description":f"调用大模型报错：{output_stream.text}","func_names":[],"func_params":[],"thought_inference":"","qa_type":None,"action_output":f"调用大模型报错：{output_stream.text}"}
                        result = json.dumps(result, ensure_ascii=False)
                        yield [Message(role=ASSISTANT, content=result)]
                        break

                    # elif "FINAL ANSWER" in response.upper():
                    #     break

                    else:
                        response = re.sub('<think>.*?</think>', '', response, flags=re.DOTALL)
                        response = re.sub('<THINK>.*?</THINK>', '', response, flags=re.DOTALL)

                    logger.info(f"第{t}次调用的response:\n{response}\n")
                    ###剔除思考过程后提取工具调用结果 [EN] ##Extract tool call results after eliminating the thinking process
                    response = re.sub('<tool>.*?</tool>', '', response, flags=re.DOTALL)
                    response = re.sub('<think>.*?</think>', '', response, flags=re.DOTALL | re.IGNORECASE)
                    ##结构化提取大模型识别的参数：兼容元景模型及ds模型（输出内容不一致） [EN] #Structured extraction of parameters for large model recognition: compatible with Yuanjing model and ds model (output content is inconsistent)
                    counts = {
                        "OBSERVATION": response.upper().count("OBSERVATION") + response.upper().count("OBSERVATION:"),
                        "ACTION": response.upper().count("ACTION") + response.upper().count("ACTION:")
                    }
                    num = counts["OBSERVATION"] if counts["OBSERVATION"] else counts["ACTION"]
                    logger.info(f"需要调用call的次数为：{num}")
                    #logger.info(f"处理了的response:{response}") [EN] logger.info(f"Processed response:{response}")
                    action_results = find_tools(response,num) 
                    action = action_results['func_name']
                    action_input = action_results['func_params']
                    thought = action_results['final_result']
                    
                    
                    #######判断提取的参数个数是否完整及有效 [EN] ###### Determine whether the number of extracted parameters is complete and valid
                    ###1、识别提取出的参数 [EN] ##1. Identify the extracted parameters
                    if action and (action_input and all(str(v).strip() not in Invalid_character for v in action_input)):
                        keys_identify = [key for key in action_input[-1].keys()]
                    logger.info(f"keys_identify：{keys_identify}")

                    ####2、识别原始API的参数名称及参数描述 [EN] ###2. Identify the parameter names and parameter descriptions of the original API
                    if action and all(item in function_items for item in action_results['func_name']):
                        if self.action_type == "modelscope_agent":   
                            function_to_call = parse_tool_selection_response(action_results['func_name'][0],functions_total_lists)
                        else:
                            function_to_call = action[-1]
                            
                        #####插件的索引位置及提取 [EN] #### Index location and extraction of plug-in
                        if  function_to_call and function_to_call not  in function_calls_names :
                            selected_function = next((f for f in functions if f.name == function_to_call), None)
                            param_descriptions = "\n".join([f"参数 {param['name']}: {param.get('description', '无描述')}" for param in selected_function.parameters if "value" not in param and param.get('required')==True])
                        ####function_to_call的索引位置及提取 [EN] ###Index position and extraction of function_to_call
                        else:
                            index_function = function_calls_names.index(function_to_call)
                            param_descriptions = "\n".join([f"参数 {key}: {value.get('description', '无描述')}" for key,value in self.function_calls_list[index_function].get('parameters', {}).get('properties', {}).items() if key in self.function_calls_list[index_function].get('required')])
                        param_des.append(param_descriptions)
                        logger.info(f"【{function_to_call}】的必填参数描述为：{param_des}")
                        matches = re.findall(param_pattern, param_descriptions)
                        logger.info(f"对应的必填参数名称为：{matches}")  
                        
                    #####3、比较原始及识别出的action及action_input的差异：兼容元景与ds返回格式 [EN] ####3. Compare the difference between original and recognized action and action_input: compatible with Yuanjing and ds return formats  
                    ####1) 有效调用环节：输出包含Action且参数不为空时，才需进一步比较差异，否则直接走输出环节 [EN] ###1) Effective calling link: Only when the output contains Action and the parameters are not empty, further differences need to be compared, otherwise the output link will be used directly.
                    if ("Action:" in response or "Action" in response ) and action_results['func_name']:
                        if not all(item in function_items for item in action_results['func_name']):
                            action_results['action_code'] = 1
                            action_results['description'] = "action名称识别有误"
                        elif (not matches) and action_results['action_code']== 2:
                            action_results['action_code'] = 0
                            action_results['description'] = "API无必填参数"
                        elif not all(element in keys_identify for element in matches):
                            action_results['action_code'] = 2
                            action_results['description'] = "参数识别缺失"
                        elif action_results['action_code'] == 2 and all(key in matches and value not in Invalid_character for key,value in action_input[-1].items()):
                            action_results['action_code'] = 0
                            action_results['description'] = "必填参数提取成功"
                        elif  self.action_type == "action_agent" and action_results['action_code'] == 2 :
                            ####qwen_agent 允许react多反思一步 [EN] ###qwen_agent allows react to reflect one more step
                            count_qs += 1
                            if count_qs< 3:
                                observation = f"\nObservation: 【{action[-1]}】的必填参数描述为：{param_des},请思考下一步的action，请注意在识别参数时需按照用户提供的信息来确定参数，不要随意发挥\n"
                                action_result = thought + f"\nAction: {action[-1]}\nAction Input: {action_input[-1]}" + observation
                                thought_inference += action_result
                                text_messages[-1].content += action_result
                                continue
                    ####输出环节：共计三种情形 [EN] ###Output link: three situations in total
                    ###1）大模型自身就能回答 [EN] ##1) The large model can answer by itself
                    ###2）调用插件后的结果能回答用户问题 [EN] ##2) The result after calling the plug-in can answer user questions
                    ###3）调用插件后的结果不能回答用户问题，但是是对用户问题的进一步补充（用户问题不清晰时生效） [EN] ##3) The result after calling the plug-in cannot answer the user's question, but is a further supplement to the user's question (effective when the user's question is unclear)
                    ###4）无法回答时，code返回1 [EN] ##4) When unable to answer, code returns 1
                    else:
                        if t == 1:
                            action_results['action_code'] = 1
                            action_results['description'] = "大模型识别为自行回答，无需调用插件"
                            func_params.append({"input":query})
                            result = {"action_code": 1,"description":action_results['description'],"func_names":["yuanjing"],"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":f""}
                            result = json.dumps(result, ensure_ascii=False)
                            logger.info(f"action最终的输出结果为：{result}")
                            yield [Message(role=ASSISTANT, content=result)]
                            break 
                        time_flag1 = time.time()
                        can_answer_text = response    
                        logger.info(f"can_answer_text：{can_answer_text}") 
                        flag = can_answer_question(query,can_answer_text,model_name = self.model_name,model_url = self.model_url)
                        time_flag2 = time.time()
                        time_flag = time_flag2 - time_flag1
                        logger.info(f"can_answer_question的响应时间为：{time_flag}")
                        final_result = can_answer_text
                        
                        if "Final Answer" in can_answer_text or "Final Answer:" in can_answer_text:
                            ###剔除think的思考过程 [EN] ##Eliminate the thinking process of think
                            can_answer_text = re.sub('<tool>.*?</tool>', '', can_answer_text, flags=re.DOTALL)
                            final_result = can_answer_text.split("Final Answer:", 1)[1].strip() if "Final Answer:" in can_answer_text else can_answer_text.split("Final Answer", 1)[1].strip()
                            
                        if int(flag) and ("Final Answer" in can_answer_text or "Final Answer:" in can_answer_text):
                            logger.info(f"-----action调用完成，正在输出结果-------") 
                            action_results['action_code'] = 0
                            action_results['description'] = "action调用成功"
                            action_results['final_result'] = action_results['final_result']if action_results['final_result'] else final_result
                            
                        else:
                            ###判断是否是对问题的进一步追问 [EN] ##Determine whether it is a further inquiry into the question
                            time_add1  = time.time()
                            # is_add = is_answer_additional_remarks(query,action_results['final_result'],model_name = self.model_name,model_url = self.model_url)
                            is_add = is_answer_additional_remarks(query, final_result,model_name=self.model_name, model_url=self.model_url)
                            time_add2  = time.time()
                            time_add = time_add2 - time_add1
                            logger.info(f"is_answer_additional_remarks的响应时间:{time_add}")
                            logger.info(f"判断是否是对问题的进一步追问、补充:{is_add}")
                            if int(is_add['flag']) and "ACTION" not in can_answer_text.upper():
                                if t > 1:
                                    yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                result = {"action_code": 0,"description":"","func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":f"{final_result}"}
                                result = json.dumps(result, ensure_ascii=False)
                                logger.info(f"action最终的输出结果为：{result}")
                                yield [Message(role=ASSISTANT, content=result)]
                                break  
                            else:
                                action_results['action_code'] = 1
                                action_results['description'] = "action答案无法全面准确的回答问题" if not action_results['description'] else action_results['description']
                    logger.info(f"action_results:{action_results}") 

                    ####完成参数提取，启动函数调用 [EN] ###Complete parameter extraction and start function call   
                    ###1、整合action及入参 【特别地，modelscope_agent的流程为：ACTION提取、ACTION_INPUT提取、call调用三步；非modelscope_agent流程为两步：ACTION+ACTION_INPUT提取、call调用】 [EN] ##1. Integrate action and input parameters [Specifically, the process of modelscope_agent is: ACTION extraction, ACTION_INPUT extraction, and call invocation in three steps; the non-modelscope_agent process is in two steps: ACTION+ACTION_INPUT extraction, call invocation]
                    if action and action_results['action_code'] in [0,2]:
                        func_names.extend(action)
                        func_params.extend(action_input)
                    logger.info(f"func_names:{func_names}，func_params：{func_params}")
                    
                    ###2、完成ACTION提取，开始流式输出 [EN] ##2. Complete ACTION extraction and start streaming output.
                    if "<tool>" not in total_result and action:
                        total_result += "<tool>"  
                        if t == 1:
                            yield [Message(role=ASSISTANT, content=f"\n--start--\n")]
                        yield [Message(role=ASSISTANT, content=f"<tool>工具名：{action[0]}")]
                    elif "<tool>" in total_result and action:
                        if t == 1:
                            yield [Message(role=ASSISTANT, content=f"\n--start--\n")]
                        yield [Message(role=ASSISTANT, content=f"工具名：{action[0]}")]
                    
                    
                    ###3、modelscope_agent需特殊处理，提取ACTION_INPUT [EN] ##3. modelscope_agent requires special processing to extract ACTION_INPUT
                    if self.action_type == "modelscope_agent" and func_names:
                        if t<=1: 
                            param_prompt = f'''

                            You have selected the tool {function_to_call}, which requires the following parameters:\n {param_descriptions}. 
                            Please generate the required parameters based on the user’s question.

                            please Use the following format:

                            Question: the input question you must answer:{query} ::::
                            Action: the action to take 【{function_to_call}】
                            Action Input: the parameters required for the chosen action 【{function_to_call}】, formatted in standard JSON Observation .
                            Observation: the result of the action 
                            Final Answer: the final answer to the original input question
                            '''
                            text_messages[-1].content = param_prompt
                            ###移除识别出来的action及参数 [EN] ##Remove the identified actions and parameters
                            func_names.pop()
                            # func_params.pop()
                            continue
                    
                    ###4、完成ACTION_INPUT提取，开始流式输出 [EN] ##4. Complete ACTION_INPUT extraction and start streaming output
                    if action and action_input:
                        yield [Message(role=ASSISTANT, content=f"\n\n```请求参数：\n{action_input[0]}\n```\n\n")]
                        

                    ###5、识别ACTION的类型，如为function_call，则标识为qa_type为21，如为action，则标识为qa_type为20， [EN] ##5. Identify the type of ACTION. If it is function_call, it will be identified as qa_type is 21. If it is action, it will be identified as qa_type is 20.
                    ###function_call无需用走call调用，直接返回ACTION和ACTION_INPUT [EN] ##function_call does not need to be called, it directly returns ACTION and ACTION_INPUT
                    if action_results['func_name'] and all(param in function_calls_names for param in action_results['func_name']) and action_results['action_code'] == 0:
                        qa_type = 21
                        action_result = action_results['final_result'] + f"\nAction: {action_results['func_name']}\nAction Input: {action_results['func_params']}\nObservation:\n"
                        thought_inference += action_result
                        text_messages[-1].content += action_result
                    logger.info(f"当前识别出的qa_type:{qa_type}")
                    
                    
                    ###6、继续调用函数或者输出结果 [EN] ##6. Continue to call functions or output results.
                    ####1）not action时，表名输出结果 [EN] ###1) When not action, the table name output result
                    ###2）action时，表明继续执行_call_tool [EN] ##2) Action indicates that _call_tool will continue to be executed.
                    if action_results['action_code'] == 0:
                        logger.info(f"当前识别出的action:{action}  action_code :{action_results['action_code']}")
                        if not action:
                            logger.info(f"------输出最终结果------")
                            ###输出action结束符号 [EN] ##Output action end symbol
                            yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                            text_messages[-1].content += "\n请保证语义通顺"
                            start_pattern = "Use the following format"
                            end_pattern = "Begin!"
                            pattern = re.compile(re.escape(start_pattern) + r'.*?' + re.escape(end_pattern), re.DOTALL)
                            text_messages[-1].content = re.sub(pattern, "", text_messages[-1].content)
                                
                            action_output = action_results['final_result']  if qa_type == 20 else  f"已基于{action}完成参数提取"
                            result = {"action_code": 0,"description":"","func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":action_output}
                            result = json.dumps(result, ensure_ascii=False)
                            logger.info(f"action最终的输出结果为：{result}")
                            yield [Message(role=ASSISTANT, content=result)]
                            break   
                        else:
                            if qa_type == 20:
                                logger.info(f"第{t}次调用_call_tool:开始执行_call_tool")
                                action = action[-1]
                                action_input = action_input[-1] if action_input else {}   ###存在插件无需必填参数的情况 [EN] ##There are cases where the plug-in does not require required parameters.
                                is_use_api_output = "TRUE" in str(action_input.get("is_use_api_output", "false")).upper()
                                api_stream = action_input.get("stream",False)
                                logger.info(f"_call_tool的入参为：action:{action}，action_input:{action_input}\n")
                                
                                time_call_tool1 = time.time()
                                observation = self._call_tool(action, action_input, messages=messages, **kwargs)
                                logger.info(f"observation类型：{type(observation)}")
                                logger.info(f"_call_tool 的observation：{observation}")
                                
                                time_call_tool2 = time.time()
                                time_call_tool = time_call_tool2 - time_call_tool1
                                logger.info(f"第{t}次调用_call_tool的响应时间为:{time_call_tool}")
                                logger.info(f"第{t}次调用_call_tool的stream为:{api_stream}")

                                if is_use_api_output and api_stream:
                                    yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                    action_output = observation
                                    action_code = 0 if observation else 1
                                    logger.info(f"action_output type:{type(action_output)}  observation type:{type(observation)}")
                                    try:
                                        ####正常流式输出 [EN] ###Normal streaming output
                                        logger.info(f"----流式API返回成功----")
                                        for line in observation.iter_lines(decode_unicode=True):
                                        # for line in observation:
                                            # logger.info(f"--line type--:{type(line)}")
                                            # logger.info(f"--line--:{line}")
                                            yield [Message(role=ASSISTANT, content=line)]
                                        break

                                    except Exception as e:
                                        ####API内部服务异常 [EN] ###API internal service exception
                                        logger.info(f"----流式API返回异常----")
                                        observation = json.loads(observation)
                                        plugin_result = observation["description"]
                                        for line in plugin_result:
                                            yield [Message(role=ASSISTANT, content=line)]
                                        break

                                elif is_use_api_output and (not api_stream):
                                    observation = json.loads(observation)
                                    plugin_result = observation["description"]
                                    # plugin_result = re.sub('<think>.*?</think>', '', plugin_result, flags=re.DOTALL)
                                    # plugin_result = re.sub('<THINK>.*?</THINK>', '', plugin_result, flags=re.DOTALL)
                                    plugin_result = re.sub(r'<think>.*?</think>', '', plugin_result,flags=re.DOTALL | re.IGNORECASE)
                                    plugin_result_output = json.loads(plugin_result)
                                    yield [Message(role=ASSISTANT,content=f"\n\n```工具调用结果：\n{plugin_result_output}\n```\n\n")]
                                    # yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                    try:
                                        output_result = json.loads(plugin_result)
                                        action_output = output_result['data']['choices'][0]['message']['content']  if output_result.get("data",{}).get("choices",{}) else (output_result.values()[0] if len(output_result.keys()) == 1 else observation['description'])
                                    except Exception as e:
                                        action_output = plugin_result
                                    try:
                                        plugin_result = json.loads(action_output)
                                        plugin_flag = all(value is None for value in plugin_result.values())
                                        if plugin_flag:
                                            action_code = 1
                                            description = f"插件调用结果无效：{action_output}"
                                        else:
                                            action_code = 0
                                            description = action_output
                                    except Exception as e:
                                        action_code = 1
                                        description = action_output
                                    func_params = [item for item in func_params if "current_screenshot" not in item.keys()]
                                    
                                else:
                                    observation = json.loads(observation)
                                    plugin_result = observation["description"]
                                    logger.info(f"plugin_result:{plugin_result}   type:{type(plugin_result)}")
                                    plugin_result_output = re.sub(r'<think>.*?</think>', '', plugin_result,flags=re.DOTALL | re.IGNORECASE)
                                    yield [Message(role=ASSISTANT, content=f"\n\n```工具调用结果：\n{plugin_result_output}\n```\n\n")]
                                    
                                    if observation['action_code'] == 1:
                                        action_code = 1
                                        action_output = f"{observation['description']}"
                                    elif observation['action_code'] == 0 and all(value is None for value in json.loads(plugin_result).values()):
                                        action_code = 1
                                        action_output = f"插件调用结果无效：{observation['description']}"

                                logger.info(f"第{t}次调用_call_tool的action_code为:{action_code}")
                                if is_use_api_output or ( (not is_use_api_output) and action_code == 1):
                                    result = {"action_code": action_code,"description":observation['description'],"func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":action_output}
                                    result = json.dumps(result, ensure_ascii=False)
                                    logger.info(f"action最终的输出结果为：{result}")
                                    yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                    yield [Message(role=ASSISTANT, content=result)]
                                    # for line in observation.iter_lines(decode_unicode=True):
                                    #     yield [Message(role=ASSISTANT, content=line)]
                                    break
                                else:
                                    observation_result = f"\nObservation: {observation}\n"
                                    if (not text_messages[-1].content.endswith('\nThought: ')) and (not thought.startswith('\n')):
                                        text_messages[-1].content += '\n'
                                    if str(action_input).startswith('```'):
                                        # Add a newline for proper markdown rendering of code
                                        action_input = '\n' + action_input
                                    action_result = thought + f"\nAction: {action}\nAction Input: {action_input}" + observation_result + ANSWER_PROMPT
                                    thought_inference += action_result
                                    text_messages[-1].content += action_result
                                    continue
                                
                            else:
                                yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                                result = {"action_code": 0,"description":"","func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":f"已基于{action}完成参数提取"}
                                result = json.dumps(result, ensure_ascii=False)
                                logger.info(f"action最终的输出结果为：{result}")
                                yield [Message(role=ASSISTANT, content=result)]
                                break

                    elif action_results['action_code'] == 2:
                        yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                        prompt = f'''

                                    你是一个问题追问补充助手，你的主要任务是通过以下信息，简洁、准确地告诉用户缺失哪些参数以及需要输入的参数要求。

                                    已知完整的参数信息为：{param_des}，

                                    你提取出来的参数为：{func_params}

                                    输出要求

                                    ## 答案中不要出现"根据您提供的信息"、"根据提供的信息"、"根据参考信息"等之类的话术。 [EN] # Do not use words such as "based on the information you provided", "based on the information provided", "based on reference information", etc. in the answer.
                                    ## 请不要使用markdown格式。 [EN] # Please do not use markdown format.
                                    ## 回答请简洁、准确地告诉用户缺失哪些参数以及需要输入的参数要求 [EN] # Please answer concisely and accurately tell the user which parameters are missing and the parameter requirements that need to be entered.
                                    '''  
                        ###调用大模型进行回答 [EN] ##Call the large model to answer
                        messages = []
                        subjson = {}
                        subjson["role"] = "user"
                        subjson["content"] = prompt
                        messages.append(subjson)
                        response = req_unicom_llm_chat_plus(messages = messages,model_name = self.model_name)
                        logger.info(f"参数缺失场景下对应的prompt为：{prompt}")
                        action_code = 2
                        action_output = response
                    else:
                        # yield [Message(role=ASSISTANT, content=f"</tool>\n")]
                        logger.info(f"------未能准确识别，返回状态码1----")
                        action_output = f"{action_results['description']}"
                        action_code = 1
                        
                    result = {"action_code": action_code,"description":action_results['description'],"func_names":func_names,"func_params":func_params,"thought_inference":thought_inference,"qa_type":qa_type,"action_output":action_output}
                    result = json.dumps(result, ensure_ascii=False)
                    logger.info(f"action最终的输出结果为：{result}")
                    yield [Message(role=ASSISTANT, content=result)]
                    break 
                    
        except Exception as e:
            # yield [Message(role=ASSISTANT, content=f"</tool>\n")]
            logger.info(f"------调用服务报错，返回状态码1:{str(e)}----")
            result = {"action_code": 1,"description":f"调用服务报错，返回状态码1:{str(e)}","func_names":[],"func_params":[],"thought_inference":thought_inference,"qa_type":qa_type,"action_output":f"调用服务报错，返回状态码1:{str(e)}"}
            result = json.dumps(result, ensure_ascii=False)
            logger.info(f"action最终的输出结果为：{result}")
            yield [Message(role=ASSISTANT, content=result)]
            

    def _prepend_react_prompt(self, messages: List[Message], lang: Literal['en', 'zh']) -> List[Message]:
        tool_descs = []
        if self.action_type == "modelscope_agent":
            for f in self.function_map.values():
                function = f.function
                # logger.info(f"当前解析的function为：{function}") [EN] logger.info(f"The currently parsed function is: {function}")
                name = function.get('name', None)
                name_for_human = function.get('name_for_human', name)
                name_for_model = function.get('name_for_model', name)

                tool_descs.append(
                    TOOL_DESC_MODELSCOPE.format(name_for_human=name_for_human,
                                     name_for_model=name_for_model,
                                     description_for_model=function.get("description_name","")).rstrip())
                    
            for fn in self.function_calls_list:
                name = fn.get('function_name', None)
                name_for_human = fn.get('name_for_human', name)
                name_for_model = fn.get('name_for_model', name)

                tool_descs.append(
                    TOOL_DESC_MODELSCOPE.format(name_for_human=name_for_human,
                                     name_for_model=name_for_model,
                                     description_for_model=fn.get("description_name","")).rstrip())
            
        else:    
            for f in self.function_map.values():
                function = f.function
                name = function.get('name', None)
                name_for_human = function.get('name_for_human', name)
                name_for_model = function.get('name_for_model', name)
                assert name_for_human and name_for_model
                args_format = function.get('args_format', '')
                tool_descs.append(
                    TOOL_DESC.format(name_for_human=name_for_human,
                                     name_for_model=name_for_model,
                                     description_for_model=function.get("description_name",""),
                                     parameters=json.dumps(function['parameters'], ensure_ascii=False),
                                     args_format=args_format).rstrip())

            for fn in self.function_calls_list:
                # logger.info(f"当前的function_list为：{fn}") [EN] logger.info(f"The current function_list is: {fn}")
                name = fn.get('function_name', None)
                name_for_human = fn.get('name_for_human', name)
                name_for_model = fn.get('name_for_model', name)
                assert name_for_human and name_for_model
                args_format = fn.get('args_format', '此工具的输入应为JSON对象。')
                tool_descs.append(
                    TOOL_DESC.format(name_for_human=name_for_human,
                                     name_for_model=name_for_model,
                                     description_for_model=fn.get("description_name",""),
                                     parameters=json.dumps(fn['parameters'], ensure_ascii=False),
                                     args_format=args_format).rstrip())
        tool_descs = '\n\n'.join(tool_descs)
        tool_names = ','.join(tool.name for tool in self.function_map.values()) + ','.join(fn['function_name'] for fn in  self.function_calls_list)
        text_messages = [format_as_text_message(m, add_upload_info=True, lang=lang) for m in messages]
        if self.action_type == "modelscope_agent":
            text_messages[-1].content = PROMPT_MODELSCOPE.format(
            tool_descs=tool_descs,
            tool_names=tool_names,
            query=text_messages[-1].content,
        )
        else:
            text_messages[-1].content = PROMPT_REACT.format(
                tool_descs=tool_descs,
                tool_names=tool_names,
                query=text_messages[-1].content,
            )
        return text_messages
    
def find_tools(text: str,num: int) -> Tuple[bool, str, str, str]:
    ###适配非deepseek模型输出 [EN] ##Adapt non-deepseek model output
    if "Action:" in text or "Final Answer:" in text:
        special_func_token = 'Action:'
        special_args_token = '\nAction Input:'
        special_obs_token = '\nObservation:'
        special_final_token = '\nFinal Answer:'
    elif '"Action":' in text or '"Final Answer:"' in text:
        special_func_token = '"Action":'
        special_args_token = '"Action Input":'
        special_obs_token = '\nObservation:'
        special_final_token = '\nFinal Answer:'
    ####适配deepseek模型输出 [EN] ###Adapt deepseek model output
    else:
        special_func_token = 'Action'
        special_args_token = 'Action Input'
        special_obs_token = 'Observation'
        special_final_token = 'Final Answer'
    func_name, func_args = None, None
    i = -1
    j = -1
    k = -1
    t = -1
    func_tokens = []
    args_token = []
    obs_token = []
    action_names = []
    action_inputs = []
    func_tokens.append(i)
    args_token.append(j)
    obs_token.append(k)
    code = 0
    description = ""
    Invalid_character = ["无","{}","","NONE","[]","未知"]

    #####兼容deepseek的输出格式 [EN] ####Output format compatible with deepseek
    text = text.replace("https//", "https://")  ###适配大模型的链接输出 [EN] ##Adapt the link output of large models
    text = text.replace("https://192.168.0.218081","https://192.168.0.21:8081")
    text = text.replace("http//", "http://")
    text = text.replace("minio-wanwu9000","minio-wanwu:9000")

    ###适配deepseek系列模型的action及参数输出格式 [EN] ##Adapt the action and parameter output format of deepseek series models
    if special_args_token not in text and special_obs_token not in text:
        logger.info(f"-----当前text无{special_args_token}、{special_obs_token}----")
        if special_final_token in text:
            t = text.find(special_final_token, t + 1)
            final_result = text[t+len(special_final_token):].strip()
            action_result = {"func_name": "", "func_params": "", "action_code": 0, "final_result": final_result,
                             "description": ''}
            return action_result
        if special_func_token not in text:
            action_result = {"func_name": "", "func_params": "", "action_code": 0, "final_result": "",
                             "description": ''}
            return action_result

        # match = re.search(r'Action: (.*?)(?=\n|Parameters:|$)', text)
        match = re.search(r'Actions?\s*[:：]?\s*(.*?)(?=(?:Parameters|Action Input)s?\s*[:：]?\s*|$)', text, re.DOTALL)
        if match:
            action_name = match.group(1).replace("\n","").replace("【","").replace("】","")
            action_names.append(action_name)
            action_result = {"func_name": action_names, "func_params": '', "action_code": 0,"final_result":'',"description":''}
        else:
            match = re.search(r'Action (.*?)(?=\n|Parameters:|$)', text)
            if match:
                action_name = match.group(1).replace("【","").replace("】","")
                action_names.append(action_name)
                action_result = {"func_name": action_names, "func_params": '', "action_code": 0,"final_result":'',"description":''}
            else:
                action_result = {"func_name": action_names, "func_params": '', "action_code": 1,"final_result":'',"description":'未识别出action名称'}
        logger.info(f"find_tools的输出结果为：{action_result}")
        return action_result

    for _ in range(num):
        i = text.find(special_func_token, i + 1)
        j = text.find(special_args_token,j+1)
        k = text.find(special_obs_token,k+1)
        func_tokens.append(i)
        args_token.append(j)
        obs_token.append(k)

        if 0 <= i < j:  
            if k < j:  
                text = text.rstrip() + special_obs_token  
            k = text.find(special_obs_token,obs_token[-2]+1)
            func_name = text[i + len(special_func_token):j].strip().replace("【","").replace("】","")
            func_args = text[j + len(special_args_token):k].strip()
            logger.info(f"处理前的func_name:{func_name}")
            logger.info(f"处理前的func_args:{func_args}")
            ###正则化提取function_name(剔除所有的标点符号和空白字符) [EN] ##Regularize extraction function_name (remove all punctuation marks and whitespace characters)
            #pattern_funcname = r'[\W\s]+'
            #func_name = re.sub(pattern_funcname, '', func_name)
            func_name = re.sub(r'^[",]+|[",]+$', '', func_name)
            logger.info(f"正则化之后的func_name:{func_name}")
            ###正则化提取function_args(提取{}之间的内容) [EN] ##Regularization extraction function_args (extract the content between {})
            func_args = func_args.replace("'", '"')
            pattern = r'[\s\S]*({.*?})[\s\S]*'
            match = re.search(pattern, func_args)
            if match:
                func_args = match.group(1)
            logger.info(f"正则化之后的func_args:{func_args}")

            if ((not func_name) or func_name.upper() in Invalid_character ) and special_final_token.replace("\n","") not in text:
                code = 1   
                description = "未识别出action名称"
            elif ((not func_name) or func_name.upper() in Invalid_character)  and special_final_token.replace("\n","") in text:
                code = 0
                description = "完成函数调用，生成最终答案"
            else:
                try:
                    if func_name and ((not func_args) or func_args in Invalid_character):
                        func_args = json.loads(func_args) if func_args == "{}" else func_args
                        code = 2  
                        description = "未识别出参数"  
                    else:
                        func_args = json.loads(func_args)
                        if isinstance(func_args, list) and len(func_args) > 1:
                            func_args = func_args[0]
                        if func_name and func_args and any(v is None  or  (v and str(v).strip() in  Invalid_character) or (not v)  for v in func_args.values()):
                            code = 2
                            description = "参数识别缺失"
                        func_args.pop("key", None)
                except json.JSONDecodeError as e:
                    code = 1
                    description = f"JSON解析报错：{str(e)}"
                action_names.append(func_name)
                action_inputs.append(func_args)
        
    t = text.find(special_final_token, t + 1)
    final_result = text[:i].strip()
    action_result = {"func_name": action_names, "func_params": action_inputs, "action_code": code,"final_result":final_result,"description":description}
    logger.info(f"find_tools的输出结果为：{action_result}")
    return action_result

def parse_tool_selection_response(response_text: str, functions: List[OpenAPIPluginTool]):
        """
        通过检测纯文本响应中的关键词来解析工具选择的大模型回答。
        如果响应文本中包含某个工具的名称，则认为该工具被选中。
        """
        # 将响应文本转换为小写，以便进行不区分大小写的匹配 [EN] Convert response text to lowercase for case-insensitive matching
        response_text_lower = response_text.lower()
        for function in functions:
            # 同样地，将工具名称转换为小写进行匹配 [EN] Likewise, convert tool names to lowercase for matching
            if isinstance(function, OpenAPIPluginTool):
                if function.name.lower() in response_text_lower:
                    return function.name  # 返回匹配到的第一个工具名称 [EN] Returns the first matching tool name
                
            else:
                if function['function_name'].lower() in response_text_lower:
                    return function['function_name']

        # 如果没有找到任何匹配的工具名称，返回 None 或者默认值 [EN] If no matching tool name is found, returns None or the default value
        return None    
