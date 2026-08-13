package model

// UserStatus 消息的用户状态表（稀疏存储：仅记操作过的，无行=未读未删）。
// 单条已读/软删逐行 upsert；一键已读不写本表（走 ReadWatermark 水位线）。
// IsRead 与 IsDeleted 正交，互不覆盖。
type UserStatus struct {
	ID        uint32 `gorm:"primary_key"`
	UserID    string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_user_status_user_org_msg,priority:1"`
	OrgID     string `gorm:"type:varchar(64);not null;default:'';uniqueIndex:uk_notice_user_status_user_org_msg,priority:2"`
	MessageID int64 `gorm:"not null;uniqueIndex:uk_notice_user_status_user_org_msg,priority:3"`
	IsRead    bool   `gorm:"not null;default:false"`
	// 0=未读过（非 NULL 语义，与 GORM 零值一致）
	ReadAt    int64 `gorm:"not null;default:0"`
	IsDeleted bool  `gorm:"not null;default:false"`
	// 0=未删过
	DeletedAt int64 `gorm:"not null;default:0"`
}

func (UserStatus) TableName() string { return "notice_user_status" }
