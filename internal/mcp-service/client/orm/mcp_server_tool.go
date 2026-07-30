package orm

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
)

func (c *Client) GetMCPServerTool(ctx context.Context, mcpServerToolId string) (*model.MCPServerTool, *errs.Status) {
	info := &model.MCPServerTool{}
	if err := sqlopt.WithMcpServerToolId(mcpServerToolId).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("mcp_get_mcp_server_tool_info_err", err.Error())
	}
	return info, nil
}

func (c *Client) CreateMCPServerTool(ctx context.Context, mcpServerTools []*model.MCPServerTool) *errs.Status {
	if len(mcpServerTools) > 0 {
		if err := c.db.WithContext(ctx).CreateInBatches(mcpServerTools, len(mcpServerTools)).Error; err != nil {
			return toErrStatus("mcp_create_mcp_server_tool_err", err.Error())
		}
	}
	return nil
}

func (c *Client) UpdateMCPServerTool(ctx context.Context, mcpServerTool *model.MCPServerTool) *errs.Status {
	if err := sqlopt.WithMcpServerToolId(mcpServerTool.MCPServerToolId).Apply(c.db).WithContext(ctx).Model(mcpServerTool).Updates(map[string]interface{}{
		"name":        mcpServerTool.Name,
		"description": mcpServerTool.Description,
		"schema":      mcpServerTool.Schema,
	}).Error; err != nil {
		return toErrStatus("mcp_update_mcp_server_tool_err", err.Error())
	}
	return nil
}

func (c *Client) DeleteMCPServerTool(ctx context.Context, mcpServerToolId, userID, orgID string) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithMcpServerToolId(mcpServerToolId),
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
	).Apply(c.db).WithContext(ctx).Delete(&model.MCPServerTool{}).Error; err != nil {
		return toErrStatus("mcp_delete_mcp_server_tool_err", err.Error())
	}
	return nil
}

func (c *Client) ListMCPServerTools(ctx context.Context, mcpServerId string) ([]*model.MCPServerTool, *errs.Status) {
	var mcpServerTools []*model.MCPServerTool
	if err := sqlopt.SQLOptions(
		sqlopt.WithMcpServerId(mcpServerId),
	).Apply(c.db).WithContext(ctx).Find(&mcpServerTools).Error; err != nil {
		return nil, toErrStatus("mcp_get_mcp_server_tool_list_err", err.Error())
	}
	return mcpServerTools, nil
}

func (c *Client) CountMCPServerTools(ctx context.Context, mcpServerId string) (int64, *errs.Status) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithMcpServerId(mcpServerId),
	).Apply(c.db).WithContext(ctx).Model(model.MCPServerTool{}).Count(&count).Error; err != nil {
		return count, toErrStatus("mcp_get_mcp_server_tool_count_err", err.Error())
	}
	return count, nil
}
