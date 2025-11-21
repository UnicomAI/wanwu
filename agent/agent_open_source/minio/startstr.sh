#!/bin/bash

# 端口号（根据你服务运行端口修改） [EN] Port number (modify according to the port your service is running on)
PORT=15002

# 设置日志目录 [EN] Set log directory
LOG_DIR="./logs"
LOG_FILE="$LOG_DIR/str_minio_$(date +'%Y-%m-%d_%H-%M-%S').log"

# 创建日志目录（如果不存在） [EN] Create the log directory if it does not exist
mkdir -p "$LOG_DIR"

# 检查端口是否被占用 [EN] Check whether the port is occupied
PID_ON_PORT=$(lsof -t -i:$PORT)

if [ -n "$PID_ON_PORT" ]; then
  echo "端口 $PORT 被占用，尝试杀死进程 $PID_ON_PORT..."
  kill -9 $PID_ON_PORT
  echo "进程 $PID_ON_PORT 已被终止。"
fi

# 启动服务并将输出写入日志 [EN] Start the service and write the output to the log
echo "启动 Flask 服务..."
nohup python str_minio.py > "$LOG_FILE" 2>&1 &

echo "服务已启动，日志保存在：$LOG_FILE"
