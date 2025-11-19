package model

type OrgRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// 组织ID [EN] Organization ID
	OrgID uint32 `gorm:"primaryKey;index:idx_org_role_org_id;autoIncrement:false"`
	// 角色 [EN] Role
	RoleID uint32 `gorm:"primaryKey;index:idx_org_role_role_id;autoIncrement:false"`
	// 是否组织内置管理员角色（组织内唯一） [EN] Whether the organization has a built-in administrator role (unique within the organization)
	IsAdmin bool `gorm:"index:idx_org_role_is_admin"`
	// 状态 [EN] state
	Status bool `gorm:"index:idx_org_role_status"`
	// 角色名 [EN] Character name
	Name string `gorm:"index:idx_org_role_name"`
}
