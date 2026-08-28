---
name: smart-search-tables
version: "2.0.0"
user-invocable: true
description: >-
  找表/找数端到端编排：在元数据型知识网络下用 ontology bkn object-type query 检索表/视图实例，
  再在职责型知识网络下检索相关部门职责与治理边界，最后汇总为中文结论
  （候选表 + 职责要点 + 下一步）。当用户问「表在哪、哪个视图、数据资产归属、谁负责这类数据」时使用。
  所有 ontology CLI 执行均委托 ontology-core 完成；本 skill 不直接执行 CLI。
argument-hint: [找表/找数/资产定位类中文问题；可选 kn_id 覆盖]
---

# Smart Search Tables（找表 / 找数）

本 skill 定义 **固定先后顺序** 的找表工具链，完全由 `ontology` **native 子命令** 实现：先在元数据型 KN 中检索表/视图实例，再在职责型 KN 中检索相关部门职责。

在数据分析员工体系中，本 skill **必须由** [smart-data-analysis](../smart-data-analysis/SKILL.md) **总入口完成意图与 KN 编排后再进入执行**；CLI 实际执行由 [ontology-core](../ontology-core/SKILL.md) 承担。

## 调用方式（统一 ontology 命令；委托 ontology-core 执行）

本 skill 涉及的所有数据/元数据检索 **必须** 通过 `ontology` native 子命令发起。

**Never** 由本 skill 直接执行 `ontology` CLI；所有 CLI 执行均委托 [ontology-core](../ontology-core/SKILL.md) 完成。调用链固定：

```
smart-data-analysis（顶层意图 + KN 编排 + LLM 决策）
  └─ smart-search-tables（本 skill：描述要调的命令形态 + 顺序 + 总结口径）
       └─ ontology-core（实际执行 ontology 命令，返回结果）
```

### 子技能依赖

| 子技能 | 角色 | 返回 | 约束 |
|--------|------|------|------|
| [smart-data-analysis](../smart-data-analysis/SKILL.md) | 顶层意图路由与 KN 选定 | 进入本 skill 的上下文（`kn_id` / `duty_kn_id` / `accountId` / `search` 词） | Never 跳过 smart-data-analysis 直接接管流程 |
| [ontology-core](../ontology-core/SKILL.md) | smart-search-tables 的 CLI 委托 | 命令执行结果与回执 | Never 跳过 smart-search-tables 直接接管流程 |

### 委托给 ontology-core 的命令形态

本 skill 仅 **描述** 下列命令形态供 ontology-core 执行；本文档与 references 中 **不出现** 真实执行入口。

```
ontology --user-id <accountId> <command> [options]
```

具体 2 个检索步骤对应到 native 子命令：

| 步骤 | native 子命令 | 用途 |
|------|---------------|------|
| 1. 元数据 KN 检索 | `bkn object-type query <kn_id> <ot-id> '<condition-json>' [--limit n]` | 在元数据型 KN 下用语义/精确条件检索表/视图等实例 |
| 2. 职责 KN 检索 | `bkn object-type query <duty_kn_id> <ot-id> '<condition-json>' [--limit n]` | 在职责型 KN 下检索相关部门职责实例或概念 |
| —（辅助） | `bkn object-type list <kn-id>` | 在不确定 `ot-id` 时先列出对象类或语义定位 |

- **`--user-id <accountId>`**：**必传**（顶层选项，写在子命令之前；详见 ontology-core SKILL）。
- 网关（`--base-url` / `ONTOLOGY_BASE_URL`）由 ontology-core 侧承担，本 skill **不出现**该参数。
- 本部署 `ontology` CLI **无须 token**；命令体内 **不出现** `--token` / `auth.token` / `Authorization`。
- `-bd bd_public`：默认即 `bd_public`，可省。

## 必读 references（按步骤）

| 步骤 | 说明 | Reference |
|------|------|-----------|
| 1 | 元数据 KN 实例检索（找表/视图） | [references/metadata-search.md](references/metadata-search.md) |
| 2 | 职责 KN 检索（相关部门职责） | [references/duty-search.md](references/duty-search.md) |
| — | 端到端顺序示例 | [references/tool-examples.md](references/tool-examples.md) |

## 主流程（必须按序）

复制进度：

