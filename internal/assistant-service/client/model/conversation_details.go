package model

type FileInfo struct {
	FileName string `gorm:"type:varchar(500)" json:"fileName"`
	FileSize int64  `gorm:"type:bigint" json:"fileSize"`
	FileUrl  string `gorm:"type:text" json:"fileUrl"`
}

type ConversationDetails struct {
	ID             uint32     `gorm:"primarykey;column:id" json:"id"`
	AssistantId    uint32     `gorm:"column:assistant_id;comment:智能体id" json:"assistantId"`
	ConversationId string     `gorm:"type:varchar(255);index" json:"conversationId"`
	Prompt         string     `gorm:"type:text" json:"prompt"`
	SysPrompt      string     `gorm:"type:text" json:"sysPrompt"`
	Response       string     `gorm:"type:text" json:"response"`
	SearchList     string     `gorm:"type:text" json:"searchList"`
	QaType         int32      `gorm:"type:int" json:"qaType"`
	FileUrl        string     `gorm:"type:text" json:"requestFileUrls"`
	FileSize       int64      `gorm:"type:bigint" json:"fileSize"`
	FileName       string     `gorm:"type:varchar(500)" json:"fileName"`
	FileInfo       []FileInfo `gorm:"type:json" json:"fileInfo"`
	UserId         string     `gorm:"type:varchar(255);index" json:"userId"`
	OrgId          string     `gorm:"type:varchar(255);index" json:"orgId"`
	CreatedAt      int64      `gorm:"autoCreateTime:milli" json:"createdAt"`
	UpdatedAt      int64      `gorm:"autoUpdateTime:milli" json:"updatedAt"`
}
