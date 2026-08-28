#!/bin/sh
# 启动时用环境变量渲染 agent-operator-integration-secret.yaml 与 mq_config.yaml
# （执行工厂读取的 DB/Redis/MQ 凭据）到 /sysvol/config，再 exec 镜像原本的二进制。
# 参考 vega-calculate-coordinator 的 render-mysql-and-run.sh：凭据单一来源是项目根
# .env 的 WANWU_MYSQL_* / WANWU_REDIS_* / WANWU_KAFKA_*，不再把用户名密码硬编码进
# 仓库里的 example yaml 文件。
# 容器设置 CONFIG_PROFILE=/sysvol/config，故执行工厂从 /sysvol/config 读取
# agent-operator-integration-secret.yaml（默认路径为 /sysvol/secret，见 config.go）。
set -e

CFG=/sysvol/config
mkdir -p "${CFG}"

# DB/Redis 凭据 -> agent-operator-integration-secret.yaml
cat > "${CFG}/agent-operator-integration-secret.yaml" <<EOF
db:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user_name: "${DB_USER}"
  password: "${DB_PASSWORD}"
  db_name: ${DB_NAME}
  charset: utf8mb4
redis:
  connectInfo:
    username: "${REDIS_USER}"
    password: "${REDIS_PASSWORD}"
    host: ${REDIS_HOST}
    port: ${REDIS_PORT}
    poolSize: 10
EOF

# Kafka/MQ 凭据 -> mq_config.yaml
cat > "${CFG}/mq_config.yaml" <<EOF
mqType: kafka
mqHost: ${KAFKA_HOST}
mqPort: ${KAFKA_PORT}
protocol: sasl_plaintext
tenant: ${KAFKA_TENANT}
auth:
  username: "${KAFKA_USER}"
  password: "${KAFKA_PASSWORD}"
  mechanism: ${KAFKA_MECHANISM}
EOF

exec /app/operator-integration
