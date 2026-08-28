package request

import (
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
)

type ModelExperienceLlmRequest struct {
	ModelId           string                 `json:"modelId" validate:"required"`   // 模型ID
	SessionId         string                 `json:"sessionId" validate:"required"` // 会话 ID
	ModelExperienceId string                 `json:"modelExperienceId"`             // 体验对话ID（模型对比时为空）
	Content           string                 `json:"content" validate:"required"`   // 内容
	FileInfo          []ConversionStreamFile `json:"fileInfo" form:"fileInfo"`
	mp_common.LLMParams
}

func (o *ModelExperienceLlmRequest) Check() error {
	for _, f := range o.FileInfo {
		if f.FileUrl != "" {
			if err := url_util.ValidateURL(f.FileUrl); err != nil {
				return err
			}
		}
	}
	return nil
}

type ModelExperienceLlmConnectRequest struct {
	SessionId string `json:"sessionId" form:"sessionId" validate:"required"` // 会话 ID
	CommonCheck
}

type ModelExperienceLlmCancelRequest struct {
	SessionId string `json:"sessionId" form:"sessionId" validate:"required"` // 会话 ID
	CommonCheck
}

type ModelExperienceDialogRequest struct {
	ModelId      string      `json:"modelId" validate:"required"`   // 模型
	SessionId    string      `json:"sessionId" validate:"required"` // 会话 ID
	Title        string      `json:"title"`                         // 对话标题
	ModelSetting interface{} `json:"modelSetting"`                  // 模型参数配置
}

func (o *ModelExperienceDialogRequest) Check() error {
	return nil
}

type ModelExperienceDialogIDRequest struct {
	ModelExperienceId string `json:"modelExperienceId" validate:"required"` // 模型体验对话ID
}

func (o *ModelExperienceDialogIDRequest) Check() error {
	return nil
}

type ModelExperienceDialogRecordRequest struct {
	ModelExperienceId string `json:"modelExperienceId" form:"modelExperienceId" validate:"required"` // 体验对话ID
}

func (o *ModelExperienceDialogRecordRequest) Check() error {
	return nil
}
