package orm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const redisStatsExpireSeconds = 30 * 24 * 3600 // 30天

// ---------- 统计 V2 写入（Record）参数 ----------
// 以下结构体仅用于 Record*StatisticV2 落库 / 写 Redis，ORM 层定义，不依赖 proto。

// RecordModelStatisticV2Input 模型统计 V2 单次调用写入参数。
type RecordModelStatisticV2Input struct {
	TraceID string

	UserID string
	OrgID  string

	Source              string
	Module              string
	AppID               string
	AppType             string
	APIKey              string
	APIKeyID            string
	MethodPath          string
	ModuleCreatorUserID string
	ModuleCreatorOrgID  string

	ModelID            string
	Model              string
	Provider           string
	ModelType          string
	ModelCreatorUserID string
	ModelCreatorOrgID  string

	PromptTokens      int64
	CompletionTokens  int64
	TotalTokens       int64
	FirstTokenLatency int64
	Costs             int64
	IsSuccess         bool
	IsStream          bool
	StatusCode        int64
	RequestBody       string
	ResponseBody      string
	FinishReason      string
	FailureReason     string
}

// RecordAppStatisticV2Input 应用统计 V2 单次调用写入参数。
type RecordAppStatisticV2Input struct {
	TraceID             string
	UserID              string
	OrgID               string
	Source              string
	Module              string
	AppID               string
	AppType             string
	ModuleCreatorUserID string
	ModuleCreatorOrgID  string
	FirstTokenLatency   int64
	Costs               int64
	IsSuccess           bool
	IsStream            bool
	StatusCode          int64
	RequestBody         string
	ResponseBody        string
	FinishReason        string
	FailureReason       string
	Question            string
	Answer              string
}

// RecordAPIKeyStatisticV2Input API Key 统计 V2 单次调用写入参数。
type RecordAPIKeyStatisticV2Input struct {
	TraceID             string
	UserID              string
	OrgID               string
	APIKeyID            string
	MethodPath          string
	CalledAt            int64
	Source              string
	Module              string
	ModuleCreatorUserID string
	ModuleCreatorOrgID  string
	AppID               string
	AppType             string
	IsStream            bool
	Costs               int64
	FirstTokenLatency   int64
	IsSuccess           bool
	StatusCode          int64
	FailureReason       string
	RequestBody         string
	ResponseBody        string
}

type StatisticOverviewItem struct {
	Value            float32 `json:"value"`
	PeriodOverPeriod float32 `json:"periodOverPeriod"`
}

// StatisticV2CallOverview 应用 / API Key 共用的调用类概览（调用次数、流式/非流式、日均、平均耗时）。
type StatisticV2CallOverview struct {
	CallCount              StatisticOverviewItem
	CallFailure            StatisticOverviewItem
	DailyAvgCallCount      StatisticOverviewItem
	DailyAvgCallFailure    StatisticOverviewItem
	DailyAvgStreamCount    StatisticOverviewItem
	DailyAvgNonStreamCount StatisticOverviewItem
	AvgFirstTokenLatency   StatisticOverviewItem
	AvgCosts               StatisticOverviewItem
	StreamCount            StatisticOverviewItem
	NonStreamCount         StatisticOverviewItem
}

// AppStatisticV2Overview / APIKeyStatisticV2Overview 与 StatisticV2CallOverview 同构，保留别名兼容既有调用。
type AppStatisticV2Overview = StatisticV2CallOverview
type APIKeyStatisticV2Overview = StatisticV2CallOverview

// StatisticChart / Line / Item 为模型、应用、API Key 趋势图共用结构。
type StatisticChart struct {
	Name  string               `json:"name"`
	Lines []StatisticChartLine `json:"lines"`
}

type StatisticChartLine struct {
	Name  string                   `json:"name"`
	Items []StatisticChartLineItem `json:"items"`
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}

// ---------- 列表行共用嵌套：身份 ----------

