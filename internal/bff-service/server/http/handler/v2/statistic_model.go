package v2

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// ========== JSON 查询类 ==========

// GetModelStatisticOverview
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型统计概览
//	@Description	10张指标卡（含日均Tokens）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2Req	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.ModelStatisticV2Overview}
//	@Router			/statistic/model/overview [post]
func GetModelStatisticOverview(ctx *gin.Context) {
	var req request.ModelStatisticV2Req
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2Overview(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticChart
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型统计趋势+排行
//	@Description	tokens用量/调用趋势 + 模型/用户/组织三维度排行
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2ChartReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.ModelStatisticV2Chart}
//	@Router			/statistic/model/chart [post]
func GetModelStatisticChart(ctx *gin.Context) {
	var req request.ModelStatisticV2ChartReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2Chart(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型调用统计列表
//	@Description	按模型聚合的调用统计（分页）。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2ListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ModelStatisticV2ListItem}}
//	@Router			/statistic/model/list [post]
func GetModelStatisticList(ctx *gin.Context) {
	var req request.ModelStatisticV2ListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticUserList
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型用户使用列表
//	@Description	指定模型下的用户使用统计（分页）。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2UserListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ModelStatisticV2UserListItem}}
//	@Router			/statistic/model/list/user [post]
func GetModelStatisticUserList(ctx *gin.Context) {
	var req request.ModelStatisticV2UserListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2UserList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticAppList
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型应用使用列表
//	@Description	指定模型下的应用使用统计（分页）。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2AppListReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ModelStatisticV2AppListItem}}
//	@Router			/statistic/model/list/app [post]
func GetModelStatisticAppList(ctx *gin.Context) {
	var req request.ModelStatisticV2AppListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2AppList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型调用明细列表
//	@Description	模型调用明细记录（分页，不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2RecordReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ModelStatisticV2RecordItem}}
//	@Router			/statistic/model/record/list [post]
func GetModelStatisticRecord(ctx *gin.Context) {
	var req request.ModelStatisticV2RecordReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}

// ========== 导出类 ==========

// ExportModelStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		导出模型调用统计列表
//	@Description	导出模型调用统计列表。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.ModelStatisticV2ExportListReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/model/list/export [post]
func ExportModelStatisticList(ctx *gin.Context) {
	var req request.ModelStatisticV2ExportListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportModelStatisticV2List(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("模型统计_调用统计_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportModelStatisticUserList
//
//	@Tags			app_observability.statistic
//	@Summary		导出模型用户使用列表
//	@Description	导出模型用户使用列表。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.ModelStatisticV2UserExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/model/list/user/export [post]
func ExportModelStatisticUserList(ctx *gin.Context) {
	var req request.ModelStatisticV2UserExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportModelStatisticV2UserList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("模型统计_用户使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportModelStatisticAppList
//
//	@Tags			app_observability.statistic
//	@Summary		导出模型应用使用列表
//	@Description	导出模型应用使用列表。sortField 可选数值指标：totalTokens、promptTokens、completionTokens、callCount、callFailure、failureRate、avgFirstTokenLatency、avgCosts
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.ModelStatisticV2AppExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/model/list/app/export [post]
func ExportModelStatisticAppList(ctx *gin.Context) {
	var req request.ModelStatisticV2AppExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportModelStatisticV2AppList(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("模型统计_应用使用_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// ExportModelStatisticRecord
//
//	@Tags			app_observability.statistic
//	@Summary		导出模型调用明细
//	@Description	导出模型调用明细（不支持用户排序；后端固定按调用时间倒序）
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	body		request.ModelStatisticV2RecordExportReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/statistic/model/record/export [post]
func ExportModelStatisticRecord(ctx *gin.Context) {
	var req request.ModelStatisticV2RecordExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	file, err := service.ExportModelStatisticV2Record(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	writeExcelExport(ctx, file, fmt.Sprintf("模型统计_调用明细_%v-%v.xlsx", req.StartDate, req.EndDate))
}

// GetModelStatisticSelect
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型统计下拉列表
//	@Description	viewScope=published 查主表（我发布的）；used 查聚合表（我使用的）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelStatisticV2SelectReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ModelInfo}}
//	@Router			/statistic/model/select [post]
func GetModelStatisticSelect(ctx *gin.Context) {
	var req request.ModelStatisticV2SelectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelStatisticV2Select(ctx, &req, getUserID(ctx), getOrgID(ctx), isAdmin(ctx), isSystem(ctx))
	gin_util.Response(ctx, resp, err)
}
