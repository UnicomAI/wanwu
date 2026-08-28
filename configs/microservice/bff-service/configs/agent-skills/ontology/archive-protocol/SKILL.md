---
name: archive-protocol
version: "1.1.0"
description: 全局归档协议。只要任务需要写入任何文件（含 PLAN.md、报告、JSON 等归档物），必须按本技能执行 Session→ARCHIVE_ID、TIMESTAMP、双轨路径（根段须为 archives/）、回读校验与状态回执；WebUI 的 archive_grid 必须用 Markdown 中语言标识为 json 的围栏代码块输出。
---

# 全局归档协议

只要当前任务需要写入任何文件，就必须执行本协议。

归档物包括但不限于：`PLAN.md`、`report.md`、`summary.md`、`result.json`、`notes.md`。

## 【身份文件】

若任务需要人格/身份设定，只认：

- `SOUL.md`
- `IDENTITY.md`

未确认可读或已注入上下文前，不得声称「已成功读取」。

# 归档协议 (Archive Protocol)

## 【核心规则：必须全量归档】

**所有生成的成果性内容（无论是单个文件还是整个目录/文件夹），在任务完成后必须彻底搬移（Move）至归档区。**

工作区（Workspace）仅作为临时加工场，任务结束后的合规状态是：**工作区不留任何生成物，全部进入 `archives/`。**

## 【绝对路径硬性边界】

**所有文件（包括工作区临时文件、PLAN.md、归档物、报告、JSON 等）的保存路径必须落在以下根目录之下，禁止越界：**

```
/home/root/workspace/{WORKSPACE_ID}/workspace/
```

- `{WORKSPACE_ID}` 由当前会话上下文给定，由 `archive` 工具拼接，**不得手填、不得使用其它前缀**。
- 工作区与归档区都必须以该绝对前缀开头：
  - 工作区临时文件：`/home/root/workspace/{WORKSPACE_ID}/workspace/<相对路径>`
  - 归档区文件：`/home/root/workspace/{WORKSPACE_ID}/workspace/archives/{ARCHIVE_ID}/...`
- 任何写入到该前缀**之外**的路径（如 `/tmp/...`、`/root/...`、`./xxx`、其它绝对路径）都视为越界，必须立即终止并报错；禁止用软链接、相对路径 `..` 跳出该根。
- `archive` 工具返回卡片中的 `archive_root` 必须以 `/home/root/workspace/` 开头；不满足则视为执行异常。

## 【归档路径规则】

路径执行双轨制，严禁混淆，并且**只能**由 `archive` 工具写入。**注意：归档操作会将文件或目录从工作区物理搬移（Move）至归档区，原位置将不再存在。**

1. **计划文件（PLAN.md 专属）**：工具参数 `{"kind":"plan","sourcePath":"PLAN.md"}`。归档后，如需修改计划，应**直接编辑** `archives/{ARCHIVE_ID}/PLAN.md`（绝对路径：`/home/root/workspace/{WORKSPACE_ID}/workspace/archives/{ARCHIVE_ID}/PLAN.md`）。
2. **普通归档物（单个文件或整个目录）**：工具参数 `{"kind":"file","sourcePath":"result_dir"}`。支持归档单个文件或包含多个文件的目录。

## 【目录归档模式】

当你的输出包含多个文件（如前端项目、代码包、多份分析报告）时，**必须执行以下流程**：
1. 在工作区创建一个专用文件夹（如 `output/`），位置必须在 `/home/root/workspace/{WORKSPACE_ID}/workspace/` 之下。
2. 将所有相关文件放入该文件夹。
3. **调用 `archive` 工具对整个文件夹进行归档**（`sourcePath` 指向文件夹路径）。
4. 归档成功后，整个文件夹及其内容将移入 `archives/{ARCHIVE_ID}/{TIMESTAMP}/`（绝对路径：`/home/root/workspace/{WORKSPACE_ID}/workspace/archives/{ARCHIVE_ID}/{TIMESTAMP}/`）。

## 【写入后强制校验】

任何归档文件写入后，必须立即回读校验。未完成校验前，不得声称「已成功归档」。调度 `archive` 工具会在写入后自动回读并输出状态卡片。

## 输出
请在最终输出的结尾，把archive工具的输出卡片也输出出来。必须以markdown的json代码块输出。

## 【工具执行指引】

**所有归档物生成都必须通过 `archive` 工具** 完成；禁止直接写 `archives/` 目录。推荐流程：

1. **准备源文件**：在工作区临时目录编写内容。
2. **调用工具**（工具会自动解析 session，生成 `TIMESTAMP`）：
   - **单个文件归档**：
     ```json
     {
       "name": "archive",
       "arguments": {
         "kind": "file",
         "sourcePath": "result.json"
       }
     }
     ```
   - **整个目录归档（推荐用于多文件结果）**：
     ```json
     {
       "name": "archive",
       "arguments": {
         "kind": "file",
         "sourcePath": "my_results_folder"
       }
     }
     ```
3. **后续操作**：归档成功后，工作区原文件/目录已删除。若需再次查看或编辑，**必须使用工具返回的 JSON 卡片中的 `subpath` 或 `archive_root` 拼接出的完整路径** 访问归档区文件。
4. **异常处理**：工具返回错误提示（如 `Archive blocked`）时须立刻停止归档流程，引用错误原因告知用户并等待下一步指示；禁止在失败后自行补写文件或回执。