// StatisticV2ModelRef 模型维度身份（列表 / 钻取 / 下拉 / 明细复用）。
type StatisticV2ModelRef struct {
	ModelId            string
	Model              string
	Provider           string
	ModelType          string
	ModelCreatorUserId string
	ModelCreatorOrgId  string
}

// StatisticV2AppRef 应用 / 板块维度身份。
type StatisticV2AppRef struct {
	Source              string
	Module              string
	AppId               string
	AppType             string
	ModuleCreatorUserId string
	ModuleCreatorOrgId  string
}

// StatisticV2UserRef 调用方用户 / 组织。
type StatisticV2UserRef struct {
	UserId string
	OrgId  string
}

// StatisticV2APIKeyRef API Key + 路径 + 归属用户。
type StatisticV2APIKeyRef struct {
	ApiKeyId   string
	MethodPath string
	OrgId      string
	UserId     string
}

// ---------- 列表行共用嵌套：度量 ----------

type StatisticV2CallMetrics struct {
	CallCount            int32
	CallFailure          int32
	FailureRate          float32
	AvgCosts             float32
	AvgFirstTokenLatency float32
}

type StatisticV2TokenMetrics struct {
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
}

type StatisticV2StreamMetrics struct {
	StreamCount    int32
	NonStreamCount int32
}

// ModelStatisticV2Metrics 模型列表行度量（tokens + 调用）。
type ModelStatisticV2Metrics struct {
	StatisticV2TokenMetrics
	StatisticV2CallMetrics
}

// AppStatisticV2Metrics 应用列表行度量（tokens + 调用 + 流式）。
type AppStatisticV2Metrics struct {
	StatisticV2TokenMetrics
	StatisticV2CallMetrics
	StatisticV2StreamMetrics
}

// APIKeyStatisticV2Metrics API Key 列表行度量（调用 + 流式）。
type APIKeyStatisticV2Metrics struct {
	StatisticV2CallMetrics
	StatisticV2StreamMetrics
}

// roundFloat2 保留小数点后 2 位（四舍五入）。统计浮点展示精度统一在 app-service 取整，BFF 透传。
func roundFloat2(v float32) float32 {
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", v), 32)
	return float32(value)
}

func calculatePoP(current, previous float32) float32 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return roundFloat2(((current - previous) / previous) * 100)
}

func calculateFailureRate(failureCount, totalCount int32) float32 {
	if totalCount == 0 {
		return 0
	}
	return roundFloat2(float32(failureCount) / float32(totalCount) * 100)
}

func calculateAvg(totalCosts int64, successCount int32) float32 {
	if successCount <= 0 {
		return 0
	}
	return roundFloat2(float32(totalCosts) / float32(successCount))
}

func buildDateRange(startDate, endDate string) ([]string, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}
	return util.DateRange(startTs, endTs), nil
}

// buildRecordCreatedAtOpts 明细表（RecordV2）按 created_at 过滤，无 date 列。
func buildRecordCreatedAtOpts(startDate, endDate string) ([]sqlopt.SQLOption, error) {
	dates, err := buildDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(dates) == 0 {
		return nil, fmt.Errorf("invalid date range %v-%v", startDate, endDate)
	}
	startTs, _ := util.Date2Time(dates[0])
	endTs, _ := util.Date2Time(dates[len(dates)-1])
	endTs = endTs + int64(24*time.Hour/time.Millisecond) - 1
	return []sqlopt.SQLOption{
		sqlopt.StartCreatedAt(startTs),
		sqlopt.EndCreatedAt(endTs),
	}, nil
}

// statisticCostsForAvg 聚合平均耗时只计入成功调用；失败记 0，与 calculateAvg(..., successCount) 对齐。
// 明细表仍保留真实耗时，便于单次失败排查。
func statisticCostsForAvg(statusCode int64, firstTokenLatency, costs int64) (ftl, c int64) {
	if !statistic.IsSuccess(statusCode) {
		return 0, 0
	}
	return firstTokenLatency, costs
}
