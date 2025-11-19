package model

type User struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 是否系统内置管理员（目前只表示是否系统唯一内置用户） [EN] Whether the system has a built-in administrator (currently it only indicates whether the system is the only built-in user)
	IsAdmin bool `gorm:"index:idx_user_is_admin"`
	// 状态 [EN] state
	Status bool `gorm:"index:idx_user_status"`
	// 创建人 [EN] Creator
	CreatorID uint32 `gorm:"index:idx_user_creator_id"`
	// 用户名 [EN] username
	Name string `gorm:"index:idx_user_name"`
	// 昵称 [EN] Nick name
	Nick string `gorm:"index:idx_user_nick"`
	// 性别 [EN] gender
	Gender string `gorm:"index:idx_user_gender"`
	// 电话 [EN] Telephone
	Phone string `gorm:"index:idx_user_phone"`
	// 邮箱 [EN] Mail
	Email string `gorm:"index:idx_user_email"`
	// 公司 [EN] company
	Company string `gorm:"index:idx_user_company"`
	// 备注 [EN] Remark
	Remark string
	// 密码 [EN] password
	Password string
	// 最后更新密码时间 [EN] Last password update time
	LastUpdatePasswordAt int64
	// 最后一次登录时间（毫秒时间戳） [EN] Last login time (millisecond timestamp)
	LastLoginAt int64
	// 最新token有效时间（毫秒时间戳，此前生成的token都无效） [EN] The latest token validity time (millisecond timestamp, previously generated tokens are invalid)
	LastTokenAt int64
	// 最后一次操作时间（毫秒时间戳） [EN] Last operation time (millisecond timestamp)
	LastExecAt int64
	// 用户语言 [EN] User language
	Language string `gorm:"index:idx_user_language"`
	// 用户头像 [EN] User avatar
	AvatarPath string `gorm:"index:idx_user_avatar_path"`
	// isEmailCheck
	IsEmailCheck bool `gorm:"default:false;not null"`
}
