# 项目交付领域提取参考

## 典型对象类候选（canonical）

- 项目（`project`）
- 里程碑（`milestone`）
- 任务（`task`）
- 交付物（`deliverable`）
- 风险（`risk`）
- 问题（`issue`）
- 资源（`resource`）
- 工时记录（`timesheet`）

## 典型关系类候选

- 项目 `包含` 里程碑（`project -> milestone`，`1:N`）
- 里程碑 `包含` 任务（`milestone -> task`，`1:N`）
- 任务 `产出` 交付物（`task -> deliverable`，`1:N`）
- 任务 `依赖` 任务（`task -> task`，`N:N`）
- 项目 `登记` 风险（`project -> risk`，`1:N`）
- 项目 `跟踪` 问题（`project -> issue`，`1:N`）
- 资源 `执行` 任务（`resource -> task`，`N:N`）
- 资源 `提交` 工时记录（`resource -> timesheet`，`1:N`）

## 对象提取三层（MUST）

- `explicit_objects`：文本显式对象
- `inferred_objects`：由流程链路补全对象
- `pending_objects`：证据不足待确认对象

推断对象必须带 `inference_reason`，并与显式对象分开回显。

## 对象命名归一化（MUST）

- 使用 canonical 对象名（如 `milestone`、`deliverable`）
- 原文叫法写入 `aliases`
- 含 `->`、`:`、`/`、`，`、`与` 的复合表达不得直接作为对象名
- “任务延期风险”“交付验收问题”等短语优先拆为对象 + 关系，而非合并对象名

## 关系识别提示

- 出现“延期、阻塞、依赖、排期冲突”时优先检查任务依赖与资源分配关系
- 出现“验收、交付、阶段完成”时优先检查里程碑与交付物关系
- 关系类输出时，`name` 必须使用中文业务名（如“包含”“产出”“跟踪”）；英文仅可放在 `relation_id`

## 领域闭环完整性检查（MUST）

识别到以下片段但对象缺失时，必须补全为 `inferred_objects`：

- 命中 `milestone` 且缺少 `task` -> 补全 `task`
- 命中 `task` 且出现“交付/验收”语义但缺少 `deliverable` -> 补全 `deliverable`
- 命中“阻塞/延期/风险”语义且缺少 `risk` 或 `issue` -> 补全缺失对象
- 命中“资源投入/工时”语义且缺少 `resource` 或 `timesheet` -> 补全缺失对象

补全结果在确认阶段必须分组回显并请求用户确认。

## 主键候选提示

- 编码类：`project_code`、`milestone_id`、`task_id`、`deliverable_id`
- 单据类：`risk_no`、`issue_no`
- 人员/资源类：`resource_id`、`timesheet_id`

## 输出建议结构

```yaml
objects:
  explicit_objects: []
  inferred_objects: []
  pending_objects: []
  rejected_candidates: []
relations: []
```
