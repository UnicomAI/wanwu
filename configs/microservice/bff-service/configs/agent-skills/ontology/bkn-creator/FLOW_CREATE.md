# 新增流程（Create）

## 流程目标

从业务输入生成可推送的 BKN，并完成绑定、推送与验证。

## 验证阶段：流程路由确认（进入新增流程前）

- MUST：先回显“识别结果 + 路由目标 + 输入摘要”，并发起确认请求
- MUST：仅当用户明确确认“进入新增流程”后，才可进入阶段一
- Never：在验证阶段未确认前输出“阶段一/Step 1”或任何提取动作

## 阶段序列（5 阶段）

1. 阶段一：建模意图澄清
2. 阶段二：生成 BKN 草案
3. 阶段三：环境检查（执行前）
4. 阶段四：数据视图绑定与属性回灌
5. 阶段五：推送与验证

## 阶段门禁

| 阶段 | 进入条件 | 退出条件（MUST） |
|------|----------|-------------------|
| 阶段一 | 用户确认进入新增流程 | 视角确认（A1/B1）+ 清单确认（A2/B2 或 C1）+ 业务规则检查完成（或用户确认带风险继续） |
| 阶段二 | 阶段一完成 | `ontology bkn validate` 通过 + 用户确认草案 |
| 阶段三 | 阶段二完成 | 连通性完成 + 同名/相似网络检查完成 |
| 阶段四 | 阶段三完成 | 完成“绑定决议 -> Step 4.4 属性二次生成 -> Step 4.6 完备性放行 -> 回填确认”全链路 + 用户确认阶段四结果 |
| 阶段五 | 阶段四完成 | 推送成功 + 完整性检查 + 孤悬检查反馈 |

> 详细确认语义与全局约束见 `./COMMON_RULES.md`。对用户可见输出必须使用“用户回显模板（统一格式）”。

## 展示策略（MUST）

- 默认仅输出摘要回显（数量、状态、风险、下一步），避免重复展开同一数据
- 同一轮仅允许一种主展示格式，禁止同一数据同时用表格与 YAML/JSON 重复展示
- 仅当用户明确请求时才输出详情（如“展开详情”“导出 YAML”“给我表格”）
- 详情输出一次仅允许一种格式（表格 或 YAML/JSON）
- 涉及内部编码枚举的对用户展示，统一遵循 `./COMMON_RULES.md` 的“字段显示词典（用户可见）”

## 用户文案去术语化（MUST）

- 对用户回显时，Never 直接展示内部流程编号（如 `A1/B1/A2/B2/C1`）
- 对用户回显时，Never 使用“门禁”一词，统一改为“确认步骤”或“确认请求”
- 推荐替换映射：
  - `A1/B1` -> `建模视角确认`
  - `A2/B2/C1` -> `建模清单确认`
  - `阶段四门禁确认` -> `阶段四结果确认`
- 内部文档可保留编号用于执行控制，但用户可见文案必须采用业务可读表达

## 阶段一：建模意图澄清

进入条件（前置）：验证阶段路由确认已通过。

### Step 1：输入类型判定

- A 结构化文档：可直接提取实体
- B 部分信息：可识别出 3+ 候选对象
- C 委托建模顾问：输入或中间结果不稳定，需要建模顾问收敛

### 路径 A（结构化文档）

1. 默认采用“对象-属性-视图映射”主视角（Object-Property-View）
2. 门禁确认 A1：用户确认视角后再提取
3. 提取对象类/关系类清单（先经 `./references/DOMAIN_ROUTING.md` 统一调度）
4. 门禁确认 A2：用户确认清单后进入阶段二

### 路径 B（部分信息）

1. 先给出主视角（对象-属性-视图映射），再按需补充 1-2 个候选视角对比
2. 门禁确认 B1：用户确认所选视角
3. 提取对象类/关系类清单
4. 门禁确认 B2：用户确认清单后进入阶段二

### A1/B1 门禁硬约束（MUST）

- 在 A1 或 B1 未确认前，Never 执行任何对象类/关系类提取动作
- 在 A1 或 B1 未确认前，Never 执行领域识别（DOMAIN_ROUTING）或展示领域识别结果
- A1/B1 视角确认必须单独占一轮对话，该轮 **仅** 输出视角确认请求，不得同时输出领域识别、提取结果或其他门禁确认
- 若误执行了提取或领域识别，必须立即中止并回滚到 A1/B1 确认步骤

