package model

type Org struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 状态
	Status bool `gorm:"index:idx_org_status"`
	// 创建人
	CreatorID uint32
	// 上级组织id（0表示一级组织）
	ParentID uint32 `gorm:"index:idx_org_parent_id"`
	// 组织名（上级组织的所有下级中唯一）
	Name string `gorm:"index:idx_org_name"`
	// 描述
	Remark string
	// 组织头像
	AvatarPath string
}
