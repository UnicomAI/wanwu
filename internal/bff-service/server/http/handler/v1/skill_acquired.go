package v1

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAcquiredSkillList
//
//	@Tags			resource.skill
//	@Summary		获取我添加的skill列表
//	@Description	获取资源库中我添加的skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AcquiredSkillInfo}}
//	@Router			/agent/skill/acquired/list [get]
func GetAcquiredSkillList(ctx *gin.Context) {
	resp, err := service.GetAcquiredSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// DeleteAcquiredSkill
//
//	@Tags			resource.skill
//	@Summary		删除我添加的skill
//	@Description	删除资源库中我添加的skill
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	body		request.DeleteAcquiredSkillReq	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/acquired [delete]
func DeleteAcquiredSkill(ctx *gin.Context) {
	var req request.DeleteAcquiredSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteAcquiredSkill(ctx, getUserID(ctx), getOrgID(ctx), req.SkillId)
	gin_util.Response(ctx, nil, err)
}

// GetAcquiredSkillDetail
//
//	@Tags			resource.skill
//	@Summary		获取我添加的skill详情
//	@Description	获取资源库中我添加的skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.AcquiredSkillDetail}
//	@Router			/agent/skill/acquired/detail [get]
func GetAcquiredSkillDetail(ctx *gin.Context) {
	resp, err := service.GetAcquiredSkill(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadAcquiredSkill
//
//	@Tags			resource.skill
//	@Summary		下载我添加的skill
//	@Description	下载资源库中我添加的最新版本skill
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/acquired/download [get]
func DownloadAcquiredSkill(ctx *gin.Context) {
	skillId := ctx.Query("skillId")
	fileName := fmt.Sprintf("%s.zip", skillId)
	resp, err := service.DownloadAcquiredSkill(ctx, getUserID(ctx), getOrgID(ctx), skillId)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, resp)
}

// GetAcquiredSkillVersionList
//
//	@Tags			resource.skill
//	@Summary		获取我添加的skill历史版本列表
//	@Description	获取资源库中我添加的skill历史版本列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SkillVersionInfo}}
//	@Router			/agent/skill/acquired/version/list [get]
func GetAcquiredSkillVersionList(ctx *gin.Context) {
	resp, err := service.GetAcquiredSkillVersionList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// CreateAcquiredSkillConfig
//
//	@Tags			resource.skill
//	@Summary		新增我添加的skill配置
//	@Description	新增资源库中我添加的skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/acquired/config [post]
func CreateAcquiredSkillConfig(ctx *gin.Context) {
	var req request.SkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateAcquiredSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// UpdateAcquiredSkillConfig
//
//	@Tags			resource.skill
//	@Summary		编辑我添加的skill配置
//	@Description	编辑资源库中我添加的skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/acquired/config [put]
func UpdateAcquiredSkillConfig(ctx *gin.Context) {
	var req request.UpdateSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateAcquiredSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// DeleteAcquiredSkillConfig
//
//	@Tags			resource.skill
//	@Summary		删除我添加的skill配置
//	@Description	删除资源库中我添加的skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/acquired/config [delete]
func DeleteAcquiredSkillConfig(ctx *gin.Context) {
	var req request.DeleteSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteAcquiredSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}