### 视角选择策略（MUST）

- 默认主视角：`对象-属性-视图映射`，用于对象属性定义、字段映射与可用性判定
- `实体-关系（ER）` 视角降级为“条件触发的补充视角”，不作为默认主视角
- ER 触发条件（满足任一即可启用）：
  - 预计关系类数量 `> 5`
  - 出现 `N:M` 关系或“关系本身有属性”
  - 存在跨对象关键约束（审批链、层级/BOM、跨流程一致性）
- 启用 ER 时的输出边界（MUST）：
  - 仅用于关系骨架梳理（关系方向、基数、约束）
  - 不主导对象属性生成
  - 不替代阶段四的属性映射完备性检查与可用性判定
- 进入 A1/B1 视角确认时，MUST 在回显中明确：
  - `selected_view`（最终主视角）
  - `er_triggered`（`是 | 否`）
  - `er_trigger_reasons`（当 `er_triggered=是` 时必填）

### 路径 C（委托建模顾问）

触发条件（满足任一即可，MUST）：

1. 输入模糊：对象/关系不清晰
2. 领域冲突：候选领域分差小且用户未确认主领域
3. 待确认项过多：`pending_objects >= 3`
4. 清单质量不足：关系方向冲突、主键缺失或命名异常导致无法稳定进入阶段二

1. 读取并委托：`../bkn-modeling-advisor/SKILL.md`
2. 接收建模清单（对象/关系/操作）
3. 门禁确认 C1：展示并确认后进入阶段二

### 领域识别与统一调度（A1/B1 确认后触发）

仅在用户回复明确确认 A1/B1 之后的下一轮才可执行领域识别。Never 在 A1/B1 确认请求的同一轮输出中执行领域路由。

若 A1/B1 已确认，MUST 先读取 `./references/DOMAIN_ROUTING.md` 进行领域识别与路由，再执行提取：

1. 命中领域：读取 `./references/<domain>/domain_*.md` 执行领域提取  
2. 候选冲突：先请求用户确认主领域，再进入领域提取  
3. 未命中领域：读取 `./references/common/generic_extraction.md` 执行通用提取

### 业务规则提取检查清单（阶段一末尾，MUST）

阶段一结束前，必须完成业务规则检查并形成结构化结果（默认以摘要回显给用户）：

| 检查项 | 提取位置 | 完成标准 |
|--------|----------|----------|
| 主键/外键规则 | 对象描述 / Data Properties | 已标注关联字段与主键候选 |
| 数据过滤规则 | 对象描述 / 关系证据 | 已记录过滤条件 |
| 状态枚举规则 | Data Properties Description | 已列出关键枚举值 |
| 计算/判定规则 | Logic Properties / 对象描述 | 已记录公式或判断条件 |
| 层级/BOM规则 | 对象描述 | 已说明层级与替代/版本约束 |

内部结构格式（示例，默认不对用户直出）：

```yaml
rule_extraction_check:
  - object: bom
    rules_extracted: [替代料规则(alt_priority), 版本规则(bom_version)]
    status: complete
  - object: mrp_plan_order
    rules_extracted: [过滤规则(closestatus_title)]
    status: complete
  - object: inventory
    rules_extracted: []
    status: warning
```

门禁要求（MUST）：

- 存在 `warning` 时，必须先输出风险提示并请求用户确认是否带风险继续
- 未获得“带风险继续”明确确认前，不得进入阶段二
- 默认仅回显检查摘要（完整/告警数量 + 关键告警项）；如用户明确要求，再展示完整结构详情

## 阶段二：生成 BKN 草案

