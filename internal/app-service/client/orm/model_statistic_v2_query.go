package orm

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

// ---------- V2 读查询返回结构（gRPC 层据此转换为 proto 消息） ----------

// ModelStatisticV2Overview V2 概览（含日均 tokens）
type ModelStatisticV2Overview struct {
	TotalTokens              StatisticOverviewItem
	PromptTokens             StatisticOverviewItem
	CompletionTokens         StatisticOverviewItem
	DailyAvgTotalTokens      StatisticOverviewItem
	DailyAvgPromptTokens     StatisticOverviewItem
	DailyAvgCompletionTokens StatisticOverviewItem
	CallCount                StatisticOverviewItem
	CallFailure              StatisticOverviewItem
	AvgCosts                 StatisticOverviewItem
	AvgFirstTokenLatency     StatisticOverviewItem
}

// ModelStatisticV2Trend V2 趋势
type ModelStatisticV2Trend struct {
	TokensUsage StatisticChart
	ModelCalls  StatisticChart
}

// ModelStatisticV2Rank V2 排行
type ModelStatisticV2Rank struct {
	ByModel []ModelStatisticV2RankByModelItem
	ByUser  []ModelStatisticV2RankByUserItem
	ByOrg   []ModelStatisticV2RankByOrgItem
}

// ModelStatisticV2Chart V2 趋势 + 排行
type ModelStatisticV2Chart struct {
	Trend ModelStatisticV2Trend
	Rank  ModelStatisticV2Rank
}

type ModelStatisticV2RankByModelItem struct {
	ModelId            string
	Model              string
	Provider           string
	ModelType          string
	ModelCreatorUserId string
	ModelCreatorOrgId  string
	TotalTokens        int64
}

type ModelStatisticV2RankByUserItem struct {
	UserId      string
	OrgId       string
	TotalTokens int64
}

type ModelStatisticV2RankByOrgItem struct {
	OrgId       string
	CallCount   int32
	TotalTokens int64
}

// ModelStatisticV2ListItem 主表行
type ModelStatisticV2ListItem struct {
	StatisticV2ModelRef
	Metrics ModelStatisticV2Metrics
}

// ModelStatisticV2UserListItem 用户钻取行
type ModelStatisticV2UserListItem struct {
	StatisticV2ModelRef
	StatisticV2UserRef
	Metrics ModelStatisticV2Metrics
}

// ModelStatisticV2AppListItem 应用钻取行
type ModelStatisticV2AppListItem struct {
	StatisticV2ModelRef
	StatisticV2AppRef
	Metrics ModelStatisticV2Metrics
}

// ModelStatisticV2RecordItem 明细列表行（含详情字段）
type ModelStatisticV2RecordItem struct {
	Id                uint32
	TraceId           string
	CreatedAt         int64
	IsSuccess         bool
	StatusCode        int64
	FailureReason     string
	RequestBody       string
	ResponseBody      string
	FinishReason      string
	FirstTokenLatency int64
	Costs             int64
	StatisticV2ModelRef
	StatisticV2AppRef
	StatisticV2UserRef
	StatisticV2TokenMetrics
}

// ---------- 公共 viewScope 过滤 ----------

