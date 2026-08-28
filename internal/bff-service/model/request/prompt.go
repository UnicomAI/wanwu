package request

import (
	"errors"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/util"
)

type CustomPromptCreate struct {
	Avatar Avatar `json:"avatar"`                     // 图标
	Name   string `json:"name" validate:"required"`   // 名称
	Desc   string `json:"desc" validate:"required"`   // 描述
	Prompt string `json:"prompt" validate:"required"` // 提示词
}

func (c *CustomPromptCreate) Check() error {
	if err := util.ValidateName(&c.Name, util.SubjectPrompt); err != nil {
		return err
	}
	if err := util.ValidateDesc(&c.Desc, util.SubjectPrompt); err != nil {
		return err
	}
	if len(strings.TrimSpace(c.Prompt)) == 0 {
		return errors.New("prompt is empty")
	}
	return nil
}

type CustomPromptIDReq struct {
	CustomPromptID string `json:"customPromptId" validate:"required"` // 自定义提示词ID
}

func (req *CustomPromptIDReq) Check() error {
	return nil
}

type UpdateCustomPrompt struct {
	CustomPromptIDReq
	Avatar Avatar `json:"avatar"`                     // 图标
	Name   string `json:"name" validate:"required"`   // 名称
	Desc   string `json:"desc" validate:"required"`   // 描述
	Prompt string `json:"prompt" validate:"required"` // 提示词
}

func (u *UpdateCustomPrompt) Check() error {
	// name/desc 校验在 service 层按"未改名跳过校验"处理
	if len(strings.TrimSpace(u.Prompt)) == 0 {
		return errors.New("prompt is empty")
	}
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

type PromptReasonReq struct {
	Prompt  string `json:"prompt" validate:"required"`
	ModelId string `json:"modelId" validate:"required"`
}

func (req *PromptReasonReq) Check() error { return nil }

type PromptEvaluateReq struct {
	Answer         string `json:"answer" validate:"required"`
	ExpectedOutput string `json:"expectedOutput" validate:"required"`
	ModelId        string `json:"modelId" validate:"required"`
}

func (req *PromptEvaluateReq) Check() error { return nil }
