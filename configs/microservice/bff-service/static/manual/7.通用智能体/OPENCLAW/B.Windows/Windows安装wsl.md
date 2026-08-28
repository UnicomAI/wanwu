# Windows安装wsl

## 一、前置条件

| 项目 | 要求 |
| --- | --- |
| 操作系统 | Windows 10 版本 2004（Build 19041 或更高）<br>或 Windows 11 |
| 架构 | 64 位（x64）或 ARM64（部分支持） |
| 虚拟化 | BIOS/UEFI 中启用 虚拟化（Intel VT-x / AMD-V） |
| 磁盘空间 | 至少 5 GB 可用空间（推荐 10 GB+） |

打开开始菜单，在开始菜单中输入 `启用或关闭 Windows 功能`，在弹出的窗口中勾选`Hyper-V`、`虚拟机平台（Virtual Machine Platform)` 和 `适用于 Linux 的 Windows 子系统`，确定之后重启系统。

![image.png](assets/6d75c04d-75b8-4132-898d-4110a2ec70ee.png)

## 二、快速安装WSL

右键单击并选择“以 **管理员** 身份运行”，在管理员模式下打开 PowerShell，输入 wsl --install 命令，然后重新启动计算机。

```shell
wsl --install
```

可选参数

*   `--no-distribution`：安装 WSL 时不要安装linux发行版。
    
*   `--distribution`：指定要安装的 Linux 发行版。 
    
*   `--no-launch`：安装 Linux 发行版，但是不自动启动它。
    
*   `--web-download`：从联机源安装，而不是使用 Microsoft Store。
    
*   `--location`：指定要将 WSL 发行版安装到哪个文件夹。
    

安装过程

![image.png](assets/cfec2100-8fa7-4d5a-aab0-c29254462442.png)

### 更改安装的默认 Linux 发行版

默认情况下，已安装的 Linux 发行版将为 Ubuntu。 可以通过使用`-d`标志来更改这一点。

*   若要更改安装的发行版，请输入：
    
    ```plaintext
    wsl --install [Distro]
    ```
    
    将 `[Distro]` 替换为您想要安装的发行版名称，例如：Ubuntu
    
*   若要查看可通过在线商店下载的可用 Linux 发行版列表，请输入：
    

    ```plaintext
    wsl --list --online
    ```

若要安装未列为可用的Linux发行版，可以使用 .tar文件[导入任何 Linux 发行版](https://learn.microsoft.com/zh-cn/windows/wsl/use-custom-distro) 。 

## 三、分步安装WSL

### 下载 Linux 内核更新包：

*   [WSL2 Linux 内核更新包适用于 x64 计算机](https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_x64.msi)
    
*   [适用于 ARM64 计算机的 WSL2 Linux 内核更新包](https://wslstorestorage.blob.core.windows.net/wslblob/wsl_update_arm64.msi)
    

### 将 WSL 2 设置为默认版本

以管理员模式打开 PowerShell 并运行以下命令，在安装新的 Linux 发行版时将 WSL 2 设置为默认版本：

```plaintext
wsl --set-default-version 2
```

### 安装所选 Linux 发行版

打开 Microsoft Store 并选择你喜欢的 Linux 发行版。

![image.png](assets/3b3aafe5-0ea9-4e42-abec-e43bcd46b504.png)

## 四、启动WSL

在搜索栏通过搜索对应的linux发行版名称可以直接启动

![image.png](assets/754d10df-4a91-4080-961d-39cc98d0dfec.png)