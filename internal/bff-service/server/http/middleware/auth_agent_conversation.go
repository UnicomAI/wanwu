package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// AuthAgentConversationAccess 校验请求中 conversationId 属于当前用户。
// conversationIdField: 请求中字段路径，支持顶层（"conversationId"）和嵌套（"data.conversationId"）。
// 注意：/assistant/stream/connect 与 /assistant/stream/cancel 的下游没有身份校验，
// 本中间件是这两条链路唯一的会话归属守卫，不要从路由中移除。
func AuthAgentConversationAccess(conversationIdField string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 取 conversationId 值（复用 auth_model.go 的 extractFieldsFromRequest，支持 query/body/嵌套路径）
		conversationIds := extractFieldsFromRequest(ctx, []string{conversationIdField})
		if len(conversationIds) == 0 || conversationIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("conversationId is required"))
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
		if err := service.CheckAgentConversationAccess(ctx, conversationIds[0], userId, orgId); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
