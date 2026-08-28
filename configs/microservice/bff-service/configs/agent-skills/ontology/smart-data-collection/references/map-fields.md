# Phase 0b：字段语义映射 + 用户逐条确认

输入：0a 步骤产出的 Markdown 表 + 字段样值 JSON。
输出：用户确认后的字段映射 JSON，作为 0c 写出 CSV 的唯一依据。

本步骤**不**生成 INSERT SQL、**不**连 mysql、**不**写 `.sql` 文件、**不**做事务控制——这些都是参考稿 §3 / §4 的内容，已被本 skill 显式排除。

## 2.1 确定知识网络

- **默认 KN**：`e0e2ed66-8eff-4f14-8e15-9fd6171f8e53`。
- **用户传入优先**：用户在调用 smart-data-collection 时显式给的 `kn_id` 覆盖默认值。
- **参数名统一**：`kn_id`（沿用 smart-data-analysis 的「知识网络声明」表，下游 skill 入参一致）。

## 2.2 取对象类字段定义

委托 [ontology-core](../../ontology-core/SKILL.md) 执行 ontology CLI（注意：`--user-id` 必传，且写在子命令前）。

### 2.2.1 列出全部对象类，按主题挑候选

```bash
ontology --user-id <accountId> bkn object-type list <kn_id> \
  | jq '.entries[] | {id, name, comment, primary_keys, display_key}'
```

- 先用 `list` 拿到当前 KN 全部 object_type 摘要，再用 0a 的"表名 / 文档标题"在 `name` / `comment` 里做语义匹配。
- 单 KN 全量 list 默认含完整 `data_properties`，常达 30~50KB 易截断，必须用 jq 投影到摘要（id / name / comment / primary_keys / display_key）。定位到目标后再用 §2.2.2 的 `bkn object-type get <kn_id> <ot_id>` 取完整字段。
- 命中多个候选时，由 smart-data-analysis 顶层路由让用户挑（不在本 reference 里做选型）。

### 2.2.2 取字段定义（映射目标）

```bash
ontology --user-id <accountId> bkn object-type get <kn_id> <ot_id>
```

从返回里取 `data_properties[]`，每个字段包含：
- `name`（字段名，例 `product_id`）；
- `display_name`（中文显示名，例 "产品编号"）；
- `data_type`（`STRING` / `INT` / `DECIMAL` / `DATETIME` / `BOOL` / `ENUM`...）；
- `required`（true/false）；
- 可能的 `enum_values[]` / `length` / `precision` 等约束。

**`data_properties` 是映射目标的唯一字段定义来源**——别从 dataview 的字段去推，别从用户口述去推。

## 2.3 映射规则 + 强制确认场景

把 0a 的「字段样值 JSON」与 `data_properties[]` 做语义对齐：

1. **字段名匹配**：优先按 `name` / `display_name` 命中（含中英文同义、单复数、缩写、大小写）。
2. **数据类型匹配**：字符串 → STRING；纯数字 → INT/DECIMAL；`YYYY-MM-DD[ HH:MM:SS]` → DATETIME；`true/false/0/1` → BOOL；枚举值需落在 `enum_values[]` 白名单。
3. **必填校验**：`required=true` 的字段必须有对应解析数据。
4. **生成映射对照表**。

**强制确认（任一触发，必须停下来）**：

| 触发场景 | 例子 |
|---|---|
| 名称不一致 | 解析字段 "数量" vs 对象类字段 `quantity` |
| 类型不一致 | 解析样值 `"12000.00"` vs 对象类 `INT`；日期格式 `2026/06/11` vs 对象类要求 `YYYY-MM-DD` |
| 必填字段缺失对应解析数据 | `required=true` 的 `customer_id` 在 0a 里没出现 |
| 歧义（一对多） | 解析字段 "金额" 同时能映射到 `amount` 和 `total_price` |

未取得用户决策前，**禁止**进入 0c 写 CSV、**禁止**进入第 1 步主流程。

## 2.4 用户确认流程

### 步骤 1：出示对照表

逐行展示所有字段（完全一致的也列出，便于用户整体核对）：

