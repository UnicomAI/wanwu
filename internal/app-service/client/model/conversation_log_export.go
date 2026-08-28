package model

const (
	ConversationLogExportInit      = 0 //待处理
	ConversationLogExportExporting = 1 //导出中
	ConversationLogExportSuccess   = 2 //成功
	ConversationLogExportFail      = 3 //失败
)

// ConversationLogExportTaskParams 导出任务参数（落库到 export_params）
type ConversationLogExportTaskParams struct {
	AppId   string   `json:"appId"`
	AppType string   `json:"appType"`
	LogIds  []string `json:"logIds"`
	OrgIds  []string `json:"orgIds"`
	UserIds []string `json:"userIds"`
}

// ConversationLogExportTask 对话日志导出任务，字段语义与知识库 KnowledgeExportTask 对齐，仅把 KnowledgeId 换成 BusinessId。
type ConversationLogExportTask struct {
	Id             uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`
	ExportId       string `gorm:"uniqueIndex:idx_unique_conv_log_export_id;column:export_id;type:varchar(64)" json:"exportId"`
	AppId          string `gorm:"column:app_id;type:varchar(64);not null;comment:应用id:智能体、知识问答应用id"`
	AppType        string `gorm:"column:app_type;type:varchar(64);not null;comment:应用类型:智能体、知识问答"`
	Title          string `gorm:"column:title;type:varchar(64);not null;comment:导出任务标题(智能体名称)"`
	ExportFilePath string `gorm:"column:export_file_path;type:text;not null;comment:'导出文件地址'" json:"exportFilePath"`
	ExportFileSize int64  `gorm:"column:export_file_size;type:bigint(20);not null;comment:'导出文件大小'" json:"exportFileSize"`
	Status         int    `gorm:"column:status;type:tinyint(1);not null;comment:'0-任务待处理；1-任务导出中 ；2-任务完成；3-任务失败'" json:"status"`
	SuccessCount   int    `gorm:"column:success_count;type:bigint(20);default:0;comment:'成功数量'" json:"successCount"`
	TotalCount     int    `gorm:"column:total_count;type:bigint(20);default:0;comment:'导出数量，当在导出过程中出现重启，则total为0'" json:"totalCount"`
	ErrorMsg       string `gorm:"column:error_msg;type:longtext;not null;comment:'导出的错误信息'" json:"errorMsg"`
	ExportParams   string `gorm:"column:export_params;type:text;not null;comment:'导出信息(json，包含用户入参）'" json:"exportParams"`
	CreatedAt      int64  `gorm:"index:idx_conversation_log_export_task_create_at;autoCreateTime:milli;comment:创建时间"`
	UpdatedAt      int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
	UserId         string `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId          string `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (ConversationLogExportTask) TableName() string { return "conversation_log_export_task" }
