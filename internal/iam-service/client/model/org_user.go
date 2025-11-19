package model

type OrgUser struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// 组织ID [EN] Organization ID
	OrgID uint32 `gorm:"primaryKey;index:idx_org_user_org_id;autoIncrement:false"`
	// 用户ID [EN] User ID
	UserID uint32 `gorm:"primaryKey;index:idx_org_user_user_id;autoIncrement:false"`
	// 状态 [EN] state
	Status string `gorm:"index:idx_org_user_status"`
}
