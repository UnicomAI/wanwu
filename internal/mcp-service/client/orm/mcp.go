package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	pkg_db "github.com/UnicomAI/wanwu/pkg/db"
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
		// 检查是否已存在来自广场
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
		// 检查是否已存在相同的记录
		if err := sqlopt.SQLOptions(
			sqlopt.WithName(tab.Name),
			sqlopt.WithOrgID(tab.OrgID),
			sqlopt.WithUserID(tab.UserID),
		).Apply(tx).First(&model.MCPClient{}).Error; err == nil {
			return toErrStatus("mcp_create_duplicate_name")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_create_err", err.Error())
		}
		// 创建
		if err := tx.Create(tab).Error; err != nil {
			return toErrStatus("mcp_create_err", err.Error())
		}
		return nil
	})
}

func (c *Client) UpdateMCP(ctx context.Context, tab *model.MCPClient) *errs.Status {
	var info model.MCPClient
	if err := sqlopt.SQLOptions(
		sqlopt.WithName(tab.Name),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).First(&info).Error; err == nil {
		if tab.ID != info.ID {
			return toErrStatus("mcp_update_duplicate_name")
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("mcp_update_err", err.Error())
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
	).Apply(c.db).WithContext(ctx).Model(tab).Updates(map[string]interface{}{
		"name":           tab.Name,
		"from":           tab.From,
		"desc":           tab.Desc,
		"sse_url":        tab.SseUrl,
		"streamable_url": tab.StreamableUrl,
		"transport":      tab.Transport,
		"avatar_path":    tab.AvatarPath,
		"auth_json":      tab.AuthJSON,
		"headers":        tab.Headers,
	}).Error; err != nil {
		return toErrStatus("mcp_update_err", err.Error())
	}
	return nil
}

func (c *Client) DeleteMCP(ctx context.Context, mcpID uint32, userID, orgID string) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(mcpID),
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
	).Apply(c.db).WithContext(ctx).Delete(&model.MCPClient{}).Error; err != nil {
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
	).Apply(c.db).WithContext(ctx).Order("updated_at DESC").Find(&mcpInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return mcpInfos, nil
}

// ListMCPsAdmin 管理员中心跨组织查询自定义MCP列表（SQL分页）。
// orgIDs / userIDs 为空表示不按该维度过滤。pageNo 从1开始，pageSize<=0 不分页。
// 返回当前页数据与总条数。
func (c *Client) ListMCPsAdmin(ctx context.Context, orgIDs, userIDs []string, name string, pageNo, pageSize int) ([]*model.MCPClient, int64, *errs.Status) {
	db := sqlopt.SQLOptions(
		sqlopt.WithOrgIDList(orgIDs),
		sqlopt.WithUserIDList(userIDs),
		sqlopt.LikeName(pkg_db.EscapeLike(name)),
	).Apply(c.db).WithContext(ctx).Model(&model.MCPClient{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("mcp_get_custom_mcp_list_err", err.Error())
	}
	var mcpInfos []*model.MCPClient
	if err := db.Limit(pageSize).
		Offset((pageNo - 1) * pageSize).
		Order("updated_at DESC").
		Find(&mcpInfos).Error; err != nil {
		return nil, 0, toErrStatus("mcp_get_custom_mcp_list_err", err.Error())
	}
	return mcpInfos, total, nil
}

func (c *Client) ListMCPsByMCPIdList(ctx context.Context, mcpIDList []uint32) ([]*model.MCPClient, *errs.Status) {
	var mcpInfos []*model.MCPClient
	if err := sqlopt.WithIDs(mcpIDList).Apply(c.db).WithContext(ctx).Order("id DESC").Find(&mcpInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return mcpInfos, nil
}
