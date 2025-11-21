package model

type Org struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// state
	Status bool `gorm:"index:idx_org_status"`
	// Creator
	CreatorID uint32
	// Parent organization id (0 indicates first-level organization)
	ParentID uint32 `gorm:"index:idx_org_parent_id"`
	// Organization name (unique among all subordinates of the parent organization)
	Name string `gorm:"index:idx_org_name"`
	// describe
	Remark string
}
