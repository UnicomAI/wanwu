package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access"
	"github.com/gromitlee/access/pkg/perm"
	"gorm.io/gorm"
)

func (c *Client) GetAdminUser(ctx context.Context) (uint32, error) {
	user := &model.User{}
	err := sqlopt.WithAdmin(true).Apply(c.db.WithContext(ctx)).First(user).Error
	return user.ID, err
}

func (c *Client) GetUser(ctx context.Context, userID, orgID uint32) (*UserInfo, *errs.Status) {
	var ret *UserInfo
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}
		err := sqlopt.WithID(userID).Apply(tx).First(user).Error
		if err != nil {
			return toErrStatus("iam_user_get", util.Int2Str(userID), err.Error())
		}
		if orgID == 0 {
			ret = toUserInfo(user)
			return nil
		}
		// check org user 用户在组织中的状态校验
		orgUser := &model.OrgUser{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithUserID(user.ID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).First(orgUser).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_user_org_check", util.Int2Str(user.ID),
					util.Int2Str(orgID), err.Error())
			}
		} else if orgUser.Status == sqlopt.OrgUserStatusDisabled {
			return toErrStatus("iam_user_org_check", util.Int2Str(user.ID),
				util.Int2Str(orgID), "user disabled")
		}
		// org tree
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_user_get", util.Int2Str(userID), err.Error())
		}
		ret, err = toUserInfoTx(tx, user, orgTree, orgID == config.TopOrgID(), orgID)
		if err != nil {
			return toErrStatus("iam_user_get", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

func (c *Client) GetUsers(ctx context.Context, orgID uint32, name, email string, roleIDs []uint32, offset, limit int32) ([]*UserInfo, int64, *errs.Status) {
	var ret []*UserInfo
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// 系统组织(顶级组织)下返回全系统用户：org_users / user_roles 查询不限组织；
		// 非系统组织只返回本组织用户。与 toUserInfoTx 的 allOrg 判断保持一致。
		allOrg := orgID == config.TopOrgID()
		// users
		var users []*model.User
		orgUsersQuery := tx.Select("user_id").Table("org_users")
		if !allOrg {
			orgUsersQuery = sqlopt.WithOrgID(orgID).Apply(orgUsersQuery)
		}
		query := tx.Where("id IN (?)", orgUsersQuery)
		// name and email filter: when name == email, use OR (name LIKE ? OR email LIKE ?)
		if name != "" && email != "" && name == email {
			query = query.Where("name LIKE ? OR email LIKE ?", "%"+name+"%", "%"+email+"%")
		} else {
			query = sqlopt.LikeName(name).Apply(query)
			query = sqlopt.LikeEmail(email).Apply(query)
		}
		if len(roleIDs) > 0 {
			var roleUsersQuery *gorm.DB
			if allOrg {
				roleUsersQuery = sqlopt.WithRoles(roleIDs).Apply(tx).Select("user_id").Table("user_roles")
			} else {
				roleUsersQuery = sqlopt.SQLOptions(
					sqlopt.WithOrgID(orgID),
					sqlopt.WithRoles(roleIDs),
				).Apply(tx).Select("user_id").Table("user_roles")
			}
			query = query.Where("id IN (?)", roleUsersQuery)
		}
		if err := query.
			Offset(int(offset)).Limit(int(limit)).Order("id DESC").Find(&users).
			Offset(-1).Limit(-1).Count(&count).Error; err != nil {
			return toErrStatus("iam_users_get", util.Int2Str(orgID), err.Error())
		}
		// org tree
		orgTree, err := getOrgTree(tx)
		if err != nil {
			return toErrStatus("iam_users_get", util.Int2Str(orgID), err.Error())
		}
		for _, user := range users {
			info, err := toUserInfoTx(tx, user, orgTree, orgID == config.TopOrgID(), orgID)
			if err != nil {
				return toErrStatus("iam_users_get", util.Int2Str(orgID), err.Error())
			}
			ret = append(ret, info)
		}
		return nil
	})

}

func (c *Client) SelectUsersNotInOrg(ctx context.Context, orgID uint32, name string) ([]IDName, *errs.Status) {
	var ret []IDName
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var users []*model.User
		orgUsersQuery := sqlopt.WithOrgID(orgID).Apply(tx).Select("user_id").Table("org_users")
		notOrgUsersQuery := sqlopt.WithOrgID(config.TopOrgID()).Apply(tx).Select("user_id").Where("user_id NOT IN (?)", orgUsersQuery).Table("org_users")
		if err := sqlopt.LikeName(name).Apply(tx).Where("id IN (?)", notOrgUsersQuery).Find(&users).Error; err != nil {
			return toErrStatus("iam_users_select_not_in_org", util.Int2Str(orgID), err.Error())
		}
		for _, user := range users {
			ret = append(ret, IDName{ID: user.ID, Name: user.Name})
		}
		return nil
	})

}

func (c *Client) SelectUsersByUserIDs(ctx context.Context, userIDs []uint32) ([]IDName, *errs.Status) {
	var users []*model.User
	if err := sqlopt.WithIDs(userIDs).Apply(c.db.WithContext(ctx)).Find(&users).Error; err != nil {
		return nil, toErrStatus("iam_user_select_by_ids", err.Error())
	}
	var ret []IDName
	for _, user := range users {
		ret = append(ret, IDName{ID: user.ID, Name: user.Name})
	}
	return ret, nil
}

