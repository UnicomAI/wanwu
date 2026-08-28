---
name: smart-data-analysis
version: "2.0.0"
user-invocable: true
description: >-
  数据分析员工（Data Analyst Agent）的唯一总入口：凡与数据资产、取数、指标、表/视图、
  治理职责、知识网络、统计或分析相关的问题，必须先经本 skill 做编排与路由，再进入找表或问数等子流程。
  负责 kn 分域、上下文注入（accountId / date）、多候选 KN 时的 LLM 决策、
  问数分支的 SQL 生成；与 smart-search-tables / smart-ask-data / ontology-core 的交接。
  当用户提出任何数据类自然语言任务、或需在多条业务 KN 间切换时使用；
  所有 ontology CLI 执行均委托 ontology-core 完成，本 skill 不直接执行 CLI。
argument-hint: [自然语言指令或带 kn 上下文的任务描述]
---

# Smart Data Analysis（总编排）

本 skill 是 **数据分析员工角色的总入口**：在本体数据技能栈中，**所有数据相关问题必须先经过本 skill**，完成 **KN 与上下文对齐、意图路由** 后，再委派至 **找表** 或 **问数**（或其它数据子 skill），**禁止**在未做编排判断时直接跳到 `smart-search-tables`、`smart-ask-data` 或零散工具调用。

**Never** 由本 skill 直接执行 `ontology` CLI；所有 CLI 调用（包括为选 KN 取候选元数据的 `bkn list/get`，与问数分支的 `bkn object-type get / dataview query --sql` 等）均委托 [ontology-core](../ontology-core/SKILL.md) 执行。

调用链：

```
smart-data-analysis（顶层意图 + LLM 决策：选 KN / 生成 SQL）
  ├─ ontology-core（直接委托：bkn list/get 等用于自身决策）
  └─ smart-search-tables 或 smart-ask-data（子流程）
       └─ ontology-core（子流程的 CLI 委托）
```


## 总入口原则（必须遵守）

1. **先编排，后执行**：识别用户问题是否属于「数据域」（资产/表/视图/指标/SQL/图表/职责/元数据/多 KN 等）；若是，**先走本节下方「编排总流程」与「路由识别」**，再打开对应子 skill 或工具链。
2. **单一前门**：同一轮对话中新增的数据子任务，仍应 **回到本 skill 的编排逻辑** 决定是延续当前分支还是切换找表/问数。
3. **按最终意图路由**：先判断用户最终想要的是“定位数据资产（表/视图）”还是“拿到数据结果（指标/明细/统计/图表）”。前者走找表；后者走问数，必要时在问数分支内先找表再生成 SQL。
4. **交接清晰**：转入 [smart-search-tables](../smart-search-tables/SKILL.md) 或 [smart-ask-data](../smart-ask-data/SKILL.md) 时，在内部上下文中保留已解析的 **`kn_id_*`、时间口径**，避免子 skill 重复猜 KN。
5. **非数据问题**：与数据无关时 **不必** 强行套用本 skill；若用户一句话里混有数据与非数据，**数据部分**仍按上述原则经本 skill 编排（可分段回答）。
6. **禁止交叉兜底**：本轮路由为 **问数** 时，若问数走不通，**禁止**改走 **找表** 分支代替交付；路由为 **找表** 时，若找表走不通，**禁止**改走 **问数** 分支代替交付。两种情况下均应 **直接输出走不通的原因**（缺 KN、无命中、平台错误等）及用户侧可采取的修复条件。细则见下方「分支走不通时的处理（禁止交叉兜底）」。

## 知识网络声明（KN id 单一来源）

KN id 在下表中直接声明，**由本 skill 路由时透传到下游**（smart-ask-data / smart-search-tables）。下游 skill **不** 自行选 KN、**不** 调 `bkn list/get` 枚举或决策。

| 分支 | 用途 | KN ID |
|------|------|-------|
| 找表 | 表/视图实例检索（必填） | `<填入KN id>` |
| 找表 | 部门职责检索（可选） | `<填入KN id>` |
| 问数 | SQL 取数（必填） | `<填入KN id>` |

约束：

