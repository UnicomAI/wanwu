# 分两阶段验证

这些文件只用于测试，不修改或替换万悟业务实现。测试结果及已覆盖范围见
[../VALIDATION.md](../VALIDATION.md)。

## 原生 Linux x64 Docker 主机

准备与目标部署相同版本的完整 FinSafe Linux x64 运行包、OpenCode 可执行文件，
并核对发行方的 SHA-256。正式用例使用 FinSafe 0.11.19、OpenCode 1.19.5。
FinSafe 下载/校验方式也可以复用 `../Dockerfile` 的 `finsafe-runtime` 构建阶段。

使用绝对路径配置以下环境变量。cgroup 委派的建立方式见上一级部署说明；
路径位于 Docker daemon 所在主机。

```bash
export FINSAFE_RUNTIME_DIR=/absolute/path/to/finsafe-runtime
export OPENCODE_BINARY=/absolute/path/to/opencode
export FINSAFE_CGROUP_ROOT=/sys/fs/cgroup/finsafe-wanwu

# 第一阶段：裸 OpenCode 与 FinSafe 包裹的 OpenCode 跑相同用例。
sh deploy/finsafe/tests/run.sh compatibility
```

用例真实调用 OpenCode 的健康接口、创建会话、Bash 工具 API、SSE 事件流和删除会话。
通过外部检查文件是否存在来验证写入结果，避免把工具返回的 completed 或 exit code
误当作越界写入被拒绝的证明。包裹后还检查 Supervisor 管理的实际 OpenCode PID
是否进入 FinSafe cgroup，以及 NoNewPrivs 和 seccomp 是否生效。

第一阶段通过后，构建并运行第二阶段驱动。可复用本机 Go 缓存；以下构建只读挂载源码：

```bash
mkdir -p /tmp/wanwu-finsafe-driver
docker run --rm -e CGO_ENABLED=0 -e GOOS=linux -e GOARCH=amd64 \
  -e GOPROXY=https://goproxy.cn,direct \
  -v "$PWD:/workspace:ro" -w /workspace \
  -v wanwu-finsafe-go-mod:/go/pkg/mod \
  -v wanwu-finsafe-go-cache:/root/.cache/go-build \
  -v /tmp/wanwu-finsafe-driver:/out golang:1.24.13-bookworm \
  go build -mod=readonly -o /out/wanwu-integration ./deploy/finsafe/tests/wanwu

export WANWU_INTEGRATION_BINARY=/tmp/wanwu-finsafe-driver/wanwu-integration
sh deploy/finsafe/tests/run.sh wanwu
```

驱动调用仓库中的真实 `pkg/wga-sandbox.Run`：准备工作区、生成 OpenCode 模型配置、
上传输入文件、创建会话、订阅并转换 SSE、触发模型/工具循环、下载产物、清理沙箱。
OpenCode 及 FinSafe 均为真实进程。

`services.py` 是测试替身，提供本地 OpenAI-compatible 模型响应和 AIO 的 shell/file
传输协议。模型返回确定的 Bash tool call，由真实 OpenCode 执行，再返回最终文本。
检查上传文件经工具处理后回传的内容，并验证越界写入失败。
这覆盖万悟 Go 适配路径，不等于完整 BFF/UI、真实 AIO 镜像或客户模型的验收。

脚本不会暴露端口到宿主机，并会删除本次测试容器。不要把测试服务部署到生产。

## ARM Docker 主机上的 x64 内核测试

直接通过 Rosetta/QEMU 用户态模拟运行 x64 容器，可能不支持 Landlock。
不能将这种失败认定为原生 Linux x64 不兼容。本次使用 Docker 承载 QEMU 系统模拟，
运行完整 x64 Linux 内核、Ubuntu 24.04 用户态和上述相同 x64 发行包。

可重复建立此环境的步骤：

1. 构建 `tests/Dockerfile` 的 amd64 镜像及 `tests/Qemu.Dockerfile` 的本机架构镜像。
2. 用 `docker create`、`docker export` 将 amd64 测试镜像导出到新建 Docker volume。
3. 将完整 FinSafe 包复制到 rootfs 的 `/opt/finsafe`，OpenCode 复制到
   `/root/.opencode/bin/opencode`；将 `compatibility.py`、`opencode-baseline.conf`、
   `model-config.json` 和正式的 `../opencode.conf`（改名 `opencode-confined.conf`）复制到 `/tests`。
4. 将 `qemu-init.sh` 放到 `/init` 并赋予执行权限，建立 `/dev/console` 字符设备
   （major 5、minor 1）。把 rootfs 用 `cpio --format=newc` 和 gzip 打包为 initramfs。
5. 第二阶段还需将编译好的 x64 `wanwu-integration` 驱动及 `services.py` 复制到 `/tests`。
6. 使用官方 x64 Linux 内核启动 QEMU，保留内核版本、SHA-256 和完整控制台记录。

