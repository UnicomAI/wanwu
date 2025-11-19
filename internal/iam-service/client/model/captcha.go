package model

// Captcha 验证码 [EN] Captcha verification code
type Captcha struct {
	// 客户端key [EN] client key
	ID string `gorm:"primaryKey"`
	// 验证码 [EN] Verification code
	Code string
	// 本次验证码交互开始时间，并不是当前验证码的创建时间 [EN] The start time of this verification code interaction is not the creation time of the current verification code.
	StartAt int64
	// 当前验证码的创建时间 [EN] The creation time of the current verification code
	RefreshAt int64
	// 从start_at开始的刷新次数 [EN] Number of refreshes starting from start_at
	RefreshCnt int32
}
