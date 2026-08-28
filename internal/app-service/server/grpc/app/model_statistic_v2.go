package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) RecordModelStatisticV2(ctx context.Context, req *app_service.RecordModelStatisticV2Req) (*emptypb.Empty, error) {
	if err := s.cli.RecordModelStatisticV2(ctx, toRecordModelStatisticV2Input(req)); err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetModelStatisticV2Overview(ctx context.Context, req *app_service.GetModelStatisticV2ReadReq) (*app_service.ModelStatisticV2Overview, error) {
	overview, err := s.cli.GetModelStatisticV2Overview(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2Overview(overview), nil
}

func (s *Service) GetModelStatisticV2Chart(ctx context.Context, req *app_service.GetModelStatisticV2ChartReq) (*app_service.ModelStatisticV2Chart, error) {
	chart, err := s.cli.GetModelStatisticV2Chart(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope, req.Limit)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return &app_service.ModelStatisticV2Chart{
		Trend: convertModelStatisticV2Trend(&chart.Trend),
		Rank:  convertModelStatisticV2Rank(&chart.Rank),
	}, nil
}

func (s *Service) GetModelStatisticV2List(ctx context.Context, req *app_service.GetModelStatisticV2ListReq) (*app_service.GetModelStatisticV2ListResp, error) {
	items, total, err := s.cli.GetModelStatisticV2List(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2ListResp(items, total), nil
}

func (s *Service) GetModelStatisticV2UserList(ctx context.Context, req *app_service.GetModelStatisticV2UserListReq) (*app_service.GetModelStatisticV2UserListResp, error) {
	items, total, err := s.cli.GetModelStatisticV2UserList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope, req.ModelId, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2UserListResp(items, total), nil
}

func (s *Service) GetModelStatisticV2AppList(ctx context.Context, req *app_service.GetModelStatisticV2AppListReq) (*app_service.GetModelStatisticV2AppListResp, error) {
	items, total, err := s.cli.GetModelStatisticV2AppList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope, req.ModelId, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2AppListResp(items, total), nil
}

func (s *Service) GetModelStatisticV2Record(ctx context.Context, req *app_service.GetModelStatisticV2RecordReq) (*app_service.GetModelStatisticV2RecordResp, error) {
	items, total, err := s.cli.GetModelStatisticV2Record(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, req.ViewScope, req.ModelId, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2RecordResp(items, total), nil
}

func (s *Service) GetModelStatisticV2Select(ctx context.Context, req *app_service.GetModelStatisticV2SelectReq) (*app_service.GetModelStatisticV2SelectResp, error) {
	items, err := s.cli.ListModelStatisticV2Select(ctx, req.OrgIds, req.UserIds, req.ModelType, req.ViewScope)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticV2SelectResp(items), nil
}

// --- internal ---

func toRecordModelStatisticV2Input(req *app_service.RecordModelStatisticV2Req) *orm.RecordModelStatisticV2Input {
	if req == nil {
		return nil
	}
	return &orm.RecordModelStatisticV2Input{
		TraceID:             req.TraceId,
		UserID:              req.UserId,
		OrgID:               req.OrgId,
		Source:              req.Source,
		Module:              req.Module,
		AppID:               req.AppId,
		AppType:             req.AppType,
		APIKey:              req.ApiKey,
		APIKeyID:            req.ApiKeyId,
		MethodPath:          req.MethodPath,
		ModuleCreatorUserID: req.ModuleCreatorUserId,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgId,
		ModelID:             req.ModelId,
		Model:               req.Model,
		Provider:            req.Provider,
		ModelType:           req.ModelType,
		ModelCreatorUserID:  req.ModelCreatorUserId,
		ModelCreatorOrgID:   req.ModelCreatorOrgId,
		PromptTokens:        req.PromptTokens,
		CompletionTokens:    req.CompletionTokens,
		TotalTokens:         req.TotalTokens,
		FirstTokenLatency:   req.FirstTokenLatency,
		Costs:               req.Costs,
		IsSuccess:           req.IsSuccess,
		IsStream:            req.IsStream,
		StatusCode:          req.StatusCode,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
		FinishReason:        req.FinishReason,
		FailureReason:       req.FailureReason,
	}
}

func convertModelStatisticV2OverviewItem(item orm.StatisticOverviewItem) *app_service.ModelStatisticV2OverviewItem {
	return &app_service.ModelStatisticV2OverviewItem{
		Value:            item.Value,
		PeriodOverPeriod: item.PeriodOverPeriod,
	}
}

func convertModelStatisticV2Overview(o *orm.ModelStatisticV2Overview) *app_service.ModelStatisticV2Overview {
	if o == nil {
		return &app_service.ModelStatisticV2Overview{}
	}
	return &app_service.ModelStatisticV2Overview{
		TotalTokens:              convertModelStatisticV2OverviewItem(o.TotalTokens),
		PromptTokens:             convertModelStatisticV2OverviewItem(o.PromptTokens),
		CompletionTokens:         convertModelStatisticV2OverviewItem(o.CompletionTokens),
		DailyAvgTotalTokens:      convertModelStatisticV2OverviewItem(o.DailyAvgTotalTokens),
		DailyAvgPromptTokens:     convertModelStatisticV2OverviewItem(o.DailyAvgPromptTokens),
		DailyAvgCompletionTokens: convertModelStatisticV2OverviewItem(o.DailyAvgCompletionTokens),
		CallCount:                convertModelStatisticV2OverviewItem(o.CallCount),
		CallFailure:              convertModelStatisticV2OverviewItem(o.CallFailure),
		AvgCosts:                 convertModelStatisticV2OverviewItem(o.AvgCosts),
		AvgFirstTokenLatency:     convertModelStatisticV2OverviewItem(o.AvgFirstTokenLatency),
	}
}

func convertModelStatisticV2Trend(t *orm.ModelStatisticV2Trend) *app_service.ModelStatisticV2Trend {
	if t == nil {
		return &app_service.ModelStatisticV2Trend{}
	}
	return &app_service.ModelStatisticV2Trend{
		TokensUsage: convertStatisticChart(t.TokensUsage),
		ModelCalls:  convertStatisticChart(t.ModelCalls),
	}
}

func convertModelStatisticV2Rank(r *orm.ModelStatisticV2Rank) *app_service.ModelStatisticV2Rank {
	if r == nil {
		return &app_service.ModelStatisticV2Rank{}
	}
	byModel := make([]*app_service.ModelStatisticV2RankByModelItem, 0, len(r.ByModel))
	for _, m := range r.ByModel {
		byModel = append(byModel, &app_service.ModelStatisticV2RankByModelItem{
			ModelId:            m.ModelId,
			Model:              m.Model,
			Provider:           m.Provider,
			ModelType:          m.ModelType,
			ModelCreatorUserId: m.ModelCreatorUserId,
			ModelCreatorOrgId:  m.ModelCreatorOrgId,
			TotalTokens:        m.TotalTokens,
		})
	}
	byUser := make([]*app_service.ModelStatisticV2RankByUserItem, 0, len(r.ByUser))
	for _, u := range r.ByUser {
		byUser = append(byUser, &app_service.ModelStatisticV2RankByUserItem{
			UserId:      u.UserId,
			OrgId:       u.OrgId,
			TotalTokens: u.TotalTokens,
		})
	}
	byOrg := make([]*app_service.ModelStatisticV2RankByOrgItem, 0, len(r.ByOrg))
	for _, o := range r.ByOrg {
		byOrg = append(byOrg, &app_service.ModelStatisticV2RankByOrgItem{
			OrgId:       o.OrgId,
			TotalTokens: o.TotalTokens,
		})
	}
	return &app_service.ModelStatisticV2Rank{
		ByModel: byModel,
		ByUser:  byUser,
		ByOrg:   byOrg,
	}
}

func convertModelStatisticV2Metrics(m orm.ModelStatisticV2Metrics) *app_service.ModelStatisticV2Metrics {
	return &app_service.ModelStatisticV2Metrics{
		TotalTokens:          m.TotalTokens,
		PromptTokens:         m.PromptTokens,
		CompletionTokens:     m.CompletionTokens,
		CallCount:            m.CallCount,
		CallFailure:          m.CallFailure,
		FailureRate:          m.FailureRate,
		AvgCosts:             m.AvgCosts,
		AvgFirstTokenLatency: m.AvgFirstTokenLatency,
	}
}

func convertModelStatisticV2ListResp(items []orm.ModelStatisticV2ListItem, total int32) *app_service.GetModelStatisticV2ListResp {
	pbItems := make([]*app_service.ModelStatisticV2ListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.ModelStatisticV2ListItem{
			ModelId:            it.ModelId,
			Model:              it.Model,
			Provider:           it.Provider,
			ModelType:          it.ModelType,
			ModelCreatorUserId: it.ModelCreatorUserId,
			ModelCreatorOrgId:  it.ModelCreatorOrgId,
			Metrics:            convertModelStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetModelStatisticV2ListResp{Items: pbItems, Total: total}
}

func convertModelStatisticV2UserListResp(items []orm.ModelStatisticV2UserListItem, total int32) *app_service.GetModelStatisticV2UserListResp {
	pbItems := make([]*app_service.ModelStatisticV2UserListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.ModelStatisticV2UserListItem{
			ModelId:            it.ModelId,
			Model:              it.Model,
			Provider:           it.Provider,
			ModelType:          it.ModelType,
			UserId:             it.UserId,
			OrgId:              it.OrgId,
			Metrics:            convertModelStatisticV2Metrics(it.Metrics),
			ModelCreatorUserId: it.ModelCreatorUserId,
			ModelCreatorOrgId:  it.ModelCreatorOrgId,
		})
	}
	return &app_service.GetModelStatisticV2UserListResp{Items: pbItems, Total: total}
}

func convertModelStatisticV2AppListResp(items []orm.ModelStatisticV2AppListItem, total int32) *app_service.GetModelStatisticV2AppListResp {
	pbItems := make([]*app_service.ModelStatisticV2AppListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.ModelStatisticV2AppListItem{
			ModelId:             it.ModelId,
			Model:               it.Model,
			Provider:            it.Provider,
			ModelType:           it.ModelType,
			Source:              it.Source,
			Module:              it.Module,
			AppId:               it.AppId,
			AppType:             it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId,
			ModuleCreatorOrgId:  it.ModuleCreatorOrgId,
			Metrics:             convertModelStatisticV2Metrics(it.Metrics),
			ModelCreatorUserId:  it.ModelCreatorUserId,
			ModelCreatorOrgId:   it.ModelCreatorOrgId,
		})
	}
	return &app_service.GetModelStatisticV2AppListResp{Items: pbItems, Total: total}
}

func convertModelStatisticV2RecordResp(items []orm.ModelStatisticV2RecordItem, total int32) *app_service.GetModelStatisticV2RecordResp {
	pbItems := make([]*app_service.ModelStatisticV2RecordItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.ModelStatisticV2RecordItem{
			Id:                  it.Id,
			TraceId:             it.TraceId,
			ModelId:             it.ModelId,
			Model:               it.Model,
			Provider:            it.Provider,
			ModelType:           it.ModelType,
			AppId:               it.AppId,
			AppType:             it.AppType,
			OrgId:               it.OrgId,
			UserId:              it.UserId,
			ModelCreatorUserId:  it.ModelCreatorUserId,
			ModelCreatorOrgId:   it.ModelCreatorOrgId,
			Source:              it.Source,
			Module:              it.Module,
			ModuleCreatorUserId: it.ModuleCreatorUserId,
			ModuleCreatorOrgId:  it.ModuleCreatorOrgId,
			TotalTokens:         it.TotalTokens,
			CalledAt:            util.Time2Str(it.CreatedAt),
			IsSuccess:           it.IsSuccess,
			StatusCode:          it.StatusCode,
			PromptTokens:        it.PromptTokens,
			CompletionTokens:    it.CompletionTokens,
			FailureReason:       it.FailureReason,
			RequestBody:         it.RequestBody,
			ResponseBody:        it.ResponseBody,
			FinishReason:        it.FinishReason,
			FirstTokenLatency:   it.FirstTokenLatency,
			Costs:               it.Costs,
		})
	}
	return &app_service.GetModelStatisticV2RecordResp{Items: pbItems, Total: total}
}

func convertModelStatisticV2SelectResp(items []orm.ModelStatisticV2SelectItem) *app_service.GetModelStatisticV2SelectResp {
	pbItems := make([]*app_service.ModelStatisticV2SelectItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.ModelStatisticV2SelectItem{
			ModelId:            it.ModelId,
			Model:              it.Model,
			Provider:           it.Provider,
			ModelType:          it.ModelType,
			ModelCreatorUserId: it.ModelCreatorUserId,
			ModelCreatorOrgId:  it.ModelCreatorOrgId,
		})
	}
	return &app_service.GetModelStatisticV2SelectResp{Items: pbItems}
}

func convertStatisticChart(chart orm.StatisticChart) *common.StatisticChart {
	pbChart := &common.StatisticChart{
		TableName:  chart.Name,
		ChartLines: make([]*common.StatisticChartLine, 0, len(chart.Lines)),
	}
	for _, line := range chart.Lines {
		pbLine := &common.StatisticChartLine{
			LineName: line.Name,
			Items:    make([]*common.StatisticChartLineItem, 0, len(line.Items)),
		}
		for _, item := range line.Items {
			pbLine.Items = append(pbLine.Items, &common.StatisticChartLineItem{
				Key:   item.Key,
				Value: item.Value,
			})
		}
		pbChart.ChartLines = append(pbChart.ChartLines, pbLine)
	}
	return pbChart
}
