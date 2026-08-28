package factory

import (
	"context"

	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_converter "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-converter"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	"github.com/cloudwego/eino/adk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// sandboxAgent 在沙箱容器中执行的智能体。
type sandboxAgent struct {
	cfg     *config.Agent
	options option.Options
}

func newSandboxAgent(_ context.Context, cfg *config.Agent, options option.Options) (adk.Agent, error) {
	return &sandboxAgent{
		cfg:     cfg,
		options: options,
	}, nil
}

func (a *sandboxAgent) Name(_ context.Context) string {
	return a.cfg.ID
}

func (a *sandboxAgent) Description(_ context.Context) string {
	return a.cfg.Description
}

func (a *sandboxAgent) Run(ctx context.Context, agentInput *adk.AgentInput, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	// 创建 sandbox 执行 span
	ctx, span := trace_util.StartAgentSpan(ctx, "sandbox", a.cfg.Name, a.cfg.ID)
	span.SetAttributes(attribute.String("wga.agent.type", "sandbox"))

	sandboxOpts, err := a.buildSandboxOpts(ctx, agentInput.Messages)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return wga_sandbox_converter.ConvertToEinoIteratorWithError(ctx, wga_sandbox_option.RunnerTypeOpencode, err)
	}

	_, outputCh, err := wga_sandbox.Run(ctx, sandboxOpts...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return wga_sandbox_converter.ConvertToEinoIteratorWithError(ctx, wga_sandbox_option.RunnerTypeOpencode, err)
	}

	// 包装迭代器：在迭代结束时自动结束 sandbox span
	iter := wga_sandbox_converter.ConvertToEinoIterator(ctx, wga_sandbox_option.RunnerTypeOpencode, outputCh)
	return trace_util.WrapIteratorWithSpan(iter, span)
}

