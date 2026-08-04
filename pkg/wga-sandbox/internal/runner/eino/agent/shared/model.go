package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// isO1Series 判断是否为 OpenAI o1 系列推理模型。
// o1 系列收到 max_tokens 会报 400（unsupported_parameter），需豁免只传 MaxCompletionTokens。
// 判断基于 OpenAI reasoning 模型现有命名规律（o1/o3/o4 前缀），未来出新前缀需补。
func isO1Series(modelID string) bool {
	m := strings.ToLower(modelID)
	return strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4")
}

// 创建关闭思考模式的 ChatModel。
func NewNoReasonChatModel(ctx context.Context, cfg AppConfig) (model.ToolCallingChatModel, error) {
	chatCfg := &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.ModelID,
		ExtraFields: map[string]any{
			"enable_thinking": false,
			"thinking": map[string]any{
				"type": "disabled",
			},
			"chat_template_kwargs": map[string]any{
				"enable_thinking": false,
			},
		},
	}

	// 按启用开关设采样参数。指针为 nil（*Enable=false）时不设，chat model 用模型默认值。
	// MaxTokens 字段策略：默认 MaxTokens 与 MaxCompletionTokens 同值都传（覆盖认任一字段的模型），
	// 只对 o1 系列只传 MaxCompletionTokens（o1 收到 max_tokens 会报 400）。
	if p := cfg.Params; p != nil {
		if p.TemperatureEnable {
			v := float32(p.Temperature)
			chatCfg.Temperature = &v
		}
		if p.TopPEnable {
			v := float32(p.TopP)
			chatCfg.TopP = &v
		}
		if p.FrequencyPenaltyEnable {
			v := float32(p.FrequencyPenalty)
			chatCfg.FrequencyPenalty = &v
		}
		if p.PresencePenaltyEnable {
			v := float32(p.PresencePenalty)
			chatCfg.PresencePenalty = &v
		}
		if p.MaxTokensEnable {
			v := int(p.MaxTokens)
			if isO1Series(cfg.ModelID) {
				chatCfg.MaxCompletionTokens = &v
			} else {
				chatCfg.MaxTokens = &v
				chatCfg.MaxCompletionTokens = &v
			}
		}
	}

	m, err := openai.NewChatModel(ctx, chatCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}
	return m, nil
}
