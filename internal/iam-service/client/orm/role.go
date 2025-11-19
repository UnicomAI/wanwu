package orm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access"
	"github.com/gromitlee/access/pkg/perm"
	"gorm.io/gorm"
)

func (c *Client) GetAdminRole(ctx context.Context) (uint32, error) {
	var roleID uint32
	return roleID, c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// top org
		topOrg := &model.Org{}
		if err := sqlopt.WithParentID(0).Apply(c.db.WithContext(ctx)).First(topOrg).Error; err != nil {
			return err
		}
		// admin role
		orgRole := &model.OrgRole{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(topOrg.ID),
			sqlopt.WithAdmin(true),
		).Apply(tx).First(orgRole).Error; err != nil {
			return err
		}
		roleID = orgRole.RoleID
		return nil
	})
}

func (c *Client) GetRole(ctx context.Context, orgID, roleID uint32) (*RoleInfo, *errs.Status) {
	var ret *RoleInfo
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// org role
		orgRole := &model.OrgRole{}
		err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).First(orgRole).Error
		if err != nil {
			return toErrStatus("iam_role_get", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// role
		ret, err = getRoleInfoTx(tx, orgRole)
		if err != nil {
			return toErrStatus("iam_role_get", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		return nil
	})

}

func (c *Client) GetRoles(ctx context.Context, orgID uint32, name string, offset, limit int32) ([]*RoleInfo, int64, *errs.Status) {
	var ret []*RoleInfo
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var orgRoles []*model.OrgRole
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.LikeName(name),
		).Apply(tx).
			Offset(int(offset)).Limit(int(limit)).Order("role_id DESC").Find(&orgRoles).
			Offset(-1).Limit(-1).Count(&count).Error; err != nil {
			return toErrStatus("iam_roles_get", util.Int2Str(orgID), err.Error())
		}
		for _, orgRole := range orgRoles {
			info, err := getRoleInfoTx(tx, orgRole)
			if err != nil {
				return toErrStatus("iam_roles_get", util.Int2Str(orgID), err.Error())
			}
			ret = append(ret, info)
		}
		return nil
	})

}

func (c *Client) SelectRoles(ctx context.Context, orgID uint32) ([]RoleIDName, *errs.Status) {
	var ret []RoleIDName
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var orgRoles []*model.OrgRole
		if err := sqlopt.WithOrgID(orgID).Apply(tx).Find(&orgRoles).Error; err != nil {
			return toErrStatus("iam_roles_select", util.Int2Str(orgID), err.Error())
		}
		for _, orgRole := range orgRoles {
			ret = append(ret, RoleIDName{
				ID:       orgRole.RoleID,
				Name:     orgRole.Name,
				IsAdmin:  orgRole.IsAdmin,
				IsSystem: orgRole.OrgID == config.TopOrgID(),
			})
		}
		return nil
	})

}

func (c *Client) CreateRole(ctx context.Context, orgID, creatorID uint32, name, remark string, perms []perm.Perm) (uint32, error) {
	var roleID uint32
	var err error
	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roleID, err = createRole(tx, orgID, creatorID, name, remark, false, perms)
		return err
	})
	return roleID, err
}

func createRole(tx *gorm.DB, orgID, creatorID uint32, name, remark string, isOrgAdmin bool, perms []perm.Perm) (uint32, error) {
	// check org
	if err := sqlopt.WithID(orgID).Apply(tx).First(&model.Org{}).Error; err != nil {
		return 0, fmt.Errorf("create org %v role check org err: %v", orgID, err)
	}
	// check creator
	var isSysAdmin bool
	if creatorID != 0 {
		// 正常创建角色 [EN] Create role normally
		if err := sqlopt.WithID(creatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return 0, fmt.Errorf("create org %v role check creator %v err: %v", orgID, creatorID, err)
		}
		// check name 角色名在组织内唯一 [EN] check name The role name is unique within the organization
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithName(name),
		).Apply(tx).First(&model.OrgRole{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return 0, fmt.Errorf("create org %v role check name %v err: %v", orgID, name, err)
		}
		// check isOrgAdmin
		if isOrgAdmin {
			if err := sqlopt.WithOrgID(orgID).Apply(tx).First(&model.OrgRole{}).Error; err != gorm.ErrRecordNotFound {
				if err == nil {
					err = errors.New("cannot be org admin role")
				}
				return 0, fmt.Errorf("create org %v role check org admin role err: %v", orgID, err)
			}
		}
	} else {
		// 创建系统顶级组织的管理员角色，此时系统内不能存在任何角色 [EN] Create an administrator role for the top-level organization of the system. No role can exist in the system at this time.
		if err := tx.First(&model.OrgRole{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return 0, fmt.Errorf("create admin role check org role err: %v", err)
		}
		isSysAdmin = true
	}
	// create role
	if roleDetail, err := access.RBAC0CreateRole(tx, 0, int64(creatorID), name, remark, isSysAdmin, perms...); err != nil {
		return 0, fmt.Errorf("create org %v role create role err: %v", orgID, err)
	} else {
		// create org role
		if err := tx.Create(&model.OrgRole{
			OrgID:   orgID,
			RoleID:  uint32(roleDetail.Role),
			IsAdmin: isOrgAdmin,
			Status:  true,
			Name:    name,
		}).Error; err != nil {
			return 0, fmt.Errorf("create org %v role err: %v", orgID, err)
		}
		return uint32(roleDetail.Role), nil
	}
}

func (c *Client) UpdateRole(ctx context.Context, orgID, roleID uint32, name, remark string, perms []perm.Perm) *errs.Status {
	if orgID == 0 || roleID == 0 {
		return toErrStatus("iam_role_update", util.Int2Str(orgID),
			util.Int2Str(roleID), "update role but org id or role id 0")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org role
		orgRole := &model.OrgRole{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).First(orgRole).Error; err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// check name
		var orgRoles []*model.OrgRole
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithName(name),
		).Apply(tx).Find(&orgRoles).Error; err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		for _, orgRole := range orgRoles {
			if orgRole.RoleID != roleID {
				return toErrStatus("iam_role_update", util.Int2Str(orgID),
					util.Int2Str(roleID), "name already exist")
			}
		}
		// update role
		if err := access.RBAC0UpdateRole(tx, perm.Role(roleID), name, remark); err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// update org role
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).Model(&model.OrgRole{}).Updates(map[string]interface{}{
			"name": name,
		}).Error; err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// 组织内置管理员，不修改权限 [EN] The organization has built-in administrators and does not need to modify permissions.
		if orgRole.IsAdmin {
			return nil
		}
		// clear perms
		if err := access.RBAC0CleanRolePerms(tx, perm.Role(roleID)); err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// grant perms
		if len(perms) > 0 {
			if err := access.RBAC0GrantRolePerms(tx, perm.Role(roleID), perms); err != nil {
				return toErrStatus("iam_role_update", util.Int2Str(orgID),
					util.Int2Str(roleID), err.Error())
			}
		}
		return nil
	})

}

