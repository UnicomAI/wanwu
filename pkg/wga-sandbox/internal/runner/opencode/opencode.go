// Package opencode 提供 opencode 智能体的运行器实现（基于 HTTP API）。
//
// 基于 opencode >= v1.4.11 的 SSE 事件格式实现，兼容 v1.14.51、v1.15.13、v1.16.2。
//
// 版本基线说明：
//
//	v1.1.65 仅支持 BusEvent 单格式 SSE（src/server/routes/global.ts 仅使用
//	BusEvent.payloads()）。v1.4.11 引入 BusEvent/SyncEvent 双格式（参见 e8f0faee），
//	SSE schema 变为 z.union([BusEvent.payloads(), SyncEvent.payloads()])，奠定了
//	当前 SSE 事件解析的架构基础。后续版本增量添加事件和字段，线上格式始终向后兼容：
//	- v1.14.51：BusEvent payload 新增 id 字段（上升标识符），GlobalBus 新增 project 字段，
//	  引入 session.next.* 细粒度事件（实验性标志，默认关闭）
//	- v1.15.13：引入 EventV2 核心系统（packages/core/src/event.ts），EventV2Bridge
//	  通过 legacy Bus/SyncEvent 间接输出，线上格式不变
//	- v1.16.2：EventV2Bridge 直接输出至 GlobalBus，session.next.* 事件默认开启，
//	  线上 BusEvent/SyncEvent 格式仍不变
//
// 通过 SSE 连接接收 opencode 事件流，转换为统一的 JSON 格式输出。
//
// v1.16.2 兼容性说明：
//   - v1.16.2 将内部事件系统从 BusEvent 迁移至 EventV2，但 EventV2Bridge 保持了
//     向后兼容的 SSE 输出格式：BusEvent 格式 {id, type, properties} 和 SyncEvent
//     格式 {type:"sync", syncEvent:{id, type, seq, aggregateID, data}} 均未变。
//   - HTTP API 端点（/session, /session/{id}/prompt_async, /question/{id}/reply 等）
//     路径和请求/响应格式未变。
//   - Part 类型（text, reasoning, tool, step-start, step-finish 等）字段结构未变。
//   - session.idle 事件仍向后兼容发送（同时发送 session.status）。
//   - 新增 SSE 字段（如 payload.project, payload.workspace）被 Go json.Unmarshal
//     自动忽略，无需代码修改。
//   - 新增 session.next.* 系列细粒度事件（SyncEvent），当前未处理，不影响现有逻辑。
//   - 注意：未来版本可能移除 BusEvent 兼容层，届时需适配 EventV2 直出格式。
package opencode

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// ============================================================================
// 常量
// ============================================================================

const (
	// opencode.json 配置文件模板
	configTemplate = `{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "*": "allow"{{if .EnableHumanInTheLoop}},
    "question": "ask"{{else}},
    "question": "deny"{{end}}
  },
  "agent": {
    "title": { "disable": true }
  },
  "provider": {
    "{{.Provider}}": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "{{.ProviderName}}",
      "options": {
        "baseURL": "{{.BaseURL}}",
        "apiKey": "{{.APIKey}}"
      },
      "models": {
        "{{.Model}}": {
          "name": "{{.ModelName}}"
        }
      }
    }
  }{{if .MCPs}},
  "mcp": {
{{range $i, $mcp := .MCPs}}    "{{$mcp.Name}}": {
      "type": "remote",
      "url": "{{$mcp.URL}}",
      "description": "{{$mcp.Description}}",
      "enabled": true{{if $mcp.Headers}},
      "headers": {
{{- range $k, $v := $mcp.Headers}}        "{{$k}}": "{{$v}}"{{end}}
      }{{end}}
    }{{if ne $i (sub (len $.MCPs) 1)}},{{end}}
{{end}}  }{{end}}
}`

	// 系统提示词模板
	systemTemplate = `# 任务要求

{{if .Instruction}}---

## 系统提示词

{{.Instruction}}

{{end}}{{if .OverallTask}}---

## 整体任务

{{.OverallTask}}

{{end}}{{if .Messages}}---

## 历史信息

{{range .Messages}}### role: {{.Role}}

{{.ReasoningContent}}{{if .UserInputMultiContent}}{{$parts := listParts .UserInputMultiContent}}{{range $i, $part := $parts}}{{if $i}}

{{end}}{{$part}}{{end}}{{else}}{{.Content}}{{end}}

{{end}}{{end}}`
)

// ============================================================================
// 类型
// ============================================================================

// 确保 Runner 实现 runner.Runner 接口
var _ runner.Runner = (*Runner)(nil)

// Runner 实现 opencode 智能体运行器（基于 HTTP API）。
// 通过 SSE 连接接收事件流，转换为 JSON 格式输出。
type Runner struct {
	sb              sandbox.Sandbox
	opt             wga_sandbox_option.RunOption
	sessionID       string
	userMsgIDs      map[string]bool // 用户消息 ID 集合，用于过滤
	logPrefix       string
	partTypeTracker *partTypeTracker // partID -> part 类型映射，用于 delta 事件判断 text/reasoning
}

// partTypeTracker 跟踪 partID 到 part 类型的映射。
// message.part.delta 事件只携带 partID 而不携带 part 类型，
// 需要通过 message.part.updated 事件中记录的 part 类型来推断 delta 的类型。
type partTypeTracker struct {
	types map[string]string // partID -> type ("text" | "reasoning" | "tool" | ...)
}

func newPartTypeTracker() *partTypeTracker {
	return &partTypeTracker{types: make(map[string]string)}
}

// set 记录 partID 对应的类型。
func (t *partTypeTracker) set(partID, partType string) {
	if partID != "" && partType != "" {
		t.types[partID] = partType
	}
}

