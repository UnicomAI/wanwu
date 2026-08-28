---
name: mcp2skill
description: 将 MCP Server 的工具列表转换为 Skill 格式的结构化 Markdown 文档
---

# mcp2skill

使用mcp2skill命令(已内置)将 MCP Server 的工具定义转换为 Skill 格式，生成 SKILL.md 概览和每个工具的详情文档。

## Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `streamableUrl` | string | Conditional | MCP Server Streamable HTTP 端点 URL（与 sseUrl 至少填一个） |
| `sseUrl` | string | Conditional | MCP Server SSE 端点 URL（与 streamableUrl 至少填一个） |
| `name` | string | No | Skill 名称，留空时从工具列表自动推断 |
| `description` | string | No | Skill 描述，留空时自动生成 |
| `transport` | string | No | 传输类型：`streamable`（默认）或 `sse` |
| `output` | string | No | 输出目录，默认当前目录 |
| `timeout` | string | No | 连接超时时间，默认 30s |
| `apiAuth` | object | No | API 认证配置 |
| `headers` | object | No | 自定义 HTTP 请求头，键值对格式 |

### `apiAuth` Properties

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `apiAuth.authType` | string | No | 认证类型：`none`（默认）、`api_key_query`（URL 参数）、`api_key_header`（请求头） |
| `apiAuth.apiKeyQueryParam` | string | No | URL 查询参数名（authType 为 api_key_query 时使用，如 `key`） |
| `apiAuth.apiKeyValue` | string | No | API Key 值（连接时使用，生成输出中会被替换为占位符） |
| `apiAuth.apiKeyHeader` | string | No | 请求头名称（authType 为 api_key_header 时使用，默认 `Authorization`） |
| `apiAuth.apiKeyHeaderPrefix` | string | No | 请求头前缀：`bearer`、`basic` 或自定义 |

## Example

### 通过 streamable URL 转换

```bash
mcp2skill name=天气查询 streamableUrl=http://192.168.0.21:8081/mcp/server/streamable?key=xxx description="查询天气" output=./skills
```

### 通过 SSE URL 转换

```bash
mcp2skill name=地图服务 sseUrl=http://192.168.0.21:8081/mcp/server/sse transport=sse description="地图服务" output=./skills
```

### 使用 API Key 查询参数认证

```bash
mcp2skill name=天气查询 streamableUrl=http://example.com/mcp/server/streamable description="查询天气" 'apiAuth={"authType":"api_key_query","apiKeyQueryParam":"key","apiKeyValue":"my-secret-key"}'
```

### 使用请求头认证（Bearer Token）

```bash
mcp2skill name=天气查询 streamableUrl=http://example.com/mcp/server/streamable description="查询天气" 'apiAuth={"authType":"api_key_header","apiKeyHeaderPrefix":"bearer","apiKeyValue":"my-token"}'
```

### 使用自定义请求头

```bash
mcp2skill name=天气查询 streamableUrl=http://example.com/mcp/server/streamable description="查询天气" 'headers={"X-Custom-Header":"value"}'
```

### 输出结构

转换成功后，在输出目录下生成：

```
{output}/{skillName}/
├── SKILL.md                        # Skill 入口概览
├── scripts/
│   └── mcp_client.py               # 自动生成的 MCP Python 客户端
└── references/
    └── operations/                 # 每个工具一个详情文件
        └── {tool-name}.md
```
