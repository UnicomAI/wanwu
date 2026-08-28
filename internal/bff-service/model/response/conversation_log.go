package response

type ConversationLogInfo struct {
	LogId                string  `json:"logId"`                // 日志Id
	Source               string  `json:"source"`               // 来源
	UserId               string  `json:"userId"`               // 使用者ID
	UserName             string  `json:"userName"`             // 使用者名称
	Title                string  `json:"title"`                // 会话标题
	ConversationId       string  `json:"conversationId"`       // 会话ID
	MessageCount         int64   `json:"messageCount"`         // 消息数量
	CreateAt             string  `json:"createAt"`             // 创建时间
	UpdateAt             string  `json:"updateAt"`             // 最后对话时间
	AvgCosts             float64 `json:"avgCosts"`             // 平均耗时
	AvgFirstTokenLatency float64 `json:"avgFirstTokenLatency"` // 平均首token时延
	LikeCount            int64   `json:"likeCount"`            // 点赞数量
	DislikeCount         int64   `json:"dislikeCount"`         // 点踩数量
	ErrorCount           int64   `json:"errorCount"`           // 错误数量
	Version              string  `json:"version"`              // 版本
}

// ConversationLogExportRecordResp 对话日志导出记录
type ConversationLogExportRecordResp struct {
	ExportRecordId string `json:"exportRecordId"` // 导出记录id
	ExportTime     string `json:"exportTime"`     // 导出时间
	FilePath       string `json:"filePath"`       // 导出文件下载路径
	FileName       string `json:"fileName"`       // 导出文件名
	Author         string `json:"author"`         // 导出人
	Status         int    `json:"status"`         // 状态
	ErrorMsg       string `json:"errorMsg"`       // 导出状态错误信息
}
