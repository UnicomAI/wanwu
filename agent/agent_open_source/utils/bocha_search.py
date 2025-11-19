import requests
import json
import re
from utils.uni_id import generate_unique_id
from utils.timing import advanced_timing_decorator
import configparser

config = configparser.ConfigParser()
config.read('config.ini', encoding='utf-8')

BOCHA_SUBSCRIPTION_KEY = config["BING"]["BOCHA_SUBSCRIPTION_KEY"]
BOCHA_SEARCH_URL = config["BING"]["BOCHA_SEARCH_URL"]

def clean_text(text):
    """清除文本中的特殊字符和多余的空白，以及HTML标签。"""
    patterns = [
        r'\xa0+', r'\u3000', r'\t+', r'\r+', r'\n+',   # Clear special whitespace characters and multiline newlines
        r'<[/]?b>|&gt;|&lt;'                        # Clear HTML tags
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

@advanced_timing_decorator(task_name="bocha_cleaned_search")
def bocha_cleaned_search(query,search_url,search_key, result_len):
    url = search_url

    payload = json.dumps({
        "query": query,
        "summary": True,
        "count": result_len,
        "page": 1
    })

    headers = {
        'Authorization': search_key,  # Please replace with your own API key
        'Content-Type': 'application/json'
    }

    try:
        # Make a request and check the status code
        response = requests.request("POST", url, headers=headers, data=payload)
        response.raise_for_status()  # If the returned status code is not 2xx, an exception will be thrown
        response_data = response.json()  # Try to parse JSON
        
        # print("Return data:", response_data)
        # print()

        # Mapping returns results
        if 'data' in response_data and 'webPages' in response_data['data']:
            results = response_data['data']['webPages']['value']
            mapped_results = [
                {
                    "type": "SE",
                    "id": item.get("id", ""),
                    "title": clean_text(item.get("name", "")),
                    "snippet": clean_text(item.get("snippet", "")),
                    "link": item.get("url", ""),
                    "datePublished": item.get("datePublished", ""),
                    "dateLastCrawled": item.get("dateLastCrawled", "")
                }
                for item in results
            ]
            return mapped_results
        else:
            print("返回数据中没有 webPages 数据")
            return []
    
    except requests.exceptions.RequestException as req_err:
        print(f"请求发生错误: {req_err}")
        return []

    except json.JSONDecodeError as json_err:
        print(f"JSON 解码错误: {json_err}")
        return []

    except KeyError as key_err:
        print(f"缺少预期的键: {key_err}")
        return []

    except Exception as err:
        # Special handling of different HTTP error codes
        if response.status_code == 403:
            print("错误 403: 余额不足。请前往 https://open.bochaai.com 进行充值")
        elif response.status_code == 400:
            error_message = response.json().get("message", "")
            if "Missing parameter query" in error_message:
                print("错误 400: 请求参数缺失 - 缺少查询参数（query）")
            elif "The API KEY is missing" in error_message:
                print("错误 400: 权限校验失败 - Header 缺少 Authorization")
        elif response.status_code == 401:
            print("错误 401: API KEY 无效，权限校验失败")
        elif response.status_code == 429:
            print("错误 429: 请求频率达到限制，请稍后再试，具体限制详见 API 定价")
        elif response.status_code == 500:
            print("错误 500: 服务器内部错误，请稍后重试")
        else:
            print(f"未知错误发生: {err}")

        return []


if __name__ == '__main__':
    query = "联通的王利民去中国移动了吗"
    result_len = 5
    result = bocha_cleaned_search(query, result_len)
    print(result)