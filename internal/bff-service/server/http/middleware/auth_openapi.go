package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
)

func AuthOpenAPIKey() func(*gin.Context) {
	return func(ctx *gin.Context) {
		token, err := getApiKey(ctx)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		apiKey, err := service.GetApiKeyByKey(ctx, token)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusUnauthorized, err)
			ctx.Abort()
			return
		}
		if !apiKey.Status {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "api key disabled")
			ctx.Abort()
			return
		}
		if apiKey.ExpiredAt != 0 && apiKey.ExpiredAt < time.Now().UnixMilli() {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "api key expired")
			ctx.Abort()
			return
		}
		ctx.Set(gin_util.USER_ID, apiKey.UserId)
		ctx.Set(gin_util.X_ORG_ID, apiKey.OrgId)
		ctx.Set(gin_util.API_KEY_ID, apiKey.KeyId)

		// OpenAPI 的身份组织必须来自 API Key 本身。
		// middleware.getOrgID 优先读取请求头 X-Org-Id；若不覆盖该请求头，
		// 调用方可以伪造 X-Org-Id 绕过 AuthKnowledge 等基于组织的数据权限校验。
		ctx.Request.Header.Set(gin_util.X_ORG_ID, apiKey.OrgId)
	}
}

func AuthAppKeyByQuery(appType string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		token, err := getAppKeyByQuery(ctx)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		appKey, err := service.GetAppKeyByKey(ctx, token)
		if err != nil {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, err.Error())
			ctx.Abort()
			return
		}
		if appKey.AppType != appType {
			gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFAuth), nil, "invalid appType")
			ctx.Abort()
			return
		}
		ctx.Set(gin_util.USER_ID, appKey.UserId)
		ctx.Set(gin_util.X_ORG_ID, appKey.OrgId)
		ctx.Set(gin_util.APP_ID, appKey.AppId)
	}

}

// --- internal ---
func getApiKey(ctx *gin.Context) (string, error) {
	authorization := ctx.Request.Header.Get("Authorization")
	if authorization != "" {
		tks := strings.Split(authorization, " ")
		if len(tks) > 1 && tks[0] == "Bearer" {
			return tks[1], nil
		} else {
			return "", fmt.Errorf("not Bearer token format")
		}
	} else {
		return "", fmt.Errorf("token is nil")
	}
}

func getAppKeyByQuery(ctx *gin.Context) (string, error) {
	key := ctx.Query("key")
	if key != "" {
		return key, nil
	} else {
		return "", fmt.Errorf("token is nil")
	}
}
