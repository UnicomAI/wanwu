import os
import re
from typing import List, Optional
import logging
import json
import requests
from jsonschema import RefResolver
from utils.actions.tools.base import BaseTool, register_tool
from pydantic import BaseModel, ValidationError
from requests.exceptions import RequestException, Timeout
from utils.actions.tools.base import TOOL_REGISTRY

MAX_RETRY_TIMES = 3

logger = logging.getLogger(__name__)

class ParametersSchema(BaseModel):
    name: str
    description: str
    required: Optional[bool] = True
    type: str


class ToolSchema(BaseModel):
    name: str
    description: str
    parameters: List[ParametersSchema]


@register_tool('openapi_plugin')
class OpenAPIPluginTool(BaseTool):
    """
     openapi schema tool
    """
    name: str = 'openapi_plugin'
    description: str = 'This is a api tool that ...'
    parameters: list = []

    def __init__(self, cfg, name):
        super().__init__(cfg)
        self.name = name
        self.cfg = cfg.get(self.name, {})
        self.is_remote_tool = self.cfg.get('is_remote_tool', False)
        # remote call
        self.url = self.cfg.get('url', '')
        self.token = self.cfg.get('token', '')
        self.header = self.cfg.get('header', '')
        self.method = self.cfg.get('method', '')
        self.parameters = self.cfg.get('parameters', [])
        self.description = self.cfg.get('description',
                                        'This is a api tool that ...')
        self.description_name = self.cfg.get('description_name','')
        self.responses_param = self.cfg.get('responses_param', [])
        super().__init__(cfg)

    def call(self, params: dict, **kwargs):
        # logger.info(f"call函数的入参：{params}") [EN] logger.info(f"Input parameters of call function: {params}")
        if self.url == '':
            result = {
                "action_code":1,
                "description":f"{self.name}没有找到endpoint"

            }
            # raise ValueError(
            #     f"Could not use remote call for {self.name} since this tool doesn't have a remote endpoint"
            # )
            return result
#         if not params:
#             result = {
#                 "action_code":1,
#                 "description":f"大模型识别的参数为空" [EN] "description":f"The parameters identified by the large model are empty"

