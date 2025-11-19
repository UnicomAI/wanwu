import requests
import json

# Flask 服务的地址 [EN] The address of the Flask service
url = "http://172.17.0.1:1990/net_search"

# 要测试的请求数据 [EN] Request data to test
payload = {
    "query": "联通董事长是谁",
    "search_url":"https://api.bochaai.com/v1/web-search",
    "search_key":"sk-e698027f1ad34c3a8a8d405f9c0f5ec4",
    "search_rerank_id":'11'
}

headers = {
    "Content-Type": "application/json"
}

# 发送 POST 请求 [EN] Send POST request
response = requests.post(url, headers=headers, data=json.dumps(payload))
print(response.text)
