package orm

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

// ---------- API Key V2 读查询返回结构 ----------

type APIKeyStatisticV2Trend struct {
	ApiKeyCalls StatisticChart
	CallResult  StatisticChart
}

type APIKeyStatisticV2RankItem struct {
	ApiKeyId  string
	OrgId     string
	UserId    string
	CallCount int32
}

type APIKeyStatisticV2Rank struct {
	ByApi []APIKeyStatisticV2RankItem
}

type APIKeyStatisticV2Chart struct {
	Trend APIKeyStatisticV2Trend
	Rank  APIKeyStatisticV2Rank
}

type APIKeyStatisticV2ListItem struct {
	StatisticV2APIKeyRef
	Metrics APIKeyStatisticV2Metrics
}

type APIKeyStatisticV2AppListItem struct {
	StatisticV2APIKeyRef
	StatisticV2AppRef
	Metrics APIKeyStatisticV2Metrics
}

type APIKeyStatisticV2ModelListItem struct {
	StatisticV2APIKeyRef
	StatisticV2ModelRef
	Metrics AppStatisticV2Metrics
}

type APIKeyStatisticV2RecordItem struct {
	Id uint32
	StatisticV2APIKeyRef
	CallTime          int64
	IsSuccess         bool
	IsStream          bool
	FirstTokenLatency int64
	Costs             int64
	FailureReason     string
	RequestBody       string
	ResponseBody      string
	StatisticV2AppRef
}

func apiKeyV2BaseOpts(orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string) []sqlopt.SQLOption {
	return []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithAPIKeyIds(apiKeyIds),
		sqlopt.WithMethodPaths(methodPaths),
	}
}

func buildAPIKeyV2Metrics(s *model.StatisticApiKey) APIKeyStatisticV2Metrics {
	return APIKeyStatisticV2Metrics{
		StatisticV2CallMetrics: StatisticV2CallMetrics{
			CallCount:            s.CallCount,
			CallFailure:          s.CallFailure,
			FailureRate:          calculateFailureRate(s.CallFailure, s.CallCount),
			AvgCosts:             calculateAvg(s.Costs, int32(s.NonStreamCount-s.NonStreamFailure)),
			AvgFirstTokenLatency: calculateAvg(s.FirstTokenLatency, int32(s.StreamCount-s.StreamFailure)),
		},
		StatisticV2StreamMetrics: StatisticV2StreamMetrics{
			StreamCount:    s.StreamCount,
			NonStreamCount: s.NonStreamCount,
		},
	}
}

func apiKeyV2OverviewByDateRange(ctx context.Context, db *gorm.DB, orgIds, userIds []string, dates []string, apiKeyIds, methodPaths []string) (*APIKeyStatisticV2Overview, int, error) {
	if len(dates) == 0 {
		return &APIKeyStatisticV2Overview{}, 0, nil
	}
	startDate, endDate := dates[0], dates[len(dates)-1]
	opts := apiKeyV2BaseOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	var stat model.StatisticApiKey
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticApiKey{})
	if err := query.Select(
		"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure, " +
			"SUM(first_token_latency) as first_token_latency, SUM(costs) as costs").
		First(&stat).Error; err != nil {
		return nil, 0, fmt.Errorf("api key v2 overview [%v, %v] err: %v", startDate, endDate, err)
	}
	avgFTL := calculateAvg(stat.FirstTokenLatency, int32(stat.StreamCount-stat.StreamFailure))
	avgCosts := calculateAvg(stat.Costs, int32(stat.NonStreamCount-stat.NonStreamFailure))

	return &APIKeyStatisticV2Overview{
		CallCount:            StatisticOverviewItem{Value: float32(stat.CallCount)},
		CallFailure:          StatisticOverviewItem{Value: float32(stat.CallFailure)},
		AvgFirstTokenLatency: StatisticOverviewItem{Value: avgFTL},
		AvgCosts:             StatisticOverviewItem{Value: avgCosts},
		StreamCount:          StatisticOverviewItem{Value: float32(stat.StreamCount)},
		NonStreamCount:       StatisticOverviewItem{Value: float32(stat.NonStreamCount)},
	}, len(dates), nil
}

