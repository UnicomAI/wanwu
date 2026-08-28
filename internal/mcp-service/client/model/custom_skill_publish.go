package model

type CustomSkillPublish struct {
	ID                 int64  `gorm:"primaryKey;column:id;type:bigint(20);autoIncrement;comment:主键id" json:"id"`
	CreatedAt          int64  `gorm:"column:created_at;type:bigint(20);autoCreateTime:milli;not null;comment:创建时间"`
	UpdatedAt          int64  `gorm:"column:updated_at;type:bigint(20);autoUpdateTime:milli;not null;comment:更新时间"`
	UserId             string `gorm:"column:user_id;type:varchar(64);comment:用户id"`
	OrgId              string `gorm:"column:org_id;type:varchar(64);comment:组织id"`
	SkillID            string `gorm:"column:skill_id;uniqueIndex:idx_skill_id_version;type:varchar(64);not null;comment:skillId" json:"skillId" `
	Version            string `gorm:"column:version;uniqueIndex:idx_skill_id_version;type:varchar(64);not null;comment:版本号(与skillId构成唯一复合索引)"`
	VersionDescription string `gorm:"column:version_description;type:varchar(255);comment:版本描述;default:''"`
	Markdown           string `gorm:"column:markdown;type:longtext;comment:markdown内容"`
	ObjectPath         string `gorm:"column:object_path;comment:skill数据minio对象路径(zip压缩包)"`
}
