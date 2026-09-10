# Metric 命令参考

指标（Metric）定义管理：CRUD、搜索、校验、查询数据、试运行。

指标是知识网络级别的计算定义，独立于对象类的 `logic_properties`。通过 `ontology metric` 命令组进行管理。

## 概览

```bash
ontology --user-id <accountId> metric list <kn_id>                              # 列出指标
ontology --user-id <accountId> metric get <kn_id> <metric_id>                   # 获取指标
ontology --user-id <accountId> metric create <kn_id> --body '<json>'            # 创建指标
ontology --user-id <accountId> metric update <kn_id> <metric_id> --body '<json>' # 更新指标
ontology --user-id <accountId> metric delete <kn_id> <metric_id> [-y]           # 删除指标
ontology --user-id <accountId> metric search <kn_id> --body '<json>'            # 搜索指标
ontology --user-id <accountId> metric validate <kn_id> --body '<json>'          # 校验指标定义
ontology --user-id <accountId> metric query <kn_id> <metric_id> --body '<json>' # 查询已保存指标的数据
ontology --user-id <accountId> metric dry-run <kn_id> --body '<json>'           # 试运行指标定义（不保存）
```

## MetricDefinition 主体结构

创建/更新指标时，`--body` 传一个 **MetricDefinition JSON 对象**。CLI 会自动包装成 `{"entries": [<body>]}` 发给后端。

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | ✅ | 指标名称 |
| `metric_type` | string | ✅ | 指标类型：`atomic`（原子）、`derived`（派生）、`composite`（复合） |
| `scope_type` | string | ✅ | 作用域类型：`object_type`（对象类级） |
| `scope_ref` | string | ✅ | 绑定的对象类 ID |
| `calculation_formula` | object | ✅ | 计算公式（见下方） |
| `unit_type` | string | | 单位类型：`numUnit` / `countUnit` / `percent` / `currencyUnit` 等 |
| `unit` | string | | 度量单位：`none` / `K` / `bit` / `%` / `piece` / `ton` 等 |
| `tags` | string[] | | 标签列表 |
| `comment` | string | | 描述说明 |
| `time_dimension` | object | | 时间维度配置（见下方） |
| `analysis_dimensions` | object[] | | 分析维度列表，每项 `{"name": "<字段名>"}` |

### calculation_formula 结构

```json
{
  "aggregation": {
    "property": "<聚合属性名>",
    "aggr": "<聚合方式>"
  },
  "group_by": [
    { "property": "<分组字段名>" }
  ],
  "order_by": [
    { "property": "<排序字段名>", "direction": "desc" }
  ],
  "having": {
    "field": "__value",
    "operation": ">",
    "value": 100
  },
  "condition": { }
}
```

**聚合方式（aggr）枚举**：

| 值 | 语义 |
|----|------|
| `count_distinct` | 去重计数（COUNT DISTINCT） |
| `count` | 计数（COUNT） |
| `sum` | 求和（SUM） |
| `max` | 最大值（MAX） |
| `min` | 最小值（MIN） |
| `avg` | 平均值（AVG） |

### time_dimension 结构

```json
{
  "property": "<时间字段名>",
  "default_range_policy": "last_24h"
}
```

`default_range_policy` 可选值：`last_1h` / `last_24h` / `calendar_day` / `none`

## list — 列出指标

```bash
ontology --user-id <accountId> metric list <kn_id> [--limit <n>] [--offset <n>] [--branch <b>] [-bd value] [--pretty]
```

## get — 获取指标

```bash
ontology --user-id <accountId> metric get <kn_id> <metric_id> [--branch <b>] [-bd value] [--pretty]
```

## create — 创建指标

```bash
ontology --user-id <accountId> metric create <kn_id> --body '<json>' [--strict-mode] [--branch <b>] [-bd value] [--pretty]
```

- `--body`：单个 MetricDefinition JSON 对象（CLI 自动包装 entries 数组）
- `--strict-mode`：严格模式校验

**示例**：

```bash
ontology --user-id <accountId> metric create d4rt3135s3q8va76m8fd --body '{
  "name": "各仓库物品种类数",
  "metric_type": "atomic",
  "scope_type": "object_type",
  "scope_ref": "ot_abc123",
  "unit_type": "countUnit",
  "unit": "none",
  "calculation_formula": {
    "aggregation": {
      "property": "item_code",
      "aggr": "count_distinct"
    },
    "group_by": [
      { "property": "warehouse_name" }
    ]
  },
  "analysis_dimensions": [
    { "name": "warehouse_name" }
  ]
}'
```

