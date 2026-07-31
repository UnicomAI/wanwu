# BKN 命令参考

知识网络管理：schema CRUD、构建、导出、推送/拉取。

与 CLI 一致：运行 `ontology bkn`、`ontology bkn validate --help`、`ontology bkn push --help` 可查看与当前版本同步的用法（含 `validate`/`push` 的编码相关选项）。

## 概览

```bash
ontology bkn list                     # 列出知识网络
ontology bkn get <kn_id> [--stats] [--export]
```

## 知识网络

```bash
ontology bkn list [--name <n>] [--name-pattern <p>] [--tag <t>] [--sort update_time] [--direction desc] [--offset 0] [--limit 50] [-v]
ontology bkn get <kn_id> [--stats] [--export]
ontology bkn stats <kn_id>
ontology bkn export <kn_id>
ontology bkn create [options]                         # 创建空知识网络（或 --body-file）
ontology bkn create-from-ds <ds_id> --name <name> [--tables <t1,t2>] [--build/--no-build] [--timeout 300]  # 每张表先等待已有原子视图（keyword 搜索 + 轮询），否则再创建；单独调试见 `ontology dataview`
ontology bkn create-from-csv <ds_id> --files <glob> --name <name> [--table-prefix <p>] [--build/--no-build] [--timeout 300]
ontology bkn update <kn_id> [--name <n>] [--description <d>] [--tag <t> ...]
ontology bkn build <kn_id> [--wait/--no-wait] [--timeout 300]
ontology bkn delete <kn_id> [--yes]
ontology bkn validate <directory> [--detect-encoding/--no-detect-encoding] [--source-encoding <name>]
ontology bkn push <directory> [--detect-encoding/--no-detect-encoding] [--source-encoding <name>]
# 规范上 .bkn 应为 UTF-8。默认会检测 .bkn 编码并规范为 UTF-8 再校验/打包；可用 --no-detect-encoding 要求严格 UTF-8；
# 已知为 GBK/GB18030 等时可 --source-encoding gb18030（整目录统一）。检测置信度不足时会报错并提示指定编码。
ontology bkn pull <kn_id> [<directory>]  # 下载 BKN tar 并解压
```

## Object Type

```bash
ontology bkn object-type list <kn_id>
ontology bkn object-type get <kn_id> <ot_id>                # -v 显示完整 data_properties
ontology bkn object-type create <kn_id> --name <n> --dataview-id <dv> --primary-key <pk> --display-key <dk> [--property '<json>' ...]
# update：合并模式会先 GET 当前对象类，再合并参数后 PUT（与 Studio 行为一致）
ontology bkn object-type update <kn_id> <ot_id> [--name <n>] [--display-key <dk>] [--add-property|--update-property '<json>' ...] [--remove-property <name> ...] [--tags '["标签"]'] [--comment <s>] [--icon <i>] [--color <c>]
# update：原始 JSON 模式（整段 PUT body，勿与上面合并参数同时使用）
ontology bkn object-type update <kn_id> <ot_id> '<full-json-body>'
ontology bkn object-type delete <kn_id> <ot_ids> [--yes/-y]
ontology bkn object-type query <kn_id> <ot_id> ['<json>']   # 查询实例（支持 --limit/--search-after）
ontology bkn object-type properties <kn_id> <ot_id> '<json>' # 查询实例属性
```

### object-type update 说明（数据属性：增 / 改 / 删）

| 操作 | CLI 方式 | 说明 |
|------|----------|------|
| **添加** | `--add-property '<json>'` | `name` 在现有列表中不存在则**追加**一条属性。 |
| **更新** | `--update-property '<json>'` 或 `--add-property '<json>'` | 与添加共用逻辑：`name` **已存在**则**整段替换**该属性（改 display_name、type、mapped_field 等）。`--update-property` 与 `--add-property` 等价，仅语义区分。 |
| **删除** | `--remove-property <name>` | 按 `name` 从 `data_properties` 中移除；可重复多次删多个。 |

- **`--tags`**：替换标签，值为 JSON 字符串数组，如 `'["足球","球员"]'`。
- **原始 JSON（整表替换）**：第三个参数传入完整对象类 JSON（或以 `ontology call` PUT），适合从 Studio 复制整段 body；勿与上面合并参数混用。
- **`ontology call -d`**：带 body 时会自动设置 **`Content-Type: application/json`**（未手动指定 `-H Content-Type` 时）。

## object-type query 条件过滤

`query` 的 JSON body 支持 `condition` 字段，用于服务端过滤，避免全量拉取后本地筛选。

### 语法

```jsonc
// 单条件（limit 不超过 30）
{"limit": 20, "condition": {"field": "quantity", "operation": ">", "value": 4000}}

// 多条件组合（AND）
{"limit": 20, "condition": {
  "operation": "and",
  "sub_conditions": [
    {"field": "quantity", "operation": ">", "value": 4000},
    {"field": "status", "operation": "==", "value": "active"}
  ]
}}

// OR 组合
{"limit": 20, "condition": {
  "operation": "or",
  "sub_conditions": [
    {"field": "category", "operation": "==", "value": "A"},
    {"field": "category", "operation": "==", "value": "B"}
  ]
}}
```

### 支持的 operation

