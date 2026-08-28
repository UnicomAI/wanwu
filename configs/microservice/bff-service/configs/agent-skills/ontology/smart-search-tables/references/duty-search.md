# 步骤 2：职责 KN 检索（相关部门职责）

在 **找表主流程第 2 步** 调用：基于第 1 步得到的 **部门线索 / 主题域 / 表候选**，在职责型 KN 下检索职责与治理边界，用于回答「谁负责、职责是什么、如何协同」。

> **执行委托**：本文件仅 **描述** 命令形态；**Never** 由 smart-search-tables 直接执行 `ontology` CLI，实际调用统一由 [ontology-core](../../ontology-core/SKILL.md) 完成。

## 与本流程的衔接

- **输入**：`accountId`、`duty_kn_id`、`ot-id`、由步骤 1 派生的检索词或 condition。
- **输出**：相关部门职责条目（部门、职责动作、适用范围），与步骤 1 的表候选合并后给出总结。
- **可跳过**：若上游未提供 `duty_kn_id`，跳过本步并在总结中说明"未检索职责"。

## 派生检索词

从步骤 1 结果中提取以下信号合成 `<search>`：

- 主线索：**部门名**（最优先；如"数据管理部"、"信息技术部"）
- 辅线索：**主题域 / 业务对象**（如"采购订单"、"指标口径"）
- 业务语境词（如"治理"、"管理"、"维护"、"协同"）

示例：`「数据管理部 采购主题 数据治理 职责」`、`「信息技术部 企业数据资产 指标口径 管理」`。

若步骤 1 无部门线索：用用户原问题中的部门/组织词，或先 **简要反问** 用户「这条数据治理的归口部门是？」。

## 命令形态（委托 ontology-core 执行）

### 主路径 A：实例检索（推荐）

```bash
ontology --user-id <accountId> bkn object-type query <duty_kn_id> <ot-id> \
  '<condition-json>' [--limit 50] [-bd bd_public] [--pretty]
```

`<condition-json>` 沿用 ontology-query API 的 condition 体，常用 `match` + `knn` 的 `or` 组合，字段以职责 KN 的实际属性为准（常见为 `embeddings_text` 之类的语义文本字段）：

```json
{
  "limit": 50,
  "need_total": true,
  "properties": ["embeddings_text"],
  "condition": {
    "operation": "or",
    "sub_conditions": [
      { "field": "embeddings_text", "operation": "match", "value": "<search>" },
      { "limit_value": 500, "limit_key": "k", "field": "embeddings_text", "operation": "knn", "value": "<search>" }
    ]
  }
}
```

- **结构约束**：与步骤 1 同构；仅允许变动 `condition.sub_conditions[*].value` 与 `limit` / `properties`。
- **`<ot-id>`**：不确定时先 `ontology --user-id <accountId> bkn object-type list <duty_kn_id>` 列出，由 smart-data-analysis 选定。

## 响应处理示意

抽取 **部门 / 职责动作 / 适用范围 / 与具体表（或主题）的关系**；去重后与第 1 步候选表/视图对照。

示例字段：

| 字段 | 含义 |
|------|------|
| 部门 | 职责主体 |
| 职责动作 | "维护"、"管理"、"治理"、"审核" 等 |
| 适用范围 | 主题域 / 业务对象 / 数据类型 |
| 关联资产（如有） | 与步骤 1 候选表的对齐说明 |

## 总结合并要点

- **事实发现**（来自步骤 1）：候选表/视图清单（`business_name` 全称 + `technical_name`）。
- **治理描述**（来自步骤 2）：相关部门 + 职责动作 + 范围。
- **关系说明**：能直接对齐时写出"部门 → 资产"；不能时如实标注"未在职责库中直接关联"。
- **空结果**：本步为空时，先尝试改写 `<search>`（补部门名 / 主题词）；仍空则在总结中说明"未检索到对应职责"，不编造。

## 注意事项

- `--user-id <accountId>` **必传**。
- 命令体内 **不出现** `--token` / `auth.token` / `Authorization`（本部署 ontology CLI 无须 token）。
- 命令报错时直接如实反馈用户，不要伪造结果或尝试登录刷新凭证。
- **`duty_kn_id` 与 `kn_id` 不可混用**：两个 KN 通常服务于不同语义；smart-data-analysis 在 KN 选定阶段已做区分。
