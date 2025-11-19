package request

type CustomPromptCreate struct {
	Avatar Avatar `json:"avatar"`                     // 图标 [EN] icon
	Name   string `json:"name" validate:"required"`   // 名称 [EN] name
	Desc   string `json:"desc" validate:"required"`   // 描述 [EN] describe
	Prompt string `json:"prompt" validate:"required"` // 提示词 [EN] prompt word
}

func (req *CustomPromptCreate) Check() error {
	return nil
}

type CustomPromptIDReq struct {
	CustomPromptID string `json:"customPromptId" validate:"required"` // 自定义提示词ID [EN] Custom prompt word ID
}

func (req *CustomPromptIDReq) Check() error {
	return nil
}

type UpdateCustomPrompt struct {
	CustomPromptIDReq
	Avatar Avatar `json:"avatar"`                     // 图标 [EN] icon
	Name   string `json:"name" validate:"required"`   // 名称 [EN] name
	Desc   string `json:"desc" validate:"required"`   // 描述 [EN] describe
	Prompt string `json:"prompt" validate:"required"` // 提示词 [EN] prompt word
}

func (req *UpdateCustomPrompt) Check() error {
	return nil
}

type CreatePromptByTemplateReq struct {
	TemplateId string `json:"templateId" validate:"required"`
	AppBriefConfig
}

func (req *CreatePromptByTemplateReq) Check() error { return nil }

type PromptOptimizeReq struct {
	Prompt  string `json:"prompt" validate:"required"`
	ModelId string `json:"modelId" validate:"required"`
}

func (req *PromptOptimizeReq) Check() error { return nil }
