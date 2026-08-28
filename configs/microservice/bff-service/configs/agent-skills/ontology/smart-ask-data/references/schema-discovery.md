# 步骤 2：Schema 发现（候选对象类与字段）

在 **问数主流程第 2 步** 调用：在已解析的 `kn_id` 下，定位与用户问题相关的 **对象类**（object-type），并取出每个对象类的 **字段** 与 **绑定的 dataview-id**（从 `data_source` 提取），供编排层 LLM 在第 3 步生成 SQL 使用。

> **执行委托**：本文件仅 **描述** 命令形态；**Never** 由 smart-ask-data 直接执行 `ontology` CLI，实际调用统一由 [ontology-core](../../ontology-core/SKILL.md) 完成。

## 与本流程的衔接

- **输入**：步骤 1 选定的 `kn_id`。
- **产出**：交给编排层 LLM 的"schema 摘要"——
  - 相关 object-type 列表（id / name / description）
  - 每个相关 object-type 的字段清单（properties）
  - 每个 object-type 后端绑定的 dataview-id（从 `data_source.id` 取，要求 `data_source.type == "data_view"`；步骤 4 SQL 执行需要）

## 命令形态（委托 ontology-core 执行）

### 2.1 列出全部对象类

```bash
ontology --user-id <accountId> bkn object-type list <kn-id> [-bd bd_public] [--pretty]
```

- 返回 KN 中所有 object-type 的 schema 摘要。
- **输出体积警告**：默认返回每个对象类完整 `data_properties`（含索引配置），单 KN 常达 30~50KB，易被截断。**仅为选定相关 ot-id 时**用 jq 投影到轻量摘要：

  ```bash
  ontology --user-id <accountId> bkn object-type list <kn-id> [-bd bd_public] \
    | jq '.entries[] | {id, name, comment, primary_keys, display_key}'
  ```

  从数十 KB 压到 ~2KB；选定相关 ot-id 后再用 2.3 `bkn object-type get` 拿完整字段。

### 2.2 取单个对象类的字段与 dataview-id

对步骤 2.1 / 2.2 选出的每个相关 object-type，取详情：

```bash
ontology --user-id <accountId> bkn object-type get <kn-id> <ot-id> [-bd bd_public] [--pretty]
```

返回（以网关为准）通常包含：

- `id` / `name` / `display_key` / `primary_key`
- `data_source`：对象类绑定的后端数据源，含 `type` / `id` / `name`。**关键** — 当 `type == "data_view"` 时 `data_source.id` 即步骤 4 `dataview query --sql` 需要的 dataview-id；其它 `type` 不可用于 `dataview query --sql`
- `properties`：字段名 + 类型 + 描述
- `logic_properties`：逻辑属性（计算字段）定义数组，每项含 `name`（逻辑属性名）、`display_name`、`type`（如 `tool`）、`execution_unit`（box 引擎执行单元）、`parameters`（入参绑定：哪个字段映射到哪个参数、常量值等）。**当用户问题中的指标名命中此数组时，走逻辑属性取值路径（`bkn object-type properties`），不走 SQL。**
- `tags` / `comment` / 关联关系类型 等

### 2.3 拿三段式表名（生成 SQL 必须）

`dataview query --sql` 的 SQL 中 `FROM/JOIN` **必须用三段式表名** `<catalog>.<schema>.<table>`；裸表名（如 `FROM bom_event`）会返回空集或报错。

每个 dataview 调一次：

```bash
ontology --user-id <accountId> dataview get <dataview-id>
```

返回里关注：

- `meta_table_name`：完整三段式表名，形如 `mysql_xxxxxxxx."demo"."bom_event"`，**直接抄进 SQL 的 FROM 子句**（双引号可保留，也可去掉）
- `fields[]`：与 2.3 的 `properties` 等价的字段清单（name / type / display_name / comment），单独使用本步即可获得 schema，不必再回 2.3
- `datasource_id`：如需进一步确认 catalog，可 `ds get <datasource-id>` 看 `bin_data.catalog_name`（通常已等于 `meta_table_name` 第一段）

> **2.3 是 SQL 生成的硬前置**：只要后续会跑 `dataview query --sql`，**每个参与的 dataview 都必须先调一次 `dataview get`** 并把 `meta_table_name` 缓存到上下文；不许凭 dataview-id 或对话记忆猜表名。
> 仅当不打算生成 SQL（如纯用 `bkn object-type query` 单表过滤）时，才能省 2.4。
> 2.2 与 2.3 的字段清单等价；如仅生成 SQL，**可只走 2.3 替代 2.2**。

## LLM 交付契约（给编排层）

smart-ask-data 把以下结构化"schema 摘要"交回给 smart-data-analysis 的 LLM 用于生成 SQL：

```text
KN: <kn_id> (<kn_name>)
相关对象类：
  - <ot-id-1> (<name>)
      dataview-id: <dv-id-1>            # 取自 data_source.id（type == "data_view"）
      meta_table_name: <catalog>.<schema>.<table>   # 来自 2.4 dataview get；SQL FROM/JOIN 直接抄
      字段: <prop1> (<type>, <desc>), <prop2> (...), ...
      逻辑属性: <lp1> (<display_name>, type=<tool>, ...), <lp2> (...), ...   # 取自 logic_properties
  - <ot-id-2> (<name>)
      dataview-id: <dv-id-2>
      meta_table_name: ...
      字段: ...
      逻辑属性: ...
关系（如有）:
  - <ot-id-1> -[<rel-name>]→ <ot-id-2>
```

- 字段表必须 **来自命令返回**，不得编造或脑补类型/含义。
- 若 `data_source` 缺失，或 `data_source.type` 不是 `data_view`，对应 object-type **不能** 用于步骤 4 SQL 执行；需要换 object-type 或换 KN。

## 注意事项

- `--user-id <accountId>` **必传**。
- **禁止** 强行在不匹配的 KN 上继续 schema 发现；中止本任务并回到步骤 1 重选 KN，或让用户改换 KN。
- 命令报错时直接如实反馈给用户，不要伪造结果或尝试登录刷新凭证。本部署 ontology CLI **无须 token**。
