package orm

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access"
	"github.com/gromitlee/access/pkg/perm"
	"gorm.io/gorm"
)

// GetGlobalRole 获取全局角色详情
// orgID 为统计范围：0 或顶级组织(TopOrgID)时统计全系统关联用户；否则只统计 user_roles.org_id == orgID 的本组织关联用户。
func (c *Client) GetGlobalRole(ctx context.Context, orgID, roleID uint32) (*RoleInfo, *errs.Status) {
	var ret *RoleInfo
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		globalRole := &model.GlobalRole{}
		if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
			return toErrStatus("iam_global_role_get", util.Int2Str(roleID), err.Error())
		}
		roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(globalRole.RoleID))
		if err != nil {
			return toErrStatus("iam_global_role_get", util.Int2Str(roleID), err.Error())
		}
		// user count（按 orgID 范围统计关联用户）
		counts, err := getRoleUserCountsTx(tx, []uint32{globalRole.RoleID}, orgID)
		if err != nil {
			return toErrStatus("iam_global_role_get", util.Int2Str(roleID), err.Error())
		}
		ret = &RoleInfo{
			ID:         uint32(roleDetail.Role),
			IsAdmin:    false,
			IsSystem:   true,
			IsGlobal:   true,
			Name:       roleDetail.Name,
			Remark:     roleDetail.Desc,
			Status:     roleDetail.Enable,
			CreatedAt:  roleDetail.CreatedAt,
			AvatarPath: globalRole.AvatarPath,
			UserCount:  counts[globalRole.RoleID],
		}
		if roleDetail.Creator != 0 {
			creator, err := getCreatorTx(tx, uint32(roleDetail.Creator))
			if err != nil {
				return toErrStatus("iam_global_role_get", util.Int2Str(roleID), err.Error())
			}
			ret.Creator = creator
		}
		for _, p := range roleDetail.Perms {
			ret.Perms = append(ret.Perms, Perm{Perm: string(p.Obj)})
		}
		return nil
	})
}

// GetGlobalRoles 获取全局角色列表（分页）
// orgID 为统计范围：0 或顶级组织(TopOrgID)时统计全系统关联用户；否则只统计 user_roles.org_id == orgID 的本组织关联用户。
func (c *Client) GetGlobalRoles(ctx context.Context, orgID uint32, name string, offset, limit int32) ([]*RoleInfo, int64, *errs.Status) {
	var ret []*RoleInfo
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var globalRoles []*model.GlobalRole
		query := tx
		if name != "" {
			query = sqlopt.LikeName(name).Apply(query)
		}
		if err := query.
			Offset(int(offset)).Limit(int(limit)).Order("role_id DESC").Find(&globalRoles).
			Offset(-1).Limit(-1).Count(&count).Error; err != nil {
			return toErrStatus("iam_global_roles_get", err.Error())
		}
		// user counts（批量预取，避免 N+1，按 orgID 范围统计）
		roleIDs := make([]uint32, 0, len(globalRoles))
		for _, gr := range globalRoles {
			roleIDs = append(roleIDs, gr.RoleID)
		}
		counts, err := getRoleUserCountsTx(tx, roleIDs, orgID)
		if err != nil {
			return toErrStatus("iam_global_roles_get", err.Error())
		}
		for _, gr := range globalRoles {
			roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(gr.RoleID))
			if err != nil {
				return toErrStatus("iam_global_roles_get", err.Error())
			}
			info := &RoleInfo{
				ID:         uint32(roleDetail.Role),
				IsAdmin:    false,
				IsSystem:   true,
				IsGlobal:   true,
				Name:       roleDetail.Name,
				Remark:     roleDetail.Desc,
				Status:     roleDetail.Enable,
				CreatedAt:  roleDetail.CreatedAt,
				AvatarPath: gr.AvatarPath,
				UserCount:  counts[gr.RoleID],
			}
			if roleDetail.Creator != 0 {
				creator, err := getCreatorTx(tx, uint32(roleDetail.Creator))
				if err != nil {
					return toErrStatus("iam_global_roles_get", err.Error())
				}
				info.Creator = creator
			}
			for _, p := range roleDetail.Perms {
				info.Perms = append(info.Perms, Perm{Perm: string(p.Obj)})
			}
			ret = append(ret, info)
		}
		return nil
	})
}

// SelectGlobalRoles 获取全局角色下拉列表（仅启用的）
func (c *Client) SelectGlobalRoles(ctx context.Context) ([]RoleIDName, *errs.Status) {
	var ret []RoleIDName
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var globalRoles []*model.GlobalRole
		if err := sqlopt.WithStatus(true).Apply(tx).Find(&globalRoles).Error; err != nil {
			return toErrStatus("iam_global_roles_select", err.Error())
		}
		for _, gr := range globalRoles {
			ret = append(ret, RoleIDName{
				ID:         gr.RoleID,
				Name:       gr.Name,
				AvatarPath: gr.AvatarPath,
				IsAdmin:    false,
				IsSystem:   true,
				IsGlobal:   true,
			})
		}
		return nil
	})
}

