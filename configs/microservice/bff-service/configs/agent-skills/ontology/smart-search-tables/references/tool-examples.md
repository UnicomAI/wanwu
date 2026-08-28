# 找表编排：端到端示例（2 步 + 总结）

> **执行委托**：本文件仅展示命令形态供 [ontology-core](../../ontology-core/SKILL.md) 执行；**Never** 由 smart-search-tables 直接执行 `ontology` CLI。
>
> **角色分工**：
> - **smart-data-analysis** — 顶层意图、选定 `kn_id` / `duty_kn_id` / `<ot-id>` / 提供 `accountId`
> - **smart-search-tables**（本 skill） — 步骤顺序与总结口径
> - **ontology-core** — 实际执行 ontology 子命令

下面以"采购订单宽表在哪个库、叫什么，谁在数据治理里负责？"为例。逐步详细约束见 [metadata-search.md](metadata-search.md) / [duty-search.md](duty-search.md)。

约定：`<X>` 为账户 id（必传 `--user-id`，由调用方提供；不得编造）。假定上游已选定 `<MK>` = `kn_id`、`<MT>` = 元数据对象类 id、`<DK>` = `duty_kn_id`、`<DT>` = 职责对象类 id。

## 1. 元数据 KN 检索：找候选表/视图

```bash
ontology --user-id <X> bkn object-type query <MK> <MT> \
  '{
    "limit": 100,
    "need_total": true,
    "properties": ["embeddings_text"],
    "condition": {
      "operation": "or",
      "sub_conditions": [
        {"field": "embeddings_text", "operation": "match", "value": "采购订单 宽表 销售域"},
        {"limit_value": 1000, "limit_key": "k", "field": "embeddings_text", "operation": "knn", "value": "采购订单 宽表 销售域"}
      ]
    }
  }'
```

整理候选（按 `_score` 由高到低取 Top 几条）：

```text
候选表/视图：
  - business_name: 采购订单宽表（销售域）   technical_name: dwd_sales_purchase_order_wide
      所属部门: 数据管理部   主题域: 销售
  - business_name: 采购订单明细表           technical_name: ods_sales_purchase_order_detail
      所属部门: 信息技术部   主题域: 采购
```

派生职责检索的关键词：**数据管理部 / 信息技术部 / 采购 / 销售域**。

## 2. 职责 KN 检索：找相关部门职责

```bash
ontology --user-id <X> bkn object-type query <DK> <DT> \
  '{
    "limit": 50,
    "need_total": true,
    "properties": ["embeddings_text"],
    "condition": {
      "operation": "or",
      "sub_conditions": [
        {"field": "embeddings_text", "operation": "match", "value": "数据管理部 采购 数据治理 职责"},
        {"limit_value": 500, "limit_key": "k", "field": "embeddings_text", "operation": "knn", "value": "数据管理部 采购 数据治理 职责"}
      ]
    }
  }'
```

整理职责条目：

```text
职责命中：
  - 部门: 数据管理部   动作: 维护 / 治理   范围: 销售域、采购主题数据资产
  - 部门: 信息技术部   动作: 管理 / 维护   范围: 企业数据资产目录、指标口径
```

> 若 `<DK>` 未提供 → 跳过本步；在总结里写"未检索职责"。

## 3. 总结（中文回复结构）

回复中 **必须包含**：

- **候选表/视图**（以 `business_name` 全称为主，附 `technical_name`；缺失标注"暂无"）：

  | 业务名（business_name） | 技术名（technical_name） | 所属部门 | 主题域 |
  |------------------------|--------------------------|----------|--------|
  | 采购订单宽表（销售域）  | dwd_sales_purchase_order_wide | 数据管理部 | 销售 |
  | 采购订单明细表          | ods_sales_purchase_order_detail | 信息技术部 | 采购 |

- **相关部门职责**（部门 + 动作 + 范围）：

  | 部门 | 职责动作 | 适用范围 |
  |------|----------|----------|
  | 数据管理部 | 维护 / 治理 | 销售域、采购主题数据资产 |
  | 信息技术部 | 管理 / 维护 | 企业数据资产目录、指标口径 |

- **关系说明**：能直接对齐时写出"部门 → 资产"；不能时如实标注"未在职责库中直接关联"。
- **下一步建议**：
  - 用户要取数 → 转 [smart-ask-data](../../smart-ask-data/SKILL.md)
  - 想换 KN → 回到 [smart-data-analysis](../../smart-data-analysis/SKILL.md)
- **不暴露完整调试 URL**。

若结果为空，直接说"未查询到符合条件的数据"，并提示放宽 `search` / 换 KN / 二次澄清；**不得**编造。
