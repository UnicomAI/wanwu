package model

// BuiltinSkill 内置skill下载统计
type BuiltinSkill struct {
	ID            uint32 `gorm:"primarykey"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli"`
	SkillId       string `gorm:"column:skill_id;size:64;uniqueIndex;comment:内置skill id"`
	DownloadCount int32  `gorm:"column:download_count;default:0;not null;comment:下载次数"`
}
