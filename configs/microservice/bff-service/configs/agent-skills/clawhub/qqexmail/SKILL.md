---
title: 腾讯企业邮箱
description: 通过 IMAP/SMTP 收发腾讯企业邮箱（exmail.qq.com）邮件。支持发送邮件、收取邮件列表、获取邮件正文。凭证从环境变量读取。
---

## 何时使用

用户要使用 **腾讯企业邮箱**（exmail.qq.com）**发邮件**、**收邮件**、**查邮件**时使用本 skill。

## 凭证（环境变量）

| 变量 | 说明 |
|------|------|
| **EXMAIL_ACCOUNT** | 腾讯企业邮箱账号（完整地址，如 name@company.com） |
| **EXMAIL_AUTH_CODE** | 腾讯企业邮箱授权码（在腾讯企业邮箱「设置 → 账户 → 账户安全」中生成，**非邮箱登录密码**；勿提交到仓库） |

脚本会校验，缺失时报错并退出。

## 腾讯企业邮箱服务器

- **IMAP**：`imap.exmail.qq.com`，端口 993（SSL）
- **SMTP**：`smtp.exmail.qq.com`，端口 465（SSL）

## 脚本

| 脚本 | 作用 |
| --- | --- |
| `scripts/send.js` | 从环境变量读凭证，用 nodemailer 连接腾讯企业邮箱 SMTP 发信；支持收件人、主题、正文（CLI 参数）。 |
| `scripts/receive.js` | 从环境变量读凭证，用 imap + mailparser 连接腾讯企业邮箱 IMAP 收信；支持「最近 N 条」或「最近 N 天」，输出主题、发件人、日期、**UID**、正文摘要。 |
| `scripts/get-body.js` | 按 **UID** 获取指定邮件的**完整正文**（纯文本，无摘要截断）。必须传入 `--uid`（值为收信列表中的 UID）。 |

## 发信流程

在 skill 根目录下执行（需已 `npm install`）：

```bash
node scripts/send.js <收件人> <主题> <正文>
```

正文若含空格，请用引号包裹；或只传收件人和主题，正文从 stdin 读入（见脚本 `--stdin`）。

**示例**：

```bash
node scripts/send.js "recipient@example.com" "测试主题" "邮件正文内容"
```

## 收信流程

```bash
# 收取最近 10 条（默认）
node scripts/receive.js

# 收取最近 N 条
node scripts/receive.js --limit 20

# 收取最近 N 天的邮件（如 7、30、90）
node scripts/receive.js --days 7
```

输出：每封邮件的主题、发件人、日期、**UID**（收件箱内唯一标识，用于按 UID 取正文）、正文摘要（前约 200 字），便于查看。

## 获取邮件正文

需要某封邮件的**完整正文**时，使用 `get-body.js`，传入收信列表中该邮件的 **UID**：

```bash
node scripts/get-body.js --uid 12345
```

未传 `--uid` 时会提示并退出。UID 与收件箱绑定，邮件移动或删除后可能失效。

- **输出**：完整正文输出到 stdout（纯文本；若原邮件仅有 HTML，会做简单去标签后输出）。可重定向到文件或管道给其它命令。
- **环境变量**：与收信相同，需 `EXMAIL_ACCOUNT`、`EXMAIL_AUTH_CODE`。

## 可选能力

- **收取时间范围**：通过 `--days 7` / `--days 30` / `--days 90` 使用 IMAP SINCE 条件。
- **收取「我的文件夹」**：当前脚本默认 INBOX；若需自定义文件夹，可扩展脚本中的 `openBox`。

## 安全提醒

- 腾讯企业邮箱授权码需在「设置 → 客户端设置」中开启 IMAP/SMTP 服务后生成，与邮箱登录密码不同，不要混淆。
- 不要将 `EXMAIL_ACCOUNT`、`EXMAIL_AUTH_CODE` 的真实值写入代码或提交到仓库；仅通过环境变量或本地 `.env` 配置。
