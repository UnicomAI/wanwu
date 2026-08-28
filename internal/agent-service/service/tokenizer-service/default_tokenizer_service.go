package tokenizer_service

import "errors"

const (
	defaultRatio = 1.7
)

// CustomTokenizer 基于 字符/token 比例估算的自定义 tokenizer
type CustomTokenizer struct {
	// ratio 字符到 token 的转换比例（默认 1.7）
	// 中文环境下通常 1个 token ≈ 1.5 ~ 2.0 个字符
	ratio float64
}

// NewCustomTokenizer 创建自定义 tokenizer
func NewCustomTokenizer() *CustomTokenizer {
	return &CustomTokenizer{ratio: defaultRatio}
}

// Type 返回 tokenizer 类型
func (t *CustomTokenizer) Type() TokenizerType {
	return DefaultTokenizer
}

// CountTokens 计算输入文本的 token 数量
// 公式: token数量 = 字符数量 / 转换比例
func (t *CustomTokenizer) CountTokens(text string) (int, error) {
	return int(float64(len(text)) / t.ratio), nil
}

// TruncateText 按最大 token 数量截断文本
func (t *CustomTokenizer) TruncateText(text string, maxTokens int) (string, error) {
	if maxTokens < 1 {
		return "", errors.New("maxTokens 必须大于 0")
	}

	// 反向计算最大允许的字符数
	maxChars := int(float64(maxTokens) * t.ratio)

	if len(text) <= maxChars {
		return text, nil
	}

	return text[:maxChars], nil
}
