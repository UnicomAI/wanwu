---
name: smart-ask-data
version: "2.0.0"
user-invocable: true
description: >-
  问数端到端编排（native CLI 版）：从候选 KN 选定知识网络，用 bkn object-type 发现对象类与字段，
  由编排层 LLM 生成 SQL，再由 ontology dataview query 执行取数；
  最后输出中文结论与口径说明。
  当用户需要指标、统计、趋势、SQL 取数或数据查询时使用。
argument-hint: [中文问数问题；可选已有 kn_id 或候选 kn 列表]
---

# Smart Ask Data（问数）

本 skill 定义 **固定先后顺序** 的问数工具链，完全由 `ontology` **native 子命令**（`bkn` / `dataview` / `ds`）实现。

在数据分析员工体系中，本 skill **必须由** [smart-data-analysis](../smart-data-analysis/SKILL.md) **总入口完成意图、选 KN、生成 SQL 等 LLM 决策后再进入执行**；CLI 实际执行由 [ontology-core](../ontology-core/SKILL.md) 承担。

## 调用方式（统一 ontology 命令；委托 ontology-core 执行）

本 skill 涉及的所有数据/Schema 访问 **必须** 通过 `ontology` native 子命令发起。

**Never** 由本 skill 直接执行 `ontology` CLI；所有 CLI 执行均委托 [ontology-core](../ontology-core/SKILL.md) 完成。调用链固定：

```
smart-data-analysis（顶层意图 + LLM 决策：选 KN、生成 SQL）
  └─ smart-ask-data（本 skill：描述要调的命令形态 + 顺序 + 约束）
       └─ ontology-core（实际执行 ontology 命令，返回结果）
```

### 子技能依赖

| 子技能 | 角色 | 返回 | 约束 |
|--------|------|------|------|
| [smart-data-analysis](../smart-data-analysis/SKILL.md) | 顶层意图路由、KN 选择、SQL 生成 | 进入本 skill 的上下文（`kn_id` / 生成的 SQL / accountId 等） | Never 跳过 smart-data-analysis 直接接管流程 |
| [ontology-core](../ontology-core/SKILL.md) | smart-ask-data 的 CLI 委托 | 命令执行结果与回执 | Never 跳过 smart-ask-data 直接接管流程 |

### 委托给 ontology-core 的命令形态

本 skill 仅 **描述** 下列命令形态供 ontology-core 执行；本文档与 references 中 **不出现** 真实执行入口。

```
ontology --user-id <accountId> <command> [options]
```

具体 4 个步骤对应到 native 子命令（详见各 reference）：

| 步骤 | native 子命令 | 用途 |
|------|---------------|------|
| 1. 找 KN |  `bkn get <kn-id>` |  取详情供 LLM 选择 |
| 2. 取字段与 dataview-id | `bkn object-type list <kn-id>` / `bkn object-type get <kn-id> <ot-id>` | 拿字段 + 后端 `dataview-id` |
| 3. 执行 SQL | `dataview query <dataview-id> --sql "..."` | LLM 生成的 SELECT/WITH SQL（mdl-uniquery） |
| —（简单单表） | `bkn object-type query <kn-id> <ot-id> '<filter-json>'` | 不需要 SQL 时的实例过滤 + 分页 |

- **`--user-id <accountId>`**：**必传**（顶层选项，写在子命令之前；发送为 `x-account-id`；详见 ontology-core SKILL）。
- 网关（`--base-url` / `ONTOLOGY_BASE_URL`）由 ontology-core 侧承担，本 skill **不出现**该参数。
- 本部署 `ontology` CLI **无须 token**；本 skill 命令体内 **不出现** `--token` / `auth.token` / `Authorization`。
- `-bd bd_public`：默认即 `bd_public`，可省。

## 必读 references（按步骤）

| 步骤 | 说明 | Reference |
|------|------|-----------|
| 1 | 知识网络选择（条件执行） | [references/kn-resolve.md](references/kn-resolve.md) |
| 2 | Schema 发现：候选对象类与字段 | [references/schema-discovery.md](references/schema-discovery.md) |
| 3 | SQL 生成（编排层 LLM）+ 执行 | [references/sql-execute.md](references/sql-execute.md) |
| — | 端到端顺序示例 | [references/tool-examples.md](references/tool-examples.md) |

## 主流程（必须按序）

复制进度：