func (c *Client) DeleteRole(ctx context.Context, orgID, roleID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org role
		orgRole := &model.OrgRole{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).First(orgRole).Error; err != nil {
			return toErrStatus("iam_role_delete", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// 组织内置管理员，不能被删除 [EN] The organization has built-in administrators and cannot be deleted.
		if orgRole.IsAdmin {
			return toErrStatus("iam_role_delete", util.Int2Str(orgID),
				util.Int2Str(roleID), "cannot delete org admin role")
		}
		// delete user role
		if err := sqlopt.WithRoleID(roleID).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_role_delete", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// delete org role
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).Delete(&model.OrgRole{}).Error; err != nil {
			return toErrStatus("iam_role_delete", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// delete role
		if err := access.RBAC0DeleteRole(tx, perm.Role(roleID)); err != nil {
			return toErrStatus("iam_role_delete", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		return nil
	})

}

func (c *Client) ChangeRoleStatus(ctx context.Context, orgID, roleID uint32, status bool) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org role
		orgRole := &model.OrgRole{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).First(orgRole).Error; err != nil {
			return toErrStatus("iam_role_change", util.Int2Str(orgID),
				util.Int2Str(roleID), strconv.FormatBool(status), err.Error())
		}
		// 组织内置管理员，不能修改状态 [EN] The organization has built-in administrators and cannot modify the status.
		if orgRole.IsAdmin {
			return toErrStatus("iam_role_change", util.Int2Str(orgID),
				util.Int2Str(roleID), strconv.FormatBool(status), "cannot change org admin role status")
		}
		// change org role status
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithRoleID(roleID),
		).Apply(tx).Model(&model.OrgRole{}).Updates(map[string]interface{}{
			"status": status,
		}).Error; err != nil {
			return toErrStatus("iam_role_change", util.Int2Str(orgID),
				util.Int2Str(roleID), strconv.FormatBool(status), err.Error())
		}
		// change role status
		if status {
			if err := access.RBAC0EnableRole(tx, perm.Role(roleID)); err != nil {
				return toErrStatus("iam_role_change", util.Int2Str(orgID),
					util.Int2Str(roleID), strconv.FormatBool(status), err.Error())
			}
			return nil
		}
		if err := access.RBAC0DisableRole(tx, perm.Role(roleID)); err != nil {
			return toErrStatus("iam_role_change", util.Int2Str(orgID),
				util.Int2Str(roleID), strconv.FormatBool(status), err.Error())
		}
		return nil
	})

}

// --- internal function ---

func getRoleInfoTx(tx *gorm.DB, orgRole *model.OrgRole) (*RoleInfo, error) {
	roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(orgRole.RoleID))
	if err != nil {
		return nil, fmt.Errorf("get role %v err: %v", orgRole.RoleID, err)
	}
	return toRoleInfoTx(tx, orgRole, roleDetail)
}

func toRoleInfoTx(tx *gorm.DB, orgRole *model.OrgRole, roleDetail *perm.RolePerms) (*RoleInfo, error) {
	ret := &RoleInfo{
		ID:        uint32(roleDetail.Role),
		IsAdmin:   orgRole.IsAdmin,
		IsSystem:  orgRole.OrgID == config.TopOrgID(),
		Name:      roleDetail.Name,
		Remark:    roleDetail.Desc,
		Status:    roleDetail.Enable,
		CreatedAt: roleDetail.CreatedAt,
	}
	// creator
	if roleDetail.Creator != 0 {
		creator, err := getCreatorTx(tx, uint32(roleDetail.Creator))
		if err != nil {
			return nil, err
		}
		ret.Creator = creator
	}
	// perms
	for _, perm := range roleDetail.Perms {
		ret.Perms = append(ret.Perms, Perm{Perm: string(perm.Obj)})
	}
	return ret, nil
}
