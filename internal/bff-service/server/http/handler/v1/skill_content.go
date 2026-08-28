package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetSkillContentFiles 获取 Skill 内容文件列表。
//
//	@Tags			resource.skill
//	@Summary		获取Skill内容文件列表
//	@Description	已发布skill返回最新发布版本的文件树；从未发布的skill返回草稿工作区文件树
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			customSkillId	query		string	true	"Skill ID"
//	@Success		200				{object}	response.Response{data=response.SkillContentFilesResp}
//	@Router			/agent/skill/content/files [get]
func GetSkillContentFiles(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetSkillContentFilesReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSkillContentFiles(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetSkillContentFile 读取 Skill 内容文件。
//
//	@Tags			resource.skill
//	@Summary		读取Skill内容文件
//	@Description	已发布skill读取最新发布版本中的文件；从未发布的skill读取草稿工作区文件
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			customSkillId	query		string	true	"Skill ID"
//	@Param			path			query		string	true	"文件路径（相对于skill内容根目录）"
//	@Success		200				{object}	response.Response{data=response.SkillContentFileResp}
//	@Router			/agent/skill/content/file [get]
func GetSkillContentFile(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetSkillContentFileReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSkillContentFile(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}


