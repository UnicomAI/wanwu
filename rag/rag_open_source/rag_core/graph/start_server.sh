#!/bin/bash

# 定义日志文件名称 [EN] Define log file name
BASE_LOG_FILE="graph_server_"

for PORT in {20050..20050}; do

  # ps -ef | grep $PORT | grep -v grep | awk '{print $2}' | xargs kill -9
  # sleep 2

  PID=$(lsof -i:$PORT -t)
  # 如果存在使用当前端口的进程，则杀掉这些进程 [EN] If there are processes using the current port, kill these processes
  if [ ! -z "$PID" ]; then
    echo "$PORT端口已被占用，进程ID为$PID，正在尝试杀掉..."
    kill -9 $PID
    echo "进程已被杀掉。"
  fi
  sleep 2

  # 启动FastAPI应用，并将输出重定向到指定的日志文件，同时在后台运行 [EN] Start the FastAPI application and redirect the output to the specified log file while running in the background
  echo "正在启动FastAPI应用，端口号为$PORT..."
  echo $BASE_LOG_FILE$PORT.log
    # GRAPH_SERVER_LOG_FILE=$BASE_LOG_FILE$PORT nohup uvicorn graph_server:app --workers 5 --host 0.0.0.0 --port $PORT --timeout-keep-alive 3 &
    GRAPH_SERVER_LOG_FILE=$BASE_LOG_FILE$PORT nohup gunicorn -w 10 -k uvicorn.workers.UvicornWorker -t 600 graph_server:app --bind 0.0.0.0:$PORT > log.out 2>&1 &
  echo "FastAPI应用启动成功，日志文件为./logs/$BASE_LOG_FILE$PORT.log。"
done