func (c *Client) GetAPIKeyStatisticV2Overview(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string) (*APIKeyStatisticV2Overview, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_api_key_statistic_v2_overview", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAPIKeyStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync api key v2 overview today %v err: %v", today, err)
	}
	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, toErrStatus("app_api_key_statistic_v2_overview", err.Error())
	}
	current, dayCount, err := apiKeyV2OverviewByDateRange(ctx, c.db, orgIds, userIds, currPeriod, apiKeyIds, methodPaths)
	if err != nil {
		return nil, toErrStatus("app_api_key_statistic_v2_overview", err.Error())
	}
	previous, prevDayCount, err := apiKeyV2OverviewByDateRange(ctx, c.db, orgIds, userIds, prevPeriod, apiKeyIds, methodPaths)
	if err != nil {
		return nil, toErrStatus("app_api_key_statistic_v2_overview", err.Error())
	}
	currDays := int32(dayCount)
	if currDays <= 0 {
		currDays = 1
	}
	prevDays := int32(prevDayCount)
	if prevDays <= 0 {
		prevDays = 1
	}
	current.DailyAvgCallCount = StatisticOverviewItem{Value: roundFloat2(current.CallCount.Value / float32(currDays))}
	current.DailyAvgCallFailure = StatisticOverviewItem{Value: roundFloat2(current.CallFailure.Value / float32(currDays))}
	current.DailyAvgStreamCount = StatisticOverviewItem{Value: roundFloat2(current.StreamCount.Value / float32(currDays))}
	current.DailyAvgNonStreamCount = StatisticOverviewItem{Value: roundFloat2(current.NonStreamCount.Value / float32(currDays))}
	previous.DailyAvgCallCount = StatisticOverviewItem{Value: roundFloat2(previous.CallCount.Value / float32(prevDays))}
	previous.DailyAvgCallFailure = StatisticOverviewItem{Value: roundFloat2(previous.CallFailure.Value / float32(prevDays))}
	previous.DailyAvgStreamCount = StatisticOverviewItem{Value: roundFloat2(previous.StreamCount.Value / float32(prevDays))}
	previous.DailyAvgNonStreamCount = StatisticOverviewItem{Value: roundFloat2(previous.NonStreamCount.Value / float32(prevDays))}

	current.CallCount.PeriodOverPeriod = calculatePoP(current.CallCount.Value, previous.CallCount.Value)
	current.CallFailure.PeriodOverPeriod = calculatePoP(current.CallFailure.Value, previous.CallFailure.Value)
	current.DailyAvgCallCount.PeriodOverPeriod = calculatePoP(current.DailyAvgCallCount.Value, previous.DailyAvgCallCount.Value)
	current.DailyAvgCallFailure.PeriodOverPeriod = calculatePoP(current.DailyAvgCallFailure.Value, previous.DailyAvgCallFailure.Value)
	current.DailyAvgStreamCount.PeriodOverPeriod = calculatePoP(current.DailyAvgStreamCount.Value, previous.DailyAvgStreamCount.Value)
	current.DailyAvgNonStreamCount.PeriodOverPeriod = calculatePoP(current.DailyAvgNonStreamCount.Value, previous.DailyAvgNonStreamCount.Value)
	current.AvgFirstTokenLatency.PeriodOverPeriod = calculatePoP(current.AvgFirstTokenLatency.Value, previous.AvgFirstTokenLatency.Value)
	current.AvgCosts.PeriodOverPeriod = calculatePoP(current.AvgCosts.Value, previous.AvgCosts.Value)
	current.StreamCount.PeriodOverPeriod = calculatePoP(current.StreamCount.Value, previous.StreamCount.Value)
	current.NonStreamCount.PeriodOverPeriod = calculatePoP(current.NonStreamCount.Value, previous.NonStreamCount.Value)
	return current, nil
}