| 类型 | operation | SQL 视图兼容 |
|------|-----------|:----------:|
| 比较 | `==`, `!=`, `>`, `>=`, `<`, `<=` | ✅ |
| 范围 | `in`, `not_in` | ✅ |
| 文本 | `like`, `not_like` | ✅ |
| 逻辑组合 | `and`, `or`（配合 `sub_conditions` 数组） | ✅ |
| 范围（索引） | `range`, `out_range` | ⚠️ 仅 OpenSearch |
| 文本（索引） | `contain`, `not_contain`, `prefix`, `regex` | ⚠️ 仅 OpenSearch |
| 全文搜索 | `match`, `match_phrase`, `multi_match` | ❌ SQL 视图报 500 |
| 空值/存在 | `exist`, `not_exist`, `null`, `not_null`, `empty`, `not_empty` | ⚠️ 仅 OpenSearch |

> **重要**：大部分数据源（MySQL / PostgreSQL）通过 SQL 视图查询。上表"SQL 视图兼容"列为 ✅ 的操作符可安全使用；标 ⚠️ 或 ❌ 的仅在 OpenSearch 索引模式下可用，SQL 视图下会返回 400 或 500 错误。
>
> **常见错误**：`eq`、`gt`、`lt`、`gte`、`lte` **不是合法操作符**，会返回 400 InvalidParameter。正确写法是 `==`、`>`、`<`、`>=`、`<=`。

### 分页与大数据量查询

对于字段多或数据量大的对象类型，单次查询返回数据可能非常大。建议：

- **设置合理的 limit**：默认 limit=30，字段多的宽表建议 10~20
- **使用 `search_after` 游标分页**：首次查询返回 `search_after` 字段，传入下一次查询实现翻页

```bash
# 首次查询
ontology bkn object-type query <kn-id> <ot-id> '{"limit":20}'
# 返回 {"datas": [...], "search_after": ["val1","val2","val3"]}

# 翻页：将上次返回的 search_after 传入
ontology bkn object-type query <kn-id> <ot-id> '{"limit":20,"search_after":["val1","val2","val3"]}'
```

### 示例

```bash
# 数量大于 4000 的产品
ontology bkn object-type query <kn-id> <ot-id> '{"limit":30,"condition":{"field":"quantity","operation":">","value":4000}}'

# 名称模糊匹配"手机"且价格 >= 1000
ontology bkn object-type query <kn-id> <ot-id> '{"limit":30,"condition":{"operation":"and","sub_conditions":[{"field":"name","operation":"like","value":"%手机%"},{"field":"price","operation":">=","value":1000}]}}'

# 状态为 active 或 pending（用 in）
ontology bkn object-type query <kn-id> <ot-id> '{"limit":30,"condition":{"field":"status","operation":"in","value":["active","pending"]}}'
```

> **注意**：不支持 `{"field": {">": value}}` 这种简写语法，必须使用 `{"field": ..., "operation": ..., "value": ...}` 结构。

## Object-type query strategy for LLM and Agent

**所有 `object-type query` 调用必须遵守以下规则，否则返回数据过大会导致 JSON 截断和解析失败：**

1. **limit 不要超过 30**。CLI 默认 limit=30。字段多的宽表建议 limit 10~20，按实际单条记录大小调整
2. **需要更多数据时使用 `search_after` 分页**，不要加大 limit：
   ```
   # 第一页
   ontology bkn object-type query <kn> <ot> '{"limit":20}'
   # → 返回 search_after: ["v1","v2","v3"]
   # 第二页
   ontology bkn object-type query <kn> <ot> '{"limit":20,"search_after":["v1","v2","v3"]}'
   ```
3. **尽量使用 `condition` 过滤**，缩小返回集：按编号、名称、状态等精确或模糊过滤，避免全表扫描
4. **优先使用 `==`、`like`、`in` 操作符**（SQL 视图兼容）。`match`/`contain` 仅 OpenSearch 索引支持，SQL 视图下会报错
5. **自动裁剪**：当查询结果超过 100KB 时，CLI 会自动裁剪 `datas` 数组并附加 `_truncated` 字段。如果返回中包含 `_truncated`，按其中的 `hint` 提示执行下一轮查询：
   ```json
   {"_truncated": {"returned": 176, "total_fetched": 200, "remaining": 24,
     "next_search_after": ["v1","v2","v3"],
     "hint": "Pass --search-after '[...]' --limit 176 to fetch the next page."}}
   ```
   直接使用 `next_search_after` 的值作为下一次查询的 `--search-after` 参数即可

## Relation Type

```bash
ontology bkn relation-type list <kn_id>
ontology bkn relation-type get <kn_id> <rt_id>
ontology bkn relation-type create <kn_id> --name <n> --source <ot_id> --target <ot_id> [--mapping src:tgt ...]
ontology bkn relation-type update <kn_id> <rt_id> [--name <n>]
ontology bkn relation-type delete <kn_id> <rt_ids> [--yes/-y]
```

## Subgraph

```bash
ontology bkn subgraph <kn_id> '<json>'   # 子图查询
```

## Action Type / Log / Execution

```bash
ontology bkn action-type list <kn_id>
ontology bkn action-type query <kn_id> <at_id> '<json>'
ontology bkn action-type execute <kn_id> <at_id> '<json>'   # 有副作用，执行前确认
ontology bkn action-execution get <kn_id> <execution_id> [--wait/--no-wait] [--timeout 300]
ontology bkn action-log list <kn_id> [--offset 0] [--limit 20] [--sort create_time] [--direction desc]
ontology bkn action-log get <kn_id> <log_id>
ontology bkn action-log cancel <kn_id> <log_id> [--yes/-y]
```

## 端到端示例

```bash
# 接入数据源 → 创建 KN → 查询
ontology ds connect mysql db.example.com 3306 erp --account root --password pass
ontology bkn create-from-ds <ds_id> --name "erp-kn" --build
ontology bkn object-type list <kn_id>
```