// statisticV2ScopeOptions 将 filter 中的 org/user 映射到正确列：
// used → 调用人 user_id / org_id；published → 模型发布者 model_creator_user_id / model_creator_org_id。
func statisticV2ScopeOptions(viewScope string, orgIds, userIds []string) []sqlopt.SQLOption {
	switch viewScope {
	case statistic.ViewScopePublished:
		return []sqlopt.SQLOption{
			sqlopt.WithModelCreatorOrgIDs(orgIds),
			sqlopt.WithModelCreatorUserIDs(userIds),
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

// ---------- 1. Overview ----------

// GetModelStatisticV2Overview 10 指标卡（含日均 tokens）
func (c *Client) GetModelStatisticV2Overview(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string) (*ModelStatisticV2Overview, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_model_statistic_v2_overview", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for today %v err: %v", today, err)
	}

	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_v2_overview", err.Error())
	}

	current, dayCount, err := statisticV2OverviewByDateRange(ctx, c.db, orgIds, userIds, currPeriod, modelIds, modelType, viewScope)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_v2_overview", err.Error())
	}
	previous, prevDayCount, err := statisticV2OverviewByDateRange(ctx, c.db, orgIds, userIds, prevPeriod, modelIds, modelType, viewScope)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_v2_overview", err.Error())
	}

	// 日均 tokens
	currDayCount := int32(dayCount)
	if currDayCount <= 0 {
		currDayCount = 1
	}
	prevDayCountI := int32(prevDayCount)
	if prevDayCountI <= 0 {
		prevDayCountI = 1
	}
	current.DailyAvgTotalTokens = StatisticOverviewItem{Value: roundFloat2(current.TotalTokens.Value / float32(currDayCount))}
	current.DailyAvgPromptTokens = StatisticOverviewItem{Value: roundFloat2(current.PromptTokens.Value / float32(currDayCount))}
	current.DailyAvgCompletionTokens = StatisticOverviewItem{Value: roundFloat2(current.CompletionTokens.Value / float32(currDayCount))}
	previous.DailyAvgTotalTokens = StatisticOverviewItem{Value: roundFloat2(previous.TotalTokens.Value / float32(prevDayCountI))}
	previous.DailyAvgPromptTokens = StatisticOverviewItem{Value: roundFloat2(previous.PromptTokens.Value / float32(prevDayCountI))}
	previous.DailyAvgCompletionTokens = StatisticOverviewItem{Value: roundFloat2(previous.CompletionTokens.Value / float32(prevDayCountI))}

	// PoP
	current.TotalTokens.PeriodOverPeriod = calculatePoP(current.TotalTokens.Value, previous.TotalTokens.Value)
	current.PromptTokens.PeriodOverPeriod = calculatePoP(current.PromptTokens.Value, previous.PromptTokens.Value)
	current.CompletionTokens.PeriodOverPeriod = calculatePoP(current.CompletionTokens.Value, previous.CompletionTokens.Value)
	current.DailyAvgTotalTokens.PeriodOverPeriod = calculatePoP(current.DailyAvgTotalTokens.Value, previous.DailyAvgTotalTokens.Value)
	current.DailyAvgPromptTokens.PeriodOverPeriod = calculatePoP(current.DailyAvgPromptTokens.Value, previous.DailyAvgPromptTokens.Value)
	current.DailyAvgCompletionTokens.PeriodOverPeriod = calculatePoP(current.DailyAvgCompletionTokens.Value, previous.DailyAvgCompletionTokens.Value)
	current.CallCount.PeriodOverPeriod = calculatePoP(current.CallCount.Value, previous.CallCount.Value)
	current.CallFailure.PeriodOverPeriod = calculatePoP(current.CallFailure.Value, previous.CallFailure.Value)
	current.AvgCosts.PeriodOverPeriod = calculatePoP(current.AvgCosts.Value, previous.AvgCosts.Value)
	current.AvgFirstTokenLatency.PeriodOverPeriod = calculatePoP(current.AvgFirstTokenLatency.Value, previous.AvgFirstTokenLatency.Value)
	return current, nil
}