#             }
#             # raise ValueError(
#             #     f"Could not use remote call for {self.name} since this tool doesn't have a remote endpoint"
#             # )
#             return result

        # json_params = json.loads(params)
        full_params = self._remote_parse_input(**params)
        # params = self._verify_args(json.dumps(full_params))
        params = full_params
        if isinstance(params, str):
            result = {
                "action_code": 1,
                "description": f"参数错误，不是json格式"

            }
            return result
        
        is_use_api_output = params.get("is_use_api_output",False)
        stream = params.get("stream",False)
        logger.info(f"call函数的is_use_api_output：{is_use_api_output}")
        logger.info(f"call函数的stream：{stream}")

        # origin_result = None
        if self.method == 'POST':
            logger.info(f"------post请求--------")
            retry_times = MAX_RETRY_TIMES
            while retry_times:
                retry_times -= 1
                logger.info(f"------retry_times：{retry_times}--------")
                logger.info(f"------url：{self.url}--------")
                logger.info(f"------headers：{self.header}--------")
                try:
                    logger.info(f"------开始请求服务--------")
                    if self.header['Content-Type']=='multipart/form-data':
                        self.header.pop('Content-Type', None)
                        logger.info(f"------Content-Type 为 multipart/form-data 的headers更改：{self.header}--------")
                        response = requests.request(
                            'POST',
                            url=self.url,
                            headers=self.header,
                            data=full_params,timeout=600)
                    else:
                        response = requests.request(
                            'POST',
                            url=self.url,
                            headers=self.header,
                            data=json.dumps(full_params),timeout=600)
                    logger.info(f"response.status_code：{response.status_code}")
                    logger.info(f"response类型：{type(response)}")
                    # logger.info(f"response：{response.text}")

                    if response.status_code == requests.codes.ok and response.content.decode('utf-8'):
                        if is_use_api_output and stream:
                            # logger.info(f"is_use_api_output：{is_use_api_output}")
                            # try:
                            #     for line in response.iter_lines(decode_unicode=True):
                            #         # logger.info(f"call line：{line}")
                            #         yield line
                            # except Exception as e:
                            #     logger.info(f"call response Exception：{str(e)}  response.text:{response.text}")
                            #     for line in response.text:
                            #         logger.info(f"call line：{line}")
                            #         yield line
                            # break
                            return response

                        else:
                            result = {
                                    "action_code": 0,
                                    "description":response.content.decode('utf-8')
                                }
                            
                            return result
                    elif response.status_code == requests.codes.ok and not response.content.decode('utf-8') :
                        result = {
                            "action_code": 1,
                            "description": "action调用返回结果为空"

                        }
                        return result

                    else:
                        result = {
                            "action_code": 1,
                            "description": response.content.decode('utf-8')

                        }
                        logger.info(f"call 返回异常：{result}")
                        return result

                # except Timeout:
                #     continue
                except RequestException as e:
                    logger.info(f"------RequestException ：{str(e)}--------")
                    result = {
                        "action_code": 1,
                        "description": f'RequestException：{str(e)}'
                    }
                    return result
                except Exception as e:
                    logger.info(f"------Exception ：{str(e)}--------")
                    result = {
                        "action_code": 1,
                        "description": f'{str(e)}'

                    }
                    return result

        elif self.method == 'GET':
            retry_times = MAX_RETRY_TIMES
            new_url = self.url
            matches = re.findall(r'\{(.*?)\}', self.url)
            for match in matches:
                if match in full_params:
                    new_url = new_url.replace('{' + match + '}', params[match])
                else:
                    logger.info(
                        f'The parameter {match} was not generated by the model.'
                    )

            while retry_times:
                retry_times -= 1
                try:
                    logger.info(f"当前调用的插件地址为：{ new_url}")
                    logger.info(f"当前入参为：{params}")
                    logger.info(f"当前请求的header为：{self.header}")

                    response = requests.request(
                        'GET', url=new_url, headers=self.header, params=params,timeout=600)

                    if response.status_code == requests.codes.ok and response.content.decode('utf-8'):
                        result = {
                            "action_code": 0,
                            "description": response.content.decode('utf-8')

                        }
                        return result

                    elif response.status_code == requests.codes.ok and not response.content.decode('utf-8') :
                        result = {
                            "action_code": 1,
                            "description": "action调用返回结果为空"

                        }
                        return result
                    else:
                        result = {
                            "action_code": 1,
                            "description": response.content.decode('utf-8')

                        }
                        return result

                # except Timeout:
                #     continue
                except RequestException as e:
                    logger.info(f"------Exception ：{str(e)}--------")
                    result = {
                        "action_code": 1,
                        "description": f'status_code:{e.response.status_code},error message:  {e.response.content.decode("utf-8")}{params}{new_url}'
                    }
                    return result

                except Exception as e:
                    logger.info(f"------Exception ：{str(e)}--------")
                    result = {
                        "action_code": 1,
                        "description": f'{str(e)}'

                    }
                    return result

        else:
            result = {
                "action_code": 1,
                "description": "调用方法错误，必须为POST或GET"

            }
            return result
    def _remote_parse_input(self, *args, **kwargs):
        # 先构建一个字典来存储预设参数值 [EN] First build a dictionary to store the preset parameter values
        preset_params = {param['name']: param.get('value', '') for param in self.parameters if 'value' in param}
        # 更新预设参数值字典，用 kwargs 中的值覆盖预设值（如果存在） [EN] Update the preset parameter value dictionary, overwriting the preset value with the value in kwargs (if it exists)
        preset_params.update(kwargs)
        
        # 初始化一个新的字典用于存储最终的请求参数 [EN] Initialize a new dictionary to store the final request parameters
        restored_dict = {}
        for key, value in preset_params.items():
            if '.' in key:
                # 如果键包含"."，则分割键并创建嵌套的字典结构 [EN] If the key contains ".", split the key and create a nested dictionary structure
                keys = key.split('.')
                temp_dict = restored_dict
                for k in keys[:-1]:
                    temp_dict = temp_dict.setdefault(k, {})
                temp_dict[keys[-1]] = value
            else:
                # 如果键不包含"."，直接将键值对存储到restored_dict中 [EN] If the key does not contain ".", store the key-value pair directly in restored_dict
                restored_dict[key] = value
        
        # logger.info(f"传给tool的参数：{restored_dict}") [EN] logger.info(f"Parameters passed to tool: {restored_dict}")
        return restored_dict


# openapi_schema_convert,register to tool_config.json
def extract_references(schema_content):
    references = []
    if isinstance(schema_content, dict):
        if '$ref' in schema_content:
            references.append(schema_content['$ref'])
        for key, value in schema_content.items():
            references.extend(extract_references(value))
    elif isinstance(schema_content, list):
        for item in schema_content:
            references.extend(extract_references(item))
    return references


