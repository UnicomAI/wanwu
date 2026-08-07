package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) RecordAppStatisticV2(ctx context.Context, req *app_service.RecordAppStatisticV2Req) (*emptypb.Empty, error) {
	if err := s.cli.RecordAppStatisticV2(ctx, toRecordAppStatisticV2Input(req)); err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetAppStatisticV2Overview(ctx context.Context, req *app_service.GetAppStatisticV2ReadReq) (*app_service.AppStatisticV2Overview, error) {
	overview, err := s.cli.GetAppStatisticV2Overview(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.Source)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2Overview(overview), nil
}

func (s *Service) GetAppStatisticV2Chart(ctx context.Context, req *app_service.GetAppStatisticV2ChartReq) (*app_service.AppStatisticV2Chart, error) {
	chart, err := s.cli.GetAppStatisticV2Chart(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.Source, req.Limit)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2Chart(chart), nil
}

func (s *Service) GetAppStatisticV2List(ctx context.Context, req *app_service.GetAppStatisticV2ListReq) (*app_service.GetAppStatisticV2ListResp, error) {
	items, total, err := s.cli.GetAppStatisticV2List(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.Source, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2ListResp(items, total), nil
}

func (s *Service) GetAppStatisticV2UserList(ctx context.Context, req *app_service.GetAppStatisticV2UserListReq) (*app_service.GetAppStatisticV2UserListResp, error) {
	items, total, err := s.cli.GetAppStatisticV2UserList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.Source, req.AppId, req.ModuleCreatorUserId, req.ModuleCreatorOrgId, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2UserListResp(items, total), nil
}

func (s *Service) GetAppStatisticV2ModelList(ctx context.Context, req *app_service.GetAppStatisticV2ModelListReq) (*app_service.GetAppStatisticV2ModelListResp, error) {
	items, total, err := s.cli.GetAppStatisticV2ModelList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.Source, req.AppId, req.ModuleCreatorUserId, req.ModuleCreatorOrgId, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2ModelListResp(items, total), nil
}

func (s *Service) GetAppStatisticV2Record(ctx context.Context, req *app_service.GetAppStatisticV2RecordReq) (*app_service.GetAppStatisticV2RecordResp, error) {
	items, total, err := s.cli.GetAppStatisticV2Record(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.Module, req.Apps, req.ViewScope, req.AppId, req.Source, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2RecordResp(items, total), nil
}

func (s *Service) GetAppStatisticV2Select(ctx context.Context, req *app_service.GetAppStatisticV2SelectReq) (*app_service.GetAppStatisticV2SelectResp, error) {
	items, err := s.cli.ListAppStatisticV2Select(ctx, req.OrgIds, req.UserIds, req.Module, req.ViewScope)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertAppStatisticV2SelectResp(items), nil
}

// --- internal ---

func toRecordAppStatisticV2Input(req *app_service.RecordAppStatisticV2Req) *orm.RecordAppStatisticV2Input {
	if req == nil {
		return nil
	}
	return &orm.RecordAppStatisticV2Input{
		TraceID:             req.TraceId,
		UserID:              req.UserId,
		OrgID:               req.OrgId,
		Source:              req.Source,
		Module:              req.Module,
		AppID:               req.AppId,
		AppType:             req.AppType,
		ModuleCreatorUserID: req.ModuleCreatorUserId,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgId,
		FirstTokenLatency:   req.FirstTokenLatency,
		Costs:               req.Costs,
		IsSuccess:           req.IsSuccess,
		IsStream:            req.IsStream,
		StatusCode:          req.StatusCode,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
		FinishReason:        req.FinishReason,
		FailureReason:       req.FailureReason,
		Question:            req.Question,
		Answer:              req.Answer,
	}
}

func convertAppStatisticV2OverviewItem(item orm.StatisticOverviewItem) *app_service.AppStatisticV2OverviewItem {
	return &app_service.AppStatisticV2OverviewItem{
		Value: item.Value, PeriodOverPeriod: item.PeriodOverPeriod,
	}
}

func convertAppStatisticV2Overview(o *orm.AppStatisticV2Overview) *app_service.AppStatisticV2Overview {
	if o == nil {
		return &app_service.AppStatisticV2Overview{}
	}
	return &app_service.AppStatisticV2Overview{
		CallCount:              convertAppStatisticV2OverviewItem(o.CallCount),
		CallFailure:            convertAppStatisticV2OverviewItem(o.CallFailure),
		DailyAvgCallCount:      convertAppStatisticV2OverviewItem(o.DailyAvgCallCount),
		DailyAvgCallFailure:    convertAppStatisticV2OverviewItem(o.DailyAvgCallFailure),
		DailyAvgStreamCount:    convertAppStatisticV2OverviewItem(o.DailyAvgStreamCount),
		DailyAvgNonStreamCount: convertAppStatisticV2OverviewItem(o.DailyAvgNonStreamCount),
		AvgFirstTokenLatency:   convertAppStatisticV2OverviewItem(o.AvgFirstTokenLatency),
		AvgCosts:               convertAppStatisticV2OverviewItem(o.AvgCosts),
		StreamCount:            convertAppStatisticV2OverviewItem(o.StreamCount),
		NonStreamCount:         convertAppStatisticV2OverviewItem(o.NonStreamCount),
	}
}

func convertAppStatisticV2Chart(c *orm.AppStatisticV2Chart) *app_service.AppStatisticV2Chart {
	if c == nil {
		return &app_service.AppStatisticV2Chart{}
	}
	convertRank := func(items []orm.AppStatisticV2RankItem) []*app_service.AppStatisticV2RankItem {
		out := make([]*app_service.AppStatisticV2RankItem, 0, len(items))
		for _, it := range items {
			out = append(out, &app_service.AppStatisticV2RankItem{
				AppId: it.AppId, AppType: it.AppType,
				ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
				CallCount: it.CallCount,
			})
		}
		return out
	}
	return &app_service.AppStatisticV2Chart{
		Trend: &app_service.AppStatisticV2Trend{
			CallResult: convertStatisticChart(c.Trend.CallResult),
			CallTrend:  convertStatisticChart(c.Trend.CallTrend),
		},
		Rank: &app_service.AppStatisticV2Rank{
			ByAgent: convertRank(c.Rank.ByAgent), ByWorkflow: convertRank(c.Rank.ByWorkflow),
			ByChatflow: convertRank(c.Rank.ByChatflow), ByRag: convertRank(c.Rank.ByRag),
		},
	}
}

func convertAppStatisticV2Metrics(m orm.AppStatisticV2Metrics) *app_service.AppStatisticV2Metrics {
	return &app_service.AppStatisticV2Metrics{
		TotalTokens: m.TotalTokens, PromptTokens: m.PromptTokens, CompletionTokens: m.CompletionTokens,
		CallCount: m.CallCount, CallFailure: m.CallFailure, FailureRate: m.FailureRate,
		AvgCosts: m.AvgCosts, AvgFirstTokenLatency: m.AvgFirstTokenLatency,
		StreamCount: m.StreamCount, NonStreamCount: m.NonStreamCount,
	}
}

func convertAppStatisticV2ListResp(items []orm.AppStatisticV2ListItem, total int32) *app_service.GetAppStatisticV2ListResp {
	pbItems := make([]*app_service.AppStatisticV2ListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.AppStatisticV2ListItem{
			Source: it.Source, Module: it.Module, AppId: it.AppId, AppType: it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
			Metrics: convertAppStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAppStatisticV2ListResp{Items: pbItems, Total: total}
}

func convertAppStatisticV2UserListResp(items []orm.AppStatisticV2UserListItem, total int32) *app_service.GetAppStatisticV2UserListResp {
	pbItems := make([]*app_service.AppStatisticV2UserListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.AppStatisticV2UserListItem{
			Source: it.Source, Module: it.Module, AppId: it.AppId, AppType: it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
			UserId: it.UserId, OrgId: it.OrgId, Metrics: convertAppStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAppStatisticV2UserListResp{Items: pbItems, Total: total}
}

func convertAppStatisticV2ModelListResp(items []orm.AppStatisticV2ModelListItem, total int32) *app_service.GetAppStatisticV2ModelListResp {
	pbItems := make([]*app_service.AppStatisticV2ModelListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.AppStatisticV2ModelListItem{
			Source: it.Source, Module: it.Module, AppId: it.AppId, AppType: it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
			ModelId: it.ModelId, Model: it.Model, Provider: it.Provider, ModelType: it.ModelType,
			ModelCreatorUserId: it.ModelCreatorUserId, ModelCreatorOrgId: it.ModelCreatorOrgId,
			Metrics: convertAppStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAppStatisticV2ModelListResp{Items: pbItems, Total: total}
}

func convertAppStatisticV2RecordResp(items []orm.AppStatisticV2RecordItem, total int32) *app_service.GetAppStatisticV2RecordResp {
	pbItems := make([]*app_service.AppStatisticV2RecordItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.AppStatisticV2RecordItem{
			Id: it.Id, TraceId: it.TraceId, AppId: it.AppId, AppType: it.AppType,
			UserId: it.UserId, OrgId: it.OrgId, Source: it.Source, Module: it.Module,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
			CalledAt: util.Time2Str(it.CallTime), IsSuccess: it.IsSuccess, StatusCode: it.StatusCode,
			Costs: it.Costs, FirstTokenLatency: it.FirstTokenLatency,
			FailureReason: it.FailureReason, RequestBody: it.RequestBody, ResponseBody: it.ResponseBody,
			Question: it.Question, Answer: it.Answer,
		})
	}
	return &app_service.GetAppStatisticV2RecordResp{Items: pbItems, Total: total}
}

func convertAppStatisticV2SelectResp(items []orm.AppStatisticV2SelectItem) *app_service.GetAppStatisticV2SelectResp {
	pbItems := make([]*app_service.AppStatisticV2SelectItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.AppStatisticV2SelectItem{
			AppId:               it.AppId,
			AppType:             it.AppType,
			Source:              it.Source,
			Module:              it.Module,
			ModuleCreatorUserId: it.ModuleCreatorUserId,
			ModuleCreatorOrgId:  it.ModuleCreatorOrgId,
		})
	}
	return &app_service.GetAppStatisticV2SelectResp{Items: pbItems}
}