// statisticV2OverviewByDateRange 返回某区间概览 + 该区间天数
func statisticV2OverviewByDateRange(ctx context.Context, db *gorm.DB, orgIds, userIds []string, dates []string, modelIds []string, modelType, viewScope string) (*ModelStatisticV2Overview, int, error) {
	if len(dates) == 0 {
		return &ModelStatisticV2Overview{}, 0, nil
	}
	startDate, endDate := dates[0], dates[len(dates)-1]
	var stat model.StatisticModel
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("SUM(prompt_tokens) as prompt_tokens, " +
			"SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, " +
			"SUM(call_count) as call_count, " +
			"SUM(call_failure) as call_failure, " +
			"SUM(stream_count) as stream_count, " +
			"SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, " +
			"SUM(non_stream_failure) as non_stream_failure, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(costs) as costs").
		First(&stat).Error; err != nil {
		return nil, 0, fmt.Errorf("statistic v2 overview [%v, %v] err: %v", startDate, endDate, err)
	}
	avgCosts := calculateAvg(stat.Costs, int32(stat.NonStreamCount-stat.NonStreamFailure))
	avgFirstTokenLatency := calculateAvg(stat.FirstTokenLatency, int32(stat.StreamCount-stat.StreamFailure))
	return &ModelStatisticV2Overview{
		TotalTokens:          StatisticOverviewItem{Value: float32(stat.TotalTokens)},
		PromptTokens:         StatisticOverviewItem{Value: float32(stat.PromptTokens)},
		CompletionTokens:     StatisticOverviewItem{Value: float32(stat.CompletionTokens)},
		CallCount:            StatisticOverviewItem{Value: float32(stat.CallCount)},
		CallFailure:          StatisticOverviewItem{Value: float32(stat.CallFailure)},
		AvgCosts:             StatisticOverviewItem{Value: avgCosts},
		AvgFirstTokenLatency: StatisticOverviewItem{Value: avgFirstTokenLatency},
	}, len(dates), nil
}

// ---------- 2. Trend ----------

// statisticV2Trend 趋势图（tokens 用量 + 调用趋势）。无 sync/校验，供 chart 入口复用。
func statisticV2Trend(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string) (*ModelStatisticV2Trend, error) {
	dates, err := buildDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("date, SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(total_tokens) as total_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(prompt_tokens) as prompt_tokens").
		Group("date").Order("date").Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("statistic v2 trend err: %v", err)
	}

	dateMap := make(map[string]model.StatisticModel, len(stats))
	for _, s := range stats {
		dateMap[s.Date] = s
	}

	var callTotal, callSuccess, callFailure []StatisticChartLineItem
	var totalTokens, completionTokens, promptTokens []StatisticChartLineItem
	for _, date := range dates {
		s := dateMap[date] // 缺日为零值
		callTotal = append(callTotal, StatisticChartLineItem{Key: date, Value: float32(s.CallCount)})
		callSuccess = append(callSuccess, StatisticChartLineItem{Key: date, Value: float32(s.CallCount - s.CallFailure)})
		callFailure = append(callFailure, StatisticChartLineItem{Key: date, Value: float32(s.CallFailure)})
		totalTokens = append(totalTokens, StatisticChartLineItem{Key: date, Value: float32(s.TotalTokens)})
		completionTokens = append(completionTokens, StatisticChartLineItem{Key: date, Value: float32(s.CompletionTokens)})
		promptTokens = append(promptTokens, StatisticChartLineItem{Key: date, Value: float32(s.PromptTokens)})
	}

	return &ModelStatisticV2Trend{
		ModelCalls: StatisticChart{
			Name: "app_statistic_model_call_trend",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_call_count_total", Items: callTotal},
				{Name: "app_statistic_call_success", Items: callSuccess},
				{Name: "app_statistic_call_failure", Items: callFailure},
			},
		},
		TokensUsage: StatisticChart{
			Name: "app_statistic_model_tokens_usage_trend",
			Lines: []StatisticChartLine{
				{Name: "app_statistic_total_tokens", Items: totalTokens},
				{Name: "app_statistic_completion_tokens", Items: completionTokens},
				{Name: "app_statistic_prompt_tokens", Items: promptTokens},
			},
		},
	}, nil
}

// ---------- 3. Rank ----------