def parse_nested_parameters(param_name, param_info, parameters_list, content):
    param_type = param_info['type']
    param_description = param_info.get('example',param_info.get('description',
                                       f'用户输入的{param_name}'))  # 按需更改描述 [EN] Change description as needed
    param_required = param_name in content['required']
    try:
        if param_type == 'object':
            properties = param_info.get('properties')
            if properties:
                # If the argument type is an object and has a non-empty "properties" field,
                # its internal properties are parsed recursively
                for inner_param_name, inner_param_info in properties.items():
                    inner_param_type = inner_param_info['type']
                    inner_param_description = inner_param_info.get("example",inner_param_info.get(
                        'description', f'用户输入的{param_name}.{inner_param_name}'))
                    inner_param_required = param_name.split(
                        '.')[0] in content['required']

                    # Recursively call the function to handle nested objects
                    if inner_param_type == 'object':
                        parse_nested_parameters(
                            f'{param_name}.{inner_param_name}',
                            inner_param_info, parameters_list, content)
                    else:
                        parameters_list.append({
                            'name':
                            f'{param_name}.{inner_param_name}',
                            'description':
                            inner_param_description,
                            'required':
                            inner_param_required,
                            'type':
                            inner_param_type,
                            'enum':
                            inner_param_info.get('enum', '')
                        })
        else:
            # Non-nested parameters are added directly to the parameter list
            parameters_list.append({
                'name': param_name,
                'description': param_description,
                'required': param_required,
                'type': param_type,
                'enum': param_info.get('enum', '')
            })
    except Exception as e:
        raise ValueError(f'{e}:schema结构出错')


def parse_responses_parameters(param_name, param_info, parameters_list):
    param_type = param_info['type']
    param_description = param_info.get('description',
                                       f'调用api返回的{param_name}')  # 按需更改描述 [EN] Change description as needed
    try:
        if param_type == 'object':
            properties = param_info.get('properties')
            if properties:
                # If the argument type is an object and has a non-empty "properties"
                # field, its internal properties are parsed recursively

                for inner_param_name, inner_param_info in properties.items():
                    param_type = inner_param_info['type']
                    param_description = inner_param_info.get(
                        'description',
                        f'调用api返回的{param_name}.{inner_param_name}')
                    parameters_list.append({
                        'name': f'{param_name}.{inner_param_name}',
                        'description': param_description,
                        'type': param_type,
                    })
        else:
            # Non-nested parameters are added directly to the parameter list
            parameters_list.append({
                'name': param_name,
                'description': param_description,
                'type': param_type,
            })
    except Exception as e:
        raise ValueError(f'{e}:schema结构出错')

def add_auth_to_config_entry(config_entry, auth):
    # logger.info(auth)
    if auth['type'] == 'apiKey':
        if auth['in'] == 'header':
            config_entry['header'][auth['name']] = auth['value']
        elif auth['in'] == 'query':
            # Assuming 'parameters' is a list of dictionaries for query parameters
            config_entry['parameters'].append({
                'name': auth['name'],
                'in': 'query',
                'description':"鉴权密钥",
                'required': True,
                'schema': {'type': 'string'},
                'value': auth['value']
            })
        elif auth['in'] == 'body':
            # Ensure there's a structure for the body if it's not already there
            body_structure = config_entry.get('body', {})
            body_structure[auth['name']] = auth['value']
            config_entry['body'] = body_structure

    elif auth['type'] == 'basic':
        import base64
        user_pass = f"{auth['username']}:{auth['password']}"
        encoded_credentials = base64.b64encode(user_pass.encode('utf-8')).decode('utf-8')
        config_entry['header']['Authorization'] = f"Basic {encoded_credentials}"

    elif auth['type'] == 'bearer':
        config_entry['header']['Authorization'] = f"Bearer {auth['token']}"
    elif auth['type'] == 'None':
        config_entry['header']['Authorization'] = None

    return config_entry


