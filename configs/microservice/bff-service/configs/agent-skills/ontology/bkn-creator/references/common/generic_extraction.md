# 通用对象类/关系类提取规则

## 目标

当领域未识别或领域知识不足时，提供稳定的通用提取流程。

## 步骤

1. 切分输入为“对象候选语句”和“关系候选语句”
2. 从名词短语提取对象类；过滤纯属性词（如状态、名称、备注）
3. 从动词或动宾短语提取关系类（如“提交申请”“关联订单”“依赖库存”）
4. 为每个对象类推断候选主键（优先编码/单号/ID）
5. 为每个关系类补齐方向、基数和证据

## 对象类判定规则

- SHOULD：可独立存在、可被唯一标识、生命周期超过一次动作
- SHOULD NOT：临时动作、单次事件结果、纯展示字段
- MUST：对象名使用规范化名称（推荐小写下划线）；原文称呼放入 `aliases`
- MUST：名称含 `->`、`:`、`/`、`，`、`与` 的复合表达不得直接作为对象类，转入 `rejected_candidates`

## 关系类判定规则

- MUST：关系名称可读且可解释
- MUST：关系 `name` 使用中文业务名（如“下达给”“关联”“驱动”）
- SHOULD：若需要英文技术标识，使用独立字段 `relation_id`（如 `assign_to_supplier`）
- MUST：有明确 `source_object` 与 `target_object`
- SHOULD：优先一跳关系，避免一次表达多跳链路

## 关系方向判定规则

### 判定优先级（默认策略）

当同一关系存在“溯源视角”和“驱动视角”两种可解释方向时，按以下优先级判定：

1. 溯源视角优先：若目标对象存在可识别的来源字段（如 `src*` / `source*` / `parent*` / `root*`），优先采用“当前对象 -> 来源对象”
2. 驱动视角次之：若语义明确描述“上游驱动下游生成”，可采用“驱动对象 -> 被驱动对象”

### 判定步骤（MUST）

1. 检查关联字段所在对象与字段命名语义
2. 根据优先级给出推荐方向和理由
3. 若两种方向均合理且证据冲突，必须标记为“方向歧义”
4. 出现“方向歧义”时，必须发起用户确认；在确认前不得固化方向

### 歧义回显模板（对用户）

```text
### 关系方向确认（通用提取 | 请确认）
说明：
- 关系：{relation_name}
- 候选方向A：{direction_a}（{reason_a}）
- 候选方向B：{direction_b}（{reason_b}）
- 推荐方向：{recommended_direction}（{recommended_reason}）
下一步：请确认关系方向（回复：采用推荐 / 采用方向A / 采用方向B / 自定义方向）。
```

## 输出模板

```yaml
objects:
  explicit_objects:
    - name: ""
      business_meaning: ""
      candidate_primary_key: "unknown"
      aliases: []
      evidence: ""
  inferred_objects:
    - name: ""
      business_meaning: ""
      candidate_primary_key: "unknown"
      aliases: []
      evidence: ""
      inference_reason: ""
  pending_objects: []
  rejected_candidates:
    - candidate: ""
      reason: ""
relations:
  - name: ""
    relation_id: ""
    source_object: ""
    target_object: ""
    cardinality: "unknown"
    join_fields: []
    business_meaning: ""
    evidence: ""
```

## 质量检查

- 对象名去重（同义词合并为主名 + aliases）
- 关系名去重（语义一致合并）
- 每条关系引用的对象必须在对象清单中存在
- 提取结果必须分组为 `explicit_objects` / `inferred_objects` / `pending_objects`
- `inferred_objects` 必须包含推断依据 `inference_reason`
- 无法确认主键/基数时标记 `unknown`，不得臆造