// statisticV2Rank 三维度排行。无 sync/校验，供 chart 入口复用。
func statisticV2Rank(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, limit int32) (*ModelStatisticV2Rank, error) {
	if limit <= 0 {
		limit = 5
	}
	byModel, err := statisticV2RankByModel(ctx, db, orgIds, userIds, startDate, endDate, modelIds, modelType, viewScope, limit)
	if err != nil {
		return nil, err
	}
	byUser, err := statisticV2RankByUser(ctx, db, orgIds, userIds, startDate, endDate, modelIds, modelType, viewScope, limit)
	if err != nil {
		return nil, err
	}
	byOrg, err := statisticV2RankByOrg(ctx, db, orgIds, userIds, startDate, endDate, modelIds, modelType, viewScope, limit)
	if err != nil {
		return nil, err
	}
	return &ModelStatisticV2Rank{ByModel: byModel, ByUser: byUser, ByOrg: byOrg}, nil
}

// ---------- Chart (Trend + Rank) ----------

// GetModelStatisticV2Chart 趋势 + 排行合并查询（单次 sync）。
func (c *Client) GetModelStatisticV2Chart(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, limit int32) (*ModelStatisticV2Chart, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_model_statistic_v2_chart", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for today %v err: %v", today, err)
	}

	trend, err := statisticV2Trend(ctx, c.db, orgIds, userIds, startDate, endDate, modelIds, modelType, viewScope)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_v2_chart", err.Error())
	}
	rank, err := statisticV2Rank(ctx, c.db, orgIds, userIds, startDate, endDate, modelIds, modelType, viewScope, limit)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_v2_chart", err.Error())
	}
	return &ModelStatisticV2Chart{Trend: *trend, Rank: *rank}, nil
}

func statisticV2RankByModel(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, limit int32) ([]ModelStatisticV2RankByModelItem, error) {
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, " +
			"ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, " +
			"ANY_VALUE(model_creator_org_id) as model_creator_org_id, " +
			"SUM(total_tokens) as total_tokens").
		Group("model_id").Order("SUM(total_tokens) DESC").Limit(int(limit)).Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("rank by model err: %v", err)
	}
	items := make([]ModelStatisticV2RankByModelItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, ModelStatisticV2RankByModelItem{
			ModelId:            s.ModelID,
			Model:              s.Model,
			Provider:           s.Provider,
			ModelType:          s.ModelType,
			ModelCreatorUserId: s.ModelCreatorUserID,
			ModelCreatorOrgId:  s.ModelCreatorOrgID,
			TotalTokens:        s.TotalTokens,
		})
	}
	return items, nil
}

func statisticV2RankByUser(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, limit int32) ([]ModelStatisticV2RankByUserItem, error) {
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("user_id, ANY_VALUE(org_id) as org_id, SUM(total_tokens) as total_tokens").
		Group("user_id, org_id").Order("SUM(total_tokens) DESC").Limit(int(limit)).Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("rank by user err: %v", err)
	}
	items := make([]ModelStatisticV2RankByUserItem, 0, len(stats))
	for _, s := range stats {
		if s.UserID == "" {
			continue
		}
		items = append(items, ModelStatisticV2RankByUserItem{
			UserId:      s.UserID,
			OrgId:       s.OrgID,
			TotalTokens: s.TotalTokens,
		})
	}
	return items, nil
}

func statisticV2RankByOrg(ctx context.Context, db *gorm.DB, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, limit int32) ([]ModelStatisticV2RankByOrgItem, error) {
	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("org_id, SUM(call_count) as call_count, SUM(total_tokens) as total_tokens").
		Group("org_id").Order("SUM(total_tokens) DESC").Limit(int(limit)).Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("rank by org err: %v", err)
	}
	items := make([]ModelStatisticV2RankByOrgItem, 0, len(stats))
	for _, s := range stats {
		if s.OrgID == "" {
			continue
		}
		items = append(items, ModelStatisticV2RankByOrgItem{
			OrgId:       s.OrgID,
			CallCount:   s.CallCount,
			TotalTokens: s.TotalTokens,
		})
	}
	return items, nil
}

// ---------- 4. List (主表，按 model_id 聚合) ----------