// IsUserOrgAdmin 查询用户在系统任一组织中是否拥有组织管理员角色。
// user_roles.is_admin 在分配组织内置管理员角色时置 true，故只要存在任一 is_admin=true 记录即为 true。
// 同时排除用户在该组织被禁用（org_users.status=disable）的记录，与 checkUserIsAdmin 语义保持一致。
func (c *Client) IsUserOrgAdmin(ctx context.Context, userID uint32) (bool, *errs.Status) {
	var count int64
	if err := c.db.WithContext(ctx).Model(&model.UserRole{}).
		Joins("JOIN org_users ON org_users.user_id = user_roles.user_id AND org_users.org_id = user_roles.org_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.is_admin = ?", true).
		Where("org_users.status IS NULL OR org_users.status != ?", sqlopt.OrgUserStatusDisabled).
		Count(&count).Error; err != nil {
		return false, toErrStatus("iam_user_is_org_admin", util.Int2Str(userID), err.Error())
	}
	return count > 0, nil
}

// IsAdminInOrgs 查询用户对指定组织列表是否都拥有管理员权限（含祖先组织继承）。
// 所有组织都通过才返回 true。
func (c *Client) IsAdminInOrgs(ctx context.Context, userID uint32, orgIDs []uint32) (bool, *errs.Status) {
	if len(orgIDs) == 0 {
		return false, toErrStatus("iam_user_is_admin_in_orgs", "orgIDs empty")
	}
	var isAdmin bool
	return isAdmin, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var orgs []*model.Org
		if err := sqlopt.WithIDs(orgIDs).Apply(tx).Find(&orgs).Error; err != nil {
			return toErrStatus("iam_user_is_admin_of_org", util.Int2Str(userID), err.Error())
		}
		for _, org := range orgs {
			admin, _, err := checkUserIsAdmin(tx, userID, org)
			if err != nil {
				return toErrStatus("iam_user_is_admin_of_org", util.Int2Str(userID), util.Int2Str(org.ID), err.Error())
			}
			if !admin {
				return nil
			}
		}
		isAdmin = true
		return nil
	})
}

func (c *Client) CreateUser(ctx context.Context, user *model.User, orgID uint32, roleIDs []uint32) (uint32, *errs.Status) {
	if user.ID != 0 {
		return 0, toErrStatus("iam_user_create", "create user but id not 0")
	}
	return user.ID, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		return createUserTx(tx, user, orgID, roleIDs)
	})
}

func (c *Client) CreateUsers(ctx context.Context, users []*UsersInfo, creatorID, orgID uint32) (*CreateUsersResult, *errs.Status) {
	return createUsersTx(c.db.WithContext(ctx), users, creatorID, orgID)
}

