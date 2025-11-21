package model

type User struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// Whether the system has a built-in administrator (currently it only indicates whether the system is the only built-in user)
	IsAdmin bool `gorm:"index:idx_user_is_admin"`
	// state
	Status bool `gorm:"index:idx_user_status"`
	// Creator
	CreatorID uint32 `gorm:"index:idx_user_creator_id"`
	// username
	Name string `gorm:"index:idx_user_name"`
	// Nick name
	Nick string `gorm:"index:idx_user_nick"`
	// gender
	Gender string `gorm:"index:idx_user_gender"`
	// Telephone
	Phone string `gorm:"index:idx_user_phone"`
	// Mail
	Email string `gorm:"index:idx_user_email"`
	// company
	Company string `gorm:"index:idx_user_company"`
	// Remark
	Remark string
	// password
	Password string
	// Last password update time
	LastUpdatePasswordAt int64
	// Last login time (millisecond timestamp)
	LastLoginAt int64
	// The latest token validity time (millisecond timestamp, previously generated tokens are invalid)
	LastTokenAt int64
	// Last operation time (millisecond timestamp)
	LastExecAt int64
	// User language
	Language string `gorm:"index:idx_user_language"`
	// User avatar
	AvatarPath string `gorm:"index:idx_user_avatar_path"`
	// isEmailCheck
	IsEmailCheck bool `gorm:"default:false;not null"`
}
