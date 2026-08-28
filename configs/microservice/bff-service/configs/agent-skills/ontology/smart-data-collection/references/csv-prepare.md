# CSV 准备、字段对齐与主键预检

`ds import-csv` 走平台 data-flow 通道写入，**仅支持 INSERT，不支持 UPDATE/UPSERT/REPLACE**。
任何与底表主键冲突的行都会让整个 batch 失败（MySQL `Error 1062`），所以在调用 `import-csv` 之前必须：
1. 把 CSV 列名与 `dataview.fields[].name` 对齐；
2. 通过 `dataview query` 拉出底表已有主键，与 CSV 做差集；
3. 复述写入计划，征得用户确认。

## 编码与格式

- UTF-8（无 BOM 优先；带 BOM 也能接受，但需在回执里提示一下）
- 首行必须是表头
- 分隔符：逗号 `,`；如源文件是 TSV / 分号分隔，先转换再导入
- 行尾换行符：`\n` 或 `\r\n` 均可
- 字符串字段如含逗号 / 引号，需按 RFC 4180 用 `"` 包裹并转义

## 表头与 dataview.fields 对齐

| 检查项 | 规则 |
|--------|------|
| 必需列 | `primary_keys[]` 中的每一列都必须出现在 CSV 表头里 |
| 命名 | 表头必须与 `dataview.fields[].name` 严格一致（大小写敏感） |
| 类型 | CSV 是文本流；对 `type=string` 字段无需转换；数值/日期型需符合下游 SQL 接受的字面量 |
| 多余列 | 不在 `fields[]` 中的表头列须告警；用户确认丢弃后再继续 |
| 缺失列 | 仅缺非主键列：可放行（写 NULL）；缺主键列：拒绝 |

## 主键预检（pre-existence dedup）

### 1. 用对的 SQL 形态

**`dataview query` 的位置参数是 dataview UUID**，SQL 内的表名用 `dataview.meta_table_name` **去掉所有双引号** 后的形式。

✅ 正确（已在生产环境验证）：

```bash
ontology --user-id <accountId> dataview query c1a934eb-7011-40f9-8a7c-c5ca6cb392a6 \
  --sql 'SELECT product_id FROM mysql_jtjlumy4.demo.product_entity'
```

```json
{
  "entries": [
    {"product_id": "P0001"},
    {"product_id": "P0002"},
    ...
    {"product_id": "P0010"}
  ],
  "vega_duration_ms": 145
}
```

当需要按 CSV 主键集缩小范围时（SQL 内含单引号字符串字面量，外层用双引号包裹即可）：

```bash
ontology --user-id <accountId> dataview query c1a934eb-7011-40f9-8a7c-c5ca6cb392a6 \
  --sql "SELECT product_id FROM mysql_jtjlumy4.demo.product_entity WHERE product_id IN ('P0001','P0002','P9999')"
```

### 2. 不要这样写

| 反例 | 失败原因 |
|------|----------|
| `... FROM mysql_jtjlumy4.demo.product_entity\ LIMIT 50` | 表名与 `LIMIT` 之间出现反斜杠（多来自手工 / JSON 转义），sqlglot 直接 `ExtractTables failed, Invalid expression / Unexpected token`。**SQL 字符串里任何位置都不要出现 `\`。** |
| `... FROM mysql_jtjlumy4."demo"."product_entity"` | 保留了 `meta_table_name` 原始的双引号，**必须全部去除**。 |
| `... FROM "product_entity"` | 丢了 `catalog.schema` 前缀，查询解析不到表。 |
| `... FROM <dataview_uuid>` | UUID 不是合法 SQL 标识符，不能直接当表名。 |
| `INSERT INTO ... ` / `UPDATE ... ` / `DELETE FROM ...` | `dataview query` 只允许 `SELECT/WITH`，写操作必须走 `ds import-csv`。 |

### 3. IN 列表分批

- 单批 `IN(...)` 主键值数量 ≤ 500，多于此分多次查询，客户端汇总结果。
- CSV 总行数 > 10000 时直接 `SELECT <pk> FROM <table>`（不带 `WHERE`）全量拉回，在客户端做差集，避免拼超长 `IN` 子句。

### 4. 差集

```
pk_in_csv      = {从 CSV 读出的所有主键值}
pk_in_table    = {上一步 SELECT 返回的主键值}
pk_to_write    = pk_in_csv - pk_in_table
```

- `pk_to_write` 为空：**直接终止**，告知用户"所有主键均已存在；本通道不支持更新"。不要进入 `import-csv`。
- 非空：把 `pk_to_write` 对应的行写到一个临时 CSV（例如 `test_dedup.csv`），后续 `import-csv` 用这个文件。

## 行数与批次

- `--batch-size`：默认 100，最大建议 1000
- 总行数 > 10000：必须分段（多次调用 `import-csv` 或单次大文件按 batch 内部分批），并把每段的 `summary` 都展示给用户
- 行数预估方法：客户端读 CSV 行数（不算表头）；与回执 `summary.succeeded + summary.failed` 对齐

## 复述写入计划（向用户确认）

进入实际写入前必须打印类似如下结构，由用户确认：

```text
即将写入：
  KN:           <kn_id>
  Object Type:  <ot_id> (<name>)
  Dataview:     <dataview_id> (<name>)
  Datasource:   <datasource_id> (<name>, type=<mysql/...>)
  Table:        <table_name_for_import>
  CSV File:     <abs_path>            （差集后的临时 CSV，不是原始文件）
  Rows:         <n>                   （= pk_to_write 行数）
  Skipped:     <m> 行因主键已存在被剔除
  Batch Size:   <batch_size>
通道：ontology ds import-csv（平台 data-flow，不直连数据库）
```

用户未明确同意之前不得调用 `ontology ds import-csv`。
