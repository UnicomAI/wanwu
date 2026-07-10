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
		// org tree
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_role_get", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// user count
		counts, err := getRoleUserCountsTx(tx, []uint32{orgRole.RoleID}, orgID)
		if err != nil {
			return toErrStatus("iam_role_get", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// role
		ret, err = getRoleInfoTx(tx, orgRole, orgTree, counts[orgRole.RoleID])
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
		// org tree（循环外预取，避免 N+1）
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_roles_get", util.Int2Str(orgID), err.Error())
		}
		// user counts（批量预取，避免 N+1）
		roleIDs := make([]uint32, 0, len(orgRoles))
		for _, orgRole := range orgRoles {
			roleIDs = append(roleIDs, orgRole.RoleID)
		}
		counts, err := getRoleUserCountsTx(tx, roleIDs, orgID)
		if err != nil {
			return toErrStatus("iam_roles_get", util.Int2Str(orgID), err.Error())
		}
		for _, orgRole := range orgRoles {
			info, err := getRoleInfoTx(tx, orgRole, orgTree, counts[orgRole.RoleID])
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

func (c *Client) CreateRole(ctx context.Context, orgID, creatorID uint32, name, remark, avatarPath string, perms []perm.Perm) (uint32, error) {
	var roleID uint32
	var err error
	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roleID, err = createRole(tx, orgID, creatorID, name, remark, avatarPath, false, perms)
		return err
	})
	return roleID, err
}

func createRole(tx *gorm.DB, orgID, creatorID uint32, name, remark, avatarPath string, isOrgAdmin bool, perms []perm.Perm) (uint32, error) {
	// check org
	if err := sqlopt.WithID(orgID).Apply(tx).First(&model.Org{}).Error; err != nil {
		return 0, fmt.Errorf("create org %v role check org err: %v", orgID, err)
	}
	// check creator
	var isSysAdmin bool
	if creatorID != 0 {
		// 正常创建角色
		if err := sqlopt.WithID(creatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return 0, fmt.Errorf("create org %v role check creator %v err: %v", orgID, creatorID, err)
		}
		// check name 角色名在组织内唯一
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
		// 创建系统顶级组织的管理员角色，此时系统内不能存在任何角色
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
			OrgID:      orgID,
			RoleID:     uint32(roleDetail.Role),
			IsAdmin:    isOrgAdmin,
			Status:     true,
			Name:       name,
			AvatarPath: avatarPath,
		}).Error; err != nil {
			return 0, fmt.Errorf("create org %v role err: %v", orgID, err)
		}
		return uint32(roleDetail.Role), nil
	}
}

func (c *Client) UpdateRole(ctx context.Context, orgID, roleID uint32, name, remark, avatarPath string, perms []perm.Perm) *errs.Status {
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
			"name":        name,
			"avatar_path": avatarPath,
		}).Error; err != nil {
			return toErrStatus("iam_role_update", util.Int2Str(orgID),
				util.Int2Str(roleID), err.Error())
		}
		// 组织内置管理员，不修改权限
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
		// 组织内置管理员，不能被删除
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
		// 组织内置管理员，不能修改状态
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

// FindOrgRoleByRoleID 通过 roleID 查找 OrgRole（roleId 全局唯一）
func (c *Client) FindOrgRoleByRoleID(ctx context.Context, roleID uint32) (*model.OrgRole, error) {
	orgRole := &model.OrgRole{}
	if err := sqlopt.WithRoleID(roleID).Apply(c.db.WithContext(ctx)).First(orgRole).Error; err != nil {
		return nil, fmt.Errorf("find org role by roleID %v err: %v", roleID, err)
	}
	return orgRole, nil
}

