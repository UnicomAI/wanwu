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
//	@Summary Query the knowledge base delimiter list
//	@Description Query the knowledge base delimiter list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetKnowledgeSplitterReq false "Query knowledge base separator list parameter"
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
//	@Summary creates knowledge base delimiter
//	@Description creates knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CreateKnowledgeSplitterReq true "Create knowledge base separator request parameter"
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
//	@Summary Modify the knowledge base separator
//	@Description Modify the knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateKnowledgeSplitterReq true "Modify knowledge base separator request parameters"
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
//	@Summary Remove knowledge base separator
//	@Description removes knowledge base separator
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteKnowledgeSplitterReq true "Delete knowledge base separator request parameter"
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
