---
name: ontology-core
description: >-
  操作 知识网络（BKN）— 构建知识网络、查询 Schema/实例、语义搜索、执行 Action。
  操作数据源与数据视图 — 数据源连接与查询、原子/自定义视图浏览与 SQL 查询。
  操作 Vega 可观测平台 — 查询 Catalog/资源/连接器类型、健康巡检。
  当用户提到"知识网络"、"知识图谱"、"查询对象类"、"执行 Action"、
  "数据源"、"数据视图"、"原子视图"、"Catalog"、"Vega"、
  "健康检查"、"巡检"等意图时自动使用。
allowed-tools: Bash(ontology *)
argument-hint: [自然语言指令]
---

# Ontology CLI

平台的命令行工具 `ontology`，覆盖知识网络管理与查询、数据源、数据视图、Vega 可观测、通用 API 调用。

## 安装

**无需安装**。`ontology` 已内置在执行环境，直接调用即可。

## 使用方式

```bash
ontology <command> [subcommand] [options]
```

**完整子命令与参数以当前 CLI 为准**：运行 `ontology --help`（或 `-h`）查看与代码同步的用法列表；查版本用 `ontology --version` / `-V`。子命令细节用 `ontology <group> <subcommand> --help`（例如 `ontology bkn push --help`、`ontology dataview find --help`）。

本 skill 下的 `references/*.md` 与 CLI 行为对齐；**表格与 reference 为速查**。

**别名**：`ontology curl` 等同于 `ontology call`。

**业务域（business domain）**：ontology CLI 没有 `config` 命令组；business domain **按命令传 `-bd <value>` 参数**。本部署默认使用 `bd_public`，无需显式传入。

## 使用前提

**直接执行 `ontology <command>` 即可。

CLI 在同网络下会按服务名直连后端（`vega-backend:13014`、`vega-bkn-backend:13014`、`vega-mdl-data-model:13020`、`vega-mdl-uniquery:13011`、`vega-data-connection:8098`、`vega-ontology-query:13018`、`vega-agent:5000`），**无需设置 `ONTOLOGY_BASE_URL`**。

### 调用约定

```bash
ontology --user-id <accountId> <command> [options]
```

- **必须显式传 `--user-id <accountId>`**（顶层选项，写在子命令之前）。本部署下所有命令组（`bkn` / `ds` / `dataview` / `vega` / `call`，含 `curl` 别名）都适用；缺省时调用会失败或语义错误
  - 若未拿到 accountId，先向用户索要，**不要**自行编造或省略
- ontology CLI **没有** `auth` / `token` / `config` 命令组，无需任何登录或登录状态检查
- 出现 401/认证类错误时，直接报错，**不要**尝试刷新 token 或引导用户登录

### 业务域优先级

1. 命令行 `-bd <value>` 参数（每条命令显式传入）
2. 默认 `bd_public`

## 常见 CLI 坑（编排层必读）

下列三类陷阱在编排层频繁误用，会直接导致语法报错或全量输出截断。编排前先核对，不要等失败后再读 references。

| 命令 | 错误用法 | 正确用法 |
|------|----------|----------|
| `bkn object-type query` 的 `condition` | `{"field":{"op":"val"}}` 等嵌套写法 | 扁平 `{"field":"f","operation":"op","value":"v"}`；多条件用 `{"operation":"and"/"or","sub_conditions":[...]}` |
| `bkn object-type get` | 用 `--name "中文名"` 选对象类 | 必须传 `<ot-id>`；不知道 id 时先 `bkn object-type list` 拿 id |
| `bkn object-type list <kn-id>` 直接全量 | 直接消费 full output | 单次返回常 30~50KB 易截断；只为路由/选 ot 时，用 jq 投影到摘要（见下方"查询策略"） |

**jq 投影通用规则**：不同 `ontology` 命令的返回结构不一致，套同一个模板会拿到 `null` 或 `Cannot iterate` 报错。常用命令对照：

| 命令 | 顶层结构 | jq 入口 |
|------|----------|---------|
| `bkn object-type list` | `{entries:[...]}` | `.entries[] \| ...` |
| `bkn object-type get` | `{entries:[obj]}`（单元素也包一层 entries） | `.entries[0] \| ...` 或 `.entries[] \| ...` |
| `bkn object-type query` | `{datas:[...], search_after:[...]}` | `.datas[] \| ...` |
| `bkn list` / `relation-type list` 等 | 多数为 `{entries:[...]}` | `.entries[] \| ...` |
| `dataview list` | 裸数组 `[...]` | `.[] \| ...` |
| `dataview get` | 裸对象 `{...}` | `.` 或直接访问字段 |

