package model

type ReportStatus int

const (
	ReportInit          ReportStatus = 0   //Community report not processed
	ReportSuccess       ReportStatus = 120 //Community report generated successfully
	ReportLoadFail      ReportStatus = 121 //Community report failed to load
	ReportExtractFail   ReportStatus = 122 //Community report generation failed
	ReportStoreFail     ReportStatus = 123 //Community reports persistent storage failure
	ReportProcessing    ReportStatus = 130 //Community report is being generated
	ReportInterruptFail ReportStatus = 139 //Community report processing outage
)

type KnowledgeBase struct {
	Id                   uint32       `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`       // Primary Key
	KnowledgeId          string       `gorm:"uniqueIndex:idx_unique_knowledge_id;column:knowledge_id;type:varchar(64)" json:"knowledgeId"` // Business Primary Key
	Name                 string       `gorm:"column:name;index:idx_user_id_name,priority:2;type:varchar(256);not null;default:''" json:"name"`
	RagName              string       `gorm:"column:rag_name;type:varchar(256);not null;default:''" json:"ragName"`
	Description          string       `gorm:"column:description;type:text;comment:'Knowledge base description';" json:"description"`
	DocCount             int          `gorm:"column:doc_count;type:int(11);not null;default:0;comment:'Document count';" json:"docCount"`
	ShareCount           int          `gorm:"column:share_count;type:int(11);not null;default:0;comment:'Shared document count';" json:"shareCount"`
	DocSize              int64        `gorm:"column:doc_size;type:bigint(20);not null;default:0;comment:'Document size in bytes';" json:"docSize"`
	EmbeddingModel       string       `gorm:"column:embedding_model;type:longtext;not null;comment:'Embedding model information';" json:"embeddingModel"`
	KnowledgeGraphSwitch int          `gorm:"column:knowledge_graph_switch;type:tinyint(1);not null;default:0;comment:'Knowledge graph switch, 0: off, 1: on';" json:"knowledgeGraphSwitch"`
	KnowledgeGraph       string       `gorm:"column:knowledge_graph;type:longtext;not null;comment:'Knowledge graph configuration';" json:"knowledgeGraph"`
	ReportCreateCount    int          `gorm:"column:report_create_count;type:int(11);not null;default:0;comment:'Community report generation count'" json:"reportCreateCount"`
	ReportStatus         ReportStatus `gorm:"column:report_status;type:int(11);not null;comment:'0-pending, 120-success, 130-generating, 121-load failed, 122-generation failed, 123-storage failed, reserved 120~140';" json:"reportStatus"`
	CreatedAt            int64        `gorm:"column:create_at;type:bigint(20);not null;" json:"createAt"` // Create Time
	UpdatedAt            int64        `gorm:"column:update_at;type:bigint(20);not null;" json:"updateAt"` // Update Time
	UserId               string       `gorm:"column:user_id;index:idx_user_id_name,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId                string       `gorm:"column:org_id;type:varchar(64);not null;default:'';" json:"orgId"`
	Deleted              int          `gorm:"column:deleted;type:tinyint(1);not null;default:0;comment:'Is logically deleted';" json:"deleted"`
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
