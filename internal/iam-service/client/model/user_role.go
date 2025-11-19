package model

type UserRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// 组织ID [EN] Organization ID
	OrgID uint32 `gorm:"index:idx_user_role_org_id"`
	// 用户ID [EN] User ID
	UserID uint32 `gorm:"primaryKey;index:idx_user_role_user_id;autoIncrement:false"`
	// 角色 [EN] Role
	RoleID uint32 `gorm:"primaryKey;index:idx_user_role_role_id;autoIncrement:false"`
	// 是否组织内置管理员角色 [EN] Whether the organization has built-in administrator roles
	IsAdmin bool `gorm:"index:idx_user_role_is_admin"`
}