// getRoleUserCountsTx 批量统计多个角色各自的关联用户数（按 user_id 去重）。
// 返回 roleID -> 用户数；未关联用户的 roleID 不在 map 中（调用方按 0 兜底）。
// 语义与 GetRoleUsers 一致：同一用户在多个组织的同一角色记录只计一次。
// orgID 为组织范围：0 或顶级组织(TopOrgID)时不限组织（全系统统计）；
// 否则只统计 user_roles.org_id == orgID 的记录（本组织关联数）。
func getRoleUserCountsTx(tx *gorm.DB, roleIDs []uint32, orgID uint32) (map[uint32]int32, error) {
	ret := make(map[uint32]int32)
	if len(roleIDs) == 0 {
		return ret, nil
	}
	type roleCount struct {
		RoleID uint32 `gorm:"column:role_id"`
		Count  int64  `gorm:"column:count"`
	}
	var rows []roleCount
	query := tx.Model(&model.UserRole{}).
		Select("role_id, COUNT(DISTINCT user_id) AS count").
		Where("role_id IN ?", roleIDs)
	if orgID != 0 && orgID != config.TopOrgID() {
		query = query.Where("org_id = ?", orgID)
	}
	if err := query.Group("role_id").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get role user counts err: %v", err)
	}
	for _, r := range rows {
		ret[r.RoleID] = int32(r.Count)
	}
	return ret, nil
}

func getRoleInfoTx(tx *gorm.DB, orgRole *model.OrgRole, orgTree *model.OrgNode, userCount int32) (*RoleInfo, error) {
	roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(orgRole.RoleID))
	if err != nil {
		return nil, fmt.Errorf("get role %v err: %v", orgRole.RoleID, err)
	}
	return toRoleInfoTx(tx, orgRole, roleDetail, orgTree, userCount)
}

func toRoleInfoTx(tx *gorm.DB, orgRole *model.OrgRole, roleDetail *perm.RolePerms, orgTree *model.OrgNode, userCount int32) (*RoleInfo, error) {
	ret := &RoleInfo{
		ID:         uint32(roleDetail.Role),
		IsAdmin:    orgRole.IsAdmin,
		IsSystem:   orgRole.OrgID == config.TopOrgID(),
		IsGlobal:   false,
		Name:       roleDetail.Name,
		Remark:     roleDetail.Desc,
		Status:     roleDetail.Enable,
		CreatedAt:  roleDetail.CreatedAt,
		AvatarPath: orgRole.AvatarPath,
		OrgName:    orgTree.GetFullName(orgRole.OrgID),
		UserCount:  userCount,
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

// GetRoleUsers 获取角色关联用户列表（分页）
// 合并同一用户的多条 UserRole 记录为一条
// orgID 为组织范围：0 或顶级组织(TopOrgID)时不限组织（全系统，Orgs 列出该用户持有此角色的所有组织）；
// 否则只查 user_roles.org_id == orgID 的记录（本组织，Orgs 只含当前组织）。
func (c *Client) GetRoleUsers(ctx context.Context, roleID, orgID uint32, name string, offset, limit int32) ([]RoleUser, int64, *errs.Status) {
	var ret []RoleUser
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var userRoles []*model.UserRole
		userRoleQuery := sqlopt.WithRoleID(roleID).Apply(tx)
		if orgID != 0 && orgID != config.TopOrgID() {
			userRoleQuery = sqlopt.WithOrgID(orgID).Apply(userRoleQuery)
		}
		if err := userRoleQuery.Find(&userRoles).Error; err != nil {
			return toErrStatus("iam_role_users_get", util.Int2Str(roleID), err.Error())
		}
		if len(userRoles) == 0 {
			return nil
		}
		// collect unique userIDs and build userRole mapping
		userIDSet := make(map[uint32]bool)
		// userOrgRoles: userID -> list of orgIDs where user has this role
		userOrgRoles := make(map[uint32][]uint32)
		for _, ur := range userRoles {
			userIDSet[ur.UserID] = true
			userOrgRoles[ur.UserID] = append(userOrgRoles[ur.UserID], ur.OrgID)
		}
		var userIDs []uint32
		for uid := range userIDSet {
			userIDs = append(userIDs, uid)
		}
		// query users with name filter
		var users []*model.User
		query := sqlopt.WithIDs(userIDs).Apply(tx)
		if name != "" {
			query = sqlopt.LikeName(name).Apply(query)
		}
		// count total unique users (after name filter)
		if err := query.Model(&model.User{}).Count(&count).Error; err != nil {
			return toErrStatus("iam_role_users_get", util.Int2Str(roleID), err.Error())
		}
		// paginate
		if offset >= 0 && limit > 0 {
			query = query.Offset(int(offset)).Limit(int(limit))
		}
		if err := query.Find(&users).Error; err != nil {
			return toErrStatus("iam_role_users_get", util.Int2Str(roleID), err.Error())
		}
		userMap := make(map[uint32]*model.User)
		for _, u := range users {
			userMap[u.ID] = u
		}
		// get org tree for org names
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_role_users_get", util.Int2Str(roleID), err.Error())
		}
		// build result
		for _, u := range users {
			orgIDs := userOrgRoles[u.ID]
			ru := RoleUser{
				UserID:     u.ID,
				UserName:   u.Name,
				Phone:      u.Phone,
				Email:      u.Email,
				AvatarPath: u.AvatarPath,
			}
			for _, oid := range orgIDs {
				ru.Orgs = append(ru.Orgs, IDName{
					ID:   oid,
					Name: orgTree.GetFullName(oid),
				})
			}
			ret = append(ret, ru)
		}
		return nil
	})
}

