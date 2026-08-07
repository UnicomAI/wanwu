package model

// StatisticApp 应用统计聚合表（按日聚合，应用会话粒度）。
// call_count = 应用会话次数（从 AppRecordV2 聚合，非模型调用次数）。
// 不含 Token 字段——AppRecordV2 无 Token；App 维度 Token 从 StatisticModel 按 app_id GROUP BY 查。
// FirstTokenLatency / Costs 为成功调用的累加值（ms），读侧再除以成功次数得到均值。

type StatisticApp struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_statistic_app_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID               string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:1"`
	UserID              string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:2"`
	Source              string `gorm:"size:32;uniqueIndex:idx_statistic_app_unique,priority:3;index:idx_statistic_app_source"`
	Module              string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:4;index:idx_statistic_app_module"`
	AppID               string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:5;index:idx_statistic_app_app_id"`
	AppType             string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:6;index:idx_statistic_app_app_type"`
	ModuleCreatorUserID string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:7;index:idx_statistic_app_module_creator_user_id"`
	ModuleCreatorOrgID  string `gorm:"size:64;uniqueIndex:idx_statistic_app_unique,priority:8;index:idx_statistic_app_module_creator_org_id"`
	Date                string `gorm:"size:16;uniqueIndex:idx_statistic_app_unique,priority:9;index:idx_statistic_app_date"`

	CallCount         int32 `gorm:"index:idx_statistic_app_call_count"`
	CallFailure       int32 `gorm:"index:idx_statistic_app_call_failure"`
	StreamCount       int32 `gorm:"index:idx_statistic_app_stream_count"`
	NonStreamCount    int32 `gorm:"index:idx_statistic_app_non_stream_count"`
	StreamFailure     int32 `gorm:"index:idx_statistic_app_stream_failure"`
	NonStreamFailure  int32 `gorm:"index:idx_statistic_app_non_stream_failure"`
	FirstTokenLatency int64 `gorm:"index:idx_statistic_app_first_token_latency"` // 成功流式调用首 token 时延累加 ms
	Costs             int64 `gorm:"index:idx_statistic_app_costs"`               // 成功非流式调用耗时累加 ms
}
