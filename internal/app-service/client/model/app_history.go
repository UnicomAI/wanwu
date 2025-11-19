package model

type AppHistory struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_app_history_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli;index:idx_app_history_updated_at"`
	// User ID
	UserID string `gorm:"index:idx_app_history_user_id"`
	// APP ID
	AppID string `gorm:"index:idx_app_history_app_id"`
	// Application type
	AppType string `gorm:"index:idx_app_history_app_type"`
}
