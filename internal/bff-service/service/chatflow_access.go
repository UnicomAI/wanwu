package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckChatflowConversationAccess 校验会话属于请求者，且会话所属 app 与当前 workflow 的应用映射一致。
func CheckChatflowConversationAccess(ctx *gin.Context, conversationId, workflowId, userId, orgId string) error {
	conversation, err := app.GetConversationByID(ctx.Request.Context(), &app_service.GetConversationByIDReq{
		ConversionId: conversationId,
	})
	if err != nil {
		return err
	}
	if conversation.GetUserId() != userId ||
		conversation.GetOrgId() != orgId ||
		conversation.GetAppType() != constant.AppTypeChatflow {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission")
	}

	appInfo, err := app.GetChatflowApplication(ctx.Request.Context(), &app_service.GetChatflowApplicationReq{
		OrgId:      orgId,
		UserId:     userId,
		WorkflowId: workflowId,
	})
	if err != nil {
		return err
	}
	if appInfo.GetApplicationId() == "" || appInfo.GetApplicationId() != conversation.GetAppId() {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission")
	}
	return nil
}
