package request

import (
	"fmt"
	"mime/multipart"

	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type WorkflowIDReq struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

func (r *WorkflowIDReq) Check() error {
	return nil
}

type GetWorkflowListReq struct {
	UserId string `form:"userId" json:"userId" validate:"required" `
	OrgId  string `form:"orgId" json:"orgId" validate:"required" `
}

func (g *GetWorkflowListReq) Check() error {
	return nil
}

type CreateWorkflowReq struct {
	AppType string `json:"appType" validate:"required"` // workflow 或 chatflow
	AppBriefConfig
}

func (r *CreateWorkflowReq) Check() error {
	// 校验 appType
	if r.AppType != constant.AppTypeWorkflow && r.AppType != constant.AppTypeChatflow {
		return fmt.Errorf("appType must be 'workflow' or 'chatflow'")
	}
	// 根据 appType 选择 subject
	subject := util.SubjectWorkflow
	if r.AppType == constant.AppTypeChatflow {
		subject = util.SubjectChatflow
	}
	return util.ValidateBriefCreate(&r.Name, &r.Desc, subject)
}

type CreateChatflowReq struct {
	AppBriefConfig
}

func (r *CreateChatflowReq) Check() error {
	return util.ValidateBriefCreate(&r.Name, &r.Desc, util.SubjectChatflow)
}

type CreateWorkflowByTemplateReq struct {
	TemplateId string `json:"templateId" validate:"required"`
	AppBriefConfig
}

func (r *CreateWorkflowByTemplateReq) Check() error {
	return util.ValidateBriefCreate(&r.Name, &r.Desc, util.SubjectWorkflow)
}

type WorkflowUploadFileReq struct {
	File *multipart.FileHeader `form:"file" json:"file" validate:"required"` // 二进制格式
}

func (u *WorkflowUploadFileReq) Check() error {
	return nil
}

type WorkflowConvertReq struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

func (r *WorkflowConvertReq) Check() error {
	return nil
}

type WorkflowRunReq struct {
	WorkflowID string         `json:"workflow_id" validate:"required"`
	Input      map[string]any `json:"input" `
}

func (r *WorkflowRunReq) Check() error {
	return nil
}

type ChatflowApplicationListReq struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
}

func (r *ChatflowApplicationListReq) Check() error {
	return nil
}

type ChatflowApplicationInfoReq struct {
	IntelligenceID   string `json:"intelligence_id" validate:"required"`
	IntelligenceType int64  `json:"intelligence_type" validate:"required"`
}

func (r *ChatflowApplicationInfoReq) Check() error {
	return nil
}

type ChatflowConversationCreateReq struct {
	WorkflowID       string `json:"workflow_id" validate:"required"`
	AppID            string `json:"app_id" validate:"required"`
	ConnectorID      string `json:"connector_id" `
	ConversationName string `json:"conversation_name" validate:"required"`
	DraftMode        bool   `json:"draft_mode"`
	GetOrCreate      bool   `json:"get_or_create"`
}

func (r *ChatflowConversationCreateReq) Check() error {
	return nil
}

type ChatflowConversationDeleteReq struct {
	ProjectId string `json:"project_id" validate:"required"`
	UniqueId  string `json:"unique_id" validate:"required"`
}

func (r *ChatflowConversationDeleteReq) Check() error {
	return nil
}
