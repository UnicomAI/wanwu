---
name: IMA笔记
description: |
  IMA笔记 - 腾讯IMA个人笔记管理工具。

  **当用户想要「搜索IMA笔记」「查找笔记」「列出笔记本」「查看某个笔记本下的笔记」「从markdown新建笔记」「追加内容到笔记」「获取笔记纯文本」时，使用此技能。**

  功能：搜索笔记内容（支持标题/正文检索）、列出笔记本、按笔记本拉取笔记列表、从 Markdown 导入新建笔记、追加内容到已有笔记、获取笔记纯文本内容。
metadata: {"openclaw": {"requires": {"env": ["IMA_OPENAPI_CLIENTID", "IMA_OPENAPI_APIKEY"]}, "primaryEnv": "IMA_OPENAPI_CLIENTID", "homepage": "https://ima.qq.com"}}
---

# IMA笔记 API

## ⚠️ 必读约束

### 🔒 认证

所有请求必须携带 Header：

```
ima-openapi-clientid: {IMA_OPENAPI_CLIENTID}
ima-openapi-apikey: {IMA_OPENAPI_APIKEY}
Content-Type: application/json
```

### 🔒 安全规则

- 笔记属于用户隐私，**不要在群聊中主动展示笔记内容**。
- 仅响应授权用户的笔记操作请求。

---

## 快速决策

| 用户意图 | 接口别名 |
|---------|---------|
| 「搜索笔记」「找包含XX的笔记」 | `/openapi/note/v1/search_note_book` |
| 「列出笔记本」「有哪些笔记本」 | `/openapi/note/v1/list_note_folder_by_cursor` |
| 「查看XX笔记本里的笔记」 | `/openapi/note/v1/list_note_by_folder_id` |
| 「从markdown新建笔记」「导入笔记」 | `/openapi/note/v1/import_doc` |
| 「追加内容到笔记」「在笔记末尾添加」 | `/openapi/note/v1/append_doc` |
| 「获取笔记纯文本」「读取笔记内容」 | `/openapi/note/v1/get_doc_content` |

---

## 数据结构

---

#### DocBasicInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `basic_info` | DocBasic | 见 [DocBasic](#docbasic) |

---

#### DocBasic

| 字段 | 类型 | 说明 |
|------|------|------|
| `docid` | string | 文章 id |
| `title` | string | 标题 |
| `summary` | string | 简介 |
| `create_time` | int64 | |
| `modify_time` | int64 | |
| `status` | DocStatus | 文章状态，`0`=正常，`1`=已删除 |
| `folder_id` | string | 文件夹 id |
| `folder_name` | string | 文件夹名称 |
| `summary_style` | map\<string, string\> | 简介样式 |

---

### FolderItem（笔记本条目）

`list_note_folder_by_cursor` 返回的笔记本对象，字段如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| `folder_id` | string | 笔记本唯一 ID |
| `name` | string | 笔记本名称 |
| `note_number` | int64 | 笔记本内笔记数量 |
| `create_time` | int64 | 创建时间（Unix 毫秒） |
| `modify_time` | int64 | 修改时间（Unix 毫秒） |
| `parent_folder_id` | string | 上级笔记本 ID（支持嵌套） |
| `folder_type` | int | 类型：`0`=用户自建，`1`=全部笔记，`2`=未分类 |
| `status` | int | 状态：`0`=正常，`1`=已删除 |

---

#### QueryInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 标题 query |
| `content` | string | 正文 query |

---

#### SearchedDoc