- 维护者一次性通过 ontology-core 调 `bkn list` 取得平台可用 KN id 后，**直接编辑本表**填入；运行时不再做 KN 列举或选择。
- 若占位仍为 `<填入...>`：进入对应分支时直接告知用户"未在 smart-data-analysis/SKILL.md 中配置对应 KN"，**不得**用其它 KN 凑数。
- 用户明确说「用 XX 知识网络问数 / 找表」→ 仅当该 id 等于上表已声明值时采用；否则提示用户先把该 KN 添加到本表再用。

## 与其它技能的分工

| 能力 | 本 skill（编排） | 专用 skill（执行细节） |
|------|------------------|------------------------|
| 路由、kn 切换、上下文注入 | ✅ 主责 | 配合 |
| KN id 声明 + 透传到下游 | ✅ 主责（见上方「知识网络声明」表） | 接收注入，不自行选 KN |
| 找表 / 定位 / 职责 / 澄清 | 定义路由与交接 | [smart-search-tables/SKILL.md](../smart-search-tables/SKILL.md) |
| 问数：SQL **生成**（基于 schema 摘要） | ✅ 主责（LLM） | 输出 SELECT/WITH SQL 给 smart-ask-data 步骤 4 |
| 问数：步骤顺序、口径约束、SQL **执行** | 定义路由与交接 | [smart-ask-data/SKILL.md](../smart-ask-data/SKILL.md) |
| ontology native CLI 执行（`bkn / ds / dataview / vega / call`） | **不直接执行** | [ontology-core/SKILL.md](../ontology-core/SKILL.md) |

## 子技能依赖

| 子技能 | 角色 | 返回 | 约束 |
|--------|------|------|------|
| [ontology-core](../ontology-core/SKILL.md) | smart-data-analysis 的 CLI 委托 | 命令执行结果与回执 | Never 跳过 smart-data-analysis 直接接管流程 |
| [smart-search-tables](../smart-search-tables/SKILL.md) | 找表分支执行 | 候选表/视图清单、归属、职责要点 | 走找表分支时由 smart-data-analysis 路由进入；本 skill 需向其透传 `accountId / kn_id / duty_kn_id / <ot-id> / search` 等入参；不直接执行 ontology CLI |
| [smart-ask-data](../smart-ask-data/SKILL.md) | 问数分支执行 | SQL 执行结果 + 口径 | 走问数分支时由 smart-data-analysis 路由进入；本 skill 需向其透传 `accountId / kn_id / 生成的 SQL` 三件套 |

## 职能矩阵（谁做什么）

| 角色/技能 | 负责（Do） | 不负责（Don't） | 典型输出 |
|-----------|------------|-----------------|----------|
| `smart-data-analysis`（总编排） | 识别主意图；**从「知识网络声明」表读取 KN id 并透传给下游**；组织 `date` / `accountId` 等上下文；问数分支基于 schema 摘要生成 SELECT SQL | 不直接执行 ontology CLI；不在运行时用 `bkn list` 决策 KN；不直接做对象检索细节；不替代子 skill 结果 | 路由决策、KN 透传、生成 SQL、交接约束 |
| `smart-search-tables`（找表执行） | 接收 smart-data-analysis 透传的 `kn_id` / `duty_kn_id` / `<ot-id>` / `accountId`；按 2 步顺序触发 `bkn object-type query`（必要时 `bkn object-type list`）等 native 命令的委托；产出候选表/视图清单 + 职责要点 | 不承担跨流程总路由；不在 skill 内自行选 KN；不直接执行 ontology CLI | 候选表清单、归属、职责说明 |
| `smart-ask-data`（问数执行） | 接收 smart-data-analysis 透传的 `kn_id` 与 SQL；按 5 步顺序触发 `bkn object-type get` / `dataview query --sql` 等 native 命令的委托；产出取数结果与口径展示 | 不承担跨流程总路由；不在 skill 内自行选 KN 或生成 SQL；不直接执行 ontology CLI | 数值结果、口径说明、SQL 与依据 |
| `ontology-core`（CLI 委托执行） | 实际承载所有 `ontology` 命令的执行（`bkn / ds / dataview / vega / call`）；统一处理 `--user-id / --base-url / -bd` 等横向约束 | 不替代业务编排判断；不直接定义问数口径 | 命令执行结果与回执 |

