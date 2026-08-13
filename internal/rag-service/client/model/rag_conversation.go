package model

// RagConversation 知识问答会话，问答明细按 conversation_id 落 ES
type RagConversation struct {
	ID               int64  `json:"id" gorm:"primaryKey;type:bigint(20);autoIncrement"`
	ConversationID   string `json:"conversationId" gorm:"uniqueIndex:idx_unique_rag_conversation_id;column:conversation_id;type:varchar(255);comment:会话id"`
	RagID            string `json:"ragId" gorm:"column:rag_id;index:idx_rag_conversation_rag_id;type:varchar(255);comment:知识问答id"`
	Title            string `json:"title" gorm:"column:title;type:text;comment:会话标题"`
	ConversationType string `json:"conversationType" gorm:"column:conversation_type;index:idx_rag_conversation_type;type:varchar(64);comment:会话来源，published(已发布)/draft(草稿)"`
	PublicModel
}

func (RagConversation) TableName() string {
	return "rag_conversation"
}
