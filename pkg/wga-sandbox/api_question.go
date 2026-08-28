package wga_sandbox

import (
	"context"
	"fmt"
	"net/url"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

// workspaceBase 与 pkg/wga-sandbox/internal/sandbox/sandbox.go 中的定义一致
const workspaceBase = "/home/root/workspace"

// ReplyQuestion 回答问题（Human-in-the-Loop）。
// 向 OpenCode 发送回答请求，解除 AI 阻塞等待。
// runID 用于计算 OpenCode 实例的 directory 参数。
func ReplyQuestion(ctx context.Context, cfg wga_sandbox_option.SandboxConfig, runID string, questionID string, answers [][]string) error {
	// directory 格式: /home/root/workspace/{runID}/workspace
	// 与 pkg/wga-sandbox/internal/sandbox/reuse.go 中 Prepare() 生成的 workDir 一致
	directory := fmt.Sprintf("%s/%s/workspace", workspaceBase, runID)
	urlStr := fmt.Sprintf("%s/question/%s/reply?directory=%s", cfg.OpencodeEndpoint(), questionID, url.QueryEscape(directory))
	_, err := trace_util.NewResty(ctx).R().
		SetContext(ctx).
		SetBody(map[string]interface{}{"answers": answers}).
		Post(urlStr)
	return err
}

// RejectQuestion 拒绝问题（Human-in-the-Loop）。
// 向 OpenCode 发送拒绝请求，AI 将收到 RejectedError。
// runID 用于计算 OpenCode 实例的 directory 参数。
func RejectQuestion(ctx context.Context, cfg wga_sandbox_option.SandboxConfig, runID string, questionID string) error {
	// directory 格式: /home/root/workspace/{runID}/workspace
	// 与 pkg/wga-sandbox/internal/sandbox/reuse.go 中 Prepare() 生成的 workDir 一致
	directory := fmt.Sprintf("%s/%s/workspace", workspaceBase, runID)
	urlStr := fmt.Sprintf("%s/question/%s/reject?directory=%s", cfg.OpencodeEndpoint(), questionID, url.QueryEscape(directory))
	_, err := trace_util.NewResty(ctx).R().
		SetContext(ctx).
		SetBody(map[string]interface{}{}).
		Post(urlStr)
	return err
}
