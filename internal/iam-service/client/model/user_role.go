package model

type UserRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// Organization ID
	OrgID uint32 `gorm:"index:idx_user_role_org_id"`
	// User ID
	UserID uint32 `gorm:"primaryKey;index:idx_user_role_user_id;autoIncrement:false"`
	// Role
	RoleID uint32 `gorm:"primaryKey;index:idx_user_role_role_id;autoIncrement:false"`
	// Whether the organization has built-in administrator roles
	IsAdmin bool `gorm:"index:idx_user_role_is_admin"`
}
