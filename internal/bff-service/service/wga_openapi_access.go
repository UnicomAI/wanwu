package service

import (
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckWgaConversationOwner 校验 WGA thread 属于请求者。查不到即拒绝，不自动创建。
func CheckWgaConversationOwner(ctx *gin.Context, threadID, userId, orgId string) error {
	threadID = strings.TrimSpace(threadID)
	resp, err := assistant.WgaConversationExists(ctx.Request.Context(), &assistant_service.WgaConversationExistsReq{
		ThreadId: threadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}
	if !resp.GetExists() {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_conversation_no_permission")
	}
	return nil
}
