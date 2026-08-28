package model

type CustomSkill struct {
	ID              uint32 `gorm:"primarykey"`
	UserID          string `gorm:"column:user_id;index:idx_custom_skill_user_id;comment:用户id"`
	OrgID           string `gorm:"column:org_id;index:idx_custom_skill_org_id;comment:组织id"`
	Name            string `gorm:"column:name;index:idx_custom_skill_name;comment:skill名称"`
	Desc            string `gorm:"column:desc;comment:skill描述"`
	Avatar          string `gorm:"column:avatar;comment:skill头像"`
	Author          string `gorm:"column:author;comment:作者"`
	ObjectPath      string `gorm:"column:object_path;comment:skill数据minio对象路径(zip压缩包)"`
	Markdown        string `gorm:"column:markdown;type:text;comment:skill markdown内容"`
	WgaThreadID     string `gorm:"column:wga_thread_id;index:idx_custom_skill_wga_thread_id;comment:WGA线程id"`
	PreviewThreadID string `gorm:"column:preview_thread_id;index:idx_custom_skill_preview_thread_id;comment:预览线程id"`
	DownloadCount   int32  `gorm:"column:download_count;default:0;not null;comment:下载次数"`
	AcquiredCount   int32  `gorm:"column:acquired_count;default:0;not null;comment:添加次数"`
	CreatedAt       int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt       int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
