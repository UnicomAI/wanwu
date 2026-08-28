package v1

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetCustomSkillList
//
//	@Tags			resource.skill
//	@Summary		获取自定义skill列表
//	@Description	获取自定义skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.PublishedSkillInfo}}
//	@Router			/agent/skill/custom/list [get]
func GetCustomSkillList(ctx *gin.Context) {
	resp, err := service.GetCustomSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetCustomSkillDetail
//
//	@Tags			resource.skill
//	@Summary		获取自定义skill详情
//	@Description	获取自定义skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.PublishedSkillDetail}
//	@Router			/agent/skill/custom/detail [get]
func GetCustomSkillDetail(ctx *gin.Context) {
	resp, err := service.GetCustomSkill(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteCustomSkill
//
//	@Tags			resource.skill
//	@Summary		删除自定义skill
//	@Description	删除自定义skill
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	body		request.DeleteCustomSkillReq	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom [delete]
func DeleteCustomSkill(ctx *gin.Context) {
	var req request.DeleteCustomSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteCustomSkill(ctx, getUserID(ctx), getOrgID(ctx), req.SkillId)
	gin_util.Response(ctx, nil, err)
}

// DownloadCustomSkillVersion
//
//	@Tags			resource.skill
//	@Summary		下载自定义skill指定版本
//	@Description	下载我创建的自定义skill指定版本ZIP包
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	query		request.CustomSkillVersionDownloadReq	true	"skill版本"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom/version/download [get]
func DownloadCustomSkillVersion(ctx *gin.Context) {
	var req request.CustomSkillVersionDownloadReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	fileName := fmt.Sprintf("%s-%s.zip", req.SkillId, req.Version)
	resp, err := service.DownloadCustomSkillVersion(ctx, getUserID(ctx), getOrgID(ctx), req.SkillId, req.Version)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, resp)
}

// CreateCustomSkillConfig
//
//	@Tags			resource.skill
//	@Summary		新增自定义skill配置
//	@Description	新增自定义skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom/config [post]
func CreateCustomSkillConfig(ctx *gin.Context) {
	var req request.SkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateCustomSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// UpdateCustomSkillConfig
//
//	@Tags			resource.skill
//	@Summary		编辑自定义skill配置
//	@Description	编辑自定义skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom/config [put]
func UpdateCustomSkillConfig(ctx *gin.Context) {
	var req request.UpdateSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateCustomSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// DeleteCustomSkillConfig
//
//	@Tags			resource.skill
//	@Summary		删除自定义skill配置
//	@Description	删除自定义skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom/config [delete]
func DeleteCustomSkillConfig(ctx *gin.Context) {
	var req request.DeleteSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteCustomSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// CheckCustomSkill
//
//	@Tags			resource.skill
//	@Summary		校验自定义skill zip包
//	@Description	校验自定义skill zip包是否有效（包含SKILL.md文件）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CheckCustomSkillReq	true	"zip包URL"
//	@Success		200		{object}	response.Response{data=response.CustomSkillCheckResp}
//	@Router			/agent/skill/custom/check [post]
func CheckCustomSkill(ctx *gin.Context) {
	var req request.CheckCustomSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CheckCustomSkill(ctx, getUserID(ctx), getOrgID(ctx), req.ZipUrl)
	gin_util.Response(ctx, resp, err)
}

// GetSkillSelect
//
//	@Tags			resource.skill
//	@Summary		获取skill选择列表
//	@Description	获取skill选择列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"skill名称"
//	@Param			skillType	query		string	false	"skill类型(builtin/custom/acquired)"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.SkillInfo}}
//	@Router			/agent/skill/select [get]
func GetSkillSelect(ctx *gin.Context) {
	resp, err := service.GetSkillSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"), ctx.Query("skillType"), false)
	gin_util.Response(ctx, resp, err)
}

// GetBuiltinSkillList
//
//	@Tags			resource.skill
//	@Summary		获取内置skill列表
//	@Description	获取资源库内置skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.BuiltinSkillInfo}}
//	@Router			/agent/skill/builtin/list [get]
func GetBuiltinSkillList(ctx *gin.Context) {
	resp, err := service.GetBuiltinSkillList(ctx, ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetBuiltinSkillDetail
//
//	@Tags			resource.skill
//	@Summary		获取内置skill详情
//	@Description	获取资源库内置skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.BuiltinSkillDetail}
//	@Router			/agent/skill/builtin/detail [get]
func GetBuiltinSkillDetail(ctx *gin.Context) {
	resp, err := service.GetBuiltinSkillDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadBuiltinSkill
//
//	@Tags			resource.skill
//	@Summary		下载内置skill
//	@Description	下载内置skill ZIP包
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/builtin/skill/download [get]
func DownloadBuiltinSkill(ctx *gin.Context) {
	skillId := ctx.Query("skillId")
	fileName := fmt.Sprintf("%s.zip", skillId)
	resp, err := service.DownloadBuiltinSkill(ctx, skillId)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, resp)
}

// CreateBuiltinSkillConfig
//
//	@Tags			resource.skill
//	@Summary		新增内置skill配置
//	@Description	新增内置skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/builtin/config [post]
func CreateBuiltinSkillConfig(ctx *gin.Context) {
	var req request.SkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateBuiltinSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// UpdateBuiltinSkillConfig
//
//	@Tags			resource.skill
//	@Summary		编辑内置skill配置
//	@Description	编辑内置skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/builtin/config [put]
func UpdateBuiltinSkillConfig(ctx *gin.Context) {
	var req request.UpdateSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateBuiltinSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// DeleteBuiltinSkillConfig
//
//	@Tags			resource.skill
//	@Summary		删除内置skill配置
//	@Description	删除内置skill变量配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteSkillConfigReq	true	"skill配置"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/builtin/config [delete]
func DeleteBuiltinSkillConfig(ctx *gin.Context) {
	var req request.DeleteSkillConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteBuiltinSkillConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}
