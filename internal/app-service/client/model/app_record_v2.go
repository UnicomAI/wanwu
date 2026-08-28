package model

// AppRecordV2 应用调用记录明细表
type AppRecordV2 struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_app_record_v2_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	TraceID string `gorm:"size:128;index:idx_app_record_v2_trace_id"`

	OrgID  string `gorm:"size:64;index:idx_app_record_v2_org_id"`  // 使用人所属组织
	UserID string `gorm:"size:64;index:idx_app_record_v2_user_id"` // 使用人

	Source              string `gorm:"size:32;index:idx_app_record_v2_source"`
	Module              string `gorm:"size:64;index:idx_app_record_v2_module"`
	ModuleCreatorUserID string `gorm:"size:64;index:idx_app_record_v2_module_creator_user_id"` // 应用作者
	ModuleCreatorOrgID  string `gorm:"size:64;index:idx_app_record_v2_module_creator_org_id"`  // 应用所属组织

	AppID   string `gorm:"size:64;index:idx_app_record_v2_app_id"`
	AppType string `gorm:"size:64;index:idx_app_record_v2_app_type"` // 应用类型

	IsStream bool `gorm:"index:idx_app_record_v2_is_stream"`

	Costs             int64 `gorm:"index:idx_app_record_v2_costs"`               // 非流式耗时 ms
	FirstTokenLatency int64 `gorm:"index:idx_app_record_v2_first_token_latency"` // 流式首 token 时延 ms

	StatusCode    int64  `gorm:"index:idx_app_record_v2_status_code"` // 状态码
	RequestBody   string `gorm:"type:longtext"`                       // 输入内容
	ResponseBody  string `gorm:"type:longtext"`                       // 输出内容
	FinishReason  string `gorm:"type:longtext"`                       // 完成原因
	FailureReason string `gorm:"type:longtext"`                       // 失败原因
	Question      string `gorm:"type:longtext"`                       // 用户问题（精简）
	Answer        string `gorm:"type:longtext"`                       // 回复（精简；本阶段流式可为空）
}
