package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// AuthRagConversationAccess 校验请求中 conversationId 属于指定 ragId 且属于当前用户
func AuthRagConversationAccess(ragIdField, conversationIdField string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 取 conversationId / ragId 值（复用 auth_model.go 的 extractFieldsFromRequest，支持 query/body/嵌套路径）
		conversationIds := extractFieldsFromRequest(ctx, []string{conversationIdField})
		if len(conversationIds) == 0 || conversationIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("conversationId is required"))
			ctx.Abort()
			return
		}
		ragIds := extractFieldsFromRequest(ctx, []string{ragIdField})
		if len(ragIds) == 0 || ragIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("ragId is required"))
			ctx.Abort()
			return
		}

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

		// 3. 校验会话归属
		if err := service.CheckRagConversationAccess(ctx, conversationIds[0], ragIds[0], userId, orgId); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