func apiKeyV2Trend(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string) (*APIKeyStatisticV2Trend, error) {
	dates, err := buildDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	opts := apiKeyV2BaseOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	var stats []model.StatisticApiKey
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticApiKey{})
	if err := query.Select("date, SUM(call_count) as call_count, SUM(call_failure) as call_failure").
		Group("date").Order("date").Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("api key v2 trend err: %v", err)
	}
	dateMap := make(map[string]model.StatisticApiKey, len(stats))
	for _, s := range stats {
		dateMap[s.Date] = s
	}

	var callTotal, callSuccess, callFailure []StatisticChartLineItem
	for _, date := range dates {
		s := dateMap[date] // 缺日为零值
		callTotal = append(callTotal, StatisticChartLineItem{Key: date, Value: float32(s.CallCount)})
		callSuccess = append(callSuccess, StatisticChartLineItem{Key: date, Value: float32(s.CallCount - s.CallFailure)})
		callFailure = append(callFailure, StatisticChartLineItem{Key: date, Value: float32(s.CallFailure)})
	}
	return &APIKeyStatisticV2Trend{
		ApiKeyCalls: StatisticChart{
			Name: "api_key_statistic_calls",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_call_count_total", Items: callTotal},
				{Name: "app_statistic_call_success", Items: callSuccess},
				{Name: "app_statistic_call_failure", Items: callFailure},
			},
		},
		CallResult: StatisticChart{
			Name: "app_statistic_call_result",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_call_success", Items: callSuccess},
				{Name: "app_statistic_call_failure", Items: callFailure},
			},
		},
	}, nil
}

func (c *Client) GetAPIKeyStatisticV2Chart(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string, limit int32) (*APIKeyStatisticV2Chart, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_api_key_statistic_v2_chart", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAPIKeyStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync api key v2 chart today %v err: %v", today, err)
	}
	trend, err := apiKeyV2Trend(ctx, c.db, orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	if err != nil {
		return nil, toErrStatus("app_api_key_statistic_v2_chart", err.Error())
	}
	if limit <= 0 {
		limit = 5
	}
	opts := apiKeyV2BaseOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	var stats []model.StatisticApiKey
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApiKey{})
	// 排行按 API Key 聚合（跨 method_path 合计），保留 org/user 归属
	if err := query.Select("api_key_id, org_id, user_id, SUM(call_count) as call_count").
		Group("api_key_id, org_id, user_id").
		Order("SUM(call_count) DESC").Limit(int(limit)).Find(&stats).Error; err != nil {
		return nil, toErrStatus("app_api_key_statistic_v2_chart", err.Error())
	}
	byApi := make([]APIKeyStatisticV2RankItem, 0, len(stats))
	for _, s := range stats {
		byApi = append(byApi, APIKeyStatisticV2RankItem{
			ApiKeyId: s.APIKeyID,
			OrgId:    s.OrgID, UserId: s.UserID, CallCount: s.CallCount,
		})
	}
	return &APIKeyStatisticV2Chart{Trend: *trend, Rank: APIKeyStatisticV2Rank{ByApi: byApi}}, nil
}

func (c *Client) GetAPIKeyStatisticV2List(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string, sortExpr, sortOrder string, offset, limit int32) ([]APIKeyStatisticV2ListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAPIKeyStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync api key v2 list today %v err: %v", today, err)
	}
	opts := apiKeyV2BaseOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApiKey{})
	if err := countQuery.Select("COUNT(DISTINCT api_key_id, method_path, org_id, user_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_list", fmt.Sprintf("count err: %v", err))
	}
	var stats []model.StatisticApiKey
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApiKey{})
	query = query.Select(
		"api_key_id, method_path, org_id, user_id, SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure, " +
			"SUM(first_token_latency) as first_token_latency, SUM(costs) as costs").
		Group("api_key_id, method_path, org_id, user_id").
		Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]APIKeyStatisticV2ListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, APIKeyStatisticV2ListItem{
			StatisticV2APIKeyRef: StatisticV2APIKeyRef{
				ApiKeyId: s.APIKeyID, MethodPath: s.MethodPath, OrgId: s.OrgID, UserId: s.UserID,
			},
			Metrics: buildAPIKeyV2Metrics(&s),
		})
	}
	return items, int32(total), nil
}

