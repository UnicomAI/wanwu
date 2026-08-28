package chatmodel

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/shared"
	"github.com/cloudwego/eino/adk"
)

type App struct {
	runner    *adk.Runner
	workspace string
}

func NewApp(ctx context.Context, cfg shared.AppConfig) (shared.AgentApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cfg.ApplyDefaults()

	chatModel, err := shared.NewNoReasonChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}

	skillMW, err := shared.NewSkillMiddleware(ctx, cfg.Workspace)
	if err != nil {
		return nil, err
	}

	bashMW, err := shared.NewBashMiddleware(cfg.Workspace, cfg.Halt)
	if err != nil {
		return nil, err
	}

	instruction, err := renderInstruction(cfg.Workspace)
	if err != nil {
		return nil, err
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Model:         chatModel,
		Name:          "ChatModelAgent",
		Description:   "一个带有skills的智能体助手",
		Instruction:   instruction,
		Middlewares:   []adk.AgentMiddleware{skillMW, bashMW},
		MaxIterations: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	return &App{runner: runner, workspace: cfg.Workspace}, nil
}

func (a *App) Query(ctx context.Context, messages []adk.Message) *adk.AsyncIterator[*adk.AgentEvent] {
	log.Printf("[App] Query with %d messages | Workspace: %s", len(messages), shared.SanitizeForLog(a.workspace))
	return a.runner.Run(ctx, messages)
}

func (a *App) Close() error {
	return nil
}

func renderInstruction(workspace string) (string, error) {
	tmpl, err := template.New("instruction").Parse(instructionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse instruction template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Workspace   string
		CurrentTime string
	}{
		Workspace:   workspace,
		CurrentTime: time.Now().Format("2006/01/02 Mon"),
	}); err != nil {
		return "", fmt.Errorf("failed to render instruction template: %w", err)
	}
	return buf.String(), nil
}
