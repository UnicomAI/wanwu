package orm

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access"
	"github.com/gromitlee/access/pkg/perm"
	"gorm.io/gorm"
)

func (c *Client) CheckUserOK(ctx context.Context, userID uint32, genTokenAt int64) (bool, string, int64, *errs.Status) {
	var needLogin bool
	var language string
	var lastUpdatePasswordAt int64

	return needLogin, language, lastUpdatePasswordAt, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_perm_check_user_ok", util.Int2Str(userID), err.Error())
		}
		// check need login
		if !checkUserNeedLogin(user, genTokenAt) {
			needLogin = true
			return toErrStatus("iam_perm_relogin")
		}
		// check status
		if !user.Status {
			return toErrStatus("iam_perm_user_disable")
		}
		// last_exec_at
		if err := tx.Model(user).Updates(map[string]interface{}{
			"last_exec_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return toErrStatus("iam_perm_check_user_ok", util.Int2Str(userID), err.Error())
		}
		language = user.Language
		lastUpdatePasswordAt = user.LastUpdatePasswordAt
		return nil
	})

}

func (c *Client) CheckUserPerm(ctx context.Context, userID uint32, genTokenAt int64, orgID uint32, oneOfPerms []perm.Perm) (bool, bool, string, int64, *errs.Status) {
	var needLogin, isAdmin bool
	var language string
	var lastUpdatePasswordAt int64
	return needLogin, isAdmin, language, lastUpdatePasswordAt, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
				util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), err.Error())
		}
		// check need login
		if !checkUserNeedLogin(user, genTokenAt) {
			needLogin = true
			return toErrStatus("iam_perm_relogin")
		}
		// check status
		if !user.Status {
			return toErrStatus("iam_perm_user_disable")
		}
		// org
		org := &model.Org{}
		if err := sqlopt.WithID(orgID).Apply(tx).First(org).Error; err != nil {
			return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
				util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), err.Error())
		}
		// check org user 用户在组织中的状态校验
		orgUser := &model.OrgUser{}
		if err := sqlopt.SQLOptions(
			sqlopt.WithUserID(userID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx).First(orgUser).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
					util.Int2Str(orgID), fmt.Sprintf("get org %v user %v", orgID, userID), err.Error())
			}
		} else if orgUser.Status == sqlopt.OrgUserStatusDisabled {
			return toErrStatus("iam_perm_user_disable")
		}
		// check if user is org admin
		var userRoles []*model.UserRole
		var err error
		isAdmin, userRoles, err = checkUserIsAdmin(tx, userID, org)
		if err != nil {
			return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
				util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), err.Error())
		} else if !isAdmin {
			// admin_center 命名空间（管理员中心）对任意组织的内置管理员放行：
			// 组织内置管理员角色 perms 为空，无法通过下方 perms 命中校验，但其管理范围本就涵盖
			// admin_center 接口（菜单显隐已由 isOrgAdmin 控制，此处接口鉴权保持一致语义）。
			// 注意：仅限 admin_center 命名空间，其它权限仍严格按当前组织 perms 校验，避免越权。
			if isAnyOrgAdmin(tx, userID) && isAllAdminCenterPerms(oneOfPerms) {
				// 放行，跳过 perms 命中校验
			} else {
				// check perm
				var ok bool
				for _, userRole := range userRoles {
					roleDetail, err := access.RBAC0GetRolePerms(tx, perm.Role(userRole.RoleID))
					if err != nil {
						return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
							util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), fmt.Sprintf("get role %v err: %v", userRole.RoleID, err.Error()))
					}
					if !roleDetail.Enable {
						return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
							util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), fmt.Sprintf("role %v status false", userRole.RoleID))
					}
					if roleDetail.IsAdmin {
						return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
							util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), fmt.Sprintf("role %v is admin", userRole.RoleID))
					}
					for _, onePerm := range oneOfPerms {
						for _, p := range roleDetail.Perms {
							if p == onePerm {
								ok = true
								break
							}
						}
						if ok {
							break
						}
					}
					if ok {
						break
					}
				}
				if !ok {
					return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
						util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), "no perm")
				}
			}
		}
		// last_exec_at
		if err := tx.Model(user).Updates(map[string]interface{}{
			"last_exec_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return toErrStatus("iam_perm_check_user_perm", util.Int2Str(userID),
				util.Int2Str(orgID), fmt.Sprintf("%v", oneOfPerms), err.Error())
		}
		language = user.Language
		lastUpdatePasswordAt = user.LastUpdatePasswordAt
		return nil
	})

}

// --- internal ---

func checkUserNeedLogin(user *model.User, genTokenAt int64) bool {
	return user.LastTokenAt <= genTokenAt
}

// isAnyOrgAdmin 判断用户在系统任一组织中是否拥有内置管理员角色（user_roles.is_admin=true）。
// 用于 admin_center 命名空间接口的放行：组织管理员的管理范围本就涵盖管理员中心。
// 同时排除用户在该组织被禁用（org_users.status=disable）的记录，与 checkUserIsAdmin 语义保持一致。
func isAnyOrgAdmin(tx *gorm.DB, userID uint32) bool {
	var count int64
	if err := tx.Model(&model.UserRole{}).
		Joins("JOIN org_users ON org_users.user_id = user_roles.user_id AND org_users.org_id = user_roles.org_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.is_admin = ?", true).
		Where("org_users.status IS NULL OR org_users.status != ?", sqlopt.OrgUserStatusDisabled).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// isAllAdminCenterPerms 判断请求的所有权限是否均属于 admin_center 命名空间。
// admin_center 顶层及其子分类（admin_center.setting / admin_center.oauth）均以 "admin_center" 为前缀。
// 仅当全部权限都属该命名空间时才适用「任意组织管理员放行」语义，避免越权放行其它权限。
func isAllAdminCenterPerms(perms []perm.Perm) bool {
	if len(perms) == 0 {
		return false
	}
	for _, p := range perms {
		obj := string(p.Obj)
		if obj != "admin_center" {
			return false
		}
	}
	return true
}
