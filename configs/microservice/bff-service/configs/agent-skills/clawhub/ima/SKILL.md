---
name: ima-note
description: |
  IMA 个人笔记服务 API skill，用于管理用户的 IMA 笔记。支持搜索笔记、浏览笔记本、获取笔记内容、新建笔记和追加内容。
  当用户提到笔记、备忘录、记事、知识库，或者想要查找、阅读、创建、编辑笔记内容时，使用此 skill。
  即使用户没有明确说"笔记"，只要意图涉及个人文档的存取（如"帮我记一下"、"我之前写过一个关于XX的东西"、"把这段内容保存下来"），也应触发此 skill。
homepage: https://ima.qq.com
metadata:
  {
    "openclaw":
      {
        "emoji": "📝",
        "requires": { "env": ["IMA_OPENAPI_CLIENTID", "IMA_OPENAPI_APIKEY"] },
        "primaryEnv": "IMA_OPENAPI_CLIENTID"
      },
  }
---

# ima-note

通过 IMA OpenAPI 管理用户个人笔记，支持读取（搜索、列表、获取内容）和写入（新建、追加）。

完整的数据结构和接口参数详见 `references/api.md`。

## Setup

1. 请打开 https://ima.qq.com/agent-interface 获取 **Client ID** 和 **Api Key**
2. 配置环境变量：

```bash
export IMA_OPENAPI_CLIENTID="your_client_id"
export IMA_OPENAPI_APIKEY="your_api_key"
```

> 建议将上述 export 语句写入 `~/.zshrc` 或 `~/.bashrc`，避免每次重开终端失效。

## 凭证预检

每次调用 API 前，先确认凭证可用。如果环境变量未设置，停止操作并提示用户按 Setup 步骤配置。

```bash
if [ -z "$IMA_OPENAPI_CLIENTID" ] || [ -z "$IMA_OPENAPI_APIKEY" ]; then
  echo "缺少 IMA 凭证，请按 Setup 步骤配置环境变量 IMA_OPENAPI_CLIENTID 和 IMA_OPENAPI_APIKEY"
  exit 1
fi
```

## API 调用模板

所有请求统一为 **HTTP POST + JSON Body**，Base URL 为 `https://ima.qq.com/openapi/note/v1`。

定义辅助函数避免重复 header：

```bash
ima_api() {
  local endpoint="$1" body="$2"
  curl -s -X POST "https://ima.qq.com/openapi/note/v1/$endpoint" \
    -H "ima-openapi-clientid: $IMA_OPENAPI_CLIENTID" \
    -H "ima-openapi-apikey: $IMA_OPENAPI_APIKEY" \
    -H "Content-Type: application/json" \
    -d "$body"
}
```

> **隐私规则：** 笔记内容属于用户隐私，在群聊场景中只展示标题和摘要，禁止展示笔记正文。

## 接口决策表

| 用户意图 | 调用接口 | 关键参数 |
|---------|---------|---------|
| 搜索/查找笔记 | `search_note_book` | `query_info`（QueryInfo 对象） |
| 查看笔记本列表 | `list_note_folder_by_cursor` | `cursor`(必填，首页传`"0"`) + `limit`(必填) |
| 浏览某笔记本里的笔记,当用户表述"最新"、"最近"之类的通用限定，没有指明笔记本时，都应该直接在全部笔记里去拉 | `list_note_by_folder_id` | `folder_id`(选填,空为全部笔记本) + `cursor`(必填，首次传`""`) + `limit`(必填) |
| 读取笔记正文 | `get_doc_content` | `doc_id` + `target_content_format`(必填，推荐`0`纯文本) |
| 新建一篇笔记 | `import_doc` | `content` + `content_format`(必填，固定`1`) + 可选 `folder_id` |
| 往已有笔记追加内容 | `append_doc` | `doc_id` + `content` + `content_format`(必填，固定`1`) |

## 常用工作流

### 查找并阅读笔记

先搜索获取 `docid`，再用 `get_doc_content` 读取正文：

