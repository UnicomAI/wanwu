package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// ============ Overview ============

func GetAPIKeyStatisticV2Overview(ctx *gin.Context, req *request.APIKeyStatisticV2Req, userId, orgId string, isAdmin, isSystem bool) (*response.APIKeyStatisticV2Overview, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAPIKeyStatisticV2Overview(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2ReadReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
	})
	if err != nil {
		return nil, err
	}
	return convertAPIKeyV2Overview(resp), nil
}

// ============ Chart (Trend + Rank) ============

func GetAPIKeyStatisticV2Chart(ctx *gin.Context, req *request.APIKeyStatisticV2ChartReq, userId, orgId string, isAdmin, isSystem bool) (*response.APIKeyStatisticV2Chart, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	limit := int32(req.Limit)
	if limit <= 0 {
		limit = 5
	}
	resp, err := app.GetAPIKeyStatisticV2Chart(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2ChartReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}
	rank, err := buildAPIKeyStatisticV2RankResponse(ctx, scope, resp.GetRank())
	if err != nil {
		return nil, err
	}
	return &response.APIKeyStatisticV2Chart{
		Trend: response.APIKeyStatisticV2Trend{
			ApiKeyCalls: convertStatisticChart(ctx, resp.GetTrend().GetApiKeyCalls()),
			CallResult:  convertStatisticChart(ctx, resp.GetTrend().GetCallResult()),
		},
		Rank: *rank,
	}, nil
}

func buildAPIKeyStatisticV2RankResponse(ctx *gin.Context, scope *statisticScope, rank *app_service.APIKeyStatisticV2Rank) (*response.APIKeyStatisticV2Rank, error) {
	if rank == nil {
		return &response.APIKeyStatisticV2Rank{}, nil
	}
	infoMap := getAPIKeyInfoMap(ctx, scope)
	var orgIds, userIds []string
	for _, it := range rank.GetByApi() {
		orgIds = append(orgIds, it.GetOrgId())
		userIds = append(userIds, it.GetUserId())
	}
	orgNameMap, _, err := buildStatisticOrgMaps(ctx, orgIds, false)
	if err != nil {
		return nil, err
	}
	userNameMap, userAvatarMap, err := buildStatisticUserMaps(ctx, userIds, true)
	if err != nil {
		return nil, err
	}
	byApi := make([]response.APIKeyStatisticV2RankItem, 0, len(rank.GetByApi()))
	for _, it := range rank.GetByApi() {
		info := getAPIKeyDisplayInfo(ctx, infoMap, it.GetApiKeyId())
		byApi = append(byApi, response.APIKeyStatisticV2RankItem{
			ApiName:       info.name,
			CallCount:     it.GetCallCount(),
			UserBriefInfo: buildUserBriefInfo(ctx, it.GetUserId(), it.GetOrgId(), userNameMap, orgNameMap, userAvatarMap),
		})
	}
	return &response.APIKeyStatisticV2Rank{ByApi: byApi}, nil
}

// ============ List (主表) ============

