package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateAssistantTool(ctx context.Context, assistantId uint32, toolId, toolType, actionName string, userId, orgID string) *err_code.Status {
	// 检查是否已存在 [EN] Check if it already exists
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithToolId(toolId),
		sqlopt.WithToolType(toolType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantTool{}).
		Count(&count).Error; err != nil {
		return toErrStatus("assistant_tool_create", err.Error())
	}
	if count > 0 {
		return toErrStatus("assistant_tool_create", "tool already exists")
	}

	err := c.db.WithContext(ctx).Create(&model.AssistantTool{
		AssistantId: assistantId,
		ToolId:      toolId,
		ToolType:    toolType,
		ActionName:  actionName,
		Enable:      true, // 默认打开 [EN] Open by default
		UserId:      userId,
		OrgId:       orgID,
	}).Error

	if err != nil {
		return toErrStatus("assistant_tool_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteAssistantTool(ctx context.Context, assistantId uint32, toolId, toolType, actionName string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithToolId(toolId),
		sqlopt.WithToolType(toolType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantTool{}).Error; err != nil {
		return toErrStatus("assistant_tool_delete", err.Error())
	}
	return nil
}

func (c *Client) UpdateAssistantTool(ctx context.Context, tool *model.AssistantTool) *err_code.Status {
	// 更新 [EN] renew
	result := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(tool.AssistantId),
		sqlopt.WithToolId(tool.ToolId),
		sqlopt.WithToolType(tool.ToolType),
		sqlopt.WithActionName(tool.ActionName),
	).Apply(c.db.WithContext(ctx)).
		Model(&model.AssistantTool{}).
		Updates(map[string]interface{}{
			"enable": tool.Enable,
		})
	if result.Error != nil {
		return toErrStatus("assistant_tool_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("assistant_tool_update", "tool not exists")
	}

	return nil
}

func (c *Client) UpdateAssistantToolConfig(ctx context.Context, assistantId uint32, toolId, toolConfig string) *err_code.Status {
	// 更新 [EN] renew
	result := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithToolId(toolId),
	).Apply(c.db.WithContext(ctx)).
		Model(&model.AssistantTool{}).
		Updates(map[string]interface{}{
			"tool_config": toolConfig,
		})
	if result.Error != nil {
		return toErrStatus("assistant_tool_config_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("assistant_tool_config_update", "tool not exists")
	}

	return nil
}

func (c *Client) GetAssistantTool(ctx context.Context, assistantId uint32, toolId, toolType, actionName string) (*model.AssistantTool, *err_code.Status) {
	tool := &model.AssistantTool{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithToolId(toolId),
		sqlopt.WithToolType(toolType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).First(tool).Error; err != nil {
		return nil, toErrStatus("assistant_tool_get", err.Error())
	}
	return tool, nil
}

func (c *Client) GetAssistantToolList(ctx context.Context, assistantId uint32) ([]*model.AssistantTool, *err_code.Status) {
	var toolList []*model.AssistantTool
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
	).Apply(c.db.WithContext(ctx)).Find(&toolList).Error; err != nil {
		return nil, toErrStatus("assistant_tool_get", err.Error())
	}
	return toolList, nil
}

func (c *Client) DeleteAssistantToolByToolId(ctx context.Context, toolId, toolType string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithToolId(toolId),
		sqlopt.WithToolType(toolType),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantTool{}).Error; err != nil {
		return toErrStatus("assistant_tool_delete", err.Error())
	}
	return nil
}