// get 获取 partID 对应的类型。
func (t *partTypeTracker) get(partID string) string {
	return t.types[partID]
}

// stepStart 在新的 step 开始时调用，清理旧的 part 类型映射。
func (t *partTypeTracker) stepStart(stepID string) {
	// 新 step 开始时清理上一轮的映射，避免内存泄漏
	t.types = make(map[string]string)
}

// stepFinish 在 step 结束时调用。
func (t *partTypeTracker) stepFinish() {
	// step 结束时不清理，因为可能还有后续 delta 事件
}

// ============================================================================
// 公开方法
// ============================================================================

// NewRunner 创建 opencode 运行器实例。
func NewRunner(sb sandbox.Sandbox, opt wga_sandbox_option.RunOption) runner.Runner {
	logPrefix := fmt.Sprintf("[wga-sandbox][%s]", opt.RunSession.RunID)
	return &Runner{
		sb:              sb,
		opt:             opt,
		userMsgIDs:      make(map[string]bool),
		logPrefix:       logPrefix,
		partTypeTracker: newPartTypeTracker(),
	}
}

// BeforeRun 执行前准备工作：
// 1. 创建 opencode 配置文件
// 2. 复制 skills 和 tools
// 3. 复制输入文件
// 4. 创建 opencode session
// 注意：沙箱环境已在 Manager.Create 时通过 Prepare 初始化，此处不再调用
func (r *Runner) BeforeRun(ctx context.Context) error {
	// 恢复 trace 上下文
	if r.opt.TraceContext != nil {
		ctx = trace_util.InjectTraceHeaders(ctx, r.opt.TraceContext)
	}

	if err := r.setupConfig(ctx); err != nil {
		return err
	}

	if err := r.setupSkills(ctx); err != nil {
		return err
	}

	if err := r.setupTools(ctx); err != nil {
		return err
	}

	if r.opt.InputDir != "" {
		if err := r.sb.CopyToSandbox(ctx, r.opt.InputDir); err != nil {
			return fmt.Errorf("failed to copy input to workspace: %w", err)
		}
	}

	if err := r.createSession(ctx); err != nil {
		return fmt.Errorf("failed to create opencode session: %w", err)
	}

	return nil
}

// Run 执行智能体任务，返回 JSON 格式事件流。
// 通过 SSE 连接接收 opencode 事件，过滤并转换为 JSON 格式输出。
func (r *Runner) Run(ctx context.Context) (<-chan string, error) {
	// 恢复 trace 上下文，确保 HTTP 请求传播 traceparent 头
	if r.opt.TraceContext != nil {
		ctx = trace_util.InjectTraceHeaders(ctx, r.opt.TraceContext)
	}

	sseCh, err := r.connectSSE(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SSE: %w", err)
	}

	if err := r.sendPromptAsync(ctx); err != nil {
		return nil, fmt.Errorf("failed to send prompt: %w", err)
	}

	outputCh := make(chan string, 1024)

	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)
		r.processSSEEvents(ctx, sseCh, outputCh)
	}()

	return outputCh, nil
}

// AfterRun 执行后处理：
// 1. 删除 opencode session
// 2. 复制输出文件到本地（如果指定了 OutputDir）
// 沙箱清理由外部统一管理，不在此处处理
func (r *Runner) AfterRun(ctx context.Context) error {
	r.deleteSession(ctx)

	if r.opt.OutputDir != "" {
		return r.copyOutput(ctx)
	}
	return nil
}

// ============================================================================
// 生命周期 - 准备
// ============================================================================

// setupConfig 创建 opencode 配置文件。
func (r *Runner) setupConfig(ctx context.Context) error {
	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", ".opencode"); err != nil {
		return fmt.Errorf("failed to create .opencode directory: %w", err)
	}

	// 如果存在 trace 上下文，将 traceId/spanId 编码到 baseURL 中
	// 这样 opencode 进程调用模型 API 时，URL 会变成
	// {baseURL}/trace/{traceId}/span/{spanId}/chat/completions
	// BFF 侧新增带 /trace/:traceId/span/:spanId/ 参数的路由来接收这种请求
	modelConfig := r.opt.ModelConfig
	if r.opt.TraceContext != nil {
		if tp, ok := r.opt.TraceContext["traceparent"]; ok {
			parts := strings.Split(tp, "-")
			if len(parts) == 4 {
				traceID := parts[1]
				spanID := parts[2]
				modelConfig.BaseURL = modelConfig.BaseURL + "/trace/" + traceID + "/span/" + spanID
			}
		}
	}

	content, err := renderConfig(modelConfig, r.opt.MCPs, r.opt.EnableHumanInTheLoop)
	if err != nil {
		return fmt.Errorf("failed to render config: %w", err)
	}
	if err := r.sb.WriteFile(ctx, ".opencode/opencode.json", []byte(content)); err != nil {
		return fmt.Errorf("failed to create opencode.json: %w", err)
	}

	return nil
}

// setupSkills 复制 skills 到工作目录，并注入变量到 SKILL.md。
func (r *Runner) setupSkills(ctx context.Context) error {
	if len(r.opt.Skills) == 0 {
		return nil
	}

	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", ".opencode/skills"); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	for _, skill := range r.opt.Skills {
		dirName := path.Base(skill.Dir)
		if err := r.sb.CopyToSandbox(ctx, skill.Dir, ".opencode/skills"); err != nil {
			return fmt.Errorf("failed to copy skill %s to workspace: %w", dirName, err)
		}

		skillDir := ".opencode/skills/" + dirName
		skillPath := fmt.Sprintf("%s/SKILL.md", skillDir)

		// 追加变量信息到 SKILL.md
		if len(skill.Variables) > 0 {
			varsContent := formatVariablesContent(skill.Variables)
			encoded := base64.StdEncoding.EncodeToString([]byte(varsContent))
			cmd := fmt.Sprintf("echo '%s' | base64 -d >> \"%s\"", encoded, skillPath)
			if _, err := r.sb.ExecuteSync(ctx, cmd); err != nil {
				return fmt.Errorf("failed to update SKILL.md for skill %s: %w", dirName, err)
			}
		}

	}

	return nil
}

