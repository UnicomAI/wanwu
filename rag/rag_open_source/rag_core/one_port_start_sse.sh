#!/bin/bash



# 定义日志文件名称 [EN] Define log file name
BASE_LOG_FILE="kb_sse_"

# 循环10次，从10876到10885的端口范围内启动FastAPI应用 [EN] Loop 10 times and start the FastAPI application in the port range from 10876 to 10885
for PORT in {10885..10885}; do
  # 检查当前端口是否被占用，并获取使用该端口的进程ID [EN] Check whether the current port is occupied and get the process ID using the port
  PID=$(lsof -i:$PORT -t)

  # 如果存在使用当前端口的进程，则杀掉这些进程 [EN] If there are processes using the current port, kill these processes
  if [ ! -z "$PID" ]; then
    echo "$PORT端口已被占用，进程ID为$PID，正在尝试杀掉..."
    kill -9 $PID
    echo "进程已被杀掉。"
  fi

  # 启动FastAPI应用，并将输出重定向到指定的日志文件，同时在后台运行 [EN] Start the FastAPI application and redirect the output to the specified log file while running in the background
  echo "正在启动FastAPI应用，端口号为$PORT..."
  echo $BASE_LOG_FILE$PORT.log
    # APP_PORT=10885 uvicorn know_sse:app --host 0.0.0.0 --port $APP_PORT
    LOG_FILE=$BASE_LOG_FILE$PORT uvicorn know_sse:app --host 0.0.0.0 --port $PORT
    # APP_PORT=$PORT nohup uvicorn know_sse:app --host 0.0.0.0 --port $PORT >> $LOG_FILE$PORT.log 2>&1 &
  echo "FastAPI应用启动成功，日志文件为./logs/$BASE_LOG_FILE$PORT.log。"
done
