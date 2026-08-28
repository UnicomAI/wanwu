# AG-UI Util

AG-UI 协议事件转换，将不同来源的事件流转换为 AG-UI 格式。

## 包结构

```
pkg/ag-ui-util/
├── message_types.go      # 常量和类型定义
├── message_state.go      # 消息状态管理（TEXT_MESSAGE/REASONING 状态）
├── event_state.go        # 基础状态管理（RUN_STARTED/FINISHED）
├── translator_eino.go    # EinoTranslator（eino AgentEvent 转换）
├── translator_opencode.go # OpencodeTranslator（opencode JSON 转换）
├── stream_processor.go   # 事件流处理（清洗 + 聚合）
├── README.md
└── EVENT_HISTORY.md
```

## AG-UI 协议规范

### 核心概念

AG-UI 采用**事件流驱动**的架构，通过事件流逐步构建消息对象。

#### 消息（Message）

消息是最终呈现给用户的内容单元，通过事件流逐步构建：

| 消息类型 | role 字段 | 说明 |
|---------|----------|------|
| AssistantMessage | `assistant` | AI 回复，可包含 content 和 toolCalls |
| UserMessage | `user` | 用户输入 |
| ToolMessage | `tool` | 工具执行结果 |
| ReasoningMessage | `reasoning` | AI 思考过程 |
| SystemMessage | `system` | 系统指令 |
| DeveloperMessage | `developer` | 开发者消息 |

消息结构：
```typescript
// 基础消息接口
interface Message {
  id: string;              // 消息唯一标识（messageId）
  role: string;            // 消息角色，决定消息类型
  content?: string;        // 文本内容
}

// UserMessage - 用户输入
interface UserMessage extends Message {
  role: "user";
  content: string | InputContent[];  // 纯文本或多模态内容（文本+文件）
}

// AssistantMessage - AI 回复
interface AssistantMessage extends Message {
  role: "assistant";
  content?: string;
  toolCalls?: ToolCall[];  // 可包含工具调用
}

// ToolMessage - 工具执行结果
interface ToolMessage extends Message {
  role: "tool";
  toolCallId: string;      // 关联到对应的工具调用
  content: string;         // 工具返回内容
}

// ReasoningMessage - AI 思考过程
interface ReasoningMessage extends Message {
  role: "reasoning";
  content: string;
}

// ToolCall - 工具调用（嵌入在 AssistantMessage 中）
interface ToolCall {
  id: string;              // toolCallId
  type: "function";
  function: {
    name: string;          // 工具名称
    arguments: string;     // JSON 格式的参数
  };
}
```

#### 用户上传文件

UserMessage 的 `content` 可以是多模态内容数组，支持文本和文件：

```typescript
// InputContent 联合类型
type InputContent = TextInputContent | BinaryInputContent;

// 文本内容
interface TextInputContent {
  type: "text";
  text: string;
}

// 二进制内容（文件/图片）
interface BinaryInputContent {
  type: "binary";
  mimeType: string;        // 如 "image/png", "application/pdf"
  id?: string;             // 文件引用 ID
  url?: string;            // 文件 URL
  data?: string;           // Base64 编码数据
  filename?: string;       // 文件名
}
```

**必须提供 `id`、`url` 或 `data` 中至少一个。**

示例：
```json
// 用户上传图片
{
  "id": "msg-123",
  "role": "user",
  "content": [
    { "type": "text", "text": "请分析这张图片" },
    { 
      "type": "binary", 
      "mimeType": "image/png", 
      "url": "https://example.com/image.png" 
    }
  ]
}

// 用户上传 PDF 文件（Base64）
{
  "id": "msg-456",
  "role": "user",
  "content": [
    { "type": "text", "text": "请总结这个文档" },
    { 
      "type": "binary", 
      "mimeType": "application/pdf",
      "filename": "report.pdf",
      "data": "JVBERi0xLjQK..."
    }
  ]
}
```

#### 事件（Event）

事件是流式传输的基本单元，用于增量更新消息状态：

