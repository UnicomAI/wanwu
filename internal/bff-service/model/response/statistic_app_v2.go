package response

// AppStatisticV2Chart 趋势 + 排行（对齐 UI 图1）
type AppStatisticV2Chart struct {
	Trend AppStatisticV2Trend `json:"trend"`
	Rank  AppStatisticV2Rank  `json:"rank"`
}

// AppStatisticV2Trend 趋势图表
type AppStatisticV2Trend struct {
	CallResult StatisticChart `json:"callResult"` // 应用调用次数统计表（成功/失败堆叠柱）
	CallTrend  StatisticChart `json:"callTrend"`  // 应用调用趋势表（总数/Web/OpenAPI/WebURL 折线）
}

// AppStatisticV2Rank 应用排行（按应用类型分榜单）
type AppStatisticV2Rank struct {
	ByAgent           []StatisticV2RankItem `json:"byAgent"`           // 智能体使用量排行
	ByWorkflow        []StatisticV2RankItem `json:"byWorkflow"`        // 工作流使用排行
	ByChatflow        []StatisticV2RankItem `json:"byChatflow"`        // 对话流使用排行（图1 红字新增）
	ByRag             []StatisticV2RankItem `json:"byRag"`             // 知识问答使用排行
	ByDigitalEmployee []StatisticV2RankItem `json:"byDigitalEmployee"` // 数字员工使用量排行
}

// AppStatisticV2ListItem 调用统计主表行（聚合维度；耗时为 avgCosts / avgFirstTokenLatency）
type AppStatisticV2ListItem struct {
	ModuleBriefInfo
	StatisticV2Metrics
}

// AppStatisticV2UserListItem 按用户钻取行（对齐 UI 图4）
type AppStatisticV2UserListItem struct {
	ModuleBriefInfo
	UserBriefInfo
	StatisticV2Metrics
}

// AppStatisticV2ModelListItem 按模型钻取行（对齐 UI 图5）
// 模型前缀复用 ModelBriefInfo（含图5 红字新增的模型名称/供应商/类型/发布者/所属组织）。
type AppStatisticV2ModelListItem struct {
	ModuleBriefInfo
	ModelBriefInfo
	StatisticV2Metrics
}

// StatisticV2AppRecordBase 应用调用明细/详情公共基础（Id/时间/状态/应用 / 使用人）。
type StatisticV2AppRecordBase struct {
	Id         uint32 `json:"id"`
	TraceId    string `json:"traceId"`
	CalledAt   string `json:"calledAt"`
	IsSuccess  bool   `json:"isSuccess"`
	StatusCode int64  `json:"statusCode"`
	ModuleBriefInfo
	UserBriefInfo
}

// AppStatisticV2RecordItem 调用明细列表行（单次调用；耗时为 firstTokenLatency / costs，非平均）
type AppStatisticV2RecordItem struct {
	StatisticV2AppRecordBase
	StatisticV2RecordPerformance
	FailureReason string `json:"failureReason"`
	RequestBody   string `json:"requestBody"`  // 输入内容
	ResponseBody  string `json:"responseBody"` // 输出内容
	Question      string `json:"question"`     // 用户问题（精简）；导出用
	Answer        string `json:"answer"`       // 回复（精简）；导出用；流式底层强制为空
}
