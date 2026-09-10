package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetParseTemplateList
//
//	@Tags			knowledge.parseTemplate
//	@Summary		查询解析模板列表
//	@Description	查询解析模板列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ParseTemplateListReq	true	"解析模板列表查询请求参数"
//	@Success		200		{object}	response.ParseTemplateListResp
//	@Router			/knowledge/parseTemplate [get]
func GetParseTemplateList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ParseTemplateListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetParseTemplateList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// CreateParseTemplate
//
//	@Tags			knowledge.parseTemplate
//	@Summary		新建解析模板
//	@Description	新建解析模板
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateParseTemplateReq	true	"新建解析模板请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/parseTemplate [post]
func CreateParseTemplate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateParseTemplateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateParseTemplate(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateParseTemplate
//
//	@Tags			knowledge.parseTemplate
//	@Summary		编辑解析模板
//	@Description	编辑解析模板
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateParseTemplateReq	true	"编辑解析模板请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/parseTemplate [put]
func UpdateParseTemplate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateParseTemplateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateParseTemplate(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteParseTemplate
//
//	@Tags			knowledge.parseTemplate
//	@Summary		删除解析模板
//	@Description	删除解析模板
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteParseTemplateReq	true	"删除解析模板请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/parseTemplate [delete]
func DeleteParseTemplate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteParseTemplateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteParseTemplate(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
