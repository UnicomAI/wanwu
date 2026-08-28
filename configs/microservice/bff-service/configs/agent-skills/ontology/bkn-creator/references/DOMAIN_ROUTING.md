# 领域识别路由表（对象类/关系类提取）

## 使用方式（评分制）

1. 从输入文本中提取关键词、专业缩写、流程节点
2. 依据“关键词分级权重”给各领域累计得分
3. 计算“原始分 / 领域满分 / 归一化得分（0-100）”
4. 按“阈值与判定规则”决定：直接命中 / 需确认 / 走通用流程

## 关键词分级权重

- `high_signal`：3 分（强领域词，如专有单据、专有缩写）
- `mid_signal`：2 分（领域常见词）
- `low_signal`：1 分（弱语义词，需与其他词共现）

> 同一关键词在同一轮仅计分一次；同义词视为同一命中项。

## 满分定义（解决“分数高低”歧义）

- 领域满分 = 该领域所有关键词全部命中时的理论最高分
- 全局满分 = 各领域满分中的最大值（用于快速参考）
- MUST 同时回显：
  - `raw_score`（原始分）
  - `domain_max_score`（该领域满分）
  - `normalized_score`（归一化百分比，`raw_score / domain_max_score * 100`）

## 领域映射

| 领域 | high_signal（3分） | mid_signal（2分） | low_signal（1分） | 参考文件 |
|------|--------------------|-------------------|-------------------|---------|
| `supply_chain` | MRP、BOM、PO、PR、MPS、齐套、缺料 | 采购、供应商、库存、到货、入库、领料、提前期 | 计划、排程、物料、工单、交期 | `./supply_chain/domain_supply_chain.md` |
| `crm_sales` | 线索、商机、报价单、赢单、回款、销售漏斗 | 客户、合同、商机阶段、签约、续约、订单转化 | 销售、跟进、客户经理、账户 | `./crm_sales/domain_crm_sales.md` |
| `project_delivery` | 里程碑、交付物、阻塞、依赖任务、工时记录 | 项目、任务、排期、资源、风险、问题单 | 交付、延期、验收、阶段完成 | `./project_delivery/domain_project_delivery.md` |

## 领域理论满分（当前词表）

- `supply_chain`：`7*3 + 7*2 + 5*1 = 40`
- `crm_sales`：`6*3 + 6*2 + 4*1 = 34`
- `project_delivery`：`5*3 + 6*2 + 4*1 = 31`
- 全局满分参考：`40`

## 阈值与判定规则

- 基础门槛：`raw_score_top >= 4`（避免单弱词误判）
- 高置信命中：`normalized_top >= 20` 且 `normalized_top - normalized_second >= 8`
- 候选冲突：`normalized_top >= 12` 且分差小于 8
- 未识别：不满足以上条件，走通用提取流程

## 输出建议格式

```yaml
domain_routing:
  top_domain: "supply_chain | crm_sales | project_delivery | generic"
  raw_score_top: 0
  raw_score_second: 0
  domain_max_score_top: 0
  domain_max_score_second: 0
  normalized_top: 0
  normalized_second: 0
  confidence: "high | medium | low"
  evidence:
    - keyword: ""
      weight: 0
      matched_text: ""
  next_action: "domain_extract | ask_user_confirm | generic_extract"
extract_output_contract:
  objects:
    explicit_objects: []
    inferred_objects: []
    pending_objects: []
    rejected_candidates: []
  relations: []
  constraints:
    - "对象名使用 canonical 命名，原文称呼写入 aliases"
    - "复合表达（如 A -> B）不得直接作为对象名"
    - "inferred_objects 必须包含 inference_reason"
    - "关系类 name 使用中文业务名；英文仅可用于 relation_id"
    - "pending_objects 非空时，必须先走待确认对象处理门禁"
```

## 回显文案模板（对用户）

### 模板：直接命中主领域（需确认）

```text
### 领域识别确认（领域识别 | 请确认）
说明：
- 主领域：{top_domain}
- 原始得分：{raw_score_top}/{domain_max_score_top}（次高：{raw_score_second}/{domain_max_score_second}）
- 归一化得分：{normalized_top}%（次高：{normalized_second}%）
- 置信度：{confidence}
- 命中证据：{keyword_1}({weight_1})，{keyword_2}({weight_2})...
下一步：请确认是否按该主领域继续提取（回复：确认按该领域提取 / 改为通用提取）。
```

### 模板：候选冲突（需用户确认）

```text
### 领域识别冲突确认（领域识别 | 请确认）
说明：
- 候选 1：{top_domain}（{raw_score_top}/{domain_max_score_top}，{normalized_top}%）
- 候选 2：{second_domain}（{raw_score_second}/{domain_max_score_second}，{normalized_second}%）
- 主要证据：{evidence_summary}
下一步：请确认主领域：1) {top_domain} 2) {second_domain} 3) 走通用提取。
```

### 模板：未识别（走通用流程）

```text
### 领域未命中处理（领域识别 | 风险提示）
说明：
- 最高得分：{raw_score_top}/{domain_max_score_top}（{normalized_top}%）
- 判定依据：未达到高置信阈值（`normalized >= 20` 且领先 `>= 8`）
- 主要证据：{evidence_summary}
下一步：进入 `./references/common/generic_extraction.md`，输出对象类与关系类最小可用清单并标记低置信度项。
```

## 兜底策略

- 未命中任何领域：进入 `./references/common/generic_extraction.md`
- 同时命中多个领域：先给出候选排序，再请求用户指定主领域
- 命中但证据不足：按命中领域 + 通用规则混合提取，并标注低置信度项