本次内核来自 [Debian 官方 netboot 镜像](https://deb.debian.org/debian/dists/trixie/main/installer-amd64/current/images/netboot/debian-installer/amd64/)，
其校验值在验证记录中；不把这个结果表述为 Ubuntu 原厂内核或 CentOS 的验收。
下面是 Docker 中实际使用的 QEMU 参数（`/fixture` 挂载下载的内核及打包的 initrd）：

```bash
qemu-system-x86_64 -machine q35 -accel tcg,thread=multi -cpu max \
  -smp 2 -m 4096 -kernel /fixture/linux -initrd /fixture/initrd.gz \
  -append 'console=ttyS0 rdinit=/init lsm=landlock quiet panic=-1' \
  -nographic -no-reboot -net none
```

所有 API 调用在 guest loopback 内完成。没有网络外联、真实模型费用或生产数据。
以 `WANWU_COMPATIBILITY_X64=PASS`、`WANWU_INTEGRATION_X64=PASS` 及 `WANWU_VM_EXIT=0`
为成功标志；只看 QEMU 进程退出码不足以证明测试通过。

## 完整派生镜像与真实 DeepSeek

`full-docker-init.sh` 是另一个测试环境的 PID 1：启动真实 x64 Docker daemon，
从第二块 virtio 只读磁盘导入 `docker save` 生成的完整派生镜像，运行其原始
AIO 入口及 Supervisor。第一块 ext4 virtio 磁盘保存 Docker 数据，使用 overlay2；
需要额外装入匹配内核的 e1000、virtio_blk、overlay、ext4 及依赖模块，以及
x64 静态 Docker 工具。此环境与上面的无网络 fixture 测试独立。
因为测试主机从 initramfs 启动，脚本为 dockerd 设置 `DOCKER_RAMDISK=1`；
这只适配测试主机的容器根目录切换，不改变 FinSafe 策略或允许 fallback。

`live_gateway.py` 在 QEMU 外层容器的 loopback 8000 端口运行，读取只读挂载的
本机 key 文件，将请求转发到官方 DeepSeek API，返回真实响应。它不模拟模型、
不记录 key 或消息体；guest 中只使用无效的占位凭据，通过 `10.0.2.2:8000/v1`
访问该转发器。代理可通过 `--proxy` 指定，默认使用本机已有的 1087 HTTP 代理。
每轮最多 12 次请求，每次最多 2048 输出 token；会产生真实模型用量。

Go 驱动设置 `WANWU_LIVE_MODEL=1` 后使用自然语言任务，要求真实模型执行指定的
Bash 脚本。断言包含：SSE 中实际观察到携带唯一越界路径的 Bash 命令、回传文件
内容正确、拒绝标记正确，以及从独立 `docker exec` 检查越界文件未产生。
这避免只依赖模型最终回答来判断约束是否生效。

完整测试的成功标志为 `WANWU_FULL_IMAGE_LIVE_MODEL=PASS`；仅导入镜像或服务启动
不算通过。测试不启动 BFF/UI/数据库，也不覆盖独立 AIO API 的权限边界。

## Eino Skill 执行端增量验证

`eino_compatibility.py` 连接完整镜像中真实的 Eino `/chat` 与 AIO shell API，
不替换 Eino 服务。完整结果见 [EINO-VALIDATION.md](../EINO-VALIDATION.md)。
先在只包裹 OpenCode 的旧镜像运行 baseline，再切换到同时包裹两个运行时的新镜像
运行 confined；两个模式使用相同逻辑的脚本，唯一探测路径每次随机生成。

边界对照需要在执行容器可解析的 Docker 网络中启动模型 fixture，不发布宿主端口：

```bash
docker run -d --name finsafe-eino-fixture --network wanwu-net \
  --mount type=bind,src="$PWD/deploy/finsafe/tests/eino_compatibility.py",dst=/test.py,readonly \
  --entrypoint python3 python:3.12-bookworm /test.py --serve-fixture

# 在可以访问真实 AIO/Eino 的 Python 客户端容器中运行；替换 URL 为实际地址。
python3 eino_compatibility.py --mode confined \
  --aio-url http://sandbox:8080 --eino-url http://sandbox:4097 \
  --evidence /tmp/eino-confined.json
```

fixture 只控制模型返回的 Bash 指令，实际脚本仍由真实 Eino 执行。
独立 AIO 检查工作区产物、原脚本 SHA-256 和越界文件是否存在。
不关闭 Eino 的原有命令/Skill 变量保护；不读取真实凭据。
`baseline` 期望越界写入成功，`confined` 期望拒绝且合法写入继续成功。

真实模型验证不使用 fixture，复用平台已配置的模型回调（这里的 ID 是数据库
模型 ID，不是 UUID），使用前确认模型名为 `deepseek-flash`：

```bash
python3 eino_compatibility.py --mode live \
  --aio-url http://sandbox:8080 --eino-url http://sandbox:4097 \
  --live-model-url http://bff-service:6668/callback/v1/model/1 \
  --evidence /tmp/eino-deepseek.json

python3 eino_compatibility.py --opencode-regression \
  --aio-url http://sandbox:8080 --opencode-url http://sandbox:4096 \
  --evidence /tmp/opencode-regression.json

docker rm -f finsafe-eino-fixture
```

live 模式会产生真实 DeepSeek 用量，要求实际 Skill+Bash 调用和计算结果 53。
测试工作区只含占位 key，真实凭据由现有平台模型代理使用。
OpenCode 回归调用真实 session/Bash API，不发起模型请求。
默认 URL 为本次 Mac Docker/QEMU 测试环境的 TCP 映射；原生 Linux 请显式传入 URL。
这些用例不宣称覆盖远程 MCP、独立 AIO 或整个 `agent-service` 的安全边界。
