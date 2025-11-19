package model

type OauthApp struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// Register the user ID of this client
	UserID uint32 `gorm:"index:idx_oauth_app_user_id"`
	// Client application name
	Name string `gorm:"index:idx_oauth_app_name"`
	// client unique identifier
	ClientID string `gorm:"index:idx_oauth_app_client_id"`
	// client key
	ClientSecret string
	// OAuth callback address (redirect_uri)
	RedirectURI string
	// Status (enabled/disabled)
	Status bool `gorm:"index:idx_oauth_app_status"`
	// Application description
	Description string
}
