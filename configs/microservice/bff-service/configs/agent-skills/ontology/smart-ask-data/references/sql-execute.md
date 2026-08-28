# 步骤 3–4：SQL 生成 + 执行

**步骤 3（SQL 生成）** 在 **smart-data-analysis 的 LLM** 内完成；本 skill 内 **不内置 LLM**，也 **不生成 SQL**。
**步骤 4（SQL 执行）** 由本 skill 描述命令形态，委托 [ontology-core](../../ontology-core/SKILL.md) 执行。

> **执行委托**：本文件仅 **描述** 命令形态；**Never** 由 smart-ask-data 直接执行 `ontology` CLI。

## 步骤 3：SQL 生成（由 smart-data-analysis 完成）

smart-data-analysis 基于：

- 用户问题（中文问数意图）
- 步骤 2 的 schema 摘要（对象类、字段、dataview-id、关系）

生成 **SELECT/WITH** SQL，须满足：

- **未拿到 `meta_table_name` 不许进入 SQL 生成**：步骤 2.4 `dataview get` 必须为每个参与的 dataview 都跑过一次，把返回的 `meta_table_name` 缓存进上下文。若发现 schema 摘要里某个 dataview 缺 `meta_table_name`，必须回 2.4 补齐；**不允许**用 dataview-id、对话历史里见过的旧表名、或者凭命名直觉拼一个表名出来。
- **只允许 SELECT/WITH**：`dataview query --sql` 默认拒绝写操作；不得使用 `--raw-sql` 绕过。
- 字段、表名必须 **完全来自步骤 2 schema 摘要**；禁止编造。
- **`FROM/JOIN` 必须用三段式表名** `<catalog>.<schema>.<table>`，取自步骤 2.4 `dataview get` 返回的 `meta_table_name`（形如 `mysql_xxxxxxxx."demo"."bom_event"`）。**禁止**写裸表名（如 `FROM bom_event`）——裸名 mdl-uniquery 解析为空集或报错。
- 表（FROM/JOIN）必须是步骤 2 取到的 dataview-id 对应的视图；多对象类时跨 dataview JOIN 须保证字段一致。
- **标识符过滤的防御式写法**：用户给的形如"业务编号 / 编码 / 短码"的字符串作为 WHERE 过滤值时，目标 schema 里往往同时存在多列标识符（典型如 `*_id` / `*_code`，也可能有 `*_no` / `*_sku` 等），列名和列注释都未必能区分这些列各自持有的值格式。除非已通过其它手段确认值落在具体哪一列，否则在 SQL 中对该实体的所有候选标识符列做 OR 兜底：

  ```sql
  WHERE (<id_col> = '<value>' OR <code_col> = '<value>' [OR <其它候选列> = '<value>'])
  ```

  候选列从步骤 2 schema 摘要里挑该实体的标识符类列；代价是多几个 OR 条件，收益是消除字段错配导致整查询返回空集的整类错误。
- 含聚合时同时给出 `GROUP BY`；含时间范围时谓词与字段类型对齐。
- 输出限制：默认在 SQL 末尾或外层加 `LIMIT`（建议 ≤ 200）；命令侧也可用 `--limit` 兜底。

生成结果作为 **入参** 传入步骤 4。

## 步骤 4：SQL 执行

### 命令形态（默认：跨表/聚合用 SQL）

```bash
ontology --user-id <accountId> dataview query <dataview-id> \
  --sql "<LLM 生成的 SELECT SQL>" \
  [--limit 200] [--offset 0] [--need-total] [-bd bd_public] [--pretty]
```

- **`<dataview-id>`**：取自步骤 2 中相关 object-type 的 `data_source.id`（要求 `data_source.type == "data_view"`）。
- **`--sql`**：完整 SELECT/WITH 语句；省略 `--sql` 时使用 view 默认 SQL。
- **`--limit / --offset`**：分页；优先用 `--limit` 兜底，避免一次拉过大。
- **`--need-total`**：需要总数时附加（按 mdl-uniquery 行为为准）。
- 引擎：`dataview query` 路由到 mdl-uniquery。

### 跨 dataview 场景

若 SQL 涉及多个 dataview，需先确定一个 **主 dataview**（通常含主事实表），将其它对象类的字段以 `JOIN <other-dataview-name>` 写入 SQL；具体表名以步骤 2 schema 摘要里 dataview 的实际名称为准。

> 多 dataview JOIN 写法依赖 mdl-uniquery 的命名空间约定；若不可用，回退到先按对象类分别拉取，再由 smart-data-analysis 用提示工程在结果层合并。

### 备选命令形态（简单单表过滤；非 SQL）

对 **单一对象类、纯过滤 + 分页** 的简单问题，可不走 SQL，直接用对象类查询：

```bash
ontology --user-id <accountId> bkn object-type query <kn-id> <ot-id> '<filter-json>' \
  [--limit 50] [--search-after '<json-array>'] [-bd bd_public] [--pretty]
```

- `<filter-json>`：以网关约定为准（`{"_instance_identities":[...]}` 等）。
- 不支持聚合 / GROUP BY / JOIN；遇此类需求回到 SQL 路径。

### 逻辑属性取值路径（计算字段结算）

当步骤 2 的 `bkn object-type get` 返回的 `logic_properties` 数组非空，且用户要的指标名命中其中某项的 `name` 时，走此路径。此路径 **不生成 SQL**，直接用固定 body 结构的取值命令，由服务端 box 引擎结算后返回计算值。

```bash
ontology --user-id <accountId> bkn object-type properties <kn-id> <ot-id> \
  '{"_instance_identities":[{"<primary-key>":"<value>"}],"properties":["<logic_property_name>"]}' \
  [-bd bd_public] [--pretty]
```

- **`<kn-id>` / `<ot-id>`**：来自步骤 2 schema 发现。
- **`_instance_identities`**：目标实例的主键标识数组。主键字段名取自 `bkn object-type get` 返回的 `primary_keys`（如 `inventory_id`）。
- **`properties`**：要结算的逻辑属性 `name` 列表，取自 `bkn object-type get` 返回的 `logic_properties[].name`。
- 返回的是 **服务端 box 引擎结算后的结果值**（不是入参绑定）。

> **逻辑属性路径 vs SQL 路径**：用户问题中的指标名出现在 `logic_properties` 中 → 走逻辑属性路径；否则走 SQL 路径。两条路径禁止混用。

## 结果回执契约

ontology-core 把 `dataview query` 或 `bkn object-type properties` 的原始 JSON 返回给 smart-ask-data；本 skill 把如下两块交给最终回复：

1. **执行的 SQL 或逻辑属性取值命令**（可脱敏，不可省略）。
2. **关键结果数据**（表格 / 行记录 / 聚合数值 / 计算值；按"注意事项"原样呈现）。

## 注意事项

- **结果展示硬约束**：若返回非空结果，最终回复 **必须同时** 给出执行的 SQL 或逻辑属性取值命令 + 关键结果数据，不得只给口头结论。
- **结果为空**：直接说"未查询到符合条件的数据"，不得编造；并建议下一步（调整时间范围、口径、或换 KN）。
- **写操作禁止**：不允许 `INSERT / UPDATE / DELETE / DDL`；不允许 `--raw-sql`。
- **`--user-id <accountId>` 必传**。
- **网关**：由 ontology-core 侧 `ONTOLOGY_BASE_URL` 承担，本 skill 命令体内不出现 `--base-url`。
- 本部署 ontology CLI **无须 token**；命令体内 **不出现** `--token` / `auth.token` / `Authorization`。
- 命令报错时直接如实反馈给用户，不要伪造结果或尝试登录刷新凭证。
