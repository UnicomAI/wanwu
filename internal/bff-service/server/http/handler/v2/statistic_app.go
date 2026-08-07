package v2

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// ========== JSON 查询类 ==========

// GetAppStatisticOverview
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用统计概览
//	@Description	10张指标卡（调用次数、Tokens、流式/非流式维度）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2Req	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.AppStatisticV2Overview}
//	@Router			/statistic/app/overview [post]
func GetAppStatisticOverview(ctx *gin.Context) {
	var req request.AppStatisticV2Req
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2Overview(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAppStatisticChart
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用统计趋势+排行
//	@Description	调用趋势 + 智能体/工作流/对话流/知识问答排行
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2ChartReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.AppStatisticV2Chart}
//	@Router			/statistic/app/chart [post]
func GetAppStatisticChart(ctx *gin.Context) {
	var req request.AppStatisticV2ChartReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2Chart(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAppStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用调用统计列表
//	@Description	按应用聚合的调用统计（分页）。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2ListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.AppStatisticV2ListItem}}
//	@Router			/statistic/app/list [post]
func GetAppStatisticList(ctx *gin.Context) {
	var req request.AppStatisticV2ListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAppStatisticUserList
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用用户使用列表
//	@Description	指定应用下的用户使用统计（分页）。钻取需传当前主表行的 source/appId/moduleCreatorUserId/moduleCreatorOrgId。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2UserListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.AppStatisticV2UserListItem}}
//	@Router			/statistic/app/list/user [post]
func GetAppStatisticUserList(ctx *gin.Context) {
	var req request.AppStatisticV2UserListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2UserList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAppStatisticModelList
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用模型使用列表
//	@Description	指定应用下的模型使用统计（分页）。钻取需传当前主表行的 source/appId/moduleCreatorUserId/moduleCreatorOrgId。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2ModelListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.AppStatisticV2ModelListItem}}
//	@Router			/statistic/app/list/model [post]
func GetAppStatisticModelList(ctx *gin.Context) {
	var req request.AppStatisticV2ModelListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2ModelList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetAppStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用调用明细列表
//	@Description	应用调用明细记录（分页，不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2RecordReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.AppStatisticV2RecordItem}}
//	@Router			/statistic/app/record/list [post]
func GetAppStatisticRecord(ctx *gin.Context) {
	var req request.AppStatisticV2RecordReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// ========== 导出类 ==========

// ExportAppStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		导出应用调用统计列表
//	@Description	导出应用调用统计列表。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.AppStatisticV2ExportListReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/app/list/export [post]
func ExportAppStatisticList(ctx *gin.Context) {
	var req request.AppStatisticV2ExportListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAppStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("应用统计_调用统计_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAppStatisticUserList
//
//	@Tags			app_observability.statistic
//	@Summary		导出应用用户使用列表
//	@Description	导出应用用户使用列表。sortField 可选数值指标：callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.AppStatisticV2UserExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/app/list/user/export [post]
func ExportAppStatisticUserList(ctx *gin.Context) {
	var req request.AppStatisticV2UserExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAppStatisticV2UserList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("应用统计_用户使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAppStatisticModelList
//
//	@Tags			app_observability.statistic
//	@Summary		导出应用模型使用列表
//	@Description	导出应用模型使用列表。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、streamCount、nonStreamCount、streamFailure、nonStreamFailure、failureRate、avgFirstTokenLatency、avgCosts、firstTokenLatency、costs
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.AppStatisticV2ModelExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/app/list/model/export [post]
func ExportAppStatisticModelList(ctx *gin.Context) {
	var req request.AppStatisticV2ModelExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAppStatisticV2ModelList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("应用统计_模型使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportAppStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		导出应用调用明细
//	@Description	导出应用调用明细（不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.AppStatisticV2RecordExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/app/record/export [post]
func ExportAppStatisticRecord(ctx *gin.Context) {
	var req request.AppStatisticV2RecordExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportAppStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("应用统计_调用明细_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// GetAppStatisticSelect
//
//	@Tags			app_observability.statistic
//	@Summary		获取应用统计下拉列表
//	@Description	viewScope=published 查主表（我发布的）；used 查聚合表（我使用的）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppStatisticV2SelectReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MyAppItem}}
//	@Router			/statistic/app/select [post]
func GetAppStatisticSelect(ctx *gin.Context) {
	var req request.AppStatisticV2SelectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAppStatisticV2Select(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}
