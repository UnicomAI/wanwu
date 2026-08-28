# WGA Sandbox

沙箱容器交互包，支持在隔离环境中执行智能体任务。

## 架构

```
api.go
  ├── Run(ctx, opts...)      执行任务，返回 <-chan string
  └── Cleanup(ctx, runID)    清理沙箱

api_opencode.go
  ├── ParseOpencodeEvent(data)            → *OpencodeEvent
  ├── ParseOpencodeTextPart(data)         → *TextPart
  ├── ParseOpencodeToolPart(data)         → *ToolPart
  ├── ParseOpencodeReasoningPart(data)    → *ReasoningPart
  ├── ParseOpencodeStepStartPart(data)    → *StepStartPart
  ├── ParseOpencodeStepFinishPart(data)   → *StepFinishPart
  ├── ParseOpencodeFilePart(data)         → *FilePart
  ├── ParseOpencodeSnapshotPart(data)     → *SnapshotPart
  ├── ParseOpencodeAgentPart(data)        → *AgentPart
  ├── ParseOpencodePartPatchPart(data)    → *PartPatchPart
  ├── ParseOpencodePartRetryPart(data)    → *PartRetryPart
  └── ParseOpencodeErrorPart(data)        → *ErrorPart

wga-sandbox-converter/
  ├── eino_converter.go
  │   └── EinoConverter 接口
  ├── eino_iterator.go
  │   ├── ConvertToEinoIterator()       JSON 流 → AgentEvent 迭代器
  │   └── ConvertToEinoIteratorWithError() 错误 → AgentEvent 迭代器
  └── opencode.go
      └── opencodeConverter 实现

sandbox.Manager
  ├── Create(ctx, runID, cfg)  创建沙箱
  ├── Get(runID)               获取实例
  └── Cleanup(ctx, runID)      清理沙箱

runner.Runner
  ├── BeforeRun(ctx)  准备环境
  ├── Run(ctx)        执行任务
  └── AfterRun(ctx)   复制输出
```

## 沙箱模式

| 模式 | 说明 | 状态 |
|------|------|------|
| reuse | 复用已启动容器，通过 Host 指定容器地址 | 完整实现 |
| oneshot | 每次启动新容器，通过 ImageName 指定镜像 | 接口定义 |

## 使用

```go
ctx := context.Background()

runSession, outputCh, _ := wga_sandbox.Run(ctx,
    // 模型配置（必须）
    wga_sandbox_option.WithModelConfig(wga_sandbox_option.ModelConfig{
        Provider:     "yuanjing",
        ProviderName: "YuanJing",
        BaseURL:      "https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1",
        APIKey:       "sk-xxx",
        Model:        "glm-5",
        ModelName:    "GLM-5",
    }),
    // 沙箱配置（必须）
    wga_sandbox_option.WithSandbox(
        wga_sandbox_option.SandboxReuse("localhost"),  // 复用模式
        // 或 wga_sandbox_option.SandboxOneshot("image:tag"),  // 一次性模式
    ),
    // 运行器类型（必须）
    wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
    // 消息列表（必须，最后一条必须是 User 消息）
    wga_sandbox_option.WithMessages([]adk.Message{
        &schema.Message{Role: schema.User, Content: "生成一个 HTTP 服务器"},
    }),
    // 会话标识
    wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
        ThreadID: "thread-123",
        RunID:    "run-456",
    }),
    // MCP 服务器
    wga_sandbox_option.WithMCPs([]wga_sandbox_option.MCP{
        {Name: "Jira工单", URL: "https://jira.example.com/sse"},
    }),
)

for line := range outputCh {
    event, _ := wga_sandbox.ParseOpencodeEvent([]byte(line))
    switch event.Type {
    case wga_sandbox.OpencodeEventTypeText:
        part, _ := wga_sandbox.ParseOpencodeTextPart(event.Part)
        fmt.Println(part.Text)
    case wga_sandbox.OpencodeEventTypeToolUse:
        part, _ := wga_sandbox.ParseOpencodeToolPart(event.Part)
        fmt.Printf("Tool: %s, Status: %s\n", part.Tool, part.State.Status)
    }
}
```

## Eino 集成

```go
import (
    wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
    "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-converter"
    wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/schema"
)

runSession, outputCh, _ := wga_sandbox.Run(ctx,
    wga_sandbox_option.WithModelConfig(modelConfig),
    wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse("localhost")),
    wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
    wga_sandbox_option.WithMessages([]adk.Message{
        &schema.Message{Role: schema.User, Content: "任务描述"},
    }),
    wga_sandbox_option.WithMCPs([]wga_sandbox_option.MCP{
        {Name: "Jira工单", URL: "https://jira.example.com/sse"},
    }),
)

// 转换为 eino AgentEvent 迭代器
iter := wga_sandbox_converter.ConvertToEinoIterator(ctx, wga_sandbox_option.RunnerTypeOpencode, outputCh)
for {
    event, ok := iter.Next()
    if !ok {
        break
    }
    if event.Err != nil {
        // 处理错误
    }
    if event.Output != nil && event.Output.MessageOutput != nil {
        fmt.Println(event.Output.MessageOutput.Message.Content)
    }
}
```

