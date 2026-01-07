package model

type KnowledgeTagRelation struct {
	Id          uint32 `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id';" json:"id"`                   // Primary Key
	TagId       string `gorm:"column:tag_id;index:idx_tag_id;type:varchar(64);not null;default:'';" json:"tagId"`                   // tagId
	KnowledgeId string `gorm:"column:knowledge_id;index:idx_knowledge_id;type:varchar(64);not null;default:'';" json:"knowledgeId"` // knowledgeId
	CreatedAt   int64  `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"`                         // Create Time
	UpdatedAt   int64  `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"`                         // Update Time
	UserId      string `gorm:"column:user_id;index:idx_user_id_tag_name,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId       string `gorm:"column:org_id;type:varchar(64);not null;default:'';" json:"orgId"`
}

func (KnowledgeTagRelation) TableName() string {
	return "knowledge_tag_relation"
}
