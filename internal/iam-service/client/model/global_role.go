package model

// GlobalRole 全局角色表，独立于组织角色(OrgRole)
// 全局角色由系统管理员创建，可跨组织分配给用户
type GlobalRole struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`
	// 角色ID（联合主键）
	RoleID uint32 `gorm:"primaryKey;index:idx_global_role_role_id;autoIncrement:false"`
	// 角色名（全局唯一）
	Name string `gorm:"index:idx_global_role_name"`
	// 状态
	Status bool `gorm:"index:idx_global_role_status"`
	// 创建人ID
	CreatorID uint32 `gorm:"index:idx_global_role_creator_id"`
	// 角色头像
	AvatarPath string
}