func createUsersTx(tx *gorm.DB, users []*UsersInfo, creatorID, orgID uint32) (*CreateUsersResult, *errs.Status) {
	result := &CreateUsersResult{
		Total: len(users),
	}

	// 前置校验（组织、创建者）
	if err := sqlopt.WithID(orgID).Apply(tx).First(&model.Org{}).Error; err != nil {
		return nil, toErrStatus("iam_user_batch_create", err.Error())
	}
	if creatorID != 0 {
		if err := sqlopt.WithID(creatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return nil, toErrStatus("iam_user_batch_create", err.Error())
		}
	}

	// 获取角色映射（组织角色 + 全局角色）
	var orgRoles []*model.OrgRole
	if err := sqlopt.WithOrgID(orgID).Apply(tx).Find(&orgRoles).Error; err != nil {
		return nil, toErrStatus("iam_roles_get", util.Int2Str(orgID), err.Error())
	}
	roleNameMap := make(map[string]*model.OrgRole)
	for i := range orgRoles {
		roleNameMap[orgRoles[i].Name] = orgRoles[i]
	}
	var globalRoles []*model.GlobalRole
	if err := tx.Find(&globalRoles).Error; err != nil {
		return nil, toErrStatus("iam_global_roles_get", err.Error())
	}
	globalRoleNameMap := make(map[string]*model.GlobalRole)
	for i := range globalRoles {
		globalRoleNameMap[globalRoles[i].Name] = globalRoles[i]
	}

	// 预先查询已存在的用户名和电话号码
	existingNames := make(map[string]bool)
	existingPhones := make(map[string]bool)
	for _, userInfo := range users {
		var user model.User
		if sqlopt.WithName(userInfo.UserName).Apply(tx).First(&user).Error == nil {
			existingNames[userInfo.UserName] = true
		}
		if userInfo.Phone != "" {
			if sqlopt.WithPhone(userInfo.Phone).Apply(tx).First(&user).Error == nil {
				existingPhones[userInfo.Phone] = true
			}
		}
	}

	// 遍历处理每个用户
	for i, userInfo := range users {
		// 检查用户名和电话号码是否都已存在（跳过）
		if existingNames[userInfo.UserName] && existingPhones[userInfo.Phone] {
			result.Errors = append(result.Errors, CreateUserError{
				Index:  i,
				Reason: "用户名和电话号码都已存在",
			})
			continue
		}

		// 检查用户名是否已存在
		if existingNames[userInfo.UserName] {
			result.Errors = append(result.Errors, CreateUserError{
				Index:  i,
				Reason: "用户名已存在",
			})
			continue
		}

		// 检查电话号码是否已存在
		if userInfo.Phone != "" && existingPhones[userInfo.Phone] {
			result.Errors = append(result.Errors, CreateUserError{
				Index:  i,
				Reason: fmt.Sprintf("电话号码%s已存在", userInfo.Phone),
			})
			continue
		}

		// 获取角色ID
		var roleID uint32
		var isAdmin bool
		if userInfo.RoleName != "" {
			orgRole, ok := roleNameMap[userInfo.RoleName]
			if ok {
				roleID = orgRole.RoleID
				isAdmin = orgRole.IsAdmin
			} else {
				// not found in org roles, check global roles
				globalRole, ok := globalRoleNameMap[userInfo.RoleName]
				if !ok {
					result.Errors = append(result.Errors, CreateUserError{
						Index:  i,
						Reason: fmt.Sprintf("角色'%s'不存在", userInfo.RoleName),
					})
					continue
				}
				if !globalRole.Status {
					result.Errors = append(result.Errors, CreateUserError{
						Index:  i,
						Reason: fmt.Sprintf("角色'%s'已禁用", userInfo.RoleName),
					})
					continue
				}
				roleID = globalRole.RoleID
			}
		}

		// 创建用户（使用保存点，失败不影响其他用户）
		savePoint := fmt.Sprintf("user_create_%d", i)
		if err := tx.SavePoint(savePoint).Error; err != nil {
			return nil, toErrStatus("iam_user_batch_create", err.Error())
		}

		user := &model.User{
			CreatorID: creatorID,
			Name:      userInfo.UserName,
			Phone:     userInfo.Phone,
			Email:     userInfo.Email,
			Company:   userInfo.Company,
			Remark:    userInfo.Remark,
			Password:  util.SHA256(userInfo.Password),
			Status:    true,
		}
		if err := tx.Create(user).Error; err != nil {
			tx.RollbackTo(savePoint)
			result.Errors = append(result.Errors, CreateUserError{
				Index:  i,
				Reason: "创建用户失败: " + err.Error(),
			})
			continue
		}

		// 创建组织用户关联
		orgUsers := []*model.OrgUser{{OrgID: orgID, UserID: user.ID}}
		if orgID != config.TopOrgID() {
			orgUsers = append(orgUsers, &model.OrgUser{OrgID: config.TopOrgID(), UserID: user.ID})
		}
		if err := tx.Create(orgUsers).Error; err != nil {
			tx.RollbackTo(savePoint)
			result.Errors = append(result.Errors, CreateUserError{
				Index:  i,
				Reason: "创建组织关联失败",
			})
			continue
		}

		// 创建用户角色关联
		if roleID != 0 {
			userRole := &model.UserRole{
				OrgID:   orgID,
				UserID:  user.ID,
				RoleID:  roleID,
				IsAdmin: isAdmin,
			}
			if err := tx.Create(userRole).Error; err != nil {
				tx.RollbackTo(savePoint)
				result.Errors = append(result.Errors, CreateUserError{
					Index:  i,
					Reason: "创建角色关联失败",
				})
				continue
			}
		}

		// 成功：更新已存在记录，防止后续重复
		existingNames[userInfo.UserName] = true
		if userInfo.Phone != "" {
			existingPhones[userInfo.Phone] = true
		}
		result.Success++
	}

	result.Failed = len(result.Errors)
	return result, nil
}

