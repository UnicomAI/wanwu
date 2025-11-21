package model

type OrgRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// Organization ID
	OrgID uint32 `gorm:"primaryKey;index:idx_org_role_org_id;autoIncrement:false"`
	// Role
	RoleID uint32 `gorm:"primaryKey;index:idx_org_role_role_id;autoIncrement:false"`
	// Whether the organization has a built-in administrator role (unique within the organization)
	IsAdmin bool `gorm:"index:idx_org_role_is_admin"`
	// state
	Status bool `gorm:"index:idx_org_role_status"`
	// Character name
	Name string `gorm:"index:idx_org_role_name"`
}
