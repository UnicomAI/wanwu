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

- `--body`：指标定义 JSON
- `--strict-mode`：严格模式校验

## update — 更新指标

```bash
ontology --user-id <accountId> metric update <kn_id> <metric_id> --body '<json>' [--branch <b>] [-bd value] [--pretty]
```

## delete — 删除指标

```bash
ontology --user-id <accountId> metric delete <kn_id> <metric_id> [-y] [--branch <b>] [-bd value]
```

- `-y`：跳过确认提示

## search — 搜索指标

```bash
ontology --user-id <accountId> metric search <kn_id> --body '<json>' [--branch <b>] [-bd value] [--pretty]
```

## validate — 校验指标定义

```bash
ontology --user-id <accountId> metric validate <kn_id> --body '<json>' [--strict-mode] [--branch <b>] [-bd value] [--pretty]
```

## query — 查询已保存指标的数据

```bash
ontology --user-id <accountId> metric query <kn_id> <metric_id> --body '<json>' [--fill-null] [--branch <b>] [-bd value] [--pretty]
```

- `--fill-null`：填充空值

## dry-run — 试运行指标定义

```bash
ontology --user-id <accountId> metric dry-run <kn_id> --body '<json>' [--fill-null] [--branch <b>] [-bd value] [--pretty]
```

试运行不保存指标定义，直接执行查询并返回结果。

## 端到端示例

```bash
# 列出知识网络下的指标
ontology --user-id <accountId> metric list <kn_id>

# 查询已保存指标的数据
ontology --user-id <accountId> metric query <kn_id> <metric_id> --body '{"time":{"start":"2024-01-01","end":"2024-12-31"}}'

# 试运行一个指标定义（不保存）
ontology --user-id <accountId> metric dry-run <kn_id> --body '{"metric_config":{...},"time":{...}}'

# 创建新指标
ontology --user-id <accountId> metric create <kn_id> --body '{"name":"月销售额","metric_type":"atomic",...}'

# 校验指标定义
ontology --user-id <accountId> metric validate <kn_id> --body '{"name":"月销售额",...}'
```
