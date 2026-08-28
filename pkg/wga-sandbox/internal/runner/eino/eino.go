// Package eino 提供 eino-agent 智能体的运行器实现（基于 HTTP API）。
//
// Runner 在沙箱内驱动一个独立的 eino-agent HTTP 服务（见 ./agent/），
// 通过 SSE 接收事件流并转发给上层调用方。生命周期分三段：
//
//   - BeforeRun: 在沙箱内准备 .env、skills/、input/、output/、tmp/ 目录。
//   - Run:       打开 SSE 连接，转发事件并保证最终 emit 一条 assistant+stop 兜底消息。
//   - AfterRun:  把 output/ 复制回宿主机并清理临时文件。
package eino

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

// 确保 Runner 实现 runner.Runner 接口。
var _ runner.Runner = (*Runner)(nil)

// Runner 是 eino-agent 智能体在 wga-sandbox 中的运行器。
type Runner struct {
	sb        sandbox.Sandbox
	req       wga_sandbox_option.RunOption
	agentType string
	logPrefix string
}

// NewRunner 创建 eino-agent 运行器实例。
func NewRunner(sb sandbox.Sandbox, req wga_sandbox_option.RunOption, agentType string) runner.Runner {
	return &Runner{
		sb:        sb,
		req:       req,
		agentType: agentType,
		logPrefix: fmt.Sprintf("[wga-sandbox][%s]", req.RunSession.RunID),
	}
}

// BeforeRun 在沙箱内准备 .env、skills/、input/、output/、tmp/ 目录。
func (r *Runner) BeforeRun(ctx context.Context) error {
	ctx = r.withTraceHeaders(ctx)

	log.Infof("%s BeforeRun - skills=%d inputDir=%s outputDir=%s workDir=%s",
		r.logPrefix, len(r.req.Skills), r.req.InputDir, r.req.OutputDir, r.sb.WorkDir())
	for i, skill := range r.req.Skills {
		log.Infof("%s BeforeRun - skill[%d]: Dir=%s", r.logPrefix, i, skill.Dir)
	}

	if err := r.setupEnv(ctx); err != nil {
		return err
	}
	if err := r.setupWorkspaceDirs(ctx); err != nil {
		return err
	}
	// setupWorkspaceDirs 必须先把 skill 目录 copy 进 workspace/skills/，
	// 之后才能向各 skill 的 SKILL.md 追加 Variables 段。
	if err := r.injectEnvVariables(ctx); err != nil {
		return err
	}
	if err := r.setupSkillVariables(ctx); err != nil {
		return err
	}
	return nil
}

// Run 通过 SSE 连接 eino-agent HTTP 服务的 /chat 端点，转发事件并保证最终
// 在 outputCh 上 emit 一条 assistant+stop 兜底消息。
func (r *Runner) Run(ctx context.Context) (<-chan string, error) {
	ctx = r.withTraceHeaders(ctx)

	sseCh, streamErr, err := r.connectSSE(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect SSE: %w", err)
	}

	outputCh := make(chan string, 1024)

	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)

		// sawFinal 与 fatalErr 通过指针出参传入 forwardSSEStream，
		// 因为下面的兜底 defer 必须在事件循环结束后读取它们的最终值。
		sawFinal := false
		var fatalErr error

		defer func() {
			if sawFinal {
				return
			}
			var err error
			if rec := recover(); rec != nil {
				err = fmt.Errorf("panic: %v", rec)
			} else {
				err = fatalErr
				if err == nil && ctx.Err() != nil {
					err = ctx.Err()
				}
			}
			line := buildFinalSSELine(err)
			if line == "" {
				return
			}
			// 兜底消息可能因 ctx cancel 时 wga-sandbox 转发层无人消费而被丢弃，
			// 这里先打 Warn 日志确保 沙箱调用方 容器日志一定能看到兜底原因。
			log.Warnf("%s final fallback constructed: %v", r.logPrefix, err)
			select {
			case outputCh <- line:
			default:
				log.Warnf("%s drop final fallback line: outputCh full", r.logPrefix)
			}
		}()

		r.forwardSSEStream(ctx, sseCh, streamErr, outputCh, &sawFinal, &fatalErr)
	}()

	return outputCh, nil
}

// AfterRun 把 output/ 复制回宿主机（如指定了 OutputDir）。
func (r *Runner) AfterRun(ctx context.Context) error {
	log.Infof("%s AfterRun start, OutputDir: %s", r.logPrefix, r.req.OutputDir)

	if r.req.OutputDir == "" {
		log.Infof("%s AfterRun skipped (no OutputDir)", r.logPrefix)
		return nil
	}

	if err := r.copyOutput(ctx); err != nil {
		log.Errorf("%s AfterRun failed: %v", r.logPrefix, err)
		return err
	}
	log.Infof("%s AfterRun completed", r.logPrefix)
	return nil
}

// withTraceHeaders 在已有 trace 上下文时注入 traceparent 头，
// 确保所有沙箱内 HTTP 调用都能传播 trace。
func (r *Runner) withTraceHeaders(ctx context.Context) context.Context {
	if r.req.TraceContext == nil {
		return ctx
	}
	return trace_util.InjectTraceHeaders(ctx, r.req.TraceContext)
}
