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
		// check org user Check the status of the user in the organization
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

func (c *Client) GetUsers(ctx context.Context, orgID uint32, name string, offset, limit int32) ([]*UserInfo, int64, *errs.Status) {
	var ret []*UserInfo
	var count int64
	return ret, count, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// users
		var users []*model.User
		orgUsersQuery := sqlopt.WithOrgID(orgID).Apply(tx).Select("user_id").Table("org_users")
		if err := sqlopt.LikeName(name).Apply(tx).Where("id IN (?)", orgUsersQuery).
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

func (c *Client) CreateUser(ctx context.Context, user *model.User, orgID uint32, roleIDs []uint32) (uint32, *errs.Status) {
	if user.ID != 0 {
		return 0, toErrStatus("iam_user_create", "create user but id not 0")
	}
	return user.ID, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		return createUserTx(tx, user, orgID, roleIDs)
	})
}

func createUserTx(tx *gorm.DB, user *model.User, orgID uint32, roleIDs []uint32) *errs.Status {
	// check org
	var adminRoles, nonAdminRoles []uint32
	if err := sqlopt.WithID(orgID).Apply(tx).First(&model.Org{}).Error; err != nil {
		return toErrStatus("iam_user_create", err.Error())
	} else if len(roleIDs) > 0 {
		// check org role
		for _, roleID := range roleIDs {
			orgRole := &model.OrgRole{}
			if err := sqlopt.SQLOptions(
				sqlopt.WithOrgID(orgID),
				sqlopt.WithRoleID(roleID),
			).Apply(tx).First(orgRole).Error; err != nil {
				return toErrStatus("iam_user_create", err.Error())
			}
			if orgRole.IsAdmin {
				adminRoles = append(adminRoles, roleID)
			} else {
				nonAdminRoles = append(nonAdminRoles, roleID)
			}
		}
	}
	// check creator
	if user.CreatorID != 0 {
		// Create user normally
		if user.IsAdmin {
			return toErrStatus("iam_user_create", "cannot create admin user")
		}
		if err := sqlopt.WithID(user.CreatorID).Apply(tx).First(&model.User{}).Error; err != nil {
			return toErrStatus("iam_user_create", err.Error())
		}
	} else {
		// Create an administrator within the system. There cannot be any users in the system at this time.
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
	if orgID != config.TopOrgID() { // Also join top-level organizations by default
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
		// check org user Check the status of the user in the organization
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
		var adminRoles, nonAdminRoles []uint32
		if len(roleIDs) > 0 {
			for _, roleID := range roleIDs {
				orgRole := &model.OrgRole{}
				if err := sqlopt.SQLOptions(
					sqlopt.WithOrgID(orgID),
					sqlopt.WithRoleID(roleID),
				).Apply(tx).First(orgRole).Error; err != nil {
					return toErrStatus("iam_user_update", err.Error())
				}
				if orgRole.IsAdmin {
					adminRoles = append(adminRoles, roleID)
				} else {
					nonAdminRoles = append(nonAdminRoles, roleID)
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
		// check password
		if user.Password != util.SHA256(pwd) {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), "password error")
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
	// user role
	var userRoles []*model.UserRole
	orgRolesQuery := sqlopt.WithStatus(true).Apply(tx).Select("role_id").Table("org_roles")
	if err := sqlopt.WithUserID(userID).Apply(tx).Where("role_id IN (?)", orgRolesQuery).Find(&userRoles).Error; err != nil {
		return false, nil, fmt.Errorf("get user %v role err: %v", userID, err)
	}
	if len(userRoles) == 0 {
		return false, nil, nil
	}
	// check if user is org admin
	orgTree, err := getOrgTree(tx)
	if err != nil {
		return false, nil, err
	}
	if orgTree.IsAdmin(org.ID, userRoles) {
		return true, nil, nil
	}
	var ret []*model.UserRole
	for _, userRole := range userRoles {
		if userRole.OrgID == org.ID {
			ret = append(ret, userRole)
		}
	}
	return false, ret, nil
}

func getUserPermission(tx *gorm.DB, userID, orgID uint32) (*Permission, error) {
	// org
	org := &model.Org{}
	if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
		return nil, fmt.Errorf("get org %v err: %v", orgID, err)
	}
	// check org user Check the status of the user in the organization
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
		for _, orgRole := range orgRoles {
			if orgRole.RoleID == userRole.RoleID {
				ret.Roles = append(ret.Roles, RoleIDName{
					ID:       orgRole.RoleID,
					Name:     orgRole.Name,
					IsAdmin:  orgRole.IsAdmin,
					IsSystem: orgRole.OrgID == config.TopOrgID(),
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
