package local_agent

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_http_client "github.com/UnicomAI/wanwu/internal/agent-service/pkg/http"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

const (
	DefaultTemperature      float32 = 0.7
	DefaultTopP             float32 = 1.0
	DefaultFrequencyPenalty float32 = 0.0
	DefaultPresencePenalty  float32 = 0.0
	DefaultMaxTokens        int     = 1024
)

type LocalAgentService interface {
	//CreateChatModel 创建chatModel
	CreateChatModel(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo) (model.ToolCallingChatModel, error)
	//BuildAgentInput 构造会话消息
	BuildAgentInput(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, agentInput *adk.AgentInput, generator *adk.AsyncGenerator[*adk.AgentEvent]) (*adk.AgentInput, error)
}

func CreateLocalAgentService(ctx context.Context, req *request.AgentChatParams, agentChatInfo *service_model.AgentChatInfo, chatContext *request.AgentChatContext) LocalAgentService {
	////如果有特殊输出或者逻辑的模型可以仿照，vision_chat 实现，不过目前主流的的vision_chat 都是openai格式,ChatAgent都可兼容
	//if agentChatInfo.VisionSupport {
	//	return &VisionChatAgent{}
	//}
	return &ChatAgent{ChatContext: chatContext}
}

func CreateChatModel(ctx context.Context, agentChatInfo *service_model.AgentChatInfo, req *request.AgentChatParams) (model.ToolCallingChatModel, error) {
	modelInfo := agentChatInfo.ModelInfo
	modelConfig := modelInfo.Config
	params := req.ModelParams
	enableThinking := req.ModelParams.EnableThinking
	var extraFields map[string]any
	if enableThinking != nil {
		var thinking = false
		if *enableThinking == 1 {
			thinking = true
		}
		extraFields = map[string]any{"enable_thinking": thinking}
	}

	if config.GetToolTemplateConfig().DeepSeekReasoning(modelInfo.Provider, modelInfo.Model) {
		return deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			HTTPClient:       agent_http_client.GetClient().Client,
			APIKey:           modelConfig.ApiKey,
			BaseURL:          modelConfig.EndpointUrl,
			Model:            modelInfo.Model,
			Temperature:      buildFloat32(params.Temperature, DefaultTemperature),
			TopP:             buildFloat32(params.TopP, DefaultTopP),
			FrequencyPenalty: buildFloat32(params.FrequencyPenalty, DefaultFrequencyPenalty),
			PresencePenalty:  buildFloat32(params.PresencePenalty, DefaultPresencePenalty),
			MaxTokens:        buildInt(params.MaxTokens, DefaultMaxTokens),
			ThinkingConfig:   buildDsThinkingConfig(enableThinking),
		})
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		HTTPClient:          agent_http_client.GetClient().Client,
		APIKey:              modelConfig.ApiKey,
		BaseURL:             modelConfig.EndpointUrl,
		Model:               modelInfo.Model,
		Temperature:         params.Temperature,
		TopP:                params.TopP,
		FrequencyPenalty:    params.FrequencyPenalty,
		PresencePenalty:     params.PresencePenalty,
		ExtraFields:         extraFields,
		MaxCompletionTokens: params.MaxTokens,
	})
}

func buildFloat32(value *float32, defaultValue float32) float32 {
	if value == nil {
		return defaultValue
	}
	return *value
}

func buildInt(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return *value
}

// buildDsThinkingConfig 构建deepseek thinking配置,
func buildDsThinkingConfig(enableThinking *int) *deepseek.ThinkingConfig {
	if enableThinking != nil {
		var thinking = "disabled"
		if *enableThinking == 1 {
			thinking = "enabled"
		}
		return &deepseek.ThinkingConfig{
			Type: thinking,
		}
	}
	return nil
}
