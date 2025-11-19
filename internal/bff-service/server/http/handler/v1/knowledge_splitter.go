package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetKnowledgeSplitterSelect
//
//	@Tags			knowledge.splitter
//	@Summary		查询知识库分隔符列表 [EN] @Summary Query the knowledge base delimiter list
//	@Description	查询知识库分隔符列表 [EN] @Description Query the knowledge base delimiter list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetKnowledgeSplitterReq	false	"查询知识库分隔符列表参数" [EN] @Param data query request.GetKnowledgeSplitterReq false "Query knowledge base separator list parameter"
//	@Success		200		{object}	response.Response{data=response.KnowledgeSplitterListResp}
//	@Router			/knowledge/splitter [get]
func GetKnowledgeSplitterSelect(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetKnowledgeSplitterReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeSplitterList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// CreateKnowledgeSplitter
//
//	@Tags			knowledge.splitter
//	@Summary		创建知识库分隔符 [EN] @Summary creates knowledge base delimiter
//	@Description	创建知识库分隔符 [EN] @Description creates knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateKnowledgeSplitterReq	true	"创建知识库分隔符请求参数" [EN] @Param data body request.CreateKnowledgeSplitterReq true "Create knowledge base separator request parameter"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/splitter [post]
func CreateKnowledgeSplitter(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateKnowledgeSplitterReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateKnowledgeSplitter(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateKnowledgeSplitter
//
//	@Tags			knowledge.splitter
//	@Summary		修改知识库分隔符 [EN] @Summary Modify the knowledge base separator
//	@Description	修改知识库分隔符 [EN] @Description Modify the knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateKnowledgeSplitterReq	true	"修改知识库分隔符请求参数" [EN] @Param data body request.UpdateKnowledgeSplitterReq true "Modify knowledge base separator request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/splitter [put]
func UpdateKnowledgeSplitter(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKnowledgeSplitterReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeSplitter(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledgeSplitter
//
//	@Tags			knowledge.splitter
//	@Summary		删除知识库分隔符 [EN] @Summary Remove knowledge base separator
//	@Description	删除知识库分隔符 [EN] @Description removes knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteKnowledgeSplitterReq	true	"删除知识库分隔符请求参数" [EN] @Param data body request.DeleteKnowledgeSplitterReq true "Delete knowledge base separator request parameter"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/splitter [delete]
func DeleteKnowledgeSplitter(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKnowledgeSplitterReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeSplitter(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
