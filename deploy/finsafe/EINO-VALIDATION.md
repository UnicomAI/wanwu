# Eino 执行端 FinSafe 增量验证

验证日期：2026-11-24（Asia/Shanghai）。在既有 OpenCode 接入上增加 `eino-agent`，
仅修改部署、策略和测试文件，业务 Go 源码不变。

## 接入位置

仓库的 Eino ADK 依赖为 `github.com/cloudwego/eino v0.7.37`。
`internal/agent-service/service/agent-tool/skill_service.go` 为普通智能体的 Skill
选择 `RunnerTypeEinoChatModel`；任务实际经 HTTP 调用沙箱中的 `eino-agent`。
`pkg/wga-sandbox/internal/runner/eino/agent/` 提供 chat-model、Skill、Bash 执行实现。

本次通过 Supervisor 配置分别启动：

```text
Supervisor
  ├─ FinSafe self-confine → OpenCode serve :4096 → 工具子进程
  ├─ FinSafe self-confine → eino-agent serve :4097 → Skill / Bash / Python
  └─ AIO :8080（未包裹）
```

没有包裹整个 `agent-service`，也不需要修改或 fork Eino 框架。
Eino 的 `DISABLE_EINO_GENERIC_GUARD`、`DISABLE_EINO_SKILL_VAR_GUARD` 保持未开启。
FinSafe 禁止无保护降级；两个服务分别使用自己的 cgroup 和策略。

## 实际环境与镜像

- Docker 中运行完整 x64 Linux 内核（Debian 6.12.117+deb13-amd64），
  Landlock ABI 6、cgroup v2、seccomp；Mac ARM 宿主通过 QEMU 系统模拟运行。
- 原始完整沙箱镜像为 Ubuntu 22.04；保留 `/opt/gem/run.sh` 和原 AIO/Supervisor。
- FinSafe 官方 Linux x64 0.11.19、OpenCode 1.19.5、原镜像的 `eino-agent`。
- 当前镜像标签 `wanwu-wga-finsafe:0.11.19`。
- Docker 导出索引：`sha256:84cf3fe9ec27382f7137791dbb95efa8111b93f631f52f29731d33dc7346dfde`。
- 实际镜像配置 ID：`sha256:31999e26ded86afc6bd43e73926ad122e67eca7621e51dc2671ce35c102da016`。
- 旧版只包裹 OpenCode 的主机标签保留为 `wanwu-wga-finsafe:0.11.19-opencode-only`。

原镜像 glibc 2.35 上仍有 FinSafe 弱版本 `GLIBC_2.39` 警告；本次实际用例通过
不能消除发行包兼容性风险，原生目标 Linux 主机仍需复验。

## 已完成的边界验证

使用 `tests/eino_compatibility.py` 连接实际 Eino `/chat` 与 AIO shell API。
边界对照只替换模型响应为确定的 Bash 指令，不替换 Eino、AIO 或 FinSafe。
探测均使用合成文件，不读取真实秘密。

| 用例 | 结果 |
| --- | --- |
| 包裹前：Eino 通过 Bash 执行 Python，工作区写入与计算 | 通过，合计 53 |
| 包裹前：写入唯一 `/var/lib/eino-compat-*` | 成功，证明原进程具备写权限 |
| 包裹后：相同脚本逻辑写入 `/var/lib/eino-compat-*` | 被拒绝，独立检查确认文件不存在 |
| 包裹后：拒绝发生后的合法工作区写入 | 成功，合计 53 |
| 实际 Eino PID 89 / OpenCode PID 88 | 各有 FinSafe cgroup、NoNewPrivs=1、Seccomp=2 |
| Eino 原有命令和 Skill 变量检查 | 保持开启 |
| OpenCode 回归：session/Bash API、写读、越界拒绝、拒绝后合法写入 | 通过，无模型调用 |
| 两个策略独立探测：工作区 `.env`、子进程写入、拒绝越界和策略覆盖 | 通过 |

基线与包裹后分别检查执行事件、脚本 SHA-256 和实际产物；不是仅依赖模型口头回答。
独立策略探测中的合成文件读取被拒绝，但未据此扩展承诺所有 `deny_read_paths`。

证据：

- [未包裹 Eino 基线](tests/evidence/eino-baseline-x64.json)
- [已包裹 Eino 对照](tests/evidence/eino-confined-x64.json)
- [OpenCode 回归](tests/evidence/opencode-eino-regression-x64.json)
- [两个策略独立检查](tests/evidence/eino-opencode-policy-smoke.json)

## 真实 DeepSeek Skill 验证

复用已启动主平台的模型配置，经 BFF `/callback/v1/model/1` 调用官方
`deepseek-flash`。回调路径使用数据库模型 ID，而非模型 UUID。
测试工作区只有占位 key，真实 key 留在平台模型代理中。

真实 `/chat` 事件流中观察到 `skill` 和 `bash` 调用，模型加载 `eino-compat`
Skill、执行预置 Python 脚本、读取结果文件；独立 AIO 检查确认合计为 **53**、
原脚本 SHA-256 未变，实际两个服务进程的 FinSafe 状态仍符合要求。
证据见 [eino-deepseek-x64.json](tests/evidence/eino-deepseek-x64.json)。

测试调试时曾把 UUID 用作模型 ID，平台返回 record not found；另有两轮因
模型加入 `cd`、stderr 捕获及结果查看而被过严的完整命令字符串断言拒绝。
测试现解析实际脚本调用的命令片段，并继续独立核对脚本哈希和产物。
上述调试失败未记作通过，也未修改 Eino 或关闭任何原有保护来通过测试。

## 覆盖边界

本次保护的是 OpenCode、`eino-agent` 本身和各自直接派生的进程树。
普通智能体的模型编排仍位于 `agent-service`；其 Skill 交给 Eino 时进入此次边界。
不能据此将整个智能体的所有工具或所有网络操作称为受保护。

独立 AIO shell/file API、远程 MCP、BFF、数据库未包裹；没有出站网络限制、
新的 PID 命名空间、每任务独立 cgroup 或会话间强隔离。多个会话共享服务进程树
和资源额度。未启动本体扩展，也未验证生产 Kubernetes 或客户技能库。

11 月 23 日完整平台的真实 OpenCode 界面任务见 [UI-VALIDATION.md](UI-VALIDATION.md)；
本次 Eino 测试直接调用真实执行服务和平台模型回调，不重复宣称完成新的界面验收。