| 解析字段 | 解析类型 | 解析样值 | → | 对象类字段名 | 对象类字段显示名称 | 对象类类型 | 是否必填 | 差异类型 |
|---|---|---|---|---|---|---|---|---|
| 订单号 | STRING | O001 | → | `order_id` | 订单编号 | STRING | 是 | 名称不一致 |
| 数量 | STRING("2") | 2 | → | `quantity` | 数量 | INT | 是 | 名称不一致 + 类型不一致 |
| 金额 | STRING("12000") | 12000 | → | `amount` / `total_price`? | 金额 / 总价? | DECIMAL | 是 | 歧义 |
| 客户名称 | STRING | 张三 | → | `customer_name` | 客户名称 | STRING | 否 | 一致（无需确认） |

### 步骤 2：逐条等待用户决策

**一次只就一条差异项向用户提问，拿到该条决策后再问下一条。严禁把多条差异揉成一个 `是否确认以上映射？` 的总确认问题（无论单选还是列表），那等于没有逐条确认。** 整表只在步骤 1 用于通览，不能拿整表当一次性确认入口。

每条差异项必须单独向用户给出以下四个动作让其选一个：

1. **确认**：采用建议映射。
2. **改为指定字段**：用户指定改映射到对象类的另一个字段。
3. **跳过**：该字段不写入数据库（必填字段被跳过时需用户**再次确认**）。
4. **终止**：放弃整个采集任务。

「一致（无需确认）」的行不必逐条问；只有 §2.3 触发强制确认的差异项（名称/类型/必填/歧义）才需逐条走本步骤。

### 步骤 3：阻断条件

以下情况下**不得**继续进入 0c：

- 任一差异项未取得用户决策。
- 必填字段被跳过但用户未二次确认。
- 用户选择「终止」。

### 步骤 4：记录映射

将用户最终确认后的映射关系写入下方 **输出契约** 的 JSON 结构，作为 0c 写出 CSV 的唯一依据。

## 输出契约（0b 步骤的统一产物）

```json
{
  "kn_id": "e0e2ed66-8eff-4f14-8e15-9fd6171f8e53",
  "object_type_id": "d50jqrr5s3q8ofn0dscg",
  "object_type_name": "sales_order",
  "mappings": [
    {
      "source_field": "订单号",
      "target_field": "order_id",
      "type": "STRING",
      "required": true,
      "action": "confirmed"
    },
    {
      "source_field": "数量",
      "target_field": "quantity",
      "type": "INT",
      "required": true,
      "action": "confirmed",
      "transform": "to_int"
    },
    {
      "source_field": "金额",
      "target_field": "total_price",
      "type": "DECIMAL",
      "required": true,
      "action": "user_chose_target"
    },
    {
      "source_field": "备注",
      "target_field": null,
      "required": false,
      "action": "skip"
    }
  ]
}
```

字段含义：

- `source_field`：0a 解析出的原始字段名。
- `target_field`：用户确认后映射到的 object-type 字段；`action=skip` 时为 `null`。
- `type` / `required`：直接抄 `data_properties[]`。
- `action`：`confirmed` / `user_chose_target` / `skip` / `terminated` 之一。
- `transform`（可选）：字符串到目标类型的转换提示，例 `to_int` / `parse_date(YYYY-MM-DD)` / `enum_check`。

0c 步骤读这个 JSON，按 `mappings` 把 0a 的 Markdown 表写成 CSV：
- 列名 = `target_field`（`action=skip` 的列直接丢掉）；
- 值按 `transform` 做最小转换；
- 编码 UTF-8、字段按 RFC 4180 转义、首行表头。

## 明确不做的事

1. **不**生成 `INSERT INTO ...` SQL。
2. **不**生成 `.sql` 文件。
3. **不**调用 `mysql` 命令、**不**用 `pymysql` / `sqlalchemy+pymysql` 等直连库。
4. **不**做事务控制（`START TRANSACTION` / `COMMIT` / `ROLLBACK`）。
5. **不**读 / **不**解密 `ds get` 返回的 `bin_data.password`。

以上全部是 SKILL.md IRON RULE 的延伸；0c 完成后，第 1～10 步主流程依然由 `ontology ds import-csv` 走 data-flow 写入。
