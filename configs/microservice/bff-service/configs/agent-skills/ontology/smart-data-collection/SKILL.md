---
name: smart-data-collection
version: "1.0.0"
user-invocable: true
description: >-
  智能数据采集技能，用于从图片或文档（PDF、Word、Excel）中提取结构化数据，基于知识网络完成字段映射，生成SQL并写入数据库。当用户提到"数据采集"、"从文档提取数据"、"图片转数据"、"数据导入"、"文档数据入库"、"批量数据提取"或需要从非结构化文件中提取结构化数据并存储时，自动使用此技能。
---

# Smart Data Collection（数据采集 / 写入）

本 skill 定义 **固定顺序** 的数据写入工具链，完全由 `ontology` **native 子命令**（`bkn` / `dataview` / `ds`）实现；
所有 CLI 执行均委托 [ontology-core](../ontology-core/SKILL.md)。

## 安全红线（IRON RULE）

1. **严禁直连数据库**：不得使用 `mysql` / `mysqldump` / `pymysql` / `sqlalchemy+pymysql` / 直接 JDBC 连接等方式向 `data_source` 写入数据。
   - `ontology ds get` 返回的 `bin_data.password` 是 **平台加密串**，不是明文密码；尝试解密或直连均视为违规。
   - 直连绕过平台审计、权限、字段映射、索引刷新与事务一致性，会导致 BKN 索引与底表脱节。
2. **唯一允许的写入入口**：`ontology ds import-csv <datasource_id> --table-name <table> --file <csv> [...]`
   - 由 CLI 调用后端 `POST /api/automation/v1/data-flow/flow`，由平台 data-flow 服务统一执行写入与回执。
3. **`import-csv` 仅支持 INSERT，不支持 UPDATE/UPSERT/REPLACE**：
   - 每行执行一次 `INSERT`；若主键已存在，MySQL 直接报 `Error 1062 (23000): Duplicate entry '<pk>' for key '<table>.PRIMARY'`，整批失败、`summary.failed=1`、`rows_written=0`。
   - **导入前必须先查"哪些主键已存在"，把这些行从 CSV 剔除**；不要假设 data-flow 会自动覆盖或合并。
   - 修改/删除已有行：本通道不支持，直接拒绝并告知用户。
4. **写操作需明确确认**：导入前必须先复述本次将写入的 `kn / object_type / dataview / datasource / table_name / row_count`，得到用户确认后再执行。
5. **不得用 `dataview query --sql` / `--raw-sql` 执行 INSERT/UPDATE/DELETE**：该入口在本部署只允许 `SELECT/WITH`，不是写入通道。

## 调用方式（统一 ontology 命令；委托 ontology-core 执行）

```
smart-data-collection（本 skill：定义写入工具链与顺序）
  └─ ontology-core（实际执行 ontology 命令）
```

```
ontology --user-id <accountId> <command> [options]
```

- **`--user-id <accountId>`**：必传（顶层选项，写在子命令之前）。
- 网关 / `ONTOLOGY_BASE_URL` 由 ontology-core 承担。
- 本部署 CLI **无 token**；本 skill 不出现 `--token` / `Authorization`。
- `-bd bd_public`：默认值，可省。

## Phase 0：非 CSV 输入的前置解析（图片 / PDF / Word / Excel）

当用户给的不是结构化 CSV，而是 `.jpg / .jpeg / .png / .bmp / .pdf / .doc / .docx / .xls / .xlsx`，必须先经 Phase 0 转成 CSV，再进入下面的第 1～10 步主流程。Phase 0 由本 skill 编排，**绝不**新增「直接生成 INSERT SQL」或「直连 mysql」通道，IRON RULE 全部保留。

- **默认 KN**：`e0e2ed66-8eff-4f14-8e15-9fd6171f8e53`；用户指定时优先。参数名统一 `kn_id`。
- **Phase 0 子步骤**：
  - **0a. 文件 → Markdown**
    - 图片：通过 Skill 工具显式调用 `yj-ocr-parser` skill（契约见 `skills/yj-ocr-parser/SKILL.md`），输出 Markdown（含表头、表格行）。
    - 文档（PDF / Word / Excel）：调用文档解析工具（具体工具**待集成**），输出统一 Markdown 表 + 字段样值。
    - 完整契约见 [references/parse-input.md](references/parse-input.md)。
  - **0b. 字段语义映射 + 用户逐条确认**
    - 委托 ontology-core 调 `bkn object-type query/get` 拿 `data_properties[]`（字段名 / 类型 / required / enum），作为映射目标的**唯一字段定义来源**。
    - 名称不一致 / 类型不一致 / 必填缺失 / 歧义 任一触发 → **必须停下来给用户出对照表，再一次只就一条差异项单独确认**，拿到该条决策才问下一条；**严禁把多条揉成一个 `是否确认以上映射？` 的总确认**。未确认前禁止写出 CSV。
    - 完整契约见 [references/map-fields.md](references/map-fields.md)。
  - **0c. 写出 CSV**
    - 按用户确认后的映射结果落成 **UTF-8 + RFC 4180** 的 CSV；列名 = **映射后字段名**（与 `dataview.fields[].name` 对齐，便于第 4 步无歧义消费）。
    - 输出文件路径作为后续第 4 步 CSV 准备的输入。
