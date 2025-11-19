import requests
from datetime import datetime, timedelta
import json
import configparser
import logging

config = configparser.ConfigParser()
config.read('config.ini',encoding='utf-8')

APP_ID = config["AUTH"]["APP_ID"]
API_KEY = config["AUTH"]["API_KEY"]
SECRET_KEY = config["AUTH"]["SECRET_KEY"]
AUTH_BASE_URL = config["AUTH"]["AUTH_BASE_URL"]


logger = logging.getLogger(__name__)





class AccessTokenManager:

    _instance = None
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(AccessTokenManager, cls).__new__(cls)
            cls._instance.access_token = None
            cls._instance.create_time = None
            cls._instance.AUTH_TYPE = None
            cls._instance.tokens = {}
        return cls._instance
        
    def add_auth(self, auth_type, access_token, create_time):
        """
        添加一个新的认证类型及其数据到tokens字典。

        :param auth_type: 认证类型标识符
        :param access_token: 访问令牌
        :param create_time: 令牌创建时间
        """
        if auth_type not in self.tokens:
            self.tokens[auth_type] = {
                "access_token": access_token,
                "create_time": create_time
            }
        else:
            logger.info(f"Auth type {auth_type} already exists.")
        


    def is_token_expired(self,AUTH_TYPE="gz"):
        # 检查token是否超期（20天） [EN] Check whether the token has expired (20 days)
        self.AUTH_TYPE = AUTH_TYPE
        create_time = self.tokens.get(f"{self.AUTH_TYPE}",{}).get("create_time") 
        if not create_time or (datetime.now() - create_time >= timedelta(days=20)):
            return True
        '''
        if not self.create_time or (datetime.now() - self.create_time >= timedelta(days=20)):
            return True
            '''
        return False

    def get_access_token(self,AUTH_TYPE="gz"):
        self.AUTH_TYPE = AUTH_TYPE
        if self.AUTH_TYPE == "gz":
            APP_ID = config["AUTH"]["APP_ID"]
            API_KEY = config["AUTH"]["API_KEY"]
            SECRET_KEY = config["AUTH"]["SECRET_KEY"]
            AUTH_BASE_URL = config["AUTH"]["AUTH_BASE_URL"]
        else:
            APP_ID = config["AUTH"]["APP_ID_HH"]
            API_KEY = config["AUTH"]["API_KEY_HH"]
            SECRET_KEY = config["AUTH"]["SECRET_KEY_HH"]
            AUTH_BASE_URL = config["AUTH"]["AUTH_BASE_URL_HH"]
        
        if self.is_token_expired(AUTH_TYPE):
            url = f"{AUTH_BASE_URL}/{APP_ID}/token"
            # logger.info(f"token url:{url}")

            headers = {
                "Content-Type": "application/json",
            }
            payload = json.dumps({
                "grant_type": "client_credentials",
                "client_id": API_KEY,
                "client_secret": SECRET_KEY,
            })

            response = requests.post(url, headers=headers, data=payload, verify=False)

            if response.status_code == 200:
                data = response.json().get("data", {})
                self.access_token = data.get("access_token", None)
                self.create_time = datetime.now()
                self.add_auth(self.AUTH_TYPE, self.access_token, self.create_time)
                
        # logger.info(f"self.tokens:{self.tokens}")
        token = self.tokens.get(f"{self.AUTH_TYPE}",{}).get("access_token")
        
        return token

    
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
    # print(access_token_manager.get_access_token())
    print(access_token_manager.get_openai_auth())
