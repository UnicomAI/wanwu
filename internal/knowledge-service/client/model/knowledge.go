package model

type ReportStatus int

const (
	ReportInit          ReportStatus = 0   //社区报告未处理 [EN] Community report not processed
	ReportSuccess       ReportStatus = 120 //社区报告生成成功 [EN] Community report generated successfully
	ReportLoadFail      ReportStatus = 121 //社区报告加载失败 [EN] Community report failed to load
	ReportExtractFail   ReportStatus = 122 //社区报告生成失败 [EN] Community report generation failed
	ReportStoreFail     ReportStatus = 123 //社区报告持久化存储失败 [EN] Community reports persistent storage failure
	ReportProcessing    ReportStatus = 130 //社区报告生成中 [EN] Community report is being generated
	ReportInterruptFail ReportStatus = 139 //社区报告处理中断 [EN] Community report processing outage
)

type KnowledgeBase struct {
	Id                   uint32       `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`       // Primary Key
	KnowledgeId          string       `gorm:"uniqueIndex:idx_unique_knowledge_id;column:knowledge_id;type:varchar(64)" json:"knowledgeId"` // Business Primary Key
	Name                 string       `gorm:"column:name;index:idx_user_id_name,priority:2;type:varchar(256);not null;default:''" json:"name"`
	RagName              string       `gorm:"column:rag_name;type:varchar(256);not null;default:''" json:"ragName"`
	Description          string       `gorm:"column:description;type:text;comment:'知识库描述';" json:"description"`
	DocCount             int          `gorm:"column:doc_count;type:int(11);not null;default:0;comment:'文档数量';" json:"docCount"`
	ShareCount           int          `gorm:"column:share_count;type:int(11);not null;default:0;comment:'文档共享数量';" json:"shareCount"`
	DocSize              int64        `gorm:"column:doc_size;type:bigint(20);not null;default:0;comment:'文档大小单位：字节';" json:"docSize"`
	EmbeddingModel       string       `gorm:"column:embedding_model;type:longtext;not null;comment:'embedding模型信息';" json:"embeddingModel"`
	KnowledgeGraphSwitch int          `gorm:"column:knowledge_graph_switch;type:tinyint(1);not null;default:0;comment:'知识图谱开关，方便查询过滤，0：关闭，1：开启';" json:"knowledgeGraphSwitch"`
	KnowledgeGraph       string       `gorm:"column:knowledge_graph;type:longtext;not null;comment:'知识图谱配置';" json:"knowledgeGraph"`
	ReportCreateCount    int          `gorm:"column:report_create_count;type:int(11);not null;default:0;comment:'社区报告生成数量'" json:"reportCreateCount"`
	ReportStatus         ReportStatus `gorm:"column:report_status;type:int(11);not null;comment:'0-待处理， 120- 生成成功， 130-生成中，121-社区报告加载图谱失败，122-生成社区报告失败，123-社区报告持久化存储失败，预留120~140';" json:"reportStatus"`
	CreatedAt            int64        `gorm:"column:create_at;type:bigint(20);not null;" json:"createAt"` // Create Time
	UpdatedAt            int64        `gorm:"column:update_at;type:bigint(20);not null;" json:"updateAt"` // Update Time
	UserId               string       `gorm:"column:user_id;index:idx_user_id_name,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId                string       `gorm:"column:org_id;type:varchar(64);not null;default:'';" json:"orgId"`
	Deleted              int          `gorm:"column:deleted;type:tinyint(1);not null;default:0;comment:'是否逻辑删除';" json:"deleted"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_base"
}

func ErrorReportStatus(status ReportStatus) bool {
	return status != ReportSuccess && status != ReportInit && status != ReportProcessing
}

func InReportStatus(status int) bool {
	reportStatus := ReportStatus(status)
	return reportStatus >= ReportSuccess && reportStatus <= ReportInterruptFail
}
