package shared

import (
	"context"

	"github.com/cloudwego/eino/adk"
)

// LLMParams 模型采样参数。
// 这是 mp_common.LLMParams（pkg/model-provider/mp-common）的结构等价本地副本——
// 字段名、类型、json tag 完全一致，保证跨进程 JSON round-trip 互通。
// 不直接 import mp_common：该包会经 trace-util 拖入 api/proto/common、redis、otel 等重依赖，
// 而 eino-agent 二进制的 Dockerfile 构建上下文只 COPY pkg/（无 api/），import mp_common 会导致
// 编译期 "no required module provides package .../api/proto/common"。宿主侧仍用 mp_common.LLMParams
// 序列化写 .env，沙箱内用本类型反序列化，两侧 JSON 契约一致即可。
type LLMParams struct {
	Temperature            float64 `json:"temperature"`              // 温度
	TemperatureEnable      bool    `json:"temperatureEnable"`        // 温度(开关)
	TopP                   float64 `json:"topP"`                     // Top P
	TopPEnable             bool    `json:"topPEnable"`               // Top P(开关)
	FrequencyPenalty       float64 `json:"frequencyPenalty"`         // 频率惩罚
	FrequencyPenaltyEnable bool    `json:"frequencyPenaltyEnable"`   // 频率惩罚(开关)
	PresencePenalty        float64 `json:"presencePenalty"`          // 存在惩罚
	PresencePenaltyEnable  bool    `json:"presencePenaltyEnable"`    // 存在惩罚(开关)
	MaxTokens              int32   `json:"maxTokens"`                // 最大标记
	MaxTokensEnable        bool    `json:"maxTokensEnable"`          // 最大标记(开关)
	ThinkingEnable         *bool   `json:"thinkingEnable,omitempty"` // 思考过程(开关)
}

// 模型相关默认值，用于 AppConfig.ApplyDefaults 兜底。
const (
	defaultBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	defaultModel   = "mimo-v2-flash"
)

// AppConfig 提供构建 App 所需的全部配置。
type AppConfig struct {
	Workspace string
	APIKey    string
	BaseURL   string
	ModelID   string
	// Params 模型采样参数（temperature/topP/maxTokens 等），由宿主经 .env 注入。
	// 为 nil 时沙箱内 chat model 用模型默认值。
	Params *LLMParams
	// Halt 连续 [BLOCKED:...] 计数与熔断回调。可空——为 nil 时不启用熔断
	// （用于 tests + oneshot 沙箱路径的向后兼容）。
	Halt *HaltState
}

// ApplyDefaults 填充默认值。
func (c *AppConfig) ApplyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.ModelID == "" {
		c.ModelID = defaultModel
	}
}

// Validate 校验必填字段。
func (c *AppConfig) Validate() error {
	// if c.APIKey == "" {
	// 	return fmt.Errorf("OPENAI_API_KEY is required")
	// }
	return nil
}

// AgentApp 定义 agent 应用的统一接口。
type AgentApp interface {
	Query(ctx context.Context, messages []adk.Message) *adk.AsyncIterator[*adk.AgentEvent]
	Close() error
}
