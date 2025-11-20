package model

type GraphStatus int

const (
	DocWaitingForUpload = -2 //Document to be uploaded
	DocInit             = 0  //Document pending
	DocSuccess          = 1  //Document processing completed
	DocProcessing       = 3  //Document processing
	DocFail             = 5  //Document pending

	GraphInit          GraphStatus = 0   //The spectrum is not processed
	GraphSuccess       GraphStatus = 100 //Map generated successfully
	GraphChunkFail     GraphStatus = 101 //Failed to generate chunk text from graph
	GraphExtractFail   GraphStatus = 102 //Failed to generate and extract the map
	GraphStoreFail     GraphStatus = 103 //Graph persistence storage failed
	GraphProcessing    GraphStatus = 110 //The graph begins to be parsed
	GraphInterruptFail GraphStatus = 119
)

type KnowledgeDoc struct {
	Id           uint32      `json:"id" gorm:"primary_key;type:bigint(20) auto_increment;not null;comment:'id';"` // Primary Key
	DocId        string      `gorm:"uniqueIndex:idx_unique_doc_id;column:doc_id;type:varchar(64)" json:"docId"`   // Business Primary Key
	ImportTaskId string      `gorm:"column:batch_id;type:varchar(64);not null;default:'';comment:'Import task id'" json:"importTaskId"`
	KnowledgeId  string      `gorm:"column:knowledge_id;index:idx_user_id_knowledge_id_name,priority:2;index:idx_user_id_knowledge_id_tag,priority:2;type:varchar(64);not null;default:''" json:"knowledgeId"`
	FilePathMd5  string      `gorm:"column:file_path_md5;type:varchar(64);not null;default:'';comment:'File MD5 value'" json:"filePathMd5"`
	FilePath     string      `gorm:"column:file_path;type:text;not null" json:"filePath"`
	Name         string      `gorm:"column:name;index:idx_user_id_knowledge_id_name,priority:3;type:varchar(256);not null;default:''" json:"name"`
	FileType     string      `gorm:"column:file_type;type:varchar(20);not null;default:''" json:"fileType"`
	FileSize     int64       `gorm:"column:file_size;type:bigint(20);COMMENT:'File size in bytes'" json:"fileSize"`
	Status       int         `gorm:"column:status;type:tinyint(1);not null;comment:'0-pending, 1-completed, 2-reviewing(not used), 3-processing, 4-review failed(not used), 5-failed';" json:"status"`
	GraphStatus  GraphStatus `gorm:"column:graph_status;type:int(11);not null;comment:'0-pending, 100-success, 101-chunk generation failed, 102-extraction failed, 103-storage failed, reserved 100~120';" json:"graphStatus"`
	ErrorMsg     string      `gorm:"column:error_msg;type:longtext;not null;comment:'Parsing error message'" json:"errorMsg"`
	CreatedAt    int64       `gorm:"column:create_at;type:bigint(20);not null;" json:"createAt"` // Create Time
	UpdatedAt    int64       `gorm:"column:update_at;type:bigint(20);not null;" json:"updateAt"` // Update Time
	UserId       string      `gorm:"column:user_id;index:idx_user_id_knowledge_id_name,priority:1;index:idx_user_id_knowledge_id_tag,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
	Deleted      int         `gorm:"column:deleted;type:tinyint(1);not null;default:0;comment:'Is logically deleted';" json:"deleted"`
}

func (KnowledgeDoc) TableName() string {
	return "knowledge_doc"
}

func SuccessGraphStatus(status int) bool {
	return GraphStatus(status) == GraphSuccess
}

// BuildGraphShowStatus report display status 0: pending, 1. parsing, 2. parsing successful, 3. parsing failed
func BuildGraphShowStatus(status GraphStatus) (int, string) {
	switch status {
	case GraphInit:
		return 0, ""
	case GraphProcessing:
		return 1, ""
	case GraphSuccess:
		return 2, ""
	}
	return 3, buildErrorMessage(status)
}

// todo multi-language is not processed
func buildErrorMessage(status GraphStatus) string {
	switch status {
	case GraphChunkFail:
		return "Graph generation chunk text failed"
	case GraphExtractFail:
		return "Graph generation extraction failed"
	case GraphStoreFail:
		return "Graph persistence storage failed"
	case GraphInterruptFail:
		return "Graph interruption failed"

	}
	return ""
}

func InGraphStatus(status int) bool {
	graphStatus := GraphStatus(status)
	return graphStatus >= GraphSuccess && graphStatus <= GraphInterruptFail
}
