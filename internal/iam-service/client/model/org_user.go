package model

type OrgUser struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// Organization ID
	OrgID uint32 `gorm:"primaryKey;index:idx_org_user_org_id;autoIncrement:false"`
	// User ID
	UserID uint32 `gorm:"primaryKey;index:idx_org_user_user_id;autoIncrement:false"`
	// state
	Status string `gorm:"index:idx_org_user_status"`
}
