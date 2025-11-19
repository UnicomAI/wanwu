package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type PromptTemplateDetail struct {
	TemplateId string `json:"templateId" validate:"required"`
	request.AppBriefConfig
	Category string `json:"category"` // Template classification
	Author   string `json:"author"`   // author
	Prompt   string `json:"prompt"`   // prompt word
}

type PromptIDData struct {
	PromptId string `json:"promptId"`
}
