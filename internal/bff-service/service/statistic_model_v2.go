package service

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// ============ Overview ============

func GetModelStatisticV2Overview(ctx *gin.Context, req *request.ModelStatisticV2Req, userId, orgId string, isAdmin, isSystem bool) (*response.ModelStatisticV2Overview, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2Overview(ctx.Request.Context(), &app_service.GetModelStatisticV2ReadReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
	})
	if err != nil {
		return nil, err
	}
	return convertV2Overview(resp), nil
}

// ============ Chart (Trend + Rank) ============

func GetModelStatisticV2Chart(ctx *gin.Context, req *request.ModelStatisticV2ChartReq, userId, orgId string, isAdmin, isSystem bool) (*response.ModelStatisticV2Chart, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	limit := int32(req.Limit)
	if limit <= 0 {
		limit = 5
	}
	resp, err := app.GetModelStatisticV2Chart(ctx.Request.Context(), &app_service.GetModelStatisticV2ChartReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	rank, err := buildModelStatisticV2RankResponse(ctx, resp.GetRank())
	if err != nil {
		return nil, err
	}
	return &response.ModelStatisticV2Chart{
		Trend: response.ModelStatisticV2Trend{
			TokensUsage: convertStatisticChart(ctx, resp.GetTrend().GetTokensUsage()),
			ModelCalls:  convertStatisticChart(ctx, resp.GetTrend().GetModelCalls()),
		},
		Rank: *rank,
	}, nil
}

// buildModelStatisticV2RankResponse 富化 rank 响应（名称/头像）。
func buildModelStatisticV2RankResponse(ctx *gin.Context, rank *app_service.ModelStatisticV2Rank) (*response.ModelStatisticV2Rank, error) {
	// 收集 ID 用于富化
	var modelIds, userIds, orgIds []string
	for _, m := range rank.GetByModel() {
		modelIds = append(modelIds, m.GetModelId())
	}
	for _, u := range rank.GetByUser() {
		userIds = append(userIds, u.GetUserId())
		if u.GetOrgId() != "" {
			orgIds = append(orgIds, u.GetOrgId())
		}
	}
	for _, o := range rank.GetByOrg() {
		orgIds = append(orgIds, o.GetOrgId())
	}
	orgNameMap, orgAvatarMap, err := buildStatisticOrgMaps(ctx, orgIds, true)
	if err != nil {
		return nil, err
	}
	userNameMap, userAvatarMap, err := buildStatisticUserMaps(ctx, userIds, true)
	if err != nil {
		return nil, err
	}
	modelMap := getModelInfoMap(ctx, modelIds)
	// 模型发布者名 = 模型创建者 user 名
	var creatorUserIds []string
	for _, m := range modelMap {
		if m.creatorUserId != "" {
			creatorUserIds = append(creatorUserIds, m.creatorUserId)
		}
	}
	creatorUserNameMap, _, err := buildStatisticUserMaps(ctx, creatorUserIds, false)
	if err != nil {
		return nil, err
	}

	byModel := make([]response.ModelStatisticV2RankByModelItem, 0, len(rank.GetByModel()))
	for _, m := range rank.GetByModel() {
		info := modelMap[m.GetModelId()]
		byModel = append(byModel, response.ModelStatisticV2RankByModelItem{
			ModelBriefInfo: response.ModelBriefInfo{
				ModelId:     m.GetModelId(),
				Model:       pickModelDisplayName(info, m.GetModel()),
				Provider:    m.GetProvider(),
				ModelAvatar: cacheModelAvatar(ctx, info.modelIconPath),
				ModelType:   info.ModelType,
				StatisticV2ModelCreator: response.StatisticV2ModelCreator{
					ModelCreatorUserId:   info.creatorUserId,
					ModelCreatorUserName: creatorUserNameMap[info.creatorUserId],
					ModelCreatorOrgId:    info.creatorOrgId,
					ModelCreatorOrgName:  orgNameMap[info.creatorOrgId],
				},
			},
			TotalTokens: m.GetTotalTokens(),
		})
	}
	byUser := make([]response.ModelStatisticV2RankByUserItem, 0, len(rank.GetByUser()))
	for _, u := range rank.GetByUser() {
		byUser = append(byUser, response.ModelStatisticV2RankByUserItem{
			UserId:      u.GetUserId(),
			UserName:    userNameMap[u.GetUserId()],
			Avatar:      userAvatarMap[u.GetUserId()],
			OrgId:       u.GetOrgId(),
			OrgName:     orgNameMap[u.GetOrgId()],
			TotalTokens: u.GetTotalTokens(),
		})
	}
	byOrg := make([]response.ModelStatisticV2RankByOrgItem, 0, len(rank.GetByOrg()))
	for _, o := range rank.GetByOrg() {
		byOrg = append(byOrg, response.ModelStatisticV2RankByOrgItem{
			OrgId:       o.GetOrgId(),
			OrgName:     orgNameMap[o.GetOrgId()],
			Avatar:      orgAvatarMap[o.GetOrgId()],
			TotalTokens: o.GetTotalTokens(),
		})
	}
	return &response.ModelStatisticV2Rank{ByModel: byModel, ByUser: byUser, ByOrg: byOrg}, nil
}

// ============ List (主表) ============

func GetModelStatisticV2List(ctx *gin.Context, req *request.ModelStatisticV2ListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2List(ctx.Request.Context(), &app_service.GetModelStatisticV2ListReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
		SortField: resolveSortExpr(sortFieldAggregateFull, req.SortField),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var modelIds, creatorUserIds, creatorOrgIds []string
	for _, it := range resp.GetItems() {
		modelIds = append(modelIds, it.GetModelId())
		if it.GetModelCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModelCreatorUserId())
		}
		if it.GetModelCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModelCreatorOrgId())
		}
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}
	modelMap := getModelInfoMap(ctx, modelIds)

	items := make([]response.ModelStatisticV2ListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		items = append(items, response.ModelStatisticV2ListItem{
			ModelBriefInfo: buildModelBriefInfo(it.GetModelId(), it.GetModel(), it.GetProvider(), it.GetModelType(),
				it.GetModelCreatorUserId(), it.GetModelCreatorOrgId(), info, userNameMap, orgNameMap),
			ModelStatisticV2Metrics: convertV2Metrics(it.GetMetrics()),
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

func GetModelStatisticV2UserList(ctx *gin.Context, req *request.ModelStatisticV2UserListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2UserList(ctx.Request.Context(), &app_service.GetModelStatisticV2UserListReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
		ModelId:   req.ModelId,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
		SortField: resolveSortExpr(sortFieldAggregateFull, req.SortField),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var userIds, orgIds []string
	var modelIds []string
	for _, it := range resp.GetItems() {
		userIds = append(userIds, it.GetUserId())
		orgIds = append(orgIds, it.GetOrgId())
		modelIds = append(modelIds, it.GetModelId())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, orgIds, userIds)
	if err != nil {
		return nil, err
	}
	modelMap := getModelInfoMap(ctx, modelIds)
	// 模型发布者名 = 模型创建者 user/org 名
	var creatorUserIds, creatorOrgIds []string
	for _, m := range modelMap {
		if m.creatorUserId != "" {
			creatorUserIds = append(creatorUserIds, m.creatorUserId)
		}
		if m.creatorOrgId != "" {
			creatorOrgIds = append(creatorOrgIds, m.creatorOrgId)
		}
	}
	creatorOrgNameMap, creatorUserNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}

	items := make([]response.ModelStatisticV2UserListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		items = append(items, response.ModelStatisticV2UserListItem{
			ModelBriefInfo: buildModelBriefInfo(it.GetModelId(), it.GetModel(), it.GetProvider(), it.GetModelType(),
				info.creatorUserId, info.creatorOrgId, info, creatorUserNameMap, creatorOrgNameMap),

			UserBriefInfo:           buildUserBriefInfo(it.GetUserId(), it.GetOrgId(), userNameMap, orgNameMap, nil),
			ModelStatisticV2Metrics: convertV2Metrics(it.GetMetrics()),
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.GetTotal()),
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

// ============ AppList ============

func GetModelStatisticV2AppList(ctx *gin.Context, req *request.ModelStatisticV2AppListReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2AppList(ctx.Request.Context(), &app_service.GetModelStatisticV2AppListReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
		ModelId:   req.ModelId,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
		SortField: resolveSortExpr(sortFieldAggregateFull, req.SortField),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	var creatorUserIds, creatorOrgIds []string
	var modelIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
		modelIds = append(modelIds, it.GetModelId())
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, creatorOrgIds, creatorUserIds)
	if err != nil {
		return nil, err
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)
	modelMap := getModelInfoMap(ctx, modelIds)
	// 模型发布者名
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

	items := make([]response.ModelStatisticV2AppListItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		items = append(items, response.ModelStatisticV2AppListItem{
			ModelBriefInfo: buildModelBriefInfo(it.GetModelId(), it.GetModel(), it.GetProvider(), it.GetModelType(),
				info.creatorUserId, info.creatorOrgId, info, modelCreatorUserNameMap, modelCreatorOrgNameMap),
			ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
				it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			ModelStatisticV2Metrics: convertV2Metrics(it.GetMetrics()),
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

func GetModelStatisticV2Record(ctx *gin.Context, req *request.ModelStatisticV2RecordReq, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2Record(ctx.Request.Context(), &app_service.GetModelStatisticV2RecordReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		ModelIds:  req.Models,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
		PageNo:    int32(req.PageNo),
		PageSize:  int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	var modelIds, userIds, orgIds, creatorUserIds, creatorOrgIds []string
	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		modelIds = append(modelIds, it.GetModelId())
		userIds = append(userIds, it.GetUserId())
		orgIds = append(orgIds, it.GetOrgId())
		if it.GetModelCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModelCreatorUserId())
		}
		if it.GetModelCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModelCreatorOrgId())
		}
		if it.GetModuleCreatorUserId() != "" {
			creatorUserIds = append(creatorUserIds, it.GetModuleCreatorUserId())
		}
		if it.GetModuleCreatorOrgId() != "" {
			creatorOrgIds = append(creatorOrgIds, it.GetModuleCreatorOrgId())
		}
		addStatisticAppIdByType(appIdsByType, it.GetAppId(), it.GetAppType(), it.GetModule())
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx,
		append(orgIds, creatorOrgIds...), append(userIds, creatorUserIds...))
	if err != nil {
		return nil, err
	}
	modelMap := getModelInfoMap(ctx, modelIds)
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)

	items := make([]response.ModelStatisticV2RecordItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		items = append(items, response.ModelStatisticV2RecordItem{
			Id:               util.Int2Str(it.GetId()),
			TraceId:          it.GetTraceId(),
			TotalTokens:      it.GetTotalTokens(),
			PromptTokens:     it.GetPromptTokens(),
			CompletionTokens: it.GetCompletionTokens(),
			CalledAt:         it.GetCalledAt(),
			IsSuccess:        it.GetIsSuccess(),
			StatusCode:       it.GetStatusCode(),
			FailureReason:    it.GetFailureReason(),
			RequestBody:      it.GetRequestBody(),
			ResponseBody:     it.GetResponseBody(),
			FinishReason:     it.GetFinishReason(),
			ModelBriefInfo: buildModelBriefInfo(it.GetModelId(), it.GetModel(), it.GetProvider(), it.GetModelType(),
				it.GetModelCreatorUserId(), it.GetModelCreatorOrgId(), info, userNameMap, orgNameMap),
			ModuleBriefInfo: buildStatisticV2AppInfo(it.GetSource(), it.GetModule(), it.GetAppId(), it.GetAppType(),
				it.GetModuleCreatorUserId(), it.GetModuleCreatorOrgId(), appBriefMap, orgNameMap, userNameMap),
			UserBriefInfo: response.UserBriefInfo{
				UserId:   it.GetUserId(),
				UserName: userNameMap[it.GetUserId()],
				OrgId:    it.GetOrgId(),
				OrgName:  orgNameMap[it.GetOrgId()],
			},
			StatisticV2RecordPerformance: response.StatisticV2RecordPerformance{
				FirstTokenLatency: it.GetFirstTokenLatency(),
				Costs:             it.GetCosts(),
			},
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

func ExportModelStatisticV2List(ctx *gin.Context, req *request.ModelStatisticV2ExportListReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.ModelStatisticV2ListReq{
		ModelStatisticV2Req: req.ModelStatisticV2Req,
		StatisticSortReq:    req.StatisticSortReq,
		PageNo:              -1,
		PageSize:            -1,
	}
	resp, err := GetModelStatisticV2List(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.ModelStatisticV2ListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "模型统计V2_调用统计"
	title := []any{"模型名称", "供应商", "模型类型", "发布者", "发布者组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均耗时(非流式)(ms)", "平均首Token时延(流式)(ms)"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.Model, it.Provider, it.ModelType,
			it.ModelCreatorUserName, it.ModelCreatorOrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgCosts, it.AvgFirstTokenLatency,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportModelStatisticV2UserList(ctx *gin.Context, req *request.ModelStatisticV2UserExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.ModelStatisticV2UserListReq{
		ModelStatisticV2Req: req.ModelStatisticV2Req,
		ModelId:             req.ModelId,
		StatisticSortReq:    req.StatisticSortReq,
		PageNo:              -1,
		PageSize:            -1,
	}
	resp, err := GetModelStatisticV2UserList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.ModelStatisticV2UserListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "模型统计V2_用户使用"
	title := []any{"模型名称", "供应商", "模型类型", "发布者", "发布者组织", "用户", "用户组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均耗时(非流式)(ms)", "平均首Token时延(流式)(ms)"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.Model, it.Provider, it.ModelType,
			it.ModelCreatorUserName, it.ModelCreatorOrgName,
			it.UserName, it.OrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgCosts, it.AvgFirstTokenLatency,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportModelStatisticV2AppList(ctx *gin.Context, req *request.ModelStatisticV2AppExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.ModelStatisticV2AppListReq{
		ModelStatisticV2Req: req.ModelStatisticV2Req,
		ModelId:             req.ModelId,
		StatisticSortReq:    req.StatisticSortReq,
		PageNo:              -1,
		PageSize:            -1,
	}
	resp, err := GetModelStatisticV2AppList(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.ModelStatisticV2AppListItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "模型统计V2_应用使用"
	title := []any{"模型名称", "供应商", "模型类型", "发布者", "发布者组织", "来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "总Tokens", "Prompt Tokens", "Completion Tokens", "调用次数", "失败次数", "失败率(%)", "平均耗时(非流式)(ms)", "平均首Token时延(流式)(ms)"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.Model, it.Provider, it.ModelType,
			it.ModelCreatorUserName, it.ModelCreatorOrgName,
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.CallCount, it.CallFailure, it.FailureRate,
			it.AvgCosts, it.AvgFirstTokenLatency,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

func ExportModelStatisticV2Record(ctx *gin.Context, req *request.ModelStatisticV2RecordExportReq, userId, orgId string, isAdmin, isSystem bool) (*util.Workbook, error) {
	listReq := request.ModelStatisticV2RecordReq{
		ModelStatisticV2Req: req.ModelStatisticV2Req,
		PageNo:              1,
		PageSize:            10000, // 导出最大行数
	}
	resp, err := GetModelStatisticV2Record(ctx, &listReq, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	items, ok := resp.List.([]response.ModelStatisticV2RecordItem)
	if !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "export list type error")
	}
	sheet := "模型统计V2_调用明细"
	title := []any{"模型名称", "供应商", "模型类型", "发布者", "发布者组织", "来源", "板块", "应用ID", "应用名称", "应用类型", "作者", "作者组织", "用户", "用户组织", "总Tokens(个)", "Prompt Tokens(个)", "Completion Tokens(个)", "流式耗时(ms)", "非流式耗时(ms)", "调用时间", "调用结果", "状态码", "失败原因", "结束原因", "请求体", "响应体"}
	var rows [][]any
	for _, it := range items {
		rows = append(rows, []any{
			it.Model, it.Provider, it.ModelType,
			it.ModelCreatorUserName, it.ModelCreatorOrgName,
			it.SourceName, it.ModuleName, it.AppId, it.AppName, it.AppType,
			it.ModuleCreatorUserName, it.ModuleCreatorOrgName,
			it.UserName, it.OrgName,
			it.TotalTokens, it.PromptTokens, it.CompletionTokens,
			it.FirstTokenLatency, it.Costs,
			it.CalledAt, statisticExportSuccessLabel(it.IsSuccess), it.StatusCode,
			it.FailureReason, it.FinishReason, it.RequestBody, it.ResponseBody,
		})
	}
	return util.WriteSheet(sheet, title, rows)
}

// ============ helpers ============

type modelBriefInfo struct {
	displayName   string
	modelIconPath string
	ModelType     string
	creatorUserId string
	creatorOrgId  string
}

// getModelInfoMap 批量查询模型信息（displayName/modelIconPath/modelType/creatorUserId/creatorOrgId）
func getModelInfoMap(ctx context.Context, modelIds []string) map[string]modelBriefInfo {
	result := map[string]modelBriefInfo{}
	if len(modelIds) == 0 {
		return result
	}
	resp, err := model.ListModelsByIds(ctx, &model_service.ListModelsByIdsReq{
		ModelIds: modelIds,
	})
	if err != nil {
		log.Warnf("getModelInfoMap ListModelsByIds err: %v, modelIds: %v", err, modelIds)
		return result
	}
	for _, m := range resp.GetModels() {
		result[m.GetModelId()] = modelBriefInfo{
			displayName:   m.GetDisplayName(),
			modelIconPath: m.GetModelIconPath(),
			ModelType:     m.GetModelType(),
			creatorUserId: m.GetUserId(),
			creatorOrgId:  m.GetOrgId(),
		}
	}
	return result
}

func pickModelDisplayName(info modelBriefInfo, fallback string) string {
	if info.displayName != "" {
		return info.displayName
	}
	return fallback
}

func buildModelBriefInfo(modelId, model, provider, modelType, creatorUserId, creatorOrgId string,
	info modelBriefInfo, userNameMap, orgNameMap map[string]string) response.ModelBriefInfo {
	return response.ModelBriefInfo{
		ModelId:   modelId,
		Model:     pickModelDisplayName(info, model),
		Provider:  provider,
		ModelType: modelType,
		// ModelAvatar 仅 chart.rank 填充；list/record 不返回头像
		StatisticV2ModelCreator: response.StatisticV2ModelCreator{
			ModelCreatorUserId:   creatorUserId,
			ModelCreatorUserName: userNameMap[creatorUserId],
			ModelCreatorOrgId:    creatorOrgId,
			ModelCreatorOrgName:  orgNameMap[creatorOrgId],
		},
	}
}

func convertV2Overview(o *app_service.ModelStatisticV2Overview) *response.ModelStatisticV2Overview {
	if o == nil {
		return &response.ModelStatisticV2Overview{}
	}
	return &response.ModelStatisticV2Overview{
		TotalTokens:              convertStatisticOverviewItem(o.GetTotalTokens()),
		PromptTokens:             convertStatisticOverviewItem(o.GetPromptTokens()),
		CompletionTokens:         convertStatisticOverviewItem(o.GetCompletionTokens()),
		DailyAvgTotalTokens:      convertStatisticOverviewItem(o.GetDailyAvgTotalTokens()),
		DailyAvgPromptTokens:     convertStatisticOverviewItem(o.GetDailyAvgPromptTokens()),
		DailyAvgCompletionTokens: convertStatisticOverviewItem(o.GetDailyAvgCompletionTokens()),
		CallCount:                convertStatisticOverviewItem(o.GetCallCount()),
		CallFailure:              convertStatisticOverviewItem(o.GetCallFailure()),
		AvgCosts:                 convertStatisticOverviewItem(o.GetAvgCosts()),
		AvgFirstTokenLatency:     convertStatisticOverviewItem(o.GetAvgFirstTokenLatency()),
	}
}

func convertV2Metrics(m *app_service.ModelStatisticV2Metrics) response.ModelStatisticV2Metrics {
	if m == nil {
		return response.ModelStatisticV2Metrics{}
	}
	return response.ModelStatisticV2Metrics{
		TotalTokens:          m.GetTotalTokens(),
		PromptTokens:         m.GetPromptTokens(),
		CompletionTokens:     m.GetCompletionTokens(),
		CallCount:            m.GetCallCount(),
		CallFailure:          m.GetCallFailure(),
		FailureRate:          m.GetFailureRate(),
		AvgCosts:             m.GetAvgCosts(),
		AvgFirstTokenLatency: m.GetAvgFirstTokenLatency(),
	}
}

// resolveStatisticAppType 列表补全用：优先 appType；knowledge 历史写入可能为空，回退 module。
func resolveStatisticAppType(appType, module string) string {
	if appType != "" {
		return appType
	}
	if module == constant.BizModuleResourceKnowledge {
		return constant.BizModuleResourceKnowledge
	}
	return ""
}

// addStatisticAppIdByType 按可解析的 appType 收集 appId，供 getAppBriefMapMulti 批量补全。
func addStatisticAppIdByType(dst map[string]map[string]struct{}, appId, appType, module string) {
	t := resolveStatisticAppType(appType, module)
	if appId == "" || t == "" {
		return
	}
	m, ok := dst[t]
	if !ok {
		m = map[string]struct{}{}
		dst[t] = m
	}
	m[appId] = struct{}{}
}
