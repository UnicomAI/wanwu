# 8GB 主平台低内存验证

日期：2026-11-24，Asia/Shanghai。目标是保留主平台组件，以较小的堆、缓存、
worker 数和后台并发运行；本体扩展维持关闭，模型使用外部官方 DeepSeek。
业务 Go/Python 源码未修改，沿用 `wanwu-wga-finsafe:0.11.19` 镜像。

## 最终配置

- Kafka：实际 JVM 参数 `-Xms128m -Xmx256m`；修正原配置变量未作用到 JVM 的问题。
- Elasticsearch：`-Xms384m -Xmx384m`，容器 1024MiB。
- MySQL：64MiB buffer pool、50 连接、256MiB 容器额度；关闭可选的 performance_schema。
- RAG：四个 gunicorn 入口各 1 worker，普通和图谱异步入库各 1 个进程，合计额度 2GiB。
- callback：1 worker、2 线程、192MiB；Go 服务设置 GC 目标及独立容器额度。
- 所有常驻服务的 RAM+swap 上限与 RAM 上限相等；实际沙箱共同父组也不得使用 swap。
- AIO、OpenCode、Eino 和工具进程放在同一 1536MiB 父组下，避免 FinSafe 将进程
  移出 Docker cgroup 后漏计资源。两个运行时原有 FinSafe 约束继续生效。
- 常驻额度合计 7296MiB，约 7.13GiB；不表示预分配内存。配置见
  [configuration.json](evidence/configuration.json)。

完整性检查必须枚举全部 21 个主平台服务，不能只枚举正在运行的容器；
后者曾漏掉数据库重启后处于 created 状态的 app-service 和 operate-service。
已补齐两者，并将缺失/未运行任意主服务作为采样失败条件。

单纯压低容器上限会把部分匿名内存换到磁盘。早期采样虽显示约 5.49GiB RAM，
仍有约 641MiB swap，记录保留在
[memory-samples-initial-swapping.json](evidence/memory-samples-initial-swapping.json)，
不将这项读数当作最终无换页结果。MySQL 的较大缓存组合曾在无 swap 的 256MiB
额度下退出 137；最终改用 64MiB pool，并验证空库初始化。

## 实际测试环境

Mac ARM64 的 Docker daemon 内存额度约 7.75GiB。主平台其他组件使用原有 ARM
镜像，实际 FinSafe 沙箱使用 Docker 承载的完整 x64 QEMU 内核环境：
Ubuntu 22.04 沙箱、Debian 6.12.117+deb13-amd64 内核、Landlock ABI 6、cgroup v2、seccomp。
这不是一台原生 8GB Linux x64 物理服务器的验收。

QEMU guest 内存 2560MiB；测试外层容器上限 4GiB、swap 禁用。
4GiB 包含模拟器开销，不是原生部署额外需要的服务。
实际沙箱内的共同父 cgroup 上限独立为 1536MiB。
不能将 Docker stats 的容器一行当作整个 FinSafe 进程树内存；实际还读取共同父组
的 memory.current、memory.peak、memory.max 和 OOM 计数。

## 功能检查

| 检查 | 已观察结果 |
| --- | --- |
| MySQL 同版本、同镜像、全新独立数据卷初始化 | 256MiB、无 swap 成功，初始化后有 14 个 schema |
| RAG 六个原入口 | 8681、10891、20041、20050、8613、15000 均监听 |
| RAG 实际文本切分 API | 返回正确的 2 个文本块 |
| 真实 DeepSeek → Eino Skill → Bash → Python | 计算结果 53，原脚本 SHA-256 未变 |
| OpenCode 真实 session/Bash API 回归 | 写读成功、越界写入拒绝、拒绝后合法写入成功 |
| 实际 OpenCode / Eino 进程 | FinSafe cgroup、NoNewPrivs=1、Seccomp=2，Eino 原有检查开启 |

独立数据库测试没有使用生产数据卷，未暴露端口；临时容器及其专用数据卷已清理。
真实模型测试的工作区只有占位 key，由原平台 BFF 模型回调使用既有 DeepSeek 配置。
本次直接验证执行服务及模型回调，未重复声称完成新的 UI 全流程验收。

证据：

- [MySQL 空库初始化](evidence/mysql-fresh-init.json)
- [RAG 入口和实际文本处理](evidence/rag-text.json)
- [真实 DeepSeek Skill](evidence/eino-deepseek.json)
- [OpenCode 回归](evidence/opencode-regression.json)

## 适用范围

仅验证轻任务和有限观测时段，没有进行多用户压力测试、大文件解析、OCR、
知识库批量入库或全部工作流节点验收。worker 降低会减少处理并发；遇到大任务应
提高对应组件额度。8GB 配置不等于原 32GB 推荐环境的容量，也不承诺 4GB 完整运行。

cgroupfs 路径在本测试 guest 中实际验证；systemd helper 按 Docker/systemd 的父组
和委派机制编写，仍需在目标服务器复验。FinSafe 既有 glibc 弱版本警告、独立 AIO/MCP
覆盖范围和会话隔离限制仍见 [FinSafe 部署说明](../finsafe/README.md)。

## 最终完整主平台采样

- 明确枚举并验证全部 **21 个主平台服务**，另计 1 个 QEMU 测试容器。
- Docker daemon 内存额度为 7.748GiB。
- 3 次短时采样中，主平台及 QEMU 合计 Docker stats 峰值为
  **6.150GiB**；包含模拟开销，不是整机或长期负载峰值保证。
- 实际沙箱共同父组的内核 memory.peak 为 **1536.0MiB**，上限 1536MiB。
- 所有 22 个容器的最终 swap 为 0，cgroup oom_kill 为 0；观测期间容器重启
  计数未增长，设置了健康检查的运行服务全部健康。
- 在完整主平台均运行时，再次通过真实 DeepSeek Skill＋Bash＋Python 任务。

原始结果见 [memory-samples.json](evidence/memory-samples.json)。
此前漏计两个服务的样本保留为 `memory-samples-incomplete-services.json`，
不作为完整主平台的验收结果。
