import requests
import json

# The address of the Flask service
url = "http://172.17.0.1:1990/net_search"

# Request data to test
payload = {
    "query": "联通董事长是谁",
    "search_url":"https://api.bochaai.com/v1/web-search",
    "search_key":"sk-e698027f1ad34c3a8a8d405f9c0f5ec4",
    "search_rerank_id":'11'
}

headers = {
    "Content-Type": "application/json"
}

# Send POST request
response = requests.post(url, headers=headers, data=json.dumps(payload))
print(response.text)