func createUserTx(tx *gorm.DB, user *model.User, orgID uint32, roleIDs []uint32) *errs.Status {
	// check org
	var adminRoles, nonAdminRoles, globalRoles []uint32
	if err := sqlopt.WithID(orgID).Apply(tx).First(&model.Org{}).Error; err != nil {
		return toErrStatus("iam_user_create", err.Error())
	} else if len(roleIDs) > 0 {
		// check roles: org roles first, then global roles
		for _, roleID := range roleIDs {
			orgRole := &model.OrgRole{}
			if err := sqlopt.SQLOptions(
				sqlopt.WithOrgID(orgID),
				sqlopt.WithRoleID(roleID),
			).Apply(tx).First(orgRole).Error; err == nil {
				// org role found
				if orgRole.IsAdmin {
					adminRoles = append(adminRoles, roleID)
				} else {
					nonAdminRoles = append(nonAdminRoles, roleID)
				}
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("iam_user_create", err.Error())
			} else {
				// not found in org_roles, check global_roles
				globalRole := &model.GlobalRole{}
				if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
					return toErrStatus("iam_user_create", util.Int2Str(orgID), util.Int2Str(roleID), err.Error())
				}
				if !globalRole.Status {
					return toErrStatus("iam_user_create", util.Int2Str(orgID), util.Int2Str(roleID), "global role disabled")
				}
				globalRoles = append(globalRoles, roleID)
			}
		}
	}
	// check creator
	if user.CreatorID != 0 {
		// 正常创建用户
		if user.IsAdmin {
			return toErrStatus("iam_user_create", "cannot create admin user")
		}
		if err := sqlopt.WithID(user.CreatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return toErrStatus("iam_user_create", err.Error())
		}
	} else {
		// 创建系统内管理员，此时系统内不能存在任何用户
		if !user.IsAdmin {
			return toErrStatus("iam_user_create", "create admin user but is not admin")
		}
		if err := tx.First(&model.User{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				err = errors.New("already exist")
			}
			return toErrStatus("iam_user_create", err.Error())
		}
	}
	// check name
	if err := sqlopt.WithName(user.Name).Apply(tx).First(&model.User{}).Error; err != gorm.ErrRecordNotFound {
		if err == nil {
			return toErrStatus("iam_user_create_name")
		}
		return toErrStatus("iam_user_create_name_err", user.Name, err.Error())
	}
	// check phone
	if user.Phone != "" {
		if err := sqlopt.WithPhone(user.Phone).Apply(tx).First(&model.User{}).Error; err != gorm.ErrRecordNotFound {
			if err == nil {
				return toErrStatus("iam_user_create_phone")
			}
			return toErrStatus("iam_user_create_phone_err", user.Phone, err.Error())
		}
	}
	// password encode
	user.Password = util.SHA256(user.Password)
	// create user
	if err := tx.Create(user).Error; err != nil {
		return toErrStatus("iam_user_create", err.Error())
	}
	// create org user
	var orgUsers []*model.OrgUser
	orgUsers = append(orgUsers, &model.OrgUser{
		OrgID:  orgID,
		UserID: user.ID,
	})
	if orgID != config.TopOrgID() { // 默认也加入顶级组织
		orgUsers = append(orgUsers, &model.OrgUser{
			OrgID:  config.TopOrgID(),
			UserID: user.ID,
		})
	}
	if err := tx.Create(orgUsers).Error; err != nil {
		return toErrStatus("iam_user_create", err.Error())
	}
	// create user role
	if len(roleIDs) > 0 {
		var userRoles []*model.UserRole
		for _, roleID := range adminRoles {
			userRoles = append(userRoles, &model.UserRole{
				OrgID:   orgID,
				UserID:  user.ID,
				RoleID:  roleID,
				IsAdmin: true,
			})
		}
		for _, roleID := range nonAdminRoles {
			userRoles = append(userRoles, &model.UserRole{
				OrgID:  orgID,
				UserID: user.ID,
				RoleID: roleID,
			})
		}
		for _, roleID := range globalRoles {
			userRoles = append(userRoles, &model.UserRole{
				OrgID:  orgID,
				UserID: user.ID,
				RoleID: roleID,
			})
		}
		if err := tx.Create(userRoles).Error; err != nil {
			return toErrStatus("iam_user_create", err.Error())
		}
	}
	return nil
}

