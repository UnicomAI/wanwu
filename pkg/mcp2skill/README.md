# mcp2skill

将 MCP (Model Context Protocol) Server 的工具列表转换为 Agent Skills 格式的结构化 Markdown 文档，并生成可直接调用的 Python 客户端脚本，供 AI Agent 按需加载工具信息。

<br />

# 原理

### 为什么需要 Skill 格式？

大语言模型的上下文窗口有限。一个包含大量工具的 MCP Server，其工具描述可能占据大量 token，直接灌入会挤占 Agent 的推理空间。Skill 格式采用**渐进式披露**策略：

1. **SKILL.md** — 入口文件，仅包含工具概览和 MCP Server 连接信息（极小）
2. **scripts/mcp_client.py** — 自动生成的 Python 客户端，可直接调用 MCP Server 的工具
3. **references/operations/** — 每个工具一个详情文件，包含参数、输出、示例等完整信息

Agent 只需先读取 SKILL.md 了解全局，再按需加载特定工具的详情，实现**按需读取、精准投递**。

### 转换流程

```
MCP Server (SSE / Streamable HTTP)
        │
        ▼
┌─────────────────┐
│  ListTools API  │ ──▶ []*protocol.Tool
└─────────────────┘
        │
        ▼
┌─────────────┐
│   Parser    │ ──▶ SkillDocument (IR，中间表示)
└─────────────┘
        │
        ▼
┌─────────────┐
│  Renderer   │ ──▶ Markdown + Python 客户端
└─────────────┘
        │
        ▼
┌─────────────┐
│   Writer    │ ──▶ 文件系统
└─────────────┘
```

**连接阶段** 通过 MCP 协议的 `ListTools` API 从 MCP Server 获取工具列表，支持两种传输方式：

- **Streamable HTTP** — 推荐的现代传输方式，默认选项
- **SSE** — Server-Sent Events 传输方式

**Parser** 阶段将 MCP 工具列表解析为 `SkillDocument` IR，包含：

- **Meta** — Skill 名称、描述、工具数量
- **Tools** — 工具列表，每个工具包含名称、描述、参数、输出 Schema、注解
- **ServerInfo** — MCP Server URL 和传输类型

**Renderer** 阶段将 IR 渲染为：

- `SKILL.md` — 概览文档，含工具表和快速开始指引
- `references/operations/*.md` — 每个工具的详情文档
- `scripts/mcp_client.py` — 自动生成的 Python 客户端脚本，包含每个工具的类型化辅助函数

**Writer** 阶段将渲染结果写入文件系统。

### 安全：URL Key 脱敏

当 MCP Server URL 包含认证参数（如 `key`、`token`、`secret`、`apikey`、`api_key`）时，生成的输出会自动将敏感值替换为 `<YOUR_KEY>` 占位符，提醒用户需要填入自己的凭证。

例如：
```
原始 URL: http://example.com/mcp?key=abc123
输出 URL: http://example.com/mcp?key=<YOUR_KEY>
```

### 自动生成的 Python 客户端

`scripts/mcp_client.py` 包含：

1. **类型化辅助函数** — 每个工具生成一个 Python 函数，必填参数作为函数参数，选填参数通过 `**kwargs` 传递
2. **核心调用函数** — `call_tool()` 和 `list_tools()` 处理与 MCP Server 的通信
3. **命令行接口** — 支持 `--list` 列出工具、`--tool` 调用工具、`--arguments` 传递参数

## 生成的目录结构

```
{outputDir}/{skillName}/
├── SKILL.md                                  # Skill 入口概览
├── scripts/
│   └── mcp_client.py                         # 自动生成的 Python MCP 客户端
└── references/
    └── operations/                           # 每个工具一个详情文件
        └── {tool-name}.md
```

`{skillName}` 的来源优先级：

1. `ConvertOptions.SkillName` 显式指定
2. `MCPConfig.Name` 配置指定（经 `toFileName` 处理）
3. 从首个工具名称自动推断（取下划线前的前缀）

## 使用方式

### 1. 通过 JSON 配置转换（推荐）

最常用的方式，通过 JSON 配置文件指定 MCP Server 连接信息：

```go
import "github.com/UnicomAI/wanwu/pkg/mcp2skill"

err := mcp2skill.ConvertFromConfig(ctx, "mcp-config.json", "./skills")
```

配置文件格式（`mcp-config.json`）：

```json
{
  "name": "天气查询",
  "description": "根据地点获取当前的天气情况",
  "streamableUrl": "http://192.168.0.21:8081/mcp/server/streamable?key=xxx",
  "sseUrl": "",
  "transport": "streamable"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 否 | Skill 名称，留空时自动推断 |
| `description` | string | 否 | Skill 描述，留空时自动生成 |
| `streamableUrl` | string | 否 | Streamable HTTP 端点 URL |
| `sseUrl` | string | 否 | SSE 端点 URL |
| `transport` | string | 否 | 传输类型：`"streamable"`（默认）或 `"sse"` |

### 2. 通过 MCPConfig 对象转换

在代码中直接构造配置对象，无需读取文件：

```go
cfg := &mcp2skill.MCPConfig{
    Name:          "search",
    StreamableUrl: "http://localhost:8080/mcp",
    Transport:     "streamable",
}

err := mcp2skill.ConvertFromMCPConfig(ctx, cfg, "./skills")
```

### 3. 通过 Server 配置转换

适用于已有 MCP Server URL 但不想使用 JSON 配置的场景：

```go
err := mcp2skill.ConvertFromServer(ctx, mcp2skill.MCPServerConfig{
    URL:           "http://localhost:8080/mcp",
    TransportType: "streamable",
}, mcp2skill.ConvertOptions{
    OutputDir:   "./skills",
    SkillName:   "my-mcp-tools",
    Description: "My MCP tools",
})
```

### 4. 从已有工具列表转换

已通过其他方式获取了 MCP 工具列表时，跳过连接阶段直接转换：

```go
// tools 是 []*protocol.Tool 类型的工具列表
err := mcp2skill.ConvertFromTools(tools, mcp2skill.ConvertOptions{
    OutputDir:     "./skills",
    SkillName:     "my-tools",
    Description:   "My custom tools",
    ServerURL:     "http://localhost:8080/mcp",
    TransportType: "streamable",
})
```

### 5. 仅解析为 IR（不写文件）

需要程序化使用解析结果时使用，不会产生任何文件 I/O：

```go
doc := mcp2skill.ParseToIR(tools, "my-skill")

fmt.Printf("Skill: %s (%d tools)\n", doc.Meta.Name, doc.Meta.ToolCount)
for _, tool := range doc.Tools {
    fmt.Printf("  %s: %s\n", tool.Name, tool.Description)
    for _, param := range tool.Parameters {
        fmt.Printf("    - %s (%s) required=%v\n", param.Name, param.Type, param.Required)
    }
}
```

### 6. 仅渲染（自定义输出）

获取 Markdown 字符串但不写文件，适合集成到其他系统：

```go
doc := mcp2skill.ParseToIR(tools, "my-skill")
renderer := mcp2skill.NewRenderer()

skillMd  := renderer.RenderSkill(doc)                    // SKILL.md 内容
opMd     := renderer.RenderOperation(doc.Tools[0])       // 工具详情
clientPy := renderer.RenderMCPClient(doc)                // Python 客户端脚本
```

## 核心数据结构

### SkillDocument（IR 顶层）

```
SkillDocument
├── Meta                      # Skill 元信息
│   ├── Name                  # Skill 名称
│   ├── Description           # 描述
│   └── ToolCount             # 工具数量
├── Tools[]                   # 工具列表
│   └── ToolDocument
│       ├── Name              # 工具名称
│       ├── Description       # 工具描述
│       ├── Parameters[]      # 参数列表
│       │   └── ParameterDocument
│       │       ├── Name, Type, Description
│       │       ├── Required  # 是否必填
│       │       ├── Enum[]    # 枚举值
│       │       ├── Default   # 默认值
│       │       ├── Properties[]  # 嵌套对象属性
│       │       └── Items         # 数组元素类型
│       ├── Required[]        # 必填参数名列表
│       ├── OutputSchema      # 输出 Schema
│       └── Annotations       # 工具注解
│           ├── Title
│           ├── ReadOnlyHint
│           ├── DestructiveHint
│           ├── IdempotentHint
│           └── OpenWorldHint
└── ServerInfo                # MCP Server 信息
    ├── URL                   # Server URL（敏感信息已脱敏）
    └── TransportType         # 传输类型（sse / streamable）
```

### MCPConfig（输入配置）

```go
type MCPConfig struct {
    Name          string `json:"name"`           // Skill 名称
    Description   string `json:"description"`    // 描述（可选）
    StreamableUrl string `json:"streamableUrl"`  // Streamable HTTP URL
    SseUrl        string `json:"sseUrl"`         // SSE URL
    Transport     string `json:"transport"`      // "streamable" 或 "sse"
}
```

## 与 openapi2skill 的对比

| <br /> | openapi2skill | mcp2skill |
| ------ | ------------- | --------- |
| 输入来源 | OpenAPI 3.x 规范文件 | MCP Server 工具列表 |
| 输入方式 | 读取本地文件 | 连接远程 MCP Server 或传入工具列表 |
| 输出格式 | Markdown 文档 | Markdown 文档 + Python 客户端脚本 |
| 使用场景 | AI Agent 按需阅读 API 文档 | AI Agent 调用 MCP 工具 |
| 核心能力 | 生成渐进式可导航的 API 文档 | 生成可导航的工具文档 + 可执行的客户端 |
| Schema 处理 | 保留结构，按前缀分组 | 保留结构，嵌套展开 |
| 分组方式 | 按标签/路径分为资源组 | 按工具名称平铺 |
| 认证 | 提取 OpenAPI Security Schemes | URL Key 自动脱敏 |
| 依赖 | kin-openapi | go-mcp |