| 字段 | 类型 | 说明 |
|------|------|------|
| `doc` | DocBasicInfo | 笔记 basic 数据，见 [DocBasicInfo](#docbasicinfo) |
| `highlight_info` | map\<string, string\> | 该条笔记匹配的高亮词，key: `doc_title`（文档标题），value: 包含 `<em>高亮词</em>` 的字段值 |

---

#### NoteBookFolder

| 字段 | 类型 | 说明 |
|------|------|------|
| `folder` | NoteBookFolderBasicInfo | 笔记本信息，非笔记本为空，见 [NoteBookFolderBasicInfo](#notebookfolderbasicinfo) |

---

#### NoteBookFolderBasicInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `basic_info` | NoteBookFolderBasic | 见 [NoteBookFolderBasic](#notebookfolderbasic) |

---

#### NoteBookFolderBasic

| 字段 | 类型 | 说明                              |
|------|------|---------------------------------|
| `folder_id` | string | 文件夹 id                          |
| `name` | string | 笔记本名称                           |
| `status` | DocStatus | 笔记本状态，`0`=正常，`1`=已删除            |
| `create_time` | int64 | 创建时间                            |
| `modify_time` | int64 | 修改时间                            |
| `note_number` | int64 | 笔记数量                            |
| `folder_type` | FolderType | 文件夹类型：`0`=用户自建，`1`=全部笔记，`2`=未分类 |

---

#### NoteBookInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `basic_info` | DocBasicInfo | 笔记基础信息，见 [DocBasicInfo](#docbasicinfo) |

---

## 接口详情

### 1. 搜索笔记
POST /openapi/note/v1/search_note_book

**触发场景**：用户说「搜索」「找笔记」「查找包含XX的内容」

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|----|----|
| `search_type` | SearchType | 否  | 检索方式，默认为标题，`0`=标题，`1`=正文 |
| `sort_type` | SortType | 否  | 排序方式，默认为更新时间，`0`=更新时间，`1`=创建时间，`2`=标题，`3`=大小 |
| `query_info` | QueryInfo | 否  | 用户 query，见 [QueryInfo](#queryinfo) |
| `start` | int64 | 是  | 翻页字段，起始编号 |
| `end` | int64 | 是  | 翻页字段，终止编号，与start相差不超过20 |
| `query_id` | string | 否  | queryid |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `docs` | SearchedDoc[] | 检索到的笔记 list，见 [SearchedDoc](#searcheddoc) |
| `is_end` | bool | 是否为最后一批数据 |
| `total_hit_num` | int64 | 检索命中结果总数 |

---

### 2. 列出笔记本
POST /openapi/note/v1/list_note_folder_by_cursor

**触发场景**：用户说「列出笔记本」「有哪些分类」「查看笔记本目录」

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `cursor` | string | **是** | 游标，第一页传 `"0"`，后续传后台返回的值 |
| `limit` | uint64 | **是** | 获取笔记数量限制，不要超过20 |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `note_book_folders` | NoteBookFolder[] | 见 [NoteBookFolder](#notebookfolder) |
| `next_cursor` | string | 下次请求的起始游标 |
| `is_end` | bool | 是否为最后一批数据 |

---

### 3. 按笔记本拉取笔记列表
POST /openapi/note/v1/list_note_by_folder_id

**触发场景**：用户说「查看XX笔记本的笔记」「列出这个笔记本里的内容」

> 全部笔记根目录的 `folder_id` 为 `user_list_{userid}`，可从「列出笔记本」返回的 `folder_id` 获取。

#### 请求参数       

| 字段 | 类型 | 必填 | 说明              |
|------|------|---|-----------------|
| `folder_id` | string | 否 | 笔记本 ID，根目录为空    |
| `cursor` | string | **是** | 当前游标，首次传空字符串 `""` |
| `limit` | uint64 | **是** | 获取笔记数量限制，不要超过20 |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `note_book_list` | NoteBookInfo[] | 见 [NoteBookInfo](#notebookinfo) |
| `next_cursor` | string | 下次请求的起始游标 |
| `is_end` | bool | 是否为最后一批数据 |

---

### 4. 从 Markdown 新建笔记
POST /openapi/note/v1/import_doc

**触发场景**：用户说「从 Markdown 新建笔记」「导入笔记」「把这段 Markdown 保存为笔记」

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|---|----|
| `content_format` | int | **是** | 文本类型：`1`=Markdown（默认）目前仅支持 `MARKDOWN`（值为 `1`） |
| `content` | string | **是** | 笔记正文内容, 只支持markdown格式, 不要传空值 |
| `folder_id` | string | 否 | 关联的笔记本id |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|----|
| `doc_id` | string | 新doc的唯一ID |

---

### 5. 追加内容到笔记
POST /openapi/note/v1/append_doc

**触发场景**：用户说「在这篇笔记末尾追加内容」「把 XX 添加到笔记里」

#### 请求参数

| 字段 | 类型 | 必填 | 说明                  |
|------|------|-----|---------------------|
| `doc_id` | string | **是** | 目标笔记的唯一ID, 需要是本人的笔记 |
| `content_format` | int | **是** | 文本类型：`1`=Markdown（默认）目前仅支持 `MARKDOWN`（值为 `1`） |
| `content` | string | **是** | 要追加的文本内容, 只支持markdown格式 |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|--|
| `doc_id` | string | 目标笔记的唯一ID |

---

### 6. 获取笔记纯文本
POST /openapi/note/v1/get_doc_content

**触发场景**：用户说「读取笔记内容」「获取这篇笔记的纯文本」「把笔记转成 Markdown」

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| `doc_id` | string | **是** | 目标笔记的唯一 ID, 需要是本人的笔记 |
| `target_content_format` | int | **是** | 目标文本类型：`0`=纯文本（推荐），`1`=Markdown（不支持），`2`=JSON |

#### 返回字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `content` | string | 笔记的文本内容（按 `target_content_format` 格式返回） |

---

## 枚举值

### `sort_type`（排序方式）

| 值 | 说明 |
|----|------|
| `0` | 更新时间（默认） |
| `1` | 创建时间 |
| `2` | 标题 |
| `3` | 大小 |

### `search_type`（检索方式）

| 值 | 说明 |
|----|------|
| `0` | 标题检索（默认） |
| `1` | 正文检索 |

### `content_format`（文本类型）

| 值 | 说明 |
|----|------|
| `0` | PLAINTEXT - 纯文本 |
| `1` | MARKDOWN - Markdown 格式 |
| `2` | JSON - JSON 格式 |

### `FolderType`
| 值 | 说明   |
|----|------|
| `0` | 用户自建 |
| `1` | 全部笔记 |
| `2` | 未分类  |

---

## 游标翻页使用规范

1. **首次请求**：`cursor` 传空字符串 `""`
2. 检查返回的 `is_end`：`false` 表示还有更多数据
3. 将返回的 `next_cursor` 作为下次请求的 `cursor`
4. `is_end = true` 时停止翻页

---

## 错误码

| 错误码 | 说明                        |
|--------|---------------------------|
| 0 | 成功                        |
| 100001 | 参数错误                      |
| 100002 | 携带无效的 ID                  |
| 100003 | 服务器内部错误                   |
| 100004 | 拉取的 size 不合法（超出范围）/ 用户空间不够 |
| 100005 | 不能获取私有笔记的访客信息 / 不是笔记的作者   |
| 100006 | 笔记已被删除                    |
| 100008 | 版本冲突                      |
| 100009 | 单篇笔记超过最大限制                |
| 310001 | 笔记本不存在                    |
| 20002 | apiKey超过最大限频              |
| 20004 | apikey鉴权失败                |
| 110037 | apikey 过期       |