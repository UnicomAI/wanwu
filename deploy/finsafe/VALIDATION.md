# FinSafe / OpenCode / 万悟接入验证

验证日期：2026-11-23（Asia/Shanghai）。业务源码未修改，未部署生产服务。

2026-11-24 增量：已将 `eino-agent` 纳入独立 FinSafe 策略；当前镜像及验证结果见
[EINO-VALIDATION.md](EINO-VALIDATION.md)。以下保留 11 月 23 日原始验收记录。

## 结论

两阶段用例通过：

1. Supervisor → FinSafe 0.11.19 → OpenCode 1.19.5 的兼容性用例通过。
2. 万悟真实 `pkg/wga-sandbox.Run` → 受约束 OpenCode 的适配链路用例通过。

第二阶段使用本地模型响应和 AIO shell/file 传输测试替身。它验证现有 Go 适配代码、
真实 OpenCode 和真实 FinSafe 的结合，不代表完整万悟集群已完成验收。

另外，完整派生镜像 `wanwu-wga-finsafe:0.11.19` 已构建并运行成功。
万悟真实适配器 → 真实 AIO → FinSafe 包裹的 OpenCode → 官方 DeepSeek Flash
的联合用例也已通过，详见下文。此项与上述 fixture 测试独立记录。

完整主平台（不含本体扩展）的 DeepSeek 界面真实任务也已通过：实际执行 Python，
生成、预览并逐文件下载销售分析产物，独立数值断言通过，详见
[UI-VALIDATION.md](UI-VALIDATION.md)。历史卡片一键下载和 Mac 测试 VM 内存事件
在该记录中单独列出，不记作通过。

## 正式测试环境

本机 Docker daemon 为 ARM64。正式结果来自 **Docker 承载的 QEMU 完整 x64 Linux
内核环境**，使用真实 x64 发行包，不使用旧版 ARM 运行时替代正式结果。

| 项目 | 实际值 |
| --- | --- |
| 架构 | x86_64 |
| 用户态 | Ubuntu 24.04.5 LTS |
| 内核 | Debian `6.12.117+deb13-amd64`，启用 Landlock、cgroup v2、seccomp |
| Landlock | ABI v6 |
| FinSafe | 官方 Linux x64 0.11.19 完整包 |
| OpenCode | 官方 Linux x64 1.19.5 |
| Supervisor | Ubuntu 软件包 4.2.5 |
| Go 驱动 | Go 1.24.13，CGO_ENABLED=0，GOOS=linux，GOARCH=amd64 |
| 万悟源码 | `3e706cb945294f9c19f2243362e4850daa865d82`，原适配实现不变 |
| 外部模型调用 | 无；guest 无外部网络，仅 loopback 测试服务 |

不是 Ubuntu 原厂内核的测试，也不是客户服务器、CentOS/RHEL 或 Kubernetes 验收。
原生 Linux x64 Docker daemon 仍应按 [测试说明](tests/README.md) 在目标内核上复现。

下载校验值（SHA-256）：

```text
FinSafe finshield-0.11.19-linux-x64.tar.zst
c6ac5c709cd34a2c1168534beb969da1813968159c0444f4271224411753a394

OpenCode opencode-linux-x64.tar.gz v1.19.5
588310a336fa3b9df36b3f0225b26f679d453efaad0cf3d66c3651ef3ec1647a

Linux netboot kernel
14159c607ad9762133db41e479dafecde1d6a5fe715a8d65d292a3262e4b6bed
```

## 第一阶段：兼容性

裸 OpenCode 与包裹后的 OpenCode 使用同一组真实 HTTP 请求。

| 检查 | 结果 |
| --- | --- |
| Supervisor 启动服务，健康接口报告 1.19.5 | 通过 |
| 创建及删除会话 | 通过 |
| Bash 工具在工作区写入、读取文件 | 通过 |
| 裸服务可写同一越界目录，包裹后文件未产生且报告拒绝访问 | 通过 |
| 拒绝越界操作后，合法工具执行仍正常 | 通过 |
| SSE 收到真实工具输出 | 通过 |
| 实际 OpenCode PID 进入 `finsafe.slice/wrapper-self-confine-*` | 通过 |
| 实际 OpenCode PID 的 NoNewPrivs=1、Seccomp=2 | 通过 |

`probe` 中 helper/bwrap 完整 run 链路的 overlay/cgroup/seccomp 功能探测未全部配置，
本接入使用 `self-confine`。上述通过结论来自真实 server PID 和文件操作，不将
诊断命令存在、进程启动或单个 `--version` 当成隔离生效的证据。

## 第二阶段：万悟适配链路

Go 驱动直接调用当前仓库的公开 `Run` 方法，没有重写 HTTP 客户端或事件解析器。
按现有 BFF 调用约定，传入已存在的输出目录及以 `/.` 结尾的输入目录。

| 检查 | 结果 |
| --- | --- |
| 创建工作区并通过 AIO 协议上传模型配置、输入文件 | 通过 |
| 通过当前适配器创建 OpenCode session、提交异步 prompt | 通过 |
| 本地 OpenAI-compatible fixture 返回 Bash tool call，真实 OpenCode 执行 | 通过 |
| 当前 SSE 适配器输出 step_start=2、step_finish=2、tool_use=1、text=1 | 通过 |
| 工具处理上传的 input.txt，产物内容为 WANWU_INTEGRATION_OK | 通过 |
| 越界写入失败，回传 denial.txt 内容为 DENIED | 通过 |
| 下载/解包产物 | 通过 |

