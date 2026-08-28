# 通用 API 调用参考

直接调用 REST API，自动注入认证。类似 curl。

## 命令

```bash
ontology call <url> [-X METHOD] [-H "Name: value"] [-d BODY] [--data-raw BODY]
             [--url URL] [--pretty] [--verbose] [-bd value]
```

（`ontology curl` 为 `call` 的别名。）

| 参数 | 说明 |
|------|------|
| `<url>` | API 路径（如 `/api/ontology-manager/v1/knowledge-networks`）；也可用 `--url` 指定 |
| `--url` | 与位置参数 `<url>` 二选一，显式传入请求路径 |
| `-X`, `--request` | HTTP 方法（默认 GET） |
| `-d`, `--data`, `--data-raw` | JSON 请求体；三者等价。提供 body 且未手动设置 `Content-Type` 时，CLI 会设置 **`Content-Type: application/json`** |
| `-H`, `--header` | 额外请求头（可重复），格式 `Name: value` |
| `-bd`, `--biz-domain` | 覆盖 `x-business-domain`（默认来自当前平台配置） |
| `-v`, `--verbose` | 打印请求信息到 stderr |
| `--pretty` | Pretty-print JSON 输出（默认开启） |

## 注意事项

- 若已通过 `-H 'Content-Type: ...'` 指定 Content-Type，则**不会**被自动覆盖。
- 若 `-d` / `--data` / `--data-raw` 与 GET 同时使用，方法会自动升级为 **POST**（与 curl 行为一致）。

## 示例

```bash
# GET
ontology call /api/ontology-manager/v1/knowledge-networks

# 使用 --url
ontology call --url /api/ontology-manager/v1/knowledge-networks

# POST
ontology call /api/ontology-manager/v1/knowledge-networks -X POST -d '{"name": "test", "branch": "main"}'

# 带自定义 header
ontology call /api/some-service/v1/endpoint -H 'X-Custom: value'

# 指定 business domain
ontology call /api/ontology-manager/v1/knowledge-networks -bd my_domain
```
