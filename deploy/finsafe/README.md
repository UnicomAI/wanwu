# FinSafe Agent 沙箱部署

通过派生镜像和 Supervisor 配置包裹两个执行端，不修改万悟或 Eino 的业务源码：

```text
Supervisor → FinSafe self-confine → OpenCode :4096 → 工具子进程
Supervisor → FinSafe self-confine → eino-agent :4097 → Shell / Python
```

保留原镜像入口 `/opt/gem/run.sh`、端口、HTTP 协议和技能挂载。
两个执行端分别使用策略和 cgroup；禁止无保护 fallback，Eino 原有命令及 Skill
变量保护继续开启。策略允许任务工作区、缓存和临时目录写入，程序和技能路径只读。

## 运行要求与启用

需要原生 Linux x86_64、rootful Docker、cgroup v2、Landlock、seccomp，以及
可委派的 cpu/memory/pids 控制器。本配置固定官方 FinSafe 0.11.19 Linux x64
完整发行包并校验 SHA-256；没有提供 ARM64 或 rootless 部署方案。

使用 [8GB 主平台部署说明](../low-memory/README.md) 准备共同父 cgroup、构建镜像
并启动服务。该配置同时约束 AIO 和 FinSafe 进程树的合计内存；仅设置 Docker
`mem_limit` 会漏算被 FinSafe 移到其他 scope 的进程。

- `Dockerfile`：安装 FinSafe，覆盖两个 Supervisor 启动项，保留原入口。
- `opencode.conf` / `eino-agent.conf`：通过 `self-confine` 启动对应运行时。
- `opencode.yaml` / `eino-agent.yaml`：独立文件系统和资源策略。
- `docker-compose.finsafe.yaml`：派生镜像、只读策略挂载及组合健康检查。

健康检查要求 OpenCode `/global/health`、Eino `/health` 和 AIO 8080 可用。
启用后还应检查实际 server PID 的 `/proc/<pid>/status` 中 `NoNewPrivs=1`、
`Seccomp=2`，其 cgroup 位于 `finsafe.slice/wrapper-self-confine-*`，并在界面运行
实际模型任务。Supervisor 日志位于 `/var/log/gem/opencode.log`、`eino-agent.log`。
缺少控制器或策略初始化失败时修复部署条件，不要启用 fallback 或 audit 启动模式。

## 模型与保护范围

模型仍由万悟按任务下发；DeepSeek 使用官方地址 `https://api.deepseek.com/v1`
及实际导入的模型名，key 保存在平台配置中，不写入镜像或 Git。
OpenCode 的 HOME/XDG 固定到 `/home/root`；原 `/root` 登录状态不会自动迁移。

本次保护 OpenCode、Eino 及各自直接派生的进程树。`agent-service`、BFF、独立
AIO shell/file API、远程 MCP 不在包装范围内；跨 HTTP 调用不会继承进程约束。
`network: host` 指共享容器网络命名空间，不限制出站网络，也不启用 Docker host
networking。共享工作区、`/proc` 和服务额度不提供会话间强隔离。

任务 `.env` 保持可用，因此关闭默认敏感文件名过滤；策略中的 `deny_read_paths`
不代表已验收所有敏感读取拒绝语义。不要在共享沙箱中存放其他用户的秘密。
单运行时策略上限 4G、512 个进程、2 CPU，是整个服务的上限；8GB 配置再由共同
父组将两个运行时及 AIO 合计限制为 1536MiB，不是逐任务配额。

本地完整 x64 内核环境中，真实 DeepSeek、Skill/Python、越界写入拒绝及实际进程
约束检查已通过。原 Ubuntu 22.04 镜像仍有 `GLIBC_2.39` 弱版本警告；这些结果
不替代目标 Linux 主机验收。systemd 路径也需在目标服务器复验。
