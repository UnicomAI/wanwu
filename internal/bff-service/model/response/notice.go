package response

// NoticeUnreadCountResp 未读总数 + 分类角标。
// 头像红点用 total；各 Tab 角标用 byCategory 对应类别的值。切换组织需重新拉取。
type NoticeUnreadCountResp struct {
	// Total 当前组织上下文未读总数
	Total int32 `json:"total"`
	// ByCategory key 固定为 announcement / productService / ticket；某类别为 0 时 key 仍返回
	ByCategory map[string]int32 `json:"byCategory"`
}

// NoticeItem 消息列表项
type NoticeItem struct {
	// MessageId 消息 ID（前端用它做已读/删除入参）
	MessageId string `json:"messageId"`
	// Title 标题（列表简短展示）
	Title string `json:"title"`
	// Category 消息类别：1 公告 / 2 产品服务 / 3 工单
	Category int32 `json:"category"`
	// Content 完整正文，$${超链文字}$$ 内为可点击超链文字，前端按出现顺序与 actions 匹配
	Content string `json:"content"`
	// Actions 跳转集合，按 content 中 $${}$$ 出现顺序一一对应；无超链时为空数组
	Actions []NoticeAction `json:"actions"`
	// ReceivedAt 接收时间（毫秒时间戳），列表按此倒序
	ReceivedAt int64 `json:"receivedAt"`
	// IsRead 是否已读
	IsRead bool `json:"isRead"`
}

// NoticeAction 跳转动作
type NoticeAction struct {
	// MsgType 消息类型，前端据此判断跳转地址类型：agent/workflow/chatflow/rag/skill/knowledge/model/about/custom_url
	MsgType string `json:"msgType"`
	// ActionType 动作类型：link 当前页直接跳 / blank 新页面打开
	ActionType string `json:"actionType"`
	// ActionParams 跳转参数对象，已是对象，前端直接使用、无需 JSON.parse
	ActionParams map[string]interface{} `json:"actionParams"`
}

// NoticeDeleteResp 批量删除结果
type NoticeDeleteResp struct {
	// AffectedCount 实际删除条数（跳过的不可见/不存在 ID 不计入）
	AffectedCount int32 `json:"affectedCount"`
}
