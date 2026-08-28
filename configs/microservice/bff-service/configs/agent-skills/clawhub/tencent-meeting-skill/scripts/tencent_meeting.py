"""
腾讯会议MCP工具主入口
负责接收用户脚本调用
"""

import json
import os
import sys
from pathlib import Path
from typing import Any, Dict

import local_tools
from mcp_proxy import McpProxy
from utils import get_os_name


def load_config() -> Dict[str, Any]:
    """
    从SKILL目录下的config.json加载配置
    
    Returns:
        dict: 配置内容
    """
    # 获取SKILL目录（scripts的上级目录）
    skill_dir = Path(__file__).parent.parent.absolute()
    config_file = skill_dir / "config.json"

    if not config_file.exists():
        raise FileNotFoundError(f"配置文件不存在: {config_file}")

    with open(config_file, "r", encoding="utf-8") as f:
        return json.load(f)


def print_usage():
    """打印使用说明"""
    usage = """
使用说明：
    python3 tencent_meeting.py <method> [params]

参数说明：
    method  - JSONRPC2.0协议的method参数（必填）
    params  - JSONRPC2.0协议的params参数（选填，JSON格式字符串）
              当method为tools/call时，params为必填参数

调用示例：
    python3 tencent_meeting.py tools/list
    python3 tencent_meeting.py tools/call '{"name": "get_meeting_by_code", "arguments": {"meeting_code": "904854736", "_client_info": {"os": "macos-26", "agent": "workbuddy", "model": "GLM-5"}}}'
"""
    print(usage)


def validate_json(params_str: str) -> Dict[str, Any]:
    """
    校验并解析JSON参数
    
    Args:
        params_str: JSON格式字符串
    
    Returns:
        dict: 解析后的字典
    
    Raises:
        ValueError: JSON格式不正确时抛出
    """
    try:
        params = json.loads(params_str)
        if not isinstance(params, dict):
            raise ValueError("params必须是JSON对象格式")
        return params
    except json.JSONDecodeError as e:
        raise ValueError(f"params参数不是有效的JSON格式: {e}")


def validate_and_fill_client_info(params: Dict[str, Any]) -> Dict[str, Any]:
    """
    校验并填充_client_info字段
    
    当method为tools/call时，如果params的arguments中存在_client_info字段，
    则使用系统真实的os名称替换os字段。
    
    Args:
        params: 请求参数字典
    
    Returns:
        dict: 处理后的参数字典
    
    Raises:
        ValueError: _client_info字段格式错误时抛出
    """
    # 确保arguments字段存在
    if "arguments" not in params:
        return params
    
    arguments = params["arguments"]
    
    # 确保arguments是字典类型
    if not isinstance(arguments, dict):
        raise ValueError("arguments字段必须是JSON对象格式")
    
    # _client_info字段为选填，不存在时直接返回
    if "_client_info" not in arguments:
        return params
    
    client_info = arguments["_client_info"]
    
    # 确保client_info是字典类型
    if not isinstance(client_info, dict):
        raise ValueError("_client_info字段必须是JSON对象格式")
    
    # 获取系统真实的os名称并替换
    real_os_name = get_os_name()
    params["arguments"]["_client_info"]["os"] = real_os_name
    
    return params


def main():
    """主函数"""
    # 检查method参数（必填）
    if len(sys.argv) < 2:
        print("[错误] 缺少method参数\n")
        print_usage()
        sys.exit(1)

    # 解析method参数
    method = sys.argv[1]

    # 解析params参数（选填）
    params_str = sys.argv[2] if len(sys.argv) > 2 else ""

    # 当method为tools/call时，params为必填
    if method == "tools/call" and not params_str:
        print("[错误] method为tools/call时，params参数为必填\n")
        print_usage()
        sys.exit(1)

    # 解析params
    if params_str:
        try:
            params = validate_json(params_str)
        except ValueError as e:
            print(f"[错误] 参数错误: {e}\n")
            print_usage()
            sys.exit(1)
    else:
        params = {}

    # 当method为tools/call时，校验并处理_client_info字段
    if method == "tools/call":
        try:
            params = validate_and_fill_client_info(params)
        except ValueError as e:
            print(f"[错误] 参数错误: {e}")
            sys.exit(1)

    # 获取环境变量中的token
    user_token = os.environ.get("TENCENT_MEETING_TOKEN", "")
    if not user_token:
        print("[错误] 环境变量 TENCENT_MEETING_TOKEN 未设置")
        print("请先设置环境变量：")
        print("    export TENCENT_MEETING_TOKEN='your_token_here'")
        sys.exit(1)

    # 创建MCP代理并发送请求
    try:
        # 读取配置文件
        config = load_config()
        base_url = config.get("baseUrl", "")
        skill_version = config.get("version", "")

        if not base_url:
            print("[错误] 配置错误: config.json中缺少baseUrl字段")
            sys.exit(1)

        proxy = McpProxy(user_token, base_url, skill_version)

        # 本地伪工具调用：直接在 Skill 包内处理，不转发到 MCP Server
        if local_tools.is_local_tool_call(method, params):
            result = local_tools.handle_local_tool_call(method, params)
        else:
            result = proxy.request(method, params)
            # tools/list：在远端结果中追加本地伪工具，便于模型发现并调用
            if method == "tools/list":
                result = local_tools.inject_local_tools(result)

        # 打印结果
        # 当method为tools/call时，打印result结果content数组里面对应type=text的text字段
        if method == "tools/call" and "result" in result:
            inner = result["result"]
            # 服务端返回业务错误时，结构为 {"error": {"code": ..., "message": ...}}，无 content 字段
            if "error" in inner:
                error = inner["error"]
                print(f"[错误] {error.get('message', json.dumps(error, ensure_ascii=False))}")
            elif "content" in inner:
                for item in inner["content"]:
                    if item.get("type") == "text":
                        print(item.get("text", ""))
            else:
                print(json.dumps(result, ensure_ascii=False, indent=2))
        else:
            print(json.dumps(result, ensure_ascii=False, indent=2))
        
    except FileNotFoundError as e:
        print(f"[错误] 配置错误: {e}")
        sys.exit(1)
    except ValueError as e:
        print(f"[错误] 配置错误: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"[错误] 请求失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
