package model

// ModelRecordV2 模型调用记录明细表
type ModelRecordV2 struct {
	ID        uint64 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_model_record_v2_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// TraceID 用于链路追踪关联
	TraceID string `gorm:"size:128;index:idx_model_record_v2_trace_id"`
	// 调用人信息
	OrgID  string `gorm:"size:64;index:idx_model_record_v2_org_id"`
	UserID string `gorm:"size:64;index:idx_model_record_v2_user_id"`
	// 调用方式
	Source              string `gorm:"size:32;index:idx_model_record_v2_source"` // web | webURL | openapi
	Module              string `gorm:"size:64;index:idx_model_record_v2_module"` // wga | skill | model | knowledge | prompt | app | other
	ModuleCreatorUserID string `gorm:"size:64;index:idx_model_record_v2_module_creator_user_id"`
	ModuleCreatorOrgID  string `gorm:"size:64;index:idx_model_record_v2_module_creator_org_id"`
	// 应用 / 资源归属
	AppID   string `gorm:"size:64;index:idx_model_record_v2_app_id"`
	AppType string `gorm:"size:64;index:idx_model_record_v2_app_type"`
	// 模型信息
	ModelID            string `gorm:"size:64;index:idx_model_record_v2_model_id"`
	Model              string `gorm:"size:128;index:idx_model_record_v2_model"`
	ModelCreatorUserID string `gorm:"size:64;index:idx_model_record_v2_model_creator_user_id"`
	ModelCreatorOrgID  string `gorm:"size:64;index:idx_model_record_v2_model_creator_org_id"`
	Provider           string `gorm:"size:64;index:idx_model_record_v2_provider"`
	ModelType          string `gorm:"size:32;index:idx_model_record_v2_model_type"`
	// 是否流式调用
	IsStream bool `gorm:"index:idx_model_record_v2_is_stream"`
	// ApiKey信息
	APIKeyID   string `gorm:"size:64;index:idx_model_record_v2_api_key_id"`
	APIKey     string `gorm:"size:128;index:idx_model_record_v2_api_key"` // openapi 明文 apiKey
	MethodPath string `gorm:"size:128;index:idx_model_record_v2_method_path"`
	// Token 用量
	PromptTokens     int64 `gorm:"index:idx_model_record_v2_prompt_tokens"`
	CompletionTokens int64 `gorm:"index:idx_model_record_v2_completion_tokens"`
	TotalTokens      int64 `gorm:"index:idx_model_record_v2_total_tokens"`
	// 性能指标（ms）
	Costs             int64 `gorm:"index:idx_model_record_v2_costs"`               // 非流式总耗时
	FirstTokenLatency int64 `gorm:"index:idx_model_record_v2_first_token_latency"` // 流式首token时延
	// 结果状态
	StatusCode    int64  `gorm:"index:idx_model_record_v2_status_code"` // HTTP/模型返回状态码
	RequestBody   string `gorm:"type:longtext"`                         // 请求参数
	ResponseBody  string `gorm:"type:longtext"`                         // 响应结果
	FinishReason  string `gorm:"type:longtext"`                         // 流式调用结束原因
	FailureReason string `gorm:"type:longtext"`                         // 调用失败原因（如 provider 返回 err）
}
