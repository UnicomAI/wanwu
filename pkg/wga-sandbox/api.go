// Package wga_sandbox 提供沙箱容器交互功能，支持在隔离环境中执行智能体任务。
package wga_sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/opencode"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

var manager = sandbox.NewManager()

// Run 在沙箱容器中执行智能体任务。
func Run(ctx context.Context, opts ...wga_sandbox_option.Option) (wga_sandbox_option.RunSession, <-chan string, error) {
	var opt wga_sandbox_option.RunOption
	if err := opt.Apply(opts...); err != nil {
		return wga_sandbox_option.RunSession{}, nil, fmt.Errorf("apply options failed: %w", err)
	}

	runID := opt.RunSession.RunID
	if err := manager.Create(ctx, runID, opt.Sandbox); err != nil {
		return wga_sandbox_option.RunSession{}, nil, fmt.Errorf("create sandbox failed: %w", err)
	}

	logPrefix := fmt.Sprintf("[wga-sandbox][%s]", runID)
	if opt.AgentName != "" {
		logPrefix = fmt.Sprintf("[wga-sandbox][%s][%s]", runID, opt.AgentName)
	}

	var currentTask string
	if len(opt.Messages) > 0 {
		currentTask = opt.Messages[len(opt.Messages)-1].Content
	}
	log.Infof("%s %s", logPrefix, currentTask)

	sb, err := manager.Get(runID)
	if err != nil {
		return wga_sandbox_option.RunSession{}, nil, fmt.Errorf("get sandbox failed: %w", err)
	}
	r, err := createRunner(opt.RunnerType, sb, opt)
	if err != nil {
		return wga_sandbox_option.RunSession{}, nil, fmt.Errorf("create runner failed: %w", err)
	}
	log.Infof("%s using runner: %s", logPrefix, getRunnerName(opt.RunnerType))

	outputCh := make(chan string, 1024)

	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)

		if !opt.SkipCleanup {
			defer func() {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				_ = manager.Cleanup(cleanupCtx, runID)
			}()
		}

		log.Infof("%s preparing...", logPrefix)
		if err := r.BeforeRun(ctx); err != nil {
			log.Errorf("%s prepare failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("prepare failed: %v", err))
			return
		}

		log.Infof("%s running...", logPrefix)
		runnerOutputCh, err := r.Run(ctx)
		if err != nil {
			log.Errorf("%s run failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("run failed: %v", err))
			return
		}

		ctxCanceled := false
		for line := range runnerOutputCh {
			if ctxCanceled {
				break
			}
			select {
			case outputCh <- line:
			case <-ctx.Done():
				ctxCanceled = true
			}
		}

		log.Infof("%s finishing...", logPrefix)
		afterRunCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := r.AfterRun(afterRunCtx); err != nil {
			log.Errorf("%s finish failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("finish failed: %v", err))
			return
		}
		if opt.OutputDir != "" {
			log.Infof("%s output saved to: %s", logPrefix, opt.OutputDir)
		}
	}()

	return opt.RunSession, outputCh, nil
}

// Cleanup 清理指定 runID 的沙箱环境（用于 SkipCleanup=true 场景）。
func Cleanup(ctx context.Context, runID string) error {
	return manager.Cleanup(ctx, runID)
}

func sendErrorEvent(ch chan<- string, message string) {
	// 必须使用 OpencodeEvent 格式（含 part 字段），否则 opencodeConverter 解析时
	// event.Part 为 nil，json.Unmarshal(nil, ...) 会产生 "unexpected end of JSON input"
	evt := opencode.OpencodeEvent{
		Type:      opencode.OpencodeEventTypeError,
		Timestamp: time.Now().UnixMilli(),
	}
	errorP := opencode.ErrorPart{}
	errorP.Error.Name = "sandbox_error"
	errorP.Error.Data.Message = message
	evt.Part, _ = json.Marshal(errorP)
	data, err := json.Marshal(evt)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"type":"error","timestamp":%d,"sessionID":"","part":{"error":{"name":"sandbox_error","data":{"message":"%s"}}}}`, time.Now().UnixMilli(), message))
	}
	select {
	case ch <- string(data):
	default:
	}
}

func createRunner(t wga_sandbox_option.RunnerType, sb sandbox.Sandbox, opt wga_sandbox_option.RunOption) (runner.Runner, error) {
	switch t {
	case wga_sandbox_option.RunnerTypeEinoChatModel:
		return eino.NewRunner(sb, opt, "chat-model"), nil
	case wga_sandbox_option.RunnerTypeOpencode:
		return opencode.NewRunner(sb, opt), nil
	default:
		return nil, fmt.Errorf("unknown runner type: %s", t)
	}
}

func getRunnerName(t wga_sandbox_option.RunnerType) string {
	switch t {
	case wga_sandbox_option.RunnerTypeEinoChatModel:
		return "eino-chat-model (pkg/wga-sandbox/internal/runner/eino)"
	case wga_sandbox_option.RunnerTypeOpencode:
		return "opencode (pkg/wga-sandbox/internal/runner/opencode)"
	default:
		return "unknown runner"
	}
}
