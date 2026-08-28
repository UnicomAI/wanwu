package middleware

import (
	"errors"
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// AuthConversationOwner 校验请求中 conversationId 属于指定 assistantId 且属于当前用户。
// conversationIdField / assistantIdField: 请求中字段路径，支持顶层（"conversationId"）和嵌套（"data.conversationId"）。
func AuthConversationOwner(conversationIdField, assistantIdField string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 取 conversationId / assistantId 值（复用 auth_model.go 的 extractFieldsFromRequest，支持 query/body/嵌套路径）
		conversationIds := extractFieldsFromRequest(ctx, []string{conversationIdField})
		if len(conversationIds) == 0 || conversationIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("conversationId is required"))
			ctx.Abort()
			return
		}
		assistantIds := extractFieldsFromRequest(ctx, []string{assistantIdField})
		if len(assistantIds) == 0 || assistantIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("assistantId is required"))
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
		if err := service.CheckConversationOwner(ctx, conversationIds[0], assistantIds[0], userId, orgId); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// AuthConversationLogOwner 校验请求中 logId 列表均属于指定 appId + appType 且属于当前用户。
// logIdField / appIdField / appTypeField: 请求中字段路径，支持顶层（"logIds"）和嵌套（"data.logIds"）；
// logIdField 支持 string 单值与 []string 数组。
func AuthConversationLogOwner(logIdField, appIdField, appTypeField string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()

		// 1. 取 logId 列表 / appId / appType 值（logIdField 支持数组与单值，其余复用 string 提取）
		//    logIds 为空表示全量导出：无需逐条校验日志归属，仅校验当前用户对该 appId + appType 的访问权限。
		logIds := extractFieldListFromRequest(ctx, logIdField)
		appIds := extractFieldsFromRequest(ctx, []string{appIdField})
		if len(appIds) == 0 || appIds[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("appId is required"))
			ctx.Abort()
			return
		}
		appTypes := extractFieldsFromRequest(ctx, []string{appTypeField})
		if len(appTypes) == 0 || appTypes[0] == "" {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("appType is required"))
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

		// 3. 校验会话日志归属（logIds 为空时仅校验 app 访问权限，表示全量导出）
		if err := service.CheckConversationLogOwner(ctx, logIds, appIds[0], appTypes[0], userId, orgId); err != nil {
			gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, err)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
