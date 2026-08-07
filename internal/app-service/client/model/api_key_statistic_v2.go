package model

// StatisticApiKey API Key 统计聚合表（按日聚合，API 请求粒度）。
// call_count = API 请求次数（从 APIKeyRecordV2 聚合）。
// 不含 Token 字段——API Key 维度的 Token 从 StatisticModel 按 api_key / method_path GROUP BY 查。
// FirstTokenLatency / Costs 为成功调用的累加值（ms），读侧再除以成功次数得到均值。

type StatisticApiKey struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_statistic_api_key_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID      string `gorm:"size:64;uniqueIndex:idx_statistic_api_key_unique,priority:1"`
	UserID     string `gorm:"size:64;uniqueIndex:idx_statistic_api_key_unique,priority:2"`
	APIKeyID   string `gorm:"size:64;uniqueIndex:idx_statistic_api_key_unique,priority:3"`
	MethodPath string `gorm:"size:128;uniqueIndex:idx_statistic_api_key_unique,priority:4"`
	Date       string `gorm:"size:16;uniqueIndex:idx_statistic_api_key_unique,priority:5;index:idx_statistic_api_key_date"`

	CallCount         int32 `gorm:"index:idx_statistic_api_key_call_count"`
	CallFailure       int32 `gorm:"index:idx_statistic_api_key_call_failure"`
	StreamCount       int32 `gorm:"index:idx_statistic_api_key_stream_count"`
	NonStreamCount    int32 `gorm:"index:idx_statistic_api_key_non_stream_count"`
	StreamFailure     int32 `gorm:"index:idx_statistic_api_key_stream_failure"`
	NonStreamFailure  int32 `gorm:"index:idx_statistic_api_key_non_stream_failure"`
	FirstTokenLatency int64 `gorm:"index:idx_statistic_api_key_first_token_latency"` // 成功流式调用首 token 时延累加 ms
	Costs             int64 `gorm:"index:idx_statistic_api_key_costs"`               // 成功非流式调用耗时累加 ms
}