**第一次对某命令套 jq 模板前，先裸跑一次 `| jq 'keys'`（确认顶层键名）或 `| jq 'type'`（确认是 array 还是 object）**，再决定入口；避免凭命名直觉假设结构。

合法 `operation` 值：`==` `!=` `>` `>=` `<` `<=` `in` `not_in` `like` `not_like` 等；`eq`/`gt`/`lt` 等不是合法操作符。完整支持矩阵见 [`references/bkn.md`](references/bkn.md#object-type-query-条件过滤)。

## 命令组总览

> **本部署仅运行 Vega 一侧的服务**（vega-web / vega-backend / vega-bkn-backend / mdl-data-model / mdl-uniquery / data-connection / ontology-query 等）。
> **未部署 Decision Agent / Skill Registry / Toolbox / Dataflow 等服务**，且 ontology CLI 自身也已**移除** `auth` / `config` / `agent` / `skill` / `toolbox` / `dataflow` / `context-loader` 命令组，对应能力全部不可用。

| 命令组 | 说明 | 常用命令 | 详细参考 |
|--------|------|---------|---------|
| `bkn` | BKN 知识网络管理、Schema、查询、Action | `bkn list`、`bkn get <id>`、`bkn object-type`、`bkn validate`/`push`、`pull`、`create-from-ds`/`create-from-csv` 等 | `references/bkn.md` |
| `ds` | 数据源管理 | `ds list`、`ds get <id>`、`ds tables <id>`、`ds connect ...` | `references/ds.md` |
| `dataview` | 原子/自定义数据视图（mdl-data-model） | `dataview list`、`find --name`、`get`、`query`（SQL / mdl-uniquery）、`delete` | `references/dataview.md` |
| `vega` | Vega 可观测平台 | `vega health`、`vega stats`、`vega catalog list`、`vega resource list`、`vega connector-type list` | `references/vega.md` |
| `call` | 通用 API 调用 | `call <url> [-X POST] [-d '...']`（可用 `curl` 别名） | `references/call.md` |


## 操作指南

| 场景 | 说明 | 详细参考 |
|------|------|---------|
| 从数据库/CSV 构建 KN | 连接数据源 → CSV 导入 → 创建 KN → 构建索引 → 查询验证 | [references/build-kn-from-db.md](references/build-kn-from-db.md) |
| 列/查数据视图 | `list` 浏览；`find --name` 按名搜索（`--exact`/`--wait`）；`query` 对视图跑 SQL | [references/dataview.md](references/dataview.md) |
| Vega 巡检 | `vega health` / `vega stats` / `vega catalog list` / `vega resource list` | [references/vega.md](references/vega.md) |

**按需阅读**：需要子命令完整参数或编排示例时，读取对应的 reference 文件。
**遇到关于 agent / skill / toolbox / dataflow / context-loader 的请求**：先告知用户 ontology CLI 与本环境均未提供，不要尝试执行。

## 调用示例

```bash
ontology bkn list
ontology vega health
ontology ds list
ontology dataview list --limit -1
```


## 注意事项

- **无需任何 `export` 环境变量**：本部署 `AUTH_ENABLED=false`，CLI 也内置了所有服务的默认地址，直接执行 `ontology <command>` 即可
- **本部署 business domain 固定为 `bd_public`**（ontology CLI 默认值）；如需切换业务域，使用 `-bd <value>` 参数显式传入
- **本部署不可用的命令组**：ontology CLI 已无 `auth` / `token` / `config`；本部署也未提供 `agent` / `skill` / `toolbox` / `tool` / `dataflow` / `context-loader` 对应服务。用户提到 Agent / Skill / Toolbox / 数据流编排相关需求时，先说明环境不支持，不要尝试执行
- Action 执行有副作用，执行前向用户确认
- 出现 401 等认证错误时直接报错，不要尝试登录或刷新

## 查询策略（object-type query）

调用 `object-type query` 时必须限制 `limit`、用 `search_after` 分页、用 `condition` 过滤，避免宽表 JSON 截断。完整规则与示例见 [`references/bkn.md`](references/bkn.md#object-type-query-strategy-for-llm-and-agent)。

### Schema 摘要捷径

`bkn object-type list <kn-id>` 默认返回每个对象类完整 `data_properties`（含索引配置），单 KN 常达 30~50KB 易截断。路由/选 ot 阶段只需 id+name+comment+主键时，用 jq 投影到摘要：

```bash
ontology --user-id <accountId> bkn object-type list <kn-id> \
  | jq '.entries[] | {id, name, comment, primary_keys, display_key}'
```

输出从数十 KB 压到 ~2KB；定位到目标对象类后再用 `bkn object-type get <kn-id> <ot-id>` 取完整字段。
