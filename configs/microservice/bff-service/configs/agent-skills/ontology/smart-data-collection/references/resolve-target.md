# 写入目标解析：object-type → dataview → datasource

数据采集的第一段工作是 **解出物理写入目标**。三段必须全部成功，缺一不可。

## 步骤 1：从对象类拿 dataview-id

```bash
ontology --user-id <accountId> bkn object-type get <kn-id> <ot-id>
```

关键字段：

| 字段路径 | 用途 |
|----------|------|
| `data_source.type` | 必须为 `data_view`，否则该对象类不支持本通道写入 |
| `data_source.id` | **下一步要用的 dataview_id** |
| `data_source.name` | 人类可读名（用于复述计划时回显） |
| `primary_keys[]` | 主键列；CSV 中不可为空 |
| `data_properties[].name` | 与 CSV 表头对齐的字段名 |
| `data_properties[].mapped_field.name` | dataview 侧对应字段名（通常一致） |

> **拒绝条件**：`data_source.type` 不是 `data_view`、或 `primary_keys` 为空 — 直接告知用户当前对象类无法通过 `import-csv` 通道写入，不要尝试别的写入方式。

## 步骤 2：从 dataview 拿 datasource-id 与物理表

```bash
ontology dataview get <dataview_id>
```

关键字段：

| 字段路径 | 用途 |
|----------|------|
| `datasource_id` | **下一步 `ds get` / `ds import-csv` 要用的 id** |
| `data_source_type` | mysql / postgresql / ... — 用于风险提示（仅提示，不改变通道） |
| `query_type` | 若为 `SQL` 而非原子表，需校验 `sql_str` 是否对应单张物理表，否则拒绝写入 |
| `meta_table_name` | 物理表名（含 catalog/schema），写入命令的 `--table-name` 必须落到这里指向的表 |
| `fields[].name` | CSV 表头需要与之对齐（按名匹配，大小写敏感） |

> **拒绝条件**：`query_type=SQL` 且 `sql_str` 含 JOIN / 子查询 / 视图 — 这种 dataview 是聚合查询，不是写入目标。

## 步骤 3：datasource 只读校验

```bash
ontology ds get <datasource_id>
```

关键字段：

| 字段路径 | 用途 |
|----------|------|
| `type` | 与 dataview 的 `data_source_type` 一致 |
| `bin_data.catalog_name` / `database_name` | 与 `meta_table_name` 对齐 |
| `latest_task_status` | 非 `success` 时提示用户：数据源最近一次任务异常，写入仍可尝试但需关注回执 |
| `operations[]` | 必须包含写入相关权限（如 `modify`/`scan`），否则拒绝 |

> **严禁**：
> - 读取或解密 `bin_data.password`；
> - 用 `host` / `port` / `account` 自行拼 JDBC URL；
> - 把这些字段贴到日志或回复里。

## 解析结果交接给下一阶段

把以下结构体（不含密码字段）交给 CSV 准备阶段：

```json
{
  "kn_id": "...",
  "object_type_id": "...",
  "dataview_id": "...",
  "datasource_id": "...",
  "data_source_type": "mysql",
  "meta_table_name": "mysql_jtjlumy4.\"demo\".\"product_entity\"",
  "table_name_for_import": "product_entity",
  "primary_keys": ["product_code"],
  "fields": ["product_name", "status", "product_code", "..."]
}
```

`table_name_for_import` 是 `--table-name` 实参，通常等于 `meta_table_name` 的最后一段（去掉 catalog/schema 包裹）。