适配器和驱动按既有逻辑调用 session/工作区清理；本用例未单独断言删除后的状态。

原始完整控制台输出见 [integration-x64.txt](tests/evidence/integration-x64.txt)，
独立第一阶段输出见 [compatibility-x64.txt](tests/evidence/compatibility-x64.txt)。
成功标志均已检查：

```text
WANWU_COMPATIBILITY_X64=PASS
WANWU_INTEGRATION_X64=PASS
WANWU_VM_EXIT=0
```

## DeepSeek 配置检查

独立 OpenCode 1.19.5 使用 `deepseek/deepseek-flash`，通过本机已有 HTTP 代理
连接官方 `https://api.deepseek.com/v1`。真实模型调用 write/read 工具创建并读取
`real-model.txt`，内容断言为 `DEEPSEEK_LIVE_OK`，用例通过。
此项使用 ARM64 OpenCode，未包裹 FinSafe，仅用于先确认 provider 配置与连接。
安全元数据见 [deepseek-real.json](tests/evidence/deepseek-real.json)。

示例 [opencode-deepseek.json](examples/opencode-deepseek.json) 通过环境变量引用 key。
实际 key 未写入部署文件或镜像；镜像 metadata/history 和部署文件检查未发现该 key。

## 完整派生镜像、真实 AIO 与 DeepSeek 联调

2026-11-23 19:25:59（Asia/Shanghai）完整用例通过。此项在同一完整 x64 内核环境
内启动真实 Docker 28.5.1（overlay2、cgroup v2），运行上述完整派生镜像。
VM 主机用户态为 Ubuntu 24.04，实际沙箱镜像为原项目的 Ubuntu 22.04。
测试主机从 initramfs 启动，故按
[Docker 官方环境变量说明](https://docs.docker.com/reference/cli/dockerd/#environment-variables)
设置 `DOCKER_RAMDISK=1` 解决测试主机的 pivot_root 限制；部署配置不需要此变量。

| 检查 | 结果 |
| --- | --- |
| 原始入口及 Supervisor 启动实际派生镜像 | 通过 |
| 真实 AIO shell/file API 可用，OpenCode 健康版本 1.19.5 | 通过 |
| 实际 OpenCode PID 81 的 FinSafe cgroup、NoNewPrivs=1、Seccomp=2 | 通过 |
| 普通容器 root 能写入越界测试的父目录（基线检查） | 通过 |
| 万悟原始适配器上传输入、下发任务模型配置并调用 OpenCode | 通过 |
| 官方 DeepSeek Flash 的 3 次真实请求均返回 HTTP 200 | 通过 |
| SSE 实际观察到携带唯一越界路径的 Bash 工具命令 | 通过 |
| 回传 integration.txt=WANWU_INTEGRATION_OK，denial.txt=DENIED | 通过 |
| 独立 docker exec 确认越界文件不存在 | 通过 |
| SSE 事件 step_start=3、step_finish=3、tool_use=3、text=17 | 通过 |
| 完成后实际容器状态为 healthy | 通过 |

真实 key 只在 QEMU 外层的本地转发进程中使用，不进入 guest 或派生镜像。
该转发器连接官方 API 并透传真实响应，不模拟模型；实际 AIO 来自原镜像。
原始成功输出见 [full-image-deepseek-x64.txt](tests/evidence/full-image-deepseek-x64.txt)，
结构化结果见 [full-image-deepseek-result.json](tests/evidence/full-image-deepseek-result.json)，
官方请求状态见 [deepseek-gateway.txt](tests/evidence/deepseek-gateway.txt)。

FinSafe 在原镜像 glibc 2.35 上报告弱版本 `GLIBC_2.39` 缺失警告，但实际服务启动和
上述工具用例通过。该警告尚未由运行时发行方消除，不据此承诺其他功能路径全部兼容。
此用例只验收 OpenCode 路径，原镜像其他非必要进程不属于本次验收范围。

## 配置检查与验收范围

- Dockerfile 构建检查通过，无告警；固定 SHA 的 FinSafe runtime 阶段构建通过。
- Compose 合并检查通过：两个沙箱保留原 networks/environment/skills volumes，
  原入口未覆盖；其他服务配置未改变。未增加 privileged 或 Docker socket。
- shellcheck、shell 语法、Python 语法、Go 构建、git diff 空白检查通过。
- 完整派生镜像构建通过，基于原沙箱 Ubuntu 22.04 镜像，继承入口 `/opt/gem/run.sh`。
  镜像标记为 `wanwu-wga-finsafe:0.11.19`，镜像 ID 为
  `sha256:e03fd5784cbe6a1124dfa3d8189a09c329cc884082be2dd60326a2cb7f9c9992`。
  前述两个阶段使用独立 Ubuntu 测试镜像；完整镜像与实际 AIO 联合验收如上单独记录。
- 已启动本机主平台的 BFF/UI/数据库及相关服务，并完成一条真实界面任务；
  未验证本体扩展、客户 MCP、技能库或 Kubernetes。
- 未承诺 `deny_read_paths` 的敏感读取拒绝、出站网络隔离或会话间强隔离。
  此处通过的安全用例是直接写入边界、cgroup 归属和 seccomp/NoNewPrivs 状态。
- Eino 不继承 OpenCode 的进程约束；11 月 24 日已独立包裹 `eino-agent`，见增量记录。
  独立 AIO API、远程 MCP 仍未纳入包装。

在客户 Linux x64 主机上使用实际部署网络、模型和万悟界面复现真实任务，
仍是生产启用前需要完成的验收。本地完整镜像测试通过不替代目标环境验收。
