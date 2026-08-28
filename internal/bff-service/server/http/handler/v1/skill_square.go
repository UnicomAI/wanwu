package v1

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetSquareBuiltinSkillList
//
//	@Tags			exploration.skill
//	@Summary		获取广场内置skill列表
//	@Description	获取探索广场中的内置skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.BuiltinSkillInfo}}
//	@Router			/square/skill/builtin/list [get]
func GetSquareBuiltinSkillList(ctx *gin.Context) {
	resp, err := service.GetSquareBuiltinSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetSquareBuiltinSkillDetail
//
//	@Tags			exploration.skill
//	@Summary		获取广场内置skill详情
//	@Description	获取探索广场中的内置skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.BuiltinSkillDetail}
//	@Router			/square/skill/builtin/detail [get]
func GetSquareBuiltinSkillDetail(ctx *gin.Context) {
	resp, err := service.GetSquareBuiltinSkillDetail(ctx, ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// GetSquareShareSkillList
//
//	@Tags			exploration.skill
//	@Summary		获取广场共享skill列表
//	@Description	获取探索广场中的共享skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SharedSkillInfo}}
//	@Router			/square/skill/share/list [get]
func GetSquareShareSkillList(ctx *gin.Context) {
	resp, err := service.GetSquareShareSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// ShareSquareSkill
//
//	@Tags			exploration.skill
//	@Summary		添加共享skill到资源库
//	@Description	将共享skill添加到我的资源库
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	body		request.ShareSquareSkillReq	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/square/skill/share [post]
func ShareSquareSkill(ctx *gin.Context) {
	var req request.ShareSquareSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ShareSquareSkill(ctx, getUserID(ctx), getOrgID(ctx), req.SkillId)
	gin_util.Response(ctx, nil, err)
}

// GetSquareShareSkillDetail
//
//	@Tags			exploration.skill
//	@Summary		获取共享skill详情
//	@Description	获取共享skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.SharedSkillDetail}
//	@Router			/square/skill/share/detail [get]
func GetSquareShareSkillDetail(ctx *gin.Context) {
	resp, err := service.GetSquareShareSkillDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadSquareShareSkill
//
//	@Tags			exploration.skill
//	@Summary		下载共享skill
//	@Description	下载共享skill ZIP包
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/square/skill/share/download [get]
func DownloadSquareShareSkill(ctx *gin.Context) {
	skillId := ctx.Query("skillId")
	fileName := fmt.Sprintf("%s.zip", skillId)
	resp, err := service.DownloadSquareShareSkill(ctx, skillId)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, resp)
}

// GetSquareShareSkillVersionList
//
//	@Tags			exploration.skill
//	@Summary		获取共享skill版本列表
//	@Description	获取共享skill的版本历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SkillVersionInfo}}
//	@Router			/square/skill/share/version/list [get]
func GetSquareShareSkillVersionList(ctx *gin.Context) {
	resp, err := service.GetSquareShareSkillVersionList(ctx, ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// GetSquareCreatedSkillList
//
//	@Tags			exploration.skill
//	@Summary		获取我发布skill列表
//	@Description	获取我发布的skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.PublishedSkillInfo}}
//	@Router			/square/skill/created/list [get]
func GetSquareCreatedSkillList(ctx *gin.Context) {
	resp, err := service.GetSquareCreatedSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetSquareCreatedSkillDetail
//
//	@Tags			exploration.skill
//	@Summary		获取我发布skill详情
//	@Description	获取我发布的skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			customSkillId	query		string	true	"custom skill ID"
//	@Success		200				{object}	response.Response{data=response.PublishedSkillDetail}
//	@Router			/square/skill/created/detail [get]
func GetSquareCreatedSkillDetail(ctx *gin.Context) {
	resp, err := service.GetSquareCreatedSkillDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("customSkillId"))
	gin_util.Response(ctx, resp, err)
}

// GetSquareCreatedSkillVersionList
//
//	@Tags			exploration.skill
//	@Summary		获取我发布skill版本列表
//	@Description	获取我发布的skill版本历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			customSkillId	query		string	true	"custom skill ID"
//	@Success		200				{object}	response.Response{data=response.ListResult{list=[]response.SkillVersionInfo}}
//	@Router			/square/skill/created/version/list [get]
func GetSquareCreatedSkillVersionList(ctx *gin.Context) {
	resp, err := service.GetSquareCreatedSkillVersionList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("customSkillId"))
	gin_util.Response(ctx, resp, err)
}
