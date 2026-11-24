#!/bin/sh
# Deployment-only worker sizing. The original /model_extend/start.sh calls these
# five scripts in their existing directories and keeps all RAG endpoints running.
set -eu
workers=${WANWU_RAG_WORKERS:-1}
ingest_workers=${WANWU_RAG_INGEST_WORKERS:-1}
for count in "$workers" "$ingest_workers"; do
  case "$count" in ''|*[!0-9]*|0) echo 'Worker counts must be positive integers' >&2; exit 2 ;; esac
done

case "${0##*/}" in
  start_run.sh)
    LOG_FILE=run_logs_8681 nohup gunicorn -w "$workers" -b 0.0.0.0:8681 \
      --timeout 600 --preload run:app >/dev/null 2>&1 &
    ;;
  start_sse.sh)
    LOG_FILE=kb_sse_10891 nohup gunicorn know_sse:app -w "$workers" \
      -k uvicorn.workers.UvicornWorker -b 0.0.0.0:10891 --timeout 600 --preload >/dev/null 2>&1 &
    ;;
  start_asyn_add.sh)
    python -u asyn_doc_status_init.py >logs/init_asyn.out 2>&1 ||
      echo 'RAG status initialization failed; see logs/init_asyn.out' >&2
    i=1
    while [ "$i" -le "$ingest_workers" ]; do
      label=$(printf '%03d' "$i")
      LOG_FILE=asyn_add_$label nohup python -u asyn_add_file.py >/dev/null 2>&1 &
      LOG_FILE=graph_asyn_add_$label nohup python -u graph_asyn_add_file.py >/dev/null 2>&1 &
      i=$((i + 1))
    done
    ;;
  start_server.sh)
    GRAPH_SERVER_LOG_FILE=graph_server_20050 nohup gunicorn -w "$workers" \
      -k uvicorn.workers.UvicornWorker -t 600 graph_server:app \
      --bind 0.0.0.0:20050 --preload >log.out 2>&1 &
    ;;
  start_es_server.sh)
    PYTHONUNBUFFERED=1 nohup gunicorn --workers "$workers" --bind 0.0.0.0:20041 \
      --timeout 600 --preload es_rag_server:app >es_rag_server_console.log 2>&1 &
    ;;
  *) echo "Unsupported RAG launcher: $0" >&2; exit 2 ;;
esac