| 事件类型 | 说明 | 关键字段 |
|---------|------|---------|
| `TEXT_MESSAGE_START` | 开始文本消息 | `messageId`, `role` |
| `TEXT_MESSAGE_CONTENT` | 追加文本内容 | `messageId`, `delta` |
| `TEXT_MESSAGE_END` | 结束文本消息 | `messageId` |
| `TOOL_CALL_START` | 开始工具调用 | `toolCallId`, `toolCallName`, `parentMessageId` |
| `TOOL_CALL_ARGS` | 追加工具参数 | `toolCallId`, `delta` |
| `TOOL_CALL_END` | 结束工具调用 | `toolCallId` |
| `TOOL_CALL_RESULT` | 工具执行结果 | `messageId`, `toolCallId`, `content` |
| `REASONING_START` | 开始推理过程 | `messageId` |
| `REASONING_MESSAGE_START` | 开始推理消息 | `messageId`, `role: "reasoning"` |
| `REASONING_MESSAGE_CONTENT` | 追加推理内容 | `messageId`, `delta` |
| `REASONING_MESSAGE_END` | 结束推理消息 | `messageId` |
| `REASONING_END` | 结束推理过程 | `messageId` |
| `RUN_STARTED` / `RUN_FINISHED` | 运行生命周期 | `threadId`, `runId` |
| `ACTIVITY_SNAPSHOT` | 智能体活动快照 | `messageId`, `activityType`, `content` |

#### 关键字段说明

- **messageId**: 消息唯一标识，用于关联同一消息的所有事件
- **toolCallId**: 工具调用唯一标识，用于关联工具调用的 START/ARGS/END/RESULT
- **parentMessageId**: 工具调用所属的消息ID，用于将工具调用关联到 AssistantMessage
- **role**: 消息角色，决定消息类型（assistant/user/tool/reasoning 等）
- **delta**: 增量内容，用于流式追加文本

### 事件与消息的关系

事件通过 `messageId` 关联，最终聚合为消息对象。

#### 文本消息示例

```
事件流:
TEXT_MESSAGE_START (messageId: "msg-1", role: "assistant")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: "Hello")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: " world")
TEXT_MESSAGE_END (messageId: "msg-1")

↓ 聚合为消息:

AssistantMessage {
  id: "msg-1",
  role: "assistant",
  content: "Hello world"
}
```

#### 工具调用关联示例

`parentMessageId` 用于将工具调用关联到已有的 AssistantMessage：

```
事件流:
TEXT_MESSAGE_START (messageId: "msg-1", role: "assistant")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: "Let me search")
TEXT_MESSAGE_END (messageId: "msg-1")

TOOL_CALL_START (toolCallId: "call-1", toolCallName: "search", parentMessageId: "msg-1")
TOOL_CALL_ARGS (toolCallId: "call-1", delta: "{\"query\": \"test\"}")
TOOL_CALL_END (toolCallId: "call-1")

TOOL_CALL_RESULT (messageId: "result-1", toolCallId: "call-1", content: "found 3 items")

↓ 聚合为消息:

AssistantMessage {
  id: "msg-1",
  role: "assistant",
  content: "Let me search",
  toolCalls: [{ id: "call-1", function: { name: "search", arguments: "{\"query\": \"test\"}" } }]
}

ToolMessage {
  id: "result-1",
  role: "tool",
  toolCallId: "call-1",
  content: "found 3 items"
}
```

### 事件类型

| 事件类型 | 事件序列 | 说明 |
|---------|---------|------|
| Run | `RUN_STARTED` → ... → `RUN_FINISHED` | 一次完整的 AI 运行 |
| TextMessage | `TEXT_MESSAGE_START` → `TEXT_MESSAGE_CONTENT*` → `TEXT_MESSAGE_END` | 文本消息 |
| Reasoning | `REASONING_START` → `REASONING_MESSAGE_START` → `REASONING_MESSAGE_CONTENT*` → `REASONING_MESSAGE_END` → `REASONING_END` | 推理过程 |
| ToolCall (发起) | `TOOL_CALL_START` → `TOOL_CALL_ARGS*` → `TOOL_CALL_END` | Assistant 发起工具调用 |
| ToolCall (结果) | `TOOL_CALL_RESULT` | 工具执行返回结果（独立事件） |