func GetAPIKeyStatisticV2List(ctx *gin.Context, req *request.APIKeyStatisticV2ListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAPIKeyStatisticV2List(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2ListReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
		PageNo:      int32(req.PageNo),
		PageSize:    int32(req.PageSize),
		SortField:   resolveSortExpr(sortFieldAggregateCallsOnly, req.SortField),
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	items, err := buildAPIKeyStatisticV2ListItems(ctx, scope, resp.GetItems())
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ AppList ============

func GetAPIKeyStatisticV2AppList(ctx *gin.Context, req *request.APIKeyStatisticV2AppListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAPIKeyStatisticV2AppList(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2AppListReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
		ApiKeyId:    req.ApiKeyId,
		MethodPath:  req.MethodPath,
		PageNo:      int32(req.PageNo),
		PageSize:    int32(req.PageSize),
		SortField:   resolveSortExpr(sortFieldAggregateFromRecord, req.SortField), // AppList 从明细表现场聚合，不能用 SUM(call_count)
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	items, err := buildAPIKeyStatisticV2AppListItems(ctx, scope, resp.GetItems())
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ ModelList ============

func GetAPIKeyStatisticV2ModelList(ctx *gin.Context, req *request.APIKeyStatisticV2ModelListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAPIKeyStatisticV2ModelList(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2ModelListReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
		ApiKeyId:    req.ApiKeyId,
		MethodPath:  req.MethodPath,
		PageNo:      int32(req.PageNo),
		PageSize:    int32(req.PageSize),
		SortField:   resolveSortExpr(sortFieldAggregateFull, req.SortField),
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	items, err := buildAPIKeyStatisticV2ModelListItems(ctx, scope, resp.GetItems())
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ Record 列表 ============

func GetAPIKeyStatisticV2Record(ctx *gin.Context, req *request.APIKeyStatisticV2RecordReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAPIKeyStatisticV2Record(ctx.Request.Context(), &app_service.GetAPIKeyStatisticV2RecordReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ApiKeyIds:   normalizeAPIKeyIds(req.ApiKeyIds),
		MethodPaths: req.MethodPaths,
		PageNo:      int32(req.PageNo),
		PageSize:    int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	items, err := buildAPIKeyStatisticV2RecordItems(ctx, scope, resp.GetItems())
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ 导出 ============

func ExportAPIKeyStatisticV2List(ctx *gin.Context, req *request.APIKeyStatisticV2ExportListReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.APIKeyStatisticV2ListReq{
		APIKeyStatisticV2Req: req.APIKeyStatisticV2Req,
		StatisticSortReq:     req.StatisticSortReq,
		PageNo:               -1,
		PageSize:             -1,
	}
	resp, err := GetAPIKeyStatisticV2List(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.APIKeyStatisticV2ListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "API Key统计V2_调用统计"
	// 列序与 V1 列表页 / brief 字段一致：请求路径在组织/用户之前
	title := []any{"API Key名称", "API Key", "请求路径", "组织", "用户", "调用次数(次)", "调用失败次数(次)", "失败率(%)", "流式平均首Token(ms)", "非流式平均耗时(ms)", "调用次数(流式)(次)", "调用次数(非流式)(次)"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.ApiName, it.ApiKey, it.MethodPath, it.OrgName, it.UserName,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgFirstTokenLatency, it.AvgCosts,
			it.StreamCount, it.NonStreamCount,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportAPIKeyStatisticV2AppList(ctx *gin.Context, req *request.APIKeyStatisticV2AppExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.APIKeyStatisticV2AppListReq{
		APIKeyStatisticV2Req: req.APIKeyStatisticV2Req,
		ApiKeyId:             req.ApiKeyId,
		MethodPath:           req.MethodPath,
		StatisticSortReq:     req.StatisticSortReq,
		PageNo:               -1,
		PageSize:             -1,
	}
	resp, err := GetAPIKeyStatisticV2AppList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.APIKeyStatisticV2AppListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "API Key统计V2_应用使用"
	title := []any{"API Key名称", "API Key", "请求路径", "组织", "用户", "来源", "板块", "应用名称", "应用类型", "作者", "作者组织", "调用次数", "失败次数", "失败率(%)", "流式平均首Token(ms)", "非流式平均耗时(ms)", "调用次数(流式)(次)", "调用次数(非流式)(次)"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.ApiName, it.ApiKey, it.MethodPath, it.OrgName, it.UserName,
			it.SourceName, it.ModuleName, it.AppName, it.AppType, it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgFirstTokenLatency, it.AvgCosts,
			it.StreamCount, it.NonStreamCount,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportAPIKeyStatisticV2ModelList(ctx *gin.Context, req *request.APIKeyStatisticV2ModelExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.APIKeyStatisticV2ModelListReq{
		APIKeyStatisticV2Req: req.APIKeyStatisticV2Req,
		ApiKeyId:             req.ApiKeyId,
		MethodPath:           req.MethodPath,
		StatisticSortReq:     req.StatisticSortReq,
		PageNo:               -1,
		PageSize:             -1,
	}
	resp, err := GetAPIKeyStatisticV2ModelList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.APIKeyStatisticV2ModelListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "API Key统计V2_模型使用"
	title := []any{"API Key名称", "API Key", "请求路径", "组织", "用户", "模型ID", "模型名称", "供应商", "模型类型", "发布者", "发布者组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均首Token时延(流式)(ms)", "平均耗时(非流式)(ms)", "流式次数", "非流式次数"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.ApiName, it.ApiKey, it.MethodPath, it.OrgName, it.UserName,
			it.ModelId, it.Model, it.Provider, it.ModelType,
			it.ModelCreatorUserName, it.ModelCreatorOrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgFirstTokenLatency, it.AvgCosts,
			it.StreamCount, it.NonStreamCount,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportAPIKeyStatisticV2Record(ctx *gin.Context, req *request.APIKeyStatisticV2RecordExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.APIKeyStatisticV2RecordReq{
		APIKeyStatisticV2Req: req.APIKeyStatisticV2Req,
		PageNo:               1,
		PageSize:             10000, // 导出最大行数
	}
	resp, err := GetAPIKeyStatisticV2Record(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.APIKeyStatisticV2RecordItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "API Key统计V2_调用明细"
	title := []any{"API Key名称", "API Key", "请求路径", "组织", "用户", "调用时间", "调用结果", "来源", "板块", "应用名称", "应用类型", "作者", "作者组织", "流式耗时(ms)", "非流式耗时(ms)", "失败原因", "请求体", "响应体"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.ApiName, it.ApiKey, it.MethodPath, it.OrgName, it.UserName,
			it.CalledAt, statisticExportSuccessLabel(it.IsSuccess),
			it.SourceName, it.ModuleName, it.AppName, it.AppType, it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.FirstTokenLatency, it.Costs,
			it.FailureReason, it.RequestBody, it.ResponseBody,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

// ============ helpers ============

func buildAPIKeyStatisticV2ListItems(ctx *gin.Context, scope *statisticScope, protoItems []*app_service.APIKeyStatisticV2ListItem) ([]response.APIKeyStatisticV2ListItem, error) {
	infoMap := getAPIKeyInfoMap(ctx, scope)
	var orgIds, userIds []string
	for _, item := range protoItems {
		orgIds = append(orgIds, item.GetOrgId())
		userIds = append(userIds, item.GetUserId())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, orgIds, userIds)
	if err != nil {
		return nil, err
	}
	items := make([]response.APIKeyStatisticV2ListItem, 0, len(protoItems))
	for _, item := range protoItems {
		info := getAPIKeyDisplayInfo(ctx, infoMap, item.GetApiKeyId())
		items = append(items, response.APIKeyStatisticV2ListItem{
			ApiKeyBriefInfo: buildStatisticV2ApiKeyInfo(ctx, info.name, info.key, item.GetApiKeyId(), item.GetMethodPath(),
				item.GetOrgId(), item.GetUserId(), orgNameMap, userNameMap, nil),
			APIKeyStatisticV2Metrics: convertAPIKeyV2Metrics(item.GetMetrics()),
		})
	}
	return items, nil
}

func buildAPIKeyStatisticV2AppListItems(ctx *gin.Context, scope *statisticScope, protoItems []*app_service.APIKeyStatisticV2AppListItem) ([]response.APIKeyStatisticV2AppListItem, error) {
	infoMap := getAPIKeyInfoMap(ctx, scope)
	var orgIds, userIds, creatorOrgIds, creatorUserIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, item := range protoItems {
		orgIds = append(orgIds, item.GetOrgId())
		userIds = append(userIds, item.GetUserId())
		if item.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, item.GetModuleCreatorOrgId())
		}
		if item.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, item.GetModuleCreatorUserId())
		}
		addStatisticAppIdByType(appIdsByType, item.GetAppId(), item.GetAppType(), item.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, append(orgIds, creatorOrgIds...), append(userIds, creatorUserIds...))
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)
	items := make([]response.APIKeyStatisticV2AppListItem, 0, len(protoItems))
	for _, item := range protoItems {
		info := getAPIKeyDisplayInfo(ctx, infoMap, item.GetApiKeyId())
		items = append(items, response.APIKeyStatisticV2AppListItem{
			ApiKeyBriefInfo: buildStatisticV2ApiKeyInfo(ctx, info.name, info.key, item.GetApiKeyId(), item.GetMethodPath(),
				item.GetOrgId(), item.GetUserId(), orgNameMap, userNameMap, nil),
			ModuleBriefInfo: buildStatisticV2AppInfo(ctx, item.GetSource(), item.GetModule(), item.GetAppId(), item.GetAppType(),
				item.GetModuleCreatorUserId(), item.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			APIKeyStatisticV2Metrics: convertAPIKeyV2Metrics(item.GetMetrics()),
		})
	}
	return items, nil
}

func buildAPIKeyStatisticV2ModelListItems(ctx *gin.Context, scope *statisticScope, protoItems []*app_service.APIKeyStatisticV2ModelListItem) ([]response.APIKeyStatisticV2ModelListItem, error) {
	infoMap := getAPIKeyInfoMap(ctx, scope)
	var orgIds, userIds, modelIds []string
	for _, item := range protoItems {
		orgIds = append(orgIds, item.GetOrgId())
		userIds = append(userIds, item.GetUserId())
		modelIds = append(modelIds, item.GetModelId())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, orgIds, userIds)
	if err != nil {
		return nil, err
	}
	modelMap := getModelInfoMap(ctx, modelIds)
	var modelCreatorUserIds, modelCreatorOrgIds []string
	for _, m := range modelMap {
		if m.creatorUserId != "" {
			modelCreatorUserIds = append(modelCreatorUserIds, m.creatorUserId)
		}
		if m.creatorOrgId != "" {
			modelCreatorOrgIds = append(modelCreatorOrgIds, m.creatorOrgId)
		}
	}
	modelCreatorOrgNameMap, modelCreatorUserNameMap, err := buildStatisticOrgUserNameMaps(ctx, modelCreatorOrgIds, modelCreatorUserIds)
	if err != nil {
		return nil, err
	}
	items := make([]response.APIKeyStatisticV2ModelListItem, 0, len(protoItems))
	for _, item := range protoItems {
		info := getAPIKeyDisplayInfo(ctx, infoMap, item.GetApiKeyId())
		modelInfo := modelMap[item.GetModelId()]
		items = append(items, response.APIKeyStatisticV2ModelListItem{
			ApiKeyBriefInfo: buildStatisticV2ApiKeyInfo(ctx, info.name, info.key, item.GetApiKeyId(), item.GetMethodPath(),
				item.GetOrgId(), item.GetUserId(), orgNameMap, userNameMap, nil),
			ModelBriefInfo: buildModelBriefInfo(ctx, item.GetModelId(), item.GetModel(), item.GetProvider(), item.GetModelType(),
				modelInfo.creatorUserId, modelInfo.creatorOrgId, modelInfo, modelCreatorUserNameMap, modelCreatorOrgNameMap),
			StatisticV2Metrics: convertAppV2Metrics(item.GetMetrics()),
		})
	}
	return items, nil
}

func buildAPIKeyStatisticV2RecordItems(ctx *gin.Context, scope *statisticScope, protoItems []*app_service.APIKeyStatisticV2RecordItem) ([]response.APIKeyStatisticV2RecordItem, error) {
	infoMap := getAPIKeyInfoMap(ctx, scope)
	var orgIds, userIds, creatorOrgIds, creatorUserIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, item := range protoItems {
		orgIds = append(orgIds, item.GetOrgId())
		userIds = append(userIds, item.GetUserId())
		if item.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, item.GetModuleCreatorOrgId())
		}
		if item.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, item.GetModuleCreatorUserId())
		}
		addStatisticAppIdByType(appIdsByType, item.GetAppId(), item.GetAppType(), item.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, append(orgIds, creatorOrgIds...), append(userIds, creatorUserIds...))
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)
	items := make([]response.APIKeyStatisticV2RecordItem, 0, len(protoItems))
	for _, item := range protoItems {
		info := getAPIKeyDisplayInfo(ctx, infoMap, item.GetApiKeyId())
		items = append(items, response.APIKeyStatisticV2RecordItem{
			Id:            item.GetId(),
			CalledAt:      item.GetCalledAt(),
			IsSuccess:     item.GetIsSuccess(),
			FailureReason: item.GetFailureReason(),
			RequestBody:   item.GetRequestBody(),
			ResponseBody:  item.GetResponseBody(),
			ApiKeyBriefInfo: buildStatisticV2ApiKeyInfo(ctx, info.name, info.key, item.GetApiKeyId(), item.GetMethodPath(),
				item.GetOrgId(), item.GetUserId(), orgNameMap, userNameMap, nil),
			ModuleBriefInfo: buildStatisticV2AppInfo(ctx, item.GetSource(), item.GetModule(), item.GetAppId(), item.GetAppType(),
				item.GetModuleCreatorUserId(), item.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			StatisticV2RecordPerformance: response.StatisticV2RecordPerformance{
				FirstTokenLatency: item.GetFirstTokenLatency(), // 流式耗时 ms
				Costs:             item.GetCosts(),             // 非流式耗时 ms
			},
		})
	}
	return items, nil
}

func buildStatisticV2ApiKeyInfo(ctx *gin.Context, apiName, apiKey, apiKeyId, methodPath, orgId, userId string,
	orgNameMap, userNameMap map[string]string, userAvatarMap map[string]request.Avatar) response.ApiKeyBriefInfo {
	return response.ApiKeyBriefInfo{
		ApiName:       apiName,
		ApiKeyId:      apiKeyId,
		ApiKey:        apiKey,
		MethodPath:    methodPath,
		UserBriefInfo: buildUserBriefInfo(ctx, userId, orgId, userNameMap, orgNameMap, userAvatarMap),
	}
}

func convertAPIKeyV2Overview(o *app_service.APIKeyStatisticV2Overview) *response.APIKeyStatisticV2Overview {
	if o == nil {
		return &response.APIKeyStatisticV2Overview{}
	}
	return convertStatisticV2CallOverview(
		o.GetCallCount(), o.GetCallFailure(),
		o.GetDailyAvgCallCount(), o.GetDailyAvgCallFailure(),
		o.GetDailyAvgStreamCount(), o.GetDailyAvgNonStreamCount(),
		o.GetAvgFirstTokenLatency(), o.GetAvgCosts(),
		o.GetStreamCount(), o.GetNonStreamCount(),
	)
}

func convertAPIKeyV2Metrics(m *app_service.APIKeyStatisticV2Metrics) response.APIKeyStatisticV2Metrics {
	if m == nil {
		return response.APIKeyStatisticV2Metrics{}
	}
	return response.APIKeyStatisticV2Metrics{
		CallCount:            m.GetCallCount(),
		CallFailure:          m.GetCallFailure(),
		FailureRate:          m.GetFailureRate(),
		AvgFirstTokenLatency: m.GetAvgFirstTokenLatency(),
		AvgCosts:             m.GetAvgCosts(),
		StreamCount:          m.GetStreamCount(),
		NonStreamCount:       m.GetNonStreamCount(),
	}
}
