package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetKnowledgeKeywordsList
//
//	@Tags			knowledge.keywords
//	@Summary Query the knowledge base keyword list
//	@Description Query the knowledge base keyword list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.ListKeywordsReq true "Keyword list query request parameters"
//	@Success		200		{object}	response.GetKnowledgeKeywordListResp
//	@Router			/knowledge/keywords [get]
func GetKnowledgeKeywordsList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ListKeywordsReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeKeywordsList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// CreateKnowledgeKeywords
//
//	@Tags			knowledge.keywords
//	@Summary Added new knowledge base keywords
//	@Description Add new knowledge base keywords
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CreateKeywordsReq true "Create keyword request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/keywords [post]
func CreateKnowledgeKeywords(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateKeywordsReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateKnowledgeKeywords(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetKnowledgeKeywordsDetail
//
//	@Tags			knowledge.keywords
//	@Summary Query knowledge base keyword details
//	@Description Query knowledge base keyword details
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetKeywordsDetailReq true "Keyword list query request parameters"
//	@Success		200		{object}	response.KeywordsInfo
//	@Router			/knowledge/keywords/detail [get]
func GetKnowledgeKeywordsDetail(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetKeywordsDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeKeywordsDetail(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledgeKeywords
//
//	@Tags			knowledge.keywords
//	@Summary Edit knowledge base keywords
//	@Description Edit knowledge base keywords
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateKeywordsReq true "Modify keyword request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/keywords [put]
func UpdateKnowledgeKeywords(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKeywordsReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeKeywords(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDocCategoryKeywords
//
//	@Tags			knowledge.keywords
//	@Summary Delete knowledge base keywords
//	@Description Delete knowledge base keywords
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteKeywordsReq true "Delete knowledge base keyword request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/keywords [delete]
func DeleteDocCategoryKeywords(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKeywordsReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteDocCategoryKeywords(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
