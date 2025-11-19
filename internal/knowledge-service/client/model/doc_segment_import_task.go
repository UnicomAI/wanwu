package model

const (
	DocSegmentImportInit      = 0 //任务待处理 [EN] Task pending
	DocSegmentImportImporting = 1 //文档分段导入中 [EN] Documents are being imported in sections
	DocSegmentImportSuccess   = 2 //文档分段导入成功 [EN] Document segments imported successfully
	DocSegmentImportFail      = 3 //文档分段导入失败 [EN] Document segment import failed
)

type DocSegmentImportParams struct {
	KnowledgeName      string   `json:"knowledgeName"`      // 知识库的名称 [EN] The name of the knowledge base
	KnowledgeRagName   string   `json:"knowledgeRagName"`   // 知识库的rag名称 [EN] The rag name of the knowledge base
	KnowledgeId        string   `json:"knowledgeId"`        // 知识库的唯一ID [EN] The unique ID of the knowledge base
	KnowledgeCreatorId string   `json:"knowledgeCreatorId"` // 知识库的创建者ID [EN] Creator ID of the knowledge base
	FileName           string   `json:"fileName"`           // 与chunk关联的文件名 [EN] The file name associated with the chunk
	MaxSentenceSize    int      `json:"maxSentenceSize"`    // 最大分段长度限制 [EN] Maximum segment length limit
	FileUrl            string   `json:"fileUrl"`            //文件url [EN] file url
	SegmentMethod      string   `json:"segmentMethod"`      ////分段方法 0：通用分段；1：父子分段,字符串为空则认为是通用分段 [EN] //Segmentation method 0: universal segmentation; 1: parent-child segmentation, if the string is empty, it is considered a universal segmentation
	SubSplitter        []string `json:"subSplitter"`        // 分隔符（只有父子分段必填） [EN] Delimiter (required only for parent-child segments)
	SubMaxSplitter     int      `json:"subMaxSplitter"`     // 可分隔最大值（只有父子分段必填） [EN] Maximum separable value (required only for parent-child segments)
}

type ChildChunkConfig struct {
	Separators []string `json:"separators"` // 分隔符 [EN] delimiter
	ChunkSize  int32    `json:"chunk_size"` // 子分段大小 [EN] subsegment size
}

type DocSegmentImportTask struct {
	Id           uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:'id';" json:"id"`
	ImportId     string `gorm:"uniqueIndex:idx_unique_import_id;column:import_id;type:varchar(64)" json:"importId"` // Business Primary Key
	DocId        string `gorm:"column:doc_id;type:varchar(64);not null;index:idx_doc_id" json:"docId"`
	Status       int    `gorm:"column:status;type:tinyint(1);not null;comment:'0-任务待处理；1-任务导入中 ；2-任务完成；3-任务失败'" json:"status"`
	SuccessCount int    `gorm:"column:success_count;type:bigint(20);default:0;comment:'成功数量'" json:"successCount"`
	TotalCount   int    `gorm:"column:total_count;type:bigint(20);default:0;comment:'导入数量，当在导入过程中出现重启，则total为0'" json:"totalCount"`
	ErrorMsg     string `gorm:"column:error_msg;type:longtext;not null;comment:'解析的错误信息'" json:"errorMsg"`
	ImportParams string `gorm:"column:import_params;type:text;not null;comment:'导入信息'" json:"importParams"`
	CreatedAt    int64  `gorm:"column:create_at;type:bigint(20);not null;autoCreateTime:milli" json:"createAt"` // Create Time
	UpdatedAt    int64  `gorm:"column:update_at;type:bigint(20);not null;autoUpdateTime:milli" json:"updateAt"` // Update Time
	UserId       string `gorm:"column:user_id;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
}

func (DocSegmentImportTask) TableName() string {
	return "knowledge_doc_segment_import_task"
}