- **Phase 0 边界**：
  - 不生成任何 `.sql` 文件、不生成 `INSERT/UPDATE/DELETE` 语句、不连 mysql、不读 `bin_data.password`。
  - 用户给的就是 CSV 时，**跳过整个 Phase 0**，直接从下面第 1 步开始。

## 必读 references（按步骤）

| 步骤 | 说明 | Reference |
|------|------|-----------|
| 0a | 非 CSV 输入 → Markdown（图片和pdf文档走 yj-ocr-parser；其他文档走解析工具） | [references/parse-input.md](references/parse-input.md) |
| 0b | object-type 字段语义映射 + 用户逐条确认 → CSV | [references/map-fields.md](references/map-fields.md) |
| 1 | 从对象类解出 dataview-id | [references/resolve-target.md](references/resolve-target.md) |
| 2 | 从 dataview 解出 datasource-id 与物理表名 | [references/resolve-target.md](references/resolve-target.md) |
| 3 | CSV 列与字段对齐 + 主键预检（差集） | [references/csv-prepare.md](references/csv-prepare.md) |
| 4 | 通过 `ds import-csv` 走 data-flow 写入 | [references/import-csv.md](references/import-csv.md) |

## 主流程（必须按序）

复制进度：

```text
数据采集进度：
- [ ] 0a. （仅非 CSV 输入）解析输入文件 → Markdown：图片和pdf文档走 yj-ocr-parser；其他文档（Word/Excel等）走解析工具
- [ ] 0b. （仅非 CSV 输入）取 object-type data_properties → 与解析字段做语义映射 → 出对照表给用户逐条确认
- [ ] 0c. （仅非 CSV 输入）按确认后的映射写出 CSV（UTF-8、RFC 4180、列名 = 映射后字段名）
- [ ] 1. 解析写入目标对象类：bkn object-type get <kn-id> <ot-id>
       → 拿 data_source.id（dataview_id）、primary_keys、data_properties
- [ ] 2. 解 dataview：ontology dataview get <dataview_id>
       → 拿 datasource_id、data_source_type、meta_table_name、fields
- [ ] 3. 解 datasource：ontology ds get <datasource_id>
       → 校验 type/connect 信息，**只读校验**；不取出 bin_data.password
- [ ] 4. 准备 CSV：列名与 fields 对齐；UTF-8；主键列不可空；行数与 batch-size 估算
- [ ] 5. 主键预检：ontology --user-id <id> dataview query <dataview_id> --sql 'SELECT <pk> FROM <stripped_meta_table_name>'
       （或加 WHERE <pk> IN (...) 缩小范围）→ 取出底表已存在的主键集合
- [ ] 6. 差集：从 CSV 中剔除上一步返回的主键行；若剔除后 rows=0，直接终止并告知用户"全部已存在"
- [ ] 7. 复述写入计划并征得用户确认：kn / object_type / dataview / datasource / table / rows
- [ ] 8. 执行写入：ontology ds import-csv <datasource_id> --table-name <table> --file <csv> --batch-size <n>
- [ ] 9. 回执核对：summary.succeeded/failed；如 failed>0，定位失败行并报错，不要重试整批
- [ ] 10. （可选）触发/等待索引：必要时由上层重建 BKN 索引以保证查得到新数据
```

### 步骤约束（摘要）

