package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetOntologySkillSelect
//
//	@Tags			ontology.digital_employee
//	@Summary		获取skill选择列表(Ontology专用)
//	@Description	获取skill选择列表(Ontology专用)
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"skill名称"
//	@Param			skillType	query		string	false	"skill类型(builtin/custom/acquired)"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.SkillInfo}}
//	@Router			/ontology/skill/select [get]
func GetOntologySkillSelect(ctx *gin.Context) {
	resp, err := service.GetSkillSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"), ctx.Query("skillType"), true)
	gin_util.Response(ctx, resp, err)
}
