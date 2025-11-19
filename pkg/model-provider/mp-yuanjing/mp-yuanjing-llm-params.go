package mp_yuanjing

type LLMParams struct {
	Temperature            float32 `json:"temperature"`            // temperature
	TemperatureEnable      bool    `json:"temperatureEnable"`      // Temperature (switch)
	TopP                   float32 `json:"topP"`                   // Top P
	TopPEnable             bool    `json:"topPEnable"`             // Top P(switch)
	FrequencyPenalty       float32 `json:"frequencyPenalty"`       // frequency penalty
	FrequencyPenaltyEnable bool    `json:"frequencyPenaltyEnable"` // Frequency Penalty (Switching)
	PresencePenalty        float32 `json:"presencePenalty"`        // There is punishment
	PresencePenaltyEnable  bool    `json:"presencePenaltyEnable"`  // There is a penalty (switch)
	MaxTokens              int32   `json:"maxTokens"`              // maximum mark
	MaxTokensEnable        bool    `json:"maxTokensEnable"`        // maximum mark(switch)
}

func (cfg *LLMParams) GetParams() map[string]interface{} {
	ret := make(map[string]interface{})

	if cfg.TemperatureEnable {
		ret["temperature"] = cfg.Temperature
	}
	if cfg.TopPEnable {
		ret["top_p"] = cfg.TopP
	}
	if cfg.FrequencyPenaltyEnable {
		ret["frequency_penalty"] = cfg.FrequencyPenalty
	}
	if cfg.PresencePenaltyEnable {
		ret["presence_penalty"] = cfg.PresencePenalty
	}
	if cfg.MaxTokensEnable {
		ret["max_tokens"] = cfg.MaxTokens
	}
	return ret
}
