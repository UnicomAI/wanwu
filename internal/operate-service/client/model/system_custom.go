package model

type SystemCustom struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// Organization ID
	OrgID string `gorm:"index:idx_system_custom_org_id"`
	// User ID (creator)
	UserID string `gorm:"index:idx_system_custom_user_id"`
	// Key
	Key string `gorm:"index:idx_system_custom_key"`
	// Value
	Value string `gorm:"type:longtext"`
}
