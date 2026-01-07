package model

import "github.com/UnicomAI/wanwu/pkg/db"

type SystemCustom struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 组织ID
	OrgID string `gorm:"index:idx_system_custom_org_id"`
	// 用户ID（创建人）
	UserID string `gorm:"index:idx_system_custom_user_id"`
	// Key
	Key string `gorm:"index:idx_system_custom_key"`
	// Value
	Value db.LongText `gorm:"column:value"`
}
