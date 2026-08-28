package orm

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

// ---------- App V2 读查询返回结构 ----------

type AppStatisticV2Trend struct {
	CallResult StatisticChart
	CallTrend  StatisticChart
}

type AppStatisticV2RankItem struct {
	AppId               string
	AppType             string
	ModuleCreatorUserId string
	ModuleCreatorOrgId  string
	CallCount           int32
}

type AppStatisticV2Rank struct {
	ByAgent           []AppStatisticV2RankItem
	ByWorkflow        []AppStatisticV2RankItem
	ByChatflow        []AppStatisticV2RankItem
	ByRag             []AppStatisticV2RankItem
	ByDigitalEmployee []AppStatisticV2RankItem
}

type AppStatisticV2Chart struct {
	Trend AppStatisticV2Trend
	Rank  AppStatisticV2Rank
}

type AppStatisticV2ListItem struct {
	StatisticV2AppRef
	Metrics AppStatisticV2Metrics
}

type AppStatisticV2UserListItem struct {
	StatisticV2AppRef
	StatisticV2UserRef
	Metrics AppStatisticV2Metrics
}

type AppStatisticV2ModelListItem struct {
	StatisticV2AppRef
	StatisticV2ModelRef
	Metrics AppStatisticV2Metrics
}

// AppStatisticV2RecordItem 应用会话调用明细列表行（一次会话一条，含详情字段）
type AppStatisticV2RecordItem struct {
	Id      uint32
	TraceId string
	StatisticV2AppRef
	StatisticV2UserRef
	CallTime          int64
	IsSuccess         bool
	StatusCode        int64
	Costs             int64
	FirstTokenLatency int64
	FailureReason     string
	RequestBody       string
	ResponseBody      string
	Question          string
	Answer            string
}

func appStatisticV2ScopeOptions(viewScope string, orgIds, userIds []string) []sqlopt.SQLOption {
	switch viewScope {
	case statistic.ViewScopePublished:
		return []sqlopt.SQLOption{
			sqlopt.WithModuleCreatorOrgIDs(orgIds),
			sqlopt.WithModuleCreatorUserIDs(userIds),
		}
	case statistic.ViewScopeUsed:
		return []sqlopt.SQLOption{
			sqlopt.WithOrgIDs(orgIds),
			sqlopt.WithUserIDs(userIds),
		}
	default:
		return []sqlopt.SQLOption{
			sqlopt.WithOrgIDs(orgIds),
			sqlopt.WithUserIDs(userIds),
		}
	}
}

func appStatisticV2BaseOpts(orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope string) []sqlopt.SQLOption {
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModule(module),
		sqlopt.WithAppIDsForStatistic(apps),
		sqlopt.WithStatisticAppIDFilter(module), // wga/model/skill/knowledge/prompt 允许空 appId；module 空=全部时仍保留板块级行
	}
	opts = append(opts, appStatisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	return opts
}

// statisticDrilldownRowOpts 定位主表一行：source + appId + 应用作者。
func statisticDrilldownRowOpts(source, appId, moduleCreatorUserId, moduleCreatorOrgId string) []sqlopt.SQLOption {
	opts := []sqlopt.SQLOption{
		sqlopt.WithSource(source),
		sqlopt.WithStatisticAppID(appId),
	}
	if moduleCreatorUserId != "" {
		opts = append(opts, sqlopt.WithModuleCreatorUserIDs([]string{moduleCreatorUserId}))
	}
	if moduleCreatorOrgId != "" {
		opts = append(opts, sqlopt.WithModuleCreatorOrgIDs([]string{moduleCreatorOrgId}))
	}
	return opts
}

