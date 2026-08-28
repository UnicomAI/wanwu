# 步骤 1：元数据 KN 实例检索（找表/视图）

在 **找表主流程第 1 步** 调用：在元数据型 KN 下检索表/视图等实例，提取 `technical_name` / `business_name` 与归属部门、主题域线索，供步骤 2 派生职责检索词使用。

> **执行委托**：本文件仅 **描述** 命令形态；**Never** 由 smart-search-tables 直接执行 `ontology` CLI，实际调用统一由 [ontology-core](../../ontology-core/SKILL.md) 完成。

## 与本流程的衔接

- **输入**：`accountId`、`kn_id`、`ot-id`（每项均由上游 smart-data-analysis 提供）；用户问题提炼出的 `search` 短语。
- **输出**：表/视图候选清单，每条至少含 `technical_name` / `business_name` / `_score` / 归属部门 / 主题域。

## 命令形态（委托 ontology-core 执行）

### 主路径：实例检索

```bash
ontology --user-id <accountId> bkn object-type query <kn_id> <ot-id> \
  '<condition-json>' [--limit 100] [--search-after '<json-array>'] [-bd bd_public] [--pretty]
```

`<condition-json>` 是 ontology-query API 的 condition 体；对找表/找视图常用 `match` + `knn` 的 `or` 组合（语义+精确双路）：

```json
{
  "limit": 100,
  "need_total": true,
  "properties": ["embeddings_text"],
  "sort": [{ "direction": "", "field": "" }],
  "condition": {
    "operation": "or",
    "sub_conditions": [
      { "field": "embeddings_text", "operation": "match", "value": "<search>" },
      { "limit_value": 1000, "limit_key": "k", "field": "embeddings_text", "operation": "knn", "value": "<search>" }
    ]
  }
}
```

- **结构约束**：字段名、嵌套层级、布尔字面量保持同构；`include_logic_params` / `include_type_info` 若 KN 要求，用 JSON `false`（非字符串 `"false"`）。
- **仅允许变动的值**：
  - `condition.sub_conditions[*].value`（替换为 `<search>`）
  - 顶层 `limit`（建议 50–100；过宽下调）
  - `properties` 列表（按需增减返回字段）
- `--limit` 命令行参数与 condition 内 `limit` 都会被服务端尊重；冲突时以服务端实现为准，建议**保持一致**。

### 辅助路径

不确定 `<ot-id>` 时先列出该 KN 的对象类：

```bash
ontology --user-id <accountId> bkn object-type list <kn_id>
```

最终回到主路径的 `bkn object-type query` 走实例检索。

## 检索词（`<search>`）优化

- 包含**业务对象 + 主题域 + 关键名称片段**；避免单字、避免代词（"这个 / 那个"）。
- 例：「采购订单宽表 销售域」「示范企业信息 港务区」「招商引资 项目台账」。
- 命中过少：放宽限定词或改用同义近义词（如"客户" ↔ "企业"）。
- 命中过多：缩小到具体业务系统/部门/字段片段。

## 响应处理示意

`bkn object-type query` 返回结构通常含 `body.datas[]`（具体以 ontology-query API 为准）。每条命中典型字段：

```json
{
  "_score": 19.16,
  "_instance_id": "metadata-...",
  "embeddings_text": "视图技术名：gangwuqujianshefazhanbu_shifanqiyexinxi | 视图业务名：港务区建设发展部_示范企业信息 | 视图UUID：ec3e47b1-... | 所属部门：淮海国际港务区 | 关联信息系统： | 所属主题域：未分组 | ..."
}
```

**至少必须提取**：

| 抽取字段 | 来源（在 `embeddings_text` 中） |
|----------|-------------------------------|
| `technical_name` | "视图技术名" 或 "表技术名" 后的值 |
| `business_name` | "视图业务名" 或 "表业务名" 后的值（**完整全称**，不得截断） |
| `view_uuid` | "视图UUID" 或 "表UUID" 后的值（用作唯一标识） |
| 归属部门 | "所属部门" 后的值（供步骤 2 派生职责 query） |
| 主题域 | "所属主题域" 后的值（同上） |

按 `_score` 由高到低列 Top 候选；视图与表按同等关系处理（`view_*` ↔ `table_*` 等价归并为 `*_name`）。

## 注意事项

- `--user-id <accountId>` **必传**。
- 命令体内 **不出现** `--token` / `auth.token` / `Authorization`（本部署 ontology CLI 无须 token）。
- 命令报错时直接如实反馈用户，不要伪造结果或尝试登录刷新凭证。
- 结果为空：放宽 `search`、改 condition、或核对 `<kn_id>` 与 `<ot-id>` 是否匹配；不得编造命中。
