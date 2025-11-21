#!/bin/bash
# main.sh

echo "开始执行主脚本"
#export PATH = "/root/miniconda3/bin:/root/miniconda3/condabin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
source /root/miniconda3/etc/profile.d/conda.sh

# 启动其他脚本 [EN] Start other scripts
conda activate agent
bash start.sh

echo "开始执行chatdoc脚本"
bash startdoc.sh

cd agent_plugin
echo "开始执行action脚本"
bash start_action_server.sh

cd /agent/agent_open_source/minio
conda activate minio
echo "开始执行三个minio节点服务脚本"
bash startstr.sh
bash startminio.sh
bash startpra.sh


echo "所有脚本执行完i毕"