// GetModelStatisticV2List 主表分页（按 model_id 聚合）
func (c *Client) GetModelStatisticV2List(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope string, sortExpr, sortOrder string, offset, limit int32) ([]ModelStatisticV2ListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_model_statistic_v2_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for today %v err: %v", today, err)
	}

	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := countQuery.Select("COUNT(DISTINCT model_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_list", fmt.Sprintf("count err: %v", err))
	}

	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	query = query.
		Select("model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, " +
			"ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, " +
			"ANY_VALUE(model_creator_org_id) as model_creator_org_id, " +
			"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, SUM(costs) as costs, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("model_id").Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]ModelStatisticV2ListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, ModelStatisticV2ListItem{
			StatisticV2ModelRef: StatisticV2ModelRef{
				ModelId: s.ModelID, Model: s.Model, Provider: s.Provider, ModelType: s.ModelType,
				ModelCreatorUserId: s.ModelCreatorUserID, ModelCreatorOrgId: s.ModelCreatorOrgID,
			},
			Metrics: buildV2Metrics(&s),
		})
	}
	return items, int32(total), nil
}

// ---------- 5. UserList (用户钻取，按 model_id+user_id+org_id) ----------

// GetModelStatisticV2UserList 用户钻取分页
func (c *Client) GetModelStatisticV2UserList(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope, modelId, sortExpr, sortOrder string, offset, limit int32) ([]ModelStatisticV2UserListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_model_statistic_v2_user_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for today %v err: %v", today, err)
	}

	if modelId == "" {
		return nil, 0, toErrStatus("app_model_statistic_v2_user_list", "modelId required")
	}

	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModelID(modelId),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := countQuery.Select("COUNT(DISTINCT user_id, org_id)").Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_user_list", fmt.Sprintf("count err: %v", err))
	}

	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	query = query.
		Select("ANY_VALUE(model_id) as model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, " +
			"ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, " +
			"ANY_VALUE(model_creator_org_id) as model_creator_org_id, " +
			"user_id, org_id, " +
			"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, SUM(costs) as costs, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("user_id, org_id").Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_user_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]ModelStatisticV2UserListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, ModelStatisticV2UserListItem{
			StatisticV2ModelRef: StatisticV2ModelRef{
				ModelId: s.ModelID, Model: s.Model, Provider: s.Provider, ModelType: s.ModelType,
				ModelCreatorUserId: s.ModelCreatorUserID, ModelCreatorOrgId: s.ModelCreatorOrgID,
			},
			StatisticV2UserRef: StatisticV2UserRef{UserId: s.UserID, OrgId: s.OrgID},
			Metrics:            buildV2Metrics(&s),
		})
	}
	return items, int32(total), nil
}

// ---------- 6. AppList (应用钻取) ----------

// GetModelStatisticV2AppList 应用钻取分页
func (c *Client) GetModelStatisticV2AppList(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope, modelId, sortExpr, sortOrder string, offset, limit int32) ([]ModelStatisticV2AppListItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_model_statistic_v2_app_list", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := syncStatisticModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync statistic model stats for today %v err: %v", today, err)
	}

	if modelId == "" {
		return nil, 0, toErrStatus("app_model_statistic_v2_app_list", "modelId required")
	}

	opts := []sqlopt.SQLOption{
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModelID(modelId),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := countQuery.
		Select("COUNT(DISTINCT app_id, source, module, app_type, module_creator_user_id, module_creator_org_id)").
		Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_app_list", fmt.Sprintf("count err: %v", err))
	}

	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	query = query.
		Select("ANY_VALUE(model_id) as model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, " +
			"ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, " +
			"ANY_VALUE(model_creator_org_id) as model_creator_org_id, " +
			"source, module, app_id, app_type, " +
			"module_creator_user_id, module_creator_org_id, " +
			"SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, SUM(costs) as costs, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count, " +
			"SUM(stream_failure) as stream_failure, SUM(non_stream_failure) as non_stream_failure").
		Group("app_id, source, module, app_type, module_creator_user_id, module_creator_org_id").
		Order(buildV2OrderClause(sortExpr, sortOrder, "SUM(call_count)"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_app_list", fmt.Sprintf("list err: %v", err))
	}
	items := make([]ModelStatisticV2AppListItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, ModelStatisticV2AppListItem{
			StatisticV2ModelRef: StatisticV2ModelRef{
				ModelId: s.ModelID, Model: s.Model, Provider: s.Provider, ModelType: s.ModelType,
				ModelCreatorUserId: s.ModelCreatorUserID, ModelCreatorOrgId: s.ModelCreatorOrgID,
			},
			StatisticV2AppRef: StatisticV2AppRef{
				Source: s.Source, Module: s.Module, AppId: s.AppID, AppType: s.AppType,
				ModuleCreatorUserId: s.ModuleCreatorUserID, ModuleCreatorOrgId: s.ModuleCreatorOrgID,
			},
			Metrics: buildV2Metrics(&s),
		})
	}
	return items, int32(total), nil
}

