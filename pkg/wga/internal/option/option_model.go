package option

import (
	"context"
	"fmt"
	"strings"

	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	deepseek "github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func (options *Options) checkModel() error {
	if options.Model.Model == "" {
		return fmt.Errorf("model required")
	}
	if options.Model.BaseURL == "" {
		return fmt.Errorf("model base url empty")
	}
	return nil
}

func needsDeepSeekCompat(m ModelConfig) bool {
	model := strings.ToLower(m.Model)
	return strings.Contains(model, "deepseek-v4") || strings.Contains(model, "deepseek_v4")
}

// ToChatModel 创建聊天模型实例。
func (options *Options) ToChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if err := options.checkModel(); err != nil {
		return nil, err
	}
	if needsDeepSeekCompat(options.Model) {
		return newDeepSeekChatModel(ctx, options.Model)
	}
	return newOpenAIChatModel(ctx, options.Model)
}

func newOpenAIChatModel(ctx context.Context, m ModelConfig) (model.ToolCallingChatModel, error) {
	cfg := &openai.ChatModelConfig{
		Model:      m.Model,
		APIKey:     m.APIKey,
		BaseURL:    m.BaseURL,
		HTTPClient: http_client.Default().Client,
	}
	if m.Params != nil {
		if m.Params.TemperatureEnable {
			temp := float32(m.Params.Temperature)
			cfg.Temperature = &temp
		}
		if m.Params.TopPEnable {
			topP := float32(m.Params.TopP)
			cfg.TopP = &topP
		}
		if m.Params.FrequencyPenaltyEnable {
			fp := float32(m.Params.FrequencyPenalty)
			cfg.FrequencyPenalty = &fp
		}
		if m.Params.PresencePenaltyEnable {
			pp := float32(m.Params.PresencePenalty)
			cfg.PresencePenalty = &pp
		}
	}
	return openai.NewChatModel(ctx, cfg)
}

func newDeepSeekChatModel(ctx context.Context, m ModelConfig) (model.ToolCallingChatModel, error) {
	cfg := &deepseek.ChatModelConfig{
		APIKey:     m.APIKey,
		BaseURL:    m.BaseURL,
		Model:      m.Model,
		HTTPClient: http_client.Default().Client,
	}
	if m.Params != nil {
		if m.Params.TemperatureEnable {
			cfg.Temperature = float32(m.Params.Temperature)
		}
		if m.Params.TopPEnable {
			cfg.TopP = float32(m.Params.TopP)
		}
		if m.Params.FrequencyPenaltyEnable {
			cfg.FrequencyPenalty = float32(m.Params.FrequencyPenalty)
		}
		if m.Params.PresencePenaltyEnable {
			cfg.PresencePenalty = float32(m.Params.PresencePenalty)
		}
	}
	return deepseek.NewChatModel(ctx, cfg)
}
