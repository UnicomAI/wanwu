import os
import json
import uuid
import json
import openpyxl
import pandas as pd
import requests


def split_chunks(excel_file_path_label):
    # Select the specified worksheet
    sub_doc = []
    docs = []
    
    df = pd.read_excel(excel_file_path_label,usecols = ['问题','标准答案'])
    question_list = df['问题'].astype(str).tolist()
    answer_list = df['标准答案'].astype(str).tolist()
    data = df.values.tolist()
    for row in data:
        if len(row)!=2:
            print(row)
            continue
        question=row[0]
        answer=row[1]
        if len(answer)>1000:
            continue
        sub_doc.append({'content':answer,'embedding_content':question})
        docs.append(question+' '+answer)
    return sub_doc,docs
    
