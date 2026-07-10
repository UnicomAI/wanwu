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
