# 数据源命令参考

数据源连接、发现与管理。

## 命令

```bash
ontology ds list [--keyword <kw>] [--type <db_type>]
ontology ds get <datasource_id>
ontology ds connect <db_type> <host> <port> <database> --account <user> --password <pass> [--schema <s>] [--name <n>]
ontology ds tables <datasource_id> [--keyword <kw>]
ontology ds delete <datasource_id> [--yes]
ontology ds import-csv <datasource_id> --files <glob_or_list> [--table-prefix <p>] [--batch-size <n>]
```

## 支持的数据库类型

mysql, postgresql, sqlserver, oracle, clickhouse, hive, opensearch, elasticsearch 等。

## 端到端示例

```bash
# 连接 MySQL
ontology ds connect mysql db.example.com 3306 erp --account root --password pass123

# 查看表结构
ontology ds tables ds-abc123

# 连接后创建知识网络
ontology bkn create-from-ds ds-abc123 --name "erp-kn" --tables "orders,products" --build

# 导入 CSV 到已有数据源
ontology ds import-csv ds-abc123 --files "*.csv" --table-prefix my_

# 从 CSV 一键创建知识网络
ontology bkn create-from-csv ds-abc123 --files "物料.csv,库存.csv" --name "supply-kn"
```
