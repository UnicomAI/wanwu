# openapi2skill

将 OpenAPI 3.x 规范转换为 Agent Skills 格式的结构化 Markdown 文档，供 AI Agent 按需加载 API 信息，避免大规格文件超出上下文窗口。

<br />

# 原理

### 为什么需要 Skill 格式？

大语言模型的上下文窗口有限。一个包含几十个接口的 OpenAPI 规范文件可能超过数万 token，直接灌入会挤占 Agent 的推理空间。Skill 格式采用**渐进式披露**策略：

1. **SKILL.md** — 入口文件，仅包含 API 概览和资源列表（极小）
2. **resources/** — 每个资源一个索引文件，列出该资源下的所有操作
3. **operations/** — 每个操作一个详情文件，包含参数、请求体、响应等完整信息
4. **schemas/** — 按命名前缀分组，每个 Schema 独立文件

Agent 只需先读取 SKILL.md 了解全局，再按需加载特定资源或操作，实现**按需读取、精准投递**。

### 转换流程

```
OpenAPI Spec (JSON/YAML)
        │
        ▼
┌─────────────┐
│   Parser    │ ──▶ SkillDocument (IR，中间表示)
└─────────────┘
        │
        ▼
┌─────────────┐
│  Renderer   │ ──▶ Markdown 字符串
└─────────────┘
        │
        ▼
┌─────────────┐
│   Writer    │ ──▶ 文件系统
└─────────────┘
```

**Parser** 阶段将 OpenAPI 规范解析为 `SkillDocument` IR，包含：

- **Meta** — API 名称、版本、服务器地址、认证方案
- **Resources** — 按标签或路径前缀分组的操作集合
- **SchemaGroups** — 按命名前缀分组的 Schema 定义
- **AuthSchemes** — 认证方案详情

**Renderer** 阶段将 IR 渲染为 Markdown 字符串，支持两种 Schema 引用方式：

- `$ref` 引用 → 生成指向 `schemas/` 目录的链接
- 内联 Schema → 直接在操作文档中渲染字段表，同时提取到 `schemas/` 目录

**Writer** 阶段将 Markdown 写入文件系统的目录结构。

### Schema 提取策略

Parser 仅从 `components/schemas` 提取命名 Schema，按 PascalCase 前缀分组（如 `SearchResponse` → `Search` 组），生成独立的 Schema 文件。

对于操作 requestBody/response 中的 Schema，根据类型分别处理：

- **`$ref`** **引用**（如 `"$ref": "#/components/schemas/SearchResponse"`）→ 渲染为指向 schema 文件的链接（如 `[SearchResponse](../schemas/Search/SearchResponse.md)`）
- **内联 Schema**（直接在 requestBody/response 中定义的 object）→ 在操作文档中**直接展示字段表**，不在 `schemas/` 目录下生成单独文件，避免重复

这样对于没有 `components/schemas` 的 OpenAPI 规范（如通义千问），所有字段信息内联在操作文档中即可完整呈现；对于有命名 Schema 的规范（如博查搜索），命名 Schema 独立成文件，内联字段直接展示。

### 分组策略

| 策略  | 常量            | 说明                                     |
| --- | ------------- | -------------------------------------- |
| 按标签 | `GroupByTags` | 按 OpenAPI `tags` 分组，无 tag 归入 `default` |
| 按路径 | `GroupByPath` | 按 URL 首段分组（自动去除 `/v1/` 等版本前缀）          |
| 自动  | `GroupByAuto` | 优先用 tags，无 tags 时回退到 path（默认）          |

### 过滤选项

```go
Filter: &ParserFilter{
    IncludeTags:       []string{"pet", "store"},  // 只包含这些 tag
    ExcludeTags:       []string{"internal"},       // 排除这些 tag
    ExcludeDeprecated: true,                       // 排除已废弃的接口
    ExcludePaths:      []string{"/admin/", "/internal/"}, // 排除路径前缀
}
```

## 生成的目录结构

```
{outputDir}/{skillName}/
├── SKILL.md                                  # API 入口概览
└── references/
    ├── resources/                            # 按资源分组的操作索引
    │   └── {resource-name}.md
    ├── operations/                           # 每个操作一个详情文件
    │   └── {operation-id}.md
    ├── schemas/                              # 按命名前缀分组的 Schema
    │   ├── {Prefix}/
    │   │   ├── _index.md                     # 组索引
    │   │   └── {SchemaName}.md              # 单个 Schema 详情
    │   └── ...
    └── authentication.md                     # 认证方案（如有）
```

`{skillName}` 默认从 `info.title` 生成（经 `toFileName` 处理并转小写），可通过 `ParserOptions.SkillName` 覆盖。

## 使用方式

### 1. 完整转换（解析 + 写文件）

最常用的方式，从 OpenAPI JSON/YAML 字节直接生成 Skill 文件：

```go
import (
    "context"
    "os"
    "github.com/UnicomAI/wanwu/pkg/openapi2skill"
)

specData, _ := os.ReadFile("openapi.json")

err := openapi2skill.Convert(context.Background(), specData, openapi2skill.ConvertOptions{
    OutputDir: "./skills",
    Parser: openapi2skill.ParserOptions{
        SkillName: "my-api",                     // 可选，默认从 info.title 生成
        GroupBy:   openapi2skill.GroupByAuto,     // 可选，默认 auto
        Filter: &openapi2skill.ParserFilter{      // 可选，过滤接口
            ExcludeDeprecated: true,
            ExcludeTags:       []string{"internal"},
            ExcludePaths:      []string{"/admin/"},
        },
    },
    CaseStrategy: openapi2skill.CaseStrategyLowercase, // 可选，小写化文件名
})
```

输出到 `./skills/my-api/SKILL.md` 等文件。

### 2. 仅解析为 IR（不写文件）

需要程序化使用解析结果时使用，不会产生任何文件 I/O：

```go
doc, err := openapi2skill.ParseToIR(context.Background(), specData, openapi2skill.ParserOptions{
    GroupBy: openapi2skill.GroupByTags,
})
if err != nil {
    log.Fatal(err)
}

// 遍历资源
for _, res := range doc.Resources {
    fmt.Printf("Resource: %s (%d ops)\n", res.Tag, len(res.Operations))
    for _, op := range res.Operations {
        fmt.Printf("  %s %s - %s\n", op.Method, op.Path, op.Summary)
    }
}

// 遍历 Schema
for _, sg := range doc.SchemaGroups {
    fmt.Printf("SchemaGroup: %s\n", sg.Prefix)
    for _, s := range sg.Schemas {
        fmt.Printf("  %s (%s)\n", s.Name, s.Type)
    }
}
```

### 3. 已有 `*openapi3.T` 对象

项目中如果已经通过 `kin-openapi` 加载了文档，可以直接使用，避免重复解析：

```go
import "github.com/getkin/kin-openapi/openapi3"

doc, _ := openapi3.NewLoader().LoadFromData(specData)

// 完整转换
err := openapi2skill.ConvertDoc(doc, openapi2skill.ConvertOptions{
    OutputDir: "./skills",
})

// 或仅解析
ir := openapi2skill.ParseDocToIR(doc, openapi2skill.ParserOptions{
    GroupBy: openapi2skill.GroupByAuto,
})
```

### 4. 仅渲染（自定义输出）

获取 Markdown 字符串但不写文件，适合集成到其他系统：

```go
doc := openapi2skill.ParseDocToIR(openapiDoc, opts)
renderer := openapi2skill.NewRenderer()

skillMd     := renderer.RenderSkill(doc)                          // SKILL.md 内容
resourceMd  := renderer.RenderResource(doc.Resources[0])         // 资源索引
operationMd := renderer.RenderOperation(doc.Resources[0].Operations[0]) // 操作详情
schemaMd    := renderer.RenderSchema(doc.SchemaGroups[0].Schemas[0])   // Schema 详情
authMd      := renderer.RenderAuthentication(doc.AuthSchemes)         // 认证信息
```

### 5. 大小写策略

在某些不区分大小写的文件系统上可能出现文件名冲突，可以启用 lowercase 策略：

```go
err := openapi2skill.Convert(ctx, specData, openapi2skill.ConvertOptions{
    OutputDir:    "./skills",
    CaseStrategy: openapi2skill.CaseStrategyLowercase,
})
```

此策略会：

- 将 Schema 组前缀转为小写并合并（如 `Pet` 和 `pet` 合并）
- 将 Schema 名称转为小写
- 如有冲突则添加数字后缀（如 `pet-2`）

## ConvertOptions 配置

```go
type ConvertOptions struct {
    OutputDir    string        // 输出目录
    Parser       ParserOptions // 解析选项
    CaseStrategy CaseStrategy  // 大小写策略
}

type ParserOptions struct {
    SkillName string         // 覆盖自动生成的 Skill 名称
    Filter    *ParserFilter  // 过滤选项
    GroupBy   GroupByStrategy // 分组策略（默认 auto）
}
```

## 核心数据结构

### SkillDocument（IR 顶层）

```
SkillDocument
├── Meta                    # API 元信息（名称、版本、服务器、认证）
├── Resources[]             # 资源列表
│   └── ResourceDocument
│       ├── Tag             # 资源名称/标签
│       ├── Description     # 资源描述
│       └── Operations[]    # 操作列表
│           └── OperationDocument
│               ├── OperationID, Path, Method
│               ├── Parameters[]     # 参数列表
│               ├── RequestBody      # 请求体（含 Schema）
│               ├── Responses[]      # 响应列表（含 Schema）
│               └── Security[]       # 安全要求
├── SchemaGroups[]          # Schema 分组
│   └── SchemaGroupDocument
│       ├── Prefix          # 分组前缀（如 Search、Pet）
│       └── Schemas[]       # Schema 列表
│           └── SchemaDocument
│               ├── Name, Type, Description
│               ├── Fields[]         # object 类型的字段
│               ├── EnumValues[]     # enum 类型的值
│               ├── Composition[]    # allOf/oneOf/anyOf
│               └── Items            # array 类型的元素
└── AuthSchemes[]           # 认证方案
    └── AuthSchemeDocument
        ├── Type            # apiKey / http / oauth2 / openIdConnect
        └── 类型特定字段
```

### SchemaRefDocument（Schema 引用）

`SchemaRefDocument` 表示对 Schema 的引用，支持两种模式：

- **`Ref`** **非空** — 引用 `components/schemas` 中的命名 Schema，渲染为链接
- **`Inline`** **非空** — 内联 Schema 定义，渲染为字段表

```go
// $ref 引用
SchemaRefDocument{Ref: "SearchResponse"}
// 渲染结果: [SearchResponse](../schemas/Search/SearchResponse.md)

// 内联 Schema
SchemaRefDocument{Inline: &SchemaDocument{Name: "(inline)", Type: SchemaTypeObject, Fields: [...]}}
// 渲染结果: 直接在文档中展示字段表
```

## 与 mcp2skill 的对比

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

