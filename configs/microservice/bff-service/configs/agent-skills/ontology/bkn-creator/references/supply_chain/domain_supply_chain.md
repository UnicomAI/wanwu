# 供应链领域提取参考

## 典型对象类候选（canonical）

- 产品（`product`）
- 产品 BOM（`bom`）
- 物料（`material`）
- 需求预测（`forecast`）
- 产品需求计划（`pp`）
- 物料需求计划（`mrp`）
- 工厂生产计划（`mps`）
- 采购申请（`pr`）
- 采购订单（`po`）
- 供应商（`supplier`）
- 库存（`inventory`）
- 销售订单（`sales_order`）

## 典型关系类候选

- 产品 `包含` BOM（`product -> bom`，`1:N`）
- BOM `引用` 物料（`bom -> material`，`N:1`）
- 物料 `形成` 采购申请（`material -> pr`，`1:N`）
- 采购申请 `转换为` 采购订单（`pr -> po`，`1:N`）
- 采购订单 `下达给` 供应商（`po -> supplier`，`N:1`）
- 物料 `记录于` 库存（`material -> inventory`，`1:N`）
- 需求预测 `驱动` 产品需求计划（`forecast -> pp`，`N:1`）
- 产品需求计划 `展开为` 物料需求计划（`pp -> mrp`，`1:N`）
- 产品需求计划 `驱动` 工厂生产计划（`pp -> mps`，`1:N`）
- 销售订单 `关联` 产品（`sales_order -> product`，`N:1`）

## 对象提取三层（MUST）

- `explicit_objects`：输入文本直接出现的对象
- `inferred_objects`：根据领域闭环规则补全的对象
- `pending_objects`：证据不足、需要用户确认的候选对象

Never 将推断对象伪装成显式对象。每个推断对象必须附 `inference_reason`。

## 对象命名归一化（MUST）

- 输出对象名必须使用 canonical 名称（如 `sales_order`、`material`）
- 原文别名进入 `aliases`，不得直接作为最终对象名
- 对象名若包含 `->`、`:`、`/`、`，`、`与`，判定为“关系表达或复合短语”，转入 `rejected_candidates`
- 类似“销售订单 -> 物料”“需求物料需求”等表达，优先拆解为关系候选而非对象类

## 关系识别提示

- 出现“转单、下达、审批、到货、入库、领料、齐套”时优先检查 PR/PO/Inventory 相关关系
- 出现“预测、计划、排程、缺口、净需求”时优先检查 Forecast/PP/MRP/MPS 相关关系
- 关系类输出时，`name` 必须使用中文业务名（如“下达给”“转换为”）；英文仅可放在 `relation_id`

## 领域闭环完整性检查（MUST）

若识别到下列链路片段但对象缺失，必须生成 `inferred_objects` 补全建议：

- 命中 `pr` 或 `po`，但缺少 `supplier` -> 补全 `supplier`
- 命中 `mrp`，但缺少 `material` -> 补全 `material`
- 命中 `pp`，但缺少 `mrp` 或 `mps` -> 补全缺失计划对象
- 命中 “到货/入库/库存”语义，但缺少 `inventory` -> 补全 `inventory`

补全后必须在确认阶段回显“显式对象 vs 推断对象”分组，等待用户确认。

## 主键候选提示

- 编码类：`material_code`、`supplier_code`、`product_code`
- 单据类：`entry_id`、`billno`、`contract_number`
- 流水类：`seq_no`

## 输出建议结构

```yaml
objects:
  explicit_objects: []
  inferred_objects:
    - name: supplier
      aliases: [供应商]
      inference_reason: "命中 pr/po 链路，供应商为必需参与方"
  pending_objects: []
  rejected_candidates: []
relations: []
```
