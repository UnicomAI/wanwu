package model

type AppFavorite struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_app_favorite_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 用户ID [EN] User ID
	UserID string `gorm:"index:idx_app_favorite_user_id"`
	// APP ID
	AppID string `gorm:"index:idx_app_favorite_app_id"`
	// 应用类型 [EN] Application type
	AppType string `gorm:"index:idx_app_favorite_app_type"`
}