// ---------- 7. Record 明细列表 ----------

// GetModelStatisticV2Record 调用明细分页（查 ModelRecord 表）
func (c *Client) GetModelStatisticV2Record(ctx context.Context, orgIds, userIds []string, startDate, endDate string, modelIds []string, modelType, viewScope, modelId, sortExpr, sortOrder string, offset, limit int32) ([]ModelStatisticV2RecordItem, int32, *errs.Status) {
	if startDate > endDate {
		return nil, 0, toErrStatus("app_model_statistic_v2_record", fmt.Sprintf("startDate %v greater than endDate %v", startDate, endDate))
	}
	createdAtOpts, err := buildRecordCreatedAtOpts(startDate, endDate)
	if err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_record", err.Error())
	}

	opts := append(createdAtOpts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)
	if modelId != "" {
		opts = append(opts, sqlopt.WithModelID(modelId))
	}
	if len(modelIds) > 0 {
		opts = append(opts, sqlopt.WithModelIds(modelIds))
	}
	if modelType != "" {
		opts = append(opts, sqlopt.WithModelType(modelType))
	}

	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.ModelRecordV2{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_record", fmt.Sprintf("count err: %v", err))
	}

	var records []model.ModelRecordV2
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.ModelRecordV2{})
	query = query.Order(buildV2OrderClause(sortExpr, sortOrder, "created_at"))
	if offset >= 0 && limit > 0 {
		query = query.Offset(int(offset)).Limit(int(limit))
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, 0, toErrStatus("app_model_statistic_v2_record", fmt.Sprintf("list err: %v", err))
	}
	items := make([]ModelStatisticV2RecordItem, 0, len(records))
	for _, r := range records {
		items = append(items, ModelStatisticV2RecordItem{
			Id:      uint32(r.ID),
			TraceId: r.TraceID,
			StatisticV2ModelRef: StatisticV2ModelRef{
				ModelId: r.ModelID, Model: r.Model, Provider: r.Provider, ModelType: r.ModelType,
				ModelCreatorUserId: r.ModelCreatorUserID, ModelCreatorOrgId: r.ModelCreatorOrgID,
			},
			StatisticV2AppRef: StatisticV2AppRef{
				Source: r.Source, Module: r.Module, AppId: r.AppID, AppType: r.AppType,
				ModuleCreatorUserId: r.ModuleCreatorUserID, ModuleCreatorOrgId: r.ModuleCreatorOrgID,
			},
			StatisticV2UserRef: StatisticV2UserRef{UserId: r.UserID, OrgId: r.OrgID},
			StatisticV2TokenMetrics: StatisticV2TokenMetrics{
				TotalTokens: r.TotalTokens, PromptTokens: r.PromptTokens, CompletionTokens: r.CompletionTokens,
			},
			CreatedAt: r.CreatedAt, IsSuccess: statistic.IsSuccess(r.StatusCode), StatusCode: r.StatusCode,
			FailureReason: r.FailureReason, RequestBody: r.RequestBody, ResponseBody: r.ResponseBody,
			FinishReason: r.FinishReason, FirstTokenLatency: r.FirstTokenLatency, Costs: r.Costs,
		})
	}
	return items, int32(total), nil
}