> **说明**：`TOOL_CALL_START/ARGS/END` 和 `TOOL_CALL_RESULT` 是两个独立的事件序列。前者由 Assistant 发起调用，后者由工具执行返回结果。它们通过 `toolCallId` 关联。

## 本实现规则（AG-UI 规范子集）

本实现采用**串行处理模式**，不穿插事件。

### 活跃状态说明

"活跃"指一个事件序列已经发送了 START 事件但尚未发送 END 事件：

| 状态 | 活跃条件 | AG-UI 完整规范 | 本实现 |
|-----|---------|---------------|-------|
| TEXT_MESSAGE 活跃 | 存在 messageId，已发送 `TEXT_MESSAGE_START` 但未发送 `TEXT_MESSAGE_END` | 多个可同时活跃（不同 messageId） | 最多 1 个活跃（单个 messageId） |
| TOOL_CALL 活跃 | 存在 toolCallId，已发送 `TOOL_CALL_START` 但未发送 `TOOL_CALL_END` | 多个可同时活跃（不同 toolCallId） | 最多 1 个活跃（串行处理） |
| REASONING 活跃 | 已发送 `REASONING_START` 但未发送 `REASONING_END` | 最多 1 个活跃 | 最多 1 个活跃 ✅ |

### 事件发送顺序

收到不同类型内容时，按以下顺序发送事件：

#### Tool 消息（Role=Tool）

收到工具执行结果时，直接发送 `TOOL_CALL_RESULT`（独立事件，不需要 START/END）：

```
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
TOOL_CALL_RESULT
```

#### ToolCalls（Assistant 消息中的工具调用）

收到工具调用请求时：

```
parentMsgID = 当前 messageId              ← 保存当前消息 ID
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
TOOL_CALL_START (parentMessageId: parentMsgID)
TOOL_CALL_ARGS
TOOL_CALL_END                    ← 每个 ToolCall 完整处理
TOOL_CALL_START (parentMessageId: parentMsgID)
TOOL_CALL_ARGS
TOOL_CALL_END
...
```

#### ReasoningContent

收到推理内容时：

```
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
REASONING_START (如果 REASONING 未活跃)
REASONING_MESSAGE_START (如果 REASONING_MESSAGE 未活跃)
REASONING_MESSAGE_CONTENT
```

#### Content

收到文本内容时：

```
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_START (如果 TEXT_MESSAGE 未活跃)
TEXT_MESSAGE_CONTENT
```

#### Question（Human-in-the-Loop）

收到 question 事件时，发送 `ACTIVITY_SNAPSHOT`（独立事件，不需要 START/END）：

```
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
ACTIVITY_SNAPSHOT (activityType: "question")
```

> **说明**：question 作为独立的 activity，必须先结束所有活跃消息再发送，确保不穿插事件序列。

## 转换器

| 转换器 | 输入 | 使用场景 |
|--------|------|---------|
| `EinoTranslator` | eino AgentEvent | wga.Run() 输出转换，支持多智能体切换 |
| `OpencodeTranslator` | opencode JSON 字符串 | wga-sandbox 输出转换 |

### EinoTranslator

支持多智能体场景，通过 `ACTIVITY_SNAPSHOT` 事件标识当前运行的智能体。

**功能特性：**
- Agent 切换检测：自动检测 `AgentEvent.AgentName` 变化
- Activity 事件：发送 `ACTIVITY_SNAPSHOT` 标识智能体启动/结束
- 独立消息状态：每个 Agent 维护独立的 MessageState
- 流式工具调用：支持流式传输工具调用参数

**ActivitySnapshot 结构：**
```json
{
  "type": "ACTIVITY_SNAPSHOT",
  "messageId": "step-xxx",
  "activityType": "sub_agent",
  "content": {
    "agentName": "Plan Agent",
    "instanceNum": 1,
    "status": "started"
  }
}
```

**字段说明：**

| 字段 | 说明 |
|-----|------|
| `activityType` | 固定为 `"sub_agent"` |
| `content.agentName` | 智能体名称 |
| `content.instanceNum` | 智能体实例编号（同一智能体可能多次运行） |
| `content.status` | 状态：`"started"` 或 `"finished"` |

### OpencodeTranslator

