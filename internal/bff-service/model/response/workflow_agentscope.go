package response

type AgentScopeWorkFlowInfo struct {
	Id           string `json:"id"`           // 应用id [EN] application id
	ConfigDesc   string `json:"configDesc"`   // 应用简介 [EN] Application introduction
	ConfigENName string `json:"configENName"` // 应用英文名称 [EN] App English name
	ConfigName   string `json:"configName"`   // 应用名称 [EN] Application name
	ExampleFlag  int    `json:"example_flag"` // 示例标识 [EN] Example identification
	IsStream     int    `json:"is_stream"`    // 流式标识 [EN] Streaming Identity
	OrgID        string `json:"orgID"`        // 组织ID [EN] Organization ID
	Status       string `json:"status"`       // 应用状态 [EN] Application status
	UpdatedTime  string `json:"updatedTime"`  // 应用更新时间 [EN] App update time
	UserID       string `json:"userID"`       // 用户ID [EN] User ID
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
