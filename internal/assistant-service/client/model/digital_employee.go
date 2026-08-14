package model

// DigitalEmployeeConversationConfig 数字员工发布会话表（独立于 wga_conversation_config）。
// 与 WgaConversationConfig 的区别：多 employee_id 绑定列；thread_id 即 wga threadId，对话执行复用 wga thread 体系。
// 不可切换约束由业务层保证：employee_id 创建即固定，无更新接口。
type DigitalEmployeeConversationConfig struct {
	ID          uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:配置Id"`
	ThreadID    string `gorm:"column:thread_id;uniqueIndex:idx_de_conv_id;type:varchar(128);not null;comment:会话ID(即wga threadId)"`
	EmployeeID  string `gorm:"column:employee_id;index:idx_de_conv_employee;type:varchar(128);not null;comment:数字员工ID(外部)"`
	Title       string `gorm:"column:title;type:text;comment:会话标题"`
	ModelConfig string `gorm:"column:model_config;type:json;comment:模型配置JSON"`
	UserID      string `gorm:"column:user_id;index:idx_de_conv_user;type:varchar(64);not null;comment:用户id"`
	OrgID       string `gorm:"column:org_id;index:idx_de_conv_org;type:varchar(64);not null;comment:组织id"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
