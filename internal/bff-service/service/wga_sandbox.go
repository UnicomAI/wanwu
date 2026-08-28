package service

import (
	"context"
	"fmt"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

func WgaSandboxRun(ctx *gin.Context, req *request.WgaSandboxRunReq) error {
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: req.Model.ModelId})
	if err != nil {
		return err
	}
	if !modelInfo.IsActive {
		return grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId)
	}

	endpoint := mp.ToModelEndpoint(modelInfo.ModelId, modelInfo.Model)
	modelURL, _ := endpoint["model_url"].(string)
	modelConfig := wga_sandbox_option.ModelConfig{
		Provider:     modelInfo.Provider,
		ProviderName: modelInfo.Provider,
		BaseURL:      modelURL,
		APIKey:       "",
		Model:        modelInfo.Model,
		ModelName:    modelInfo.DisplayName,
	}

	var sandbox wga_sandbox_option.SandboxConfig
	sandboxCfg := config.Cfg().WgaSandbox.Sandbox
	switch sandboxCfg.Type {
	case string(wga_sandbox_option.SandboxTypeOneshot):
		sandbox = wga_sandbox_option.SandboxOneshot(sandboxCfg.ImageName)
	default:
		sandbox = wga_sandbox_option.SandboxReuse(sandboxCfg.Host)
	}

	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
			ThreadID: req.ThreadID,
			RunID:    req.RunID,
		}),
		wga_sandbox_option.WithModelConfig(modelConfig),
		wga_sandbox_option.WithSandbox(sandbox),
		wga_sandbox_option.WithRunnerType(wga_sandbox_option.RunnerTypeOpencode),
		wga_sandbox_option.WithInstruction(req.Instruction),
		wga_sandbox_option.WithOverallTask(req.OverallTask),
		wga_sandbox_option.WithEnableThinking(req.EnableThinking),
		wga_sandbox_option.WithSkipCleanup(req.SkipCleanup),
		wga_sandbox_option.WithAgentName(req.AgentName),
		wga_sandbox_option.WithInputDir(req.InputDir),
		wga_sandbox_option.WithOutputDir(req.OutputDir),
	}

	if len(req.Messages) > 0 {
		messages := make([]adk.Message, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = &schema.Message{
				Role:    schema.RoleType(msg.Role),
				Content: msg.Content,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMessages(messages))
	}

	if len(req.Skills) > 0 {
		skills := make([]wga_sandbox_option.Skill, len(req.Skills))
		for i, skill := range req.Skills {
			skills[i] = wga_sandbox_option.Skill{
				Dir:       skill.Dir,
				Variables: convertWgaSandboxSkillVariables(skill.Variables),
			}
		}
		opts = append(opts, wga_sandbox_option.WithSkills(skills))
	}

	if len(req.MCPs) > 0 {
		mcps := make([]wga_sandbox_option.MCP, len(req.MCPs))
		for i, mcp := range req.MCPs {
			mcps[i] = wga_sandbox_option.MCP{
				Name:        mcp.Name,
				URL:         mcp.URL,
				Description: mcp.Description,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMCPs(mcps))
	}

	if len(req.Tools) > 0 {
		tools, err := convertWgaSandboxTools(ctx.Request.Context(), req.Tools)
		if err != nil {
			return err
		}
		opts = append(opts, wga_sandbox_option.WithTools(tools))
	}

	_, outputCh, err := wga_sandbox.Run(ctx.Request.Context(), opts...)
	if err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("wga sandbox run failed: %v", err))
	}

	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[WGA][Sandbox] model %s run", req.Model.ModelId), sse_util.DONE_MSG).
		WriteStream(outputCh, nil, buildWgaSandboxLineProcessor(), nil)
	return nil
}

func WgaSandboxCleanup(ctx *gin.Context, req *request.WgaSandboxCleanupReq) error {
	if err := wga_sandbox.Cleanup(ctx.Request.Context(), req.RunID); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("wga sandbox cleanup failed: %v", err))
	}
	return nil
}

func buildWgaSandboxLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data: " + lineText + "\n\n", false, nil
	}
}

func getWgaSandboxConfig() (wga_sandbox_option.SandboxConfig, error) {
	cfg := config.Cfg().WgaSandbox.Sandbox
	if cfg.Type != "reuse" || cfg.Host == "" {
		return wga_sandbox_option.SandboxConfig{}, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "sandbox config not available or not in reuse mode")
	}
	return wga_sandbox_option.SandboxReuse(cfg.Host), nil
}

func getWgaSandboxOntologyConfig() (wga_sandbox_option.SandboxConfig, error) {
	cfg := config.Cfg().WgaSandbox.SandboxOntology
	if cfg.Type != "reuse" || cfg.Host == "" {
		return wga_sandbox_option.SandboxConfig{}, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "sandbox(ontology) config not available or not in reuse mode")
	}
	return wga_sandbox_option.SandboxReuse(cfg.Host), nil
}

func convertWgaSandboxSkillVariables(variables []request.WgaSandboxSkillVariable) []wga_sandbox_option.SkillVariable {
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

func convertWgaSandboxTools(ctx context.Context, tools []request.WgaSandboxTool) ([]wga_sandbox_option.Tool, error) {
	if len(tools) == 0 {
		return nil, nil
	}
	result := make([]wga_sandbox_option.Tool, 0, len(tools))
	for i, tool := range tools {
		doc, err := openapi3_util.LoadFromData(ctx, []byte(tool.Schema))
		if err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("tools[%d] invalid schema: %v", i, err))
		}
		var auth *openapi3_util.Auth
		if tool.ApiAuth != nil {
			auth, err = tool.ApiAuth.ToOpenapiAuth()
			if err != nil {
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("tools[%d] invalid apiAuth: %v", i, err))
			}
		}
		result = append(result, wga_sandbox_option.Tool{
			OpenAPI3Schema: doc,
			OperationIDs:   tool.OperationIDs,
			APIAuth:        auth,
		})
	}
	return result, nil
}
