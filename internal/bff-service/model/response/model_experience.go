package response

import (
	"encoding/json"
	"strings"
)

// ModelExperienceDialog 模型体验对话
type ModelExperienceDialog struct {
	ID           string `json:"id"` // modelExperienceId
	ModelId      string `json:"modelId"`
	SessionId    string `json:"sessionId"`
	Title        string `json:"title"`
	ModelSetting string `json:"modelSetting"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// ModelExperienceDialogRecord 模型体验对话记录
type ModelExperienceDialogRecord struct {
	ModelExperienceId string                 `json:"modelExperienceId"` // 模型体验 ID
	ModelId           string                 `json:"modelId"`           // 模型 ID
	SessionId         string                 `json:"sessionId"`         // Session ID
	OriginalContent   string                 `json:"originalContent"`   // 原始内容
	ReasoningContent  string                 `json:"reasoningContent"`  // 思考过程
	Role              string                 `json:"role"`              // 角色
	RequestFiles      []AssistantRequestFile `json:"requestFiles"`      // 文件信息
}

type ModelExperienceResp struct {
	Id                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int         `json:"created"`
	Model             string      `json:"model"`
	Choices           []*Choices  `json:"choices"`
	Usage             *Usage      `json:"usage"`
	ServiceTier       interface{} `json:"service_tier"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
	Code              int         `json:"code"`
	ImgId             string      `json:"img_id"`
}

func (m *ModelExperienceResp) Compact(newMsg *ModelExperienceResp) *ModelExperienceResp {

	if len(m.Choices) > 0 && len(newMsg.Choices) > 0 && m.Choices[0].Delta != nil && newMsg.Choices[0].Delta != nil {
		// 处理正文输出请情况
		if m.Choices[0].Delta.Content != "" && newMsg.Choices[0].Delta.Content != "" && m.Choices[0].Delta.ReasoningContent == "" && newMsg.Choices[0].Delta.ReasoningContent == "" {
			content := ""
			for _, nchoice := range newMsg.Choices {
				content += nchoice.Delta.Content
			}
			m.Choices[0].Delta.Content += content
			return m
		}
		// 合并 reasoningContent
		if m.Choices[0].Delta.Content == "" && newMsg.Choices[0].Delta.Content == "" && m.Choices[0].Delta.ReasoningContent != "" && newMsg.Choices[0].Delta.ReasoningContent != "" {
			reasonContent := ""
			for _, nchoice := range newMsg.Choices {
				reasonContent += nchoice.Delta.ReasoningContent
			}
			m.Choices[0].Delta.ReasoningContent += reasonContent
			return m
		}
	}
	return nil
}

func UnmarshalModelExperienceResp(data string) (*ModelExperienceResp, error) {
	data = strings.TrimPrefix(data, "data:")
	resp := ModelExperienceResp{}
	err := json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func MarshalModelExperienceResp(data *ModelExperienceResp) (string, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return "data:" + string(marshal), nil
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
type Choices struct {
	Index        int         `json:"index"`
	Delta        *Delta      `json:"delta"`
	FinishReason string      `json:"finish_reason"`
	Logprobs     interface{} `json:"logprobs"`
}
type Delta struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}
