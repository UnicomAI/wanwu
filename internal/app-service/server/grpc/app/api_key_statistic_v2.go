package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) RecordAPIKeyStatisticV2(ctx context.Context, req *app_service.RecordAPIKeyStatisticV2Req) (*emptypb.Empty, error) {
	if err := s.cli.RecordAPIKeyStatisticV2(ctx, toRecordAPIKeyStatisticV2Input(req)); err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetAPIKeyStatisticV2Overview(ctx context.Context, req *app_service.GetAPIKeyStatisticV2ReadReq) (*app_service.APIKeyStatisticV2Overview, error) {
	overview, err := s.cli.GetAPIKeyStatisticV2Overview(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2Overview(overview), nil
}

func (s *Service) GetAPIKeyStatisticV2Chart(ctx context.Context, req *app_service.GetAPIKeyStatisticV2ChartReq) (*app_service.APIKeyStatisticV2Chart, error) {
	chart, err := s.cli.GetAPIKeyStatisticV2Chart(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths, req.Limit)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2Chart(chart), nil
}

func (s *Service) GetAPIKeyStatisticV2List(ctx context.Context, req *app_service.GetAPIKeyStatisticV2ListReq) (*app_service.GetAPIKeyStatisticV2ListResp, error) {
	items, total, err := s.cli.GetAPIKeyStatisticV2List(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2ListResp(items, total), nil
}

func (s *Service) GetAPIKeyStatisticV2AppList(ctx context.Context, req *app_service.GetAPIKeyStatisticV2AppListReq) (*app_service.GetAPIKeyStatisticV2AppListResp, error) {
	items, total, err := s.cli.GetAPIKeyStatisticV2AppList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths, req.ApiKeyId, req.MethodPath, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2AppListResp(items, total), nil
}

func (s *Service) GetAPIKeyStatisticV2ModelList(ctx context.Context, req *app_service.GetAPIKeyStatisticV2ModelListReq) (*app_service.GetAPIKeyStatisticV2ModelListResp, error) {
	items, total, err := s.cli.GetAPIKeyStatisticV2ModelList(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths, req.ApiKeyId, req.MethodPath, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2ModelListResp(items, total), nil
}

func (s *Service) GetAPIKeyStatisticV2Record(ctx context.Context, req *app_service.GetAPIKeyStatisticV2RecordReq) (*app_service.GetAPIKeyStatisticV2RecordResp, error) {
	items, total, err := s.cli.GetAPIKeyStatisticV2Record(ctx, req.OrgIds, req.UserIds, req.StartDate, req.EndDate, req.ApiKeyIds, req.MethodPaths, req.SortField, req.SortOrder, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(err_code.Code_AppAPIKeyRecord, err)
	}
	return convertAPIKeyStatisticV2RecordResp(items, total), nil
}

// --- internal ---

func toRecordAPIKeyStatisticV2Input(req *app_service.RecordAPIKeyStatisticV2Req) *orm.RecordAPIKeyStatisticV2Input {
	if req == nil {
		return nil
	}
	return &orm.RecordAPIKeyStatisticV2Input{
		TraceID:             req.TraceId,
		UserID:              req.UserId,
		OrgID:               req.OrgId,
		APIKeyID:            req.ApiKeyId,
		MethodPath:          req.MethodPath,
		CalledAt:            req.CalledAt,
		Source:              req.Source,
		Module:              req.Module,
		ModuleCreatorUserID: req.ModuleCreatorUserId,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgId,
		AppID:               req.AppId,
		AppType:             req.AppType,
		IsStream:            req.IsStream,
		Costs:               req.Costs,
		FirstTokenLatency:   req.FirstTokenLatency,
		IsSuccess:           req.IsSuccess,
		StatusCode:          req.StatusCode,
		FailureReason:       req.FailureReason,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
	}
}

func convertAPIKeyStatisticV2OverviewItem(item orm.StatisticOverviewItem) *app_service.APIKeyStatisticV2OverviewItem {
	return &app_service.APIKeyStatisticV2OverviewItem{
		Value: item.Value, PeriodOverPeriod: item.PeriodOverPeriod,
	}
}

func convertAPIKeyStatisticV2Overview(o *orm.APIKeyStatisticV2Overview) *app_service.APIKeyStatisticV2Overview {
	if o == nil {
		return &app_service.APIKeyStatisticV2Overview{}
	}
	return &app_service.APIKeyStatisticV2Overview{
		CallCount:              convertAPIKeyStatisticV2OverviewItem(o.CallCount),
		CallFailure:            convertAPIKeyStatisticV2OverviewItem(o.CallFailure),
		DailyAvgCallCount:      convertAPIKeyStatisticV2OverviewItem(o.DailyAvgCallCount),
		DailyAvgCallFailure:    convertAPIKeyStatisticV2OverviewItem(o.DailyAvgCallFailure),
		DailyAvgStreamCount:    convertAPIKeyStatisticV2OverviewItem(o.DailyAvgStreamCount),
		DailyAvgNonStreamCount: convertAPIKeyStatisticV2OverviewItem(o.DailyAvgNonStreamCount),
		AvgFirstTokenLatency:   convertAPIKeyStatisticV2OverviewItem(o.AvgFirstTokenLatency),
		AvgCosts:               convertAPIKeyStatisticV2OverviewItem(o.AvgCosts),
		StreamCount:            convertAPIKeyStatisticV2OverviewItem(o.StreamCount),
		NonStreamCount:         convertAPIKeyStatisticV2OverviewItem(o.NonStreamCount),
	}
}

func convertAPIKeyStatisticV2Chart(c *orm.APIKeyStatisticV2Chart) *app_service.APIKeyStatisticV2Chart {
	if c == nil {
		return &app_service.APIKeyStatisticV2Chart{}
	}
	byApi := make([]*app_service.APIKeyStatisticV2RankItem, 0, len(c.Rank.ByApi))
	for _, it := range c.Rank.ByApi {
		byApi = append(byApi, &app_service.APIKeyStatisticV2RankItem{
			ApiKeyId: it.ApiKeyId,
			OrgId:    it.OrgId, UserId: it.UserId, CallCount: it.CallCount,
		})
	}
	return &app_service.APIKeyStatisticV2Chart{
		Trend: &app_service.APIKeyStatisticV2Trend{
			ApiKeyCalls: convertStatisticChart(c.Trend.ApiKeyCalls),
			CallResult:  convertStatisticChart(c.Trend.CallResult),
		},
		Rank: &app_service.APIKeyStatisticV2Rank{ByApi: byApi},
	}
}

func convertAPIKeyStatisticV2Metrics(m orm.APIKeyStatisticV2Metrics) *app_service.APIKeyStatisticV2Metrics {
	return &app_service.APIKeyStatisticV2Metrics{
		CallCount: m.CallCount, CallFailure: m.CallFailure, FailureRate: m.FailureRate,
		AvgFirstTokenLatency: m.AvgFirstTokenLatency, AvgCosts: m.AvgCosts,
		StreamCount: m.StreamCount, NonStreamCount: m.NonStreamCount,
	}
}

func convertAPIKeyStatisticV2ListResp(items []orm.APIKeyStatisticV2ListItem, total int32) *app_service.GetAPIKeyStatisticV2ListResp {
	pbItems := make([]*app_service.APIKeyStatisticV2ListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.APIKeyStatisticV2ListItem{
			ApiKeyId: it.ApiKeyId, MethodPath: it.MethodPath, OrgId: it.OrgId, UserId: it.UserId,
			Metrics: convertAPIKeyStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAPIKeyStatisticV2ListResp{Items: pbItems, Total: total}
}

func convertAPIKeyStatisticV2AppListResp(items []orm.APIKeyStatisticV2AppListItem, total int32) *app_service.GetAPIKeyStatisticV2AppListResp {
	pbItems := make([]*app_service.APIKeyStatisticV2AppListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.APIKeyStatisticV2AppListItem{
			ApiKeyId: it.ApiKeyId, MethodPath: it.MethodPath, OrgId: it.OrgId, UserId: it.UserId,
			Source: it.Source, Module: it.Module, AppId: it.AppId, AppType: it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
			Metrics: convertAPIKeyStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAPIKeyStatisticV2AppListResp{Items: pbItems, Total: total}
}

func convertAPIKeyStatisticV2ModelListResp(items []orm.APIKeyStatisticV2ModelListItem, total int32) *app_service.GetAPIKeyStatisticV2ModelListResp {
	pbItems := make([]*app_service.APIKeyStatisticV2ModelListItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.APIKeyStatisticV2ModelListItem{
			ApiKeyId: it.ApiKeyId, MethodPath: it.MethodPath, OrgId: it.OrgId, UserId: it.UserId,
			ModelId: it.ModelId, Model: it.Model, Provider: it.Provider, ModelType: it.ModelType,
			ModelCreatorUserId: it.ModelCreatorUserId, ModelCreatorOrgId: it.ModelCreatorOrgId,
			Metrics: convertAppStatisticV2Metrics(it.Metrics),
		})
	}
	return &app_service.GetAPIKeyStatisticV2ModelListResp{Items: pbItems, Total: total}
}

func convertAPIKeyStatisticV2RecordResp(items []orm.APIKeyStatisticV2RecordItem, total int32) *app_service.GetAPIKeyStatisticV2RecordResp {
	pbItems := make([]*app_service.APIKeyStatisticV2RecordItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &app_service.APIKeyStatisticV2RecordItem{
			Id: it.Id, ApiKeyId: it.ApiKeyId, MethodPath: it.MethodPath,
			OrgId: it.OrgId, UserId: it.UserId, CalledAt: util.Time2Str(it.CallTime),
			IsSuccess: it.IsSuccess, IsStream: it.IsStream,
			FirstTokenLatency: it.FirstTokenLatency, Costs: it.Costs,
			FailureReason: it.FailureReason, RequestBody: it.RequestBody, ResponseBody: it.ResponseBody,
			Source: it.Source, Module: it.Module, AppId: it.AppId, AppType: it.AppType,
			ModuleCreatorUserId: it.ModuleCreatorUserId, ModuleCreatorOrgId: it.ModuleCreatorOrgId,
		})
	}
	return &app_service.GetAPIKeyStatisticV2RecordResp{Items: pbItems, Total: total}
}
