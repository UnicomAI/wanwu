# OpenClaw部署文档Windows版

## 快速安装

> 以管理员模式打开终端，确保能访问git和node依赖

### Windows PowerShell

```plaintext
iwr -useb https://openclaw.ai/install.ps1 | iex
```

### CMD

```plaintext
curl -fsSL https://openclaw.ai/install.cmd -o install.cmd && install.cmd && del install.cmd
```

## npm安装

### 第一步：安装依赖

以管理员模式打开 Power Shell终端，确保 Git 和 node已经安装并能成功下载到代码和依赖。

![image.png](assets/98a417c1-a91f-4f42-8a5f-64920491a76b.png)

![image.png](assets/1336699a-20d2-4807-9c02-d95dcba0032d.png)

### 第二步：安装openclaw

```plaintext
npm install -g openclaw@latest
```

![image.png](assets/bcb40522-600b-4361-bec5-d36160dcac0f.png)

## 配置openclaw

输入下述命令配置openclaw

```plaintext
openclaw onboard
```

![image.png](assets/142d559e-20b5-4c91-a532-5bcd10b79d29.png)

配置大模型

![image.png](assets/cfea1184-c716-4291-9eda-bbe3cc0a883c.png)

配置要接入的channel

![image.png](assets/628879e3-24eb-45be-9d62-900de138eb40.png)

选择要配置的skills

![image.png](assets/d3612323-9840-4ad3-9034-d78828c26f94.png)

安装成功后 在 Windows 浏览器中访问：[http://127.0.0.1:18789](http://127.0.0.1:18789)即可进入openclaw界面

![image.png](assets/0888bcdc-6cf8-4d23-a3bd-7a723f0afb28.png)

## 常见问题排查

### Q1: 报错 npm : 无法加载文件 C:\Program Files\nodejs\npm.ps1，因为在此系统上禁止运行脚本

![image.png](assets/fc486bf0-b887-40f2-912f-5afd4547b25a.png)

**原因**：执行策略/权限受限

**解决**：终端输入 `Set-ExecutionPolicy -Scope CurrentUser` 输入`RemoteSigned`命令给用户赋予权限

![image.png](assets/5286859c-a076-4702-9859-8fc7831974db.png)