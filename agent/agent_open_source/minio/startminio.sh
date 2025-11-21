#!/bin/bash  
  
# 定义日志文件的路径，这里使用当前目录下的flask_app.log [EN] Define the path of the log file. Here, flask_app.log in the current directory is used.  
LOGFILE="minio.log"  
  
# 使用nohup在后台运行Flask应用，并将输出重定向到日志文件 [EN] Use nohup to run a Flask app in the background and redirect output to a log file  
# 注意：将下面的/path/to/your/app.py替换为你的Flask应用脚本的实际路径 [EN] NOTE: Replace /path/to/your/app.py below with the actual path to your Flask application script  

nohup python3 /agent/agent_open_source/minio/minio_open.py > "$LOGFILE" 2>&1 &  
  
echo "Flask app started in the background."
