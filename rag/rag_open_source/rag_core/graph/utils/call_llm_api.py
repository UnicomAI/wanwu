import os
import time
import json
import requests
import re

from dotenv import load_dotenv

from graph.utils.logger import logger

load_dotenv()

class LLMCompletionCall:
    def __init__(self, llm_model="", llm_base_url="", llm_api_key="", temperature=0.001):
        self.temperature = temperature
        if not llm_model or not llm_base_url or not llm_api_key:
            self.llm_model = os.getenv("LLM_MODEL", "deepseek-chat")
            self.llm_base_url = os.getenv("LLM_BASE_URL", "https://api.deepseek.com")
            self.llm_api_key = os.getenv("LLM_API_KEY", "")
        else:
            self.llm_model = llm_model
            self.llm_base_url = llm_base_url
            self.llm_api_key = llm_api_key
        if not self.llm_api_key:
            raise ValueError("LLM API key not provided")
        self.llm_timeout = int(os.getenv("LLM_TIMEOUT", "300"))
        self.llm_max_retries = int(os.getenv("LLM_MAX_RETRIES", "3"))

    def call_api(self, content: str) -> str:
        """
        Call API to generate text with retry mechanism.
        
        Args:
            content: Prompt content
            
        Returns:
            Generated text response
        """
            
        last_err = None
        for i in range(self.llm_max_retries):
            try:
                headers = {"Content-Type": "application/json", "Authorization": f"Bearer {self.llm_api_key}"}
                llm_data = {
                    "model": self.llm_model,
                    "temperature": self.temperature,
                    "stream": False,
                    "enable_thinking": False,
                    "messages": [{"role": "user", "content": content}],
                }
                response = requests.post(self.llm_base_url, json=llm_data, headers=headers, verify=False, timeout=self.llm_timeout)
                if response.status_code != 200:
                    raise RuntimeError(f"LLM http {response.status_code}")
                result_data = json.loads(response.text)
                raw = result_data["choices"][0]["message"]["content"] or ""
                clean_completion = self._clean_llm_content(raw)
                return clean_completion
            except Exception as e:
                last_err = e
                if i < self.llm_max_retries - 1:
                    time.sleep(1.5 * (2 ** i))
                else:
                    logger.error(f"LLM api calling failed. Error: {e}")
                    raise e

    def _clean_llm_content(self, text: str) -> str:
        if not isinstance(text, str):
            return ""
        t = text.replace("\r\n", "\n").replace("\r", "\n").strip()
        t = re.sub(r"[\u200B-\u200D\uFEFF]", "", t)
        end_think_re = re.compile(r"</\s*think\s*>", re.IGNORECASE)
        m_end = end_think_re.search(t)
        if m_end:
            t = t[m_end.end():].strip()
        fence_re = re.compile(r"^\s*```(?:\s*\w+)?\s*\n(?P<body>[\s\S]*?)\n\s*```\s*$", re.MULTILINE)
        m = fence_re.match(t)
        if m:
            t = m.group("body").strip()
        else:
            if t.startswith("```") and t.endswith("```") and len(t) >= 6:
                t = t[3:-3].strip()

        if t.lower().startswith("json\n"):
            t = t.split("\n", 1)[1].strip()

        return t