## AG-UI 协议

```go
import ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"

runSession, outputCh, _ := wga_sandbox.Run(ctx,
    wga_sandbox_option.WithModelConfig(modelConfig),
    wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse("localhost")),
    wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
    wga_sandbox_option.WithMessages([]adk.Message{
        &schema.Message{Role: schema.User, Content: "任务描述"},
    }),
    wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
        ThreadID: "thread-123",
        RunID:    "run-456",
    }),
)

tr := ag_ui_util.NewOpencodeTranslator("thread-123", "run-456")
eventCh := tr.TranslateStream(ctx, outputCh)
```

## API

| 函数 | 说明 |
|------|------|
| `Run(ctx, opts...)` | 执行任务，返回 `<-chan string` JSON 字符串流 |
| `Cleanup(ctx, runID)` | 清理沙箱环境 |
| `ReplyQuestion(ctx, cfg, questionID, answers)` | 回答问题（Human-in-the-Loop，仅 Reuse 模式） |
| `RejectQuestion(ctx, cfg, questionID)` | 拒绝问题（Human-in-the-Loop，仅 Reuse 模式） |

### 事件解析

| 函数 | 说明 |
|------|------|
| `ParseOpencodeEvent(data)` | 解析事件，返回 `*OpencodeEvent` |
| `ParseOpencodeTextPart(data)` | 解析文本部分 |
| `ParseOpencodeToolPart(data)` | 解析工具调用部分 |
| `ParseOpencodeReasoningPart(data)` | 解析推理部分 |
| `ParseOpencodeFilePart(data)` | 解析文件部分 |
| `ParseOpencodeSnapshotPart(data)` | 解析快照部分 |
| `ParseOpencodeAgentPart(data)` | 解析智能体部分 |
| `ParseOpencodeQuestionPart(data)` | 解析问题部分（Human-in-the-Loop） |

### Eino 转换 (wga-sandbox-converter)

| 函数 | 说明 |
|------|------|
| `NewEinoConverter(runnerType)` | 创建转换器 |
| `ConvertToEinoIterator(ctx, runnerType, outputCh)` | JSON 流 → `*adk.AsyncIterator[*adk.AgentEvent]` |
| `ConvertToEinoIteratorWithError(ctx, runnerType, err)` | 错误 → `*adk.AsyncIterator[*adk.AgentEvent]` |

## 选项

| 选项 | 说明 | 必须 |
|------|------|------|
| `WithModelConfig` | 模型配置 | 是 |
| `WithSandbox` | 沙箱配置（`SandboxReuse(host)` 或 `SandboxOneshot(imageName)`） | 是 |
| `WithRunnerType` | 运行器类型（`RunnerTypeOpencode`） | 是 |
| `WithMessages` | 消息列表（历史消息 + 当前问题，最后一条必须是 User 消息） | 是 |
| `WithRunSession` | 会话标识 | 否 |
| `WithInstruction` | 系统提示词 | 否 |
| `WithOverallTask` | 整体任务（用于子智能体） | 否 |
| `WithInputDir` | 输入目录 | 否 |
| `WithOutputDir` | 输出目录 | 否 |
| `WithTools` | 工具列表 | 否 |
| `WithSkills` | 技能列表 | 否 |
| `WithMCPs` | MCP 服务器列表 | 否 |
| `WithEnableThinking` | 思考模式 | 否 |
| `WithEnableHumanInTheLoop` | 启用人机交互（可选参数：enableCustom 设置是否允许用户自定义回答） | 否 |
| `WithSkipCleanup` | 跳过清理 | 否 |
| `WithAgentName` | 智能体名称 | 否 |

## 类型

### Tool

工具配置。

```go
type Tool struct {
    OpenAPI3Schema *openapi3.T       // OpenAPI 3.0 schema 文档（必须）
    OperationIDs   []string          // 允许的 operations，为空则全部允许
    APIAuth        *openapi3_util.Auth // API 认证（可选）
    Name           string            // 工具名称，从 schema 的 info.title 自动读取
}
```

### Skill

技能配置。

```go
type Skill struct {
    Dir string // skill 目录路径
}
```

### MCP

MCP 服务器配置。

```go
type MCP struct {
    Name string // MCP 名称
    URL  string // MCP SSE/STREAMABLE 服务器地址
}
```

## MCP 服务器

`WithMCPs` 用于配置 MCP (Model Context Protocol) 服务器，允许智能体通过 SSE 协议与外部工具交互。

**字段验证**：
- `Name`：必须非空
- `URL`：必须非空

```go
wga_sandbox_option.WithMCPs([]wga_sandbox_option.MCP{
    {Name: "Jira工单", URL: "https://jira.example.com/sse"},
    {Name: "Confluence", URL: "https://confluence.example.com/sse"},
})
```

生成的 opencode.json 配置：

```json
{
  "mcp": {
    "Jira工单": {
      "type": "remote",
      "url": "https://jira.example.com/sse",
      "enabled": true
    },
    "Confluence": {
      "type": "remote",
      "url": "https://confluence.example.com/sse",
      "enabled": true
    }
  }
}
```