- MUST 读取并委托：`../create-bkn/SKILL.md`
- MUST：目录与落盘位置统一遵循并直接委托 skill `archive-protocol`，不在本流程内定义 `archives/*` 规则
- 输出目录（`network_dir`）由 `archive-protocol` 提供与校验
- 质量检查：
  - `network.bkn` 的 `id` MUST 在生成 BKN 草案的同时本地生成一个 UUID v4 直接写入（形如 `f47ac10b-58cc-4372-a567-0e02b2c3d479`），不得留空、不得写 `null` / `~`、不得整键缺失
    - 原因：YAML 空值会被解析为 `None`，`ontology bkn validate` 与 `ontology bkn push` 都要求 `id` 为非空字符串；阶段五不再调用 `bkn create` 申请 id，直接用本地 UUID 作为最终 `kn_id`
    - 该 UUID 一经写入即为最终 `kn_id`，阶段五 push 时复用同一值；Never 在后续阶段重新生成或替换该 id（除非用户明确"重建网络"，否则视为同一网络）
    - 若 `object_types/*.bkn` 等需要引用 network id，MUST 同步写入该 UUID
    - MUST：UUID 标准为 v4 随机版本；Never 使用顺序 id、业务名拼接、占位符字符串
  - `network.bkn` MUST 补齐默认样式：`icon: icon-dip-graph`、`color: #0e5fc5`
  - 所有 `object_types/*.bkn` MUST 设置 `color`，按“随机颜色”策略分配（可复现即可，不要求真正随机源）
  - 对象类 Data Source 标记为“待绑定”，映射字段留空
  - `comment` / Description 完整
  - Description 文案纯净：仅描述稳定业务语义，不得写入映射过程信息（如“视图衍生”“待核实”“视图中使用 xxx”）
  - 映射相关不确定信息统一写入独立字段（如 `mapping_note` / `mapping_confidence` / `verify_status`），不得混入 Description
- 样式补齐门禁（MUST）：
  - 未完成网络默认样式或对象类颜色分配前，不得进入阶段三
  - 若对象类已有颜色且用户未要求覆盖，保持原值不改
- 草案人工复核门禁（MUST）：
  - 阶段二回显必须明确引导用户查看并复核草案文件：`network.bkn`、`object_types/*.bkn`、`relation_types/*.bkn`
  - 用户如需调整，必须允许其直接修改 `.bkn` 文件后再继续流程
  - 未收到“确认草案”或“已完成草案修改”前，不得进入阶段三
- 验证：委托 `ontology-core` 执行 `ontology bkn validate <network_dir>`

## 阶段三：环境检查（执行前）

- 读取并委托：`../ontology-core/SKILL.md`
- 执行：`ontology bkn list --limit 200`
- 输出：连通性状态、可用 `kn_id` 列表（完整 ID）
- 说明：若环境检查失败，保留已生成草案并暂停阶段四/五执行
- 回显模板：

```text
### 环境检查结果（阶段三 | 结果）
说明：
- 连通性：{connected | failed}
- 可用网络数：{available_kn_count}
- 同名/相似网络：{similar_network_summary | 无}
下一步：环境就绪，可回复"继续"进入阶段四；如有异常需先处理。
```

## 阶段四：数据视图绑定与属性回灌

### 设计目标（重排后）

- 将阶段四从“单线串行”改为“绑定决议线 -> 属性回灌线 -> 放行确认线”
- `Step 4.4` 改为阶段四的硬门槛，不再作为可被统计或确认步骤绕过的中间项
- 所有阶段四对外确认都统一后置到 `Step 4.9`，禁止提前确认

### 执行状态机（MUST，禁止跳步）

按以下状态顺序推进，前一状态未达成不得进入下一状态：

`S4_BIND_EVIDENCE -> S4_BIND_DECISION -> S4_PROPERTY_REGEN -> S4_MAPPING_GATE -> S4_DIFF_CONFIRM -> S4_STAGE_CONFIRM`

状态与步骤映射：

- `S4_BIND_EVIDENCE`：Step 4.1 ~ Step 4.2
- `S4_BIND_DECISION`：Step 4.3
- `S4_PROPERTY_REGEN`：Step 4.4（关键重排点）
- `S4_MAPPING_GATE`：Step 4.5 ~ Step 4.7
- `S4_DIFF_CONFIRM`：Step 4.8
- `S4_STAGE_CONFIRM`：Step 4.9

若尝试从 `Step 4.3`、`Step 4.5`、`Step 4.7` 直接输出“阶段四结果确认”，必须立即回退到缺失状态补齐后再继续。

