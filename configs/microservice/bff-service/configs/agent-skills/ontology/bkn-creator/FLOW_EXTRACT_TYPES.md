# 提取流程（Extract）

## 流程目标

从业务描述、文档或需求中提取对象类与关系类候选清单，并给出可用于后续建模的结构化结果。

## 验证阶段：流程路由确认（进入提取流程前）

- MUST：先回显“识别结果 + 路由目标 + 输入摘要”，并发起确认请求
- MUST：仅当用户明确确认“进入提取流程”后，才可进入后续阶段
- Never：在验证阶段未确认前输出 `discover/preview/confirm` 细化内容或执行任何提取动作

## 渐进式阶段

1. `discover`：识别输入来源、业务范围与候选领域
2. `preview`：回显提取策略（领域知识提取 / 通用提取）
3. `confirm`：用户确认提取范围与粒度
4. `execute`：执行提取并形成候选清单
5. `verify`：核验主键、关系方向、基数、命名归一化与分组一致性
6. `confirm_pending`：处理待确认对象（`pending_objects`）
7. `report`：输出“显式对象 + 推断对象 + 关系清单 + 待确认项”

> 统一门禁规则见 `./COMMON_RULES.md`。对用户可见输出必须使用“用户回显模板（统一格式）”。

进入条件（前置）：验证阶段路由确认已通过。

## 主控流程

### Step 1：领域识别（必须先做）

1. 读取 `./references/DOMAIN_ROUTING.md`
2. 按评分制计算各领域得分（`high/mid/low signal` 权重）
3. 输出每个候选领域的 `raw_score / domain_max_score / normalized_score`
4. 按阈值判定主领域或冲突状态
5. 回显识别结果（候选领域、原始分、满分、归一化分、命中证据、下一步动作）
6. MUST 按 `./references/DOMAIN_ROUTING.md` 的“回显文案模板”输出

### Step 2：分支确认与执行

- 命中领域：先请求用户确认“按该主领域继续提取”，确认后读取对应 `./references/<domain>/domain_*.md`
- 候选冲突：先请求用户确认主领域，再进入领域提取
- 未命中领域：先请求用户确认是否走通用提取，确认后读取 `./references/common/generic_extraction.md`

### Step 3：结果校验（verify，MUST）

完成提取后，进入统一校验：

1. 对象命名归一化：对象名必须为 canonical 名称，原文叫法保留到 `aliases`
2. 复合表达拦截：包含 `->`、`:`、`/`、`，`、`与` 的候选名称不得直接入对象类，转入 `rejected_candidates`
3. 对象分组校验：必须区分 `explicit_objects`、`inferred_objects`、`pending_objects`
4. 推断对象可解释：`inferred_objects` 每项必须包含 `inference_reason`
5. 引用完整性：关系引用对象必须存在于对象总清单（显式 + 推断 + 待确认）

### Step 4：待确认对象处理（confirm_pending，MUST）

当 `pending_objects` 非空时，必须先完成待确认对象处理，才可进入最终清单确认：

1. 回显 `pending_objects` 详情（名称、证据、不确定原因）
2. 请求用户逐项或批量选择处理动作：
   - `纳入对象清单`
   - `移出清单`
   - `保留待确认（并标注风险）`
3. 根据用户确认更新对象分组后，再输出“提取结果确认”

门禁约束（MUST）：

- `pending_objects` 非空且未完成处理确认前，Never 直接询问“清单是否完整”
- 未完成处理时，不得进入 `report` 阶段

## 领域分支规则

- MUST：同一轮提取只采用一个主领域，跨领域时先请用户确认主领域
- MUST：优先复用领域文档内的标准术语与关系方向
- SHOULD：保留原文别名（用于后续字段映射）

## 通用分支规则

- 以名词短语识别对象类（稳定业务实体）
- 以动词短语识别关系类（交互/依赖/流转）
- 关系必须明确：`source_object -> target_object`
- 每条关系应提供候选基数（`1:1` / `1:N` / `N:1` / `N:N`）

## 输出规范（MUST）

### 对象类清单

- `explicit_objects`（显式对象）
- `inferred_objects`（推断对象，含 `inference_reason`）
- `pending_objects`（待确认对象）
- `rejected_candidates`（被拒绝候选及原因）
- `inference_summary`（推断依据摘要，面向确认回显）
- `rejected_summary`（拒绝候选摘要：数量 + 前 3 条示例）

### 关系类清单

- `name`（MUST 中文业务名）
- `relation_id`（可选，英文技术名/slug）
- `source_object`
- `target_object`
- `cardinality`
- `join_fields`（可空）
- `business_meaning`
- `evidence`（原文依据）

### 展示策略（MUST）

- 默认采用“摘要回显”作为主展示格式，不重复输出同一数据的表格与 YAML/JSON
- 仅当用户明确请求时才输出详情（如“展开对象详情”“导出 YAML”“给我表格”）
- 详情输出一次仅允许一种格式（表格 或 YAML/JSON），避免同轮重复回显

## 失败与兜底

- 领域识别冲突：列出前 2 个候选领域并请求用户确认主领域
- 证据不足：输出“最小可用清单”并标记 `待补充`
- Never 伪造字段或主键；缺失时明确写 `unknown`

## 回显模板（对用户）

### 模板：流程路由确认（验证阶段）

```text
### 流程路由确认（验证阶段 | 请确认）
说明：
- 当前识别：这是"提取对象类/关系类"请求
- 将进入流程：提取流程（FLOW_EXTRACT_TYPES）
- 你提供的信息摘要：{input_summary}
- 说明：这一步只确认流程方向，还不会开始提取
下一步：你可以回复"确认进入提取流程"，或先调整提取范围后再继续。
```

### 模板：提取结果确认

```text
### 提取结果确认（提取流程 | 请确认）
说明：
- 显式对象数量：{explicit_object_count}
- 推断对象数量：{inferred_object_count}
- 推断依据摘要：{inference_summary}
- 待确认对象数量：{pending_object_count}
- 关系类数量：{relation_count}
- 已拒绝候选：{rejected_count}（示例：{rejected_top3_examples}）
- 待补充项：{pending_items}
下一步：请确认清单是否可进入下一阶段（回复：确认清单 / 调整清单）。如需详情可回复：展开对象详情 / 导出 YAML。
```

### 模板：待确认对象处理

```text
### 待确认对象处理（提取流程 | 请确认）
说明：
- 待确认对象：{pending_objects_detail}
- 不确定原因：{pending_reasons}
- 可选动作：
  1) 纳入对象清单
  2) 移出清单
  3) 保留待确认并继续（带风险）
下一步：请回复处理决策（可逐项回复：`对象A=纳入，对象B=移出`）。
```
