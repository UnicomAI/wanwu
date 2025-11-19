package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
)

func (c *Client) CreateAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string, userId, orgID string) *err_code.Status {
	// 检查是否已存在 [EN] Check if it already exists
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithMCPID(mcpId),
		sqlopt.WithMCPType(mcpType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantMCP{}).
		Count(&count).Error; err != nil {
		return toErrStatus("assistant_mcp_create", err.Error())
	}
	if count > 0 {
		return toErrStatus("assistant_mcp_create", "mcp already exists")
	}
	err := c.db.WithContext(ctx).Create(&model.AssistantMCP{
		AssistantId: assistantId,
		MCPId:       mcpId,
		MCPType:     mcpType,
		ActionName:  actionName,
		Enable:      true, // 默认开 [EN] On by default
		UserId:      userId,
		OrgId:       orgID,
	}).Error
	if err != nil {
		return toErrStatus("assistant_mcp_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithMCPID(mcpId),
		sqlopt.WithMCPType(mcpType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantMCP{}).Error; err != nil {
		return toErrStatus("assistant_mcp_delete", err.Error())
	}
	return nil
}

func (c *Client) DeleteAssistantMCPByMCPId(ctx context.Context, mcpId, mcpType string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithMCPID(mcpId),
		sqlopt.WithMCPType(mcpType),
	).Apply(c.db.WithContext(ctx)).Delete(&model.AssistantMCP{}).Error; err != nil {
		return toErrStatus("assistant_mcp_delete", err.Error())
	}
	return nil
}

func (c *Client) GetAssistantMCP(ctx context.Context, assistantId uint32, mcpId, mcpType, actionName string) (*model.AssistantMCP, *err_code.Status) {
	mcp := &model.AssistantMCP{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(assistantId),
		sqlopt.WithMCPID(mcpId),
		sqlopt.WithMCPType(mcpType),
		sqlopt.WithActionName(actionName),
	).Apply(c.db.WithContext(ctx)).First(mcp).Error; err != nil {
		return nil, toErrStatus("assistant_mcp_get", err.Error())
	}
	return mcp, nil
}

func (c *Client) GetAssistantMCPList(ctx context.Context, assistantId uint32) ([]*model.AssistantMCP, *err_code.Status) {
	var mcpList []*model.AssistantMCP
	if err := sqlopt.WithAssistantID(assistantId).Apply(c.db.WithContext(ctx)).Find(&mcpList).Error; err != nil {
		return nil, toErrStatus("assistant_mcp_list", err.Error())
	}
	return mcpList, nil
}

func (c *Client) UpdateAssistantMCP(ctx context.Context, mcp *model.AssistantMCP) *err_code.Status {
	result := sqlopt.SQLOptions(
		sqlopt.WithAssistantID(mcp.AssistantId),
		sqlopt.WithMCPID(mcp.MCPId),
		sqlopt.WithMCPType(mcp.MCPType),
		sqlopt.WithActionName(mcp.ActionName),
	).Apply(c.db.WithContext(ctx)).Model(&model.AssistantMCP{}).Updates(map[string]interface{}{
		"enable": mcp.Enable,
	})
	if result.Error != nil {
		return toErrStatus("assistant_mcp_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("assistant_mcp_update", "mcp not exists")
	}
	return nil
}