```bash
# 1. 按标题搜索
ima_api "search_note_book" '{"search_type": 0, "query_info": {"title": "会议纪要"}, "start": 0, "end": 20}'
# 从返回的 docs[].doc.basic_info.docid 中取目标笔记 ID

# 2. 读取正文（纯文本格式，Markdown 格式目前不支持）
ima_api "get_doc_content" '{"doc_id": "目标docid", "target_content_format": 0}'
```

### 浏览笔记本里的笔记

先拉笔记本列表获取 `folder_id`，再拉该笔记本下的笔记：

```bash
# 1. 列出笔记本（首页 cursor 传 "0"）
ima_api "list_note_folder_by_cursor" '{"cursor": "0", "limit": 20}'

# 2. 拉取指定笔记本的笔记（首页 cursor 传 ""）
ima_api "list_note_by_folder_id" '{"folder_id": "user_list_xxx", "cursor": "", "limit": 20}'
```

### 新建笔记

```bash
# 新建到默认位置
ima_api "import_doc" '{"content_format": 1, "content": "# 标题\n\n正文内容"}'

# 新建到指定笔记本
ima_api "import_doc" '{"content_format": 1, "content": "# 标题\n\n正文内容", "folder_id": "笔记本ID"}'
# 返回 doc_id，后续可用于 append_doc
```

### 追加内容到已有笔记

```bash
ima_api "append_doc" '{"doc_id": "笔记ID", "content_format": 1, "content": "\n## 补充内容\n\n追加的文本"}'
```

### 按正文搜索

```bash
ima_api "search_note_book" '{"search_type": 1, "query_info": {"content": "项目排期"}, "start": 0, "end": 20}'
```

## 核心响应字段

**搜索结果**（`SearchedDoc`）：笔记信息路径为 `doc.basic_info`（DocBasic），关键字段：`docid`、`title`、`summary`、`folder_id`、`folder_name`、`create_time`（Unix 毫秒）、`modify_time`、`status`。额外包含 `highlight_info`（高亮匹配，key 为 `doc_title`，value 含 `<em>高亮词</em>`）。

**笔记本条目**（`NoteBookFolder`）：信息路径为 `folder.basic_info`（NoteBookFolderBasic），关键字段：`folder_id`、`name`、`note_number`、`create_time`、`modify_time`、`folder_type`（`0`=用户自建，`1`=全部笔记，`2`=未分类）、`status`。

**笔记列表条目**（`NoteBookInfo`）：信息路径为 `basic_info.basic_info`（DocBasicInfo → DocBasic），关键字段：`docid`、`title`、`summary`、`folder_id`、`folder_name`、`create_time`、`modify_time`、`status`。

**写入结果**（`import_doc`/`append_doc`）：返回 `doc_id`（新建或目标笔记的唯一 ID）。

完整字段定义见 `references/api.md`。

## 分页

- **游标分页 — 笔记本列表**（`list_note_folder_by_cursor`）：首次 `cursor: "0"`，后续用 `next_cursor`，`is_end=true` 时停止。
- **游标分页 — 笔记列表**（`list_note_by_folder_id`）：首次 `cursor: ""`，后续用 `next_cursor`，`is_end=true` 时停止。
- **偏移量分页**（`search_note_book`）：首次 `start: 0, end: 20`，翻页时递增，`is_end=true` 时停止。

## 枚举值

- **`content_format`：** `0`=纯文本，`1`=Markdown，`2`=JSON。写入（`import_doc`/`append_doc`）目前仅支持 `1`（Markdown）。读取（`get_doc_content`）推荐 `0`（纯文本），Markdown 格式不支持。
- **`search_type`：** `0`=标题检索（默认），`1`=正文检索
- **`sort_type`：** `0`=更新时间（默认），`1`=创建时间，`2`=标题，`3`=大小（仅 `search_note_book` 使用）
- **`folder_type`：** `0`=用户自建，`1`=全部笔记（根目录），`2`=未分类

## 注意事项

