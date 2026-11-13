package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckCustomSkillOwner 校验 customSkillId 属于当前用户及组织。
func CheckCustomSkillOwner(ctx *gin.Context, skillID, userId, orgId string) error {
	resp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: []string{skillID},
	})
	if err != nil {
		return err
	}
	for _, skill := range resp.GetSkillDetails() {
		if skill.GetSkillId() != skillID {
			continue
		}
		if skill.GetUserId() == userId && skill.GetOrgId() == orgId {
			return nil
		}
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_no_permission")
	}
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_no_permission")
}
