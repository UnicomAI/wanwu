package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CheckMCPExist(ctx context.Context, orgID, userID, mcpSquareID string) (bool, *errs.Status) {
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
		sqlopt.WithMcpSquareId(mcpSquareID),
	).Apply(c.db).WithContext(ctx).First(&model.MCPClient{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, toErrStatus("mpc_check_exist_err", err.Error())
	}
	return true, nil
}

func (c *Client) GetMCP(ctx context.Context, mcpID uint32) (*model.MCPClient, *errs.Status) {
	info := &model.MCPClient{}
	if err := sqlopt.WithID(mcpID).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("mcp_get_err", err.Error())
	}
	return info, nil
}

func (c *Client) CreateMCP(ctx context.Context, tab *model.MCPClient) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// Check if already exists from square
		if tab.McpSquareId != "" {
			if err := sqlopt.SQLOptions(
				sqlopt.WithMcpSquareId(tab.McpSquareId),
				sqlopt.WithOrgID(tab.OrgID),
				sqlopt.WithUserID(tab.UserID),
			).Apply(tx).First(&model.MCPClient{}).Error; err == nil {
				return toErrStatus("mcp_create_duplicate_square")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("mcp_create_err", err.Error())
			}
		}
		// Check if the same record already exists
		if err := sqlopt.SQLOptions(
			sqlopt.WithName(tab.Name),
			sqlopt.WithOrgID(tab.OrgID),
			sqlopt.WithUserID(tab.UserID),
		).Apply(tx).First(&model.MCPClient{}).Error; err == nil {
			return toErrStatus("mcp_create_duplicate_name")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_create_err", err.Error())
		}
		// create
		if err := tx.Create(tab).Error; err != nil {
			return toErrStatus("mcp_create_err", err.Error())
		}
		return nil
	})
}

func (c *Client) UpdateMCP(ctx context.Context, tab *model.MCPClient) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
	).Apply(c.db).WithContext(ctx).Model(tab).Updates(map[string]interface{}{
		"name":        tab.Name,
		"from":        tab.From,
		"desc":        tab.Desc,
		"sse_url":     tab.SseUrl,
		"avatar_path": tab.AvatarPath,
	}).Error; err != nil {
		return toErrStatus("mcp_update_err", err.Error())
	}
	return nil
}

func (c *Client) DeleteMCP(ctx context.Context, mcpID uint32) *errs.Status {
	if err := sqlopt.WithID(mcpID).Apply(c.db).WithContext(ctx).Delete(&model.MCPClient{}).Error; err != nil {
		return toErrStatus("mcp_delete_err", err.Error())
	}
	return nil
}

func (c *Client) ListMCPs(ctx context.Context, orgID, userID, name string) ([]*model.MCPClient, *errs.Status) {
	var mcpInfos []*model.MCPClient
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Order("id DESC").Find(&mcpInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return mcpInfos, nil
}

func (c *Client) ListMCPsByMCPIdList(ctx context.Context, mcpIDList []uint32) ([]*model.MCPClient, *errs.Status) {
	var mcpInfos []*model.MCPClient
	if err := sqlopt.WithIDs(mcpIDList).Apply(c.db).WithContext(ctx).Order("id DESC").Find(&mcpInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return mcpInfos, nil
}
