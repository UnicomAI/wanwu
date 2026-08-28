package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type bashArgs struct {
	Command string `json:"command"`
}

// NewBashTool 创建 bash 执行工具，封装安全拦截、退出码与输出截断。
func NewBashTool(backend *ShellOnlyBackend) (tool.BaseTool, error) {
	bashTool, err := utils.InferTool("bash",
		"执行 shell 命令并返回输出结果",
		func(ctx context.Context, input bashArgs) (string, error) {
			result, err := backend.Execute(ctx, &filesystem.ExecuteRequest{Command: input.Command})
			if err != nil {
				return "", err
			}
			return formatBashOutput(result), nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create bash tool: %w", err)
	}
	return bashTool, nil
}

func formatBashOutput(result *filesystem.ExecuteResponse) string {
	output := result.Output

	// 检测安全拦截，返回拦截信息让 agent 处理
	if strings.HasPrefix(output, "安全拦截：") {
		return output + "\n\n[系统提示] 检测到安全违规操作，请严格遵守安全规范"
	}

	if result.ExitCode != nil && *result.ExitCode != 0 {
		output += fmt.Sprintf("\n[命令执行失败，退出码: %d]", *result.ExitCode)
	}
	if result.Truncated {
		output += "\n[输出因大小限制被截断]"
	}
	// 命令执行成功且无任何输出时（如 mv、mkdir、touch），output 为空字符串。
	// 下游 go-openai ChatCompletionMessage.Content 的 json tag 带有 omitempty，
	// 空字符串会被序列化时省略，导致发给 LLM 的 tool message 缺少 content 字段，
	// 部分模型会因此误判工具未返回结果而重复调用或直接报错。补一个占位文本避免此问题，视情况再移除。
	if output == "" {
		output = "(命令执行完毕，无输出)"
	}
	return output
}

// NewSkillMiddleware 创建技能中间件（基于 LocalBackend）。
func NewSkillMiddleware(ctx context.Context, workspace string) (adk.AgentMiddleware, error) {
	backend, err := skill.NewLocalBackend(&skill.LocalBackendConfig{
		BaseDir: workspace + "/skills",
	})
	if err != nil {
		return adk.AgentMiddleware{}, fmt.Errorf("failed to create skill backend: %w", err)
	}

	skillMiddleware, err := skill.New(ctx, &skill.Config{
		Backend:    backend,
		UseChinese: true,
	})
	if err != nil {
		return adk.AgentMiddleware{}, fmt.Errorf("failed to create skill middleware: %w", err)
	}
	return skillMiddleware, nil
}

// NewBashMiddleware 把 bash 工具包装成 AgentMiddleware，
// 与 NewSkillMiddleware 并列在调用点装配，每个中间件自包含。
//
// halt 可空——为 nil 时不启用连续 BLOCKED 熔断（用于 tests + oneshot 沙箱兼容）。
func NewBashMiddleware(workspace string, halt *HaltState) (adk.AgentMiddleware, error) {
	backend := NewShellOnlyBackend(workspace, halt)
	bashTool, err := NewBashTool(backend)
	if err != nil {
		return adk.AgentMiddleware{}, err
	}
	return adk.AgentMiddleware{
		AdditionalTools: []tool.BaseTool{bashTool},
	}, nil
}
