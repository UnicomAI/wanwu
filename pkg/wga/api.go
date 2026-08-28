// Package wga 提供 AI 智能体的统一管理和执行接口。
//
// 支持多种智能体类型：react、sandbox、sequential、loop、parallel、deep、supervisor。
package wga

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/factory"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/cloudwego/eino/adk"
	"go.opentelemetry.io/otel/codes"
)

var (
	ErrWgaNotInit     = errors.New("wga not init")
	ErrWgaAlreadyInit = errors.New("wga already init")
)

var _agents []*config.Agent

// Init 初始化智能体配置。
func Init(ctx context.Context, configPath string) error {
	if _agents != nil {
		return ErrWgaAlreadyInit
	}
	agents, err := config.LoadAgents(ctx, configPath)
	if err != nil {
		return err
	}
	_agents = agents

	// 注册 WGA agent 级别 trace 回调处理器（受 JAEGER_ENABLE 环境变量控制）
	jaegerEnabled, _ := strconv.ParseBool(os.Getenv(trace_util.TraceJaegerEnable))
	if jaegerEnabled {
		trace_util.WgaGlobalTracing()
	}

	return nil
}

// GetAgentToolCategories 获取智能体及其子智能体的工具类别配置。
func GetAgentToolCategories(id string) ([]*wga_option.ToolCategoryInfo, error) {
	agent, err := getAgent(id)
	if err != nil {
		return nil, err
	}
	return agent.CollectToolCategories(), nil
}

// CheckToolOptions 检查智能体工具配置是否满足运行要求，返回检查结果。
func CheckToolOptions(_ context.Context, id string, opts ...option.Option) (*wga_option.CheckResult, error) {
	agentCfg, err := getAgent(id)
	if err != nil {
		return nil, err
	}
	var options option.Options
	if err := options.Apply(opts...); err != nil {
		return nil, err
	}
	// 检查工具配置（包括配置文件工具条件和额外工具冲突检查）
	toolCategories, err := options.CheckTools(agentCfg)
	if err != nil {
		return nil, err
	}
	return &wga_option.CheckResult{
		ToolCategories: toolCategories,
	}, nil
}

// Run 执行智能体任务，返回会话标识和事件迭代器。
func Run(ctx context.Context, id string, opts ...option.Option) (wga_option.RunSession, *adk.AsyncIterator[*adk.AgentEvent], error) {
	agentCfg, err := getAgent(id)
	if err != nil {
		return wga_option.RunSession{}, nil, err
	}
	var options option.Options
	if err := options.Apply(opts...); err != nil {
		return wga_option.RunSession{}, nil, err
	}
	if err := options.CheckModelConfig(); err != nil {
		return wga_option.RunSession{}, nil, err
	}
	if err := options.CheckMessages(); err != nil {
		return wga_option.RunSession{}, nil, err
	}
	// 暂不在 Run 阶段进行工具检查，由业务层决定是否调用 CheckOptions 进行工具检查；wga 内部不加载配置无效的工具
	// toolCategories, err := options.CheckToolOptions(agentCfg)
	// if err != nil {
	// 	return wga_option.RunSession{}, nil, err
	// }
	// for _, tc := range toolCategories {
	// 	if !tc.Meet {
	// 		return wga_option.RunSession{}, nil, fmt.Errorf("tool category (%s) condition (%s) not meet", tc.Category, tc.Condition)
	// 	}
	// }

	// 注入 WGA trace 元数据到 context，供 Eino 回调处理器读取
	wgaCtx := &trace_util.WgaTraceContext{
		AgentID:   agentCfg.ID,
		AgentType: string(agentCfg.Type),
		AgentName: agentCfg.Name,
		ThreadID:  options.RunSession.ThreadID,
		RunID:     options.RunSession.RunID,
		Model:     options.Model.Model,
	}
	ctx = trace_util.SetWgaTraceContext(ctx, wgaCtx)

	// 创建顶层 agent 执行 span
	ctx, agentSpan := trace_util.StartAgentSpan(ctx, string(agentCfg.Type), agentCfg.Name, agentCfg.ID)

	agent, err := factory.NewAgent(ctx, agentCfg, options)
	if err != nil {
		agentSpan.RecordError(err)
		agentSpan.SetStatus(codes.Error, err.Error())
		agentSpan.End()
		return wga_option.RunSession{}, nil, err
	}

	iter := agent.Run(ctx, &adk.AgentInput{Messages: options.Messages, EnableStreaming: true})

	// 包装迭代器：在迭代结束时自动结束 agent span
	wrappedIter := trace_util.WrapIteratorWithSpan(iter, agentSpan)

	return options.RunSession, wrappedIter, nil
}

// Cleanup 清理指定 runID 的沙箱工作目录。
func Cleanup(ctx context.Context, runID string) error {
	return wga_sandbox.Cleanup(ctx, runID)
}

// ReplyQuestion 回答问题（Human-in-the-Loop）。
// 仅支持 Reuse 模式。
func ReplyQuestion(ctx context.Context, sandboxCfg wga_sandbox_option.SandboxConfig, runID string, questionID string, answers [][]string) error {
	return wga_sandbox.ReplyQuestion(ctx, sandboxCfg, runID, questionID, answers)
}

// RejectQuestion 拒绝问题（Human-in-the-Loop）。
// 仅支持 Reuse 模式。
func RejectQuestion(ctx context.Context, sandboxCfg wga_sandbox_option.SandboxConfig, runID string, questionID string) error {
	return wga_sandbox.RejectQuestion(ctx, sandboxCfg, runID, questionID)
}

// --- internal ---

func getAgent(id string) (*config.Agent, error) {
	for _, agent := range _agents {
		if agent.ID == id {
			return agent, nil
		}
	}
	return nil, fmt.Errorf("agent (%s) not found", id)
}