func (c *Client) UpdateUser(ctx context.Context, user *model.User, orgID uint32, roleIDs []uint32) *errs.Status {
	if user.ID == 0 {
		return toErrStatus("iam_user_update", "update user but id 0")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check org user 用户在组织中的状态校验
		orgUser := &model.OrgUser{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithUserID(user.ID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).First(orgUser).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_user_update", err.Error())
			}
		} else if orgUser.Status == sqlopt.OrgUserStatusDisabled {
			return toErrStatus("iam_user_update", fmt.Sprintf("org %v user %v disable", orgID, user.ID))
		}
		if user.ID == config.AdminUserID() && orgID == config.TopOrgID() {
			var exist bool
			for _, roleID := range roleIDs {
				if roleID == config.AdminRoleID() {
					exist = true
					break
				}
			}
			if !exist {
				return toErrStatus("iam_user_update", "cannot remove admin role")
			}
		}
		// check org creator: 不能移除组织创建者在该组织的管理员角色
		if orgID != config.TopOrgID() {
			org := &model.Org{}
			if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
				return toErrStatus("iam_user_update", err.Error())
			}
			if org.CreatorID == user.ID {
				// 查找当前用户在该组织的管理员角色
				var currentUserRoles []*model.UserRole
				if err := sqlopt.SQLOptions(
					sqlopt.WithOrgID(orgID),
					sqlopt.WithUserID(user.ID),
					sqlopt.WithAdmin(true),
				).Apply(tx).Find(&currentUserRoles).Error; err != nil {
					return toErrStatus("iam_user_update", err.Error())
				}
				for _, ur := range currentUserRoles {
					var found bool
					for _, roleID := range roleIDs {
						if roleID == ur.RoleID {
							found = true
							break
						}
					}
					if !found {
						return toErrStatus("iam_user_update", util.Int2Str(orgID), util.Int2Str(user.ID), "cannot remove org creator admin role")
					}
				}
			}
		}
		var adminRoles, nonAdminRoles, globalRoles []uint32
		if len(roleIDs) > 0 {
			for _, roleID := range roleIDs {
				orgRole := &model.OrgRole{}
				if err := sqlopt.SQLOptions(
					sqlopt.WithOrgID(orgID),
					sqlopt.WithRoleID(roleID),
				).Apply(tx).First(orgRole).Error; err == nil {
					// org role found
					if orgRole.IsAdmin {
						adminRoles = append(adminRoles, roleID)
					} else {
						nonAdminRoles = append(nonAdminRoles, roleID)
					}
				} else if !errors.Is(err, gorm.ErrRecordNotFound) {
					return toErrStatus("iam_user_update", err.Error())
				} else {
					// not found in org_roles, check global_roles
					globalRole := &model.GlobalRole{}
					if err := sqlopt.WithRoleID(roleID).Apply(tx).First(globalRole).Error; err != nil {
						return toErrStatus("iam_user_update", util.Int2Str(user.ID), util.Int2Str(roleID), err.Error())
					}
					if !globalRole.Status {
						return toErrStatus("iam_user_update", util.Int2Str(user.ID), util.Int2Str(roleID), "global role disabled")
					}
					globalRoles = append(globalRoles, roleID)
				}
			}
		}
		var users []*model.User
		// check phone
		if user.Phone != "" {
			if err := sqlopt.WithPhone(user.Phone).Apply(tx).Find(&users).Error; err != nil {
				return toErrStatus("iam_user_update_phone", util.Int2Str(user.ID), user.Phone, err.Error())
			}
			if len(users) > 0 {
				for _, u := range users {
					if u.ID != user.ID {
						return toErrStatus("iam_user_update_phone", util.Int2Str(user.ID), user.Phone, "already exist")
					}
				}
			}
		}
		// update user
		if err := tx.Model(user).Updates(map[string]interface{}{
			"nick":    user.Nick,
			"gender":  user.Gender,
			"phone":   user.Phone,
			"company": user.Company,
			"remark":  user.Remark,
		}).Error; err != nil {
			return toErrStatus("iam_user_update_err", util.Int2Str(user.ID), err.Error())
		}
		// check need recreate user role
		var userRoles []*model.UserRole
		if err := sqlopt.SQLOptions(
			sqlopt.WithOrgID(orgID),
			sqlopt.WithUserID(user.ID),
		).Apply(tx).Find(&userRoles).Error; err != nil {
			return toErrStatus("iam_user_update_err", util.Int2Str(user.ID), err.Error())
		}
		var recreate bool
		if len(userRoles) != len(roleIDs) {
			recreate = true
		} else {
			for _, roleID := range roleIDs {
				var exist bool
				for _, userRole := range userRoles {
					if userRole.RoleID == roleID {
						exist = true
						break
					}
				}
				if !exist {
					recreate = true
					break
				}
			}
		}
		if recreate {
			// delete user role
			if err := sqlopt.SQLOptions(
				sqlopt.WithOrgID(orgID),
				sqlopt.WithUserID(user.ID),
			).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
				return toErrStatus("iam_user_update_err", util.Int2Str(user.ID), err.Error())
			}
			// create user role
			if len(roleIDs) > 0 {
				var userRoles []*model.UserRole
				for _, roleID := range adminRoles {
					userRoles = append(userRoles, &model.UserRole{
						OrgID:   orgID,
						UserID:  user.ID,
						RoleID:  roleID,
						IsAdmin: true,
					})
				}
				for _, roleID := range nonAdminRoles {
					userRoles = append(userRoles, &model.UserRole{
						OrgID:  orgID,
						UserID: user.ID,
						RoleID: roleID,
					})
				}
				for _, roleID := range globalRoles {
					userRoles = append(userRoles, &model.UserRole{
						OrgID:  orgID,
						UserID: user.ID,
						RoleID: roleID,
					})
				}
				if err := tx.Create(userRoles).Error; err != nil {
					return toErrStatus("iam_user_update_err", util.Int2Str(user.ID), err.Error())
				}
			}
		}
		return nil
	})

}

