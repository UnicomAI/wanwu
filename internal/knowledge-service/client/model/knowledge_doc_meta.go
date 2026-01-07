package model

const (
	MetaTypeString = "string"
	MetaTypeNumber = "number"
	MetaTypeTime   = "time"
)

type KnowledgeDocMeta struct {
	Id          uint32 `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"` // Primary Key
	KnowledgeId string `gorm:"index:idx_knowledge_id;index:idx_knowledge_id_value_main,priority:1;column:knowledge_id;type:varchar(64);not null;default:''" json:"knowledgeId"`
	MetaId      string `gorm:"uniqueIndex:idx_unique_meta_id;column:meta_id;type:varchar(64)" json:"metaId"`
	DocId       string `gorm:"index:idx_doc_id;column:doc_id;type:varchar(64)" json:"docId"`
	Key         string `gorm:"index:idx_meta_key;column:key;type:varchar(64);not null;default:''" json:"key"`
	Value       string `gorm:"column:value;type:text;not null" json:"value"`
	ValueMain   string `gorm:"index:idx_knowledge_id_value_main,priority:2;column:value_main;type:varchar(128);not null;default:'';comment:'替代原value字段'" json:"valueMain"`
	ValueType   string `gorm:"column:value_type;type:varchar(64);not null;default:'string';comment:'string,number,time'" json:"valueType"`
	Rule        string `gorm:"column:rule;type:text;not null" json:"rule"`
	CreatedAt   int64  `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt   int64  `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId      string `gorm:"column:user_id;index:idx_user_id_knowledge_id_name,priority:1;index:idx_user_id_knowledge_id_tag,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId       string `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (KnowledgeDocMeta) TableName() string {
	return "knowledge_doc_meta"
}

type UpdateKeys struct {
	OldKey string `json:"oldKey"`
	NewKey string `json:"newKey"`
}
