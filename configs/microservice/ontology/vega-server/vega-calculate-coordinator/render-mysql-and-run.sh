#!/bin/sh
# 启动时用环境变量渲染 mysql.properties（openLooKeng SqlConfiguration 读取的 JDBC 下推规则库连接），
# 再 exec 镜像原本的 launcher。凭据单一来源是项目根 .env 的 WANWU_MYSQL_* / ONTOLOGY_DB_NAME，
# 不再把用户名密码硬编码进仓库里的 properties 文件。
set -e

ETC=/opt/openlookeng/hetu-server/etc

cat > "${ETC}/mysql.properties" <<EOF
jdbc.driver=com.mysql.jdbc.Driver
jdbc.url=jdbc:mysql://${DB_HOST}:${DB_PORT}/${DB_NAME}?useUnicode=true&useSSL=false&allowPublicKeyRetrieval=true&passwordCharacterEncoding=utf-8
jdbc.username=${DB_USER}
jdbc.password=${DB_PASSWORD}
max.connection=1000
min.idle=10
max.idle=50
EOF

exec /opt/openlookeng/hetu-server/bin/launcher run
