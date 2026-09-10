package model

import (
	"encoding/json"
	"strings"

	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
)

// KnowledgeParseTemplate 知识库文档解析模板，按文档类型保存可复用的解析配置
type KnowledgeParseTemplate struct {
	Id            uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`
	TemplateId    string `gorm:"uniqueIndex:idx_unique_template_id;column:template_id;type:varchar(64)" json:"templateId"` // Business Primary Key
	Name          string `gorm:"column:name;type:varchar(256);not null;default:''" json:"name"`
	DocType       string `gorm:"column:doc_type;index:idx_user_id_doc_type,priority:2;type:varchar(20);not null;default:'';comment:'文档类型'" json:"docType"`
	SegmentConfig string `gorm:"column:segment_config;type:text;not null;comment:'分段配置信息'" json:"segmentConfig"`
	DocAnalyzer   string `gorm:"column:doc_analyzer;type:text;not null;comment:'文档解析配置'" json:"docAnalyzer"`
	OcrModelId    string `gorm:"column:ocr_model_id;type:varchar(64);not null;default:'';comment:'ocr模型id'" json:"ocrModelId"`
	DocPreProcess string `gorm:"column:doc_pre_process;type:text;not null;comment:'文档预处理规则: replace_symbols,delete_links'" json:"docPreProcess"`
	BuiltIn       int    `gorm:"column:built_in;type:tinyint(1);not null;default:0;comment:'是否内置（默认）模板，每种文档类型至多一个'" json:"builtIn"`
	CreatedAt     int64  `gorm:"column:create_at;type:bigint(20);autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt     int64  `gorm:"column:update_at;type:bigint(20);autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId        string `gorm:"column:user_id;index:idx_user_id_doc_type,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId         string `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (KnowledgeParseTemplate) TableName() string {
	return "knowledge_parse_template"
}

// KnowledgeParseTemplateBind 知识库上选定的解析模板
// 模板不共享，绑定也按人存：同一个知识库，每个用户看到和使用的都是自己的那份
type KnowledgeParseTemplateBind struct {
	Id          uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;" json:"id"`
	KnowledgeId string `gorm:"column:knowledge_id;uniqueIndex:uk_knowledge_doc_user,priority:1;type:varchar(64);not null;default:''" json:"knowledgeId"`
	DocType     string `gorm:"column:doc_type;uniqueIndex:uk_knowledge_doc_user,priority:2;type:varchar(20);not null;default:''" json:"docType"`
	UserId      string `gorm:"column:user_id;uniqueIndex:uk_knowledge_doc_user,priority:3;type:varchar(64);not null;default:''" json:"userId"`
	OrgId       string `gorm:"column:org_id;uniqueIndex:uk_knowledge_doc_user,priority:4;type:varchar(64);not null;default:''" json:"orgId"`
	TemplateId  string `gorm:"column:template_id;index:idx_template_id;type:varchar(64);not null;default:''" json:"templateId"`
	CreatedAt   int64  `gorm:"column:create_at;type:bigint(20);autoCreateTime:milli;not null;" json:"createAt"`
	UpdatedAt   int64  `gorm:"column:update_at;type:bigint(20);autoUpdateTime:milli;not null;" json:"updateAt"`
}

func (KnowledgeParseTemplateBind) TableName() string {
	return "knowledge_parse_template_bind"
}

// BuiltInTemplateName 内置（默认）模板名称，由后端落库时写入
const BuiltInTemplateName = "内置（默认）"

// ParseTemplateDocTypes 支持解析模板的文档类型
var ParseTemplateDocTypes = []string{
	"pdf", "word", "ppt", "excel", "csv", "txt", "html", "markdown", "wps", "ofd", "video", "audio", "image",
}

// docTypeExtMap 文件后缀到解析模板文档类型的映射，未命中的按默认参数处理
var docTypeExtMap = map[string]string{
	"pdf":  "pdf",
	"docx": "word",
	"doc":  "word",
	"pptx": "ppt",
	"ppt":  "ppt",
	"xlsx": "excel",
	"xls":  "excel",
	"csv":  "csv",
	"txt":  "txt",
	"html": "html",
	"htm":  "html",
	"md":   "markdown",
	"wps":  "wps",
	"ofd":  "ofd",
	"avi":  "video",
	"mp4":  "video",
	"mov":  "video",
	"wmv":  "video",
	"mp3":  "audio",
	"wav":  "audio",
	"aac":  "audio",
	"m4a":  "audio",
	"png":  "image",
	"jpg":  "image",
	"jpeg": "image",
}

// DocTypeByExt 按文件后缀取解析模板的文档类型，归不了类返回空串
func DocTypeByExt(ext string) string {
	return docTypeExtMap[strings.ToLower(strings.TrimPrefix(ext, "."))]
}

// mediaDocTypes 视频/音频/图片，解析参数必须由用户显式配模型，没有内置（默认）模板兜底
var mediaDocTypes = map[string]struct{}{"video": {}, "audio": {}, "image": {}}

// IsMediaDocType 是否视频/音频/图片
func IsMediaDocType(docType string) bool {
	_, ok := mediaDocTypes[docType]
	return ok
}

// BuiltInDocTypes 需要内置（默认）模板的文档类型，媒体类型不在其中
var BuiltInDocTypes = func() []string {
	result := make([]string, 0, len(ParseTemplateDocTypes))
	for _, docType := range ParseTemplateDocTypes {
		if !IsMediaDocType(docType) {
			result = append(result, docType)
		}
	}
	return result
}()

// 内置（默认）模板的解析参数，全局唯一一份，前端不再另存
const (
	defaultSegmentType    = "0" //自动分段
	defaultSplitter       = "\n\n"
	defaultSubSplitter    = "\n"
	defaultMaxSplitter    = 1024
	defaultOverlap        = 0.2
	defaultSubMaxSplitter = 200
	defaultAnalyzer       = "text"
	defaultPreProcess     = "replaceSymbols"
)

func DefaultSegmentConfig() *SegmentConfig {
	return &SegmentConfig{
		SegmentMethod:  CommonSegmentMethod,
		SegmentType:    defaultSegmentType,
		Splitter:       []string{defaultSplitter},
		MaxSplitter:    defaultMaxSplitter,
		Overlap:        defaultOverlap,
		SubSplitter:    []string{defaultSubSplitter},
		SubMaxSplitter: defaultSubMaxSplitter,
	}
}

func DefaultDocAnalyzer() *DocAnalyzer {
	return &DocAnalyzer{AnalyzerList: []string{defaultAnalyzer}}
}

func DefaultDocPreProcess() *DocPreProcess {
	return &DocPreProcess{PreProcessList: []string{defaultPreProcess}}
}

// BuiltInTemplateId 内置模板用确定性 id，并发重复插入时撞 template_id 唯一索引而不是插出两条
func BuiltInTemplateId(userId, orgId, docType string) string {
	return "builtin-" + wanwu_util.MD5([]byte(userId+"|"+orgId+"|"+docType))
}

// NewBuiltInParseTemplate 构造某文档类型的内置（默认）模板
func NewBuiltInParseTemplate(userId, orgId, docType string) (*KnowledgeParseTemplate, error) {
	segmentConfig, err := json.Marshal(DefaultSegmentConfig())
	if err != nil {
		return nil, err
	}
	docAnalyzer, err := json.Marshal(DefaultDocAnalyzer())
	if err != nil {
		return nil, err
	}
	docPreProcess, err := json.Marshal(DefaultDocPreProcess())
	if err != nil {
		return nil, err
	}
	return &KnowledgeParseTemplate{
		TemplateId:    BuiltInTemplateId(userId, orgId, docType),
		Name:          BuiltInTemplateName,
		DocType:       docType,
		SegmentConfig: string(segmentConfig),
		DocAnalyzer:   string(docAnalyzer),
		DocPreProcess: string(docPreProcess),
		BuiltIn:       1,
		UserId:        userId,
		OrgId:         orgId,
	}, nil
}
