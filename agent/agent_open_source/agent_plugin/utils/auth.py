import requests
from datetime import datetime, timedelta
import json

import configparser
import logging

config = configparser.ConfigParser()
config.read('config.ini')

APP_ID = config["AUTH"]["APP_ID"]
API_KEY = config["AUTH"]["API_KEY"]
SECRET_KEY = config["AUTH"]["SECRET_KEY"]
AUTH_BASE_URL = config["AUTH"]["AUTH_BASE_URL"]


logger = logging.getLogger(__name__)

class AccessTokenManager:
    def __init__(self):
        self.access_token = None
        self.create_time = None

    def is_token_expired(self):
        # Check whether the token has expired (20 days)
        if not self.create_time or (datetime.now() - self.create_time >= timedelta(days=20)):
            return True
        return False

    def get_access_token(self):
        if self.is_token_expired():
            url = f"{AUTH_BASE_URL}/{APP_ID}/token"
            headers = {
                "Content-Type": "application/json",
            }
            payload = json.dumps({
                "grant_type": "client_credentials",
                "client_id": API_KEY,
                "client_secret": SECRET_KEY,
            })

            response = requests.post(url, headers=headers, data=payload, verify=False)
            # logger.info(f"token_response:{response.text}")
            self.access_token = response.json().get("data")["access_token"]
            self.create_time = datetime.now()

        return self.access_token
    
    
    def get_openai_auth(self,model="deepseek-v3"):       
        return auth_dict.get(model)


access_token_manager = AccessTokenManager()

auth_dict =  {
    "deepseek-chat": {
        "api_key":"sk-5adb495b9cc847aabd4669b31f6e054e",
        "base_url":"https://api.deepseek.com"
    },
    "deepseek-reasoner": {
        "api_key":"sk-5adb495b9cc847aabd4669b31f6e054e",
        "base_url":"https://api.deepseek.com"
    },
    "deepseek-ai/DeepSeek-V3": {
        "api_key":"sk-lourxxkjchptufhdcmomjfxkjtzvagwciibdymldbekjoukj",
        "base_url":"https://api.siliconflow.cn/v1"
    },
    "deepseek-ai/DeepSeek-R1": {
        "api_key":"sk-lourxxkjchptufhdcmomjfxkjtzvagwciibdymldbekjoukj",
        "base_url":"https://api.siliconflow.cn/v1"
    },
    "deepseek-v3": {
        "api_key":"",
        "base_url":"https://maas-gz-api.ai-yuanjing.com/openapi/compatible-mode/v1"
    },
    "deepseek-r1": {
        "api_key":"",
        "base_url":"https://maas-gz-api.ai-yuanjing.com/openapi/compatible-mode/v1"
    },  	
    "ds-r1-32b": {
        "api_key":"",
        "base_url":"https://maas-gz-api.ai-yuanjing.com/openapi/compatible-mode/v1"
    },  
    
}

if __name__ == "__main__":
    access_token_manager = AccessTokenManager()
    print(access_token_manager.get_access_token())