package model

// AssistantPublish 智能体发布状态关联表
type AssistantPublish struct {
	ID          uint32 `gorm:"primarykey;column:id;comment:主键Id"`
	AssistantID uint32 `gorm:"column:assistant_id;uniqueIndex:idx_unique_assistant_publish;comment:智能体Id"`
	PublishType string `gorm:"column:publish_type;type:varchar(32);comment:发布类型(private/organization/public)"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}

// AdminAssistantItem 管理员列表查询结果项，包含智能体及其发布类型
type AdminAssistantItem struct {
	Assistant
	PublishType string `gorm:"-"` // 发布类型(private/organization/public)，空=草稿
}