// setupTools 转换 tools 为 skills 并复制到工作目录。
func (r *Runner) setupTools(ctx context.Context) error {
	if len(r.opt.Tools) == 0 {
		return nil
	}

	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", ".opencode/tools"); err != nil {
		return fmt.Errorf("failed to create tools directory: %w", err)
	}

	if _, err := r.sb.ExecuteSync(ctx, "mkdir", "-p", ".opencode/skills"); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	for _, tool := range r.opt.Tools {
		if err := r.setupTool(ctx, tool); err != nil {
			return err
		}
	}

	return nil
}

// setupTool 处理单个 tool。
func (r *Runner) setupTool(ctx context.Context, tool wga_sandbox_option.Tool) error {
	// 写入 OpenAPI schema 文件
	schemaData, err := json.Marshal(tool.OpenAPI3Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal tool schema %s: %w", tool.Name, err)
	}

	dstFileName := fmt.Sprintf("%s.%s.json", toSkillName(tool.Name), uuid.New().String()[:8])
	dstPath := ".opencode/tools/" + dstFileName
	if err := r.sb.WriteFile(ctx, dstPath, schemaData); err != nil {
		return fmt.Errorf("failed to write tool schema %s: %w", tool.Name, err)
	}

	// 使用 openapi2skill 转换为 skill
	skillName := toSkillName(tool.Name)
	if _, err := r.sb.ExecuteSync(ctx, "openapi2skill", dstPath, "-o", ".opencode/skills", "-n", skillName, "-f"); err != nil {
		return fmt.Errorf("failed to convert tool %s to skill: %w", tool.Name, err)
	}

	// 追加 API 认证信息到 SKILL.md
	if tool.APIAuth != nil && tool.APIAuth.Type != "none" && tool.APIAuth.Value != "" {
		skillDir := ".opencode/skills/" + skillName
		skillPath := fmt.Sprintf("%s/SKILL.md", skillDir)
		authContent := formatAuthContent(tool.APIAuth)
		encoded := base64.StdEncoding.EncodeToString([]byte(authContent))
		cmd := fmt.Sprintf("echo '%s' | base64 -d >> \"%s\"", encoded, skillPath)
		if _, err := r.sb.ExecuteSync(ctx, cmd); err != nil {
			return fmt.Errorf("failed to update SKILL.md for tool %s: %w", tool.Name, err)
		}
	}

	// 追加 trace 上下文提示到 SKILL.md（引导 LLM 在 curl 调用中注入 trace 头）
	if r.opt.TraceContext != nil {
		if _, ok := r.opt.TraceContext["traceparent"]; ok {
			skillDir := ".opencode/skills/" + skillName
			skillPath := fmt.Sprintf("%s/SKILL.md", skillDir)
			traceContent := formatTraceContextContent(r.opt.TraceContext)
			encoded := base64.StdEncoding.EncodeToString([]byte(traceContent))
			cmd := fmt.Sprintf("echo '%s' | base64 -d >> \"%s\"", encoded, skillPath)
			if _, err := r.sb.ExecuteSync(ctx, cmd); err != nil {
				return fmt.Errorf("failed to update SKILL.md trace context for tool %s: %w", tool.Name, err)
			}
		}
	}
	return nil
}

// copyOutput 复制输出文件到本地，并移除隐藏文件。
//
// 替换策略：先复制到临时目录并清理隐藏文件，再用 backup+swap 原子替换 OutputDir。
// 这样可以确保沙箱中已删除/重命名的文件不会在 OutputDir 中残留，
// 同时保证任何失败都不会破坏 OutputDir 的现有内容（workspace 链不断）。
func (r *Runner) copyOutput(ctx context.Context) error {
	outputDir := r.opt.OutputDir
	tmpDir := filepath.Join(filepath.Dir(outputDir), ".swaptmp-"+filepath.Base(outputDir))
	bakDir := filepath.Join(filepath.Dir(outputDir), ".swapbak-"+filepath.Base(outputDir))

	// 清理上次失败留下的临时/备份目录
	_ = os.RemoveAll(tmpDir)
	_ = os.RemoveAll(bakDir)

	// 1. 复制沙箱内容到临时目录（失败 → OutputDir 不变，链安全）
	if err := r.sb.CopyFromSandbox(ctx, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to copy output from workspace: %w", err)
	}
	// 确保 tmpDir 存在（CopyFromSandbox 在沙箱只有隐藏文件时会跳过）
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 2. 清理临时目录中的隐藏文件
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to read temp directory: %w", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			removePath := filepath.Join(tmpDir, entry.Name())
			if err := os.RemoveAll(removePath); err != nil {
				_ = os.RemoveAll(tmpDir)
				return fmt.Errorf("failed to remove hidden file %s: %w", entry.Name(), err)
			}
		}
	}

	// 3. 原子替换：备份 OutputDir → 安装 tmpDir → 清理备份
	//    （失败 → 从备份恢复，链安全）
	if err := swapDirWithBackup(tmpDir, outputDir, bakDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to swap output directory: %w", err)
	}

	return nil
}

// ============================================================================
// Session 管理
// ============================================================================

