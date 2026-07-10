package model

type OrgRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// 组织ID
	OrgID uint32 `gorm:"primaryKey;index:idx_org_role_org_id;autoIncrement:false"`
	// 角色
	RoleID uint32 `gorm:"primaryKey;index:idx_org_role_role_id;autoIncrement:false"`
	// 是否组织内置管理员角色（组织内唯一）
	IsAdmin bool `gorm:"index:idx_org_role_is_admin"`
	// 状态
	Status bool `gorm:"index:idx_org_role_status"`
	// 角色名
	Name string `gorm:"index:idx_org_role_name"`
	// 角色头像
	AvatarPath string
}
