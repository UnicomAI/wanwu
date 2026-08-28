package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetSkillDetail
//
//	@Tags			skill
//	@Summary		获取技能详情
//	@Description	根据skillId和skillType获取技能详情
//	@Accept			json
//	@Produce		json
//	@Param			skillId		query		string	true	"skillId"
//	@Param			skillType	query		string	true	"skillType (builtin/custom)"
//	@Success		200			{object}	response.Response{data=response.CallbackSkillDetail}
//	@Router			/callback/v1/skill/detail [get]
func GetSkillDetail(ctx *gin.Context) {
	resp, err := service.GetCallbackSkillDetail(ctx, ctx.Query("skillId"), ctx.Query("skillType"))
	gin_util.Response(ctx, resp, err)
}
