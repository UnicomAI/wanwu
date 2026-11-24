# 8GB 主平台部署配置

目标是 **8GiB 内存的原生 Linux x86_64 主机、外部模型、轻并发调试**。
保留万悟主平台、数据库、RAG、工作流及 FinSafe 包裹的 OpenCode/Eino；
本体扩展默认不启动。原 README 的 32GB 建议仍适用于更大的工作负载。
不把本配置表述为 4GB 完整主平台或多人并发容量承诺。

## 内存预算

| 常驻部分 | 上限 MiB |
| --- | ---: |
| RAG，所有现有入口保留、各 1 worker | 2048 |
| Elasticsearch，JVM 堆 384MiB | 1024 |
| Kafka，JVM 堆 256MiB | 512 |
| MySQL，buffer pool 64MiB、最多 50 连接 | 256 |
| MinIO / Redis | 192 / 96 |
| BFF / agent-service | 各 192 |
| 其他 9 个 Go 微服务 | 各 96 |
| callback / workflow / nginx | 192 / 128 / 64 |
| AIO + OpenCode + Eino + 工具子进程，共享父 cgroup | 合计 1536 |
| **常驻合计** | **7296（约 7.13GiB）** |

给 8GiB 主机留下约 896MiB 的系统余量。初始化任务另有短时额度：
mysql-setup 128MiB、es-setup 512MiB；这些上限不预留物理内存。
达到上限仍可能触发 OOM，因此该预算需要结合实际任务验证，不能只看容器已启动。
各容器的 RAM+swap 上限相等，共同父组也禁止 swap，避免依赖换页通过验证。
MySQL 关闭可选的 `performance_schema` 性能诊断采集；应用数据表和 SQL 功能保留。

Kafka 使用镜像实际识别的 `KAFKA_HEAP_OPTS`。原来的
`KAFKA_CFG_HEAP_OPTS` 单独设置时，实测 JVM 仍以 `-Xmx1024m` 启动。
Go 的 `GOMEMLIMIT` 是 GC 目标，Docker/cgroup 硬上限负责限制整个进程的实际内存。
减少 worker 会降低并发处理能力；大文件、OCR、浏览器和高并发需另行增加额度。

## 在 Linux 主机启用

沿用已有 `.env`、`.env.ontology`；首次将 `.env.finsafe.example` 复制为
`.env.finsafe`，已有配置则继续复用。按照上一级 FinSafe 文档准备 Linux x64、cgroup v2、
Landlock 和 seccomp；派生镜像名为 `wanwu-wga-finsafe:0.11.19`；首次部署需构建，已有同版本镜像可复用。
RAG 启动器通过只读挂载调整 worker 数，不改变业务 Python/Go 源码。

在 **Docker daemon 所在的 Linux 主机**运行一次：

```bash
sudo sh deploy/low-memory/prepare-cgroup.sh > .env.low-memory.new &&
  mv .env.low-memory.new .env.low-memory
```

脚本输出专用 cgroup 根目录、共同父组和只读预算目录三个变量；只有准备成功才
替换环境文件，失败时保留已有配置。
cgroupfs 使用 `/finsafe-wanwu`；systemd 使用专用 slice 管理总额度，另设
`Delegate=yes` 服务子树交给 FinSafe，避免直接管理 systemd 的 slice 子目录。
系统重启后要重新运行准备步骤；systemd 单元在 `/run/systemd/system` 中，不是永久安装。
现有同名运行时单元不会被覆盖，同一次启动应保留并复用 `.env.low-memory`。
不要在这些组中放其他工作负载；本体沙箱不在此 8GB 预算中。

```bash
docker compose --env-file .env.image.amd64 --env-file .env --env-file .env.ontology \
  --env-file .env.finsafe --env-file .env.low-memory \
  -f docker-compose.yaml -f docker-compose.finsafe.yaml -f docker-compose.low-memory.yaml \
  build wga-sandbox

docker compose --env-file .env.image.amd64 --env-file .env --env-file .env.ontology \
  --env-file .env.finsafe --env-file .env.low-memory \
  -f docker-compose.yaml -f docker-compose.finsafe.yaml -f docker-compose.low-memory.yaml \
  up -d
```

启动后检查 `ps`、两个 Agent 健康接口、实际任务和 cgroup OOM 计数。
不要加 `--profile ontology`；该扩展不计入本次预算。
后续管理也应带上同一组覆盖文件，以免还原 worker 数和资源设置。

## 为什么要给 Docker 和 FinSafe 设置共同父组

FinSafe self-confine 会将运行时移到自己的 scope；它可能已经离开 Docker
容器的 cgroup。只设置 `mem_limit` 会漏算这部分进程。
低内存配置将 Docker 容器和 FinSafe 委派子树放在同一个父组，父组的
`memory.max=1610612736` 对整个沙箱合计生效。
策略中原来的单运行时 4G 上限会受到共同父组的更小上限约束，不能突破它。

共同父组以只读方式挂到容器 `/run/finsafe-budget`，便于查看真实合计：

```bash
docker compose --env-file .env.image.amd64 --env-file .env --env-file .env.ontology \
  --env-file .env.finsafe --env-file .env.low-memory \
  -f docker-compose.yaml -f docker-compose.finsafe.yaml -f docker-compose.low-memory.yaml \
  exec wga-sandbox sh -c 'cat /run/finsafe-budget/memory.current /run/finsafe-budget/memory.max /run/finsafe-budget/memory.events'
```

cgroup 驱动和父组语法参考 [Docker 文档](https://docs.docker.com/reference/cli/dockerd/#default-cgroup-parent)，
委派原则参考 [systemd 文档](https://systemd.io/CGROUP_DELEGATION/)。
本机实测驱动为 cgroupfs；systemd 路径须在目标服务器完成验收。

## 验证摘要

全部 21 个主平台服务已在本机启动并完成真实 DeepSeek Skill/Python 任务和
OpenCode 读写、越界拒绝检查。短时采样峰值约 6.15GiB，包含 Mac ARM 上额外的
完整 x64 QEMU 模拟开销；最终 swap、oom_kill 为 0。MySQL 同镜像的空库初始化
在 256MiB、无 swap 下通过。这些是轻任务结果，未做多人压力测试，也不替代
原生目标 Linux 服务器验收。测试工具、日志和截图不属于部署所需文件。

## 回退

先停止沙箱，再将原 FinSafe 策略绑定切回原委派路径，去掉低内存覆盖文件重新
创建相应服务。保留数据卷；不要使用 `down -v`。服务停止且组内无进程后，
再由运维撤销本配置的 cgroup/systemd 运行时单元。单纯移除 Compose 文件不会
自动移除父组的总额度，也不会修改已有容器的运行设置。
