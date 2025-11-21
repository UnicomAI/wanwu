package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Get community reports
//	@Description Get community reports
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.KnowledgeReportSelectReq true "Get community report request parameters"
//	@Success		200		{object}	response.Response{data=response.KnowledgeReportPageResult}
//	@Router			/knowledge/report/list [get]
func GetKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GenerateKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Generate community report
//	@Description Generate community report
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.KnowledgeReportGenerateReq true "Generate community report request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/report/generate [post]
func GenerateKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportGenerateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.GenerateKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Remove community report
//	@Description Delete community report
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.KnowledgeReportDeleteReq true "Delete community report request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/report/delete [delete]
func DeleteKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportDeleteReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Editor Community Report
//	@Description Edit community report
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.KnowledgeReportUpdateReq true "Edit community report request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/report/update [post]
func UpdateKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// AddKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Add a single community report
//	@Description Add a single community report
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.KnowledgeReportAddReq true "Single new community report request parameter"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/report/add [post]
func AddKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportAddReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AddKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BatchAddKnowledgeReport
//
//	@Tags			knowledge.report
//	@Summary Add community reports in batches
//	@Description Add community reports in batches
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.KnowledgeReportBatchAddReq true "Batch add community report request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/report/batch/add [post]
func BatchAddKnowledgeReport(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeReportBatchAddReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.BatchAddKnowledgeReport(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