type appV2TokenAgg struct {
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

type appV2AggKey struct {
	Source              string
	Module              string
	AppId               string
	AppType             string
	ModuleCreatorUserId string
	ModuleCreatorOrgId  string
}

type appV2UserKey struct {
	UserId string
	OrgId  string
}

func appV2AggKeyFromApp(s *model.StatisticApp) appV2AggKey {
	return appV2AggKey{
		Source:              s.Source,
		Module:              s.Module,
		AppId:               s.AppID,
		AppType:             s.AppType,
		ModuleCreatorUserId: s.ModuleCreatorUserID,
		ModuleCreatorOrgId:  s.ModuleCreatorOrgID,
	}
}

func appV2QueryTokenMapByAppKey(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source string) (map[appV2AggKey]appV2TokenAgg, error) {
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.Select(
		"source, module, app_id, app_type, module_creator_user_id, module_creator_org_id, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens").
		Group("source, module, app_id, app_type, module_creator_user_id, module_creator_org_id").
		Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("app v2 token map by app key err: %v", err)
	}
	out := make(map[appV2AggKey]appV2TokenAgg, len(stats))
	for _, s := range stats {
		key := appV2AggKey{
			Source:              s.Source,
			Module:              s.Module,
			AppId:               s.AppID,
			AppType:             s.AppType,
			ModuleCreatorUserId: s.ModuleCreatorUserID,
			ModuleCreatorOrgId:  s.ModuleCreatorOrgID,
		}
		out[key] = appV2TokenAgg{
			PromptTokens:     s.PromptTokens,
			CompletionTokens: s.CompletionTokens,
			TotalTokens:      s.TotalTokens,
		}
	}
	return out, nil
}

func appV2QueryTokenMapByUser(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source, appId, moduleCreatorUserId, moduleCreatorOrgId string) (map[appV2UserKey]appV2TokenAgg, error) {
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	opts = append(opts, statisticDrilldownRowOpts(source, appId, moduleCreatorUserId, moduleCreatorOrgId)...)
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.Select(
		"user_id, org_id, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, SUM(total_tokens) as total_tokens").
		Group("user_id, org_id").
		Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("app v2 token map by user err: %v", err)
	}
	out := make(map[appV2UserKey]appV2TokenAgg, len(stats))
	for _, s := range stats {
		out[appV2UserKey{UserId: s.UserID, OrgId: s.OrgID}] = appV2TokenAgg{
			PromptTokens:     s.PromptTokens,
			CompletionTokens: s.CompletionTokens,
			TotalTokens:      s.TotalTokens,
		}
	}
	return out, nil
}

func buildAppV2Metrics(s *model.StatisticApp, tokens appV2TokenAgg) AppStatisticV2Metrics {
	return AppStatisticV2Metrics{
		StatisticV2TokenMetrics: StatisticV2TokenMetrics{
			TotalTokens:      tokens.TotalTokens,
			PromptTokens:     tokens.PromptTokens,
			CompletionTokens: tokens.CompletionTokens,
		},
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

func buildAppV2MetricsFromRecordAgg(callCount, callFailure, streamCount, nonStreamCount, streamFailure, nonStreamFailure int32, promptTokens, completionTokens, totalTokens, firstTokenLatency, costs int64) AppStatisticV2Metrics {
	return AppStatisticV2Metrics{
		StatisticV2TokenMetrics: StatisticV2TokenMetrics{
			TotalTokens:      totalTokens,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
		},
		StatisticV2CallMetrics: StatisticV2CallMetrics{
			CallCount:            callCount,
			CallFailure:          callFailure,
			FailureRate:          calculateFailureRate(callFailure, callCount),
			AvgCosts:             calculateAvg(costs, int32(nonStreamCount-nonStreamFailure)),
			AvgFirstTokenLatency: calculateAvg(firstTokenLatency, int32(streamCount-streamFailure)),
		},
		StatisticV2StreamMetrics: StatisticV2StreamMetrics{
			StreamCount:    streamCount,
			NonStreamCount: nonStreamCount,
		},
	}
}

func appV2OverviewByDateRange(ctx context.Context, db *gorm.DB, orgIds, userIds []string, dates []string, module string, apps []string, viewScope, source string) (*AppStatisticV2Overview, int, error) {
	if len(dates) == 0 {
		return &AppStatisticV2Overview{}, 0, nil
	}
	startDate, endDate := dates[0], dates[len(dates)-1]
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	var stat model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticApp{})
	if err := query.Select(
		"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure, " +
			"SUM(first_token_latency) as first_token_latency, SUM(costs) as costs").
		First(&stat).Error; err != nil {
		return nil, 0, fmt.Errorf("app v2 overview [%v, %v] err: %v", startDate, endDate, err)
	}
	avgCosts := calculateAvg(stat.Costs, int32(stat.NonStreamCount-stat.NonStreamFailure))
	avgFTL := calculateAvg(stat.FirstTokenLatency, int32(stat.StreamCount-stat.StreamFailure))
	return &AppStatisticV2Overview{
		CallCount:            StatisticOverviewItem{Value: float32(stat.CallCount)},
		CallFailure:          StatisticOverviewItem{Value: float32(stat.CallFailure)},
		AvgFirstTokenLatency: StatisticOverviewItem{Value: avgFTL},
		AvgCosts:             StatisticOverviewItem{Value: avgCosts},
		StreamCount:          StatisticOverviewItem{Value: float32(stat.StreamCount)},
		NonStreamCount:       StatisticOverviewItem{Value: float32(stat.NonStreamCount)},
	}, len(dates), nil
}

func (c *Client) GetAppStatisticV2Overview(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source string) (*AppStatisticV2Overview, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_statistic_v2_overview", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAppStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync app statistic v2 for overview today %v err: %v", today, err)
	}
	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_overview", err.Error())
	}
	current, dayCount, err := appV2OverviewByDateRange(ctx, c.db, orgIds, userIds, currPeriod, module, apps, viewScope, source)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_overview", err.Error())
	}
	previous, prevDayCount, err := appV2OverviewByDateRange(ctx, c.db, orgIds, userIds, prevPeriod, module, apps, viewScope, source)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_overview", err.Error())
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