将 opencode JSON 事件流转换为 AG-UI 事件。

**事件类型映射：**

| opencode 事件 | AG-UI 事件 |
|--------------|-----------|
| `text` | `TEXT_MESSAGE_CONTENT` |
| `reasoning` | `REASONING_MESSAGE_CONTENT` |
| `tool_use` | `TOOL_CALL_START/ARGS/END/RESULT` |
| `error` | `TEXT_MESSAGE_CONTENT`（带 `[error]` 前缀） |

## 使用示例

### EinoTranslator - 多智能体场景

```go
import ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"

// 创建转换器
runSession, iter, _ := wga.Run(ctx, agentID, opts...)
tr := ag_ui_util.NewEinoTranslator(runSession.ThreadID, runSession.RunID)
eventCh := tr.TranslateStream(ctx, iter)

// 转换为 JSON 字符串流（用于 SSE）
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)
```

### OpencodeTranslator - Sandbox 场景

```go
// 创建转换器
runSession, outputCh, _ := wga_sandbox.Run(ctx, opts...)
tr := ag_ui_util.NewOpencodeTranslator(runSession.ThreadID, runSession.RunID)
eventCh := tr.TranslateStream(ctx, outputCh)

// 转换为 JSON 字符串流
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)
```

### StreamProcessor - 事件清洗和历史聚合

```go
// 创建处理器配置
config := &ag_ui_util.ProcessorConfig{
    ToolNameMapper: map[string]string{
        "transfer_to_agent": "正在交给专业智能体",
    },
    ExcludedAgentNames: []string{"default", "Supervisor Agent"},
    ResultFormatters: map[string]func(string) string{
        "bochaWebSearch": formatBochaResult,
    },
}

// 创建处理器
processor := ag_ui_util.NewStreamProcessor(config)

// 处理事件流
cleanedEventCh, historyCh := processor.Process(ctx, eventCh)

// 实时处理清洗后的事件
go func() {
    for event := range cleanedEventCh {
        sendToClient(event)
    }
}()

// 收集历史记录
for msg := range historyCh {
    saveHistory(msg)
}
```

## API

### 转换器

| 函数 | 说明 |
|------|------|
| `NewEinoTranslator(threadID, runID)` | 创建 eino 转换器（支持多智能体） |
| `NewOpencodeTranslator(threadID, runID)` | 创建 opencode 转换器 |

### EinoTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, *adk.AsyncIterator[*adk.AgentEvent])` | 转换 eino AgentEvent 迭代器 |

### OpencodeTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, <-chan string)` | 转换 opencode JSON 字符串流 |

### StreamProcessor

| 函数/方法 | 说明 |
|----------|------|
| `NewStreamProcessor(config)` | 创建事件流处理器 |
| `Process(ctx, in)` | 处理事件流，返回清洗后的事件流和历史消息流 |

### 辅助函数

| 函数 | 说明 |
|------|------|
| `EventsToJSONChannel(ctx, events)` | 事件流 → JSON 字符串流 |
| `FormatJSONResult(result)` | 格式化 JSON 字符串 |
| `TruncateResult(maxLen)` | 创建截断函数 |
| `MaskSensitiveFields(fields)` | 创建脱敏函数 |
| `RemovePrefixes(prefixes)` | 创建前缀移除函数 |

### 消息类型

```go
// 常量
const (
    RoleAssistant = "assistant"
    RoleUser      = "user"
    RoleTool      = "tool"
    RoleReasoning = "reasoning"
    RoleSystem    = "system"
    
    ActivityTypeSubAgent  = "sub_agent"
    ActivityTypeWorkspace = "workspace"
)

// 消息类型
type TextMessage struct { ... }
type ReasoningMessage struct { ... }
type ToolMessage struct { ... }
type ToolCall struct { ... }
type Activity struct { ... }

// 活动内容类型
type WorkspaceActivityContent struct {
    RunID     string `json:"runId"`
    ThreadID  string `json:"threadId"`
    FileCount int    `json:"fileCount"`
    TotalSize int64  `json:"totalSize"`
    Timestamp int64  `json:"timestamp"`
}

// 工具结果格式化类型
type WebSearchResult struct { ... }
type WebPage struct { ... }
```
