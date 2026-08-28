package tokenizer_service

import service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"

// TokenizerType tokenizer 类型
type TokenizerType int

const (
	DefaultTokenizer TokenizerType = 0 // 默认 tokenizer（基于字符比例估算）
	limitTokenRate   float64       = 0.7
)

// Tokenizer 定义 token 计算器接口
type Tokenizer interface {
	// Type 返回 tokenizer 类型
	Type() TokenizerType
	// CountTokens 计算输入文本的 token 数量
	CountTokens(text string) (int, error)
	// TruncateText 按最大 token 数量截断文本
	TruncateText(text string, maxTokens int) (string, error)
}

// NewTokenizer 创建 tokenizer 的工厂方法
func NewTokenizer(tokenizerType TokenizerType) Tokenizer {
	switch tokenizerType {
	case DefaultTokenizer:
		return NewCustomTokenizer()
	default:
		return nil
	}
}

func TokenLimit(chatInfo *service_model.AgentChatInfo) int {
	return int(float64(buildTokenLimit(chatInfo)) * limitTokenRate)
}

func buildTokenLimit(agentChatInfo *service_model.AgentChatInfo) int {
	if agentChatInfo.ModelInfo != nil && agentChatInfo.ModelInfo.Config != nil {
		return agentChatInfo.ModelInfo.Config.ContextSize
	}
	return 0
}