func appV2Trend(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source string) (*AppStatisticV2Trend, error) {
	dates, err := buildDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	var stats []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticApp{})
	if err := query.Select("date, SUM(call_count) as call_count, SUM(call_failure) as call_failure, source").
		Group("date, source").Order("date").Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("app v2 trend err: %v", err)
	}

	type dayAgg struct {
		total, success, failure float32
		web, openapi, webURL    float32
	}
	dayMap := make(map[string]*dayAgg)
	for _, s := range stats {
		if dayMap[s.Date] == nil {
			dayMap[s.Date] = &dayAgg{}
		}
		d := dayMap[s.Date]
		d.total += float32(s.CallCount)
		d.success += float32(s.CallCount - s.CallFailure)
		d.failure += float32(s.CallFailure)
		switch s.Source {
		case constant.BizSourceWeb:
			d.web += float32(s.CallCount)
		case constant.BizSourceOpenAPI:
			d.openapi += float32(s.CallCount)
		case constant.BizSourceWebUrl:
			d.webURL += float32(s.CallCount)
		}
	}

	var callSuccess, callFailure []StatisticChartLineItem
	var callTotal, sourceWeb, sourceOpenAPI, sourceWebURL []StatisticChartLineItem
	for _, date := range dates {
		d := dayMap[date]
		if d == nil {
			d = &dayAgg{}
		}
		callSuccess = append(callSuccess, StatisticChartLineItem{Key: date, Value: d.success})
		callFailure = append(callFailure, StatisticChartLineItem{Key: date, Value: d.failure})
		callTotal = append(callTotal, StatisticChartLineItem{Key: date, Value: d.total})
		sourceWeb = append(sourceWeb, StatisticChartLineItem{Key: date, Value: d.web})
		sourceOpenAPI = append(sourceOpenAPI, StatisticChartLineItem{Key: date, Value: d.openapi})
		sourceWebURL = append(sourceWebURL, StatisticChartLineItem{Key: date, Value: d.webURL})
	}
	return &AppStatisticV2Trend{
		CallResult: StatisticChart{
			Name: "app_statistic_call_result",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_call_success", Items: callSuccess},
				{Name: "app_statistic_call_failure", Items: callFailure},
			},
		},
		CallTrend: StatisticChart{
			Name: "app_statistic_call_trend",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_call_count_total", Items: callTotal},
				{Name: "app_statistic_source_web", Items: sourceWeb},
				{Name: "app_statistic_source_openapi", Items: sourceOpenAPI},
				{Name: "app_statistic_source_weburl", Items: sourceWebURL},
			},
		},
	}, nil
}

