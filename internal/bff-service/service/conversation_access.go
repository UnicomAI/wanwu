package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckConversationOwner 校验 conversationId 属于指定 assistantId 且属于当前用户。
// 应用 owner 可访问该应用下所有用户的会话（与会话日志的权限模型一致）；非 owner 仅可访问属于自己的会话。
// 任一条件不符（会话不存在、属于其他智能体、属于其他用户/组织）即返回错误。
func CheckConversationOwner(ctx *gin.Context, conversationId, assistantId, userId, orgId string) error {
	resp, err := assistant.GetConversationOwner(ctx.Request.Context(), &assistant_service.GetConversationOwnerReq{
		ConversationId: conversationId,
	})
	if err != nil {
		return err
	}
	if resp.GetAssistantId() != assistantId {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission", "conversation not belong to current assistant")
	}
	// 应用 owner 放行
	ownerUserId, ownerOrgId, err := OwnerInfo(ctx, constant.BizModuleAppAgent, assistantId)
	if err != nil {
		return err
	}
	isOwner := ownerUserId == userId && ownerOrgId == orgId
	if !isOwner && (resp.GetUserId() != userId || resp.GetOrgId() != orgId) {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission", "conversation not belong to current user")
	}
	return nil
}

// CheckConversationLogOwner 校验 logIds 列表均属于指定 appId + appType 且属于当前用户。
// appId / appType 始终校验；若当前请求者为应用 owner，则不校验日志 userId/orgId，否则需校验日志属主与请求者一致。
// logIds 为空表示全量导出：跳过逐条日志归属校验，仅校验当前用户对该 appId + appType 的访问权限。
// 任一条件不符（日志不存在、数量不符、属于其他应用、属于其他用户/组织、无 app 访问权限）即返回错误。
func CheckConversationLogOwner(ctx *gin.Context, logIds []string, appId, appType, userId, orgId string) error {
	// 全量导出：仅校验当前用户对该 appId + appType 的访问权限，由后续业务层根据 owner 范围限定导出数据。
	if len(logIds) == 0 {
		return CheckAppPublishAccess(ctx, appId, appType, userId, orgId)
	}
	resp, err := app.GetConversationLogByLogIds(ctx.Request.Context(), &app_service.GetConversationLogByLogIdsReq{
		LogIds: logIds,
	})
	if err != nil {
		return err
	}
	if len(resp.GetItems()) != len(logIds) {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_log_no_permission", "conversation log not belong to current user or app")
	}
	// 应用 owner 可见该应用下所有用户的日志，非 owner 仅可见自身日志
	ownerUserId, ownerOrgId, err := OwnerInfo(ctx, appType, appId)
	if err != nil {
		return err
	}
	isOwner := ownerUserId == userId && ownerOrgId == orgId
	for _, item := range resp.GetItems() {
		if item.GetAppId() != appId || item.GetAppType() != appType {
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_log_no_permission", "conversation log not belong to current user or app")
		}
		if !isOwner && (item.GetUserId() != userId || item.GetOrgId() != orgId) {
			return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_log_no_permission", "conversation log not belong to current user or app")
		}
	}
	return nil
}
