package model

type App struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_app_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// User ID
	UserID string `gorm:"index:idx_app_user_id"`
	// Organization ID
	OrgID string `gorm:"index:idx_app_org_id"`
	// APP ID
	AppID string `gorm:"index:idx_app_app_id"`
	// Application type
	AppType string `gorm:"index:idx_app_app_type"`
	// Release type
	PublishType string `gorm:"index:idx_app_publish_type"`
}
