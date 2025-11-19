package model

// Captcha verification code
type Captcha struct {
	// client key
	ID string `gorm:"primaryKey"`
	// Verification code
	Code string
	// The start time of this verification code interaction is not the creation time of the current verification code.
	StartAt int64
	// The creation time of the current verification code
	RefreshAt int64
	// Number of refreshes starting from start_at
	RefreshCnt int32
}
