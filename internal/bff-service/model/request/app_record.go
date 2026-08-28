package request

type AppRecordRequest struct {
	UserID string `json:"userId"`

	OrgID string `json:"orgId"`

	AppID string `json:"appId"`

	AppType string `json:"appType"`

	Module string `json:"module"` // 板块 agent|rag|workflow|knowledge|...；空则按 appType 推断

	// StatusCode 调用方透传的 HTTP 语义状态码；成功传 200，失败传对应码（如 500）。
	StatusCode int64 `json:"statusCode"`

	// FailureReason 失败原因；成功时传空。
	FailureReason string `json:"failureReason"`

	IsStream bool `json:"isStream"`

	StreamCosts int64 `json:"streamCosts"`

	NonStreamCosts int64 `json:"nonStreamCosts"`

	Source string `json:"source"`

	RequestBody string `json:"requestBody"` // 整体请求 JSON；受 statistic.record_body 控制

	ResponseBody string `json:"responseBody"` // 整体响应 JSON；流式忽略；受 statistic.record_body 控制

	Question string `json:"question"` // 用户提问（精简）；始终落库

	Answer string `json:"answer"` // 回复（精简）；始终落库；流式可空
}

func (a *AppRecordRequest) Check() error { return nil }