- `folder_id` 不可为 `"0"`，根目录 ID 格式为 `user_list_{userid}`（从 `folder_type=1` 的笔记本条目获取）
- 笔记内容有大小上限，超过时返回 `100009`，可拆分为多次 `append_doc` 写入
- 展示笔记列表时只展示标题、摘要和修改时间，不要主动展示正文
- 时间字段是 Unix 毫秒时间戳，展示时转为可读格式
- 返回数据为嵌套结构：搜索结果取 `docs[].doc.basic_info.docid`，笔记本取 `note_book_folders[].folder.basic_info.folder_id`，笔记列表取 `note_book_list[].basic_info.basic_info.docid`，注意按层级解析
- **UTF-8 编码**：内容写入前必须确保为 UTF-8 编码。当内容来自临时文件、WebFetch 或外部来源时，按运行环境选择合适的方式转码：

  **Python（推荐，几乎所有环境都有）：**
  ```bash
  # 读取文件，自动检测编码并转为 UTF-8
  content=$(python3 -c "
  import sys
  data = open('tmpfile', 'rb').read()
  for enc in ['utf-8', 'gbk', 'gb2312', 'big5', 'latin-1']:
      try:
          sys.stdout.write(data.decode(enc))
          break
      except (UnicodeDecodeError, LookupError):
          continue
  " 2>/dev/null)

  # 如果内容已在变量中，清洗非法 UTF-8 字节
  content=$(printf '%s' "$content" | python3 -c "import sys; sys.stdout.write(sys.stdin.buffer.read().decode('utf-8','ignore'))")
  ```

  **Node.js：**
  ```bash
  content=$(node -e "const fs=require('fs');const buf=fs.readFileSync('tmpfile');process.stdout.write(buf.toString('utf8'))")
  # 已知编码（如 GBK）：
  content=$(node -e "const fs=require('fs');process.stdout.write(new TextDecoder('gbk').decode(fs.readFileSync('tmpfile')))")
  ```

  **Unix (macOS/Linux)：**
  ```bash
  content=$(iconv -f "$(file -b --mime-encoding tmpfile)" -t UTF-8 tmpfile 2>/dev/null || cat tmpfile)
  ```

  **Windows PowerShell：**
  ```powershell
  # 读取非 UTF-8 文件并转码
  $content = [System.IO.File]::ReadAllText('tmpfile', [System.Text.Encoding]::Default)
  [System.IO.File]::WriteAllText('tmpfile.utf8', $content, [System.Text.Encoding]::UTF8)
  ```

  > **PowerShell 5.1 发送请求注意事项：**
  > `Invoke-RestMethod` 传入字符串 Body 时，默认使用系统 ANSI 编码（中文 Windows 为 GBK），而非 UTF-8，即使设置了 `Content-Type: charset=utf-8` 也无效。必须**显式转为 UTF-8 字节数组**再传入 `-Body`，同时使用 `ConvertTo-Json` 构建 JSON 以避免手动拼接的转义风险：
  >
  > ```powershell
  > $body = @{ title = "标题"; content = $content; content_format = 1 } | ConvertTo-Json -Depth 10
  > $utf8Bytes = [System.Text.Encoding]::UTF8.GetBytes($body)
  > Invoke-RestMethod -Uri $url -Method Post -Body $utf8Bytes -ContentType "application/json; charset=utf-8" -Headers $headers
  > ```
  >
  > PowerShell 7+ 默认使用 UTF-8，无需额外处理。

  标题同理，确保传入 API 的所有字符串字段均为合法 UTF-8。


## 错误处理

| 错误码 | 含义              | 建议处理                  |
|--------|-----------------|-----------------------|
| 0 | 成功              | —                     |
| 100001 | 参数错误            | 检查请求参数格式和必填字段         |
| 100002 | 无效 ID           | 检查凭证配置                |
| 100003 | 服务器内部错误         | 等待后重试                 |
| 100004 | size 不合法 / 空间不够 | 检查参数范围                |
| 100005 | 无权限             | 确认操作的是用户自己的笔记         |
| 100006 | 笔记已删除           | 告知用户该笔记不存在            |
| 100008 | 版本冲突            | 重新获取内容后再操作            |
| 100009 | 超过大小限制          | 拆分为多次 `append_doc` 写入 |
| 310001 | 笔记本不存在          | 检查 `folder_id` 是否正确   |
| 20002 | apiKey超过最大限频    |
| 20004 | apikey 鉴权失败     | 检查凭证配置是否正确            |
| 110037 | apikey 过期       | 请获取最新apikey：https://ima.qq.com/agent-interface          |