// createSession 通过 API 创建 opencode session。
func (r *Runner) createSession(ctx context.Context) error {
	var result struct {
		ID string `json:"id"`
	}
	resp, err := trace_util.NewResty(ctx).R().
		SetContext(ctx).
		SetQueryParam("directory", r.sb.WorkDir()).
		SetBody(map[string]any{}).
		SetResult(&result).
		Post(r.opt.Sandbox.OpencodeEndpoint() + "/session")
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	if resp.StatusCode() >= 300 {
		return fmt.Errorf("create session failed: [%d] %s", resp.StatusCode(), resp.String())
	}
	r.sessionID = result.ID
	return nil
}

// deleteSession 通过 API 删除 opencode session。
func (r *Runner) deleteSession(ctx context.Context) {
	if r.sessionID == "" {
		return
	}
	resp, err := trace_util.NewResty(ctx).R().
		SetContext(ctx).
		SetQueryParam("directory", r.sb.WorkDir()).
		Delete(fmt.Sprintf("%s/session/%s", r.opt.Sandbox.OpencodeEndpoint(), r.sessionID))
	if err != nil {
		log.Warnf("%s failed to delete session %s: %v", r.logPrefix, r.sessionID, err)
		return
	}
	if resp.StatusCode() >= 300 {
		log.Warnf("%s delete session %s failed: [%d] %s", r.logPrefix, r.sessionID, resp.StatusCode(), resp.String())
	}
}

// ============================================================================
// SSE 连接
// ============================================================================

// connectSSE 连接到 opencode 全局事件流。
func (r *Runner) connectSSE(ctx context.Context) (<-chan string, error) {
	sseCh := make(chan string, 1024)
	errCh := make(chan error, 1)
	connected := make(chan struct{})

	go func() {
		defer util.PrintPanicStack()
		defer close(sseCh)
		defer close(errCh)

		resp, err := trace_util.NewResty(ctx).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			SetTimeout(0).
			R().
			SetContext(ctx).
			SetHeader("Accept", "text/event-stream").
			SetDoNotParseResponse(true).
			Get(r.opt.Sandbox.OpencodeEndpoint() + "/global/event")
		if err != nil {
			errCh <- fmt.Errorf("SSE connect failed: %w", err)
			return
		}
		defer func() {
			if resp != nil && resp.RawResponse != nil {
				_ = resp.RawResponse.Body.Close()
			}
		}()

		// context 已取消，直接返回
		select {
		case <-ctx.Done():
			return
		default:
		}

		if resp.StatusCode() >= 300 {
			b, _ := io.ReadAll(resp.RawResponse.Body)
			errCh <- fmt.Errorf("SSE connect failed: [%d] %s", resp.StatusCode(), string(b))
			return
		}

		// 连接成功，通知主 goroutine
		close(connected)
		r.readSSEStream(resp.RawResponse.Body, sseCh, ctx)
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-connected:
		return sseCh, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// readSSEStream 读取 SSE 流，提取 data 字段并发送到通道。
func (r *Runner) readSSEStream(body io.ReadCloser, sseCh chan<- string, ctx context.Context) {
	scanner := util.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			select {
			case sseCh <- data:
			case <-ctx.Done():
				return
			}
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF && err != context.Canceled {
		log.Warnf("%s SSE stream error: %v", r.logPrefix, err)
	}
}

// processSSEEvents 处理 SSE 事件流，过滤并转换为 JSON 输出。
func (r *Runner) processSSEEvents(ctx context.Context, sseCh <-chan string, outputCh chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-sseCh:
			if !ok {
				return
			}
			line, isIdle := r.handleEvent(data)
			if line != "" {
				select {
				case outputCh <- line:
				case <-ctx.Done():
					return
				}
			}
			if isIdle {
				return
			}
		}
	}
}

// ============================================================================
// 提示词
// ============================================================================

