package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetKnowledgeTagSelect
//
//	@Tags			knowledge.tag
//	@Summary Query the knowledge base tag list
//	@Description Query the knowledge base tag list
//	@Security		JWT
//	@Accept			json
//	@Param data query request.KnowledgeTagSelectReq true "Query knowledge base request parameters"
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.KnowledgeTagListResp}
//	@Router			/knowledge/tag [get]
func GetKnowledgeTagSelect(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeTagSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeTagList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// SelectKnowledgeTagBindCount
//
//	@Tags			knowledge.tag
//	@Summary Query the number of tag binding knowledge bases
//	@Description Query the number of tag binding knowledge bases
//	@Security		JWT
//	@Accept			json
//	@Param data query request.KnowledgeTagBindCountReq true "Query tag binding number parameter request parameter"
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.KnowledgeTagListResp}
//	@Router			/knowledge/tag/bind/count [get]
func SelectKnowledgeTagBindCount(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeTagBindCountReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeTagBindCount(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// CreateKnowledgeTag
//
//	@Tags			knowledge.tag
//	@Summary Create knowledge base tags
//	@Description creates a knowledge base tag
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CreateKnowledgeTagReq true "Create knowledge base tag request parameters"
//	@Success		200		{object}	response.Response{data=response.CreateKnowledgeTagResp}
//	@Router			/knowledge/tag [post]
func CreateKnowledgeTag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateKnowledgeTagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateKnowledgeTag(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledgeTag
//
//	@Tags			knowledge.tag
//	@Summary Modify knowledge base tags
//	@Description Modify knowledge base tags
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateKnowledgeTagReq true "Modify knowledge base tag request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/tag [put]
func UpdateKnowledgeTag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKnowledgeTagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeTag(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledgeTag
//
//	@Tags			knowledge.tag
//	@Summary Remove knowledge base tag
//	@Description Delete knowledge base tag
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteKnowledgeTagReq true "Delete knowledge base tag request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/tag [delete]
func DeleteKnowledgeTag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKnowledgeTagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeTag(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BindKnowledgeTag
//
//	@Tags			knowledge.tag
//	@Summary binds knowledge base tags
//	@Description binds knowledge base tags
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.BindKnowledgeTagReq true "Bind knowledge base tag request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/tag/bind [post]
func BindKnowledgeTag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.BindKnowledgeTagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.BindKnowledgeTag(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
