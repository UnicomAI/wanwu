package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type PromptTemplateDetail struct {
	TemplateId string `json:"templateId" validate:"required"`
	request.AppBriefConfig
	Category string `json:"category"` // 模板分类 [EN] Template classification
	Author   string `json:"author"`   // 作者 [EN] author
	Prompt   string `json:"prompt"`   // 提示词 [EN] prompt word
}

type PromptIDData struct {
	PromptId string `json:"promptId"`
}