// sendPromptAsync 异步发送提示词到 opencode session。
func (r *Runner) sendPromptAsync(ctx context.Context) error {
	system, prompt, err := r.buildSystemAndPrompt()
	if err != nil {
		return fmt.Errorf("failed to build system and prompt: %w", err)
	}

	reqBody := map[string]any{
		"parts": []map[string]any{
			{"type": "text", "text": prompt},
		},
	}
	if system != "" {
		reqBody["system"] = system
	}

	resp, err := trace_util.NewResty(ctx).R().
		SetContext(ctx).
		SetQueryParam("directory", r.sb.WorkDir()).
		SetBody(reqBody).
		Post(fmt.Sprintf("%s/session/%s/prompt_async", r.opt.Sandbox.OpencodeEndpoint(), r.sessionID))
	if err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}
	if resp.StatusCode() >= 300 && resp.StatusCode() != 204 {
		return fmt.Errorf("send prompt failed: [%d] %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// buildSystemAndPrompt 构建系统提示词和用户提示词。
// 当 SystemMessageStrategy 为 merge 时，提取 messages 中的 system 消息并合并到 instruction 中。
// historyMessages: 倒序找到第一条 role=user 之前的所有消息
// prompt: 倒序找到第一条 role=user 消息，将这条及之后的所有消息拼接成 prompt
func (r *Runner) buildSystemAndPrompt() (system string, prompt string, err error) {
	messages := r.opt.Messages

	// 提取 system 消息到 instruction（仅 merge 策略）
	var extraSystemContent string
	var otherMessages []adk.Message
	if r.opt.SystemMessageStrategy == wga_sandbox_option.SystemMessageStrategyMerge {
		for _, msg := range messages {
			if msg.Role == schema.System && msg.Content != "" {
				if extraSystemContent != "" {
					extraSystemContent += "\n\n"
				}
				extraSystemContent += msg.Content
			} else {
				otherMessages = append(otherMessages, msg)
			}
		}
	} else {
		otherMessages = messages
	}

	instruction := r.opt.Instruction
	if extraSystemContent != "" {
		if instruction != "" {
			instruction += "\n\n"
		}
		instruction += extraSystemContent
	}

	// 从后往前找到第一条 role=user 的消息
	firstUserIndex := -1
	for i := len(otherMessages) - 1; i >= 0; i-- {
		if otherMessages[i].Role == schema.User {
			firstUserIndex = i
			break
		}
	}

	var historyMessages []adk.Message
	var promptMessages []adk.Message

	if firstUserIndex >= 0 {
		// historyMessages: 第一条 user 之前的消息
		if firstUserIndex > 0 {
			historyMessages = otherMessages[:firstUserIndex]
		}
		// promptMessages: 第一条 user 及之后的消息
		promptMessages = otherMessages[firstUserIndex:]
	} else {
		// 没有 user 消息，全部作为 prompt
		promptMessages = otherMessages
	}

	system, err = renderSystem(instruction, r.opt.OverallTask, historyMessages)
	if err != nil {
		return "", "", err
	}

	// 拼接 promptMessages
	for i, msg := range promptMessages {
		if i > 0 {
			prompt += "\n\n"
		}
		if len(msg.UserInputMultiContent) > 0 {
			b, _ := json.Marshal(msg.UserInputMultiContent)
			prompt += string(b)
		} else {
			prompt += msg.Content
		}
	}

	return system, prompt, nil
}

// ============================================================================
// 事件转换
// ============================================================================

// handleEvent 处理 SSE 事件，返回 (输出内容, 是否会话结束)。
func (r *Runner) handleEvent(data string) (string, bool) {
	var event sseEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return "", false
	}

	// 过滤其他沙箱的事件
	if event.Directory != r.sb.WorkDir() {
		return "", false
	}

	// 区分 BusEvent 和 SyncEvent
	if event.Payload.Type == "sync" && event.Payload.SyncEvent != nil {
		return r.handleSyncEvent(&event)
	}

	return r.handleBusEvent(&event)
}

// handleBusEvent 处理 BusEvent 类型的事件（对应 opencode/src/bus/bus-event.ts）。
// BusEvent 用于实时增量推送，包括：
// - message.part.delta: 流式文本/推理增量
// - session.idle: 会话空闲
// - session.error: 会话错误
func (r *Runner) handleBusEvent(event *sseEvent) (string, bool) {
	switch event.Payload.Type {
	case "message.part.delta":
		// 流式增量事件（对应 MessageV2.Event.PartDelta）
		if event.Payload.Properties.SessionID != r.sessionID {
			return "", false
		}
		return r.convertDeltaEvent(event), false
	case "session.idle":
		// 会话空闲事件（对应 SessionStatus.Event.Idle）
		// session.idle 已标记 deprecated（仍向后兼容发送），同时发送 session.status。
		// 若后续版本停止发送 session.idle，需切换至 session.status：
		//   收到 session.status 且 status.type == "idle" 时视作会话结束
		if event.Payload.Properties.SessionID != r.sessionID {
			return "", false
		}
		return "", true
	case "session.error":
		// 会话错误事件（对应 Session.Event.Error）
		if event.Payload.Properties.SessionID != r.sessionID {
			return "", false
		}
		return r.convertErrorEvent(event), false
	case string(OpencodeEventTypeQuestionAsked):
		// 问题提出事件（Human-in-the-Loop）
		return r.convertQuestionEvent(event, OpencodeEventTypeQuestionAsked, "pending"), false
	case string(OpencodeEventTypeQuestionReplied):
		// 问题已回答事件（Human-in-the-Loop）
		return r.convertQuestionEvent(event, OpencodeEventTypeQuestionReplied, "answered"), false
	case string(OpencodeEventTypeQuestionRejected):
		// 问题被拒绝事件（Human-in-the-Loop）
		return r.convertQuestionEvent(event, OpencodeEventTypeQuestionRejected, "rejected"), false
	default:
		return "", false
	}
}

// handleSyncEvent 处理 SyncEvent 类型的事件（对应 opencode/src/sync/index.ts）。
// SyncEvent 用于状态变更通知，包括：
// - message.updated.1: 消息创建/更新
// - message.part.updated.1: Part 状态更新
func (r *Runner) handleSyncEvent(event *sseEvent) (string, bool) {
	syncEvent := event.Payload.SyncEvent
	if syncEvent == nil {
		return "", false
	}

	switch syncEvent.Type {
	case "message.updated.1":
		// 消息更新事件（对应 MessageV2.Event.Updated）
		r.trackUserMessageIDFromSyncEvent(syncEvent)
		return "", false
	case "message.part.updated.1":
		// Part 更新事件（对应 MessageV2.Event.PartUpdated）
		if syncEvent.Data.SessionID != r.sessionID {
			return "", false
		}
		return r.convertMessagePartEventFromSync(syncEvent), false
	default:
		return "", false
	}
}

// trackUserMessageIDFromSyncEvent 从 message.updated 事件中记录用户消息 ID。
// 用于过滤用户消息的 delta 输出（用户消息不需要输出）。
func (r *Runner) trackUserMessageIDFromSyncEvent(syncEvent *sseSyncEvent) {
	info := syncEvent.Data.Info
	if info != nil && info.Role == "user" && info.ID != "" {
		r.userMsgIDs[info.ID] = true
	}
}

// convertDeltaEvent 转换 message.part.delta 事件（对应 MessageV2.Event.PartDelta）。
// delta 事件用于流式传输文本和推理内容的增量，只携带 partID 而不携带 part 类型，
// 需要通过 partTypeTracker 查找类型（从 message.part.updated 事件中记录）。
func (r *Runner) convertDeltaEvent(event *sseEvent) string {
	props := event.Payload.Properties
	delta := props.Delta
	if delta == "" {
		return ""
	}

	// 通过 messageID 判断是否为用户消息的 delta（用户消息不需要输出）
	if r.userMsgIDs[props.MessageID] {
		return ""
	}

	// 根据 part 类型判断是 text 还是 reasoning
	partType := r.partTypeTracker.get(props.PartID)
	switch partType {
	case "reasoning":
		if !r.opt.EnableThinking {
			return ""
		}
		return r.buildReasoningEvent(delta)
	default:
		// 未知 part 类型或 text 类型，作为 text 处理
		return r.buildTextEvent(delta)
	}
}

// convertMessagePartEventFromSync 转换 message.part.updated 事件（对应 MessageV2.Event.PartUpdated）。
// PartUpdated 事件用于通知 Part 状态变更，包括：
// - step-start/step-finish: 步骤标记
// - text/reasoning: 跟踪 part 类型供 delta 事件使用
// - tool: 工具调用状态变更
func (r *Runner) convertMessagePartEventFromSync(syncEvent *sseSyncEvent) string {
	part := syncEvent.Data.Part
	if part == nil {
		return ""
	}

	if r.userMsgIDs[part.MessageID] {
		return ""
	}

	switch part.Type {
	case "step-start":
		// 跟踪当前 step 的 part 列表
		r.partTypeTracker.stepStart(part.ID)
		return r.convertStepStartEvent(part)
	case "step-finish":
		r.partTypeTracker.stepFinish()
		return r.convertStepFinishEvent(part)
	case "text":
		// 跟踪 part 类型，供 delta 事件使用
		r.partTypeTracker.set(part.ID, "text")
		// SyncEvent 中 text part.updated 不携带增量文本
		return ""
	case "reasoning":
		// 跟踪 part 类型，供 delta 事件使用
		r.partTypeTracker.set(part.ID, "reasoning")
		return ""
	case "tool":
		return r.convertToolEvent(part)
	default:
		return ""
	}
}

// convertStepStartEvent 转换步骤开始事件。
func (r *Runner) convertStepStartEvent(part *sseEventPart) string {
	event := OpencodeEvent{
		Type:      OpencodeEventTypeStepStart,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}
	stepP := stepStartPart{Type: "step_start"}
	event.Part, _ = json.Marshal(stepP)

	data, _ := json.Marshal(event)
	return string(data)
}

// convertStepFinishEvent 转换步骤结束事件。
func (r *Runner) convertStepFinishEvent(part *sseEventPart) string {
	event := OpencodeEvent{
		Type:      OpencodeEventTypeStepFinish,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}
	finishP := stepFinishPart{
		Type:   "step_finish",
		Reason: part.Reason,
		Tokens: stepFinishPartTokens{
			Input:     part.Tokens.Input,
			Output:    part.Tokens.Output,
			Reasoning: part.Tokens.Reasoning,
			Cache: struct {
				Read  float64 `json:"read,omitempty"`
				Write float64 `json:"write,omitempty"`
			}{
				Read:  part.Tokens.Cache.Read,
				Write: part.Tokens.Cache.Write,
			},
		},
		Cost: part.Cost,
	}
	event.Part, _ = json.Marshal(finishP)

	data, _ := json.Marshal(event)
	return string(data)
}

// buildTextEvent 构建文本增量事件。
func (r *Runner) buildTextEvent(delta string) string {
	event := OpencodeEvent{
		Type:      OpencodeEventTypeText,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}
	textP := textPart{Type: "text", Text: delta}
	event.Part, _ = json.Marshal(textP)

	data, _ := json.Marshal(event)
	return string(data)
}

// buildReasoningEvent 构建推理增量事件。
func (r *Runner) buildReasoningEvent(delta string) string {
	event := OpencodeEvent{
		Type:      OpencodeEventTypeReasoning,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}
	reasoningP := reasoningPart{Type: "reasoning", Text: delta}
	event.Part, _ = json.Marshal(reasoningP)

	data, _ := json.Marshal(event)
	return string(data)
}

// convertToolEvent 转换工具调用事件。
// 只发送 completed 或 error 状态的事件。
func (r *Runner) convertToolEvent(part *sseEventPart) string {
	if part.State.Status != "completed" && part.State.Status != "error" {
		return ""
	}

	callID := part.CallID
	if callID == "" {
		callID = part.ID
	}

	event := OpencodeEvent{
		Type:      OpencodeEventTypeToolUse,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}

	toolP := toolPart{
		Type:   "tool_use",
		CallID: callID,
		Tool:   part.Tool,
		State: toolState{
			Status: part.State.Status,
			Input:  part.State.Input,
			Output: part.State.Output,
			Error:  part.State.Error,
		},
	}
	event.Part, _ = json.Marshal(toolP)

	data, _ := json.Marshal(event)
	return string(data)
}

// convertErrorEvent 转换错误事件。
func (r *Runner) convertErrorEvent(event *sseEvent) string {
	errInfo := event.Payload.Properties.Error
	evt := OpencodeEvent{
		Type:      OpencodeEventTypeError,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}
	errorP := errorPart{}
	errorP.Error.Name = errInfo.Name
	errorP.Error.Data.Message = errInfo.Data.Message
	evt.Part, _ = json.Marshal(errorP)

	data, _ := json.Marshal(evt)
	return string(data)
}

// convertQuestionEvent 转换问题事件（Human-in-the-Loop）。
func (r *Runner) convertQuestionEvent(event *sseEvent, eventType OpencodeEventType, status string) string {
	evt := OpencodeEvent{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		SessionID: r.sessionID,
	}

	props := event.Payload.Properties
	questionID := props.ID
	if questionID == "" {
		questionID = props.RequestID
	}

	questionP := questionPart{
		Type:       "question",
		QuestionID: questionID,
		SessionID:  props.SessionID,
		Status:     status,
		Questions:  make([]questionItem, 0, len(props.Questions)),
	}

	for _, q := range props.Questions {
		custom := r.opt.EnableHumanInTheLoopCustom
		if q.Custom != nil {
			custom = *q.Custom
		}
		item := questionItem{
			Question: q.Question,
			Header:   q.Header,
			Multiple: q.Multiple,
			Custom:   custom,
			Options:  make([]questionOption, 0, len(q.Options)),
		}
		for _, opt := range q.Options {
			item.Options = append(item.Options, questionOption(opt))
		}
		questionP.Questions = append(questionP.Questions, item)
	}

	if len(props.Answers) > 0 {
		questionP.Answers = props.Answers
	}

	evt.Part, _ = json.Marshal(questionP)
	data, _ := json.Marshal(evt)

	return string(data)
}

// ============================================================================
// 模板渲染
// ============================================================================

// processedMCP 渲染用的 MCP 中间结构体，在渲染前统一完成鉴权处理。
type processedMCP struct {
	Name        string
	URL         string
	Description string
	Headers     map[string]string // 最终注入到 opencode.json 的请求头
}

// renderConfig 渲染 opencode 配置文件。
func renderConfig(config wga_sandbox_option.ModelConfig, mcps []wga_sandbox_option.MCP, enableHITL bool) (string, error) {
	// 在渲染前统一处理鉴权：query 鉴权拼入 URL，header 鉴权转为 headers map
	processedMcps := make([]processedMCP, 0, len(mcps))
	for _, mcp := range mcps {
		pm := processedMCP{
			Name:        mcp.Name,
			Description: strings.Join(strings.Fields(mcp.Description), " "),
			Headers:     make(map[string]string),
		}
		// 自定义请求头优先注入
		for k, v := range mcp.Headers {
			pm.Headers[k] = v
		}
		// 处理 API 鉴权
		if mcp.ApiAuth != nil && mcp.ApiAuth.AuthType != "" && mcp.ApiAuth.AuthType != util.AuthTypeNone {
			switch mcp.ApiAuth.AuthType {
			case util.AuthTypeAPIKeyQuery:
				// query 鉴权拼入 URL
				rawUrl, err := url.Parse(mcp.URL)
				if err != nil {
					return "", fmt.Errorf("mcp [%s] parse url err: %w", mcp.Name, err)
				}
				q := rawUrl.Query()
				q.Set(mcp.ApiAuth.ApiKeyQueryParam, mcp.ApiAuth.ApiKeyValue)
				rawUrl.RawQuery = q.Encode()
				pm.URL = rawUrl.String()
			case util.AuthTypeAPIKeyHeader:
				// header 鉴权转为请求头
				value := mcp.ApiAuth.ApiKeyValue
				switch mcp.ApiAuth.ApiKeyHeaderPrefix {
				case util.ApiKeyHeaderPrefixBasic:
					value = "Basic " + mcp.ApiAuth.ApiKeyValue
				case util.ApiKeyHeaderPrefixBearer:
					value = "Bearer " + mcp.ApiAuth.ApiKeyValue
				}
				headerName := mcp.ApiAuth.ApiKeyHeader
				if headerName == "" {
					headerName = util.ApiKeyHeaderDefault
				}
				pm.Headers[headerName] = value
			}
		}
		// 未设置 URL 的情况（无 query 鉴权）
		if pm.URL == "" {
			pm.URL = mcp.URL
		}
		// 没有 headers 则设为 nil，避免模板渲染空的 headers 块
		if len(pm.Headers) == 0 {
			pm.Headers = nil
		}
		processedMcps = append(processedMcps, pm)
	}

	tmpl, err := template.New("config").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"len": func(v interface{}) int {
			switch s := v.(type) {
			case []processedMCP:
				return len(s)
			}
			return 0
		},
	}).Parse(configTemplate)
	if err != nil {
		return "", fmt.Errorf("parse config template failed: %w", err)
	}
	data := struct {
		wga_sandbox_option.ModelConfig
		MCPs                 []processedMCP
		EnableHumanInTheLoop bool
	}{
		ModelConfig:          config,
		MCPs:                 processedMcps,
		EnableHumanInTheLoop: enableHITL,
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute config template failed: %w", err)
	}
	return buf.String(), nil
}

