package model

type AcquiredSkill struct {
	ID            uint32 `gorm:"primarykey"`
	CustomSkillID string `gorm:"column:custom_skill_id;index:idx_acquired_skill_custom_skill_id;comment:skill id"`
	UserID        string `gorm:"column:user_id;index:idx_acquired_skill_user_id;comment:用户id"`
	OrgID         string `gorm:"column:org_id;index:idx_acquired_skill_org_id;comment:组织id"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
