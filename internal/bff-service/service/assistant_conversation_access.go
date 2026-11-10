package service

import (
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckAgentConversationAccess 校验 conversationId 属于当前用户及其组织。
// 仅按会话归属校验，不绑定具体 app：会话属主 userId/orgId 与调用者一致即放行。
// 任一条件不符（会话不存在、属于其他用户/组织）即返回错误。
func CheckAgentConversationAccess(ctx *gin.Context, conversationId, userId, orgId string) error {
	resp, err := assistant.GetConversationOwner(ctx.Request.Context(), &assistant_service.GetConversationOwnerReq{
		ConversationId: conversationId,
	})
	if err != nil {
		return err
	}
	if resp.GetUserId() != userId || resp.GetOrgId() != orgId {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission", "conversation not belong to current user")
	}
	return nil
}