func appV2RankByType(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, appType, source string, limit int32) ([]AppStatisticV2RankItem, error) {
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	opts = append(opts, sqlopt.WithAppType(appType))
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	var stats []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticApp{})
	if limit <= 0 {
		limit = 5
	}
	if err := query.Select("app_id, app_type, module_creator_user_id, module_creator_org_id, SUM(call_count) as call_count").
		Group("app_id, app_type, module_creator_user_id, module_creator_org_id").
		Order("SUM(call_count) DESC").Limit(int(limit)).Find(&stats).Error; err != nil {
		return nil, err
	}
	items := make([]AppStatisticV2RankItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, AppStatisticV2RankItem{
			AppId:               s.AppID,
			AppType:             s.AppType,
			ModuleCreatorUserId: s.ModuleCreatorUserID,
			ModuleCreatorOrgId:  s.ModuleCreatorOrgID,
			CallCount:           s.CallCount,
		})
	}
	return items, nil
}

func (c *Client) GetAppStatisticV2Chart(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source string, limit int32) (*AppStatisticV2Chart, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAppStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync app statistic v2 today %v err: %v", today, err)
	}
	trend, err := appV2Trend(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, source)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", err.Error())
	}
	byAgent, err := appV2RankByType(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, constant.AppTypeAgent, source, limit)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("rank by agent err: %v", err))
	}
	byWorkflow, err := appV2RankByType(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, constant.AppTypeWorkflow, source, limit)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("rank by workflow err: %v", err))
	}
	byChatflow, err := appV2RankByType(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, constant.AppTypeChatflow, source, limit)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("rank by chatflow err: %v", err))
	}
	byRag, err := appV2RankByType(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, constant.AppTypeRag, source, limit)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("rank by rag err: %v", err))
	}
	byDigitalEmployee, err := appV2RankByType(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, constant.AppTypeDigitalEmployee, source, limit)
	if err != nil {
		return nil, toErrStatus("app_statistic_v2_chart", fmt.Sprintf("rank by digital employee err: %v", err))
	}
	return &AppStatisticV2Chart{
		Trend: *trend,
		Rank: AppStatisticV2Rank{
			ByAgent: byAgent, ByWorkflow: byWorkflow, ByChatflow: byChatflow, ByRag: byRag, ByDigitalEmployee: byDigitalEmployee,
		},
	}, nil
}

func (c *Client) GetAppStatisticV2List(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source, sortExpr, sortOrder string, offset, limit int32) ([]AppStatisticV2ListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_statistic_v2_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAppStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync app statistic v2 today %v err: %v", today, err)
	}
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for app v2 list today %v err: %v", today, err)
	}
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApp{})
	if err := countQuery.Select("COUNT(DISTINCT app_id, source, module, app_type, module_creator_user_id, module_creator_org_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_list", fmt.Sprintf("count err: %v", err))
	}
	var stats []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApp{})
	query = query.Select(
		"source, module, app_id, app_type, module_creator_user_id, module_creator_org_id, " +
			"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(costs) as costs, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("source, module, app_id, app_type, module_creator_user_id, module_creator_org_id").
		Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_list", fmt.Sprintf("list err: %v", err))
	}
	tokenMap, err := appV2QueryTokenMapByAppKey(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, source)
	if err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_list", err.Error())
	}
	items := make([]AppStatisticV2ListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, AppStatisticV2ListItem{
			StatisticV2AppRef: StatisticV2AppRef{
				Source: s.Source, Module: s.Module, AppId: s.AppID, AppType: s.AppType,
				ModuleCreatorUserId: s.ModuleCreatorUserID, ModuleCreatorOrgId: s.ModuleCreatorOrgID,
			},
			Metrics: buildAppV2Metrics(&s, tokenMap[appV2AggKeyFromApp(&s)]),
		})
	}
	return items, int32(total), nil
}