```text
找表进度：
- [ ] 1. 元数据 KN 检索：bkn object-type query <kn_id> <ot-id> '<condition-json>'，得到表/视图候选与部门/主题线索
- [ ] 2. 职责 KN 检索：基于第 1 步线索构造 query 或 condition，调 bkn object-type query <duty_kn_id> <ot-id>
- [ ] 3. 总结：候选表（business_name 全称 + technical_name）+ 职责要点 + 下一步建议；不暴露完整调试 URL
```

### 步骤约束（摘要）

1. **双 KN**：第 1 步用元数据型 KN（`kn_id`，由 smart-data-analysis 选定）；第 2 步用职责型 KN（`duty_kn_id`，同样由上游指定）。两者由 smart-data-analysis 用 `bkn list/get` + LLM 决策选定，**不在本 skill 内做 KN 选择**。
2. **第 1 步先于第 2 步**：第 2 步的检索词应能由第 1 步结果**派生**（部门名、主题域、表归属等）；若第 1 步无线索，则用用户原问题中的部门/组织词，或 **简要反问** 后再调职责查询。
3. **结果合并**：总结中区分 **事实发现**（检索到的表）与 **治理描述**（职责库中的条文）；二者无法强绑定时如实说明。
4. **业务名展示约束**：候选表输出必须包含表技术名 `technical_name` 与表业务名 `business_name`（若缺失则标注"暂无"）；展示时以 `business_name` 为主，且必须使用**完整全称**，禁止截断、省略或缩写。
5. **视图/表归并**：在元数据 KN 的实例字段中，`view_tech_name` 等价于 `table_tech_name`（统一为 `technical_name`），`view_business_name` 等价于 `table_business_name`（统一为 `business_name`）。

### 检索词与 condition 约束

- **`search`**：把用户「找表 / 找数」问题提炼为短语，包含业务对象 + 主题域 + 名称片段（避免代词、避免单字）。
- **`condition-json`**：参见 [metadata-search.md](references/metadata-search.md) 的样例；字段名、嵌套层级、布尔字面量保持同构，仅替换值。
- **`limit`**：默认拉 50–100；过宽时下调，命中过少时改写 `search` 或放宽 condition。

## 注意事项（必须遵守）

1. 所有信息**必须完全来自查询结果**，不允许添加任何结果中不存在的内容。
2. 不允许猜测、推断、脑补、编造表名/字段/部门。
3. 不允许改写、美化、夸张、虚构企业信息。
4. 不使用不确定词汇，如"可能""大概""应该""据悉"。
5. 若结果为空，直接说明"未查询到符合条件的数据"，并提示放宽 `search` / 换 KN / 二次澄清。
6. 严格按原始数据呈现，不修改数字、名称、顺序；表业务名禁止缩写。

## 与 smart-data-analysis 的关系

由 [smart-data-analysis](../smart-data-analysis/SKILL.md) 路由到本 skill 时，主意图为 **找表/定位**。其需向本 skill 提供：

- `accountId`（→ `--user-id`，必传）
- `kn_id`（元数据型 KN）与 `duty_kn_id`（职责型 KN，可选——若未提供则跳过第 2 步并在总结中注明"未检索职责"）
- `<ot-id>`（每个 KN 内用于实例检索的对象类 id；不明确时上游应先用 `bkn object-type list <kn-id>` 取候选并 LLM 选定）

网关（`ONTOLOGY_BASE_URL`）由 ontology-core 侧统一承担；本部署 ontology CLI **无须 token**，本 skill 与 smart-data-analysis 均不持有任何凭证。

用户后续要 **指标与 SQL 取数** → 转 [smart-ask-data](../smart-ask-data/SKILL.md)。

## 配置

- 本 skill **统一默认配置**：[config.json](config.json)
  - **`pipeline`**：2 步检索 + 1 步总结的顺序与对应 ontology 子命令的声明
  - **`runtime_contract`**：accountId / 网关 / 认证 / KN 入参的来源契约
  - 不维护 `base_url` / 端点 url_path / `defaults.user_id`：均由 ontology-core 侧环境变量或运行时入参承担

## 调用示例

```text
/smart-search-tables 采购订单相关宽表在哪个库、叫什么，谁在数据治理里负责？
/smart-search-tables 销售域 KPI 用哪张汇总表，对应部门职责怎么说
```