// ---------- 共用 helper ----------

// buildV2Metrics 从聚合行构建通用度量
func buildV2Metrics(s *model.StatisticModel) ModelStatisticV2Metrics {
	return ModelStatisticV2Metrics{
		StatisticV2TokenMetrics: StatisticV2TokenMetrics{
			TotalTokens:      s.TotalTokens,
			PromptTokens:     s.PromptTokens,
			CompletionTokens: s.CompletionTokens,
		},
		StatisticV2CallMetrics: StatisticV2CallMetrics{
			CallCount:            s.CallCount,
			CallFailure:          s.CallFailure,
			FailureRate:          calculateFailureRate(s.CallFailure, s.CallCount),
			AvgCosts:             calculateAvg(s.Costs, int32(s.NonStreamCount-s.NonStreamFailure)),
			AvgFirstTokenLatency: calculateAvg(s.FirstTokenLatency, int32(s.StreamCount-s.StreamFailure)),
		},
	}
}

// buildV2OrderClause 构造 ORDER BY 子句。sortExpr 为 BFF 白名单映射后的安全列表达式，
// 为空时回退 defaultExpr；sortOrder 仅接受 "asc"/"desc"，其余回退 "desc"。
// 聚合列表的 defaultExpr / sortExpr 须为 SUM(...) 等形式，避免 ONLY_FULL_GROUP_BY 下裸列歧义。
func buildV2OrderClause(sortExpr, sortOrder, defaultExpr string) string {
	expr := defaultExpr
	if sortExpr != "" {
		expr = sortExpr
	}
	dir := "DESC"
	if sortOrder == "asc" {
		dir = "ASC"
	}
	return expr + " " + dir
}

// ModelStatisticV2SelectItem 模型下拉选项（聚合表 DISTINCT model_id）
type ModelStatisticV2SelectItem = StatisticV2ModelRef

// ListModelStatisticV2Select 从聚合表查询出现过的模型 ID（按 viewScope 过滤，不限日期）。
// 下拉不触发当日 Redis→DB sync，依赖 cron / 其它读路径落库。
func (c *Client) ListModelStatisticV2Select(ctx context.Context, orgIds, userIds []string, modelType, viewScope string) ([]ModelStatisticV2SelectItem, *errs.Status) {
	opts := []sqlopt.SQLOption{
		sqlopt.WithNonEmptyModelID(),
		sqlopt.WithModelType(modelType),
	}
	opts = append(opts, statisticV2ScopeOptions(viewScope, orgIds, userIds)...)

	var stats []model.StatisticModel
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.StatisticModel{})
	if err := query.
		Select("model_id, ANY_VALUE(model) as model, ANY_VALUE(provider) as provider, " +
			"ANY_VALUE(model_type) as model_type, " +
			"ANY_VALUE(model_creator_user_id) as model_creator_user_id, " +
			"ANY_VALUE(model_creator_org_id) as model_creator_org_id").
		Group("model_id").
		Order("model_id ASC").
		Find(&stats).Error; err != nil {
		return nil, toErrStatus("app_model_statistic_v2_select", fmt.Sprintf("query err: %v", err))
	}

	items := make([]ModelStatisticV2SelectItem, 0, len(stats))
	for _, s := range stats {
		items = append(items, ModelStatisticV2SelectItem{
			ModelId: s.ModelID, Model: s.Model, Provider: s.Provider, ModelType: s.ModelType,
			ModelCreatorUserId: s.ModelCreatorUserID, ModelCreatorOrgId: s.ModelCreatorOrgID,
		})
	}
	return items, nil
}
