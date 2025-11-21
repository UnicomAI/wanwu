# 定义日志文件名称 [EN] Define log file name
BASE_LOG_FILE="asyn_add_"

#kill现有asyn_add_file进程 [EN] kill existing asyn_add_file process
ps -ef | grep '[a]syn_add_file' | grep -v grep | awk '{print $2}' | xargs kill -9
# 发送重启信号 [EN] Send restart signal
eval "$(conda shell.bash hook)"
conda activate rag-new
nohup ./asyn_doc_status_init > logs/init_asyn.out 2>&1
sleep 1

#循环2次，启动asyn_add [EN] Loop 2 times and start asyn_add
for ADDID in $(seq -f "%03g" 1 2)
do
    LOG_FILE=$ASYNC_ADD_FILE_BASE_LOG_FILE$ADDID nohup ./asyn_add_file >>logs/start_asyn_add.out 2>&1 &
    echo "应用启动成功，日志文件为./logs/$BASE_LOG_FILE$ADDID.log。"
done
