"""
MCP代理模块
负责收发MCP协议
"""

import json
import urllib.error
import urllib.request
from typing import Any, Dict, Optional


class McpProxy:
    """MCP代理类，负责与MCP Server进行交互"""

    def __init__(self, user_token: str, base_url: str, skill_version: str):
        """
        初始化MCP代理
        
        Args:
            user_token: 用户令牌，用于请求头 X-Tencent-Meeting-Token
            base_url: MCP Server的URL地址
            skill_version: 技能版本号，用于请求头 X-Skill-Version
        """
        self.user_token = user_token
        self.base_url = base_url
        self.skill_version = skill_version

    def _build_headers(self) -> Dict[str, str]:
        """
        构建请求头
        
        Returns:
            dict: 请求头字典
        """
        return {
            "Content-Type": "application/json",
            "X-Tencent-Meeting-Token": self.user_token,
            "X-Skill-Version": self.skill_version,
        }

    def request(self, method: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        发送MCP请求
        
        Args:
            method: JSONRPC2.0协议的method参数
            params: JSONRPC2.0协议的params参数
        
        Returns:
            dict: 响应结果
        
        Raises:
            ValueError: baseUrl未配置时抛出
            urllib.error.URLError: 网络请求失败时抛出
        """
        if not self.base_url:
            raise ValueError("baseUrl未配置，请检查config.json中的baseUrl字段")

        # 构建JSONRPC2.0请求体
        request_body = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params or {},
            "id": 1
        }

        # 发送HTTP请求
        headers = self._build_headers()
        data = json.dumps(request_body).encode("utf-8")
        req = urllib.request.Request(
            self.base_url,
            data=data,
            headers=headers,
            method="POST"
        )

        try:
            with urllib.request.urlopen(req) as response:
                response_data = response.read().decode("utf-8")
                return json.loads(response_data)
        except urllib.error.URLError as e:
            raise urllib.error.URLError(f"MCP请求失败: {e}")
