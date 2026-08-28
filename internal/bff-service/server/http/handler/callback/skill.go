package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SearchBuiltInSkillList
//
//	@Tags			skill
//	@Summary		获取内置skill详情列表
//	@Description	获取内置skill详情列表
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SearchBuiltinSkillListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.SkillDetailListResp}
//	@Router			/callback/v1/skill/builtin/list [post]
func SearchBuiltInSkillList(ctx *gin.Context) {
	var req request.SearchBuiltinSkillListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// UserId/OrgId 为空时 GetAgentSkillListDetail 会自动跳过 vars 填充（保持老 caller 行为）。
	resp, err := service.GetAgentSkillListDetail(ctx, req.UserId, req.OrgId, req.SkillIdList)
	gin_util.Response(ctx, resp, err)
}

// SearchCustomSkillList
//
//	@Tags			skill
//	@Summary		获取自定义skill详情列表
//	@Description	获取自定义skill详情列表
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SearchCustomSkillListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.CustomSkillDetailListResp}
//	@Router			/callback/v1/skill/custom/list [post]
func SearchCustomSkillList(ctx *gin.Context) {
	var req request.SearchCustomSkillListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// req.UserId / req.OrgId 是智能体（Assistant）创建者身份，由 assistant-service 传入，
	// 不是 HTTP 调用者身份；透传给 service 层，供其观测/审计使用。
	resp, err := service.GetCustomSkillListDetail(ctx, req.UserId, req.OrgId, req.SkillIdList)
	gin_util.Response(ctx, resp, err)
}

// SearchAcquiredSkillList
//
//	@Tags			skill
//	@Summary		获取我添加skill详情列表
//	@Description	获取我添加skill详情列表
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SearchAcquiredSkillListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.CallbackAcquiredSkillDetailListResp}
//	@Router			/callback/v1/skill/acquired/list [post]
func SearchAcquiredSkillList(ctx *gin.Context) {
	var req request.SearchAcquiredSkillListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// req.UserId / req.OrgId 是智能体（Assistant）创建者身份，由 assistant-service 传入，
	// 不是 HTTP 调用者身份；透传给 service 层，供其观测/审计使用。
	resp, err := service.GetCallbackAcquiredSkillListDetail(ctx, req.UserId, req.OrgId, req.SkillIdList)
	gin_util.Response(ctx, resp, err)
}
