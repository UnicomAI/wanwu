package model

type ConversationLogExtend struct {
	ConversationIDMask uint32 `json:"conversation_id_mark"` // 旧版本对话id
	ConversationMark   uint32 `json:"conversation_mark"`    // 是否是补充的conversationID，0：不是，1：是（是否是补充的conversationID后增加的字段）`
	TokenCosts         int64  `json:"token_costs"`
	InputTokens        int64  `json:"input_tokens"`
	OutputTokens       int64  `json:"output_tokens"`
}

type ConversationLog struct {
	ID                uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;"`
	LogId             string `gorm:"uniqueIndex:idx_unique_log_id;column:log_id;type:varchar(64);comment:对话日志id"`
	ConversationId    string `gorm:"uniqueIndex:idx_unique_conversation_log_conversation_id_type,priority:1;column:conversation_id;type:varchar(64);not null;comment:会话id"`
	AppId             string `gorm:"index:idx_conversation_log_app_id_type,priority:1;column:app_id;type:varchar(64);not null;comment:应用id:智能体、知识问答应用id"`
	AppType           string `gorm:"uniqueIndex:idx_unique_conversation_log_conversation_id_type,priority:2;index:idx_conversation_log_app_id_type,priority:2;column:app_type;type:varchar(64);not null;comment:应用类型:智能体、知识问答"`
	Source            string `gorm:"column:source;type:varchar(64);not null;default:'';comment:来源(openapi、webURL、web、draft)"`
	Version           string `gorm:"column:version;type:varchar(64);not null;default:'';comment:版本" `
	Title             string `gorm:"column:title;type:varchar(128);not null;default:'';comment:标题 >128截取 " `
	MessageCount      int    `gorm:"column:message_count;type:int(11);not null;default:0;comment:消息数量"`
	Costs             int64  `gorm:"column:costs;type:bigint(20);not null;default:0;comment:平均耗时（非流式）"`
	FirstTokenLatency int64  `gorm:"column:first_token_latency;type:bigint(20);not null;default:0;comment:平均首token时延 (流式)"`
	LikeCount         int    `gorm:"column:like_count;type:int(11);not null;default:0;comment:点赞数量"`
	DisLikeCount      int    `gorm:"column:dislike_count;type:int(11);not null;default:0;comment:点踩数量"`
	ErrorCount        int    `gorm:"column:error_count;type:int(11);not null;default:0;comment:报错数量"`
	UserId            string `gorm:"index:idx_conversation_log_user_id;column:user_id;comment:用户id"`
	OrgId             string `gorm:"index:idx_conversation_log_org_id;column:org_id;comment:组织id"`
	CreatedAt         int64  `gorm:"index:idx_conversation_log_create_at;autoCreateTime:milli;comment:创建时间"`
	UpdatedAt         int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
	Ext               string `gorm:"column:ext;type:text;not null;comment: 扩展字段" `
}
