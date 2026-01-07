package model

import "github.com/UnicomAI/wanwu/pkg/db"

const (
	QAPairSuccess = 2 // 问答对文件导入成功
)

type KnowledgeQAPair struct {
	Id           uint32      `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"`              // Primary Key
	QAPairId     string      `gorm:"uniqueIndex:idx_unique_qa_pair_id;column:qa_pair_id;type:varchar(64)" json:"qaPairId"` // Business Primary Key
	ImportTaskId string      `gorm:"column:import_id;type:varchar(64);not null;default:'';comment:'导入的任务id'" json:"importId"`
	KnowledgeId  string      `gorm:"column:knowledge_id;uniqueIndex:idx_knowledge_id_md5,priority:1;;type:varchar(64);not null;default:''" json:"knowledgeId"`
	Question     db.LongText `gorm:"column:question;not null;comment:'问题'" json:"question"`
	Answer       db.LongText `gorm:"column:answer;not null;comment:'答案'" json:"answer"`
	Switch       bool        `gorm:"column:switch;not null;default:0;comment:'开关'" json:"switch"`
	Status       int         `gorm:"column:status;not null;comment:'0-待处理， 1-导入中，2-导入成功，3-导入失败';" json:"status"`
	ErrorMsg     db.LongText `gorm:"column:error_msg;not null;comment:'导入状态错误信息'" json:"errorMsg"`
	QuestionMd5  string      `gorm:"column:question_md5;uniqueIndex:idx_knowledge_id_md5,priority:2;type:varchar(64);not null;default:'';comment:'问题的md5值'" json:"questionMd5"`
	CreatedAt    int64       `gorm:"column:create_at;type:bigint;not null;autoCreateTime:milli;" json:"createAt"` // Create Time
	UpdatedAt    int64       `gorm:"column:update_at;type:bigint;not null;autoCreateTime:milli;" json:"updateAt"` // Update Time
	UserId       string      `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
	Deleted      int         `gorm:"column:deleted;not null;default:0;comment:'是否逻辑删除';" json:"deleted"`
}

func (KnowledgeQAPair) TableName() string {
	return "knowledge_qa_pair"
}
