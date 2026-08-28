# BKN 规范（LLM 生成用）

你负责生成符合 BKN 格式的 Markdown，供业务知识网络建模使用。

## 文件扩展名

- `.bkn`：BKN 定义文件（schema）

## 根文件与目录加载

- **根文件**：`network.bkn`，唯一入口
- **目录输入**：`validate network <dir>`、`load_network(dir)` 等支持传入目录，自动发现 `network.bkn`；若不存在则报错
- SDK/CLI 自动发现同目录下的 BKN 文件，无需 `includes` 声明

## 文件结构

每个 `.bkn` 文件由两部分组成：
1. **YAML Frontmatter**（元数据，以 `---` 包裹）
2. **Markdown Body**（定义内容）

## Frontmatter 类型

| type | 说明 |
|------|------|
| `knowledge_network` | 完整知识网络顶层容器 |
| `object_type` | 单个对象类型定义 |
| `relation_type` | 单个关系类型定义 |
| `action_type` | 单个行动类型定义 |
| `concept_group` | 概念分组 |

## 目录结构

```
{business_dir}/
├── SKILL.md                     # agentskills.io 标准入口
├── network.bkn                  # 网络根文件
├── CHECKSUM                     # 可选；SDK/CLI 可生成
├── object_types/
│   └── {object}.bkn
├── relation_types/
│   └── {relation}.bkn
├── action_types/
│   └── {action}.bkn
├── concept_groups/
│   └── {group}.bkn
└── data/                        # 可选实例数据（如 CSV）
```

## 网络 (Knowledge network)

```yaml
---
type: knowledge_network
id: {network_id}
name: {显示名称}
tags: [tag1, tag2]               # 可选
business_domain: {domain}        # 可选
---
```

正文：
- `# {显示名称}` + 描述
- `## Network Overview`：网络概览（对象/关系/行动列表）

## 对象类型 (ObjectType)

```yaml
---
type: object_type
id: {object_id}                  # 小写+下划线
name: {显示名称}
tags: [tag1, tag2]               # 可选
---
```

推荐顺序如下（与 [assets/templates/object_type.bkn.template](../assets/templates/object_type.bkn.template) 及 `references/examples/` 下可校验网络对齐）：
- `## ObjectType: {显示名称}` + 简短描述
- `### Data Properties`（必须）：表格，列 Name | Display Name | Type | Description | Mapped Field
- `### Keys`（必须）：
  - `Primary Keys: {key_name}`（至少一个）
  - `Display Key: {key_name}`（一个）
  - `Incremental Key: {key_name}`（可选，可为空）
- `### Logic Properties`（可选）：无内容时保留空小节；有内容时 `#### {property_name}`，含 Display/Type/Source/Description，以及 Parameter 表（列 Parameter | Type | Source | Binding | Description）
  - Source 值：`property`（对象属性）/ `input`（用户输入）/ `const`（常量）
  - Binding：Source 为 property 时填属性名，const 时填常量值，input 时填 `-`
- `### Data Source`（可选）：表格，列 Type | ID | Name，行 `data_view | {view_id} | {view_name}`；无数据视图绑定时可省略整节

### 数据类型

Type 列标准类型（大小写不敏感）：

| 类型 | 说明 |
|------|------|
| string | 字符串 |
| integer | 整数 |
| float | 浮点数 |
| decimal | 精确十进制数 |
| boolean | 布尔值 |
| date | 日期（无时间） |
| time | 时间（无日期） |
| datetime | 日期时间 |
| text | 长文本 |
| json | JSON 结构数据 |
| binary | 二进制数据 |

不在列表中的类型透传。

## 关系类型 (RelationType)

```yaml
---
type: relation_type
id: {relation_id}
name: {显示名称}
tags: [tag1, tag2]               # 可选
---
```

正文：
- `## RelationType: {显示名称}` + 简短描述
- `### Endpoint`（必须）：表格 Source | Target | Type（`direct` 或 `data_view`）
- **direct 类型**时：
  - `### Mapping Rules`：表格 Source Property | Target Property
- **data_view 类型**时：
  - `### Mapping View`：表格 Type | ID
  - `### Source Mapping`：表格 Source Property | View Property
  - `### Target Mapping`：表格 View Property | Target Property

## 行动类型 (ActionType)

```yaml
---
type: action_type
id: {action_id}
name: {显示名称}
tags: [tag1, tag2]                  # 可选
action_type: add | modify | delete  # ⚠️ 必填 —— 后端 ActionTypeMap 硬白名单
enabled: boolean                    # 可选，建议默认 false
risk_level: low | medium | high     # 可选
requires_approval: boolean          # 可选
---
```