1. **目标解析三段式**：`object-type → dataview → datasource`，三段必须全部成功才能进入写入。
2. **物理表名来源**：仅以第 2 步 `dataview.meta_table_name`（或 `fields`/`sql_str` 中显式声明的表）为准；**不得**让用户口头给的表名直接生效。
3. **字段映射**：CSV 表头必须能与 `dataview.fields[].name` 一一对齐；多余列丢弃前先告警，缺失主键列直接拒绝。
4. **批次大小**：默认 `--batch-size 100`；超过 1000 行的导入必须分批，回执按批汇总。
5. **主键预检 SQL 注意事项**：
   - ① **`dataview query` 位置参数用 dataview UUID**，SQL 内的表名用 `meta_table_name` **去掉所有双引号**后的形式（例：`mysql_jtjlumy4.demo.product_entity`）。
   - ② SQL 用 **单引号** 包裹整体传给 `--sql`；**任何位置都不要出现反斜杠**——曾出现 `product_entity\ LIMIT 50` 触发 sqlglot `ExtractTables failed` 的事故。
   - ③ 主键值较多时分批 `IN(...)`，每批 ≤ 500；总量 > 10000 时直接 `SELECT <pk> FROM <table>` 全量拉回客户端做差集，避免 `IN` 列表过长。
6. **HTTP debug**：诊断时可加 `ONTOLOGY_DEBUG_HTTP=1`，确认请求落到 `/api/automation/v1/data-flow/flow`；切勿误以为是直连 MySQL。
7. **失败处理**：`failed[]` 中任意一条都要回显具体行号/原因；**禁止**自动重试，等待用户判断。

## 注意事项（必须遵守）

1. 严禁直连 MySQL/任何数据库；唯一写入入口为 `ontology ds import-csv`。
2. 严禁尝试解密 `ds get` 返回的 `bin_data.password`。
3. 严禁用 `dataview query --sql` 执行 INSERT/UPDATE/DELETE。
4. 主键已存在的行必须在客户端剔除后再 `import-csv`；不得依赖通道侧合并。
5. 写入前必须复述计划并由用户确认；不得"看起来差不多"就直接落库。
6. 回执必须忠实展示 `summary` 与 `failed[]`，不得粉饰失败。
7. 出现 401/403 等认证类错误，直接报错；不要尝试登录或刷新 token。
8. Phase 0（图片 / PDF / Word / Excel 输入）出现字段名 / 类型 / 必填 / 歧义任一不一致时，必须停下来给用户出对照表，并**一次只就一条差异项单独确认**（严禁把多条揉成一个「是否确认以上映射？」的总确认）；未确认前**禁止**写出 CSV、**禁止**进入第 1 步。Phase 0 不允许生成 INSERT SQL 或直连 mysql。


## 调用示例

```text
/smart-data-collection 把 test.csv 导入产品对象类（kn=d4rt3135s3q8va76m8fd, ot=d50jqrr5s3q8ofn0dscg）
/smart-data-collection 给"产品"对象类追加 200 行新产品数据，文件在 ~/uploads/products.csv
```

## 真实链路速记（示例输出对照）

```bash
# 1) 对象类 → dataview_id
ontology --user-id 1 bkn object-type get d4rt3135s3q8va76m8fd d50jqrr5s3q8ofn0dscg
# data_source.id = c1a934eb-7011-40f9-8a7c-c5ca6cb392a6

# 2) dataview → datasource_id + meta_table_name
ontology dataview get c1a934eb-7011-40f9-8a7c-c5ca6cb392a6
# datasource_id   = 374c4c1b-8836-4e60-8099-572a1f0367c5
# meta_table_name = mysql_jtjlumy4."demo"."product_entity"

# 3) datasource 只读校验（不要拿 bin_data.password 做任何事）
ontology ds get 374c4c1b-8836-4e60-8099-572a1f0367c5

# 4) 主键预检：dataview UUID 作为位置参数，SQL 内的表名是 meta_table_name 去引号后的形式
ontology --user-id 1 dataview query c1a934eb-7011-40f9-8a7c-c5ca6cb392a6 \
  --sql 'SELECT product_id FROM mysql_jtjlumy4.demo.product_entity'
# 返回底表已有的 product_id 列表；客户端用它和 CSV 做差集

# 5) 经平台 data-flow 写入（唯一允许的写入入口；仅 INSERT 已差集后的新行）
ONTOLOGY_DEBUG_HTTP=1 ontology ds import-csv 374c4c1b-8836-4e60-8099-572a1f0367c5 \
  --table-name product_entity --file test_dedup.csv --batch-size 100
# 期望命中：POST http://vega-bkn-backend:13014/api/automation/v1/data-flow/flow
# 主键重复（未预检）时回执形如：
#   {"failed":["product_entity"],
#    "summary":{"succeeded":0,"failed":1},
#    "error":"Error 1062 (23000): Duplicate entry 'P0009' for key 'product_entity.PRIMARY'"}
```
