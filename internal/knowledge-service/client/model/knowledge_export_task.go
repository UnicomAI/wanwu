package model

import "github.com/UnicomAI/wanwu/pkg/db"

const (
	KnowledgeExportInit      = 0 //任务待处理
	KnowledgeExportExporting = 1 //导出中
	KnowledgeExportSuccess   = 2 //导出成功
	KnowledgeExportFail      = 3 //导出失败
)

type KnowledgeExportTaskParams struct {
	KnowledgeId string   `json:"knowledgeId"` //知识库id
	DocIdList   []string `json:"docIdList"`   //文档id列表
}

type KnowledgeExportTask struct {
	Id             uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id';" json:"id"`
	ExportId       string      `gorm:"uniqueIndex:idx_unique_export_id;column:export_id;type:varchar(64)" json:"exportId"` // Business Primary Key
	KnowledgeId    string      `gorm:"column:knowledge_id;type:varchar(64);not null;index:idx_knowledge_id" json:"knowledgeId"`
	ExportFilePath string      `gorm:"column:export_file_path;type:text;not null;comment:'导出文件地址'" json:"exportFilePath"`
	ExportFileSize int64       `gorm:"column:export_file_size;type:bigint;not null;comment:'导出文件大小'" json:"exportFileSize"`
	Status         int         `gorm:"column:status;not null;comment:'0-任务待处理；1-任务导出中 ；2-任务完成；3-任务失败'" json:"status"`
	SuccessCount   int         `gorm:"column:success_count;type:bigint;default:0;comment:'成功数量'" json:"successCount"`
	TotalCount     int         `gorm:"column:total_count;type:bigint;default:0;comment:'导出数量，当在导出过程中出现重启，则total为0'" json:"totalCount"`
	ErrorMsg       db.LongText `gorm:"column:error_msg;not null;comment:'导出的错误信息'" json:"errorMsg"`
	ExportParams   string      `gorm:"column:export_params;type:text;not null;comment:'导出信息'" json:"exportParams"`
	CreatedAt      int64       `gorm:"column:create_at;type:bigint;not null;autoCreateTime:milli" json:"createAt"` // Create Time
	UpdatedAt      int64       `gorm:"column:update_at;type:bigint;not null;autoUpdateTime:milli" json:"updateAt"` // Update Time
	UserId         string      `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId          string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (KnowledgeExportTask) TableName() string {
	return "knowledge_export_task"
}