### 执行步骤（严格顺序）

1. **Step 4.1：数据源线索提取**
   - 从输入文档与上下文提取数据视图线索，形成 `data_source_clues`
   - 产出 `data_source_map = { object_id: clue_value | null }`
   - 线索值（名称/技术名/别名）仅作为候选，不得直接写入最终绑定字段
2. **Step 4.2：现有视图匹配与存在性校验（MUST）**
   - 读取并委托 `../ontology-core/SKILL.md`，使用 `ontology dataview` 命令族完成匹配
   - 对已有绑定值（如 `mdl_id` / 技术名）先解析为视图 `id`，再校验存在性，状态仅允许：`exists | not_found | ambiguous`
   - 仅 `exists` 计入“已绑定”；其余一律转入“待决策”
   - 若原值为名称/技术名且已解析出视图 `id`，记录 `backfill_replace_candidate`
   - 未命中的对象统一进入“待决策”，由 Step 4.9 转手动提供或跳过
3. **Step 4.3：绑定决议快照（MUST）**
   - 汇总 `binding_decision_list`：`bound` / `pending` / `rejected`
   - 生成 `binding_summary_pre_regen`（阶段四中间统计，仅内部使用）
   - Never 在本步骤发起“阶段四结果确认”；本步骤只做“绑定决议”不做放行
4. **Step 4.4：对象类属性二次生成（视图回灌，MUST）**
   - 以“用户原始业务文本 + Step 4.3 已确认绑定决议 + 视图字段”重建对象属性候选
   - 二次生成结果必须分层：`core_business` / `view_derived` / `technical_excluded`
   - 必须输出差异动作：`keep` / `add` / `rename` / `merge` / `drop_candidate`
   - 对用户展示时使用中文标签：`核心业务/视图衍生/技术排除` 与 `保留/新增/重命名/合并/建议移除`
   - Description 与映射备注必须分离：稳定业务语义写 Description，不确定映射写映射备注字段
   - 当“无已绑定视图”时也不得跳过本步骤：必须输出 `step_4_4_summary`（可为空差异）并标注 `reason=no_bound_view`
5. **Step 4.5：属性映射草案生成（MUST）**
   - 基于 Step 4.4 的“对象属性定稿候选”生成 `property_mapping_draft`
   - 每个属性都必须进入三态之一：`mapped` / `waived` / `blocked`
   - `mapped` 必填：`view_id`、`field_path`、`confidence`、类型校验结果
   - `waived` 必填：`waive_reason`
   - `blocked` 必填：`block_reason`
6. **Step 4.6：属性映射完备性放行检查（MUST）**
   - 计算 `coverage = (mapped + waived) / total_properties`
   - 放行条件必须同时满足：`blocked_count = 0` 且 `coverage = 100%`
   - 未满足放行条件时必须暂停，并输出互斥策略包（3选1）：
     - 策略 A（BKN优先）：保留核心业务属性，仅补充必要映射后重跑 Step 4.5 -> Step 4.6
     - 策略 B（视图优先）：按视图重构属性，采纳 `add/rename/merge` 后重跑 Step 4.4 -> Step 4.6
     - 策略 C（混合治理）：仅对高价值差异增补，其余豁免后重跑 Step 4.5 -> Step 4.6
7. **Step 4.7：阶段四统计与完整性证据（MUST）**
   - 输出最终 `binding_summary`（`total_objects` / `bound_objects` / `unbound_objects` / `binding_rate`）
   - 输出并校验以下证据齐备：
     - `step_4_4_summary`（至少含 `keep/add/rename/merge/drop_candidate`）
     - `step_4_6_summary`（至少含“已映射/已豁免/待处理”与 `coverage`）
     - `binding_decision_list`（至少含 `bound/pending/rejected` 分组）
   - 任一证据缺失时禁止进入 Step 4.8
8. **Step 4.8：统一差异清单与回填确认（MUST）**
   - 将绑定变更、属性变更、映射状态合并成一份统一差异清单后再请求确认
   - 对已解析出视图 `id` 但仍是名称/技术名绑定值的对象，必须纳入“替换回填”清单
   - 仅当用户明确确认“执行回填”后，才可写入 BKN 文件；Never 自动回填
