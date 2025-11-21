import json
import requests
import time
import openpyxl
import pandas as pd
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)




# 设置URL  
url = 'http://localhost:8681/rag/search-knowledge-base'
user_id = "1"
kb_name = 'rag_base_test'
headers = {
            "Content-Type": "application/json",
            "X-uid": user_id
            }

excel_file_path_label =r"RAG-test.xlsx"
# OpenExcelFile
workbook = openpyxl.load_workbook(excel_file_path_label)
# 选择指定的工作Table
sheet = workbook['RAG基线数据集']   
# 读取ExcelFileData
df = pd.read_excel(excel_file_path_label,usecols = ['序号','问题'])
# 提取所需的两列Data
question_list = df['问题'].astype(str).tolist()

# question_list = [ 
#     "怎么排故" for i in range(10)
# ]

for j,question in enumerate(question_list):
    payload = json.dumps({
            "userId":user_id,
            "knowledgeBase": kb_name, 
            "question":question,
            "threshold":0,
            "topK":5,
            "history":[],
            "stream":True,
            "search_field":"emc",
            "model_name":"unicom-72b-chat",
            })
    start_time = time.time()
    # print(payload)
    response = requests.request("POST", url, headers=headers, data=payload,verify=False,stream=True)
    result_data = json.loads(response.text)
    # print(len(result_data["data"]["searchList"]))
    print(len(result_data["data"]["prompt"]))
