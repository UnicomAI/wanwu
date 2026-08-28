// Package factory 提供智能体实例创建功能。
package factory

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/schema"
)

// NewAgent 创建智能体实例。
func NewAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.Agent, error) {
	return newAgent(ctx, cfg, options)
}

func newAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.Agent, error) {
	switch cfg.Type {
	case config.AgentTypeReAct:
		return newReactAgent(ctx, cfg, options)
	case config.AgentTypeSandbox:
		return newSandboxAgent(ctx, cfg, options)
	case config.AgentTypeSequential:
		return newSequentialAgent(ctx, cfg, options)
	case config.AgentTypeLoop:
		return newLoopAgent(ctx, cfg, options)
	case config.AgentTypeParallel:
		return newParallelAgent(ctx, cfg, options)
	case config.AgentTypeSupervisor:
		return newSupervisorAgent(ctx, cfg, options)
	default:
		return nil, fmt.Errorf("agent (%v) type (%v) unsupported", cfg.ID, cfg.Type)
	}
}

func newReactAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.Agent, error) {
	instruction, err := options.FormatInstruction(ctx, cfg)
	if err != nil {
		return nil, err
	}
	model, err := options.ToChatModel(ctx)
	if err != nil {
		return nil, err
	}
	tools, err := options.ToToolsConfig(cfg.ToolCategories)
	if err != nil {
		return nil, err
	}

	// 兜底：父级模型误调未注册工具（典型如沙箱子智能体内部工具 bash/skill）时，
	// 返回可恢复的 Tool 结果，避免 ToolNode 因索引找不到工具而直接报错中断整个 ReAct 循环。
	// 注意：提示词刻意不引导 transfer_to_agent，避免父级在子智能体已收尾后再次转移造成空转重跑沙箱。
	tools.UnknownToolsHandler = func(_ context.Context, name, input string) (string, error) {
		return fmt.Sprintf(
			"工具 `%s` 不属于当前智能体的可调用工具，它只在子智能体（如沙箱）执行过程内部使用，"+
				"所需结果通常已在上文对话中给出。请直接基于已有结果继续或总结回答，"+
				"不要重复调用该工具或再次转移子智能体；确需让子智能体重新执行时，请在其指令中明确说明新的要求。",
			name), nil
	}

	config := &adk.ChatModelAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		Instruction: instruction,
		Model:       model,
		ToolsConfig: tools,

		MaxIterations: cfg.Configure.MaxIterations,
	}

	if options.SystemMessageStrategy == option.SystemMessageStrategyMerge {
		config.GenModelInput = func(ctx context.Context, instr string, input *adk.AgentInput) ([]adk.Message, error) {
			msgs := make([]adk.Message, 0, len(input.Messages)+1)

			extraSystem, otherMessages := option.ExtractSystemMessage(input.Messages)

			systemPrompt := instr
			if extraSystem != "" {
				if systemPrompt != "" {
					systemPrompt += "\n\n"
				}
				systemPrompt += extraSystem
			}

			if systemPrompt != "" {
				msgs = append(msgs, schema.SystemMessage(systemPrompt))
			}

			msgs = append(msgs, otherMessages...)

			return msgs, nil
		}
	}

	return adk.NewChatModelAgent(ctx, config)
}

func newSequentialAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,
	})
}

func newLoopAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,

		MaxIterations: cfg.Configure.MaxIterations,
	})
}

func newParallelAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,
	})
}

func newSupervisorAgent(ctx context.Context, cfg *config.Agent, options option.Options) (adk.Agent, error) {
	agent, err := newReactAgent(ctx, cfg, options)
	if err != nil {
		return nil, err
	}
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: agent,
		SubAgents:  subAgents,
	})
}
