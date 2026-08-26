package model

const (
	ConversationMarkOld = 1
)

type Conversation struct {
	ID               uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:对话Id"`
	AssistantId      uint32 `gorm:"column:assistant_id;index:idx_conversation_assistant_id;type:bigint(20);not null;comment:'智能体id'"`
	ConversationId   string `gorm:"uniqueIndex:idx_unique_conversation_id;column:conversation_id;type:varchar(64)" json:"conversationId"` // Business Primary Key
	ConversationMark uint32 `gorm:"column:conversation_mark;type:tinyint;not null;default:0;comment:'是否是补充的conversationID，0：不是，1：是（是否是补充的conversationID后增加的字段）'"`
	Title            string `gorm:"column:title;type:text;comment:'对话标题'"`
	ConversationType string `gorm:"column:conversation_type;index:idx_conversation_conversation_type;type:varchar(64);comment:对话类型" `
	UserId           string `gorm:"column:user_id;index:idx_conversation_user_id;comment:用户id"`
	OrgId            string `gorm:"column:org_id;index:idx_conversation_org_id;comment:组织id"`
	CreatedAt        int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt        int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