9. **Step 4.9：阶段四结果确认（唯一确认出口）**
   - 当 `binding_rate >= 80%`：发起阶段四确认，用户确认后进入阶段五
   - 当 `binding_rate < 80%`：暂停并要求用户选择“手动提供 / 跳过未绑定对象并继续 / 中止流程”
   - “阶段四结果确认”只能在本步骤输出，其他步骤输出一律视为流程违规

阶段四退出条件（MUST）：

- 已产出并保留：`binding_decision_list`、`step_4_4_summary`、`step_4_6_summary`、`binding_summary`
- 放行条件成立：`blocked_count = 0` 且 `coverage = 100%`
- 统一差异清单已回显，且回填已获用户确认
- 阶段四结果仅在 Step 4.9 确认一次

语义约束（MUST）：

- `Step 4.4` 是硬门槛步骤，不得因“绑定率已达标”或“用户催促进入推送”而跳过
- “跳过未绑定对象并继续”仅影响未绑定对象，不影响 `Step 4.4 ~ Step 4.7` 的必执行性
- 绑定值最终必须是视图 `id`；名称、技术名、别名不得作为最终绑定值
- `waived` 是有依据豁免，不计阻断；`blocked` 必须清零后方可进入阶段五

## 阶段五：推送与验证

id 口径（MUST）：

- `kn_id` 由阶段二生成 `network.bkn` 时本地 UUID v4 写入，阶段五**直接复用**，Never 再调用 `ontology bkn create` 申请 id
- 推送前 MUST 回读 `network.bkn` 的 `id` 字段，确认其为合法 UUID v4 且非占位符/None；不合法时停在阶段五并修复 `network.bkn`，Never 跳过校验直接 push
- 若用户明确要求"重建网络"，重新生成 UUID 覆盖原 `id` 后再 push；否则视为同一网络

执行前门禁（MUST）：

- 推送前必须先委托 `ontology-core` 检查同名/相似名称网络（基于 `ontology bkn list` 结果）
- 若存在同名或相似网络，默认使用"基名 + 最新版本号"创建新网络（如 `_v5`）；Never 尝试更新既有网络
- 版本号规则：提取同基名历史版本中的最大版本号并 `+1`；若无版本后缀但已存在同名，则从 `_v2` 起
- 在执行 `ontology bkn push` 前，必须先展示命令摘要与影响面（含 `kn_id` 与 `--user-id`）
- 仅当用户明确确认"执行推送"后，才可进入实际执行；Never 自动更新网络

推送执行口径（MUST）：

- MUST：阶段五的推送是必执行步骤，不得以"BKN 草案已生成 / 报告已生成 / 文件已落盘"为由跳过；未完成 `ontology bkn push` 之前，阶段五一律视为未完成
- MUST：所有 `ontology` 命令统一委托 `../ontology-core/SKILL.md` 执行，**必须显式传顶层选项 `--user-id <accountId>`，写在子命令之前**，缺省即视为流程违规并立即报错
- MUST：`accountId` 必须从当前会话上下文取得；若未取得，停在阶段五并向用户索要，Never 自行编造或留空
- MUST：命令摘要回显必须包含 `--user-id` 占位（如 `--user-id <accountId>`），不得在摘要中省略该参数
- Never：用未带 `--user-id` 的命令真实执行 `bkn push`
- Never：在阶段五调用 `ontology bkn create`（id 已由阶段二本地 UUID 提供）

命令模板（MUST 按此结构委托 ontology-core 执行）：

```bash
# 直接推送目录（branch 固定为 main，CLI 无需也无法指定其他分支；如需业务域显式声明，再追加 -bd <business-domain>）
# network.bkn 的 id 已由阶段二写入本地生成的 UUID v4，无需再走 bkn create
ontology --user-id <accountId> bkn push <network_dir> [-bd <business-domain>]
```

推送结果后必须执行（MUST，成功/失败均执行）：

1. 完整性检查（草案 vs 线上）
2. 生成业务知识网络 HTML 报告（构建统计 + 可用性评估）

执行口径（MUST）：

