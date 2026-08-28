package mp_ollama

import (
	"context"
	"fmt"
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type LLM struct {
	ApiKey          string `json:"apiKey"`                                              // ApiKey
	EndpointUrl     string `json:"endpointUrl"`                                         // 推理url
	FunctionCalling string `json:"functionCalling" validate:"oneof=noSupport toolCall"` // 函数调用是否支持
	VisionSupport   string `json:"visionSupport" validate:"oneof=noSupport support"`    // 视觉支持
	ThinkingSupport string `json:"thinkingSupport" validate:"oneof=noSupport support"`  // 深度思考是否支持
	MaxTokens       *int   `json:"maxTokens"`                                           // 模型回答最大tokens
	ContextSize     *int   `json:"contextSize"`                                         // 上下文长度
}

func (cfg *LLM) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagChat,
		},
	}
	tags = append(tags, mp_common.GetTagsByVisionSupport(cfg.VisionSupport)...)
	tags = append(tags, mp_common.GetTagsByFunctionCall(cfg.FunctionCalling)...)
	tags = append(tags, mp_common.GetTagsByContentSize(cfg.ContextSize)...)
	return tags
}

func (cfg *LLM) NewReq(req *mp_common.LLMReq) (mp_common.ILLMReq, error) {
	if req.MaxTokens != nil && cfg.ContextSize != nil && *req.MaxTokens > *cfg.ContextSize {
		return nil, fmt.Errorf("max_tokens too large (max allowed: %d)", *cfg.ContextSize)
	}
	m, err := req.Data()
	if err != nil {
		return nil, err
	}
	// 处理思考能力控制：优先使用 ReasoningEffort，否则从 EnableThinking 映射
	if req.ReasoningEffort != nil {
		m["reasoning_effort"] = *req.ReasoningEffort
	} else if req.EnableThinking != nil {
		if *req.EnableThinking {
			m["reasoning_effort"] = "high"
		} else {
			m["reasoning_effort"] = "none"
		}
	}
	if req.Stream != nil && *req.Stream {
		if req.StreamOptions != nil && req.StreamOptions.IncludeUsage != nil {
			m["stream_options"] = map[string]bool{"include_usage": *req.StreamOptions.IncludeUsage}
		} else {
			m["stream_options"] = map[string]bool{"include_usage": true}
		}
	}
	return mp_common.NewLLMReq(m), nil
}

func (cfg *LLM) ChatCompletions(ctx context.Context, req mp_common.ILLMReq, headers ...mp_common.Header) (mp_common.ILLMResp, <-chan mp_common.ILLMResp, error) {
	return mp_common.ChatCompletions(ctx, "ollama", cfg.ApiKey, cfg.chatCompletionsUrl(), req, mp_common.NewLLMResp, headers...)
}

func (cfg *LLM) chatCompletionsUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/chat/completions")
	return ret
}