func (c *Client) GetAppStatisticV2UserList(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source, appId, moduleCreatorUserId, moduleCreatorOrgId, sortExpr, sortOrder string, offset, limit int32) ([]AppStatisticV2UserListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_statistic_v2_user_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncAppStatisticV2Stats(ctx, today, c.db); err != nil {
		log.Errorf("sync app statistic v2 today %v err: %v", today, err)
	}
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for app v2 user list today %v err: %v", today, err)
	}
	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	opts = append(opts, statisticDrilldownRowOpts(source, appId, moduleCreatorUserId, moduleCreatorOrgId)...)
	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApp{})
	if err := countQuery.Select("COUNT(DISTINCT user_id, org_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_user_list", fmt.Sprintf("count err: %v", err))
	}
	var stats []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApp{})
	query = query.Select(
		"ANY_VALUE(source) as source, ANY_VALUE(module) as module, ANY_VALUE(app_id) as app_id, ANY_VALUE(app_type) as app_type, " +
			"ANY_VALUE(module_creator_user_id) as module_creator_user_id, ANY_VALUE(module_creator_org_id) as module_creator_org_id, " +
			"user_id, org_id, SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(costs) as costs, SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("user_id, org_id").Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_user_list", fmt.Sprintf("list err: %v", err))
	}
	tokenMap, err := appV2QueryTokenMapByUser(ctx, c.db, orgIds, userIds, startDate, endDate, module, apps, viewScope, source, appId, moduleCreatorUserId, moduleCreatorOrgId)
	if err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_user_list", err.Error())
	}
	items := make([]AppStatisticV2UserListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, AppStatisticV2UserListItem{
			StatisticV2AppRef: StatisticV2AppRef{
				Source: s.Source, Module: s.Module, AppId: s.AppID, AppType: s.AppType,
				ModuleCreatorUserId: s.ModuleCreatorUserID, ModuleCreatorOrgId: s.ModuleCreatorOrgID,
			},
			StatisticV2UserRef: StatisticV2UserRef{UserId: s.UserID, OrgId: s.OrgID},
			Metrics:            buildAppV2Metrics(&s, tokenMap[appV2UserKey{UserId: s.UserID, OrgId: s.OrgID}]),
		})
	}
	return items, int32(total), nil
}

