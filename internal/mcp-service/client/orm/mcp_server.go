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

func (c *Client) CreateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status {
	// 检查是否已存在相同的记录
	if err := sqlopt.SQLOptions(
		sqlopt.WithName(mcpServer.Name),
		sqlopt.WithOrgID(mcpServer.OrgID),
		sqlopt.WithUserID(mcpServer.UserID),
	).Apply(c.db).WithContext(ctx).First(&model.MCPServer{}).Error; err == nil {
		return toErrStatus("mcp_create_duplicate_mcp_server_err")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("mcp_create_mcp_server_err", err.Error())
	}
	// 创建 mcp server
	if err := c.db.Create(mcpServer).Error; err != nil {
		return toErrStatus("mcp_create_mcp_server_err", err.Error())
	}
	return nil
}

func (c *Client) UpdateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status {
	var mcpServerInfo model.MCPServer
	if err := sqlopt.SQLOptions(
		sqlopt.WithName(mcpServer.Name),
		sqlopt.WithOrgID(mcpServer.OrgID),
		sqlopt.WithUserID(mcpServer.UserID),
	).Apply(c.db).First(&mcpServerInfo).Error; err == nil {
		if mcpServerInfo.MCPServerID != mcpServer.MCPServerID {
			return toErrStatus("mcp_update_duplicate_mcp_server_err")
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("mcp_update_mcp_server_err", err.Error())
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithMcpServerId(mcpServer.MCPServerID),
	).Apply(c.db).WithContext(ctx).Model(mcpServer).Updates(map[string]interface{}{
		"name":        mcpServer.Name,
		"description": mcpServer.Description,
		"avatar_path": mcpServer.AvatarPath,
	}).Error; err != nil {
		return toErrStatus("mcp_update_mcp_server_err", err.Error())
	}
	return nil

}

func (c *Client) GetMCPServer(ctx context.Context, mcpServerId string) (*model.MCPServer, *errs.Status) {
	info := &model.MCPServer{}
	if err := sqlopt.WithMcpServerId(mcpServerId).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("mcp_get_mcp_server_info_err", err.Error())
	}
	return info, nil
}

func (c *Client) ListMCPServers(ctx context.Context, orgID, userID, name string) ([]*model.MCPServer, *errs.Status) {
	var mcpServerInfos []*model.MCPServer
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Order("updated_at desc").Find(&mcpServerInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_mcp_server_list_err", err.Error())
	}
	return mcpServerInfos, nil
}

// ListMCPServersAdmin 管理员中心跨组织查询MCP服务列表（SQL分页）。
// orgIDs / userIDs 为空表示不按该维度过滤。pageNo 从1开始，pageSize<=0 不分页。
// 返回当前页数据与总条数。
func (c *Client) ListMCPServersAdmin(ctx context.Context, orgIDs, userIDs []string, name string, pageNo, pageSize int) ([]*model.MCPServer, int64, *errs.Status) {
	db := sqlopt.SQLOptions(
		sqlopt.WithOrgIDList(orgIDs),
		sqlopt.WithUserIDList(userIDs),
		sqlopt.LikeName(pkg_db.EscapeLike(name)),
	).Apply(c.db).WithContext(ctx).Model(&model.MCPServer{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("mcp_get_mcp_server_list_err", err.Error())
	}
	var mcpServerInfos []*model.MCPServer
	if err := db.Limit(pageSize).
		Offset((pageNo - 1) * pageSize).
		Order("updated_at desc").
		Find(&mcpServerInfos).Error; err != nil {
		return nil, 0, toErrStatus("mcp_get_mcp_server_list_err", err.Error())
	}
	return mcpServerInfos, total, nil
}

func (c *Client) DeleteMCPServer(ctx context.Context, mcpServerId, userID, orgID string) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		if err := sqlopt.SQLOptions(
			sqlopt.WithMcpServerId(mcpServerId),
			sqlopt.WithUserID(userID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).WithContext(ctx).Delete(&model.MCPServer{}).Error; err != nil {
			return toErrStatus("mcp_delete_mcp_server_err", err.Error())
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithMcpServerId(mcpServerId),
			sqlopt.WithUserID(userID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).WithContext(ctx).Delete(&model.MCPServerTool{}).Error; err != nil {
			return toErrStatus("mcp_delete_mcp_server_tool_err", err.Error())
		}
		return nil
	})
}

func (c *Client) ListMCPServerByIdList(ctx context.Context, mcpServerIdList []string) ([]*model.MCPServer, *errs.Status) {
	var mcpServerList []*model.MCPServer
	if err := sqlopt.WithMcpServerIdList(mcpServerIdList).Apply(c.db).WithContext(ctx).Find(&mcpServerList).Error; err != nil {
		return nil, toErrStatus("mcp_get_mcp_server_tool_list_err", err.Error())
	}
	return mcpServerList, nil
}
