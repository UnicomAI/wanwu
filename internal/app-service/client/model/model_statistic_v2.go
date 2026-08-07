package model

// StatisticModel 模型调用统计聚合表（按日聚合，模型调用粒度）。
// 含全维度（model + app + source + module + openapi 的 apiKey/methodPath），是唯一的 Token 数据来源；
// StatisticApp（按 app_id）/ StatisticApiKey（按 api_key + method_path）的 Token 查询需跨表走本表 GROUP BY。
// FirstTokenLatency / Costs 为成功调用的累加值（ms），读侧再除以成功次数得到均值。

type StatisticModel struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_statistic_model_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 唯一键（每日维度聚合，与 Redis field 11 维 + date 对齐；非 OpenAPI 时 api_key/method_path 为空字符串）
	ModelID string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:1"`
	Date    string `gorm:"size:16;uniqueIndex:idx_statistic_model_unique,priority:2;index:idx_statistic_model_date"` // yyyy-mm-dd
	// 调用人
	UserID string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:3;index:idx_statistic_model_user_id"`
	OrgID  string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:4;index:idx_statistic_model_org_id"`
	// 调用方式
	Source  string `gorm:"size:32;uniqueIndex:idx_statistic_model_unique,priority:5;index:idx_statistic_model_source"` // web | webURL | openapi
	Module  string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:6;index:idx_statistic_model_module"` // wga | skill | model | knowledge | prompt | app | other
	AppID   string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:7;index:idx_statistic_model_app_id"`
	AppType string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:8;index:idx_statistic_model_app_type"`
	// OpenAPI 调用标识（明文 apiKey + 路径；非 OpenAPI 为空，仍参与唯一键）
	APIKey     string `gorm:"size:64;uniqueIndex:idx_statistic_model_unique,priority:9;index:idx_statistic_model_api_key"`
	MethodPath string `gorm:"size:128;uniqueIndex:idx_statistic_model_unique,priority:10;index:idx_statistic_model_method_path"`
	// 应用 / 资源归属
	ModuleCreatorUserID string `gorm:"size:32;uniqueIndex:idx_statistic_model_unique,priority:11;index:idx_statistic_model_module_creator_user_id"`
	ModuleCreatorOrgID  string `gorm:"size:32;uniqueIndex:idx_statistic_model_unique,priority:12;index:idx_statistic_model_module_creator_org_id"`
	// 模型信息（创建后不变，冗余存储便于查询）
	Provider           string `gorm:"size:64;index:idx_statistic_model_provider"`
	Model              string `gorm:"size:128;index:idx_statistic_model_model"`
	ModelType          string `gorm:"size:32;index:idx_statistic_model_model_type"`
	ModelCreatorUserID string `gorm:"size:32;index:idx_statistic_model_model_creator_user_id"`
	ModelCreatorOrgID  string `gorm:"size:32;index:idx_statistic_model_model_creator_org_id"`
	// Token 用量（累加）
	PromptTokens     int64 `gorm:"index:idx_statistic_model_prompt_tokens"`
	CompletionTokens int64 `gorm:"index:idx_statistic_model_completion_tokens"`
	TotalTokens      int64 `gorm:"index:idx_statistic_model_total_tokens"`
	// 成功流式调用首 token 时延累加 ms（读侧 / 成功流式次数 = 均值）
	FirstTokenLatency int64 `gorm:"index:idx_statistic_model_first_token_latency"`
	// 成功非流式调用耗时累加 ms（读侧 / 成功非流式次数 = 均值）
	Costs int64 `gorm:"index:idx_statistic_model_costs"`
	// 调用次数
	CallCount      int32 `gorm:"index:idx_statistic_model_call_count"`
	StreamCount    int32 `gorm:"index:idx_statistic_model_stream_count"`
	NonStreamCount int32 `gorm:"index:idx_statistic_model_non_stream_count"`
	// 失败次数
	CallFailure      int32 `gorm:"index:idx_statistic_model_call_failure"`
	StreamFailure    int32 `gorm:"index:idx_statistic_model_stream_failure"`
	NonStreamFailure int32 `gorm:"index:idx_statistic_model_non_stream_failure"`
}
