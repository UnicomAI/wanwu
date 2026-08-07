package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

// ModelStatisticV2Metrics 聚合调用统计指标（主表 / 钻取 list）；耗时为平均值。
type ModelStatisticV2Metrics struct {
	TotalTokens          int64   `json:"totalTokens"`
	PromptTokens         int64   `json:"promptTokens"`
	CompletionTokens     int64   `json:"completionTokens"`
	CallCount            int32   `json:"callCount"`
	CallFailure          int32   `json:"callFailure"`
	FailureRate          float32 `json:"failureRate"`
	AvgCosts             float32 `json:"avgCosts"`             // 非流式平均耗时 ms
	AvgFirstTokenLatency float32 `json:"avgFirstTokenLatency"` // 流式平均耗时 ms
}

// ModelStatisticV2Overview 10 指标卡
type ModelStatisticV2Overview struct {
	TotalTokens              StatisticOverviewItem `json:"totalTokens"`
	PromptTokens             StatisticOverviewItem `json:"promptTokens"`
	CompletionTokens         StatisticOverviewItem `json:"completionTokens"`
	DailyAvgTotalTokens      StatisticOverviewItem `json:"dailyAvgTotalTokens"`
	DailyAvgPromptTokens     StatisticOverviewItem `json:"dailyAvgPromptTokens"`
	DailyAvgCompletionTokens StatisticOverviewItem `json:"dailyAvgCompletionTokens"`
	CallCount                StatisticOverviewItem `json:"callCount"`
	CallFailure              StatisticOverviewItem `json:"callFailure"`
	AvgCosts                 StatisticOverviewItem `json:"avgCosts"`
	AvgFirstTokenLatency     StatisticOverviewItem `json:"avgFirstTokenLatency"`
}

// ModelStatisticV2Trend 趋势
type ModelStatisticV2Trend struct {
	TokensUsage StatisticChart `json:"tokensUsage"`
	ModelCalls  StatisticChart `json:"modelCalls"`
}

// ModelStatisticV2Chart 趋势 + 排行
type ModelStatisticV2Chart struct {
	Trend ModelStatisticV2Trend `json:"trend"`
	Rank  ModelStatisticV2Rank  `json:"rank"`
}

// ModelStatisticV2Rank 排行
type ModelStatisticV2Rank struct {
	ByModel []ModelStatisticV2RankByModelItem `json:"byModel"`
	ByUser  []ModelStatisticV2RankByUserItem  `json:"byUser"`
	ByOrg   []ModelStatisticV2RankByOrgItem   `json:"byOrg"`
}

type ModelStatisticV2RankByModelItem struct {
	TotalTokens int64 `json:"totalTokens"`
	ModelBriefInfo
}

type ModelStatisticV2RankByUserItem struct {
	UserId      string         `json:"userId"`
	UserName    string         `json:"userName"`
	Avatar      request.Avatar `json:"avatar"`
	OrgId       string         `json:"orgId"`
	OrgName     string         `json:"orgName"`
	TotalTokens int64          `json:"totalTokens"`
}

type ModelStatisticV2RankByOrgItem struct {
	OrgId       string         `json:"orgId"`
	OrgName     string         `json:"orgName"`
	Avatar      request.Avatar `json:"avatar"`
	TotalTokens int64          `json:"totalTokens"`
}

// ModelStatisticV2ListItem 主表行
type ModelStatisticV2ListItem struct {
	ModelBriefInfo
	ModelStatisticV2Metrics
}

// ModelStatisticV2UserListItem 用户钻取行
type ModelStatisticV2UserListItem struct {
	ModelBriefInfo
	UserBriefInfo
	ModelStatisticV2Metrics
}

// ModelStatisticV2AppListItem 应用钻取行
type ModelStatisticV2AppListItem struct {
	ModelBriefInfo
	ModuleBriefInfo
	ModelStatisticV2Metrics
}

// ModelStatisticV2RecordItem 调用明细列表行（单次调用 record；含详情字段）
type ModelStatisticV2RecordItem struct {
	Id               string `json:"id"`
	TraceId          string `json:"traceId"`
	TotalTokens      int64  `json:"totalTokens"`
	PromptTokens     int64  `json:"promptTokens"`
	CompletionTokens int64  `json:"completionTokens"`
	CalledAt         string `json:"calledAt"`
	IsSuccess        bool   `json:"isSuccess"`
	StatusCode       int64  `json:"statusCode"`
	FailureReason    string `json:"failureReason"`
	RequestBody      string `json:"requestBody"`
	ResponseBody     string `json:"responseBody"`
	FinishReason     string `json:"finishReason"`
	UserBriefInfo
	ModelBriefInfo
	ModuleBriefInfo
	StatisticV2RecordPerformance
}
