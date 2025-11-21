package model

const (
	KnowledgeImportAnalyze = 1   //Knowledge base task analysis in progress
	KnowledgeImportSubmit  = 2   //Knowledge base task has been submitted
	KnowledgeImportFinish  = 3   //Knowledge base task import completed
	KnowledgeImportError   = 4   //Knowledge base task import failed
	FileImportType         = 0   //File upload
	UrlImportType          = 1   //url upload
	UrlFileImportType      = 2   //2. Batch URL upload
	ParentSegmentMethod    = "1" //Father-son segmentation
	CommonSegmentMethod    = "0" //common segmentation
)

type SegmentConfig struct {
	SegmentMethod  string   `json:"segmentMethod"`                   ////Segmentation method 0: universal segmentation; 1: parent-child segmentation, if the string is empty, it is considered a universal segmentation
	SegmentType    string   `json:"segmentType" validate:"required"` //Segmentation mode 0: Customized segmentation; 1: Customized segmentation
	Splitter       []string `json:"splitter"`                        // Delimiter (required only for custom segments)
	MaxSplitter    int      `json:"maxSplitter"`                     // Maximum separable value (required only for custom segments)
	Overlap        float32  `json:"overlap"`                         // Overlapping values ​​(required only for custom segments)
	SubSplitter    []string `json:"subSplitter"`                     // Delimiter (required only for parent-child segments)
	SubMaxSplitter int      `json:"subMaxSplitter"`                  // Maximum separable value (required only for parent-child segments)
}

type DocAnalyzer struct {
	AnalyzerList []string `json:"analyzerList"` //Document parsing method, OCR, etc.
}

type DocPreProcess struct {
	PreProcessList []string `json:"preProcessList"` //Document preprocessing method: replace_symbols, delete_links
}

type DocImportInfo struct {
	DocInfoList []*DocInfo `json:"docInfoList"`
}

type DocInfo struct {
	DocId   string `json:"docId"`   //document id
	DocName string `json:"docName"` //file name
	DocUrl  string `json:"docUrl"`  //Document url
	DocType string `json:"docType"` // Document type
	DocSize int64  `json:"docSie"`  // Document size
}

type DocImportMetaData struct {
	DocMetaDataList []*KnowledgeDocMeta `json:"docMetaDataList"`
}

type DocMetaData struct {
	MetaId    string      `json:"metaId"`    // metadata id
	Key       string      `json:"key"`       // key
	Value     interface{} `json:"value"`     // constant
	ValueType string      `json:"valueType"` // constant type
	Rule      string      `json:"rule"`      // regular expression
}

type KnowledgeImportTask struct {
	Id            uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`
	ImportId      string `gorm:"uniqueIndex:idx_unique_import_id;column:import_id;type:varchar(64)" json:"importId"` // Business Primary Key
	KnowledgeId   string `gorm:"column:knowledge_id;type:varchar(64);not null;index:idx_knowledge_id" json:"knowledgeId"`
	ImportType    int    `gorm:"column:import_type;type:tinyint(1);not null;" json:"importType"`
	Status        int    `gorm:"column:status;type:tinyint(1);not null;comment:'0-任务待处理；1-任务解析中 ；2-任务提交算法完成；3-任务完成；4-任务失败" json:"status"`
	ErrorMsg      string `gorm:"column:error_msg;type:longtext;not null;comment:'解析的错误信息'" json:"errorMsg"`
	DocInfo       string `gorm:"column:doc_info;type:longtext;not null;comment:'文件信息'" json:"docInfo"`
	SegmentConfig string `gorm:"column:segment_config;type:text;not null;comment:'分段配置信息'" json:"segmentConfig"`
	DocAnalyzer   string `gorm:"column:doc_analyzer;type:text;not null;comment:'文档解析配置'" json:"docAnalyzer"`
	OcrModelId    string `gorm:"column:ocr_model_id;type:varchar(64);not null;default:'';comment:'ocr模型id'" json:"ocrModelId"`
	DocPreProcess string `gorm:"column:doc_pre_process;type:text;not null;comment:'文档预处理规则: replace_symbols,delete_links'" json:"docPreProcess"`
	MetaData      string `gorm:"column:meta_data;type:text;not null;comment:'元数据列表'" json:"metaData"`
	CreatedAt     int64  `gorm:"column:create_at;type:bigint(20);not null;" json:"createAt"` // Create Time
	UpdatedAt     int64  `gorm:"column:update_at;type:bigint(20);not null;" json:"updateAt"` // Update Time
	UserId        string `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId         string `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (KnowledgeImportTask) TableName() string {
	return "knowledge_import_task"
}
