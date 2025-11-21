#!/bin/bash

# ==========================================
# Flask + Gunicorn 多 Worker 启动脚本 [EN] Flask + Gunicorn multi-worker startup script
# ==========================================

# 基础配置 [EN] Basic configuration
APP_NAME="server_open"               # Flask 应用模块名（如 server_open.py） [EN] Flask application module name (such as server_open.py)
APP_VAR="app"                        # Flask 应用对象变量名 [EN] Flask application object variable name
PORT=7258                            # 监听端口 [EN] listening port
WORKERS=${WORKERS:-5}                 
THREADS=${THREADS:-2}                            # 每个 worker 的线程数 [EN] Number of threads per worker
TIMEOUT=300                          # 超时时间 [EN] timeout
LOG_DIRECTORY="./logs"               # 日志目录 [EN] Log directory
LOG_FILE="${LOG_DIRECTORY}/${APP_NAME}_console.log"

# ==========================================
# 检查日志目录 [EN] Check log directory
# ==========================================
if [ ! -d "$LOG_DIRECTORY" ]; then
  echo "日志目录不存在，正在创建..."
  mkdir -p "$LOG_DIRECTORY"
  echo "日志目录已创建：$LOG_DIRECTORY"
fi

# ==========================================
# 检查端口占用 [EN] Check port occupancy
# ==========================================
process_ids=$(lsof -ti:$PORT)
if [ -n "$process_ids" ]; then
  echo "端口 ${PORT} 已被占用，正在杀死相关进程..."
  kill $process_ids || kill -9 $process_ids
  sleep 3
  echo "已释放端口 ${PORT}"
else
  echo "端口 ${PORT} 未被占用"
fi

# ==========================================
# 启动 Flask + Gunicorn 服务 [EN] Start the Flask + Gunicorn service
# ==========================================
echo "启动 Flask 服务 (workers=${WORKERS}, threads=${THREADS}) ..."
nohup gunicorn ${APP_NAME}:${APP_VAR} \
  --bind 0.0.0.0:$PORT \
  --workers $WORKERS \
  --threads $THREADS \
  --timeout $TIMEOUT \
  --log-level info \
  > "$LOG_FILE" 2>&1 &

# ==========================================
# 启动结果检测 [EN] Start result detection
# ==========================================
sleep 3
if lsof -i:$PORT >/dev/null; then
  echo "✅ Flask 服务启动成功！"
  echo "监听端口：$PORT"
  echo "Worker 数量：$WORKERS"
  echo "日志输出：$LOG_FILE"
else
  echo "❌ 启动失败，端口 ${PORT} 未检测到进程，请检查日志：$LOG_FILE"
fi
