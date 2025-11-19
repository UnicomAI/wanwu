#!/bin/bash

# 端口号（根据你服务运行端口修改） [EN] Port number (modify according to the port your service is running on)
PORT=1991

# 设置日志目录和固定日志文件名 [EN] Set the log directory and fixed log file name
LOG_DIR="./logs"
LOG_FILE="$LOG_DIR/doc.log"

# 创建日志目录（如果不存在） [EN] Create the log directory if it does not exist
mkdir -p "$LOG_DIR"

# 检查端口是否被占用 [EN] Check whether the port is occupied
PID_ON_PORT=$(lsof -t -i:$PORT)

if [ -n "$PID_ON_PORT" ]; then
  echo "端口 $PORT 被占用，尝试杀死进程 $PID_ON_PORT..."
  kill -9 $PID_ON_PORT
  echo "进程 $PID_ON_PORT 已被终止。"
fi

# 启动服务并将输出追加写入固定日志文件 [EN] Start the service and append output to a fixed log file
echo "启动 Flask 服务..."
nohup python doc_pra.py >> "$LOG_FILE" 2>&1 &

# 保存进程 ID 到 pid 文件（可选） [EN] Save process ID to pid file (optional)
echo $! > doc.pid

echo "服务已启动，日志写入：$LOG_FILE"
echo "PID: $(cat doc.pid)"