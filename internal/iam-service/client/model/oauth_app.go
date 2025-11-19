package model

type OauthApp struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 注册该客户端的用户ID [EN] Register the user ID of this client
	UserID uint32 `gorm:"index:idx_oauth_app_user_id"`
	// 客户端应用名称 [EN] Client application name
	Name string `gorm:"index:idx_oauth_app_name"`
	// 客户端唯一标识符 [EN] client unique identifier
	ClientID string `gorm:"index:idx_oauth_app_client_id"`
	// 客户端密钥 [EN] client key
	ClientSecret string
	// OAuth回调地址（redirect_uri） [EN] OAuth callback address (redirect_uri)
	RedirectURI string
	// 状态（启用/禁用） [EN] Status (enabled/disabled)
	Status bool `gorm:"index:idx_oauth_app_status"`
	// 应用描述 [EN] Application description
	Description string
}