func (c *Client) GetAPIKeyStatisticV2AppList(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string, apiKeyId, methodPath, sortExpr, sortOrder string, offset, limit int32) ([]APIKeyStatisticV2AppListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_app_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	opts, err := apiKeyRecordV2CreatedAtOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	if err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_app_list", err.Error())
	}
	if apiKeyId != "" {
		opts = append(opts, sqlopt.WithAPIKeyID(apiKeyId))
	}
	if methodPath != "" {
		opts = append(opts, sqlopt.WithMethodPath(methodPath))
	}

	// 有 app_id 的应用调用，或仅有 module 的板块调用（如 WGA：app_id 为空、module=wga）均进入钻取。
	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.APIKeyRecordV2{}).
		Where("(app_id != '' OR module != '')")
	if err := countQuery.Select("COUNT(DISTINCT app_id, source, module, app_type, module_creator_user_id, module_creator_org_id, org_id, user_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_app_list", fmt.Sprintf("count err: %v", err))
	}
	var rows []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.APIKeyRecordV2{}).Where("(app_id != '' OR module != '')")
	query = query.Select(
		"source, module, app_id, app_type, module_creator_user_id, module_creator_org_id, org_id, user_id, " +
			"COUNT(*) as call_count, SUM(CASE WHEN status_code != 200 THEN 1 ELSE 0 END) as call_failure, " +
			"SUM(CASE WHEN is_stream = 1 THEN 1 ELSE 0 END) as stream_count, " +
			"SUM(CASE WHEN is_stream = 0 THEN 1 ELSE 0 END) as non_stream_count, " +
			"SUM(CASE WHEN is_stream = 1 AND status_code != 200 THEN 1 ELSE 0 END) as stream_failure, " +
			"SUM(CASE WHEN is_stream = 0 AND status_code != 200 THEN 1 ELSE 0 END) as non_stream_failure, " +
			// 平均耗时口径与日聚合表对齐：分子只累加成功调用，失败记 0（明细表保留真实值仅供单次排查）。
			"SUM(CASE WHEN status_code = 200 THEN first_token_latency ELSE 0 END) as first_token_latency, " +
			"SUM(CASE WHEN status_code = 200 THEN costs ELSE 0 END) as costs").
		Group("source, module, app_id, app_type, module_creator_user_id, module_creator_org_id, org_id, user_id").
		Order(buildV2OrderClause(sortExpr, sortOrder, "COUNT(*)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_app_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]APIKeyStatisticV2AppListItem, 0, len(rows))
	for _, r := range rows {
		m := model.StatisticApiKey{
			CallCount: r.CallCount, CallFailure: r.CallFailure,
			StreamCount: r.StreamCount, NonStreamCount: r.NonStreamCount,
			StreamFailure: r.StreamFailure, NonStreamFailure: r.NonStreamFailure,
			FirstTokenLatency: r.FirstTokenLatency, Costs: r.Costs,
		}
		items = append(items, APIKeyStatisticV2AppListItem{
			StatisticV2APIKeyRef: StatisticV2APIKeyRef{
				ApiKeyId: apiKeyId, MethodPath: methodPath, OrgId: r.OrgID, UserId: r.UserID,
			},
			StatisticV2AppRef: StatisticV2AppRef{
				Source: r.Source, Module: r.Module, AppId: r.AppID, AppType: r.AppType,
				ModuleCreatorUserId: r.ModuleCreatorUserID, ModuleCreatorOrgId: r.ModuleCreatorOrgID,
			},
			Metrics: buildAPIKeyV2Metrics(&m),
		})
	}
	return items, int32(total), nil
}

// GetAPIKeyStatisticV2ModelList API Key 维度钻取模型列表。
// 走 StatisticModel 聚合表（与 App V2 模型钻取一致）：当日数据由 syncStatisticModelStats 从 Redis 落库。
// StatisticModel 存明文 api_key（非 api_key_id），故先用 OpenApiKey 解析后再过滤。
func (c *Client) GetAPIKeyStatisticV2ModelList(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string, apiKeyId, methodPath, sortExpr, sortOrder string, offset, limit int32) ([]APIKeyStatisticV2ModelListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_model_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	if apiKeyId == "" || methodPath == "" {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_model_list", "apiKeyId and methodPath required")
	}
	if len(apiKeyIds) > 0 && !slices.Contains(apiKeyIds, apiKeyId) {
		return []APIKeyStatisticV2ModelListItem{}, 0, nil
	}

	var openKey model.OpenApiKey
	if err := sqlopt.WithID(util.MustU32(apiKeyId)).Apply(c.db.WithContext(ctx)).First(&openKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []APIKeyStatisticV2ModelListItem{}, 0, nil
		}
		return nil, 0, toErrStatus("app_api_key_statistic_v2_model_list", fmt.Sprintf("get api key err: %v", err))
	}
	apiKey := openKey.Key

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for api key v2 model list today %v err: %v", today, err)
	}

	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithAppKey(apiKey),
		sqlopt.WithMethodPath(methodPath),
		sqlopt.WithMethodPaths(methodPaths),
		sqlopt.WithNonEmptyModelID(),
	}

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := countQuery.Select("COUNT(DISTINCT model_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_model_list", fmt.Sprintf("count err: %v", err))
	}
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	query = query.Select(
		"model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, ANY_VALUE(model_creator_org_id) as model_creator_org_id, " +
			"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens, " +
			"SUM(costs) as costs, SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("model_id").Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_model_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]APIKeyStatisticV2ModelListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, APIKeyStatisticV2ModelListItem{
			StatisticV2APIKeyRef: StatisticV2APIKeyRef{
				ApiKeyId: apiKeyId, MethodPath: methodPath, OrgId: openKey.OrgID, UserId: openKey.UserID,
			},
			StatisticV2ModelRef: StatisticV2ModelRef{
				ModelId: s.ModelID, Model: s.Model, Provider: s.Provider, ModelType: s.ModelType,
				ModelCreatorUserId: s.ModelCreatorUserID, ModelCreatorOrgId: s.ModelCreatorOrgID,
			},
			Metrics: buildAppV2MetricsFromRecordAgg(
				s.CallCount, s.CallFailure, s.StreamCount, s.NonStreamCount, s.StreamFailure, s.NonStreamFailure,
				s.PromptTokens, s.CompletionTokens, s.TotalTokens, s.FirstTokenLatency, s.Costs,
			),
		})
	}
	return items, int32(total), nil
}