func (c *Client) DeleteUser(ctx context.Context, userID uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check user
		if userID == config.AdminUserID() {
			return toErrStatus("iam_user_delete", util.Int2Str(userID), "cannot delete admin user")
		}
		// delete user role
		if err := sqlopt.WithUserID(userID).Apply(tx).Delete(&model.UserRole{}).Error; err != nil {
			return toErrStatus("iam_user_delete", util.Int2Str(userID), err.Error())
		}
		// delete org user
		if err := sqlopt.WithUserID(userID).Apply(tx).Delete(&model.OrgUser{}).Error; err != nil {
			return toErrStatus("iam_user_delete", util.Int2Str(userID), err.Error())
		}
		// delete user
		if err := sqlopt.WithID(userID).Apply(tx).Delete(&model.User{}).Error; err != nil {
			return toErrStatus("iam_user_delete", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

func (c *Client) UpdateUserAvatar(ctx context.Context, userID uint32, avatarPath string) *errs.Status {
	if err := sqlopt.WithID(userID).Apply(c.db.WithContext(ctx)).Model(&model.User{}).Updates(map[string]interface{}{
		"avatar_path": avatarPath,
	}).Error; err != nil {
		return toErrStatus("iam_user_avatar_upload", util.Int2Str(userID), err.Error())
	}
	return nil
}

func (c *Client) ChangeUserStatus(ctx context.Context, userID, orgID uint32, status bool) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check user
		if userID == config.AdminUserID() {
			return toErrStatus("iam_user_change_status", util.Int2Str(userID), "cannot change admin user status")
		}
		// check org creator: 不能禁用组织创建者在该组织的状态
		if !status && orgID != config.TopOrgID() {
			org := &model.Org{}
			if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err == nil {
				if org.CreatorID == userID {
					return toErrStatus("iam_user_change_status", util.Int2Str(userID), "cannot disable org creator")
				}
			}
		}
		// change status
		if orgID == config.TopOrgID() {
			if err := sqlopt.WithID(userID).Apply(tx).Model(&model.User{}).Updates(map[string]interface{}{
				"status": status,
			}).Error; err != nil {
				return toErrStatus("iam_user_change_status", util.Int2Str(userID), err.Error())
			}
			return nil
		}
		var orgUserStatus string
		if !status {
			orgUserStatus = sqlopt.OrgUserStatusDisabled
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithUserID(userID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).Model(&model.OrgUser{}).Updates(map[string]interface{}{
			"status": orgUserStatus,
		}).Error; err != nil {
			return toErrStatus("iam_user_change_status", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

func (c *Client) UpdateUserPassword(ctx context.Context, userID uint32, pwd, newPwd string) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), err.Error())
		}
		// check old password
		if user.Password != util.SHA256(pwd) {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), "password error")
		}
		// check new password same as old password
		if user.Password == util.SHA256(newPwd) {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), "new password cannot be same as old password")
		}
		if err := tx.Model(user).Updates(map[string]interface{}{
			"password":                util.SHA256(newPwd),
			"last_update_password_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

func (c *Client) ResetUserPassword(ctx context.Context, userID uint32, pwd string) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_user_password_reset", util.Int2Str(userID), err.Error())
		}
		// reset password
		if err := tx.Model(user).Updates(map[string]interface{}{
			"password": util.SHA256(pwd),
		}).Error; err != nil {
			return toErrStatus("iam_user_password_reset", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

func (c *Client) GetUserPermission(ctx context.Context, userID, orgID uint32) (*Permission, *errs.Status) {
	var ret *Permission
	var err error
	return ret, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		ret, err = getUserPermission(tx, userID, orgID)
		if err != nil {
			return toErrStatus("iam_user_permission_get", util.Int2Str(userID),
				util.Int2Str(orgID), err.Error())
		}
		return nil
	})

}

func (c *Client) ChangeUserLanguage(ctx context.Context, userID uint32, language string) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_user_language_change", util.Int2Str(userID), err.Error())
		}
		if err := tx.Model(user).Updates(map[string]interface{}{
			"language": language,
		}).Error; err != nil {
			return toErrStatus("iam_user_language_change", util.Int2Str(userID), err.Error())
		}
		return nil
	})

}

// --- internal function ---

func toUserInfo(user *model.User) *UserInfo {
	return &UserInfo{
		ID:         user.ID,
		Status:     user.Status,
		Name:       user.Name,
		Nick:       user.Nick,
		Gender:     user.Gender,
		Phone:      user.Phone,
		Email:      user.Email,
		Company:    user.Company,
		Remark:     user.Remark,
		CreatedAt:  user.CreatedAt,
		Language:   user.Language,
		AvatarPath: user.AvatarPath,
	}
}

func toUserInfoTx(tx *gorm.DB, user *model.User, orgTree *model.OrgNode, allOrg bool, orgID ...uint32) (*UserInfo, error) {
	ret := &UserInfo{
		ID:         user.ID,
		Name:       user.Name,
		Nick:       user.Nick,
		Gender:     user.Gender,
		Phone:      user.Phone,
		Email:      user.Email,
		Company:    user.Company,
		Remark:     user.Remark,
		CreatedAt:  user.CreatedAt,
		Language:   user.Language,
		AvatarPath: user.AvatarPath,
	}
	// user org
	orgs, err := getUserOrgsTx(tx, user.ID, orgTree)
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		if allOrg {
			ret.Orgs = append(ret.Orgs, &UserOrg{Org: org})
			ret.Status = user.Status
		} else if util.Exist(orgID, org.ID) {
			ret.Orgs = append(ret.Orgs, &UserOrg{Org: org})
			ret.Status = org.Status
		}
	}
	// creator
	if user.CreatorID != 0 {
		creator, err := getCreatorTx(tx, user.CreatorID)
		if err != nil {
			return nil, err
		}
		ret.Creator = creator
	}
	// user role
	var userRoles []*model.UserRole
	var userRolesOpt sqlopt.SQLOption
	if allOrg {
		userRolesOpt = sqlopt.WithUserID(user.ID)
	} else {
		userRolesOpt = sqlopt.SQLOptions(
			sqlopt.WithUserID(user.ID),
			sqlopt.WithOrgs(orgID),
		)
	}
	if err := userRolesOpt.Apply(tx).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("get user %v roles err: %v", user.ID, err)
	}
	var roleIDs []uint32
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}
	// org role
	var orgRoles []*model.OrgRole
	var orgRolesOpt sqlopt.SQLOption
	if allOrg {
		orgRolesOpt = sqlopt.WithRoles(roleIDs)
	} else {
		orgRolesOpt = sqlopt.SQLOptions(
			sqlopt.WithRoles(roleIDs),
			sqlopt.WithOrgs(orgID),
		)
	}
	if err := orgRolesOpt.Apply(tx).Find(&orgRoles).Error; err != nil {
		return nil, fmt.Errorf("get org roles %v err: %v", roleIDs, err)
	}
	// global role: user_roles may reference global roles not present in org_roles.
	// Resolve any roleIDs not covered by org_roles against global_roles so users
	// with global roles don't break the read path.
	orgRoleExist := make(map[uint32]bool)
	for _, orgRole := range orgRoles {
		orgRoleExist[orgRole.RoleID] = true
	}
	var missingRoleIDs []uint32
	for _, roleID := range roleIDs {
		if !orgRoleExist[roleID] {
			missingRoleIDs = append(missingRoleIDs, roleID)
		}
	}
	globalRoleMap := make(map[uint32]*model.GlobalRole)
	if len(missingRoleIDs) > 0 {
		var globalRoles []*model.GlobalRole
		if err := sqlopt.WithRoles(missingRoleIDs).Apply(tx).Find(&globalRoles).Error; err != nil {
			return nil, fmt.Errorf("get global roles %v err: %v", missingRoleIDs, err)
		}
		for i := range globalRoles {
			globalRoleMap[globalRoles[i].RoleID] = globalRoles[i]
		}
	}
	for _, userRole := range userRoles {
		var orgExist bool
		var currentOrg *UserOrg
		for _, org := range ret.Orgs {
			if org.Org.ID == userRole.OrgID {
				orgExist = true
				currentOrg = org
				break
			}
		}
		if !orgExist {
			return nil, fmt.Errorf("user %v role %v org %v not exist in orgs", user.ID, userRole.RoleID, userRole.OrgID)
		}
		var roleExist bool
		for _, orgRole := range orgRoles {
			if orgRole.OrgID == userRole.OrgID && orgRole.RoleID == userRole.RoleID {
				roleExist = true
				currentOrg.Roles = append(currentOrg.Roles, RoleIDName{
					ID:       orgRole.RoleID,
					Name:     orgRole.Name,
					IsAdmin:  orgRole.IsAdmin,
					IsSystem: orgRole.OrgID == config.TopOrgID(),
				})
			}
		}
		if !roleExist {
			// not an org role, check global roles
			if globalRole, ok := globalRoleMap[userRole.RoleID]; ok {
				roleExist = true
				currentOrg.Roles = append(currentOrg.Roles, RoleIDName{
					ID:       globalRole.RoleID,
					Name:     globalRole.Name,
					IsGlobal: true,
				})
			}
		}
		if !roleExist {
			return nil, fmt.Errorf("user %v role %v org %v not exist in org roles", user.ID, userRole.RoleID, userRole.OrgID)
		}
	}
	return ret, nil
}