func (a *sandboxAgent) buildSandboxOpts(ctx context.Context, messages []adk.Message) ([]wga_sandbox_option.Option, error) {
	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
			ThreadID: a.options.RunSession.ThreadID,
			RunID:    a.options.RunSession.RunID,
		}),
		wga_sandbox_option.WithModelConfig(wga_sandbox_option.ModelConfig{
			Provider:     a.options.Model.Provider,
			ProviderName: a.options.Model.ProviderName,
			BaseURL:      a.options.Model.BaseURL,
			APIKey:       a.options.Model.APIKey,
			Model:        a.options.Model.Model,
			ModelName:    a.options.Model.ModelName,
			Params:       a.options.Model.Params,
		}),
		wga_sandbox_option.WithMessages(messages),
		wga_sandbox_option.WithEnableThinking(a.cfg.Configure.EnableThinking),
		wga_sandbox_option.WithEnableHumanInTheLoop(a.options.EnableHumanInTheLoop, a.options.EnableHumanInTheLoopCustom),
		wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
		wga_sandbox_option.WithAgentName(a.cfg.ID),
	}

	instruction, err := a.options.FormatInstruction(ctx, a.cfg)
	if err != nil {
		return nil, err
	}
	opts = append(opts, wga_sandbox_option.WithInstruction(instruction))

	if a.options.OverallTask != "" {
		opts = append(opts, wga_sandbox_option.WithOverallTask(a.options.OverallTask))
	}

	// 传递技能（配置文件 + 运行时）
	var allSkills []wga_sandbox_option.Skill
	for _, skill := range a.cfg.Skills {
		allSkills = append(allSkills, wga_sandbox_option.Skill{
			Dir:       skill.Dir,
			Variables: convertSkillVariables(skill.Variables),
		})
	}
	for _, skill := range a.options.Skills {
		allSkills = append(allSkills, wga_sandbox_option.Skill{
			Dir:       skill.Dir,
			Variables: convertSkillVariables(skill.Variables),
		})
	}
	if len(allSkills) > 0 {
		opts = append(opts, wga_sandbox_option.WithSkills(allSkills))
	}

	if a.options.Workspace.InputDir != "" {
		opts = append(opts, wga_sandbox_option.WithInputDir(a.options.Workspace.InputDir))
	}
	if a.options.Workspace.OutputDir != "" {
		opts = append(opts, wga_sandbox_option.WithOutputDir(a.options.Workspace.OutputDir))
	}

	if a.cfg.Configure.Sandbox != nil {
		cfg := a.cfg.Configure.Sandbox
		switch cfg.Type {
		case string(wga_sandbox_option.SandboxTypeOneshot):
			opts = append(opts, wga_sandbox_option.WithSandbox(
				wga_sandbox_option.SandboxOneshot(cfg.ImageName),
			))
		default:
			opts = append(opts, wga_sandbox_option.WithSandbox(
				wga_sandbox_option.SandboxReuse(cfg.Host),
			))
		}
	}

	var tools []wga_sandbox_option.Tool
	for _, tc := range a.cfg.ToolCategories {
		for _, toolCfg := range tc.Tools {
			var auth *openapi3_util.Auth
			var err error
			if toolCfg.AuthRequired {
				// 需要鉴权的工具必须在 options.Tools 中找到匹配项并获取认证信息；否则视为未配置该工具，不加入沙箱工具列表
				var matched bool
				for _, toolOpt := range a.options.Tools {
					if toolOpt.Title == toolCfg.Doc.Info.Title {
						if auth, err = toolOpt.APIAuth.ToOpenapiAuth(); err == nil && auth.Type != "none" {
							matched = true
						}
						break
					}
				}
				if !matched {
					continue
				}
			} // 无需鉴权的工具直接加入沙箱工具列表
			var operationIDs []string
			for _, op := range toolCfg.Operations {
				operationIDs = append(operationIDs, op.OperationID)
			}
			tools = append(tools, wga_sandbox_option.Tool{
				OpenAPI3Schema: toolCfg.Doc,
				OperationIDs:   operationIDs,
				APIAuth:        auth,
			})
		}
	}
	for _, extraTool := range a.options.ExtraTools {
		var auth *openapi3_util.Auth
		if extraTool.APIAuth != nil {
			if converted, err := extraTool.APIAuth.ToOpenapiAuth(); err == nil {
				auth = converted
			}
		}
		tools = append(tools, wga_sandbox_option.Tool{
			OpenAPI3Schema: extraTool.OpenAPI3Schema,
			OperationIDs:   nil,
			APIAuth:        auth,
		})
	}
	if len(tools) > 0 {
		opts = append(opts, wga_sandbox_option.WithTools(tools))
	}

	// 传递 MCP 服务器
	if len(a.options.MCPs) > 0 {
		mcps := make([]wga_sandbox_option.MCP, len(a.options.MCPs))
		for i, mcp := range a.options.MCPs {
			mcps[i] = wga_sandbox_option.MCP{
				Name:        mcp.Name,
				URL:         mcp.URL,
				Description: mcp.Description,
				ApiAuth:     mcp.ApiAuth,
				Headers:     mcp.Headers,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMCPs(mcps))
	}

	// 传递 SystemMessageStrategy 到 wga-sandbox
	if a.options.SystemMessageStrategy == option.SystemMessageStrategyMerge {
		opts = append(opts, wga_sandbox_option.WithSystemMessageStrategy(wga_sandbox_option.SystemMessageStrategyMerge))
	}

	// 传递 trace 上下文到 sandbox 容器，实现跨进程 trace 传播
	if traceHeaders := trace_util.ExtractTraceHeaders(ctx); len(traceHeaders) > 0 {
		opts = append(opts, wga_sandbox_option.WithTraceContext(traceHeaders))
	}

	return opts, nil
}

// convertSkillVariables 转换 Skill 变量列表类型。
func convertSkillVariables(variables []config.SkillVariable) []wga_sandbox_option.SkillVariable {
	if len(variables) == 0 {
		return nil
	}
	result := make([]wga_sandbox_option.SkillVariable, len(variables))
	for i, v := range variables {
		result[i] = wga_sandbox_option.SkillVariable{
			Name:          v.Name,
			Description:   v.Description,
			VariableKey:   v.VariableKey,
			VariableValue: v.VariableValue,
		}
	}
	return result
}