// CreateGlobalRole 创建全局角色
func (c *Client) CreateGlobalRole(ctx context.Context, creatorID uint32, name, remark, avatarPath string, perms []perm.Perm) (uint32, error) {
	var roleID uint32
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// check name unique in global roles
		var existing model.GlobalRole
		if err := sqlopt.WithName(name).Apply(tx).First(&existing).Error; err == nil {
			return fmt.Errorf("global role name %v already exist", name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check global role name err: %v", err)
		}
		// create role in RBAC0（全局角色永不是管理员，传 isAdmin=false 并直接写入 perms）。
		// 注意：不要传 isAdmin=true 来“跳过 perms 再事后 grant”——那会把 roles.is_admin 写成 true，
		// 进而在 CheckUserPerm 鉴权时被 RBAC0GetRolePerms 读回，触发 "role X is admin" 拒绝登录。
		roleDetail, err := access.RBAC0CreateRole(tx, 0, int64(creatorID), name, remark, false, perms...)
		if err != nil {
			return fmt.Errorf("create global role in RBAC0 err: %v", err)
		}
		roleID = uint32(roleDetail.Role)
		// create global role record
		if err := tx.Create(&model.GlobalRole{
			RoleID:     roleID,
			Name:       name,
			Status:     true,
			CreatorID:  creatorID,
			AvatarPath: avatarPath,
		}).Error; err != nil {
			return fmt.Errorf("create global role err: %v", err)
		}
		return nil
	})
	return roleID, err
}

// UpdateGlobalRole 更新全局角色
func (c *Client) UpdateGlobalRole(ctx context.Context, roleID uint32, name, remark, avatarPath string, perms []perm.Perm) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check global role exists
		globalRole := &model.GlobalRole{}
		if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
			return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
		}
		// check name unique
		var existingRoles []*model.GlobalRole
		if err := sqlopt.WithName(name).Apply(tx).Find(&existingRoles).Error; err != nil {
			return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
		}
		for _, r := range existingRoles {
			if r.RoleID != roleID {
				return toErrStatus("iam_global_role_update", util.Int2Str(roleID), "name already exist")
			}
		}
		// update role in RBAC0
		if err := access.RBAC0UpdateRole(tx, perm.Role(roleID), name, remark); err != nil {
			return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
		}
		// update name in global_roles table
		if err := sqlopt.WithRoleID(roleID).Apply(tx).Model(&model.GlobalRole{}).Updates(map[string]interface{}{
			"name":        name,
			"avatar_path": avatarPath,
		}).Error; err != nil {
			return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
		}
		// clear and re-grant perms
		if err := access.RBAC0CleanRolePerms(tx, perm.Role(roleID)); err != nil {
			return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
		}
		if len(perms) > 0 {
			if err := access.RBAC0GrantRolePerms(tx, perm.Role(roleID), perms); err != nil {
				return toErrStatus("iam_global_role_update", util.Int2Str(roleID), err.Error())
			}
		}
		return nil
	})
}

// DeleteGlobalRole 删除全局角色
func (c *Client) DeleteGlobalRole(ctx context.Context, roleID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check global role exists
		globalRole := &model.GlobalRole{}
		if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
			return toErrStatus("iam_global_role_delete", util.Int2Str(roleID), err.Error())
		}
		// delete user role associations for this global role across all orgs
		if err := sqlopt.WithRoleID(roleID).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_global_role_delete", util.Int2Str(roleID), err.Error())
		}
		// delete global role record
		if err := sqlopt.WithRoleID(roleID).Apply(tx).Delete(&model.GlobalRole{}).Error; err != nil {
			return toErrStatus("iam_global_role_delete", util.Int2Str(roleID), err.Error())
		}
		// delete role in RBAC0
		if err := access.RBAC0DeleteRole(tx, perm.Role(roleID)); err != nil {
			return toErrStatus("iam_global_role_delete", util.Int2Str(roleID), err.Error())
		}
		return nil
	})
}

// ChangeGlobalRoleStatus 修改全局角色状态
func (c *Client) ChangeGlobalRoleStatus(ctx context.Context, roleID uint32, status bool) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check global role exists
		globalRole := &model.GlobalRole{}
		if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
			return toErrStatus("iam_global_role_change_status", util.Int2Str(roleID), err.Error())
		}
		// change status in global_roles table
		if err := sqlopt.WithRoleID(roleID).Apply(tx).Model(&model.GlobalRole{}).Updates(map[string]interface{}{
			"status": status,
		}).Error; err != nil {
			return toErrStatus("iam_global_role_change_status", util.Int2Str(roleID), err.Error())
		}
		// change status in RBAC0
		if status {
			if err := access.RBAC0EnableRole(tx, perm.Role(roleID)); err != nil {
				return toErrStatus("iam_global_role_change_status", util.Int2Str(roleID), err.Error())
			}
		} else {
			if err := access.RBAC0DisableRole(tx, perm.Role(roleID)); err != nil {
				return toErrStatus("iam_global_role_change_status", util.Int2Str(roleID), err.Error())
			}
		}
		return nil
	})
}

// IsGlobalRole 检查 roleID 是否为全局角色
func (c *Client) IsGlobalRole(ctx context.Context, roleID uint32) (bool, error) {
	globalRole := &model.GlobalRole{}
	err := sqlopt.WithRoleID(roleID).Apply(c.db.WithContext(ctx)).First(globalRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