func getUserOrgsTx(tx *gorm.DB, userID uint32, orgTree *model.OrgNode) ([]OrgUserIDName, error) {
	var ret []OrgUserIDName
	var userOrgs []*model.OrgUser
	if err := sqlopt.WithUserID(userID).Apply(tx).Find(&userOrgs).Error; err != nil {
		return nil, fmt.Errorf("get user %v orgs err: %v", userID, err)
	}
	for _, orgUser := range userOrgs {
		var status bool
		if orgUser.Status != sqlopt.OrgUserStatusDisabled {
			status = true
		}
		ret = append(ret, OrgUserIDName{IDName: IDName{ID: orgUser.OrgID, Name: orgTree.GetFullName(orgUser.OrgID)}, Status: status})
	}
	return ret, nil
}

// ID: orgUser.OrgID, Name: orgTree.GetFullName(orgUser.OrgID),Status: orgUser.Status
func getCreatorTx(tx *gorm.DB, creatorID uint32) (IDName, error) {
	ret := IDName{ID: creatorID}
	creator := &model.User{}
	if err := sqlopt.WithID(creatorID).Apply(tx).First(creator).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return ret, fmt.Errorf("get creator %v err: %v", creatorID, err)
		}
	} else {
		ret.Name = creator.Name
	}
	if ret.Name == "" {
		ret.NameStatus = toErrStatus("iam_user_deleted")
	}
	return ret, nil
}

func checkUserIsAdmin(tx *gorm.DB, userID uint32, org *model.Org) (bool, []*model.UserRole, error) {
	// check org
	if !org.Status {
		return false, nil, fmt.Errorf("org %v status false", org.ID)
	}
	// check user admin
	if userID == config.AdminUserID() {
		return true, nil, nil
	}
	// 一次性查出用户在所有组织的 org_users 和 user_roles（is_admin=true）
	var orgUserList []*model.OrgUser
	if err := sqlopt.WithUserID(userID).Apply(tx).Find(&orgUserList).Error; err != nil {
		return false, nil, fmt.Errorf("get user %v org users err: %v", userID, err)
	}
	// org_users status map: orgID -> status
	orgUserStatusMap := make(map[uint32]string)
	for _, ou := range orgUserList {
		orgUserStatusMap[ou.OrgID] = ou.Status
	}
	// user_roles is_admin map: orgID -> true（用户在该组织有管理员角色）
	userAdminOrgMap := make(map[uint32]bool)
	var userRoles []*model.UserRole
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithAdmin(true),
	).Apply(tx).Find(&userRoles).Error; err != nil {
		return false, nil, fmt.Errorf("get user %v roles err: %v", userID, err)
	}
	for _, ur := range userRoles {
		userAdminOrgMap[ur.OrgID] = true
	}
	// 从当前组织开始，沿祖先向上找，第一个满足两个条件的节点即为管理员
	// 条件：org_users.status != disabled && user_roles.is_admin = true
	// 先查当前组织，再查祖先
	orgTree, err := getOrgTree(tx)
	if err != nil {
		return false, nil, err
	}
	checkOrgIDs := []uint32{org.ID}
	checkOrgIDs = append(checkOrgIDs, orgTree.GetAncestorIDs(org.ID)...)
	for _, oid := range checkOrgIDs {
		if orgUserStatusMap[oid] != sqlopt.OrgUserStatusDisabled && userAdminOrgMap[oid] {
			return true, nil, nil
		}
	}
	// 没有找到：返回该组织下的非管理员角色（用于后续权限检查）
	var ret []*model.UserRole
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(org.ID),
	).Apply(tx).Find(&userRoles).Error; err != nil {
		return false, nil, fmt.Errorf("get user %v org %v roles err: %v", userID, org.ID, err)
	}
	for _, ur := range userRoles {
		if !ur.IsAdmin {
			ret = append(ret, ur)
		}
	}
	// also include global roles assigned in this org
	var globalUserRoles []*model.UserRole
	globalRolesQuery := sqlopt.WithStatus(true).Apply(tx).Select("role_id").Table("global_roles")
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(org.ID),
	).Apply(tx).Where("role_id IN (?)", globalRolesQuery).Find(&globalUserRoles).Error; err != nil {
		return false, nil, fmt.Errorf("get user %v global roles err: %v", userID, err)
	}
	ret = append(ret, globalUserRoles...)
	return false, ret, nil
}

