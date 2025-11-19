package response

type AgentScopeWorkFlowInfo struct {
	Id           string `json:"id"`           // application id
	ConfigDesc   string `json:"configDesc"`   // Application introduction
	ConfigENName string `json:"configENName"` // App English name
	ConfigName   string `json:"configName"`   // Application name
	ExampleFlag  int    `json:"example_flag"` // Example identification
	IsStream     int    `json:"is_stream"`    // Streaming Identity
	OrgID        string `json:"orgID"`        // Organization ID
	Status       string `json:"status"`       // Application status
	UpdatedTime  string `json:"updatedTime"`  // App update time
	UserID       string `json:"userID"`       // User ID
}

type AgentScopeWorkFlowListResp struct {
	Code    int                           `json:"code"`
	Message string                        `json:"msg"`
	Data    *AgentScopeWorkFlowPageResult `json:"data"`
}

type AgentScopeWorkFlowPageResult struct {
	List     []AgentScopeWorkFlowInfo `json:"list"`
	Total    int64                    `json:"total"`
	PageNo   int                      `json:"pageNo"`
	PageSize int                      `json:"pageSize"`
}

type AgentScopeDeleteWorkFlowResp struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}
