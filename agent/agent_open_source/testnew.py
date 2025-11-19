import requests
import json

url = "http://172.17.0.1:7258/agent"

question = "京东下场发展外卖最近情况是什么"
question = "帮我搜索评分较高的西安凉皮店铺"
#question = "There are many types of semi-supervised learning, which one is the most widely used?"
#question = "How is China Unicom's stock price doing recently?"
#question='Sichuan restaurant near the Forbidden City in Beijing'
#question='What is an intelligent agent'
#question='Help me write a bubble sort code'
#question = "How is the weather in Beijing?"
question = '上传的这篇文章写的什么总结一下'


headers = {
    "Content-Type": "application/json",
    "X-uid":"123"
}




response = requests.post(
    url,

    json={"input": question,"upload_file_url":"https://192.168.0.21:8081/minio/download/api/public/tmpt7cc25tv.txt","system_role":'每次回答最后增加免责声明：本AI助手提供的信息基于公开数据和算法生成，仅供参考。',"plugin_list":[],"search_url":"https://api.bochaai.com/v1/web-search","search_rerank_id":'11',"search_key":"sk-e698027f1ad34c3a8a8d405f9c0f5ec4","function_call":False,"stream":True,"model":'deepseek-v3',"model_url":'http://172.17.0.1:6668/callback/v1/model/1',"use_code":False,"use_search":False,"use_know":False,"do_sample":False,"temperature":0.01,"repetition_penalty":1.1,"auto citation":False,"need_search_list":True,"kn_params":{'knowledgeBase':'123','threshold':0.7,'topk':3,'rerank_id':'','model':'','model_url':''}},
    stream=True,
    headers=headers
)





print("\n💬 答案开始：\n")

try:
    for line in response.iter_lines(decode_unicode=True):
        if line:
            print(line)

except KeyboardInterrupt:
    print("\n⏹️ 用户中断")

print("\n\n✅ 测试完成")