func getUserPermission(tx *gorm.DB, userID, orgID uint32) (*Permission, error) {
	// org
	org := &model.Org{}
	if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
		return nil, fmt.Errorf("get org %v err: %v", orgID, err)
	}
	// check org user 用户在组织中的状态校验
	orgUser := &model.OrgUser{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
	).Apply(tx).First(orgUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get org %v user %v err: %v", orgID, userID, err.Error())
		}
	} else if orgUser.Status == sqlopt.OrgUserStatusDisabled {
		return nil, fmt.Errorf("org %v user %v disable", orgID, userID)
	}
	// user role
	var userRoles []*model.UserRole
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
	).Apply(tx).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("get org %v user %v role err: %v", orgID, userID, err)
	}
	// org role
	var roleIDs []uint32
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}
	var orgRoles []*model.OrgRole
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgID),
		sqlopt.WithRoles(roleIDs),
	).Apply(tx).Find(&orgRoles).Error; err != nil {
		return nil, fmt.Errorf("get org %v roles %v err: %v", orgID, roleIDs, err)
	}
	// global role: resolve roleIDs not covered by org_roles against global_roles
	// so users with global roles get their roles and permissions.
	orgRoleExist := make(map[uint32]bool)
	for _, orgRole := range orgRoles {
		orgRoleExist[orgRole.RoleID] = true
	}
	var missingRoleIDs []uint32
	for _, roleID := range roleIDs {
		if !orgRoleExist[roleID] {
			missingRoleIDs = append(missingRoleIDs, roleID)
		}
	}
	globalRoleMap := make(map[uint32]*model.GlobalRole)
	if len(missingRoleIDs) > 0 {
		var globalRoles []*model.GlobalRole
		if err := sqlopt.WithRoles(missingRoleIDs).Apply(tx).Find(&globalRoles).Error; err != nil {
			return nil, fmt.Errorf("get global roles %v err: %v", missingRoleIDs, err)
		}
		for i := range globalRoles {
			globalRoleMap[globalRoles[i].RoleID] = globalRoles[i]
		}
	}
	// check if user is org admin
	isAdmin, validUserRoles, err := checkUserIsAdmin(tx, userID, org)
	if err != nil {
		return nil, err
	}
	ret := &Permission{
		IsAdmin:  isAdmin,
		IsSystem: orgID == config.TopOrgID(),
		Org:      IDName{ID: org.ID, Name: org.Name},
	}
	// user
	user := &model.User{}
	if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
		return nil, fmt.Errorf("get user %v err: %v", userID, err)
	}
	ret.LastUpdatePasswordAt = user.LastUpdatePasswordAt

	for _, userRole := range userRoles {
		var matched bool
		for _, orgRole := range orgRoles {
			if orgRole.RoleID == userRole.RoleID {
				matched = true
				ret.Roles = append(ret.Roles, RoleIDName{
					ID:       orgRole.RoleID,
					Name:     orgRole.Name,
					IsAdmin:  orgRole.IsAdmin,
					IsSystem: orgRole.OrgID == config.TopOrgID(),
				})
			}
		}
		if !matched {
			if globalRole, ok := globalRoleMap[userRole.RoleID]; ok {
				ret.Roles = append(ret.Roles, RoleIDName{
					ID:       globalRole.RoleID,
					Name:     globalRole.Name,
					IsGlobal: true,
					IsSystem: true,
				})
			}
		}
	}
	if isAdmin {
		return ret, nil
	}
	// perms
	for _, userRole := range validUserRoles {
		roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(userRole.RoleID))
		if err != nil {
			return nil, fmt.Errorf("get role %v err: %v", userRole.RoleID, err)
		}
		if !roleDetail.Enable {
			continue
		}
		for _, p1 := range roleDetail.Perms {
			var exist bool
			for _, p2 := range ret.Perms {
				if p2.Perm == string(p1.Obj) {
					exist = true
					break
				}
			}
			if !exist {
				ret.Perms = append(ret.Perms, Perm{Perm: string(p1.Obj)})
			}
		}
	}
	return ret, nil
}
