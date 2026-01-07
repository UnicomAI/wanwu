package model

// KnowledgeKeywords  知识库关键词映射表
type KnowledgeKeywords struct {
	Id               uint32 `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"` // Primary Key
	Name             string `json:"name" gorm:"column:name;type:varchar(255);comment:专名词"`
	Alias            string `json:"alias" gorm:"column:alias;type:varchar(255);comment:别名"`
	KnowledgeBaseIds string `json:"knowledgeBaseIds" gorm:"column:knowledge_base_ids;type:text;comment:关联的知识库id;内容格式为:[\"2\",\"3\"]"`
	UserId           string `json:"userId" gorm:"column:user_id;type:varchar(64);not null;index:idx_user_id;comment:用户id;default:''"`
	OrgId            string `json:"orgId" gorm:"column:org_id;type:varchar(64);not null;default:''"`
	CreatedAt        int64  `gorm:"autoCreateTime:milli;index:created_at;column:created_at;type:bigint" json:"createdAt"`
	UpdatedAt        int64  `gorm:"autoUpdateTime:milli;column:updated_at;type:bigint" json:"updatedAt"`
}

func (KnowledgeKeywords) TableName() string {
	return "knowledge_keywords"
}
