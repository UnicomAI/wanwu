import os
import json
import shutil
import numpy as np
import logging
from flask import Flask, jsonify, request, make_response
from flask_cors import CORS
from pathlib import Path
from functools import wraps
from logging.handlers import TimedRotatingFileHandler
from datetime import datetime
import argparse

app = Flask(__name__)
CORS(app, resources={r"/*": {"origins": "*"}})


app.config['JSON_AS_ASCII'] = False
app.config['JSONIFY_MIMETYPE'] ='application/json;charset=utf-8'

@app.route('/list_file_names', methods=['POST']) #List kb
def list_file_names():
    data={
        "code": "0",
        "message": "成功",
        "data":{
            "file_names":["文档1","文档2","文档3"]
        }
    }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response

@app.route('/list_kb_names', methods=['POST']) #List kb
def list_kb_names():
    data={
        "code": "0",
        "message": "成功",
        "data":{
            "kb_names":["知识库1","知识库2","知识库3"]
        }
    }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response

@app.route('/init_kb', methods=['POST']) #initialization
def init_kb():
    data={
        "code": "0",
        "message": "成功"
    }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response

@app.route('/del_kb', methods=['POST']) #initialization
def del_kb():
    data={
        "code": "0",
        "message": "成功"
    }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response

@app.route('/del_files', methods=['POST']) #initialization
def del_files():
    data=  {
        "code": "1",
        "message": "添加失败",
        "data":{
            "success":["文档1.txt","文档3.txt"],
            "failed":["文档2.txt"]
        }
      }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response

@app.route('/search', methods=['POST']) #initialization
def search():
    with open('examples/milvus_search_result.json', 'r') as f:
        res_data = json.loads(f.read())
    response = make_response(json.dumps(res_data, ensure_ascii=False))
    return response

@app.route('/add', methods=['POST']) #initialization
def add():
    init_info = json.loads(request.get_data())
    print(init_info)
    data={
            "code": "0",
            "message": "成功"
        }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response


@app.route('/callback', methods=['POST']) #initialization
def mycallback():
    init_info = json.loads(request.get_data())
    print(init_info)
    data={
            "code": "0",
            "message": "成功"
        }
    response = make_response(json.dumps(data, ensure_ascii=False))
    return response


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int) 
    args = parser.parse_args()
    app.run(host='0.0.0.0', port=args.port)