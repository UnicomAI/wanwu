package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateCustomTool(ctx context.Context, customTool *model.CustomTool) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// Check if the same record already exists
		if err := sqlopt.SQLOptions(
			sqlopt.WithName(customTool.Name),
			sqlopt.WithOrgID(customTool.OrgID),
			sqlopt.WithUserID(customTool.UserID),
		).Apply(tx).First(&model.CustomTool{}).Error; err == nil {
			return toErrStatus("mcp_create_duplicate_custom_tool")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_create_custom_tool_err", err.Error())
		}
		// create
		if err := tx.Create(customTool).Error; err != nil {
			return toErrStatus("mcp_create_custom_tool_err", err.Error())
		}
		return nil
	})
}

func (c *Client) GetCustomTool(ctx context.Context, customTool *model.CustomTool) (*model.CustomTool, *err_code.Status) {
	info := &model.CustomTool{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(customTool.ID),
		sqlopt.WithOrgID(customTool.OrgID),
		sqlopt.WithUserID(customTool.UserID),
	).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_info_err", err.Error())
	}
	return info, nil
}

func (c *Client) ListCustomTools(ctx context.Context, orgID, userID, name string) ([]*model.CustomTool, *err_code.Status) {
	var customToolInfos []*model.CustomTool
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Order("updated_at desc").Find(&customToolInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return customToolInfos, nil
}

func (c *Client) ListCustomToolsByCustomToolIDs(ctx context.Context, ids []uint32) ([]*model.CustomTool, *err_code.Status) {
	var customToolInfos []*model.CustomTool
	if err := sqlopt.WithIDs(ids).Apply(c.db).WithContext(ctx).Find(&customToolInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return customToolInfos, nil
}

func (c *Client) UpdateCustomTool(ctx context.Context, customTool *model.CustomTool) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// Check if the same record already exists
		if customTool.ToolSquareId == "" {
			var dbCustomToolInfo model.CustomTool
			if err := sqlopt.SQLOptions(
				sqlopt.WithName(customTool.Name),
				sqlopt.WithOrgID(customTool.OrgID),
				sqlopt.WithUserID(customTool.UserID),
			).Apply(tx).First(&dbCustomToolInfo).Error; err == nil {
				if dbCustomToolInfo.ID != customTool.ID {
					return toErrStatus("mcp_update_custom_tool_err", "custom tool name already exists")
				}
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("mcp_update_custom_tool_err", err.Error())
			}
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(customTool.ID),
		).Apply(c.db).WithContext(ctx).Model(customTool).Updates(map[string]interface{}{
			"name":           customTool.Name,
			"avatar_path":    customTool.AvatarPath,
			"description":    customTool.Description,
			"schema":         customTool.Schema,
			"privacy_policy": customTool.PrivacyPolicy,
			"auth_json":      customTool.AuthJSON,
			"api_key":        customTool.APIKey,
		}).Error; err != nil {
			return toErrStatus("mcp_update_custom_tool_err", err.Error())
		}
		return nil
	})
}

func (c *Client) DeleteCustomTool(ctx context.Context, ID uint32) *err_code.Status {
	if err := sqlopt.WithID(ID).Apply(c.db).WithContext(ctx).Delete(&model.CustomTool{}).Error; err != nil {
		return toErrStatus("mcp_delete_custom_tool_err", err.Error())
	}
	return nil
}

func (c *Client) ListBuiltinTools(ctx context.Context, orgID, userID string) ([]*model.BuiltinTool, *err_code.Status) {
	var builtinToolInfos []*model.BuiltinTool
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
	).Apply(c.db).WithContext(ctx).Find(&builtinToolInfos).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_list_err", err.Error())
	}
	return builtinToolInfos, nil
}

func (c *Client) GetBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) (*model.BuiltinTool, *err_code.Status) {
	info := &model.BuiltinTool{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithToolSquareID(builtinTool.ToolSquareId),
		sqlopt.WithOrgID(builtinTool.OrgID),
		sqlopt.WithUserID(builtinTool.UserID),
	).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("mcp_get_custom_tool_info_err", err.Error())
	}
	return info, nil
}

func (c *Client) UpdateBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// Check if the record exists
		var dbBuiltinToolInfo model.BuiltinTool
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(builtinTool.ID),
			sqlopt.WithOrgID(builtinTool.OrgID),
			sqlopt.WithUserID(builtinTool.UserID),
		).Apply(tx).First(&dbBuiltinToolInfo).Error; err != nil {
			return toErrStatus("mcp_update_builtin_tool_err", err.Error())
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(builtinTool.ID),
		).Apply(c.db).WithContext(ctx).Model(builtinTool).Updates(map[string]interface{}{
			"auth_json": builtinTool.AuthJSON,
		}).Error; err != nil {
			return toErrStatus("mcp_update_custom_tool_err", err.Error())
		}
		return nil
	})
}

func (c *Client) CreateBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// Check if the same record already exists
		if err := sqlopt.SQLOptions(
			sqlopt.WithToolSquareID(builtinTool.ToolSquareId),
			sqlopt.WithOrgID(builtinTool.OrgID),
			sqlopt.WithUserID(builtinTool.UserID),
		).Apply(tx).First(&model.BuiltinTool{}).Error; err == nil {
			return toErrStatus("mcp_create_duplicate_builtin_tool")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_create_builtin_tool_err", err.Error())
		}
		// create
		if err := tx.Create(builtinTool).Error; err != nil {
			return toErrStatus("mcp_create_builtin_tool_err", err.Error())
		}
		return nil
	})
}
