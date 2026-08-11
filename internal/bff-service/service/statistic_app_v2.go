package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// ============ Overview ============

func GetAppStatisticV2Overview(ctx *gin.Context, req *request.AppStatisticV2Req, userId, orgId string, isAdmin, isSystem bool) (*response.AppStatisticV2Overview, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAppStatisticV2Overview(ctx.Request.Context(), &app_service.GetAppStatisticV2ReadReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Module:    req.Module,
		Apps:      req.Apps,
		ViewScope: req.ViewScope,
		Source:    req.Source,
	})
	if err != nil {
		return nil, err
	}
	return convertAppV2Overview(resp), nil
}

// ============ Chart (Trend + Rank) ============

func GetAppStatisticV2Chart(ctx *gin.Context, req *request.AppStatisticV2ChartReq, userId, orgId string, isAdmin, isSystem bool) (*response.AppStatisticV2Chart, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	limit := int32(req.Limit)
	if limit <= 0 {
		limit = 5
	}
	resp, err := app.GetAppStatisticV2Chart(ctx.Request.Context(), &app_service.GetAppStatisticV2ChartReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Module:    req.Module,
		Apps:      req.Apps,
		ViewScope: req.ViewScope,
		Source:    req.Source,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	rank, err := buildAppStatisticV2RankResponse(ctx, resp.GetRank())
	if err != nil {
		return nil, err
	}
	return &response.AppStatisticV2Chart{
		Trend: response.AppStatisticV2Trend{
			CallResult: convertStatisticChart(ctx, resp.GetTrend().GetCallResult()),
			CallTrend:  convertStatisticChart(ctx, resp.GetTrend().GetCallTrend()),
		},
		Rank: *rank,
	}, nil
}

func buildAppStatisticV2RankResponse(ctx *gin.Context, rank *app_service.AppStatisticV2Rank) (*response.AppStatisticV2Rank, error) {
	if rank == nil {
		return &response.AppStatisticV2Rank{}, nil
	}
	allItems := append(rank.GetByAgent(), rank.GetByWorkflow()...)
	allItems = append(allItems, rank.GetByChatflow()...)
	allItems = append(allItems, rank.GetByRag()...)
	allItems = append(allItems, rank.GetByDigitalEmployee()...)

	appIdsByType := map[string]map[string]struct{}{}
	var creatorUserIds, creatorOrgIds []string
	for _, it := range allItems {
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), "")
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)

	convertItems := func(items []*app_service.AppStatisticV2RankItem) []response.StatisticV2RankItem {
		out := make([]response.StatisticV2RankItem, 0, len(items))
		for _, it := range items {
			out = append(out, response.StatisticV2RankItem{
				AppId:   it.GetAppId(),
				AppName: pickAppName(appBriefMap, it.GetAppType(), it.GetAppId(), it.GetAppId()),
				Avatar:  pickAppAvatar(appBriefMap, it.GetAppType(), it.GetAppId()),
				StatisticV2ModuleCreator: response.StatisticV2ModuleCreator{
					ModuleCreatorUserId:   it.GetModuleCreatorUserId(),
					ModuleCreatorUserName: userNameMap[it.GetModuleCreatorUserId()],
					ModuleCreatorOrgId:    it.GetModuleCreatorOrgId(),
					ModuleCreatorOrgName:  orgNameMap[it.GetModuleCreatorOrgId()],
				},
				CallCount: it.GetCallCount(),
			})
		}
		return out
	}
	return &response.AppStatisticV2Rank{
		ByAgent:           convertItems(rank.GetByAgent()),
		ByWorkflow:        convertItems(rank.GetByWorkflow()),
		ByChatflow:        convertItems(rank.GetByChatflow()),
		ByRag:             convertItems(rank.GetByRag()),
		ByDigitalEmployee: convertItems(rank.GetByDigitalEmployee()),
	}, nil
}

// ============ List (主表) ============

