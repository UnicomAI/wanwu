package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type AdminCheckType int

const (
	AdminCommonCheck AdminCheckType = 0 // 通用校验
	AdminOrgCheck    AdminCheckType = 1 // 参数中选中组织校验
	AdminBizCheck    AdminCheckType = 2 // 业务校验
)

type AdminCenterBiz struct {
	BizType string
	BizId   string
}

// AuthAdminCenter 管理员中心权限校验
func AuthAdminCenter(checkType AdminCheckType, adminCenterBiz *AdminCenterBiz) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
			if panicOccur {
				log.Errorf("panic occur, err %v", recoverError)
				gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("admin center system error"))
				ctx.Abort()
				return
			}
		})
		userId, orgId, err := userInfo(ctx)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		//1.校验当前用户必须是admin 或者是系统管理员角色
		err = authUserAdmin(ctx)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("user permit valid fail"))
			ctx.Abort()
			return
		}

		//2.校验当前管理员中心业务权限
		err = authAdminCenterBiz(ctx, checkType, adminCenterBiz, userId, orgId)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("user biz permit valid fail"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// 校验用户是否是超级管理员或系统管理员
func authUserAdmin(ctx *gin.Context) error {
	if isAdmin(ctx) {
		return nil
	}
	return errors.New("user is not admin")
}

func authAdminCenterBiz(ctx *gin.Context, checkType AdminCheckType, adminCenterBiz *AdminCenterBiz, userId, orgId string) error {
	if isSystem(ctx) { // 系统管理员,无需校验业务权限
		return nil
	}
	switch checkType {
	case AdminOrgCheck:
		return authAdminBizOrgID(ctx, userId, orgId)
	case AdminBizCheck:
		return authAdminBiz(ctx, adminCenterBiz, userId, orgId)
	}
	return nil
}

// 校验用户是否有选取的组织权限
func authAdminBizOrgID(ctx *gin.Context, userId, orgId string) error {
	userSelect, err := adminUserSelect(ctx)
	if err != nil {
		return err
	}
	if len(userSelect.OrgIdList) == 0 {
		if !userSelect.IsAllOrg {
			return errors.New("orgIdList can not be empty when isAllOrg is false")
		}
		// 选择全部组织,则直接跳过不校验
		return nil
	}
	return authUserOrgOrSubOrg(ctx, userId, userSelect.OrgIdList)
}

func authAdminBiz(ctx *gin.Context, adminCenterBiz *AdminCenterBiz, userId, orgId string) error {
	value := getFieldValue(ctx, adminCenterBiz.BizId)
	if len(value) == 0 {
		return errors.New("biz id params is empty")
	}
	_, ownerOrgID, err := service.OwnerInfo(ctx, adminCenterBiz.BizType, value)
	if err != nil {
		return err
	}
	if ownerOrgID == orgId {
		return nil
	}
	return authUserOrgOrSubOrg(ctx, userId, []string{ownerOrgID})
}

// 校验用户是否有组织或者子组织权限
func authUserOrgOrSubOrg(ctx *gin.Context, userId string, orgIdList []string) error {
	pass := service.IsAdminInOrgs(ctx, userId, orgIdList...)
	if !pass {
		marshal, _ := json.Marshal(orgIdList)
		log.Errorf("user org permit valid fail, userId %s, orgIdList %s", userId, string(marshal))
		return errors.New("user org permit valid fail")
	}
	return nil
}

// 请求参数中选择的用户信息
func adminUserSelect(ctx *gin.Context) (*request.AdminUserSelect, error) {
	body, err := requestBody(ctx)
	if err != nil {
		return nil, err
	}
	userSelect := request.AdminUserSelect{}
	err = json.Unmarshal([]byte(body), &userSelect)
	return &userSelect, err
}

// 获取用户信息
func userInfo(ctx *gin.Context) (userId, orgId string, err error) {
	// userID
	userID, err := getUserID(ctx)
	if err != nil {
		return "", "", err
	}

	// orgID
	orgID, err := getOrgID(ctx)
	if err != nil {
		return "", "", err
	}
	return userID, orgID, nil
}