// renderSystem 渲染系统提示词模板。
func renderSystem(instruction, overallTask string, messages []adk.Message) (string, error) {
	tmpl, err := template.New("system").Funcs(template.FuncMap{
		"formatPart": formatMessageInputPart,
		"listParts":  listMessageInputParts,
	}).Parse(systemTemplate)
	if err != nil {
		return "", fmt.Errorf("parse system template failed: %w", err)
	}

	data := struct {
		Instruction string
		OverallTask string
		Messages    []adk.Message
	}{
		Instruction: instruction,
		OverallTask: overallTask,
		Messages:    messages,
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute system template failed: %w", err)
	}
	return buf.String(), nil
}

// listMessageInputParts 将消息部分列表格式化为字符串列表，过滤空结果。
func listMessageInputParts(parts []schema.MessageInputPart) []string {
	var result []string
	for _, part := range parts {
		if s := formatMessageInputPart(part); s != "" {
			result = append(result, s)
		}
	}
	return result
}

// formatMessageInputPart 格式化消息部分为文本表示。
// text: 直接输出文本内容
// image_url/audio_url/video_url/file_url: 输出 URL
func formatMessageInputPart(part schema.MessageInputPart) string {
	switch part.Type {
	case schema.ChatMessagePartTypeText:
		return part.Text
	case schema.ChatMessagePartTypeImageURL:
		if part.Image != nil && part.Image.URL != nil {
			return fmt.Sprintf("[image](%s)", *part.Image.URL)
		}
	case schema.ChatMessagePartTypeAudioURL:
		if part.Audio != nil && part.Audio.URL != nil {
			return fmt.Sprintf("[audio](%s)", *part.Audio.URL)
		}
	case schema.ChatMessagePartTypeVideoURL:
		if part.Video != nil && part.Video.URL != nil {
			return fmt.Sprintf("[video](%s)", *part.Video.URL)
		}
	case schema.ChatMessagePartTypeFileURL:
		if part.File != nil && part.File.URL != nil {
			name := "file"
			if part.File.Name != "" {
				name = part.File.Name
			}
			return fmt.Sprintf("[%s](%s)", name, *part.File.URL)
		}
	}
	return ""
}

