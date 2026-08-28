# Vega 命令参考

Vega 可观测平台：Catalog 管理、数据资源查询、连接器类型、健康巡检。

## 概览

```bash
ontology vega                         # 帮助信息
ontology vega health                  # 服务健康检查
ontology vega stats                   # Catalog 数量统计
ontology vega inspect                 # 聚合诊断（health + catalog 数量 + 运行中的 discover 任务）
```

## Catalog

```bash
ontology vega catalog list [--status healthy|degraded|unhealthy|offline|disabled] [--limit N] [--offset N]
ontology vega catalog get <id>
ontology vega catalog health <ids...> | --all
ontology vega catalog test-connection <id>
ontology vega catalog discover <id> [--wait]
ontology vega catalog resources <id> [--category table|index|...] [--limit N]
```

## Resource

```bash
ontology vega resource list [--catalog-id <id>] [--category table] [--status active] [--limit N] [--offset N]
ontology vega resource get <id>
ontology vega resource query <id> -d '<json-body>'
ontology vega resource preview <id> [--limit N]
```

## Connector Type

```bash
ontology vega connector-type list
ontology vega connector-type get <type>
```

## 公共参数

所有子命令支持：

- `-bd, --biz-domain <s>` — 业务域（默认 `bd_public`）
- `--pretty` — 格式化 JSON 输出（默认开启）

## 端到端示例

```bash
# 巡检
ontology vega inspect
ontology vega catalog health --all

# 查看 catalog 下的资源
ontology vega catalog list
ontology vega catalog resources <catalog-id> --category table

# 预览资源数据
ontology vega resource preview <resource-id> --limit 5

# 查询资源数据
ontology vega resource query <resource-id> -d '{"page": 1, "limit": 10}'

# 查看连接器类型
ontology vega connector-type list
```
