package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateAssistantWorkflow(ctx context.Context, workflow *model.AssistantWorkflow) *err_code.Status {
	// Check if it already exists
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(workflow.AssistantId),
		sqlopt.WithWorkflowID(workflow.WorkflowId),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantWorkflow{}).
		Count(&count).Error; err != nil {
		return toErrStatus("assistant_workflow_create", err.Error())
	}
	if count > 0 {
		return toErrStatus("assistant_workflow_create", "workflow already exists")
	}

	// Create Workflow
	err := c.db.WithContext(ctx).Create(workflow).Error
	if err != nil {
		return toErrStatus("assistant_workflow_create", err.Error())
	}
	return nil
}

func (c *Client) UpdateAssistantWorkflow(ctx context.Context, workflow *model.AssistantWorkflow) *err_code.Status {
	result := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(workflow.AssistantId),
		sqlopt.WithWorkflowID(workflow.WorkflowId),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantWorkflow{}).Updates(map[string]interface{}{
		"enable": workflow.Enable,
	})
	if result.Error != nil {
		return toErrStatus("assistant_workflow_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("assistant_workflow_update", "workflow not exists")
	}
	return nil
}

func (c *Client) GetAssistantWorkflow(ctx context.Context, assistantId uint32, workflowId string) (*model.AssistantWorkflow, *err_code.Status) {
	workflow := &model.AssistantWorkflow{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithWorkflowID(workflowId),
	).Apply(c.db.WithContext(ctx)).First(workflow).Error; err != nil {
		return nil, toErrStatus("assistant_workflow_get", err.Error())
	}
	return workflow, nil
}

func (c *Client) GetAssistantWorkflowsByAssistantID(ctx context.Context, assistantId uint32) ([]*model.AssistantWorkflow, *err_code.Status) {
	var workflows []*model.AssistantWorkflow
	if err := sqlopt.WithAssistantID(assistantId).Apply(c.db.WithContext(ctx)).Find(&workflows).Error; err != nil {
		return nil, toErrStatus("assistant_workflows_get_by_assistant_id", err.Error())
	}
	return workflows, nil
}

func (c *Client) DeleteAssistantWorkflow(ctx context.Context, assistantId uint32, workflowId string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithWorkflowID(workflowId),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantWorkflow{}).Error; err != nil {
		return toErrStatus("assistant_workflow_delete", err.Error())
	}
	return nil
}

func (c *Client) DeleteAssistantWorkflowByWorkflowId(ctx context.Context, workflowId string) *err_code.Status {
	if err := sqlopt.WithWorkflowID(workflowId).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantWorkflow{}).Error; err != nil {
		return toErrStatus("assistant_workflow_delete", err.Error())
	}
	return nil
}