// ============================================================================
// 工具函数
// ============================================================================

// toSkillName 将工具名称转换为 skill 名称，替换空格为连字符，移除括号，并转为小写以与 openapi2skill 保持一致。
func toSkillName(name string) string {
	result := strings.ReplaceAll(name, " ", "-")
	result = strings.ReplaceAll(result, "(", "")
	result = strings.ReplaceAll(result, ")", "")
	return strings.ToLower(result)
}

// formatAuthContent 格式化认证信息为 Markdown 格式。
func formatAuthContent(auth *openapi3_util.Auth) string {
	if auth.Type == "none" || auth.Value == "" {
		return ""
	}
	var authDesc string
	switch auth.In {
	case "header":
		authDesc = fmt.Sprintf("Header: `%s: %s`", auth.Name, auth.Value)
	case "query":
		authDesc = fmt.Sprintf("Query Parameter: `%s=%s`", auth.Name, auth.Value)
	default:
		authDesc = fmt.Sprintf("Auth Value: `%s`", auth.Value)
	}
	return fmt.Sprintf("\n## API Key\n\n%s\n", authDesc)
}

// formatVariablesContent 格式化变量信息为 Markdown 格式。
func formatVariablesContent(variables []wga_sandbox_option.SkillVariable) string {
	var buf strings.Builder
	buf.WriteString("\n## Variables\n\n")
	buf.WriteString("The following variables are configured for this skill:\n\n")
	for _, v := range variables {
		// 转义反引号
		escapedKey := strings.ReplaceAll(v.VariableKey, "`", "\\`")
		escapedValue := strings.ReplaceAll(v.VariableValue, "`", "\\`")

		if v.Description != "" {
			escapedDesc := strings.ReplaceAll(v.Description, "`", "\\`")
			fmt.Fprintf(&buf, "- **%s** (%s): `%s` = `%s`\n",
				v.Name, escapedDesc, escapedKey, escapedValue)
		} else {
			fmt.Fprintf(&buf, "- **%s**: `%s` = `%s`\n",
				v.Name, escapedKey, escapedValue)
		}
	}
	buf.WriteString("\n")
	return buf.String()
}

