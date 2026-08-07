package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// StatisticV2CallOverview 应用 / API Key 共用的调用类概览 10 指标卡
// （调用次数、流式/非流式、日均、平均耗时；与模型 V2 Overview 的 Tokens 主指标不同）。
type StatisticV2CallOverview struct {
	CallCount              StatisticOverviewItem `json:"callCount"`              // 调用总次数
	CallFailure            StatisticOverviewItem `json:"callFailure"`            // 失败次数
	DailyAvgCallCount      StatisticOverviewItem `json:"dailyAvgCallCount"`      // 日均调用总次数
	DailyAvgCallFailure    StatisticOverviewItem `json:"dailyAvgCallFailure"`    // 日均失败次数
	DailyAvgStreamCount    StatisticOverviewItem `json:"dailyAvgStreamCount"`    // 日均调用次数(流式)
	DailyAvgNonStreamCount StatisticOverviewItem `json:"dailyAvgNonStreamCount"` // 日均调用次数(非流式)
	AvgFirstTokenLatency   StatisticOverviewItem `json:"avgFirstTokenLatency"`   // 流式平均首Token时延 ms
	AvgCosts               StatisticOverviewItem `json:"avgCosts"`               // 非流式平均耗时 ms
	StreamCount            StatisticOverviewItem `json:"streamCount"`            // 流式调用次数
	NonStreamCount         StatisticOverviewItem `json:"nonStreamCount"`         // 非流式调用次数
}

// AppStatisticV2Overview 应用统计概览（与 API Key 共用 StatisticV2CallOverview）。
type AppStatisticV2Overview = StatisticV2CallOverview

// APIKeyStatisticV2Overview API Key 统计概览（与应用共用 StatisticV2CallOverview）。
type APIKeyStatisticV2Overview = StatisticV2CallOverview

// UserBriefInfo 用户简要信息（统计 V2 跨模型/应用/API Key 复用）。
type UserBriefInfo struct {
	UserId     string         `json:"userId"`
	UserName   string         `json:"userName"`
	UserAvatar request.Avatar `json:"userAvatar"`
	OrgId      string         `json:"orgId"`
	OrgName    string         `json:"orgName"` // 调用人组织（≠ module/model creator org）
}

// ModuleBriefInfo 应用（module）简要信息（统计 V2 跨模型/应用/API Key 复用）。
type ModuleBriefInfo struct {
	Source       string         `json:"source"`     // 调用来源 code：web|openapi|webURL
	SourceName   string         `json:"sourceName"` // 调用来源展示名：Web|Openapi|WebURL
	Module       string         `json:"module"`     // 板块 code：wga|skill|knowledge|...
	ModuleName   string         `json:"moduleName"` // 板块展示名：通用智能体|智能体|知识库|...
	AppId        string         `json:"appId"`      // 应用ID
	AppName      string         `json:"appName"`    // 应用名称
	AppType      string         `json:"appType"`    // 应用类型 agent|rag|workflow|chatflow|...
	ModuleAvatar request.Avatar `json:"moduleAvatar"`
	StatisticV2ModuleCreator
}

// ModelBriefInfo 模型简要信息（统计 V2 跨模型/应用 复用）。
type ModelBriefInfo struct {
	ModelId     string         `json:"modelId"`
	Model       string         `json:"model"`
	Provider    string         `json:"provider"`
	ModelType   string         `json:"modelType"`
	ModelAvatar request.Avatar `json:"modelAvatar"`
	StatisticV2ModelCreator
}

// StatisticV2ModelCreator 模型创建人（发布者）信息，详情页等仅需创建人字段时复用。
type StatisticV2ModelCreator struct {
	ModelCreatorUserId   string `json:"modelCreatorUserId"`
	ModelCreatorUserName string `json:"modelCreatorUserName"`
	ModelCreatorOrgId    string `json:"modelCreatorOrgId"`
	ModelCreatorOrgName  string `json:"modelCreatorOrgName"`
}

