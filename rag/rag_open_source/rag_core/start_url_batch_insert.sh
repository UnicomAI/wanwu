# 定义日志文件名称 [EN] Define log file name
BASE_LOG_FILE="url_batch_insert_logs_"


#kill现有进程 [EN] kill existing process
ps -ef | grep '[u]rl_batch_insert.py' | grep -v grep | awk '{print $2}' | xargs kill -9

# 进入url_parser目录 [EN] Enter the url_parser directory
cd url_parser

#循环5次，启动 [EN] Cycle 5 times and start
for ADDID in $(seq -f "%03g" 1 4)  
do  
    LOG_FILE=$BASE_LOG_FILE$ADDID nohup /root/miniconda3/envs/rag-new/bin/python -u url_batch_insert.py &
    echo "应用启动成功，日志文件为./logs/$BASE_LOG_FILE$ADDID.log。"
done