- 当推送成功时：按标准口径执行“草案 vs 线上”完整性检查，并生成完整报告
- 当推送失败时：仍必须执行上述两步；完整性检查改为“草案 vs 最近可获取线上快照/已知状态”并标注 `integrity_mode: degraded`
- 当推送失败且无法获取任何线上快照时：完整性检查不得跳过，需输出 `integrity_status: unavailable` 与原因；报告仍必须生成，并显式记录“推送失败 + 线上快照不可用”
- 无论推送成功或失败，都必须在上述两步完成后给出阶段五执行回执
- 归档回执与归档一致性规则由 `archive-protocol` 统一负责，本流程不重复定义

阶段五退出条件（MUST，缺一不可）：

- `ontology --user-id <accountId> bkn push <network_dir>` 已**真实执行**并取得回执（成功或失败均算执行；未执行即未退出）
- `network.bkn` 的 `id` 字段为合法 UUID v4（推送前已通过回读校验，且与 push 命令使用的 `kn_id` 一致）
- 完整性检查与 HTML 报告均已产出
- 若用户未确认“执行推送”，必须停在执行前门禁继续等待，Never 以“草案已完整”“报告已生成”等理由跳过推送、宣告阶段五完成

### HTML 报告生成规范（MUST）

- 报告文件路径：`{network_dir}/reports/bkn_report.html`
- 报告生成前 MUST 先汇总结构化统计数据，并落一份同源 JSON：`{network_dir}/reports/bkn_report_data.json`
- 推荐优先复用模板：`./references/bkn_report_template.html`（允许在不改变统计口径的前提下做样式微调）
- 报告统计口径 MUST 至少包含：
  - 网络概览：`network_name`、`network_id`、`generated_at`、`branch`
  - 结构统计：`object_count`、`relation_count`、`action_count`、`property_total`
  - 映射统计：`mapped_count`、`waived_count`、`blocked_count`、`coverage`
  - 绑定统计：`total_objects`、`bound_objects`、`unbound_objects`、`binding_rate`
  - 风险统计：`unresolved_diffs`、`orphan_objects`、`warnings`
  - 对象展开统计：`object_binding_details[]`（每个对象至少含 `object_name`、`mapped_count`、`waived_count`、`blocked_count`、`coverage`、`binding_view`、`property_rows[]`）
  - 待核实明细：`pending_verification_details[]`（每项至少含 `object_name`、`property_name`、`mapped_field`、`verification_point`、`suggestion`）
  - 可用性解释：`availability_reason`、`availability_judgement[]`（逐条说明“已满足项/待改进项/阻断项”）
- 报告可用性结论 MUST 使用三态：`AVAILABLE` / `PARTIAL` / `UNAVAILABLE`
  - `AVAILABLE`：`coverage = 100%` 且 `blocked_count = 0` 且 `binding_rate >= 80%`
  - `PARTIAL`：`blocked_count = 0` 且（`coverage < 100%` 或 `binding_rate < 80%` 或 `warnings > 0`）
  - `UNAVAILABLE`：`blocked_count > 0` 或关键完整性检查失败
- 页面样式 MUST 可读且美观，至少包含：
  - 顶部摘要区（网络信息 + 结论状态色：绿/黄/红）
  - 可用性判定说明区（明确“为什么是可用/部分可用/不可用”）
  - 核心指标卡片区（结构、映射、绑定、风险）
  - 覆盖率与绑定率进度条
  - 对象类/关系类统计表
  - 属性映射明细表（状态显示为“已映射/已豁免/待处理”，并附原因）
  - 对象类可展开详情区（点击对象类展开属性映射统计与明细）
  - 待核实属性展开区（按对象分组，逐条列出待核实属性与核实要点）
  - 风险与建议区（按高/中/低优先级）
- 样式约束（MUST）：
  - 使用内联 CSS（单文件可离线打开），不得依赖外部 CDN
  - 字体、颜色、间距保持统一，不得出现难辨识对比色
  - 报告中所有百分比统一保留 1 位小数
- 报告输出前 MUST 回读校验：文件存在、非空、包含 `<html` 与关键统计字段
- 若报告生成失败，必须在阶段五回执中明确失败原因与补救建议，不得宣告流程完整成功
- 当状态为 `PARTIAL` 或 `UNAVAILABLE` 时，报告中 MUST 给出可执行改进项（按优先级）以及“达到 AVAILABLE 还缺什么”
- 若仅生成了 `bkn_report_data.json` 而未生成 `bkn_report.html`，视为报告未完成，必须阻断归档成功回执