// StatisticV2ModuleCreator 应用（module）创建人信息，详情页等仅需创建人字段时复用。
type StatisticV2ModuleCreator struct {
	ModuleCreatorUserId   string `json:"moduleCreatorUserId"`
	ModuleCreatorUserName string `json:"moduleCreatorUserName"`
	ModuleCreatorOrgId    string `json:"moduleCreatorOrgId"`
	ModuleCreatorOrgName  string `json:"moduleCreatorOrgName"`
}

// StatisticV2Metrics 聚合调用统计指标（主表 / 用户钻取 / 模型钻取等 list 接口）。
// 耗时字段为时间段内多次调用的平均值（avgCosts / avgFirstTokenLatency），不是单次 record 耗时。
type StatisticV2Metrics struct {
	TotalTokens          int64   `json:"totalTokens"`
	PromptTokens         int64   `json:"promptTokens"`
	CompletionTokens     int64   `json:"completionTokens"`
	CallCount            int32   `json:"callCount"`
	CallFailure          int32   `json:"callFailure"`
	FailureRate          float32 `json:"failureRate"`
	AvgCosts             float32 `json:"avgCosts"`             // 非流式平均耗时 ms
	AvgFirstTokenLatency float32 `json:"avgFirstTokenLatency"` // 流式平均首Token时延 ms
	StreamCount          int32   `json:"streamCount"`          // 流式调用次数
	NonStreamCount       int32   `json:"nonStreamCount"`       // 非流式调用次数
}

// StatisticV2RankItem 排行项（应用维度各板块 TopN 共用）。
type StatisticV2RankItem struct {
	AppId     string         `json:"appId"`
	AppName   string         `json:"appName"`
	Avatar    request.Avatar `json:"avatar"`
	CallCount int32          `json:"callCount"`
	StatisticV2ModuleCreator
}

// StatisticV2RecordPerformance 单次调用耗时（record 明细列表行 / 详情弹窗共用，平铺序列化）。
// 每条 record 对应一次真实调用，firstTokenLatency / costs 为该次调用的实际耗时，非聚合平均。
type StatisticV2RecordPerformance struct {
	FirstTokenLatency int64 `json:"firstTokenLatency"` // 本次流式首Token时延(TTFT) ms
	Costs             int64 `json:"costs"`             // 本次非流式耗时 ms
}

// ApiKeyBriefInfo API Key 行前缀（主表 / 应用钻取 / 模型钻取 / 明细列表共用）。
type ApiKeyBriefInfo struct {
	ApiName    string `json:"apiName"` // API Key 名称
	ApiKeyId   string `json:"apiKeyId"`
	ApiKey     string `json:"apiKey"`     // 脱敏展示 sk-xxxx****xxxx
	MethodPath string `json:"methodPath"` // 路径
	UserBriefInfo
}

// APIKeyStatisticV2Metrics API Key 聚合统计指标（主表 / 应用钻取 list）；耗时为平均值。
type APIKeyStatisticV2Metrics struct {
	CallCount            int32   `json:"callCount"`
	CallFailure          int32   `json:"callFailure"`
	FailureRate          float32 `json:"failureRate"`
	AvgFirstTokenLatency float32 `json:"avgFirstTokenLatency"` // 流式平均首Token时延 ms
	AvgCosts             float32 `json:"avgCosts"`             // 非流式平均耗时 ms
	StreamCount          int32   `json:"streamCount"`
	NonStreamCount       int32   `json:"nonStreamCount"`
}

// MyAppItem 统计看板应用下拉项（已发布应用）。
type MyAppItem struct {
	AppId       string         `json:"appId"`
	Name        string         `json:"name"`
	AppType     string         `json:"appType"`
	Avatar      request.Avatar `json:"avatar"` // 图标
	PublishType string         `json:"publishType"`
	CreatedAt   int64          `json:"createdAt"`
}
