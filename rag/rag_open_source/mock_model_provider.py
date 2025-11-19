from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/callback/v1/model/:<model_id>', methods=['GET'])
def get_model(model_id):
    # 根据 model_id 返回不同类型的模型 [EN] Return different types of models based on model_id
    if model_id == "11111":
        data = {
            "modelId": model_id,
            "model": "bge-m3",
            "modelType": "embedding",
            "provider": "OpenAI-API-compatible",
            "config": {
                "apiKey": "",
                "endpointUrl": "xxx",
                "functionCalling": "noSupport"
            },
            "publishDate": "2024-01-01",
            "updatedAt": "2024-05-01",
            "userId": "user_001"
        }
    elif model_id == "22222":
        data = {
            "modelId": model_id,
            "model": "bge-m3",
            "modelType": "rerank",
            "provider": "OpenAI-API-compatible",
            "config": {
                "apiKey": "",
                "endpointUrl": "https://xxxx.com/openapi/bge/v1"
            },
            "publishDate": "2024-01-01",
            "updatedAt": "2024-05-01",
            "userId": "user_002"
        }
    elif model_id == "33333":
        data = {
            "modelId": model_id,
            "model": "deepseek-r1",
            "modelType": "llm",
            "provider": "OpenAI-API-compatible",
            "config": {
                "apiKey": "",
                "endpointUrl": "https://xxxx.com/openapi/compatible-mode/v1"
            },
            "publishDate": "2024-01-01",
            "updatedAt": "2024-05-01",
            "userId": "user_003"
        }
    else:
        data = {}

    return jsonify({
        "code": 0 if data else 1,
        "data": data,
        "msg": "success" if data else "model_id not found"
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=11534)