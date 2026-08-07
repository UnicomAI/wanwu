package service

// V2 统计列表 sortField 白名单：仅数值类指标可排序；未命中时 ORM 回退默认排序。
//
// 聚合列表（GROUP BY）必须用 SUM(...)：MySQL ONLY_FULL_GROUP_BY 下，
// ORDER BY 同名裸列会优先绑到表列而非 SELECT 别名，触发 Error 1055。
// 调用明细（单次调用 record list）无 GROUP BY、不开放用户排序，固定 created_at 降序。
//
// 分母 <= 0 时回落 0（与 Go calculateAvg / calculateFailureRate 一致）；*1.0 避免整除截断。

// sortFieldAggregateFull 含 Tokens 的聚合表排序（模型主表 / 模型钻取等）。
var sortFieldAggregateFull = map[string]string{
	"totalTokens":          "SUM(total_tokens)",
	"promptTokens":         "SUM(prompt_tokens)",
	"completionTokens":     "SUM(completion_tokens)",
	"callCount":            "SUM(call_count)",
	"callFailure":          "SUM(call_failure)",
	"streamCount":          "SUM(stream_count)",
	"nonStreamCount":       "SUM(non_stream_count)",
	"streamFailure":        "SUM(stream_failure)",
	"nonStreamFailure":     "SUM(non_stream_failure)",
	"firstTokenLatency":    "SUM(first_token_latency)",
	"costs":                "SUM(costs)",
	"failureRate":          "(CASE WHEN SUM(call_count) > 0 THEN SUM(call_failure) * 1.0 / SUM(call_count) ELSE 0 END)",
	"avgFirstTokenLatency": "(CASE WHEN (SUM(stream_count) - SUM(stream_failure)) > 0 THEN SUM(first_token_latency) * 1.0 / (SUM(stream_count) - SUM(stream_failure)) ELSE 0 END)",
	"avgCosts":             "(CASE WHEN (SUM(non_stream_count) - SUM(non_stream_failure)) > 0 THEN SUM(costs) * 1.0 / (SUM(non_stream_count) - SUM(non_stream_failure)) ELSE 0 END)",
}

// sortFieldAggregateCallsOnly 无 Tokens 的聚合表排序（应用主表 / API Key 主表等）。
var sortFieldAggregateCallsOnly = map[string]string{
	"callCount":            "SUM(call_count)",
	"callFailure":          "SUM(call_failure)",
	"streamCount":          "SUM(stream_count)",
	"nonStreamCount":       "SUM(non_stream_count)",
	"streamFailure":        "SUM(stream_failure)",
	"nonStreamFailure":     "SUM(non_stream_failure)",
	"firstTokenLatency":    "SUM(first_token_latency)",
	"costs":                "SUM(costs)",
	"failureRate":          "(CASE WHEN SUM(call_count) > 0 THEN SUM(call_failure) * 1.0 / SUM(call_count) ELSE 0 END)",
	"avgFirstTokenLatency": "(CASE WHEN (SUM(stream_count) - SUM(stream_failure)) > 0 THEN SUM(first_token_latency) * 1.0 / (SUM(stream_count) - SUM(stream_failure)) ELSE 0 END)",
	"avgCosts":             "(CASE WHEN (SUM(non_stream_count) - SUM(non_stream_failure)) > 0 THEN SUM(costs) * 1.0 / (SUM(non_stream_count) - SUM(non_stream_failure)) ELSE 0 END)",
}

// sortFieldAggregateFromRecord 钻取列表排序：数据源是调用明细表、现场 GROUP BY 聚合
// （如 API Key→应用钻取），无 call_count 等预聚合列，须用 COUNT/CASE 现算。
// 注意：调用明细（record list）本身不开放排序，勿与本白名单混淆。
var sortFieldAggregateFromRecord = map[string]string{
	"callCount":            "COUNT(*)",
	"callFailure":          "SUM(CASE WHEN status_code != 200 THEN 1 ELSE 0 END)",
	"streamCount":          "SUM(CASE WHEN is_stream = 1 THEN 1 ELSE 0 END)",
	"nonStreamCount":       "SUM(CASE WHEN is_stream = 0 THEN 1 ELSE 0 END)",
	"streamFailure":        "SUM(CASE WHEN is_stream = 1 AND status_code != 200 THEN 1 ELSE 0 END)",
	"nonStreamFailure":     "SUM(CASE WHEN is_stream = 0 AND status_code != 200 THEN 1 ELSE 0 END)",
	"firstTokenLatency":    "SUM(first_token_latency)",
	"costs":                "SUM(costs)",
	"failureRate":          "(CASE WHEN COUNT(*) > 0 THEN SUM(CASE WHEN status_code != 200 THEN 1 ELSE 0 END) * 1.0 / COUNT(*) ELSE 0 END)",
	"avgFirstTokenLatency": "(CASE WHEN SUM(CASE WHEN is_stream = 1 AND status_code = 200 THEN 1 ELSE 0 END) > 0 THEN SUM(CASE WHEN status_code = 200 THEN first_token_latency ELSE 0 END) * 1.0 / SUM(CASE WHEN is_stream = 1 AND status_code = 200 THEN 1 ELSE 0 END) ELSE 0 END)",
	"avgCosts":             "(CASE WHEN SUM(CASE WHEN is_stream = 0 AND status_code = 200 THEN 1 ELSE 0 END) > 0 THEN SUM(CASE WHEN status_code = 200 THEN costs ELSE 0 END) * 1.0 / SUM(CASE WHEN is_stream = 0 AND status_code = 200 THEN 1 ELSE 0 END) ELSE 0 END)",
}

// resolveSortExpr 按白名单把 sortField 映射为安全 SQL 列表达式，未命中返回空。
func resolveSortExpr(whitelist map[string]string, sortField string) string {
	if expr, ok := whitelist[sortField]; ok {
		return expr
	}
	return ""
}
