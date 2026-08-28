# `ds import-csv`：经 data-flow 写入物理表

唯一允许的写入通道。前置步骤（resolve-target、csv-prepare、主键差集、用户确认）全部通过后才能调用。

## 命令形态

```bash
ontology --user-id <accountId> ds import-csv <datasource_id> \
  --table-name <physical_table> \
  --file <csv_path> \
  [--batch-size <n>]    # 默认 100
```

参数对应到前面三段解析的结果：

| 参数 | 来源 |
|------|------|
| `<datasource_id>` | `dataview get` 返回的 `datasource_id` |
| `--table-name`    | `dataview.meta_table_name` 的尾段（去掉 catalog/schema 包裹），通常等于 `dataview.name` |
| `--file`          | csv-prepare 输出的 **差集后** 临时 CSV，不是用户原始文件 |
| `--batch-size`    | ≤ 1000；CSV 总行数 > 1000 必须分段 |

> **严禁** 出现 `mysql -h ... -e "INSERT ..."` / `pymysql.connect(...)` / 任何直接 JDBC/SQL 客户端调用作为本步的"替代"或"补救"。
> 任何绕过 `ds import-csv` 的写入都视为违规，停止并报错。

## HTTP debug：验证走的是 data-flow，不是直连

```bash
ONTOLOGY_DEBUG_HTTP=1 ontology --user-id 1 ds import-csv 374c4c1b-8836-4e60-8099-572a1f0367c5 \
  --table-name product_entity --file test_dedup.csv --batch-size 100
```

期望日志包含：

```
[debug] POST http://vega-bkn-backend:13014/api/automation/v1/data-flow/flow
[debug] headers: {"accept":"application/json, text/plain, */*", ... "content-type":"application/json"}
[debug] body (first 300): {"title":"import-csv-product_entity-<ts>", ...
        "steps":[{"id":"step-trigger", ...}, {"id":"step-write","title":"Write to Database", ...}], ...}
```

如果日志命中的不是 `vega-bkn-backend:.../data-flow/flow`、或出现 `mysql://` / `jdbc:` 字样，**立即终止**并报错——说明 CLI 链路或环境被改动到非平台通道。

## 成功回执

```json
{
  "tables": ["product_entity"],
  "failed": [],
  "summary": {"succeeded": <n>, "failed": 0}
}
```

校验点：
- `summary.succeeded == 准备写入的行数（差集后）`
- `summary.failed == 0`
- `tables[]` 包含目标表名

任一项不满足均按失败处理。

## 失败回执：主键冲突（最常见）

预检遗漏 / CSV 与底表中间被并发写入时，会出现：

```json
{
  "tables": [],
  "failed": ["product_entity"],
  "summary": {"succeeded": 0, "failed": 1}
}
```

错误流（来自 CLI 标准错误）形如：

```
[product_entity] batch 1/1 error: Dataflow run failed: {
  "error_type": "insert_failed",
  "message": "insert failed on row 0: Error 1062 (23000): Duplicate entry 'P0009' for key 'product_entity.PRIMARY'",
  "details": {
    "original_batch": "Error 1062 (23000): Duplicate entry 'P0009' for key 'product_entity.PRIMARY'",
    "per_row_error": "Error 1062 (23000): Duplicate entry 'P0009' for key 'product_entity.PRIMARY'",
    "row_index": 0,
    "rows_written": 0,
    "table": "product_entity"
  }
}
```

处理规则：
1. **禁止自动重试整批**。重试只会复现 1062。
2. 把 `details.row_index`、冲突的主键值、`per_row_error` 完整回显给用户。
3. 提示路径：① 回到 csv-prepare 重做差集；② 用户改主键值；③ 用户明确要求"修改已有行"时，告知本通道不支持，请走业务流程。
4. 不要尝试 `dataview query --sql "UPDATE ..."` / `--raw-sql` 之类的 fallback——`dataview query` 在本部署只允许 `SELECT/WITH`。

## 其他失败模式（速查）

| 现象 | 含义 | 处理 |
|------|------|------|
| `failed[]` 非空但 `per_row_error` 是类型错误 | CSV 某列与字段类型不匹配（如日期格式） | 回到 csv-prepare 修正格式后重写差集 CSV |
| HTTP 401/403 | 用户无权限或 `--user-id` 错误 | 直接报错；不要尝试登录或换 token |
| HTTP 500 + `data-flow` 报错 | 平台 data-flow 服务异常 | 报错并交给用户/运维；不要降级到直连 |
| `tables` 为空且 `summary` 全 0 | CSV 文件为空或路径错误 | 检查 `--file` 是否指向差集后的临时 CSV，且行数 > 0 |

## 写入后

- 索引刷新：BKN 对象类的 ES 索引由 data-flow 异步更新，可能存在短暂"写完但查不到"的窗口；如需立刻可查，提示用户由上层触发索引重建（不属于本 skill 职责）。
- 建议在回复里同时给出：差集统计、写入回执、目标表 + dataview，方便用户复查。