func apiKeyRecordV2CreatedAtOpts(orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string) ([]sqlopt.SQLOption, error) {
	createdAtOpts, err := buildRecordCreatedAtOpts(startDate, endDate)
	if err != nil {
		return nil, err
	}
	return append(createdAtOpts,
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithAPIKeyIds(apiKeyIds),
		sqlopt.WithMethodPaths(methodPaths),
	), nil
}

func (c *Client) GetAPIKeyStatisticV2Record(ctx context.Context, orgIds, userIds []string, startDate, endDate string, apiKeyIds, methodPaths []string, sortExpr, sortOrder string, offset, limit int32) ([]APIKeyStatisticV2RecordItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_record", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	opts, err := apiKeyRecordV2CreatedAtOpts(orgIds, userIds, startDate, endDate, apiKeyIds, methodPaths)
	if err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_record", err.Error())
	}
	var total int64
	if err := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.APIKeyRecordV2{}).Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_record", fmt.Sprintf("count err: %v", err))
	}
	var records []model.APIKeyRecordV2
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.APIKeyRecordV2{})
	query = query.Order(buildV2OrderClause(sortExpr, sortOrder, "created_at"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, 0, toErrStatus("app_api_key_statistic_v2_record", fmt.Sprintf("list err: %v", err))
	}
	items := make([]APIKeyStatisticV2RecordItem, 0, len(records))
	for _, r := range records {
		items = append(items, APIKeyStatisticV2RecordItem{
			Id: uint32(r.ID),
			StatisticV2APIKeyRef: StatisticV2APIKeyRef{
				ApiKeyId: r.APIKeyID, MethodPath: r.MethodPath, OrgId: r.OrgID, UserId: r.UserID,
			},
			CallTime: r.CreatedAt, IsSuccess: statistic.IsSuccess(r.StatusCode), IsStream: r.IsStream,
			FirstTokenLatency: r.FirstTokenLatency, Costs: r.Costs,
			FailureReason: r.FailureReason, RequestBody: r.RequestBody, ResponseBody: r.ResponseBody,
			StatisticV2AppRef: StatisticV2AppRef{
				Source: r.Source, Module: r.Module, AppId: r.AppID, AppType: r.AppType,
				ModuleCreatorUserId: r.ModuleCreatorUserID, ModuleCreatorOrgId: r.ModuleCreatorOrgID,
			},
		})
	}
	return items, int32(total), nil
}
