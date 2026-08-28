---
name: wechat-article-publisher
description: 从 Markdown 文件或网页链接提取文章并发布到微信公众号。适用于需要“自动创建草稿/提交发布”、统一微信样式（standard/viral）和批量复用发布流程的场景。
---

# WeChat Article Publisher

## 何时使用

当用户明确要求以下任一任务时触发本 skill：

- 把本地 `.md` 文章发布到微信公众号
- 把网页链接（博客/新闻页）提取后发布到微信公众号
- 需要统一微信排版模板（`standard` 或 `viral`）并自动生成封面
- 需要程序化输出 `media_id`、`publish_id`、发布状态

## 工作流

1. 准备配置：
   - 编辑本目录 `config.json`，仅需填写：`wechat.app_id`、`wechat.app_secret`、`wechat.author`
2. 安装依赖：
   - `python scripts/publish_wechat.py --install`
3. 预览（不调用微信接口）：
   - `python scripts/publish_wechat.py <输入> --dry-run`
4. 创建草稿：
   - `python scripts/publish_wechat.py <输入>`
5. 直接提交发布（可选）：
   - `python scripts/publish_wechat.py <输入> --publish --status`

`<输入>` 支持：

- 本地 Markdown 文件路径
- `http://` / `https://` 网页 URL

## 命令参数

主脚本：`scripts/publish_wechat.py`

- `input`：Markdown 文件路径或网页 URL
- `--config`：配置文件路径，默认本 skill 的 `config.json`
- `--template`：覆盖模板，`standard|viral`
- `--author`：覆盖作者
- `--source-url`：覆盖原文链接
- `--cover-image`：指定本地封面图
- `--dry-run`：仅提取+渲染，不调微信 API
- `--publish`：草稿创建后调用 `freepublish/submit`
- `--status`：提交发布后查询一次发布状态

## 执行约束

- 发布前优先做 `--dry-run`，检查标题、摘要和渲染 HTML。
- 如果账号无 `freepublish` 权限，`--publish` 可能返回 `48001`，此时保留草稿手动发布。
- 若创建草稿时报 `41005 media data missing`，请通过 `--cover-image` 指定封面图。

## 输出结果

脚本标准输出 JSON，关键字段：

- `draft_media_id`
- `publish_id`（仅 `--publish`）
- `status`（仅 `--status`）
- `preview_html`（仅 `--dry-run`）