```text
问数进度：
- [ ] 1. 解析 kn_id：若已指定或仅 1 个候选 KN 则直用；多候选时由编排层 LLM 用 bkn list/get 选定（见 kn-resolve）
- [ ] 2. Schema 发现：bkn object-type list/get → 候选对象类、字段、dataview-id、**logic_properties**（逻辑属性定义）
- [ ] 3. 路由判定（二选一，不可混用）：
  - [ ] 3a. **逻辑属性路径**：用户要的指标名命中 step-2 返回的 `logic_properties[].name` → 走 `bkn object-type properties` 结算（见 references/sql-execute.md "逻辑属性取值"）
  - [ ] 3b. **SQL 路径**：聚合/JOIN/普通字段取值 → 编排层 LLM 基于第 2 步信息生成 SELECT SQL
- [ ] 4. 执行取数：
  - [ ] 4a. 逻辑属性路径：ontology bkn object-type properties <kn-id> <ot-id> '{"_instance_identities":[{...}],"properties":[...]}' → 服务端 box 引擎结算，返回计算值
  - [ ] 4b. SQL 路径：ontology dataview query <dataview-id> --sql "..." 取数
- [ ] 5. 总结：结论 + 口径 + 依据（KN/对象类/执行的命令或 SQL）
```

### 知识网络来源

- 问数使用的 `kn_id` 来自 (a) 调用方/上游 smart-data-analysis 明确指定；(b) 通过 ontology-core 调 `bkn list` 枚举。

### 步骤约束（摘要）

1. **KN 解析（条件路由）**：
   - 已传入 `kn_id`：直接使用。
   - 仅 1 个候选 KN：直接使用。
   - 候选 > 1：由 **smart-data-analysis** 的 LLM 决策（基于 `bkn list` / `bkn get` 拿到的候选元数据）。
2. **Schema 发现先于取数**：先用 `bkn object-type list/get` 锁定对象类与字段，再决定走 SQL 还是逻辑属性路径；防止 SQL 幻觉。
3. **逻辑属性路径 vs SQL 路径**：
   - 用户问题中的指标名出现在 `bkn object-type get` 返回的 `logic_properties` 数组中 → 走逻辑属性路径。
   - `logic_properties` 为空或不含目标属性 → 走 SQL 路径。
   - **禁止混用**：逻辑属性路径不生成 SQL；SQL 路径不调 `bkn object-type properties`。
4. **SQL 生成在编排层**：本 skill 不内置 LLM；SQL 由 smart-data-analysis 生成后传入。
5. **只允许 SELECT/WITH**：`dataview query --sql` 默认拒绝写操作；不得使用 `--raw-sql` 绕过。
6. **结果展示硬约束**：若执行返回非空结果，最终回复中 **必须同时展示**：
   - 生成并执行的 SQL 或逻辑属性取值命令（可脱敏，不可省略）；
   - 关键结果数据（表格或要点汇总，不可仅给口头结论）。
7. **总结**：明确时间范围、指标定义；不暴露完整调试 URL。

## 注意事项（必须遵守）

1. 所有信息**必须完全来自查询结果**，不允许添加任何结果中不存在的内容。
2. 不允许猜测、推断、脑补、编造数据。
3. 不允许改写、美化、夸张、虚构企业信息。
4. 不使用不确定词汇，如"可能""大概""应该""据悉"。
5. 若结果为空，直接说明"未查询到符合条件的数据"，不得自行编造。
6. 只做结构化整理、排序、计数、分段展示，不做逻辑外扩。
7. 严格按原始数据呈现，不修改数字、名称、顺序。

## 与 smart-data-analysis 的关系

由 [smart-data-analysis](../smart-data-analysis/SKILL.md) 做顶层路由时，进入本 skill 表示用户 **主意图为问数**。其需向本 skill 提供：

- `accountId`（→ `--user-id`，必传）
- `kn_id`（已选定的业务 KN；缺失时附带候选 `kn_ids` 并已剔除 forbidden 项，由本 skill 触发 LLM 二次决策）
- 路由判定结果（二选一）：
  - **SQL 路径**：由 LLM 基于 Schema 发现结果生成的 SELECT SQL（步骤 3b 起需要）
  - **逻辑属性路径**：`ot_id` + `logic_property_names`（从 `bkn object-type get` 返回的 `logic_properties[].name` 提取）+ `instance_identities`（目标实例主键标识）

网关（`ONTOLOGY_BASE_URL`）由 ontology-core 侧统一承担；本部署 ontology CLI **无须 token**，本 skill 与 smart-data-analysis 均不持有任何凭证。

## 配置

- 本 skill **统一默认配置**：[config.json](config.json)
  - **`pipeline`**：5 步顺序与对应 ontology 子命令的声明
  - **`runtime_contract`**：accountId / 网关 / 认证 / kn_id 等入参的来源契约
  - 不维护 `base_url` / 端点 url_path / `defaults.user_id`：均由 ontology-core 侧环境变量或运行时入参承担

## 调用示例

```text
/smart-ask-data 上个月各区域销售额，按区域汇总
/smart-ask-data 在候选知识网络里自动选 KN，查库存周转相关明细并给结论
```
