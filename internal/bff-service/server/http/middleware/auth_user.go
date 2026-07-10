package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/gin-gonic/gin"
)

var (
	accessRouter = map[string]bool{
		"/v1/user/password":         true,
		"/v1/user/permission":       true,
		"/v1/user/info":             true,
		"/v1/org/select":            true,
		"/v1/org/admin/select":      true,
		"/v1/user/login/email/code": true,
		"/v1/user/login":            true,
	}
)

func CheckUserEnable(ctx *gin.Context) {
	httpStatus := http.StatusForbidden
	// userID
	userID, err := getUserID(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	ctx.Set(gin_util.USER_ID, userID)
	// genTokenTS
	genTokenTS, err := getGenTokenTS(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	// check
	resp, err := service.CheckUserEnable(ctx, userID, genTokenTS)
	if err != nil {
		gin_util.ResponseErrWithStatus(ctx, httpStatus, err)
		ctx.Abort()
		return
	}
	err = checkPasswordUpdateAccess(ctx, resp.LastUpdatePasswordAt)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	ctx.Set(gin_util.X_LANGUAGE, resp.Language)
}

func CheckUserPerm(ctx *gin.Context) {
	httpStatus := http.StatusForbidden
	// userID
	userID, err := getUserID(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	ctx.Set(gin_util.USER_ID, userID)
	// genTokenTS
	genTokenTS, err := getGenTokenTS(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	// orgID
	orgID, err := getOrgID(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	// tags
	tags, ok := route.GetTags(ctx.FullPath(), ctx.Request.Method)
	if !ok {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFGeneral, "", fmt.Sprintf("auth path [%v]%v not found", ctx.Request.Method, ctx.FullPath()))
		ctx.Abort()
		return
	}
	// admin_center 一级命名空间：权限已由路由层中间件（CheckOrgAdmin 等）校验，
	// CheckUserPerm 只需校验用户有效性并设置上下文。
	// 子模块（admin_center.xxx）走正常权限校验。
	if len(tags) == 1 && tags[0] == "admin_center" {
		resp, err := service.CheckUserEnable(ctx, userID, genTokenTS)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, httpStatus, err)
			ctx.Abort()
			return
		}
		err = checkPasswordUpdateAccess(ctx, resp.LastUpdatePasswordAt)
		if err != nil {
			gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
			ctx.Abort()
			return
		}
		ctx.Set(gin_util.IS_ADMIN, true)
		ctx.Set(gin_util.IS_SYSTEM, orgID == config.TopOrgID)
		return
	}
	// check
	resp, err := service.CheckUserPerm(ctx, userID, genTokenTS, orgID, tags)
	if err != nil {
		gin_util.ResponseErrWithStatus(ctx, httpStatus, err)
		ctx.Abort()
		return
	}
	err = checkPasswordUpdateAccess(ctx, resp.LastUpdatePasswordAt)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, httpStatus, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}
	ctx.Set(gin_util.IS_ADMIN, resp.IsAdmin)
	ctx.Set(gin_util.IS_SYSTEM, resp.IsSystem)

}

// --- internal ---

func getUserID(ctx *gin.Context) (string, error) {
	claims, ok := ctx.Get(gin_util.CLAIMS)
	if ok {
		if customClaims, ok := claims.(*jwt_util.CustomClaims); ok {
			if userID := customClaims.UserID; userID != "" {
				return userID, nil
			}
		}
	}

	if userID := ctx.GetString(gin_util.USER_ID); userID != "" {
		return userID, nil
	}

	return "", errors.New("user id empty")
}

func getOrgID(ctx *gin.Context) (string, error) {
	if orgID := ctx.GetHeader(gin_util.X_ORG_ID); orgID != "" {
		return orgID, nil
	}
	if orgID := ctx.GetString(gin_util.X_ORG_ID); orgID != "" {
		return orgID, nil
	}
	return "", errors.New("org id empty")
}

func getGenTokenTS(ctx *gin.Context) (string, error) {
	claims, ok := ctx.Get(gin_util.CLAIMS)
	if !ok {
		return "", errors.New("jwt claims empty")
	}
	return strconv.Itoa(int(claims.(*jwt_util.CustomClaims).ExpiresAt * 1000)), nil
}

func checkPasswordUpdateAccess(ctx *gin.Context, lastUpdatePasswordAt int64) error {
	// 如果 lastUpdatePasswordAt 不为 0，则说明用户已经更新过密码 -> 放行
	if lastUpdatePasswordAt != 0 {
		return nil
	}

	// 获取当前路由
	router := ctx.FullPath()

	// 如果 lastUpdatePasswordAt == 0，但当前路由不在白名单内 → 拒绝访问
	if !accessRouter[router] {
		return errors.New("please change your password")
	}

	// 在 accessRouter 内，允许访问
	return nil
}
