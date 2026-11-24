# 本机万悟主平台界面验收

2026-11-24 的 Eino 包裹及运行时回归见 [EINO-VALIDATION.md](EINO-VALIDATION.md)。
本文件保留 11 月 23 日的 OpenCode 界面任务记录。

日期：2026-11-23，Asia/Shanghai。

## 当前状态

- 按用户确认，启动完整主平台，暂不启用本体扩展。
- 22 个主平台服务定义加 1 个沙箱 TCP 转发服务均已启动；MySQL/ES 初始化任务
  正常退出 0，其余服务运行。配置了健康检查的运行服务均健康。
- RAG 的 8681、10891、8613、20041、20050、15000，工作流 8998/8999，
  文件回调 8669 端口实际监听检查通过。
- 界面 HTTP 200，已通过真实登录表单登录本地管理员。
- **DeepSeek 界面真实任务通过。** 通过模型管理页面导入官方 DeepSeek，在
  WanwuBot 界面提交销售分析任务，主管智能体转交 General Agent，后者通过实际
  OpenCode 执行 Bash、Write、Read 和 Python 脚本。
- 界面显示任务完成、工作空间 4 个文件；通过工作空间的独立下载按钮，实际下载
  `sales.csv`、`summary.csv`、`report.md` 和 `analyze.py`，CSV 内容及报告断言通过。

启动元数据见 [ui-platform-startup.json](tests/evidence/ui-platform-startup.json)，
任务结果及文件 SHA-256 见 [ui-task-result.json](tests/evidence/ui-task-result.json)，
后端真实调用记录见 [ui-task-backend.txt](tests/evidence/ui-task-backend.txt)。

## 本机访问和管理

入口：<http://localhost:8081/aibase/>。仅发布到本机 loopback。

本机环境文件 `.env`、`.env.ontology` 已创建并被 Git 忽略。
实际部署配置及运行数据位于 `/tmp/wanwu-ui`，配置包含本地 JWT 签名密钥，
文件权限为 0600，不应提交或作为公开证据复制。

```bash
docker compose -f /tmp/wanwu-ui/compose.json ps --all
docker compose -f /tmp/wanwu-ui/compose.json stop
docker start wanwu-wga-finsafe-full-x64
docker compose -f /tmp/wanwu-ui/compose.json up -d --pull never
```

主平台使用仓库固定版本的 ARM64 官方镜像，镜像层均按 manifest SHA-256 校验后
导入。为适应本机 16GB 内存、Docker 约 8GB 上限，ES heap 调整为 512MB，Kafka
heap 上限 256MB，MySQL buffer pool 为 128MB，文件回调使用 1 个 worker。
RAG 的三组 Gunicorn 服务分别使用 1 个 worker，两类后台入库消费者各 1 个进程；
以只读挂载本地启动脚本实现，实际服务端口在调整后重新检查通过。
这些是本机临时验收配置，不改变业务源码。

## 真实沙箱连接

`wga-sandbox-wanwu` 是原样转发 TCP 的测试桥接服务，不模拟 API 或模型：

```text
万悟 BFF
  → wga-sandbox-wanwu:8080 / 4096 / 4097
  → finsafe-x64-host:18080 / 14096 / 14097
  → 完整 x64 Linux 内核中的实际沙箱容器
  → Supervisor → FinSafe → OpenCode
```

外层容器名为 `wanwu-wga-finsafe-full-x64`，内层实际镜像为
`wanwu-wga-finsafe:0.11.19`。x64 VM 内存为 2560MB。
启动后实际 OpenCode 的 cgroup、seccomp、NoNewPrivs 检查通过，实际容器健康。
guest 到 BFF 回调端口的 DNS/TCP/HTTP 可达性也已检查。

界面模式移除了独立测试使用的全局 provider allowlist 和模型转发器，模型配置
由万悟按任务下发。外层不再挂载 DeepSeek key。模型管理页面已导入：供应商
`DeepSeek`、模型 `deepseek-flash`、显示名 `DeepSeek-Flash-FinSafe`，官方地址
`https://api.deepseek.com/v1`，工具调用开启，上下文 65536，最大输出 8192。
实际 key 仅保存在本地平台运行数据中，不写入仓库或镜像。

## 实际任务与结果

界面提交的任务位于 `/tmp/wanwu-ui/task.txt`：让通用原子智能体实际使用 Python，
根据六条虚构销售数据生成 `sales.csv`、`summary.csv`、`report.md`，在界面提供
下载入口。独立预期结果为总销售额 39200；城市汇总为上海 14100、深圳 13500、
北京 11600；类别汇总为软件 24300、服务 14900。

任务于 19:54:53 提交，BFF 于 19:59:49 将实际沙箱产物保存到持久工作区。
运行 ID 为 `505ba10a-ad73-47fa-9000-c6d4a20813bd`；界面显示 General Agent 的
10 个步骤完成。下载后的原始 6 行明细与输入逐行一致，城市汇总的顺序和金额
严格匹配预期，报告包含总额和虚构数据说明。

下载产物副本见 [ui-task-artifacts](tests/evidence/ui-task-artifacts/)，界面证据见
[ui-task-completed.jpg](tests/evidence/ui-task-completed.jpg) 和
[ui-task-snapshot.txt](tests/evidence/ui-task-snapshot.txt)。

## 本次发现的限制

- 历史任务卡片的一键下载按钮未触发下载请求或浏览器下载事件；本次通过工作空间
  的逐文件下载按钮成功取回全部 4 个文件，报告预览也通过。没有把卡片一键下载
  记为通过，未修改现有前端业务源码。
- 模型中途返回的相对路径链接会导致页面跳转。实际产物应通过最终工作空间入口
  获取；本次任务的最终工作空间已持久化，历史记录可重新打开。
- 界面任务完成、产物保存后，外层测试虚拟机容器曾触发 3584MB 内存上限，退出
  137。事件见 [ui-vm-memory-event.json](tests/evidence/ui-vm-memory-event.json)。
  随后将外层测试容器额度调整为 4GiB，并将 TCG 翻译缓存明确限制为 128MiB，
  保持 guest 内存及 FinSafe 策略不变。缓存参数含义见
  [QEMU 官方说明](https://www.qemu.org/docs/master/system/qemu-manpage.html)。
  恢复后重新检查：实际沙箱 healthy，实际 OpenCode PID 81 的 cgroup、seccomp、
  NoNewPrivs 均通过，见 [ui-protected-server.json](tests/evidence/ui-protected-server.json)。
  该 Mac 模拟环境未做长时间稳定性或并发压测。
- 本次验收主平台的一条通用智能体执行链路；未启用本体扩展，未逐项验收知识库、
  工作流、OpenClaw、远程 MCP 或客户技能。FinSafe 的边界限制仍见
  [VALIDATION.md](VALIDATION.md)，不能据此宣称所有执行端均受保护。
