package v2

import (
	"net/http"

	v2 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v2"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerStatistic(apiV2 *gin.RouterGroup) {
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/overview", http.MethodPost, v2.GetModelStatisticOverview, "获取模型统计概览")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/chart", http.MethodPost, v2.GetModelStatisticChart, "获取模型统计趋势+排行")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list", http.MethodPost, v2.GetModelStatisticList, "获取模型调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list/user", http.MethodPost, v2.GetModelStatisticUserList, "获取模型用户使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list/app", http.MethodPost, v2.GetModelStatisticAppList, "获取模型应用使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/record/list", http.MethodPost, v2.GetModelStatisticRecord, "获取模型调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list/export", http.MethodPost, v2.ExportModelStatisticList, "导出模型调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list/user/export", http.MethodPost, v2.ExportModelStatisticUserList, "导出模型用户使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/list/app/export", http.MethodPost, v2.ExportModelStatisticAppList, "导出模型应用使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/record/export", http.MethodPost, v2.ExportModelStatisticRecord, "导出模型调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/model/select", http.MethodPost, v2.GetModelStatisticSelect, "获取模型统计下拉列表")

	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/overview", http.MethodPost, v2.GetAppStatisticOverview, "获取应用统计概览")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/chart", http.MethodPost, v2.GetAppStatisticChart, "获取应用统计趋势+排行")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list", http.MethodPost, v2.GetAppStatisticList, "获取应用调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list/user", http.MethodPost, v2.GetAppStatisticUserList, "获取应用用户使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list/model", http.MethodPost, v2.GetAppStatisticModelList, "获取应用模型使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/record/list", http.MethodPost, v2.GetAppStatisticRecord, "获取应用调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list/export", http.MethodPost, v2.ExportAppStatisticList, "导出应用调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list/user/export", http.MethodPost, v2.ExportAppStatisticUserList, "导出应用用户使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/list/model/export", http.MethodPost, v2.ExportAppStatisticModelList, "导出应用模型使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/record/export", http.MethodPost, v2.ExportAppStatisticRecord, "导出应用调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/app/select", http.MethodPost, v2.GetAppStatisticSelect, "获取应用统计下拉列表")

	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/overview", http.MethodPost, v2.GetAPIKeyStatisticOverview, "获取API Key统计概览")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/chart", http.MethodPost, v2.GetAPIKeyStatisticChart, "获取API Key统计趋势+排行")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list", http.MethodPost, v2.GetAPIKeyStatisticList, "获取API Key调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list/app", http.MethodPost, v2.GetAPIKeyStatisticAppList, "获取API Key应用调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list/model", http.MethodPost, v2.GetAPIKeyStatisticModelList, "获取API Key模型使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/record/list", http.MethodPost, v2.GetAPIKeyStatisticRecord, "获取API Key调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list/export", http.MethodPost, v2.ExportAPIKeyStatisticList, "导出API Key调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list/app/export", http.MethodPost, v2.ExportAPIKeyStatisticAppList, "导出API Key应用调用统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/list/model/export", http.MethodPost, v2.ExportAPIKeyStatisticModelList, "导出API Key模型使用列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/record/export", http.MethodPost, v2.ExportAPIKeyStatisticRecord, "导出API Key调用明细列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/routes", http.MethodGet, v2.GetApiKeyStatisticRoutes, "获取API Key统计路由列表")
	mid.Sub("app_observability.statistic").Reg(apiV2, "/statistic/api/select", http.MethodPost, v2.GetAPIKeyStatisticSelect, "获取API Key下拉列表")
}