func GetAppStatisticV2List(ctx *gin.Context, req *request.AppStatisticV2ListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAppStatisticV2List(ctx.Request.Context(), &app_service.GetAppStatisticV2ListReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Module:    req.Module,
		Apps:      req.Apps,
		ViewScope: req.ViewScope,
		Source:    req.Source,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
		SortField: resolveSortExpr(sortFieldAggregateCallsOnly, req.SortField),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var creatorUserIds, creatorOrgIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)

	items := make([]response.AppStatisticV2ListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, response.AppStatisticV2ListItem{
			ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
				it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			StatisticV2Metrics: convertAppV2Metrics(it.GetMetrics()),
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ UserList ============

func GetAppStatisticV2UserList(ctx *gin.Context, req *request.AppStatisticV2UserListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAppStatisticV2UserList(ctx.Request.Context(), &app_service.GetAppStatisticV2UserListReq{
		OrgIds:              scope.OrgIds,
		UserIds:             scope.UserIds,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		Module:              req.Module,
		Apps:                req.Apps,
		ViewScope:           req.ViewScope,
		Source:              req.Source,
		AppId:               req.AppId,
		ModuleCreatorUserId: req.ModuleCreatorUserId,
		ModuleCreatorOrgId:  req.ModuleCreatorOrgId,
		PageNo:              int32(req.PageNo),
		PageSize:            int32(req.PageSize),
		SortField:           resolveSortExpr(sortFieldAggregateCallsOnly, req.SortField),
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var userIds, orgIds []string
	var creatorUserIds, creatorOrgIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		userIds = append(userIds, it.GetUserId())
		orgIds = append(orgIds, it.GetOrgId())
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, append(orgIds, creatorOrgIds...), append(userIds, creatorUserIds...))
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)

	items := make([]response.AppStatisticV2UserListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, response.AppStatisticV2UserListItem{
			ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
				it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			UserBriefInfo:      buildUserBriefInfo(it.GetUserId(), it.GetOrgId(), userNameMap, orgNameMap, nil),
			StatisticV2Metrics: convertAppV2Metrics(it.GetMetrics()),
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ ModelList ============

func GetAppStatisticV2ModelList(ctx *gin.Context, req *request.AppStatisticV2ModelListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAppStatisticV2ModelList(ctx.Request.Context(), &app_service.GetAppStatisticV2ModelListReq{
		OrgIds:              scope.OrgIds,
		UserIds:             scope.UserIds,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		Module:              req.Module,
		Apps:                req.Apps,
		ViewScope:           req.ViewScope,
		Source:              req.Source,
		AppId:               req.AppId,
		ModuleCreatorUserId: req.ModuleCreatorUserId,
		ModuleCreatorOrgId:  req.ModuleCreatorOrgId,
		PageNo:              int32(req.PageNo),
		PageSize:            int32(req.PageSize),
		SortField:           resolveSortExpr(sortFieldAggregateFull, req.SortField),
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var modelIds, creatorUserIds, creatorOrgIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		modelIds = append(modelIds, it.GetModelId())
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)
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

	items := make([]response.AppStatisticV2ModelListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		items = append(items, response.AppStatisticV2ModelListItem{
			ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
				it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			ModelBriefInfo: buildModelBriefInfo(it.GetModelId(), it.GetModel(), it.GetProvider(), it.GetModelType(),
				info.creatorUserId, info.creatorOrgId, info, modelCreatorUserNameMap, modelCreatorOrgNameMap),
			StatisticV2Metrics: convertAppV2Metrics(it.GetMetrics()),
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ Record 列表 ============

func GetAppStatisticV2Record(ctx *gin.Context, req *request.AppStatisticV2RecordReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetAppStatisticV2Record(ctx.Request.Context(), &app_service.GetAppStatisticV2RecordReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Module:    req.Module,
		Apps:      req.Apps,
		ViewScope: req.ViewScope,
		AppId:     req.AppId,
		Source:    req.Source,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	appIdsByType := map[string]map[string]struct{}{}
	var userIds, orgIds, creatorUserIds, creatorOrgIds []string
	for _, it := range resp.GetItems() {
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
		userIds = append(userIds, it.GetUserId())
		orgIds = append(orgIds, it.GetOrgId())
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx,
		append(orgIds, creatorOrgIds...), append(userIds, creatorUserIds...))
	if err != nil {
		return nil, err
	}

	items := make([]response.AppStatisticV2RecordItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, response.AppStatisticV2RecordItem{
			StatisticV2AppRecordBase: response.StatisticV2AppRecordBase{
				Id:         it.GetId(),
				TraceId:    it.GetTraceId(),
				CalledAt:   it.GetCalledAt(),
				IsSuccess:  it.GetIsSuccess(),
				StatusCode: it.GetStatusCode(),
				ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
					it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
				UserBriefInfo: buildUserBriefInfo(it.GetUserId(), it.GetOrgId(), userNameMap, orgNameMap, nil),
			},
			StatisticV2RecordPerformance: response.StatisticV2RecordPerformance{
				FirstTokenLatency: it.GetFirstTokenLatency(),
				Costs:             it.GetCosts(),
			},
			FailureReason: it.GetFailureReason(),
			RequestBody:   it.GetRequestBody(),
			ResponseBody:  it.GetResponseBody(),
			Question:      it.GetQuestion(),
			Answer:        it.GetAnswer(),
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ 导出 ============

func ExportAppStatisticV2List(ctx *gin.Context, req *request.AppStatisticV2ExportListReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.AppStatisticV2ListReq{
		AppStatisticV2Req: req.AppStatisticV2Req,
		StatisticSortReq:  req.StatisticSortReq,
		PageNo:            -1,
		PageSize:          -1,
	}
	resp, err := GetAppStatisticV2List(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.AppStatisticV2ListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "应用统计V2_调用统计"
	// 与 list 响应 StatisticV2Metrics 对齐；耗时列区分流式/非流式
	title := []any{"来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均首Token时延(流式)(ms)", "平均耗时(非流式)(ms)", "流式次数", "非流式次数"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgFirstTokenLatency, it.AvgCosts,
			it.StreamCount, it.NonStreamCount,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportAppStatisticV2UserList(ctx *gin.Context, req *request.AppStatisticV2UserExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.AppStatisticV2UserListReq{
		AppStatisticV2Req:   req.AppStatisticV2Req,
		Source:              req.Source,
		AppId:               req.AppId,
		ModuleCreatorUserId: req.ModuleCreatorUserId,
		ModuleCreatorOrgId:  req.ModuleCreatorOrgId,
		StatisticSortReq:    req.StatisticSortReq,
		PageNo:              -1,
		PageSize:            -1,
	}
	resp, err := GetAppStatisticV2UserList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.AppStatisticV2UserListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "应用统计V2_用户使用"
	title := []any{"来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "用户", "用户组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均首Token时延(流式)(ms)", "平均耗时(非流式)(ms)", "流式次数", "非流式次数"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.UserName, it.OrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgFirstTokenLatency, it.AvgCosts,
			it.StreamCount, it.NonStreamCount,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportAppStatisticV2ModelList(ctx *gin.Context, req *request.AppStatisticV2ModelExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.AppStatisticV2ModelListReq{
		AppStatisticV2Req:   req.AppStatisticV2Req,
		Source:              req.Source,
		AppId:               req.AppId,
		ModuleCreatorUserId: req.ModuleCreatorUserId,
		ModuleCreatorOrgId:  req.ModuleCreatorOrgId,
		StatisticSortReq:    req.StatisticSortReq,
		PageNo:              -1,
		PageSize:            -1,
	}
	resp, err := GetAppStatisticV2ModelList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.AppStatisticV2ModelListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "应用统计V2_模型使用"
	title := []any{"来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "模型ID", "模型名称", "供应商", "模型类型", "发布者", "发布者组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均首Token时延(流式)(ms)", "平均耗时(非流式)(ms)", "流式次数", "非流式次数"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
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

func ExportAppStatisticV2Record(ctx *gin.Context, req *request.AppStatisticV2RecordExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.AppStatisticV2RecordReq{
		AppStatisticV2Req: req.AppStatisticV2Req,
		AppId:             req.AppId,
		PageNo:            1,
		PageSize:          10000, // 导出最大行数
	}
	resp, err := GetAppStatisticV2Record(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.AppStatisticV2RecordItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "应用统计V2_调用明细"
	title := []any{"来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "使用人", "使用人组织", "流式耗时(ms)", "非流式耗时(ms)", "调用结果", "状态码", "调用时间", "失败原因", "请求体", "响应体", "用户提问", "回复"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.UserName, it.OrgName,
			it.FirstTokenLatency, it.Costs,
			statisticExportSuccessLabel(it.IsSuccess), it.StatusCode, it.CalledAt,
			it.FailureReason, it.RequestBody, it.ResponseBody, it.Question, it.Answer,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

// ============ helpers ============

func buildStatisticV2AppInfo(source, module, appId, appType, authorId, authorOrgId string,
	appBriefMap map[string]map[string]appBrief, orgNameMap, userNameMap map[string]string) response.ModuleBriefInfo {
	appType = resolveStatisticAppType(appType, module)
	return response.ModuleBriefInfo{
		Source:       source,
		SourceName:   constant.BizSourceName(source),
		Module:       module,
		ModuleName:   constant.BizModuleName(module),
		AppId:        appId,
		AppName:      pickAppName(appBriefMap, appType, appId, appId),
		AppType:      appType,
		ModuleAvatar: pickAppAvatar(appBriefMap, appType, appId),
		StatisticV2ModuleCreator: response.StatisticV2ModuleCreator{
			ModuleCreatorUserId:   authorId,
			ModuleCreatorUserName: userNameMap[authorId],
			ModuleCreatorOrgId:    authorOrgId,
			ModuleCreatorOrgName:  orgNameMap[authorOrgId],
		},
	}
}

func convertAppV2Overview(o *app_service.AppStatisticV2Overview) *response.AppStatisticV2Overview {
	if o == nil {
		return &response.AppStatisticV2Overview{}
	}
	return convertStatisticV2CallOverview(
		o.GetCallCount(), o.GetCallFailure(),
		o.GetDailyAvgCallCount(), o.GetDailyAvgCallFailure(),
		o.GetDailyAvgStreamCount(), o.GetDailyAvgNonStreamCount(),
		o.GetAvgFirstTokenLatency(), o.GetAvgCosts(),
		o.GetStreamCount(), o.GetNonStreamCount(),
	)
}

func convertAppV2Metrics(m *app_service.AppStatisticV2Metrics) response.StatisticV2Metrics {
	if m == nil {
		return response.StatisticV2Metrics{}
	}
	return response.StatisticV2Metrics{
		TotalTokens:          m.GetTotalTokens(),
		PromptTokens:         m.GetPromptTokens(),
		CompletionTokens:     m.GetCompletionTokens(),
		CallCount:            m.GetCallCount(),
		CallFailure:          m.GetCallFailure(),
		FailureRate:          m.GetFailureRate(),
		AvgCosts:             m.GetAvgCosts(),
		AvgFirstTokenLatency: m.GetAvgFirstTokenLatency(),
		StreamCount:          m.GetStreamCount(),
		NonStreamCount:       m.GetNonStreamCount(),
	}
}
