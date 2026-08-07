package model

// APIKeyRecordV2 API Key 调用明细表（V2）
type APIKeyRecordV2 struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_api_key_record_v2_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`

	OrgID               string `gorm:"size:64;index:idx_api_key_record_v2_org_id"`
	UserID              string `gorm:"size:64;index:idx_api_key_record_v2_user_id"`
	APIKeyID            string `gorm:"size:64;index:idx_api_key_record_v2_api_key_id"`
	MethodPath          string `gorm:"size:128;index:idx_api_key_record_v2_method_path"`
	TraceID             string `gorm:"size:128;index:idx_api_key_record_v2_trace_id"`
	Source              string `gorm:"size:32;index:idx_api_key_record_v2_source"`
	Module              string `gorm:"size:64;index:idx_api_key_record_v2_module"`
	ModuleCreatorUserID string `gorm:"size:64;index:idx_api_key_record_v2_module_creator_user_id"`
	ModuleCreatorOrgID  string `gorm:"size:64;index:idx_api_key_record_v2_module_creator_org_id"`
	AppID               string `gorm:"size:64;index:idx_api_key_record_v2_app_id"`
	AppType             string `gorm:"size:64;index:idx_api_key_record_v2_app_type"`

	IsStream          bool  `gorm:"index:idx_api_key_record_v2_is_stream"`
	Costs             int64 `gorm:"index:idx_api_key_record_v2_costs"`
	FirstTokenLatency int64 `gorm:"index:idx_api_key_record_v2_first_token_latency"`

	StatusCode    int64  `gorm:"index:idx_api_key_record_v2_status_code"`
	FailureReason string `gorm:"type:longtext"`
	RequestBody   string `gorm:"type:longtext"`
	ResponseBody  string `gorm:"type:longtext"`
}