// formatTraceContextContent 格式化 trace 上下文信息为 Markdown 格式。
// 追加到 SKILL.md 末尾，引导 LLM 在构造 curl 命令时注入 trace 头。
func formatTraceContextContent(traceCtx map[string]string) string {
	var buf strings.Builder
	buf.WriteString("\n## Trace Context\n\n")
	buf.WriteString("**You MUST include the following headers in every HTTP request to any external API to maintain distributed tracing. Do NOT omit them:**\n")
	if tp, ok := traceCtx["traceparent"]; ok {
		fmt.Fprintf(&buf, "- `traceparent: %s`\n", tp)
	}
	if ts, ok := traceCtx["tracestate"]; ok && ts != "" {
		fmt.Fprintf(&buf, "- `tracestate: %s`\n", ts)
	}
	if bg, ok := traceCtx["baggage"]; ok && bg != "" {
		fmt.Fprintf(&buf, "- `baggage: %s`\n", bg)
	}
	buf.WriteString("\n")
	buf.WriteString("Example curl command with trace headers:\n")
	buf.WriteString("```bash\n")
	// 直接使用具体的 trace 值（不再依赖环境变量）
	curlArgs := []string{`curl`}
	if tp, ok := traceCtx["traceparent"]; ok {
		curlArgs = append(curlArgs, fmt.Sprintf(`-H "traceparent: %s"`, tp))
	}
	if ts, ok := traceCtx["tracestate"]; ok && ts != "" {
		curlArgs = append(curlArgs, fmt.Sprintf(`-H "tracestate: %s"`, ts))
	}
	if bg, ok := traceCtx["baggage"]; ok && bg != "" {
		curlArgs = append(curlArgs, fmt.Sprintf(`-H "baggage: %s"`, bg))
	}
	curlArgs = append(curlArgs, "...")
	buf.WriteString(strings.Join(curlArgs, " "))
	buf.WriteString("\n```\n")
	return buf.String()
}

// swapDirWithBackup 用备份目录实现原子替换 targetDir 为 tmpDir。
//
// 流程：
//  1. Rename targetDir → bakDir（备份当前内容）
//  2. Rename tmpDir → targetDir（安装新内容）
//  3. 如果步骤 2 失败，恢复 bakDir → targetDir（尽力恢复）
//  4. 删除 bakDir（清理）
//
// 三个目录必须在同一文件系统上，os.Rename 才能工作。
// 如果 targetDir 不存在，函数返回错误，不修改 tmpDir。
//
// 安全保证：任意步骤失败时，targetDir 要么包含原始内容（从备份恢复），
// 要么从未被修改。tmpDir 在失败时不会丢失。
func swapDirWithBackup(tmpDir, targetDir, bakDir string) error {
	if err := os.Rename(targetDir, bakDir); err != nil {
		return fmt.Errorf("failed to backup %s to %s: %w", targetDir, bakDir, err)
	}

	if err := os.Rename(tmpDir, targetDir); err != nil {
		_ = os.Rename(bakDir, targetDir)
		return fmt.Errorf("failed to install %s to %s: %w", tmpDir, targetDir, err)
	}

	_ = os.RemoveAll(bakDir)
	return nil
}
