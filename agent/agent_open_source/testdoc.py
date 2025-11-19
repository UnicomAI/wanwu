import requests
import json

# 修改为你的服务地址 [EN] Change it to your service address
URL = "http://172.17.0.1:1991/doc_pra"

# 构造测试数据 [EN] Construct test data
payload = {
    "query": "请总结文档中的主要内容。",
    "file_url": [
        "https://192.168.0.21:8081/minio/download/api/public/tmpt7cc25tv.txt"  # 替换为真实可访问的文档 URL [EN] Replaced with a real accessible document URL
    ]
}

headers = {
    "Content-Type": "application/json"
}

def test_doc_pra():
    try:
        print("发送请求到 Flask 服务...")
        response = requests.post(URL, headers=headers, data=json.dumps(payload))
        
        if response.status_code == 200:
            try:
                data = response.json()
                print("返回的 JSON 数据：")
                print(json.dumps(data, ensure_ascii=False, indent=2))
            except json.JSONDecodeError:
                print("返回非 JSON 数据：")
                print(response.text)
        else:
            print(f"请求失败，状态码: {response.status_code}")
            print("返回内容：", response.text)
    except Exception as e:
        print("测试请求出错：", str(e))


if __name__ == "__main__":
    test_doc_pra()
