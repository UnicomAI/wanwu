# 查找流程（Read）

## 流程目标

定位并展示知识网络或其对象/关系结构，不执行写操作。

## 验证阶段：流程路由确认（进入查找流程前）

- MUST：先回显“识别结果 + 路由目标 + 输入摘要”，并发起确认请求
- MUST：仅当用户明确确认“进入查找流程”后，才可进入后续阶段
- Never：在验证阶段未确认前输出 `discover/preview/confirm` 细化内容或执行查询命令

## 渐进式阶段

1. `discover`：识别查找目标（名称、ID、范围）
2. `preview`：回显查询计划（按名模糊查 / 按 ID 精确查）
3. `confirm`：用户确认查询范围
4. `execute`：委托 `ontology-core` 执行查询
5. `verify`：核对结果是否与目标一致
6. `report`：按“网络级/对象级/关系级”输出结构化回执

> 统一门禁规则见 `./COMMON_RULES.md`。对用户可见输出必须使用“用户回显模板（统一格式）”。

进入条件（前置）：验证阶段路由确认已通过。

## 展示策略（MUST）

- 默认仅输出摘要回显（命中数量、关键字段、风险与下一步）
- 同一轮仅允许一种主展示格式，禁止同一数据同时用表格与 YAML/JSON 重复展示
- 仅当用户明确请求时才输出详情（如“展开详情”“导出 YAML”“给我表格”）
- 详情输出一次仅允许一种格式（表格 或 YAML/JSON）

## 标准查询命令（由 ontology-core 执行）

```bash
ontology bkn list --name-pattern "关键词"
ontology bkn get <kn_id>
ontology bkn object-type list <kn_id>
ontology bkn relation-type list <kn_id>
```

## 输出规范

- 网络级：`kn_id`、名称、comment、更新时间
- 对象级：对象类名、主键、关键字段、数据源状态
- 关系级：关系名、源对象、目标对象、映射字段

## 风险控制

- MUST 使用完整 `kn_id`，Never 截断
- 当结果为空时，必须提供“可能原因 + 下一步检索建议”

## 回显模板（对用户）

### 模板：流程路由确认（验证阶段）

```text
### 流程路由确认（验证阶段 | 请确认）
说明：
- 当前识别：这是"查找知识网络"请求
- 将进入流程：查找流程（FLOW_READ）
- 你提供的信息摘要：{input_summary}
- 说明：这一步只确认流程方向，还不会开始查询
下一步：你可以回复"确认进入查找流程"，或先调整查询目标后再继续。
```

### 模板：查询计划确认

```text
### 查询计划确认（查找流程 | 请确认）
说明：
- 查询目标：{target_name_or_id}
- 查询范围：{network_level | object_level | relation_level}
- 查询方式：{name_pattern | exact_id}
下一步：请确认是否按以上范围执行查询（回复：确认查询 / 调整范围）。
```

### 模板：查询结果回执

```text
### 查询结果（查找流程 | 结果）
说明：
- 网络级：{network_summary}
- 对象级：{object_summary}
- 关系级：{relation_summary}
下一步：如需继续，可指定对象类/关系类进行深入查看。若需详情可回复：展开查询详情 / 导出 YAML。
```

### 模板：结果为空

```text
### 查询结果为空（查找流程 | 风险提示）
说明：
- 可能原因：{possible_reason_1}；{possible_reason_2}
- 已执行检索：{executed_query_scope}
下一步：建议改用名称关键词、完整 `kn_id` 或扩大检索范围后重试。
```
