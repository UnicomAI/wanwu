package service

import (
	"context"
	"math"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetModelStatistic(ctx *gin.Context, filter request.StatisticFilter, startDate, endDate string, modelIds []string, modelType string, userId, orgId string, isAdmin, isSystem bool) (*response.ModelStatistic, error) {
	scope, err := ResolveStatisticScope(ctx, filter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	if modelType == "" {
		modelType = mp.ModelTypeLLM
	}
	resp, err := app.GetModelStatistic(ctx.Request.Context(), &app_service.GetModelStatisticReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: startDate,
		EndDate:   endDate,
		ModelIds:  modelIds,
		ModelType: modelType,
	})
	if err != nil {
		return nil, err
	}
	return &response.ModelStatistic{
		Overview: response.ModelStatisticOverview{
			CallCountTotal:        convertModelStatisticOverviewItem(resp.Overview.GetCallCount()),
			CallFailureTotal:      convertModelStatisticOverviewItem(resp.Overview.GetCallFailure()),
			TotalTokensTotal:      convertModelStatisticOverviewItem(resp.Overview.GetTotalTokens()),
			PromptTokensTotal:     convertModelStatisticOverviewItem(resp.Overview.GetPromptTokens()),
			CompletionTokensTotal: convertModelStatisticOverviewItem(resp.Overview.GetCompletionTokens()),
			AvgCosts:              convertModelStatisticOverviewItem(resp.Overview.GetAvgCosts()),
			AvgFirstTokenLatency:  convertModelStatisticOverviewItem(resp.Overview.GetAvgFirstTokenLatency()),
		},
		Trend: response.ModelStatisticTrend{
			ModelCalls:  convertStatisticChart(ctx, resp.Trend.GetModelCalls()),
			TokensUsage: convertStatisticChart(ctx, resp.Trend.GetTokensUsage()),
		},
	}, nil
}

func GetModelStatisticList(ctx *gin.Context, filter request.StatisticFilter, startDate, endDate string, modelIds []string, modelType string, page, pageSize int32, userId, orgId string, isAdmin, isSystem bool) (*response.PageResult, error) {
	scope, err := ResolveStatisticScope(ctx, filter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	if modelType == "" {
		modelType = mp.ModelTypeLLM
	}
	resp, err := app.GetModelStatisticList(ctx.Request.Context(), &app_service.GetModelStatisticListReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		StartDate: startDate,
		EndDate:   endDate,
		ModelIds:  modelIds,
		ModelType: modelType,
		PageNo:    page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]response.ModelStatisticItem, 0, len(resp.Items))
	// 收集items中的orgId然后获取OrgIds对应的OrgNames
	var orgIDs []string
	var modelIdsRes []string
	var userIDs []string
	for _, item := range resp.Items {
		orgIDs = append(orgIDs, item.OrgId)
		modelIdsRes = append(modelIdsRes, item.ModelId)
		userIDs = append(userIDs, item.UserId)
	}
	orgNameMap, userNameMap, err := buildStatisticOrgUserNameMaps(ctx, orgIDs, userIDs)
	if err != nil {
		return nil, err
	}

	// 调用模型服务获取模型信息
	modelResp, err := model.ListModelsByIds(ctx.Request.Context(), &model_service.ListModelsByIdsReq{
		ModelIds: modelIdsRes,
	})
	if err != nil {
		return nil, err
	}
	// 创建modelId到modelInfo的映射
	displayNameMap := make(map[string]string)
	uuidMap := make(map[string]string)
	for _, model := range modelResp.Models {
		displayNameMap[model.ModelId] = model.DisplayName
		uuidMap[model.ModelId] = model.Uuid
	}
	for _, item := range resp.Items {
		roundedFailureRate := float32(math.Round(float64(item.FailureRate)*100) / 100)
		roundedAvgCosts := float32(math.Round(float64(item.AvgCosts)*100) / 100)
		roundedAvgFirstTokenLatency := float32(math.Round(float64(item.AvgFirstTokenLatency)*100) / 100)
		items = append(items, response.ModelStatisticItem{
			UUID:                 uuidMap[item.ModelId], // 前端不需要展示uuid,excel导出需要
			ModelId:              item.ModelId,
			Model:                getModelDisplayName(displayNameMap, item.ModelId),
			Provider:             item.Provider,
			OrgName:              orgNameMap[item.OrgId],
			UserName:             userNameMap[item.UserId],
			CallCount:            item.CallCount,
			CallFailure:          item.CallFailure,
			FailureRate:          roundedFailureRate,
			PromptTokens:         item.PromptTokens,
			CompletionTokens:     item.CompletionTokens,
			TotalTokens:          item.TotalTokens,
			AvgCosts:             roundedAvgCosts,
			AvgFirstTokenLatency: roundedAvgFirstTokenLatency,
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.Total),
		PageNo:   int(page),
		PageSize: int(pageSize),
	}, nil
}

func ExportModelStatisticList(ctx *gin.Context, filter request.StatisticFilter, startDate, endDate string, modelIds []string, modelType string, userId, orgId string, isAdmin, isSystem bool) (*excelize.File, error) {
	resp, err := GetModelStatisticList(ctx, filter, startDate, endDate, modelIds, modelType, -1, -1, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	return writeModelListExcel(resp.List.([]response.ModelStatisticItem))
}

func recordModelStatistic(ctx context.Context, modelInfo *model_service.ModelInfo, isSuccess bool, promptTokens, completionTokens, totalTokens, costs, firstTokenLatency int, isStream bool) {
	_, err := app.RecordModelStatistic(ctx, &app_service.RecordModelStatisticReq{
		UserId:            modelInfo.UserId,
		OrgId:             modelInfo.OrgId,
		ModelId:           modelInfo.ModelId,
		Model:             modelInfo.Model,
		Provider:          modelInfo.Provider,
		ModelType:         modelInfo.ModelType,
		PromptTokens:      int64(promptTokens),
		CompletionTokens:  int64(completionTokens),
		TotalTokens:       int64(totalTokens),
		FirstTokenLatency: int64(firstTokenLatency),
		Costs:             int64(costs),
		IsSuccess:         isSuccess,
		IsStream:          isStream,
	})
	if err != nil {
		log.Errorf("record modelId %v modelName %v modelType %v statistic err:%v", modelInfo.ModelId, modelInfo.Model, modelInfo.ModelType, err)
	}
}

func convertModelStatisticOverviewItem(item *app_service.ModelStatisticOverviewItem) response.StatisticOverviewItem {
	return response.StatisticOverviewItem{
		Value:            item.Value,
		PeriodOverPeriod: item.PeriodOverPeriod,
	}
}

func writeModelListExcel(items []response.ModelStatisticItem) (*excelize.File, error) {
	sheet := "模型统计列表"
	title := []any{"UUID", "模型", "模型供应商", "组织", "用户", "模型调用次数(次)", "模型调用失败次数(次)", "模型失败率(%)", "Prompt Tokens(个)", "Completion Tokens(个)", "总Tokens(个)", "平均耗时(非流式)(ms)", "平均首Token时延(流式)(ms)"}
	var rows [][]any
	for _, item := range items {
		rows = append(rows, []any{
			item.UUID,
			item.Model,
			item.Provider,
			item.OrgName,
			item.UserName,
			item.CallCount,
			item.CallFailure,
			item.FailureRate,
			item.PromptTokens,
			item.CompletionTokens,
			item.TotalTokens,
			item.AvgCosts,
			item.AvgFirstTokenLatency,
		})
	}
	return writeExcel(sheet, title, rows)
}

func writeExcel(sheet string, title []any, rows [][]any) (*excelize.File, error) {
	f := excelize.NewFile()
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	if err := writeExcelRow(f, sheet, 1, title); err != nil {
		return nil, err
	}
	for i, row := range rows {
		if err := writeExcelRow(f, sheet, i+2, row); err != nil {
			return nil, err
		}
	}
	return f, nil
}

func writeExcelRow(f *excelize.File, sheet string, row int, values []any) error {
	for col, value := range values {
		cell, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}

func getModelDisplayName(displayNameMap map[string]string, modelId string) string {
	if displayName, ok := displayNameMap[modelId]; ok {
		return displayName
	}
	return "该模型已被删除"
}
