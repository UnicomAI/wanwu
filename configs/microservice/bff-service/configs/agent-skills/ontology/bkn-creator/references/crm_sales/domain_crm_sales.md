# CRM/销售领域提取参考

## 典型对象类候选（canonical）

- 客户（`customer`）
- 线索（`lead`）
- 商机（`opportunity`）
- 报价单（`quotation`）
- 合同（`contract`）
- 销售订单（`sales_order`）
- 回款（`payment`）
- 销售人员（`sales_rep`）

## 典型关系类候选

- 客户 `产生` 线索（`customer -> lead`，`1:N`）
- 线索 `转化为` 商机（`lead -> opportunity`，`1:N` 或 `1:1`）
- 商机 `生成` 报价单（`opportunity -> quotation`，`1:N`）
- 报价单 `签订为` 合同（`quotation -> contract`，`1:N`）
- 合同 `履约为` 销售订单（`contract -> sales_order`，`1:N`）
- 销售订单 `对应` 回款（`sales_order -> payment`，`1:N`）
- 销售人员 `负责` 客户（`sales_rep -> customer`，`1:N`）

## 对象提取三层（MUST）

- `explicit_objects`：输入文本直接出现的对象
- `inferred_objects`：根据领域链路补全的对象
- `pending_objects`：证据不足、需确认对象

每个推断对象必须提供 `inference_reason`。

## 对象命名归一化（MUST）

- 对象输出名使用 canonical 名称（如 `sales_order`、`customer`）
- 原文保留到 `aliases`
- 含 `->`、`:`、`/`、`，`、`与` 的名称视为复合表达，不得直接入对象名
- “签约转单”“合同到回款链路”等表达优先拆解为关系，不作为对象名

## 关系识别提示

- 出现“跟进、转化、赢单、签约、续约、回款”时优先检查线索到合同链路
- 出现“客户分配、客户经理、账户归属”时优先检查销售人员与客户关系
- 关系类输出时，`name` 必须使用中文业务名（如“转化为”“履约为”）；英文仅可放在 `relation_id`

## 领域闭环完整性检查（MUST）

若识别到下列片段但对象缺失，必须补全为 `inferred_objects`：

- 命中 `quotation` 且缺少 `contract` -> 补全 `contract`
- 命中 `contract` 且缺少 `sales_order` -> 补全 `sales_order`
- 命中 `sales_order` 且缺少 `payment` -> 补全 `payment`
- 命中 “客户经理/归属/分配”语义且缺少 `sales_rep` 或 `customer` -> 补全缺失对象

补全结果必须在确认阶段分组回显，允许用户删除不采纳项。

## 主键候选提示

- 编码类：`customer_code`、`lead_code`、`opportunity_code`
- 单据类：`quotation_no`、`contract_no`、`order_no`、`payment_no`
- 人员类：`sales_rep_id`、`employee_id`

## 输出建议结构

```yaml
objects:
  explicit_objects: []
  inferred_objects: []
  pending_objects: []
  rejected_candidates: []
relations: []
```