> ⚠️ **`action_type` 字段的三个坑（必须读懂再写）**
>
> 1. **位置**：`action_type` 是 **YAML frontmatter 顶级字段**，**不是** `### Bound Object` 表格的"Action Type"列。后端 Go parser 解析 Bound Object 表时**只取第一列 Bound Object，第二列原样丢弃**。
> 2. **取值**：仅限 **`add`** / **`modify`** / **`delete`**。
> 3. **后端错误信息有 bug**：缺失或非法时 backend 返回
>    `"The action type is expected one of [create, update, delete], actual is [...]"`
>    —— **不要照这串值改！** 这是 backend 历史 bug（错误消息硬编码、未跟实际 enum 同步）。实际白名单就是 `add` / `modify` / `delete`，`create` / `update` 会被拒。
>
> **缺失即报错**：每个 ActionType 文件 frontmatter 都必须包含 `action_type:` 这一行，缺了 backend 报 `actual is []`。
>
> **只读语义不要建为 ActionType**：ActionType 只承载对实例的写操作。查询/监控/追溯/校验等只读语义请用 **object-type query / subgraph / metric / semantic search**，不要落到 `action_types/*.bkn`。
>
> **何时生成 ActionType（按需，可为空）**：
>
> - ActionType 是按需生成的，**可整组为空**；空 `action_types/` 不影响网络合法性。
> - **该建**：PRD 含对实例的明确写操作（增/改/删），且能识别出触发条件或绑定工具——每个写操作建一个 ActionType。
> - **该空**：PRD 全为只读语义（查询/监控/分析/统计/追溯）——走 object-type query / subgraph / metric / semantic search，不建 ActionType。
> - **灰色地带**（如"审批"是写操作还是状态流转）——交用户确认，不默认建也不默认跳过。
> - **一致性**：`action_types/` 是否为空，必须与 `network.bkn` Network Overview 的 ActionType 列表一致（都空，或一一对应）。

正文：
- `## ActionType: {显示名称}` + 简短描述
- `### Bound Object`（必须）：表格 Bound Object | Action Type
  - **Bound Object 列**：填实际绑定的 object_type id（后端读这一列）
  - **Action Type 列**：填 `add` / `modify` / `delete`，与 frontmatter `action_type:` 保持一致（**仅为人类可读**，backend 不解析此列）
- `### Affect Object`（可选）：表格 Affect Object
- `### Trigger Condition`（可选）：YAML 代码块，格式：
  ```yaml
  condition:
    object_type_id: {object_type_id}
    field: {property_name}
    operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
    value: {value}
  ```
- `### Pre-conditions`（可选）：表格 Object | Check | Condition | Message
  - Check 格式：`relation:{relation_id}` 或 `property:{property_name}`
- `### Scope of Impact`（可选）：表格 Object | Impact Description
- `### Tool Configuration`（必须）：
  - tool 类型：表格 Type | Toolbox ID | Tool ID
  - mcp 类型：表格 Type | MCP ID | Tool Name
- `### Parameter Binding`（必须）：表格 Parameter | Type | Source | Binding | Description
  - Source 值：`property` / `input` / `const`
- `### Schedule`（可选）：表格 Type | Expression（`FIX_RATE` 或 `CRON`）
- `### Execution Description`（可选）：编号列表描述执行流程

### 触发条件操作符

| 操作符 | 说明 |
|--------|------|
| == | 等于 |
| != | 不等于 |
| > / < / >= / <= | 比较 |
| in / not_in | 包含于/不包含于 |
| exist / not_exist | 存在/不存在 |
| range | 范围内 |

## 概念分组 (ConceptGroup)

```yaml
---
type: concept_group
id: {group_id}
name: {显示名称}
tags: [tag1, tag2]               # 可选
---
```

正文：
- `## ConceptGroup: {显示名称}` + 简短描述
- `### Object Types`（必须）：表格 ID | Name | Description

## 更新与删除（无 patch 模型）

- 定义文件导入 = add/modify（upsert）；修改即编辑文件后重新导入
- 删除元素通过 SDK/CLI delete API 执行，不通过 BKN 文件
- **不要生成 type: delete 或 type: patch 文件**

## 输出规则（必须遵守）

1. **仅输出 BKN Markdown**：含 frontmatter 和 body，无多余说明
2. **不要包裹代码块**：不要用 \`\`\`markdown 包裹整体输出
3. **引用已存在的 ID**：object/relation 引用时，使用项目中已有的 id
4. **表格格式**：按上述列名严格对齐
5. **命名**：ID 使用小写字母、数字、下划线；显示名和描述用中文（除非另有要求）
6. **必填字段**：
   - 所有类型：type、id、name
   - ObjectType：Data Properties、Keys（Primary Keys + Display Key）
   - RelationType：Endpoint、Mapping Rules（或 Mapping View + Source/Target Mapping）
   - ActionType：frontmatter `action_type`（add/modify/delete）、Bound Object、Tool Configuration、Parameter Binding
   - ConceptGroup：Object Types
7. **标题层级**：`#` 网络标题、`##` 类型定义、`###` 定义内 section、`####` 子项（逻辑属性名）
8. **业务规则放置**：
   - 网络级规则 → `network.bkn` 的 `# {name}` 后描述区
   - 类型级规则 → 对应类型文件 body 描述段（`## ObjectType: {name}` 后、第一个 `###` 前），不放 frontmatter
   - 属性级规则 → Data Properties 的 Description 列
   - **不要添加规范外的额外信息**：BKN 格式约束严格，超出标准 section 的内容在解析和导入时可能不被保存
