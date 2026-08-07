package v2

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
	"github.com/gin-gonic/gin"
)

// ========== JSON 查询类 ==========

// GetAPIKeyStatisticOverview
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key统计概览
//	@Description	10张指标卡（调用总次数、失败次数、日均调用/失败、日均流式/非流式、平均耗时、流式/非流式调用次数）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2Req	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.APIKeyStatisticV2Overview}
//	@Router			/statistic/api/overview [post]
func GetAPIKeyStatisticOverview(ctx *gin.Context) {
	var req request.APIKeyStatisticV2Req
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2Overview(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAPIKeyStatisticChart
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key统计趋势+排行
//	@Description	API Key调用趋势 + API Key调用次数排行
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2ChartReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.APIKeyStatisticV2Chart}
//	@Router			/statistic/api/chart [post]
func GetAPIKeyStatisticChart(ctx *gin.Context) {
	var req request.APIKeyStatisticV2ChartReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2Chart(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAPIKeyStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key调用统计列表
//	@Description	按API Key+路径聚合的调用统计（分页）。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2ListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.APIKeyStatisticV2ListItem}}
//	@Router			/statistic/api/list [post]
func GetAPIKeyStatisticList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2ListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAPIKeyStatisticAppList
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key应用调用统计列表
//	@Description	指定API Key下的应用调用统计（分页）。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2AppListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.APIKeyStatisticV2AppListItem}}
//	@Router			/statistic/api/list/app [post]
func GetAPIKeyStatisticAppList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2AppListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2AppList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAPIKeyStatisticModelList
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key模型使用列表
//	@Description	指定API Key下的模型使用统计（分页）。钻取需传当前主表行的 apiKeyId/methodPath。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2ModelListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.APIKeyStatisticV2ModelListItem}}
//	@Router			/statistic/api/list/model [post]
func GetAPIKeyStatisticModelList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2ModelListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2ModelList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAPIKeyStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key调用明细列表
//	@Description	API Key调用明细记录（分页，不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2RecordReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.APIKeyStatisticV2RecordItem}}
//	@Router			/statistic/api/record/list [post]
func GetAPIKeyStatisticRecord(ctx *gin.Context) {
	var req request.APIKeyStatisticV2RecordReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAPIKeyStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// ========== 导出类 ==========

// ExportAPIKeyStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		导出API Key调用统计列表
//	@Description	导出API Key调用统计列表。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.APIKeyStatisticV2ExportListReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/api/list/export [post]
func ExportAPIKeyStatisticList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2ExportListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAPIKeyStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("APIKey统计_调用统计_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAPIKeyStatisticAppList
//
//	@Tags			app_observability.statistic
//	@Summary		导出API Key应用调用统计列表
//	@Description	导出API Key应用调用统计列表。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.APIKeyStatisticV2AppExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/api/list/app/export [post]
func ExportAPIKeyStatisticAppList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2AppExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAPIKeyStatisticV2AppList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("APIKey统计_应用使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAPIKeyStatisticModelList
//
//	@Tags			app_observability.statistic
//	@Summary		导出API Key模型使用列表
//	@Description	导出API Key模型使用列表。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.APIKeyStatisticV2ModelExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/api/list/model/export [post]
func ExportAPIKeyStatisticModelList(ctx *gin.Context) {
	var req request.APIKeyStatisticV2ModelExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAPIKeyStatisticV2ModelList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("APIKey统计_模型使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAPIKeyStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		导出API Key调用明细
//	@Description	导出API Key调用明细（不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.APIKeyStatisticV2RecordExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/api/record/export [post]
func ExportAPIKeyStatisticRecord(ctx *gin.Context) {
	var req request.APIKeyStatisticV2RecordExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAPIKeyStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("APIKey统计_调用明细_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// GetApiKeyStatisticRoutes
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key统计路由列表
//	@Description	根据 openApiType 获取对应的路由信息，不传则返回所有 OpenAPI 路由
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			openApiType	query		string	false	"OpenAPI类型: agent, rag, workflow, chatflow, knowledge"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.ApiKeyStatisticRouteItem}}
//	@Router			/statistic/api/routes [get]
func GetApiKeyStatisticRoutes(ctx *gin.Context) {
	routes := route.GetApiKeyStatisticRoutes(ctx.Query("openApiType"))
	gin_util.Response(ctx, routes, nil)
}

// GetAPIKeyStatisticSelect
//
//	@Tags			app_observability.statistic
//	@Summary		获取API Key下拉列表
//	@Description	组织→用户→API Key名称级联
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.APIKeyStatisticV2SelectReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.APIKeyDetailResponse}}
//	@Router			/statistic/api/select [post]
func GetAPIKeyStatisticSelect(ctx *gin.Context) {
	var req request.APIKeyStatisticV2SelectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetStatisticAPIKeySelect(ctx, req.StatisticFilter, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}
