from flask import Flask, request, jsonify
import logging
import requests
import io
import os
import tempfile
import time
from botocore.config import Config
#from oss2.exceptions import ClientError
from fpdf import FPDF
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.lib.units import mm
import textwrap
from docx import Document

app = Flask(__name__)

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s: %(message)s', handlers=[logging.StreamHandler()])

pdfmetrics.registerFont(TTFont('SimHei', 'simhei.ttf'))
#pdfmetrics.registerFont(TTFont('Noto', '/usr/share/fonts/opentype/noto/NotoSansCJKsc-Regular.otf'))



def upload_file_to_minio(formatted_markdown,to_format,filename):
    try:
        url = 'http://localhost:15001/upload'
        path = '/agent/agent_open_source/minio/file/'
        file_path = ''
        if to_format == 'pdf':
            file_path = path+filename+".pdf"
            filename = filename            
            c = canvas.Canvas(file_path, pagesize=A4)
            width, height = A4
            margin = 20 * mm
            line_height = 18
            max_width = width - 2 * margin

            c.setFont("SimHei", 12)
            y = height - margin

            # 设置每行最多多少个字符（估算宽度） [EN] Set the maximum number of characters per line (estimated width)
            max_chars_per_line = int(max_width // 12)  # 每个中文大约占 12pt 宽度 [EN] Each Chinese character takes up approximately 12pt width

            # 按照字符数手动换行 [EN] Manually wrap lines according to the number of characters
            wrapped_lines = []
            for paragraph in formatted_markdown.splitlines():
                wrapped_lines.extend(textwrap.wrap(paragraph, width=max_chars_per_line))
                wrapped_lines.append("")  # 空行分段 [EN] Empty line segmentation

            # 写入 PDF [EN] Write to PDF
            for line in wrapped_lines:
                if y < margin:
                    c.showPage()
                    c.setFont("SimHei", 12)
                    y = height - margin
                c.drawString(margin, y, line)
                y -= line_height
            c.save()
        elif to_format == 'docx':
            doc = Document()
            doc.add_paragraph(formatted_markdown)
            file_path = path+filename+".docx"
            doc.save(file_path)
            filename = filename
        elif to_format == 'txt':
            file_path = path+filename+".txt"
            with open(file_path, "w", encoding='utf-8') as file:
                file.write(formatted_markdown)
            filename = filename
        with open(file_path, "rb") as file:
            files = {"file": file}
            data = {"bucket_name":'agent-prod',"file_name":filename}
            response = requests.post(url, files=files,data=data)
            if response.status_code == 200:
                print("File uploaded successfully")
                print("Download link:", response.json()["download_link"])
                download_link = response.json()["download_link"]
                return filename,download_link
    except Exception as e:
        logging.error(f"Error uploading file to OSS: {e}")
        print("Error uploading file to Minio:", e)
        return None


@app.route("/upload", methods=["POST"])
def upload_file():
    try:
        formatted_markdown = request.form.get('formatted_markdown')
        to_format = request.form.get('to_format')
        filename = request.form.get('title','')
        uploaded_file_name,download_link = upload_file_to_minio(formatted_markdown,to_format,filename)
        return jsonify({"download_link": download_link})
    except Exception as e:
        logging.error(f"Error in upload_file endpoint: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    logging.info("Starting Flask app on port 15002")
    app.run(host="0.0.0.0", port=15002)