若仅有顶层编排而无子 skill 正文：**仍须按下方路由规则执行**；找表/问数细节分别以 [smart-search-tables](../smart-search-tables/SKILL.md)、[smart-ask-data](../smart-ask-data/SKILL.md) 为准；CLI 与 BKN 以 `ontology-core` 的 `references/*.md` 为准（尤其 `bkn.md`）。

## 编排总流程（必须按序）

复制并勾选：

```text
编排进度：
- [ ] 1. 解析任务中的 KN / 业务域（见「知识网络分域」）
- [ ] 2. 注入公共上下文（见「上下文注入」）
- [ ] 3. 意图路由：找表 / 问数（见「路由识别」）
- [ ] 4. 进入对应分支清单并完成；歧义时先澄清
- [ ] 5. 输出结构：结论 + 依据（表/视图/KN）+ 下一步可选动作
```

## KN 切换规则

实际 KN id 见上方「[知识网络声明](#知识网络声明kn-id-单一来源)」表；切换时按以下规则透传给下游：

- 路由到 **找表分支** → 注入"找表"行的 KN id 为 `kn_id`；若"职责"行非空则同时注入 `duty_kn_id`。
- 路由到 **问数分支** → 注入"问数"行的 KN id 为 `kn_id`。
- 用户明确说「用 XX 知识网络问数 / 找表」→ 仅当 XX 等于声明表中某行的值时采用；否则要求用户先把它加进声明表。
- 同时跨找表/问数（先找表再算数）→ 两步分别注入各自分支的 KN id，不混用。

## 上下文注入

在调用检索或问数工具前，尽量在推理上下文中 **显式整理**（不必向用户冗长展示）：

| 键 | 含义 | 来源 / 用途 |
|----|------|--------------|
| `accountId` | 当前会话用户的账户 id | 由会话框架/调用方提供；委托 ontology-core 时注入 `--user-id <accountId>`（必传，缺失时向用户索要，不得编造） |
| `date` | 用户问题中的时间范围、默认「当前日期」与对比周期（同比/环比） | 解析自用户问题 |
| 路由分支 + 注入下游的 `kn_id`（找表分支额外含 `duty_kn_id`） | 当前要进入的分支决定注入哪行声明值 | 从「知识网络声明」表读取；占位为 `<填入...>` 时告知用户并停止 |

将上述一并作为后续工具调用的隐含约束，减少跨 KN 误查。网关（`ONTOLOGY_BASE_URL`）由 [ontology-core](../ontology-core/SKILL.md) 侧承担；本部署 ontology CLI **无须 token**，本 skill 与子 skill 均不持有任何凭证。


## 路由识别

按优先级匹配用户**最终意图**（一句可多标签，取最终要交付给用户的结果）：

### 走「找表（找数）」分支（最终目标是定位资产）

触发词或场景示例：表在哪、哪个视图、字段在哪个模型、主题域/部门职责、资产目录、「有没有叫…的表」、**仅**定位不做指标计算。

找表分支通过 ontology-core 委托 `bkn object-type query`（必要时 ``bkn object-type list`）：第 1 步在元数据型 KN 下做实例检索拿候选表/视图，第 2 步在职责型 KN 下做实例检索拿相关部门职责。详见 [smart-search-tables/SKILL.md](../smart-search-tables/SKILL.md)。

**分支清单**

```text
找表进度：
- [ ] 确认 `accountId`（缺失则向用户索取，不得编造）
- [ ] 从「知识网络声明」表读取找表 KN id 作为 `kn_id`；占位仍为 `<填入...>` → 告知用户并停止
- [ ] 同步读取职责 KN id 作为 `duty_kn_id`（可空；空则后续跳过职责检索）
- [ ] 选定 `<ot-id>`：每个 KN 内用于实例检索的对象类 id；不确定时先 `bkn object-type list <kn-id>` + LLM 选定
- [ ] 把用户问题提炼为 `search` 短语（业务对象 + 主题域 + 名称片段；避免单字 / 代词）
- [ ] 委托 smart-search-tables：透传 `accountId / kn_id / duty_kn_id / <ot-id> / search`，按 2 步触发 `bkn object-type query`
- [ ] 意图不清：结构化反问（业务主题？系统？表名片段？）
- [ ] 输出：候选表/视图（`business_name` 全称 + `technical_name`）+ 相关部门职责 + 若下一步要统计则引导进入问数
```

**与 smart-search-tables 的交接契约**（**本 skill 的硬责任**）：

| 字段 | 含义 | 缺失处理 |
|------|------|----------|
| `accountId` | 当前会话用户账户 id | 向用户索取；不得编造 |
| `kn_id` | 找表 KN id | 从「知识网络声明」表"找表"行读取；占位未填 → 告知用户并停止 |
| `duty_kn_id` | 职责 KN id（可选） | 从「知识网络声明」表"职责"行读取；留空则跳过职责检索并在总结中说明"未检索职责" |
| `<ot-id>` | KN 内用于实例检索的对象类 id（一般两个 KN 各一个） | 不确定时先 `bkn object-type list <kn-id>` + LLM 选定 |
| `search` | 提炼后的检索短语 | 必须由用户问题提炼，不得为代词或单字 |

### 走「问数」分支（最终目标是拿到数据结果）

触发词或场景：

- **聚合/统计类**：多少、占比、趋势、TopN、指标、统计、分析结论、查询…信息（需要结果）
- **关系/关联类**：与…关联 / 与…有关、由…组成 / 包含哪些、经过哪些 / 流向哪里、哪些 X 与 Y…、X 的下游 / 上游是什么
- **逻辑属性/派生指标类**：用户提到的指标名不在 `data_properties` 里，而是出现在 `bkn object-type get` 返回的 `logic_properties` 中（`type: tool` 的计算属性，由服务端 box 引擎结算）
- 已识别具体视图/表并需要 **聚合或 SQL 级取数**

关系类问题的最终交付是"相关实体集合"，需要跨对象类 JOIN，**必须走 SQL 问数**，不能用单 ot 的 `bkn object-type query` 拼凑代替（参见下方判别表）。

逻辑属性类问题的最终交付是"经服务端 box 引擎结算的计算值"，**必须走 `bkn object-type properties`** 结算路径，不能用 `bkn object-type query` 代替——后者只返回逻辑属性的入参绑定（`body` 里填好字段值），**不执行计算**（参见下方判别表）。

问数分支通过 ontology-core 委托 ``bkn object-type list/get` 做 Schema 发现，本 skill 在编排层基于 schema 摘要 **生成 SELECT SQL**，再由 [smart-ask-data](../smart-ask-data/SKILL.md) 委托 `dataview query --sql` 执行。绘图与二次代码加工能力不内置；如需绘图，明确告知用户由前端自渲染或独立处理。

**关键区分：SQL 问数 vs 对象类实例检索 vs 逻辑属性取值**

三者都能返回数据，但能力差很大。**默认偏好**：涉及"关系"或"聚合" → SQL 问数；单实体精确/范围过滤 → object-type query；逻辑属性（计算字段）→ object-type properties。

| 任务形态 | SQL 问数（dataview query --sql） | object-type query | object-type properties |
|----------|:----:|:----:|:----:|
| 跨对象类 JOIN | ✅ | ❌（只能单 ot，需手工拼接） | ❌ |
| 聚合统计（COUNT/SUM/AVG/GROUP BY） | ✅ | ❌ | ❌ |
| 单实体 ID 精确查找 | 任选 | ✅（更轻） | ❌ |
| 单实体多条件过滤 | 任选 | ✅ | ❌ |
| 全文检索 | ❌（SQL 视图不支持 match） | ✅ | ❌ |
| **逻辑属性取值（`logic_properties` 中的计算字段）** | ❌ | ❌（只给入参，不结算） | ✅（服务端 box 引擎结算，返回结果值） |

凡问题中出现"X 与哪些 Y 关联 / 由哪些 Y 组成"这类需要**跨对象类关系**的措辞，或出现"多少 / 占比 / 趋势 / TopN / 平均"这类**聚合**措辞，**必须**进入 SQL 问数；不要用单 ot 的 `bkn object-type query` 拼接代替。

凡问题中出现的指标名命中 `bkn object-type get` 返回的 `logic_properties` 数组（即该指标是 `type: tool` 的计算属性），**必须**走 `bkn object-type properties` 结算路径；不要用 `bkn object-type query` 代替（那只返回入参绑定，不执行计算）。

**`bkn object-type query` 的 JSON body 必须用 `condition` 外层包裹**（编排层常错点）：

```json
// 单条件
{"limit": 30, "condition": {"field": "<字段名>", "operation": "==", "value": "<值>"}}

// 多条件 AND
{"limit": 30, "condition": {
  "operation": "and",
  "sub_conditions": [
    {"field": "<f1>", "operation": "==", "value": "<v1>"},
    {"field": "<f2>", "operation": "in", "value": ["<a>","<b>"]}
  ]
}}
```

错例（缺 `condition` 外层，报 InvalidParameter "Filter conditions must be wrapped in a 'condition' structure"）：`{"field":"x","operation":"==","value":"y"}`

**分支清单**

```text
问数进度：
- [ ] 确认 `accountId`（缺失则向用户索取，不得编造）
- [ ] 从「知识网络声明」表读取问数 KN id 作为 `kn_id`；占位仍为 `<填入...>` → 告知用户并停止
- [ ] 候选 > 1：本 skill 委托 ontology-core 调 `bkn list` / `bkn get` 取候选元数据，LLM 决策选定
- [ ] 若表未就绪：先在问数流程内短循环找表/找视图（调用 smart-search-tables 或追问）后再继续
- [ ] 明确指标口径、时间粒度、维度与过滤条件
- [ ] 委托 smart-ask-data：触发 schema 发现（`bkn object-type get`）拿候选对象类、字段、dataview-id（取自 `data_source.id`，要求 `data_source.type == "data_view"`）
- [ ] **路由判定**（二选一，不可混用）：
  - [ ] **逻辑属性路径**：schema 发现返回的 `logic_properties` 非空且含用户要的计算属性 → 走 `bkn object-type properties` 结算（见下方交接契约），**不走 SQL**
  - [ ] **SQL 路径**：聚合/JOIN/普通字段取值 → 本 skill 在编排层基于 schema 摘要 + 用户口径生成 SELECT/WITH SQL（必带 LIMIT；字段表名必须来自摘要）
- [ ] 输出：执行的命令或 SQL（脱敏可，不可省）+ 关键数据 + 口径说明 + 可复核步骤
```

**与 smart-ask-data v2 的交接契约**（**本 skill 的硬责任**）：

问数分支按路由判定分为两条路径，交接内容不同：

**SQL 路径**（聚合/JOIN/普通字段取值）：

| 字段 | 含义 | 缺失处理 |
|------|------|----------|
| `accountId` | 当前会话用户账户 id | 向用户索取；不得编造或用 config 默认值 |
| `kn_id` | 问数 KN id | 从「知识网络声明」表"问数"行读取；占位未填 → 告知用户并停止 |
| 生成的 SELECT SQL | 基于 step-2 schema 摘要的 SQL；只允许 SELECT/WITH；字段表名必须来自摘要；带 LIMIT | smart-ask-data 不内置 LLM，本 skill 必须给出 |

**逻辑属性路径**（计算字段/派生指标取值）：

| 字段 | 含义 | 缺失处理 |
|------|------|----------|
| `accountId` | 当前会话用户账户 id | 向用户索取；不得编造或用 config 默认值 |
| `kn_id` | 问数 KN id | 从「知识网络声明」表"问数"行读取；占位未填 → 告知用户并停止 |
| `ot_id` | 目标对象类 id | 来自 schema 发现的 `bkn object-type get` 返回 |
| `logic_property_names` | 要结算的逻辑属性 `name` 列表（取自 schema 发现返回的 `logic_properties[].name`） | 从 `bkn object-type get` 的 `logic_properties` 数组中提取；为空 → 说明该对象类无逻辑属性，回退 SQL 路径或告知用户 |
| `instance_identities` | 目标实例的主键标识数组 | 由用户提供或从 `bkn object-type query` 查得；缺失则向用户索取 |

> 逻辑属性路径不需要本 skill 生成 SQL——`bkn object-type properties` 是固定 body 结构的取值命令，由 smart-ask-data 委托 ontology-core 直接执行。

### 歧义与复合请求（按最终意图收敛）

- **找表 + 问数** 同句出现：最终目标若是“出结果”，则归入**问数分支**，仅把找表当作前置步骤。
- **仅**「分析一下」且无对象：先澄清对象与时间范围，默认不要直接生成长文编造数字。
- **跨 KN**：找表与问数分别使用「知识网络声明」表中各自分支的 KN id；禁止混用未在声明表中出现的 KN。
- 示例：`查询企业相关信息` 若用户期望返回企业数据内容/统计结果，归入**问数**（必要时先找企业相关表/视图，再生成 SQL 查询）。

### 分支走不通时的处理（禁止交叉兜底）

- **问数走不通**：例如未指定且 `bkn list` 无可用业务 KN、`bkn object-type get` 的 `data_source` 缺失或 `data_source.type` 不是 `data_view`、`dataview query --sql` 报错或返回空等 → **只向用户说明具体原因与补齐条件**，**不得**为给出「类似答案」而切换到 **找表** 分支；找表结果不能替代用户要的指标或明细取数。
- **找表走不通**：例如「知识网络声明」表中找表行未填、`bkn object-type query` 命中为空、职责 KN 未提供或检索无果等 → **只向用户说明具体原因**，**不得**切换到 **问数** 分支用 SQL 猜测表名、编造资产位置或虚构明细。
- **与「问数内前置找表」的区分**：在 **问数分支** 内，按 [smart-ask-data](../smart-ask-data/SKILL.md) 在生成 SQL 前做的 Schema/表发现（`bkn object-type list/get`，必要时短循环找表）属于 **同一问数任务的固定子步骤**，**不属于**「问数失败后的交叉兜底」。
- **绘图 / 二次代码加工请求**：本架构不内置 `json2plot` / `execute_code_sync` 能力。若用户要求绘图或代码二次加工，直接向用户说明当前架构不支持，可用 SQL 表达的口径继续问数。

## 输出格式建议

对用户回复推荐结构：

1. **结论**（一两句）
2. **依据**（用了哪个 KN、哪些表/视图或取数路径）
3. **数据或列表**（表格或要点）
4. **可选下一步**（例如：是否在同一视图上做同比）

## 调用示例（slash / 指令）

```text
/smart-data-analysis 在当前找表用的 kn 里查有没有订单相关宽表
/smart-data-analysis 用问数 kn 算上月销售额环比，口径按销售域约定
/smart-data-analysis 找表 kn 用 A，问数 kn 用 B：先找「库存」视图再算周转
```

## 注意事项

- **不要**在未确认 `kn_id` 的情况下假设当前知识网络已正确。
- **不要**默认对纯 SQL 视图源使用 `match` 全文操作符；文本模糊优先 `like`。
- 专用 skill 中的 `references/tool-examples.md` 若存在，**在执行层优先遵循**；本文件只负责 **路由与编排约束**。
- **Never** 由本 skill 直接执行 `ontology` CLI；所有 ontology 命令（含选 KN 时的 `bkn list/get`）一律委托 [ontology-core](../ontology-core/SKILL.md) 执行。
- **`--user-id <accountId>`** 是 ontology 命令的硬性顶层选项；本 skill 在交接给下游前 **必须** 已拿到 `accountId`，缺失则向用户索取。
- **问数分支 SQL 生成是本 skill 的硬责任**：smart-ask-data v2 不内置 LLM；不要把"生成 SQL"推给下游或委托给 ontology-core。
- 本部署 ontology CLI **无须 token**；命令体内不出现 `--token` / `auth.token` / `Authorization`。命令报错时直接如实反馈，不要伪造结果或尝试登录刷新凭证。