## 常见失败恢复

- `missing required field 'id'` / `id is None`：阶段二未在 `network.bkn` 写入 UUID v4；打开 `network.bkn`，在 frontmatter 写入本地生成的 UUID v4 字符串后重新 push（Never 改回"留空"或写 `null` / `~`）
- `KnowledgeNetwork.NotFound`：检查是否使用了截断 ID
- `referential integrity`：检查关系类引用对象是否存在
- `id format invalid`：确认写入的是 UUID v4（v4 的第 13 位必须是 `4`、第 17 位为 `8/9/a/b`），重新生成一个 UUID v4 替换后再推送

## 回显模板（对用户）

### 模板：流程路由确认（验证阶段）

```text
### 流程路由确认（验证阶段 | 请确认）
说明：
- 当前识别：这是“新建知识网络”请求
- 将进入流程：新增流程（FLOW_CREATE）
- 你提供的信息摘要：{input_summary}
- 说明：这一步只确认流程方向，还不会开始提取或写入
下一步：你可以回复“确认进入新增流程”，或先补充/修改输入后再继续。
```

### 模板：建模意图澄清（阶段一）

```text
### 建模视角确认（阶段一 | 请确认）
说明：
- 当前处理路径：{A | B | C}
- 建议主视角：{selected_view}
- 是否启用 ER 补充视角：{er_triggered}
- ER 触发依据：{er_trigger_reasons | 无}
- 选择理由：{reason_summary}
下一步：如果你认可这个视角，请回复“确认采用该视角”；如果不认可，我可以马上重选。
```

### 模板：视角确认回执（阶段一）

```text
### 视角确认结果（阶段一 | 结果）
说明：
- 主视角：对象-属性-视图映射
- ER 是否启用：{是 | 否}
- ER 触发依据：{关系较复杂（>5） | 存在多对多或关系带属性 | 存在跨对象关键约束 | 无}
- 说明：ER 仅用于关系骨架梳理，不参与属性映射放行判定
下一步：如无异议，你可以回复“确认清单”，我继续进入清单确认。
```

### 模板：草案确认（阶段二）

```text
### BKN 草案确认（阶段二 | 请确认）
说明：
- 草案目录：{network_dir}
- 校验结果：{validate_summary}
- 关键差异：{draft_highlights}
- 文件复核指引：
  - `{network_dir}/network.bkn`
  - `{network_dir}/object_types/*.bkn`
  - `{network_dir}/relation_types/*.bkn`
- 说明：如果有要调整的内容，直接修改上述 `.bkn` 文件即可
下一步：你可以回复“确认草案”进入阶段三；如果改过文件，回复“已完成草案修改”后我再继续。需要更多细节可说“展开草案详情”或“导出 YAML”。
```

### 模板：阶段四结果确认（阶段四）

```text
### 阶段四结果确认（阶段四 | 请确认）
说明：
- 绑定率：{binding_rate}（{bound_objects}/{total_objects} 对象类已绑定）
- Step 4.4 摘要：{step_4_4_summary}
- Step 4.6 摘要：{step_4_6_summary}
- 关键风险：{risk_summary}
- 执行结论：{pass | warning | blocked}
下一步：如你确认进入推送阶段，请回复“确认”；如需先修复，请回复“先修复”并说明优先项。
```

### 模板：BKN 报告回执（阶段五）

```text
### BKN 报告结果（阶段五 | 结果）
说明：
- 报告文件：{network_dir}/reports/bkn_report.html
- 数据文件：{network_dir}/reports/bkn_report_data.json
- 可用性结论：{可用 | 部分可用 | 不可用}
- 关键统计：对象类 {object_count} / 关系类 {relation_count} / 属性 {property_total} / 覆盖率 {coverage}
- 风险摘要：阻断项 {blocked_count} / 未决差异 {unresolved_diffs} / 孤悬对象 {orphan_objects}
下一步：如果你希望我继续处理报告中的高优先级问题，直接说“继续修复”就行。
```