## update — 更新指标

```bash
ontology --user-id <accountId> metric update <kn_id> <metric_id> --body '<json>' [--branch <b>] [-bd value] [--pretty]
```

- `--body`：完整的 MetricDefinition JSON 对象

## delete — 删除指标

```bash
ontology --user-id <accountId> metric delete <kn_id> <metric_id> [-y] [--branch <b>] [-bd value]
```

- `-y`：跳过确认提示

## search — 搜索指标

```bash
ontology --user-id <accountId> metric search <kn_id> --body '<json>' [--branch <b>] [-bd value] [--pretty]
```

**示例**：

```bash
ontology --user-id <accountId> metric search d4rt3135s3q8va76m8fd --body '{
  "query": "仓库 物品种类",
  "limit": 10
}'
```

## validate — 校验指标定义

```bash
ontology --user-id <accountId> metric validate <kn_id> --body '<json>' [--strict-mode] [--branch <b>] [-bd value] [--pretty]
```

- `--body`：单个 MetricDefinition JSON 对象（CLI 自动包装 entries 数组）

## query — 查询已保存指标的数据

```bash
ontology --user-id <accountId> metric query <kn_id> <metric_id> --body '<json>' [--fill-null] [--branch <b>] [-bd value] [--pretty]
```

- `--fill-null`：填充空值

**请求体结构**：

```json
{
  "time": {
    "start": 1704067200000,
    "end": 1735689600000,
    "instant": true,
    "step": "day"
  },
  "condition": { },
  "analysis_dimensions": ["warehouse_name"],
  "order_by": [
    { "property": "__value", "direction": "desc" }
  ],
  "having": {
    "field": "__value",
    "operation": ">",
    "value": 0
  },
  "limit": 100
}
```

## dry-run — 试运行指标定义

```bash
ontology --user-id <accountId> metric dry-run <kn_id> --body '<json>' [--fill-null] [--branch <b>] [-bd value] [--pretty]
```

试运行不保存指标定义，直接执行查询并返回结果。

**请求体结构**（MetricDryRunRequest）：

```json
{
  "metric_config": {
    "name": "各仓库物品种类数",
    "metric_type": "atomic",
    "scope_type": "object_type",
    "scope_ref": "ot_abc123",
    "calculation_formula": {
      "aggregation": { "property": "item_code", "aggr": "count_distinct" },
      "group_by": [ { "property": "warehouse_name" } ]
    }
  },
  "time": {
    "start": 1704067200000,
    "end": 1735689600000,
    "instant": true
  },
  "limit": 100
}
```

## 端到端示例

```bash
# 列出知识网络下的指标
ontology --user-id <accountId> metric list d4rt3135s3q8va76m8fd

# 查询已保存指标的数据
ontology --user-id <accountId> metric query d4rt3135s3q8va76m8fd dafuk0426hkc73f4e0m0 --body '{"time":{"instant":true},"limit":100}'

# 试运行一个指标定义（不保存）
ontology --user-id <accountId> metric dry-run d4rt3135s3q8va76m8fd --body '{"metric_config":{"name":"test","metric_type":"atomic","scope_type":"object_type","scope_ref":"ot_xxx","calculation_formula":{"aggregation":{"property":"item_code","aggr":"count_distinct"},"group_by":[{"property":"warehouse_name"}]}},"time":{"instant":true},"limit":100}'

# 创建新指标
ontology --user-id <accountId> metric create d4rt3135s3q8va76m8fd --body '{"name":"各仓库物品种类数","metric_type":"atomic","scope_type":"object_type","scope_ref":"ot_abc123","unit_type":"countUnit","unit":"none","calculation_formula":{"aggregation":{"property":"item_code","aggr":"count_distinct"},"group_by":[{"property":"warehouse_name"}]},"analysis_dimensions":[{"name":"warehouse_name"}]}'

# 校验指标定义
ontology --user-id <accountId> metric validate d4rt3135s3q8va76m8fd --body '{"name":"各仓库物品种类数","metric_type":"atomic","scope_type":"object_type","scope_ref":"ot_abc123","calculation_formula":{"aggregation":{"property":"item_code","aggr":"count_distinct"},"group_by":[{"property":"warehouse_name"}]}}'
```
