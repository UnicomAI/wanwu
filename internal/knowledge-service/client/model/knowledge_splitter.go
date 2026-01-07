package model

type KnowledgeSplitter struct {
	Id         uint32 `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id';" json:"id"`        // Primary Key
	SplitterId string `gorm:"uniqueIndex:idx_unique_splitter_id;column:splitter_id;type:varchar(64)" json:"splitterId"` // Business Primary Key
	Name       string `gorm:"column:name;index:idx_user_id_name,priority:2;type:varchar(64);not null;default:''" json:"name"`
	Value      string `gorm:"column:value;type:varchar(64);not null;default:''" json:"value"`
	CreatedAt  int64  `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt  int64  `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId     string `gorm:"column:user_id;index:idx_user_id_name,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId      string `gorm:"column:org_id;type:varchar(64);not null;default:'';" json:"orgId"`
}

func (KnowledgeSplitter) TableName() string {
	return "knowledge_splitter"
}
