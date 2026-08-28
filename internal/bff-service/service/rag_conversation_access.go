package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckRagConversationAccess 校验会话属于请求中的 ragId 且属于请求者本人
func CheckRagConversationAccess(ctx *gin.Context, conversationId, ragId, userId, orgId string) error {
	resp, err := rag.GetRagConversationOwner(ctx.Request.Context(), &rag_service.GetRagConversationOwnerReq{
		ConversationId: conversationId,
	})
	if err != nil {
		return err
	}
	if resp.GetRagId() != ragId {
		return ragConversationNoPermission()
	}
	if resp.GetUserId() != userId || resp.GetOrgId() != orgId {
		return ragConversationNoPermission()
	}
	return nil
}

func ragConversationNoPermission() error {
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_conversation_no_permission")
}
