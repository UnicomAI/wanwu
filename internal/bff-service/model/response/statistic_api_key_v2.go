package response

// ApiKeyStatisticRouteItem OpenAPI 路由（API Key 统计筛选用）
type ApiKeyStatisticRouteItem struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// APIKeyStatisticV2Chart 趋势 + 排行
type APIKeyStatisticV2Chart struct {
	Trend APIKeyStatisticV2Trend `json:"trend"`
	Rank  APIKeyStatisticV2Rank  `json:"rank"`
}

// APIKeyStatisticV2Trend 趋势图表
type APIKeyStatisticV2Trend struct {
	ApiKeyCalls StatisticChart `json:"apiKeyCalls"` // APIKey 总/成功/失败 折线
	CallResult  StatisticChart `json:"callResult"`  // 成功/失败 堆叠柱
}

// APIKeyStatisticV2Rank API Key 调用次数排行
type APIKeyStatisticV2Rank struct {
	ByApi []APIKeyStatisticV2RankItem `json:"byApi"`
}

// APIKeyStatisticV2RankItem 单条 API Key 排行（按 apiKey 聚合，不含 methodPath）
type APIKeyStatisticV2RankItem struct {
	ApiName   string `json:"apiName"`
	CallCount int32  `json:"callCount"`
	UserBriefInfo
}

// APIKeyStatisticV2ListItem 调用统计主表行（聚合；耗时为 avgCosts / avgFirstTokenLatency）
type APIKeyStatisticV2ListItem struct {
	ApiKeyBriefInfo
	APIKeyStatisticV2Metrics
}

// APIKeyStatisticV2AppListItem 应用钻取行
type APIKeyStatisticV2AppListItem struct {
	ApiKeyBriefInfo
	ModuleBriefInfo
	APIKeyStatisticV2Metrics
}

// APIKeyStatisticV2ModelListItem 模型钻取行
type APIKeyStatisticV2ModelListItem struct {
	ApiKeyBriefInfo
	ModelBriefInfo
	StatisticV2Metrics
}

// APIKeyStatisticV2RecordItem 调用明细列表行（单次调用；耗时为 firstTokenLatency / costs，非平均）
type APIKeyStatisticV2RecordItem struct {
	Id            uint32 `json:"id"`
	CalledAt      string `json:"calledAt"`
	IsSuccess     bool   `json:"isSuccess"`
	StatusCode    int64  `json:"statusCode"`
	FailureReason string `json:"failureReason"`
	RequestBody   string `json:"requestBody"`
	ResponseBody  string `json:"responseBody"`
	ApiKeyBriefInfo
	ModuleBriefInfo
	StatisticV2RecordPerformance
}