// GetAppStatisticV2ModelList App 维度钻取模型列表。
// 走 StatisticModel 聚合表（与 v2 其它读路径一致）：当日数据由 syncStatisticModelStats 从 Redis 落库，
// 这里直接 SUM 聚合字段，不再在 ModelRecordV2 明细行上做 CASE WHEN。
func (c *Client) GetAppStatisticV2ModelList(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, source, appId, moduleCreatorUserId, moduleCreatorOrgId, sortExpr, sortOrder string, offset, limit int32) ([]AppStatisticV2ModelListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_statistic_v2_model_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for app v2 model list today %v err: %v", today, err)
	}

	opts := appStatisticV2BaseOpts(orgIds, userIds, startDate, endDate, module, apps, viewScope)
	opts = append(opts, statisticDrilldownRowOpts(source, appId, moduleCreatorUserId, moduleCreatorOrgId)...)
	opts = append(opts, sqlopt.WithNonEmptyModelID())

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := countQuery.Select("COUNT(DISTINCT model_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_model_list", fmt.Sprintf("count err: %v", err))
	}
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	query = query.Select(
		"ANY_VALUE(source) as source, ANY_VALUE(module) as module, ANY_VALUE(app_id) as app_id, ANY_VALUE(app_type) as app_type, " +
			"ANY_VALUE(module_creator_user_id) as module_creator_user_id, ANY_VALUE(module_creator_org_id) as module_creator_org_id, " +
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
		return nil, 0, toErrStatus("app_statistic_v2_model_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]AppStatisticV2ModelListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, AppStatisticV2ModelListItem{
			StatisticV2AppRef: StatisticV2AppRef{
				Source: s.Source, Module: s.Module, AppId: s.AppID, AppType: s.AppType,
				ModuleCreatorUserId: s.ModuleCreatorUserID, ModuleCreatorOrgId: s.ModuleCreatorOrgID,
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

func (c *Client) GetAppStatisticV2Record(ctx context.Context, orgIds, userIds []string, startDate, endDate, module string, apps []string, viewScope, appId, source, sortExpr, sortOrder string, offset, limit int32) ([]AppStatisticV2RecordItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_statistic_v2_record", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	createdAtOpts, err := buildRecordCreatedAtOpts(startDate, endDate)
	if err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_record", err.Error())
	}

	opts := append(createdAtOpts, sqlopt.WithModule(module), sqlopt.WithStatisticAppIDFilter(module))
	if len(apps) > 0 {
		opts = append(opts, sqlopt.WithAppIDsForStatistic(apps))
	}
	if appId != "" {
		opts = append(opts, sqlopt.WithAppID(appId))
	}
	if source != "" {
		opts = append(opts, sqlopt.WithSource(source))
	}
	opts = append(opts, appStatisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var total int64
	if err := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.AppRecordV2{}).Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_record", fmt.Sprintf("count err: %v", err))
	}
	var records []model.AppRecordV2
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.AppRecordV2{})
	query = query.Order(buildV2OrderClause(sortExpr, sortOrder, "created_at"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, 0, toErrStatus("app_statistic_v2_record", fmt.Sprintf("list err: %v", err))
	}
	items := make([]AppStatisticV2RecordItem, 0, len(records))
	for _, r := range records {
		items = append(items, AppStatisticV2RecordItem{
			Id:      uint32(r.ID),
			TraceId: r.TraceID,
			StatisticV2AppRef: StatisticV2AppRef{
				Source: r.Source, Module: r.Module, AppId: r.AppID, AppType: r.AppType,
				ModuleCreatorUserId: r.ModuleCreatorUserID, ModuleCreatorOrgId: r.ModuleCreatorOrgID,
			},
			StatisticV2UserRef: StatisticV2UserRef{UserId: r.UserID, OrgId: r.OrgID},
			CallTime:           r.CreatedAt, IsSuccess: statistic.IsSuccess(r.StatusCode), StatusCode: r.StatusCode,
			Costs: r.Costs, FirstTokenLatency: r.FirstTokenLatency,
			FailureReason: r.FailureReason, RequestBody: r.RequestBody, ResponseBody: r.ResponseBody,
			Question: r.Question, Answer: r.Answer,
		})
	}
	return items, int32(total), nil
}

// AppStatisticV2SelectItem 应用下拉选项（聚合表 DISTINCT app_id）
type AppStatisticV2SelectItem = StatisticV2AppRef

// ListAppStatisticV2Select 从聚合表查询出现过的应用 ID（按 viewScope + module 过滤，不限日期）。
// module=workflow（含前端误传的 chatflow）时同时返回 app_type=workflow|chatflow；
// agent / rag 仅对应 app_type；knowledge 不限制 app_type（历史行可能为空）。
// 下拉不触发当日 Redis→DB sync，依赖 cron / 其它读路径落库。
func (c *Client) ListAppStatisticV2Select(ctx context.Context, orgIds, userIds []string, module, viewScope string) ([]AppStatisticV2SelectItem, *errs.Status) {
	queryModule := module
	if module == constant.AppTypeChatflow {
		queryModule = constant.BizModuleAppWorkflow
	}

	opts := []sqlopt.SQLOption{
		sqlopt.WithModule(queryModule),
		sqlopt.WithStatisticAppIDFilter(queryModule),
	}
	switch queryModule {
	case constant.BizModuleAppWorkflow:
		opts = append(opts, sqlopt.WithAppTypes([]string{constant.AppTypeWorkflow, constant.AppTypeChatflow}))
	case constant.BizModuleAppAgent:
		opts = append(opts, sqlopt.WithAppType(constant.AppTypeAgent))
	case constant.BizModuleAppRag:
		opts = append(opts, sqlopt.WithAppType(constant.AppTypeRag))
	}
	opts = append(opts, appStatisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var stats []model.StatisticApp
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticApp{})
	if err := query.
		Select("app_id, ANY_VALUE(app_type) as app_type, ANY_VALUE(source) as source, " +
			"ANY_VALUE(module) as module, " +
			"ANY_VALUE(module_creator_user_id) as module_creator_user_id, " +
			"ANY_VALUE(module_creator_org_id) as module_creator_org_id").
		Group("app_id").
		Order("app_id ASC").
		Find(&stats).Error; err != nil {
		return nil, toErrStatus("app_statistic_v2_select", fmt.Sprintf("query err: %v", err))
	}

	items := make([]AppStatisticV2SelectItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, AppStatisticV2SelectItem{
			Source: s.Source, Module: s.Module, AppId: s.AppID, AppType: s.AppType,
			ModuleCreatorUserId: s.ModuleCreatorUserID, ModuleCreatorOrgId: s.ModuleCreatorOrgID,
		})
	}
	return items, nil
}
