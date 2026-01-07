package model

import "github.com/UnicomAI/wanwu/pkg/db"

const (
	DocSegmentImportInit      = 0 //任务待处理
	DocSegmentImportImporting = 1 //文档分段导入中
	DocSegmentImportSuccess   = 2 //文档分段导入成功
	DocSegmentImportFail      = 3 //文档分段导入失败
)

type DocSegmentImportParams struct {
	KnowledgeName      string   `json:"knowledgeName"`      // 知识库的名称
	KnowledgeRagName   string   `json:"knowledgeRagName"`   // 知识库的rag名称
	KnowledgeId        string   `json:"knowledgeId"`        // 知识库的唯一ID
	KnowledgeCreatorId string   `json:"knowledgeCreatorId"` // 知识库的创建者ID
	FileName           string   `json:"fileName"`           // 与chunk关联的文件名
	MaxSentenceSize    int      `json:"maxSentenceSize"`    // 最大分段长度限制
	FileUrl            string   `json:"fileUrl"`            //文件url
	SegmentMethod      string   `json:"segmentMethod"`      ////分段方法 0：通用分段；1：父子分段,字符串为空则认为是通用分段
	SubSplitter        []string `json:"subSplitter"`        // 分隔符（只有父子分段必填）
	SubMaxSplitter     int      `json:"subMaxSplitter"`     // 可分隔最大值（只有父子分段必填）
}

type ChildChunkConfig struct {
	Separators []string `json:"separators"` // 分隔符
	ChunkSize  int32    `json:"chunk_size"` // 子分段大小
}

type DocSegmentImportTask struct {
	Id           uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id';" json:"id"`
	ImportId     string      `gorm:"uniqueIndex:idx_unique_import_id;column:import_id;type:varchar(64)" json:"importId"` // Business Primary Key
	DocId        string      `gorm:"column:doc_id;type:varchar(64);not null;index:idx_doc_id" json:"docId"`
	Status       int         `gorm:"column:status;not null;comment:'0-任务待处理；1-任务导入中 ；2-任务完成；3-任务失败'" json:"status"`
	SuccessCount int         `gorm:"column:success_count;type:bigint;default:0;comment:'成功数量'" json:"successCount"`
	TotalCount   int         `gorm:"column:total_count;type:bigint;default:0;comment:'导入数量，当在导入过程中出现重启，则total为0'" json:"totalCount"`
	ErrorMsg     db.LongText `gorm:"column:error_msg;not null;comment:'解析的错误信息'" json:"errorMsg"`
	ImportParams string      `gorm:"column:import_params;type:text;not null;comment:'导入信息'" json:"importParams"`
	CreatedAt    int64       `gorm:"column:create_at;type:bigint;not null;autoCreateTime:milli" json:"createAt"` // Create Time
	UpdatedAt    int64       `gorm:"column:update_at;type:bigint;not null;autoUpdateTime:milli" json:"updateAt"` // Update Time
	UserId       string      `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (DocSegmentImportTask) TableName() string {
	return "knowledge_doc_segment_import_task"
}
