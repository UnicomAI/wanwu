# OpenClaw在wsl下面部署文档

---

本文档详细介绍了如何在 Windows WSL (Windows Subsystem for Linux) 环境中安装、配置并验证 OpenClaw。

---

## 前置条件

在开始安装之前，请确保你的 Windows WSL 环境满足以下所有要求。这些条件是 OpenClaw 正常运行的基石。

| 条件 | 要求版本/状态 | 说明 |
| --- | --- | --- |
| **操作系统** | Windows 10/11 + WSL2 | 建议使用 WSL2 以获得更好的网络兼容性。 |
| **Node.js** | **\>= 22** | OpenClaw 强制要求 Node 22 以上版本。WSL 默认源版本过低，**必须使用 NVM 安装**。 |
| **Git** | 已安装 | npm 在安装依赖包时需要调用 git。如果未安装会报 `ENOENT` 错误。 |
| **构建工具** | build-essential | 包含 GCC/G++ 编译器，用于编译部分 C++ 依赖。 |

---

## 安装步骤

### 第一步：更新系统并安装基础依赖

打开 WSL 终端（推荐 Ubuntu 20.04 或 22.04），执行以下命令来更新软件源并安装 Git 和构建工具。

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y git build-essential curl
```

**注意**：如果你之前手动卸载过 git，这一步必须重新执行。

### 第二步：安装 Node.js (使用 NVM)

由于 OpenClaw 需要 Node.js >= 22，而 `apt` 源中的版本通常较老，我们使用 `nvm` (Node Version Manager) 来安装最新版。

1.  **安装 NVM**：
    
    ```bash
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
    ```
    
    注意：如果下载安装失败，先执行以下命令：
    
    ```bash
    git config --global url."https://github.com".insteadOf ssh://git@github.com
    git config --global url."https://github.com".insteadOf git@github.com
    ```
    
2.  **激活 NVM 并安装 Node 22**：
    
    ```bash
    # 重新加载环境变量（或关闭终端重新打开）
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    # 安装 Node.js 22 最新版
    nvm install 22
    # 设置为默认版本
    nvm alias default 22
    ```
    
3.  **验证版本**：
    

```bash
node -v
# 输出应为 v22.x.x
```

### 第三步：启用 WSL Systemd 支持

为了让 OpenClaw 能够作为后台服务运行，需要开启 WSL 的 systemd 支持。

1.  **编辑 WSL 配置文件**：
    
    ```bash
    sudo vim /etc/wsl.conf
    ```
    
2.  **添加以下内容**：
    
    ```properties
    [boot]
    systemd=true
    ```
    
    _(保存：按_ `Ctrl+O`_，回车；退出：按_ `Ctrl+X`_)_
    
3.  **重启 WSL**：  
    在 Windows 的 PowerShell 或 CMD 中执行：
    

```powershell
wsl --shutdown
```

然后重新打开 WSL 终端。

### 第四步：全局安装 OpenClaw

确保 Node 环境正确后，使用 npm 安装 OpenClaw。由于使用了 nvm，无需加 `sudo`。

```bash
npm install -g openclaw@latest
```

注意：如果安装失败，执行以下命令：

```bash
npm config set registry https://registry.npmmirror.com
```

### 第五步：初始化配置

安装完成后，运行初始化向导。这将引导你配置模型 API Key、连接渠道（如 WhatsApp/Telegram）并注册系统服务。

```bash
openclaw onboard --install-daemon
```

按照提示完成配置：

配置 ~/.openclaw/openclaw.json

可以参照[《openclaw 配置元景模型样例》](https://alidocs.dingtalk.com/i/nodes/G1DKw2zgV2R0PxMkIDboGBjMVB5r9YAn?doc_type=wiki_doc)配置。

---

## 验证安装

完成安装后，请按照以下步骤验证 OpenClaw 是否正常工作。

### 1. 验证 CLI 命令

检查 OpenClaw 命令是否可用：

```bash
openclaw --version
```

_预期结果：输出版本号，如_ `openclaw/x.x.x ...`

### 2. 启动网关并访问 Web UI

如果服务未自动启动网关，可以手动启动：

```bash
openclaw gateway
```

或者

```bash
nohup openclaw gateway --port 18789 > openclaw.log 2>&1 &
```

_预期结果：终端输出显示网关已启动，监听在_ `0.0.0.0:18789` _或类似端口。_

**浏览器验证**：

在 Windows 浏览器中访问：

`http://127.0.0.1:18789` 或 `http://localhost:18789`

_预期结果：成功打开 OpenClaw 的 Web 控制面板界面。_

### 1. 关闭openclaw

pkill -f "openclaw gateway"

---

## 常见问题排查

### Q1: 报错 `npm error code ENOENT npm error syscall spawn git`

**原因**：系统未安装 Git，或 Git 路径未配置。  
**解决**：执行 `sudo apt install git`，然后重新运行 `npm install -g openclaw@latest`。

### Q2: 报错 `EBADENGINE Unsupported engine`

**原因**：Node.js 版本低于 22。  
**解决**：使用 `nvm install 22` 重新安装并切换版本。

### Q3: 浏览器访问 localhost 拒绝连接

**原因**：网关未启动，或防火墙拦截。  
**解决**：

1.  确认 WSL 终端中 `openclaw gateway` 正在运行。
    
2.  尝试使用 `http://127.0.0.1:18789` 而非 localhost。
    

### Q4： 网页聊天/仪表盘界面无法连接到网关 websocket，出现错误： `unauthorized conn= “”   remote=127.0.0.1 client=openclaw-control-ui webchat vdev reason=token\_missing`

![image.png](assets/b159c0ab-f666-4b53-8478-d550d979344d.png)

解决：

1.  cat ~/.openclaw/openclaw.json
    
2.  拷贝gateway.auth.token
    
3.  粘贴到  [http://127.0.0.1:18789/](http://127.0.0.1:18789/) \>overview > Gateway Access > Gateway Token
    

![image.png](assets/b37e6e68-737c-4f78-a78d-10179f22986b.png)

![image.png](assets/c71e5535-fb41-43b1-8429-48a556b2ce54.png)

参考链接：[https://github.com/openclaw/openclaw/issues/1690](https://github.com/openclaw/openclaw/issues/1690)