def openapi_schema_convert(schema, auth):
    try:
        resolver = RefResolver.from_schema(schema)
        servers = schema.get('servers', [])
        if servers:
            servers_url = servers[0].get('url')
        else:
            logger.info('No URL found in the schema.')
        # Extract endpoints
        endpoints = schema.get('paths', {})
        description = schema.get('info', {}).get('description',
                                                 'This is a api tool that ...')
        config_data = {}
        # Iterate over each endpoint and its contents
        for endpoint_path, methods in endpoints.items():
            for method, details in methods.items():
                summary = details.get('summary', 'No summary').replace(' ', '_')
                description_name = details.get('description', 'No description').replace(' ', '_')
                name = details.get('operationId', 'No operationId')
                url = f'{servers_url}{endpoint_path}'
                security = details.get('security', [{}])
                # Security (Bearer Token)
                authorization = ''
                if security:
                    for sec in security:
                        if 'BearerAuth' in sec:
                            api_token = auth.get('apikey',
                                                 os.environ.get('apikey', ''))
                            api_token_type = auth.get(
                                'apikey_type',
                                os.environ.get('apikey_type', 'Bearer'))
                            authorization = f'{api_token_type} {api_token}'
                if method.upper() == 'POST':
                    requestBody = details.get('requestBody', {})
                    if requestBody:
                        for content_type, content_details in requestBody.get(
                                'content', {}).items():
                            schema_content = content_details.get('schema', {})
                            references = extract_references(schema_content)
                            if references:
                                for reference in references:
                                    resolved_schema = resolver.resolve(reference)
                                    content = resolved_schema[1]
                                    parameters_list = []
                                    for param_name, param_info in content[
                                            'properties'].items():
                                        parse_nested_parameters(
                                            param_name, param_info, parameters_list,
                                            content)
                                    # logger.info(f"\nparameters_list:{parameters_list}")
                                    X_DashScope_Async = requestBody.get(
                                        'X-DashScope-Async', '')
                                    if X_DashScope_Async == '':
                                        config_entry = {
                                            'name': name,
                                            'description': description,
                                            'description_name':description_name,
                                            'is_active': True,
                                            'is_remote_tool': True,
                                            'url': url,
                                            'method': method.upper(),
                                            'parameters': parameters_list,
                                            'header': {
                                                'Content-Type': content_type,
                                                'Authorization': authorization
                                            }
                                        }
                                    else:
                                        config_entry = {
                                            'name': name,
                                            'description': description,
                                            'description_name':description_name,
                                            'is_active': True,
                                            'is_remote_tool': True,
                                            'url': url,
                                            'method': method.upper(),
                                            'parameters': parameters_list,
                                            'header': {
                                                'Content-Type': content_type,
                                                'Authorization': authorization,
                                                'X-DashScope-Async': 'enable'
                                            }
                                        }
                            else:
                                if schema_content['properties'].items():
                                    parameters_list = []
                                    for param_name, param_info in schema_content[
                                        'properties'].items():
                                        parse_nested_parameters(
                                            param_name, param_info, parameters_list,
                                            schema_content)
                                    # logger.info(f"\nparameters_list:{parameters_list}")
                                    X_DashScope_Async = requestBody.get(
                                        'X-DashScope-Async', '')
                                    if X_DashScope_Async == '':
                                        config_entry = {
                                            'name': name,
                                            'description': description,
                                            'description_name':description_name,
                                            'is_active': True,
                                            'is_remote_tool': True,
                                            'url': url,
                                            'method': method.upper(),
                                            'parameters': parameters_list,
                                            'header': {
                                                'Content-Type': content_type,
                                                'Authorization': authorization
                                            }
                                        }
                                else:
                                    config_entry = {
                                            'name': name,
                                            'description': description,
                                            'description_name':description_name,
                                            'is_active': True,
                                            'is_remote_tool': True,
                                            'url': url,
                                            'method': method.upper(),
                                            'parameters': [],
                                            'header': {
                                                'Content-Type': 'application/json',
                                                'Authorization': authorization
                                            }
                        }
                    else:
                        config_entry = {
                            'name': name,
                            'description': description,
                            'description_name':description_name,
                            'is_active': True,
                            'is_remote_tool': True,
                            'url': url,
                            'method': method.upper(),
                            'parameters': [],
                            'header': {
                                'Content-Type': 'application/json',
                                'Authorization': authorization
                            }
                        }
                elif method.upper() == 'GET':
                    parameters_list = details.get('parameters', [])
                    config_entry = {
                        'name': name,
                        'description': description,
                        'description_name':description_name,
                        'is_active': True,
                        'is_remote_tool': True,
                        'url': url,
                        'method': method.upper(),
                        'parameters': parameters_list,
                        'header': {
                            'Authorization': authorization
                        }
                    }
                else:
                    raise 'method is not POST or GET'

                # 把auth添加到config_entry中 [EN] Add auth to config_entry
                config_entry = add_auth_to_config_entry(config_entry,auth)
                config_data[summary] = config_entry
        return config_data
    except Exception as e:
        logger.info(f"插件配置报错：{str(e)}")
        return None

def add_openapi_plugin_to_additional_tool(plugin_cfgs, function_list):
    if plugin_cfgs is None or plugin_cfgs == {}:
        return function_list
    try:
        for name, _ in plugin_cfgs.items():
            openapi_plugin_object = OpenAPIPluginTool(name=name, cfg=plugin_cfgs)
            # logger.info(f"\nopenapi_plugin_object:{openapi_plugin_object}")
            TOOL_REGISTRY[name] = openapi_plugin_object
            function_list.append(name)
        return function_list
    except Exception as e:
        return e
