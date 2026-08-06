package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// AuthAppPublish 校验当前用户对请求中指定 appId 的已发布 App 的访问权限。
// appIdField: 请求中 appId 字段的路径，支持顶层（"appId"）和嵌套（"config.appId"）。
// appType: 该 App 的类型，如 constant.AppTypeAgent / AppTypeRag / AppTypeWorkflow 等。
func AuthAppPublish(appIdField, appType string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 取 appId 值（复用 auth_model.go 的 extractFieldsFromRequest，支持 query/body/嵌套路径）
		appIds := extractFieldsFromRequest(ctx, []string{appIdField})
		if len(appIds) == 0 || appIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("appId is required"))
			ctx.Abort()
			return
		}
		appId := appIds[0]

		// 2. 取请求者 userId / orgId
		userId, err := getUserID(ctx)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		orgId, err := getOrgID(ctx)
		if err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}

		// 3. 校验发布范围权限
		if err := service.CheckAppPublishAccess(ctx, appId, appType, userId, orgId); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