## 依赖

- Sandbox API 服务：通过 `SandboxConfig.Host()` 动态获取端点地址
- Opencode 服务：通过 `SandboxConfig.OpencodeEndpoint()` 动态获取 HTTP API 地址

## 事件类型

| 类型 | 说明 |
|------|------|
| `OpencodeEventTypeStepStart` | 步骤开始 |
| `OpencodeEventTypeStepFinish` | 步骤结束 |
| `OpencodeEventTypeText` | 文本输出 |
| `OpencodeEventTypeToolUse` | 工具调用 |
| `OpencodeEventTypeReasoning` | 推理过程 |
| `OpencodeEventTypeFile` | 文件操作 |
| `OpencodeEventTypeSnapshot` | 快照 |
| `OpencodeEventTypeAgent` | 智能体 |
| `OpencodeEventTypePatch` | 补丁 |
| `OpencodeEventTypeRetry` | 重试 |
| `OpencodeEventTypeSubtask` | 子任务 |
| `OpencodeEventTypeCompaction` | 压缩 |
| `OpencodeEventTypeError` | 错误 |
| `OpencodeEventTypeQuestionAsked` | 问题提出（Human-in-the-Loop） |
| `OpencodeEventTypeQuestionReplied` | 问题已回答（Human-in-the-Loop） |
| `OpencodeEventTypeQuestionRejected` | 问题被拒绝（Human-in-the-Loop） |

## Human-in-the-Loop

wga-sandbox 支持人机交互（Human-in-the-Loop），允许 AI 在执行过程中向用户提问并等待回复。

### 功能说明

当启用 HITL 后，AI 可以通过 `question.ask()` 向用户提问，系统会：
1. 发布 `question.asked` SSE 事件，前端显示问题 UI
2. AI 阻塞等待用户回复
3. 用户通过 HTTP API 提交回答或取消
4. AI 收到回复后继续执行

### 适用模式

| 模式 | HITL 支持 | 说明 |
|------|----------|------|
| **Reuse** | ✅ 支持 | 单实例 Sandbox，questionID 在 pending Map 中唯一 |
| **Oneshot** | ❌ 不支持 | 需要额外映射 questionID → runID → SandboxHost |

### 启用方式

```go
runSession, outputCh, _ := wga_sandbox.Run(ctx,
    wga_sandbox_option.WithModelConfig(modelConfig),
    wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse("localhost")),
    wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
    wga_sandbox_option.WithMessages(messages),
    wga_sandbox_option.WithEnableHumanInTheLoop(true),  // 启用人机交互
)
```

### 配置生成

启用 HITL 后，opencode.json 中会生成以下配置：

```json
{
  "permission": {
    "*": "allow",
    "question": "ask"
  }
}
```

未启用时，`permission.question` 默认为 `deny`，AI 无法调用 question 工具。

### 回答/取消问题

```go
// 回答问题
err := wga_sandbox.ReplyQuestion(ctx, sandboxCfg, runID, questionID, answers)

// 取消问题
err := wga_sandbox.RejectQuestion(ctx, sandboxCfg, runID, questionID)
```

**参数说明**：
- `ctx`: 上下文
- `sandboxCfg`: Sandbox 配置，需使用 Reuse 模式
- `runID`: 运行 ID，从 SSE 事件 `question.asked` 的 `runId` 字段获取
- `questionID`: 问题 ID，从 SSE 事件 `question.asked` 的 `questionId` 字段获取
- `answers`: 答案数组，格式为 `[][]string`

### 事件格式

当 AI 提问时，会发送 `question.asked` 事件：

```json
{
  "type": "question.asked",
  "timestamp": 1234567890123,
  "sessionID": "session-xxx",
  "part": {
    "type": "question",
    "questionId": "question_01JXYZ...",
    "sessionID": "session-xxx",
    "status": "pending",
    "questions": [{
      "question": "请选择部署环境",
      "header": "环境选择",
      "options": [
        {"label": "开发环境", "description": "用于开发和测试"},
        {"label": "生产环境", "description": "正式生产部署"}
      ],
      "multiple": false,
      "custom": false
    }]
  }
}
```

用户回答后，会发送 `question.replied` 事件：

```json
{
  "type": "question.replied",
  "timestamp": 1234567890123,
  "sessionID": "session-xxx",
  "part": {
    "type": "question",
    "questionId": "question_01JXYZ...",
    "sessionID": "session-xxx",
    "status": "answered",
    "questions": [],
    "answers": [["开发环境"]]
  }
}
```

**事件类型**：
- `question.asked`: 问题待回答
- `question.replied`: 用户已回答，`answers` 字段包含用户选择
- `question.rejected`: 用户取消

### API

| 函数 | 说明 |
|------|------|
| `ReplyQuestion(ctx, cfg, runID, questionID, answers)` | 回答问题（仅 Reuse 模式） |
| `RejectQuestion(ctx, cfg, runID, questionID)` | 拒绝问题（仅 Reuse 模式） |
