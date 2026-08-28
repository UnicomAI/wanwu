# WGA Human-in-the-Loop 设计文档

基于 OpenCode 原生 Question 系统，为 WGA 提供人机交互能力。

---

## 目录

- [1. 总体概览](#1-总体概览)
  - [1.1 为什么需要 HITL](#11-为什么需要-hitl)
  - [1.2 为什么基于 OpenCode Question](#12-为什么基于-opencode-question)
  - [1.3 核心设计原则](#13-核心设计原则)
  - [1.4 适用模式](#14-适用模式)
  - [1.5 配置项总览](#15-配置项总览)
- [2. 架构总览](#2-架构总览)
  - [2.1 分层架构](#21-分层架构)
  - [2.2 为什么不需要 BFF 层状态管理](#22-为什么不需要-bff-层状态管理)
  - [2.3 数据流（时序图）](#23-数据流时序图)
- [3. 底层实现（OpenCode + wga-sandbox）](#3-底层实现opencode--wga-sandbox)
  - [3.1 OpenCode Question 系统](#31-opencode-question-系统)
  - [3.2 wga-sandbox 配置项](#32-wga-sandbox-配置项)
  - [3.3 wga-sandbox opencode 配置动态生成](#33-wga-sandbox-opencode-配置动态生成)
  - [3.4 wga-sandbox 事件类型定义](#34-wga-sandbox-事件类型定义)
  - [3.5 wga-sandbox HTTP API（Reply/Reject）](#35-wga-sandbox-http-apireplyreject)
- [4. 中间层实现（wga 包 + ag-ui-util）](#4-中间层实现wga-包--ag-ui-util)
  - [4.1 wga 配置项](#41-wga-配置项)
  - [4.2 wga API（Run、ReplyQuestion、RejectQuestion）](#42-wga-apirunreplyquestionrejectquestion)
  - [4.3 ag-ui-util 事件类型定义](#43-ag-ui-util-事件类型定义)
- [5. 上层实现（BFF）](#5-上层实现bff)
  - [5.1 BFF Handler 与 Service](#51-bff-handler-与-service)
  - [5.2 路由注册](#52-路由注册)
  - [5.3 HTTP API 请求格式](#53-http-api-请求格式)
- [6. 前端实现](#6-前端实现)
  - [6.1 SSE 事件解析](#61-sse-事件解析)
  - [6.2 状态管理](#62-状态管理)
  - [6.3 问题组件](#63-问题组件)
- [7. 修改清单](#7-修改清单)
- [8. 注意事项](#8-注意事项)

---

## 1. 总体概览

### 1.1 为什么需要 HITL

WGA 在执行复杂任务时，需要用户确认、选择或输入信息：
- 确认高风险操作（删除文件、部署到生产环境）
- 选择部署环境（开发/测试/生产）
- 输入敏感信息（API Key、密码）

### 1.2 为什么基于 OpenCode Question

OpenCode 内置 Question 系统提供：
- **阻塞式交互**：AI 调用 `question.ask()` 后阻塞等待
- **多种问题类型**：确认型、选择型、多选型、输入型
- **SSE 事件流**：`question.asked` / `question.replied` / `question.rejected`
- **HTTP API**：`POST /question/:requestID/reply`、`POST /question/:requestID/reject`

### 1.3 核心设计原则

| 原则 | 说明 |
|------|------|
| **层级下沉** | BFF → wga → wga-sandbox → OpenCode，逐层封装 |
| **状态下沉** | 所有状态由 Sandbox 容器内的 OpenCode 管理 |
| **复用原生** | 直接使用 OpenCode 的 Question 系统和 HTTP API |
| **配置控制** | 通过 `EnableHumanInTheLoop` 选项控制是否启用 |

### 1.4 适用模式

| 模式 | HITL 支持 | 说明 |
|------|----------|------|
| **Reuse** | ✅ 支持 | 单实例 Sandbox，固定 Host，无需 Redis |
| **Oneshot** | ❌ 不支持 | 需要存储 questionID → runID → SandboxHost 映射 |

**本设计仅适用于 Reuse 模式。**

### 1.5 配置项总览

```
用户调用 wga.Run(ctx, agentID, 
    wga_option.WithEnableHumanInTheLoop(true),  // ← 用户配置
    ...
)
        │
        ▼
wga/internal/factory/agent_sandbox.go
    │ 传递 EnableHumanInTheLoop
    ▼
wga_sandbox.Run(ctx, 
    wga_sandbox_option.WithEnableHumanInTheLoop(true),
    ...
)
        │
        ▼
wga_sandbox/internal/runner/opencode/opencode.go
    │ renderConfig() 时根据配置生成
    ▼
opencode.json
{
    "permission": {
        "*": "allow",
        "question": "ask"  // ← EnableHumanInTheLoop = true
    }
}
        │
        ▼
OpenCode Question Service
    │ AI 调用 question.ask() 时
    ▼
阻塞等待用户回复
```

| 配置值 | permission.question | 行为 |
|--------|---------------------|------|
| `false`（默认） | `"deny"` | 禁止 question 工具，AI 无法调用 |
| `true` | `"ask"` | 启用 HITL，AI 调用时阻塞等待用户回复 |

---

## 2. 架构总览

### 2.1 分层架构

```
┌──────────────────────────────────────────────────────────────┐
│                         前端 (Vue)                            │
│  sse-parser.js 解析 ACTIVITY_SNAPSHOT                        │
│  QuestionBlock.vue 显示问题 UI，用户交互                      │
└──────────────────────────────────────────────────────────────┘
         │ SSE EventSource                      │ HTTP POST
         │ ACTIVITY_SNAPSHOT                    │ /api/v1/general/agent/question/reply
         ▼                                      ▼
┌──────────────────────────────────────────────────────────────┐
│                      BFF Service (无状态)                     │
│  wga_conversation_chat.go - SSE Handler                      │
│  wga.go (新增) - HTTP Handler                                │
│      └─ 调用 wga.ReplyQuestion() / wga.RejectQuestion()      │
└──────────────────────────────────────────────────────────────┘
         │
         │ 函数调用
         ▼
┌──────────────────────────────────────────────────────────────┐
│                      wga 包 (高级 API)                        │
│  api.go                                                      │
│  ├─ Run(ctx, id, opts...) - 执行智能体                        │
│  │   └─ opts.EnableHumanInTheLoop → 传递到 wga-sandbox       │
│  ├─ ReplyQuestion() - 回答问题 ← 新增                         │
│  └─ RejectQuestion() - 拒绝问题 ← 新增                        │
│      └─ 从全局配置获取 SandboxConfig，调用 wga-sandbox        │
└──────────────────────────────────────────────────────────────┘
         │
         │ 函数调用
         ▼
┌──────────────────────────────────────────────────────────────┐
│                   wga-sandbox 包 (低级 API)                   │
│  api.go                                                      │
│  └─ Run(ctx, opts...)                                        │
│      └─ opts.EnableHumanInTheLoop → 生成 opencode.json       │
│                                                              │
│  api_question.go ← 新增                                      │
│  ├─ ReplyQuestion(ctx, cfg, runID, questionID, answers)             │
│  └─ RejectQuestion(ctx, cfg, runID, questionID)                     │
│      └─ HTTP POST 到 Sandbox:4096                            │
└──────────────────────────────────────────────────────────────┘
         │
         │ HTTP POST (固定地址: wga-sandbox-wanwu:4096)
         ▼
┌──────────────────────────────────────────────────────────────┐
│                   Sandbox 容器 (单实例 - Reuse 模式)          │
│  OpenCode Server (端口 4096)                                  │
│  ├─ SSE /global/event → 发布 question.asked 事件             │
│  ├─ POST /question/:requestID/reply → 解除阻塞               │
│  └─ POST /question/:requestID/reject → 抛出 RejectedError    │
│                                                              │
│  Question Service (内存)                                      │
│  ├─ pending: Map<QuestionID, { info, deferred }>             │
│  ├─ ask() → 创建 Deferred, 发布事件, await 阻塞              │
│  ├─ reply() → Deferred.succeed(answers)                      │
│  └─ reject() → Deferred.fail(RejectedError)                  │
│                                                              │
│  opencode.json (动态生成)                                     │
│  └─ permission.question: "ask" | "deny"                      │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 为什么不需要 BFF 层状态管理

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SSE 与 HTTP API 是独立的                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  SSE 连接（长连接，建立时绑定）                                              │
│  ─────────────────────────────                                              │
│  用户 ◄───────────────────────► BFF-A ◄─────────────────────► Sandbox      │
│        SSE EventSource            wga.Run()                      SSE         │
│        (长连接)                    runner.Run()                  /global/event│
│                                    outputCh ◄─────────────────               │
│                                                                             │
│  HTTP Reply（短连接，无状态）                                                │
│  ─────────────────────────────                                              │
│  用户 ────────────────────────► BFF-B ──────────────────────► Sandbox      │
│        POST /question/reply        wga.ReplyQuestion()          POST        │
│        (短连接)                    从配置获取 Host              /question/id │
│                                                                             │
│  流程：                                                                     │
│  1. question.asked → Sandbox → BFF-A → 用户                                │
│  2. 用户回复 → BFF-B → wga.ReplyQuestion() → Sandbox HTTP API              │
│  3. Sandbox 解除阻塞，继续执行                                              │
│  4. 后续事件 → Sandbox → BFF-A (已建立的SSE连接) → 用户 ✓                   │
│                                                                             │
│  关键：                                                                     │
│  - SSE 连接在 BFF-A 上保持                                                  │
│  - HTTP reply 可以到任何 BFF 实例                                           │
│  - Reuse 模式下 Sandbox Host 固定，所有 BFF 实例连接同一个 Sandbox          │
│  - questionID 在 Sandbox 的 pending Map 中唯一，不需要额外映射              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 数据流（时序图）

#### 2.3.1 Question 提交流程

**前提**：`EnableHumanInTheLoop = true`

```
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│OpenCode │  │ Runner  │  │wga-     │  │  wga    │  │  BFF    │  │  前端   │
│ Service │  │(opencode)│  │sandbox  │  │  api    │  │ Handler │  │         │
│ (容器)  │  │         │  │         │  │         │  │         │  │         │
└────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
     │            │            │            │            │            │
     │            │            │            │ wga.Run()  │            │
     │            │            │            │ EnableHITL │            │
     │            │            │<───────────│ = true     │            │
     │            │            │            │            │            │
     │            │            │ renderConfig()          │            │
     │            │            │ permission: │            │            │
     │            │            │   question:│            │            │
     │            │            │   "ask"    │            │            │
     │            │            │            │            │            │
     │ AI 调用    │            │            │            │            │
     │ question.  │            │            │            │            │
     │ ask()      │            │            │            │            │
     ├───────────>│            │            │            │            │
     │            │            │            │            │            │
     │ 创建       │            │            │            │            │
     │ Deferred   │            │            │            │            │
     │ pending.   │            │            │            │            │
     │ set(id,    │            │            │            │            │
     │  deferred) │            │            │            │            │
     │            │            │            │            │            │
     │ 发布 SSE   │            │            │            │            │
     │ question.  │            │            │            │            │
     │ asked      │            │            │            │            │
     ├───────────>│            │            │            │            │
     │            │            │            │            │            │
     │            │ handleBus  │            │            │            │
     │            │ Event()    │            │            │            │
     │            │            │            │            │            │
     │            │ 转换为     │            │            │            │
     │            │ AG-UI Event│            │            │            │
     │            ├───────────>│            │            │            │
     │            │            │            │            │            │
     │            │            │ EinoIter.  │            │            │
     │            │            │ Translate()│            │            │
     │            │            │            │            │            │
     │            │            │ eventCh    │            │            │
     │            │            ├───────────>│            │            │
     │            │            │            │            │            │
     │            │            │            │ SSE 转发   │            │
     │            │            │            ├───────────>│            │
     │            │            │            │            │            │
     │            │            │            │            │ 渲染问题UI │
     │            │            │            │            │            │
     │ Deferred   │            │            │            │            │
     │ await 阻塞 │            │            │            │            │
     │ (等待用户  │            │            │            │            │
     │  回复...)  │            │            │            │            │
```

#### 2.3.2 Reply 流程

```
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│OpenCode │  │ Runner  │  │wga-     │  │  wga    │  │  BFF    │  │  前端   │
│ Service │  │(opencode)│  │sandbox  │  │  api    │  │ Handler │  │         │
│ (容器)  │  │         │  │         │  │         │  │         │  │         │
└────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
     │            │            │            │            │            │
     │ Deferred   │            │            │            │            │
     │ 阻塞等待...│            │            │            │            │
     │            │            │            │            │            │
     │            │            │            │            │ 用户点击  │
     │            │            │            │            │ 提交      │
     │            │            │            │            │<───────────│
     │            │            │            │            │            │
     │            │            │            │            │ POST       │
     │            │            │            │            │ /question/ │
     │            │            │            │            │ reply      │
     │            │            │            │            │<───────────│
     │            │            │            │            │            │
     │            │            │            │<───────────│            │
     │            │            │            │ wga.       │            │
     │            │            │            │ ReplyQuestion()         │
     │            │            │            │            │            │
     │            │            │<───────────│            │            │
     │            │            │ wga_sandbox.            │            │
     │            │            │ ReplyQuestion()         │            │
     │            │            │            │            │            │
     │<───────────────────────│            │            │            │
     │            │            │ HTTP POST  │            │            │
     │            │            │ cfg.OpencodeEndpoint()  │            │
     │            │            │ + /question/:id/reply   │            │
     │            │            │            │            │            │
     │ pending.   │            │            │            │            │
     │ get(id)    │            │            │            │            │
     │            │            │            │            │            │
     │ Deferred.  │            │            │            │            │
     │ succeed    │            │            │            │            │
     │ (answers)  │            │            │            │            │
     │            │            │            │            │            │
     │ 解除阻塞   │            │            │            │            │
     │ AI 继续执行│            │            │            │            │
     │            │            │            │            │            │
     │ 发布 SSE   │            │            │            │            │
     │ question.  │            │            │            │            │
     │ replied    │            │            │            │            │
     ├───────────>│            │            │            │            │
     │            │            │            │            │            │
     │            │ handleBus  │            │            │            │
     │            │ Event()    │            │            │            │
     │            │            │            │            │            │
     │            │ 转换为     │            │            │            │
     │            │ AG-UI Event│            │            │            │
     │            ├───────────>│            │            │            │
     │            │            │            │            │            │
     │            │            │ EinoIter.  │            │            │
     │            │            │ Translate()│            │            │
     │            │            │            │            │            │
     │            │            │ eventCh    │            │            │
     │            │            ├───────────>│            │            │
     │            │            │            │            │            │
     │            │            │            │ SSE 转发   │            │
     │            │            │            ├───────────>│            │
     │            │            │            │            │            │
     │            │            │            │            │ 更新UI     │
```

**调用链**：`前端 → BFF → wga → wga-sandbox → OpenCode`

---

## 3. 底层实现（OpenCode + wga-sandbox）

### 3.1 OpenCode Question 系统

#### 3.1.1 核心机制

```typescript
// packages/opencode/src/question/index.ts

interface PendingEntry {
  info: Request
  deferred: Deferred.Deferred<ReadonlyArray<Answer>, RejectedError>
}

const pending = new Map<QuestionID, PendingEntry>()

// ask - 创建 Deferred 并阻塞等待
const ask = Effect.fn("Question.ask")(function* (input) {
  const id = QuestionID.ascending()  // question_01JXYZ...
  const deferred = yield* Deferred.make<ReadonlyArray<Answer>, RejectedError>()
  pending.set(id, { info, deferred })
  yield* bus.publish(Event.Asked, info)
  
  return yield* Deferred.await(deferred)  // 阻塞等待
})

// reply - 解除阻塞（只需要 requestID）
const reply = Effect.fn("Question.reply")(function* (input) {
  const existing = pending.get(input.requestID)  // ✅ 只需要 questionID
  yield* Deferred.succeed(existing.deferred, input.answers)
})

// reject - 抛出错误
const reject = Effect.fn("Question.reject")(function* (requestID) {
  const existing = pending.get(requestID)
  yield* Deferred.fail(existing.deferred, new RejectedError())
})
```

**关键点**：
- `pending` Map 以 `questionID` 为 key，不需要额外映射
- `reply` 只需要 `requestID`（即 `questionID`），无需 `runID` 或 `sessionID`
- `Deferred.await()` 使 AI 纤程阻塞，直到用户回复

#### 3.1.2 HTTP API 端点

| 端点 | 方法 | 请求体 | 说明 |
|------|------|--------|------|
| `/question/:requestID/reply` | POST | `{"answers": [["答案"]]}` | 回答问题 |
| `/question/:requestID/reject` | POST | `{}` | 取消问题 |
| `/question` | GET | - | 列出所有待处理问题 |

#### 3.1.3 SSE 事件格式

```json
{
  "directory": "/workspace/xxx",
  "payload": {
    "type": "question.asked",
    "properties": {
      "id": "question_01JXYZ...",
      "sessionID": "session-xxx",
      "questions": [{
        "question": "请选择部署环境",
        "header": "环境选择",
        "options": [
          { "label": "开发环境", "description": "用于开发和测试" },
          { "label": "生产环境", "description": "正式生产部署" }
        ],
        "multiple": false,
        "custom": false
      }]
    }
  }
}
```

**事件类型**：
- `question.asked` - AI 提出问题，等待用户回复
- `question.replied` - 用户已回复问题
- `question.rejected` - 用户取消问题

---

### 3.2 wga-sandbox 配置项

**文件**: `pkg/wga-sandbox/wga-sandbox-option/option.go`

```go
type RunOption struct {
    // ... 现有字段
    EnableHumanInTheLoop       bool  // 是否启用人机交互
    EnableHumanInTheLoopCustom bool  // 是否允许用户自定义回答
}

// WithEnableHumanInTheLoop 启用人机交互。
// enableCustom 为可选参数，设置是否允许用户自定义回答。
func WithEnableHumanInTheLoop(enable bool, enableCustom ...bool) Option {
    return OptionFunc(func(opts *RunOption) error {
        opts.EnableHumanInTheLoop = enable
        if len(enableCustom) > 0 {
            opts.EnableHumanInTheLoopCustom = enableCustom[0]
        }
        return nil
    })
}
```

**注意**: OpenCode 的 `question.ask()` 工具参数 `Prompt` 类型没有 `custom` 字段，
但 `Info` 类型有 `custom` 字段（文档注释说默认为 `true`）。
当 OpenCode 未传递 `custom` 字段时，使用 `EnableHumanInTheLoopCustom` 配置值（默认 `false`）。

---

### 3.3 wga-sandbox opencode 配置动态生成

**文件**: `pkg/wga-sandbox/internal/runner/opencode/opencode.go`

configTemplate 修改：根据 `EnableHumanInTheLoop` 配置动态设置 `permission.question`：

| EnableHumanInTheLoop | permission.question |
|---------------------|---------------------|
| `false`（默认） | 不设置（默认 deny） |
| `true` | `"ask"` |

---

### 3.4 wga-sandbox 事件类型定义

**文件**: `pkg/wga-sandbox/internal/runner/opencode/types.go`

```go
const (
    // 现有事件类型...
    
    // 新增 question 事件类型
    OpencodeEventTypeQuestionAsked    OpencodeEventType = "question.asked"
    OpencodeEventTypeQuestionReplied  OpencodeEventType = "question.replied"
    OpencodeEventTypeQuestionRejected OpencodeEventType = "question.rejected"
)
```

---

### 3.5 wga-sandbox HTTP API（Reply/Reject）

**文件**: `pkg/wga-sandbox/api_question.go`（新增）

| 函数 | 说明 |
|------|------|
| `ReplyQuestion(ctx, cfg, questionID, answers)` | 回答问题 |
| `RejectQuestion(ctx, cfg, questionID)` | 拒绝问题 |

**设计说明**：
- 复用 `SandboxConfig.OpencodeEndpoint()` 方法（已包含端口 4096）
- 与现有 `wga_sandbox.Run()` 接口风格一致

---

## 4. 中间层实现（wga 包 + ag-ui-util）

### 4.1 wga 配置项

**文件**: `pkg/wga/internal/option/option.go`

```go
type Options struct {
    // ... 现有字段
    EnableHumanInTheLoop       bool  // 是否启用人机交互
    EnableHumanInTheLoopCustom bool  // 是否允许用户自定义回答
}

// WithEnableHumanInTheLoop 启用人机交互。
// enableCustom 为可选参数，设置是否允许用户自定义回答。
func WithEnableHumanInTheLoop(enable bool, enableCustom ...bool) Option {
    return optionFunc(func(opts *Options) error {
        opts.EnableHumanInTheLoop = enable
        if len(enableCustom) > 0 {
            opts.EnableHumanInTheLoopCustom = enableCustom[0]
        }
        return nil
    })
}
```

**文件**: `pkg/wga/wga-option/option_api.go`

```go
// 导出 WithEnableHumanInTheLoop 选项
var WithEnableHumanInTheLoop = option.WithEnableHumanInTheLoop
```

---

### 4.2 wga API（Run、ReplyQuestion、RejectQuestion）

**文件**: `pkg/wga/api.go`

| 函数 | 说明 |
|------|------|
| `Run(ctx, id, opts...)` | 执行智能体（现有） |
| `ReplyQuestion(ctx, sandboxCfg, runID, questionID, answers)` | 回答问题（新增） |
| `RejectQuestion(ctx, sandboxCfg, runID, questionID)` | 拒绝问题（新增） |

**设计说明**：
- wga 层不管理全局配置，由业务层（BFF Service）从配置中获取 SandboxConfig 并传入
- API 直接调用 `wga_sandbox.ReplyQuestion()` 和 `wga_sandbox.RejectQuestion()`
- 符合分层架构：BFF Service 负责配置管理，wga 提供纯函数 API

---

### 4.3 ag-ui-util 事件类型定义

**文件**: `pkg/ag-ui-util/message_types.go`

```go
const (
    // 现有活动类型
    ActivityTypeSubAgent  = "sub_agent"
    ActivityTypeWorkspace = "workspace"
    
    // 新增 question 活动类型
    ActivityTypeQuestion = "question"
)

// QuestionActivityContent 问题活动内容
type QuestionActivityContent struct {
    QuestionID string         `json:"questionId"`
    RunID      string         `json:"runId"`
    ThreadID   string         `json:"threadId"`
    Status     string         `json:"status"` // "pending", "answered", "rejected"
    Questions  []QuestionItem `json:"questions"`
    Answers    [][]string     `json:"answers,omitempty"` // 用户答案（status="answered" 时）
    Timestamp  int64          `json:"timestamp"`
}

// QuestionItem 问题项
type QuestionItem struct {
    Question string   `json:"question"`
    Header   string   `json:"header"`
    Options  []Option `json:"options"`
    Multiple bool     `json:"multiple"`
    Custom   bool     `json:"custom"`
}

// Option 选项
type Option struct {
    Label       string `json:"label"`
    Description string `json:"description"`
}
```

**状态流转**：
```
question.asked   → status: "pending"   → 显示问题 UI
question.replied → status: "answered"  → 关闭 UI，显示答案
question.rejected→ status: "rejected"  → 关闭 UI，显示取消
```

---

## 5. 上层实现（BFF）

### 5.1 BFF Handler 与 Service

**Handler**: `internal/bff-service/server/http/handler/v1/wga.go`

| Handler | 路由 | 说明 |
|---------|------|------|
| `GeneralAgentReplyQuestion` | POST `/general/agent/question/reply` | 回答问题 |
| `GeneralAgentRejectQuestion` | POST `/general/agent/question/reject` | 拒绝问题 |

**Service**: `internal/bff-service/service/wga.go`

| Service 函数 | 说明 |
|-------------|------|
| `GeneralAgentReplyQuestion(ctx, runID, questionID, answers)` | 回答问题 |
| `GeneralAgentRejectQuestion(ctx, runID, questionID)` | 拒绝问题 |

**配置**: `internal/bff-service/config/wga.go`

```go
type WgaConfig struct {
    // ... 现有字段
    HumanInTheLoop bool `yaml:"humanInTheLoop" json:"humanInTheLoop" mapstructure:"humanInTheLoop"`
}
```

---

### 5.2 路由注册

**文件**: `internal/bff-service/server/http/handler/router/v1/wga.go`

```go
// 人机交互相关接口
mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/question/reply", http.MethodPost, v1.GeneralAgentReplyQuestion, "回答问题")
mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/question/reject", http.MethodPost, v1.GeneralAgentRejectQuestion, "拒绝问题")
```

---

### 5.3 HTTP API 请求格式

```json
// POST /api/v1/general/agent/question/reply
{
  "runId": "run-xxx",
  "questionId": "question_01JXYZ...",
  "answers": [["开发环境"]]
}

// POST /api/v1/general/agent/question/reject
{
  "runId": "run-xxx",
  "questionId": "question_01JXYZ..."
}
```

---

## 6. 前端实现

### 6.1 SSE 事件解析

**文件**: `web/src/views/generalAgent/utils/sse-parser.js`

```javascript
// 添加 QUESTION 活动类型
export const ActivityType = {
    // ... 现有类型
    QUESTION: 'question',
};
```

---

### 6.2 状态管理

**文件**: 
- `web/src/views/generalAgent/mixins/streamStateManager.js`
- `web/src/views/generalAgent/utils/message-aggregator.js`

处理 `ActivityType.QUESTION` 类型的 `ACTIVITY_SNAPSHOT` 事件：

```javascript
// pending 状态创建新 fragment
{
    type: 'question',
    questionId: activityContent.questionId,
    runId: activityContent.runId,
    status: activityContent.status,
    questions: activityContent.questions || [],
    answers: activityContent.answers,  // 用户答案（历史记录回显）
    timestamp: activityContent.timestamp || Date.now(),
}

// answered/rejected 状态更新已存在的 fragment
existingFragment.status = status;
existingFragment.answers = activityContent.answers || existingFragment.answers;
```

---

### 6.3 问题组件

**文件**: `web/src/views/generalAgent/components/QuestionBlock.vue`（新增）

| 功能 | 说明 |
|------|------|
| 显示问题 | 展示 `questions` 数组中的所有问题 |
| 选项选择 | 支持单选/多选（根据 `multiple` 字段） |
| 自定义输入 | 支持 `custom: true` 时用户自定义输入 |
| 提交回答 | 调用 `replyQuestion` API |
| 取消问题 | 调用 `rejectQuestion` API |
| 状态显示 | 根据 `status` 显示 pending/answered/rejected |
| 答案回显 | `answers` 属性支持历史记录回显用户选择 |

---

## 7. 修改清单

### 7.1 后端

| 层级 | 文件 | 类型 | 说明 |
|------|------|------|------|
| 底层 | `pkg/wga-sandbox/wga-sandbox-option/option.go` | 修改 | 添加 `EnableHumanInTheLoop` 字段和选项函数 |
| 底层 | `pkg/wga-sandbox/internal/runner/opencode/types.go` | 修改 | 添加 question 事件类型常量 |
| 底层 | `pkg/wga-sandbox/internal/runner/opencode/opencode.go` | 修改 | 动态生成 permission、处理 question 事件 |
| 底层 | `pkg/wga-sandbox/api_question.go` | 新增 | HTTP 调用实现 |
| 底层 | `pkg/wga-sandbox/api_opencode.go` | 修改 | 添加 `ParseOpencodeQuestionPart` |
| 底层 | `pkg/wga-sandbox/wga-sandbox-converter/opencode.go` | 修改 | 添加 question 事件解析处理 |
| 中间层 | `pkg/wga/internal/option/option.go` | 修改 | 添加 `EnableHumanInTheLoop` 字段和选项函数 |
| 中间层 | `pkg/wga/wga-option/option_api.go` | 修改 | 导出 `WithEnableHumanInTheLoop` 选项 |
| 中间层 | `pkg/wga/api.go` | 修改 | 添加 `ReplyQuestion`、`RejectQuestion` 接口 |
| 中间层 | `pkg/wga/internal/factory/agent_sandbox.go` | 修改 | 传递 `EnableHumanInTheLoop` 到 wga-sandbox |
| 中间层 | `pkg/ag-ui-util/message_types.go` | 修改 | 添加 `ActivityTypeQuestion`、`QuestionActivityContent` |
| 中间层 | `pkg/ag-ui-util/translator_eino.go` | 修改 | 添加 `translateQuestionEvent` 方法 |
| 中间层 | `pkg/ag-ui-util/translator_opencode.go` | 修改 | 添加 `translateQuestion` 方法 |
| 上层 | `configs/microservice/bff-service/configs/wga/config.yaml` | 修改 | 添加 `humanInTheLoop: true` 配置 |
| 上层 | `internal/bff-service/config/wga.go` | 修改 | 添加 `HumanInTheLoop` 字段 |
| 上层 | `internal/bff-service/model/request/wga.go` | 修改 | 添加请求结构体 |
| 上层 | `internal/bff-service/service/wga.go` | 修改 | 添加 Service 函数 |
| 上层 | `internal/bff-service/service/wga_conversation_chat.go` | 修改 | 添加 `WithEnableHumanInTheLoop` 选项 |
| 上层 | `internal/bff-service/server/http/handler/v1/wga.go` | 修改 | 添加 Handler |
| 上层 | `internal/bff-service/server/http/handler/router/v1/wga.go` | 修改 | 注册路由 |

### 7.2 前端

| 序号 | 文件 | 类型 | 说明 |
|------|------|------|------|
| 1 | `web/src/views/generalAgent/utils/sse-parser.js` | 修改 | 添加 `ActivityType.QUESTION` 解析 |
| 2 | `web/src/views/generalAgent/mixins/streamStateManager.js` | 修改 | question fragment 状态管理、answers 更新 |
| 3 | `web/src/views/generalAgent/utils/message-aggregator.js` | 修改 | question 消息聚合、answers 字段处理 |
| 4 | `web/src/views/generalAgent/components/MessageItem.vue` | 修改 | 添加 `type: 'question'` fragment 渲染 |
| 5 | `web/src/views/generalAgent/components/QuestionBlock.vue` | 新增 | 问题显示组件，支持 answers 属性回显 |
| 6 | `web/src/views/generalAgent/components/ActivityBlock.vue` | 修改 | pending question 时保持展开 |
| 7 | `web/src/views/generalAgent/index.vue` | 修改 | 添加事件处理 |
| 8 | `web/src/api/generalAgent.js` | 修改 | 添加 replyQuestion/rejectQuestion API |
| 9 | `web/src/lang/zh.js` | 修改 | 添加中文翻译 |
| 10 | `web/src/lang/en.js` | 修改 | 添加英文翻译 |

---

## 8. 注意事项

### 8.1 模式限制

| 模式 | HITL 支持 | 原因 |
|------|----------|------|
| **Reuse** | ✅ 支持 | 单实例 Sandbox，questionID 在 pending Map 中唯一 |
| **Oneshot** | ❌ 不支持 | 需要额外映射 questionID → runID → SandboxHost |

### 8.2 水平扩展

**Reuse 模式下 BFF 可水平扩展**：
- SSE 连接在某个 BFF 实例上保持
- HTTP reply 可到达任意 BFF 实例
- 所有 BFF 实例调用同一个 Sandbox Host
- 不需要 Redis、不需要会话粘性

### 8.3 Reject 后 AI 行为

AI 收到 `RejectedError` 后可选择：
1. **优雅降级**：跳过需确认的步骤，继续其他操作
2. **重新提问**：换种方式重新询问
3. **终止任务**：礼貌终止并说明原因

### 8.4 API 调用条件

- `ReplyQuestion` 和 `RejectQuestion` 仅在 Reuse 模式下有效
- 若 Sandbox 配置为 Oneshot 模式或未配置，Service 层将返回错误
- 若未启用 `EnableHumanInTheLoop`，OpenCode 无 pending questions，调用这些 API 无实际效果（但不会报错）

---

## 9. 参考资料

### OpenCode 源码

| 组件 | 文件 | 说明 |
|------|------|------|
| Question Service | `packages/opencode/src/question/index.ts` | Deferred 阻塞机制 |
| Question Schema | `packages/opencode/src/question/schema.ts` | QuestionID 格式 |
| HTTP Routes | `packages/opencode/src/server/routes/instance/question.ts` | API 端点 |
