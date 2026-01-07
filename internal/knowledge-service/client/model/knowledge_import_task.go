package model

import "github.com/UnicomAI/wanwu/pkg/db"

const (
	KnowledgeImportAnalyze     = 1   //知识库任务解析中
	KnowledgeImportSubmit      = 2   //知识库任务已提交
	KnowledgeImportFinish      = 3   //知识库任务导入完成
	KnowledgeImportError       = 4   //知识库任务导入失败
	FileImportType             = 0   //文件上传
	UrlImportType              = 1   //url上传
	UrlFileImportType          = 2   //2.批量url上传
	ParentSegmentMethod        = "1" //父子分段
	CommonSegmentMethod        = "0" //通用分段
	ImportTaskTypeCreate       = 0
	ImportTaskTypeUpdateConfig = 1 //更新配置
)

type SegmentConfig struct {
	SegmentMethod  string   `json:"segmentMethod"`                   ////分段方法 0：通用分段；1：父子分段,字符串为空则认为是通用分段
	SegmentType    string   `json:"segmentType" validate:"required"` //分段方式 0：自定分段；1：自定义分段
	Splitter       []string `json:"splitter"`                        // 分隔符（只有自定义分段必填）
	MaxSplitter    int      `json:"maxSplitter"`                     // 可分隔最大值（只有自定义分段必填）
	Overlap        float32  `json:"overlap"`                         // 可重叠值（只有自定义分段必填）
	SubSplitter    []string `json:"subSplitter"`                     // 分隔符（只有父子分段必填）
	SubMaxSplitter int      `json:"subMaxSplitter"`                  // 可分隔最大值（只有父子分段必填）
}

type DocAnalyzer struct {
	AnalyzerList []string `json:"analyzerList"` //文档解析方式，ocr等
}

type DocPreProcess struct {
	PreProcessList []string `json:"preProcessList"` //文档预处理方式: replace_symbols, delete_links
}

type DocImportInfo struct {
	DocInfoList []*DocInfo `json:"docInfoList"`
}

type DocInfo struct {
	DocId       string `json:"docId"`       //文档id
	DocName     string `json:"docName"`     //文档名称
	DocUrl      string `json:"docUrl"`      //文档url
	DocType     string `json:"docType"`     // 文档类型
	DocSize     int64  `json:"docSie"`      // 文档大小
	DirFilePath string `json:"dirFilePath"` //所在文件夹中的路径
	FilePathMd5 string `json:"filePathMd5"` //文件路径md5
}

type DocImportMetaData struct {
	DocMetaDataList []*KnowledgeDocMeta `json:"docMetaDataList"`
}

type DocMetaData struct {
	MetaId    string      `json:"metaId"`    // 元数据id
	Key       string      `json:"key"`       // key
	Value     interface{} `json:"value"`     // 常量
	ValueType string      `json:"valueType"` // 常量类型
	Rule      string      `json:"rule"`      // 正则表达式
}

type KnowledgeImportTask struct {
	Id            uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id';" json:"id"`
	ImportId      string      `gorm:"uniqueIndex:idx_unique_import_id;column:import_id;type:varchar(64)" json:"importId"` // Business Primary Key
	KnowledgeId   string      `gorm:"column:knowledge_id;type:varchar(64);not null;index:idx_knowledge_id" json:"knowledgeId"`
	ImportType    int         `gorm:"column:import_type;not null;" json:"importType"`
	TaskType      int         `gorm:"column:task_type;type:tinyint(1);not null;default:0;comment:'0:创建导入，1：配置更新'" json:"taskType"`
	Status        int         `gorm:"column:status;not null;comment:'0-任务待处理；1-任务解析中 ；2-任务提交算法完成；3-任务完成；4-任务失败" json:"status"`
	ErrorMsg      db.LongText `gorm:"column:error_msg;not null;comment:'解析的错误信息'" json:"errorMsg"`
	DocInfo       db.LongText `gorm:"column:doc_info;not null;comment:'文件信息'" json:"docInfo"`
	SegmentConfig string      `gorm:"column:segment_config;type:text;not null;comment:'分段配置信息'" json:"segmentConfig"`
	DocAnalyzer   string      `gorm:"column:doc_analyzer;type:text;not null;comment:'文档解析配置'" json:"docAnalyzer"`
	OcrModelId    string      `gorm:"column:ocr_model_id;type:varchar(64);not null;default:'';comment:'ocr模型id'" json:"ocrModelId"`
	DocPreProcess string      `gorm:"column:doc_pre_process;type:text;not null;comment:'文档预处理规则: replace_symbols,delete_links'" json:"docPreProcess"`
	MetaData      string      `gorm:"column:meta_data;type:text;not null;comment:'元数据列表'" json:"metaData"`
	CreatedAt     int64       `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt     int64       `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId        string      `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId         string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (KnowledgeImportTask) TableName() string {
	return "knowledge_import_task"
}
