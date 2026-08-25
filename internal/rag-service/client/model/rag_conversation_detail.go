package model

type RagFileInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileUrl  string `json:"fileUrl"`
}

// RagTimeCost 一个阶段的起止时刻(毫秒)，两个字段均为0表示该阶段未发生
type RagTimeCost struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// RagStatistic 一轮问答的耗时与token统计
type RagStatistic struct {
	StartTime         int64 `json:"startTime"`         // 请求开始时间戳(毫秒)
	FirstTokenLatency int64 `json:"firstTokenLatency"` // 首token时延(毫秒)，整轮无输出时为0
	TotalCostTime     int64 `json:"totalCostTime"`     // 总耗时(毫秒)
	PromptTokens      int64 `json:"promptTokens"`
	CompletionTokens  int64 `json:"completionTokens"`
	TotalTokens       int64 `json:"totalTokens"`
	// 三个阶段的起止时刻，字段名与详情接口的 reasoningContent/searchList/qaSearchList 对应。
	ReasoningTimeCost    RagTimeCost `json:"reasoningTimeCost"`
	SearchListTimeCost   RagTimeCost `json:"searchListTimeCost"`
	QaSearchListTimeCost RagTimeCost `json:"qaSearchListTimeCost"`
}

// RagConversationDetail 知识问答单轮问答明细，按月落ES索引 rag_conversation_detail_infos_YYYYMM
// ReasoningContent、SearchList、QaSearchList 同一规则：nil=该阶段未发生、空串=发生过但没产出、非空=内容
type RagConversationDetail struct {
	Id               string         `json:"id"`    // 取本轮问答的msgId，与返回给调用方的msg_id一致
	RagId            string         `json:"ragId"` //
	ConversationId   string         `json:"conversationId"`
	Publish          int32          `json:"publish"` // 0草稿 1已发布
	Prompt           string         `json:"prompt"`
	Response         string         `json:"response"`
	ReasoningContent *string        `json:"reasoningContent"`
	SearchList       *string        `json:"searchList"`   // 知识库引用段落原始json
	QaSearchList     *string        `json:"qaSearchList"` // 问答库命中结果原始json
	ErrMessage       string         `json:"errMessage"`
	QaErrMessage     string         `json:"qaErrMessage"` // 问答库检索失败原因，非空表示这轮是问答库报错后转的知识库
	FileInfo         []*RagFileInfo `json:"fileInfo"`
	TraceId          string         `json:"traceId"` // 本轮问答的链路id，与应用/模型统计明细同值，可据此反查调用链
	UserId           string         `json:"userId"`
	OrgId            string         `json:"orgId"`
	CreatedAt        int64          `json:"createdAt"`
	UpdatedAt        int64          `json:"updatedAt"`
	Statistic        RagStatistic   `json:"statistic"`
	Feedback         int32          `json:"feedback"`        // 当前反馈状态: 0=无 1=点赞 2=点踩
	FeedbackContent  string         `json:"feedbackContent"` // 反馈文本内容
}

const (
	FeedBackNone    int32 = 0 // 无反馈
	FeedBackLike    int32 = 1 // 点赞
	FeedBackDislike int32 = 2 // 点踩
)
