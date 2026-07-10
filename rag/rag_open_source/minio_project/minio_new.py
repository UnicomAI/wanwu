from flask import Flask, request, jsonify
from minio import Minio
import os
import tempfile
import logging
import requests
import json

app = Flask(__name__)

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s: %(message)s', handlers=[logging.StreamHandler()])

# ======= 从环境变量中获取mimio配置信息 =======
MINIO_ADDRESS = os.getenv("MINIO_ADDRESS")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY")
if MINIO_ADDRESS is None or MINIO_ACCESS_KEY is None or MINIO_SECRET_KEY is None:
    MINIO_ADDRESS = "172.18.0.1:9000"
    MINIO_ACCESS_KEY = "root"
    MINIO_SECRET_KEY = "A2pp123456"
# ======= 从环境变量中获取mimio配置信息 =======
# 配置Minio服务器
minio_client = Minio(
    MINIO_ADDRESS,
    access_key=MINIO_ACCESS_KEY,
    secret_key=MINIO_SECRET_KEY,
    secure=False
)
default_bucket_name = "rag-public"
REPLACE_MINIO_DOWNLOAD_URL = os.getenv("PUBLIC_MINIO_DOWNLOAD_URL")
if REPLACE_MINIO_DOWNLOAD_URL is None:
    response = requests.get('http://bff-service:6667/v1/api/deploy/info')
    response_data = response.json()
    REPLACE_MINIO_DOWNLOAD_URL = response_data['data']['webBaseUrl']


def upload_file_to_minio(file_stream, original_filename, bucket_name=default_bucket_name, overwrite_file_name=None):
    try:
        # 获取文件扩展名
        _, file_extension = os.path.splitext(original_filename)
        
        # 创建临时文件
        temp_file_path = tempfile.mktemp(suffix=file_extension)
        with open(temp_file_path, "wb") as temp_file:
            temp_file.write(file_stream.read())

        # 获取临时文件的文件名
        file_name = os.path.basename(temp_file_path)

        if overwrite_file_name:
            file_name = overwrite_file_name + file_extension
        
        with open(temp_file_path, "rb") as file_data:
            file_stat = os.stat(temp_file_path)
            minio_client.put_object(bucket_name, file_name, file_data, file_stat.st_size)

        # 删除临时文件
        os.remove(temp_file_path)
        logging.info(f"File '{file_name}' uploaded to Minio bucket '{bucket_name}'")

        return file_name
    except Exception as e:
        logging.error(f"Error uploading file to Minio: {e}")
        print("Error uploading file to Minio:", e)
        return None

@app.route("/upload", methods=["POST"])
def upload_file():
    try:
        # 从请求中获取文件
        uploaded_file = request.files["file"]
        file_stream = uploaded_file.stream
        original_filename = uploaded_file.filename

        # ------ Get bucket name and file name from request data ------
        bucket_name = request.form.get("bucket_name", default_bucket_name)
        overwrite_file_name = request.form.get("file_name", None)
        # ------------------------------------------------------------

        uploaded_file_name = upload_file_to_minio(file_stream, original_filename, bucket_name, overwrite_file_name)
        if uploaded_file_name:
            download_link = f"{REPLACE_MINIO_DOWNLOAD_URL}/{bucket_name}/{uploaded_file_name}"
            logging.info(f"File uploaded successfully, download link: {download_link}")
            
            return jsonify({"download_link": download_link})
        else:
            return jsonify({"error": "Failed to upload file to Minio."}), 500
    except Exception as e:
        logging.error(f"Error in upload_file endpoint: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    logging.info("Starting Flask app on port 15000")
    app.run(host="0.0.0.0", port=15000)