// RemoveRoleUser 移除角色关联用户
// orgID 为组织范围：0 或顶级组织(TopOrgID)时不限组织（按 role_id+user_id 删除该用户的全局唯一关联，
// 用于系统组织/系统调用跨组织移除）；否则只删除 user_roles.org_id == orgID 的本组织关联。
// 注意 user_roles 主键为 (user_id, role_id)，org_id 字段记录的是赋权时的组织，
// 因此系统组织下若带 org_id=顶级 过滤，会删不掉在业务组织赋权的记录。
func (c *Client) RemoveRoleUser(ctx context.Context, roleID, userID, orgID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		opts := []sqlopt.SQLOption{
			sqlopt.WithRoleID(roleID),
			sqlopt.WithUserID(userID),
		}
		if orgID != 0 && orgID != config.TopOrgID() {
			opts = append(opts, sqlopt.WithOrgID(orgID))
		}
		// check system admin: 不能移除系统内置管理员(admin)的超级管理员角色
	// （与 UpdateUser 中的保护一致）
	if userID == config.AdminUserID() && roleID == config.AdminRoleID() {
		return toErrStatus("iam_role_user_remove", util.Int2Str(roleID), util.Int2Str(userID), "cannot remove admin user's admin role")
	}
	// check org creator: 不能移除组织创建者在所属组织的管理员角色，否则创建者会失去自己组织的管理权限
		// （与 UpdateUser 中的保护一致；orgID 为 0/TopOrgID 的系统组织调用不在此限）
		if orgID != 0 && orgID != config.TopOrgID() {
			org := &model.Org{}
			if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
				return toErrStatus("iam_role_user_remove", util.Int2Str(roleID), util.Int2Str(userID), err.Error())
			}
			if org.CreatorID == userID {
				// 仅当被移除的角色是该组织的管理员角色时才拒绝
				orgRole := &model.OrgRole{}
				if err := sqlopt.SQLOptions(
					sqlopt.WithOrgID(orgID),
					sqlopt.WithRoleID(roleID),
				).Apply(tx).First(orgRole).Error; err == nil && orgRole.IsAdmin {
					return toErrStatus("iam_role_user_remove", util.Int2Str(roleID), util.Int2Str(userID), "cannot remove org creator admin role")
				}
			}
		}
		if err := sqlopt.SQLOptions(opts...).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_role_user_remove", util.Int2Str(roleID), util.Int2Str(userID), err.Error())
		}
		return nil
	})
}
