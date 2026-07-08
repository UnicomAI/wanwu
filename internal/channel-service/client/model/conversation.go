package model

// ChannelConversation 通道会话映射
// 维护 (channelID + platformUserID + appType) -> conversationId 的映射，持久化到 DB
// appType=agent 时 ConversationID 为万悟 conversationId；appType=wga 时为 WGA threadId
type ChannelConversation struct {
	ID             uint32 `gorm:"primary_key"`
	CreatedAt      int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt      int64  `gorm:"autoUpdateTime:milli"`
	ChannelID      string `gorm:"uniqueIndex:idx_cc_ch_user_app;size:64"`
	UserID         string `gorm:"uniqueIndex:idx_cc_ch_user_app;size:64"`
	AppType        string `gorm:"uniqueIndex:idx_cc_ch_user_app;size:32"` // agent/wga
	ConversationID string `gorm:"size:128"`                               // agent=conversationId / wga=threadId
}

// TableName 表名
func (ChannelConversation) TableName() string {
	return "channel_conversations"
}
