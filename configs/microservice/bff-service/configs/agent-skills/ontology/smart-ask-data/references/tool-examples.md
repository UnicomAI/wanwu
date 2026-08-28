# 问数编排：端到端示例（5 步）

> **执行委托**：本文件仅展示命令形态供 [ontology-core](../../ontology-core/SKILL.md) 执行；**Never** 由 smart-ask-data 直接执行 `ontology` CLI。
>
> **角色分工**：
> - **smart-data-analysis**（LLM）—— 顶层意图、选 KN、生成 SQL
> - **smart-ask-data**（本 skill）—— 步骤顺序与口径约束
> - **ontology-core** —— 实际执行 ontology 子命令

下面以"上月各区域销售额，按区域汇总"为例，给出 5 步骨架。逐步详细约束见 [kn-resolve.md](kn-resolve.md) / [schema-discovery.md](schema-discovery.md) / [sql-execute.md](sql-execute.md)。

约定：`<X>` 为账户 id（必传 `--user-id`，由调用方提供；不得编造）。

## 1. 选 KN（条件执行）

若上下文 `kn_id` 已给出 → 跳过本步；否则由 **smart-data-analysis 的 LLM** 在编排层决策：

```bash
ontology --user-id <X> bkn list
# 对每个候选取详情：
ontology --user-id <X> bkn get <candidate-kn-id>
```

LLM 输出：`kn_id = d71o5e1e8q1nr9l7mb80`（假定）。

## 2. Schema 发现：候选对象类与字段

```bash
ontology --user-id <X> bkn object-type get d71o5e1e8q1nr9l7mb80 fact_sales_order
ontology --user-id <X> bkn object-type get d71o5e1e8q1nr9l7mb80 dim_region
```

整理 schema 摘要（交给步骤 3 LLM）：

```text
KN: d71o5e1e8q1nr9l7mb80
对象类：
  - fact_sales_order
      dataview-id: dv_fact_sales_order     # 取自 data_source.id（type == "data_view"）
      字段: order_id (string), region_id (string), order_month (date), amount (decimal)
  - dim_region
      dataview-id: dv_dim_region
      字段: region_id (string), region_name (string)
关系：fact_sales_order.region_id → dim_region.region_id
```

## 3. SQL 生成（编排层 LLM）

由 smart-data-analysis 基于上面 schema 摘要 + 用户问题生成 SQL（**不在本 skill 内**）：

```sql
SELECT r.region_name, SUM(o.amount) AS amount
FROM dv_fact_sales_order o
JOIN dv_dim_region r ON o.region_id = r.region_id
WHERE o.order_month = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
GROUP BY r.region_name
ORDER BY amount DESC
LIMIT 200
```

约束：只允许 SELECT/WITH；字段/表必须来自步骤 2 摘要；含聚合时带 `GROUP BY`；末尾或外层带 `LIMIT`。

## 4. 执行 SQL

```bash
ontology --user-id <X> dataview query dv_fact_sales_order \
  --sql "SELECT r.region_name, SUM(o.amount) AS amount FROM dv_fact_sales_order o JOIN dv_dim_region r ON o.region_id = r.region_id WHERE o.order_month = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') GROUP BY r.region_name ORDER BY amount DESC LIMIT 200" \
  --limit 200
```

- `dataview-id` 取主事实表对应的 dataview（这里 `dv_fact_sales_order`）。
- 引擎走 mdl-uniquery；默认拒绝写操作；**禁止** `--raw-sql`。
- 简单单表过滤可用 `bkn object-type query <kn-id> <ot-id> '<filter-json>'` 代替（不支持聚合 / JOIN）。

## 5. 总结

回复中 **必须同时** 包含：

- **执行的 SQL**（可脱敏，不可省略）。
- **关键结果数据**（表格 / 行记录 / 聚合数值）：

  | 区域 | 销售额 |
  |------|--------|
  | 华东 | 1,200,000 |
  | 华北 | 800,000 |
  | …    | … |

- **结论**（一两句业务口径，含时间范围）。
- **数据依据**：KN id / 对象类 / dataview-id；不暴露完整网关 URL。
- **限制**：未覆盖维度、采样、口径假设等。

若结果为空，直接说"未查询到符合条件的数据"，并给出下一步建议（时间范围 / 口径 / 换 KN），**不得**编造。
