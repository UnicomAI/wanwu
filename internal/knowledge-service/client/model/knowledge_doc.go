package model

type GraphStatus int

const (
	DocWaitingForUpload = -2 //文档待上传
	DocInit             = 0  //文档待处理
	DocSuccess          = 1  //文档处理完成
	DocProcessing       = 3  //文档处理中
	DocFail             = 5  //文档解析失败
	DocSuccessNew       = 10 //文档处理完成

	GraphInit           GraphStatus = 0   //图谱未处理
	GraphSuccess        GraphStatus = 100 //graph 图谱提取成功（终态·成功）
	GraphChunkFail      GraphStatus = 101 //生成图谱获取 chunk 文本失败
	GraphExtractFail    GraphStatus = 102 //提取图谱失败
	GraphStoreFail      GraphStatus = 103 //图谱持久化存储失败
	GraphSchemaFail     GraphStatus = 104 //graph schema 解析失败
	GraphProcessing     GraphStatus = 110 //图谱开始解析
	GraphSchemaSuccess  GraphStatus = 111 //graph schema 解析成功
	GraphChunkSuccess   GraphStatus = 112 //生成图谱获取 chunk 文本成功
	GraphExtractSuccess GraphStatus = 113 //提取图谱成功
	GraphStoreSuccess   GraphStatus = 114 //图谱持久化存储成功
	GraphInterruptFail  GraphStatus = 119 //重启打断执行
)

type KnowledgeDoc struct {
	Id           uint32      `json:"id" gorm:"primary_key;type:bigint(20) auto_increment;not null;comment:'id';"` // Primary Key
	DocId        string      `gorm:"uniqueIndex:idx_unique_doc_id;column:doc_id;type:varchar(64)" json:"docId"`   // Business Primary Key
	ImportTaskId string      `gorm:"column:batch_id;type:varchar(64);not null;default:'';comment:'导入的任务id'" json:"importTaskId"`
	KnowledgeId  string      `gorm:"column:knowledge_id;index:idx_user_id_knowledge_id_name,priority:2;index:idx_user_id_knowledge_id_tag,priority:2;type:varchar(64);not null;default:''" json:"knowledgeId"`
	FilePathMd5  string      `gorm:"column:file_path_md5;type:varchar(64);not null;default:'';comment:'文件的md5值'" json:"filePathMd5"`
	FilePath     string      `gorm:"column:file_path;type:text;not null" json:"filePath"`
	DirFilePath  string      `gorm:"column:dir_file_path;type:text;not null;comment:'文件在文件夹中的相对目录'" json:"dirFilePath"`
	Name         string      `gorm:"column:name;index:idx_user_id_knowledge_id_name,priority:3;type:varchar(256);not null;default:''" json:"name"`
	FileType     string      `gorm:"column:file_type;type:varchar(20);not null;default:''" json:"fileType"`
	FileSize     int64       `gorm:"column:file_size;type:bigint(20);COMMENT:'文件大小，单位byte'" json:"fileSize"`
	Status       int         `gorm:"column:status;type:tinyint(1);not null;comment:'0-待处理， 1- 处理完成， 2-正在审核中(目前没有)，3-正在解析中，4-审核未通过（目前没有），5-解析失败';" json:"status"`
	GraphStatus  GraphStatus `gorm:"column:graph_status;type:int(11);not null;comment:'0-待处理， 100- 生成成功， 101-生成图谱获取chunk文本失败，102-提取图谱失败，103-图谱持久化存储失败，预留100~120';" json:"graphStatus"`
	ErrorMsg     string      `gorm:"column:error_msg;type:longtext;not null;comment:'解析的错误信息'" json:"errorMsg"`
	CreatedAt    int64       `gorm:"column:create_at;type:bigint(20);autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt    int64       `gorm:"column:update_at;type:bigint(20);autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId       string      `gorm:"column:user_id;index:idx_user_id_knowledge_id_name,priority:1;index:idx_user_id_knowledge_id_tag,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
	Deleted      int         `gorm:"column:deleted;type:tinyint(1);not null;default:0;comment:'是否逻辑删除';" json:"deleted"`
}

func (KnowledgeDoc) TableName() string {
	return "knowledge_doc"
}

func SuccessGraphStatus(status int) bool {
	return GraphStatus(status) == GraphSuccess
}

// BuildGraphShowStatus 报告展示状态 0:待处理，1.解析中，2.解析成功，3.解析失败 4. 文档处理失败，不显示图谱状态
func BuildGraphShowStatus(status GraphStatus, docStatus int) (int, string) {
	if docStatus == DocFail {
		return 4, "文档处理失败，不显示图谱状态"
	}
	switch status {
	case GraphInit:
		return 0, ""
	case GraphProcessing, GraphSchemaSuccess, GraphChunkSuccess, GraphExtractSuccess, GraphStoreSuccess:
		return 1, ""
	case GraphSuccess:
		return 2, ""
	}
	return 3, buildErrorMessage(status)
}

// todo 多语言没有处理
func buildErrorMessage(status GraphStatus) string {
	switch status {
	case GraphChunkFail:
		return "生成图谱获取chunk文本失败"
	case GraphExtractFail:
		return "提取图谱失败"
	case GraphStoreFail:
		return "图谱持久化存储失败"
	case GraphSchemaFail:
		return "graph schema 解析失败"
	case GraphInterruptFail:
		return "重启打断执行"
	}
	return ""
}

func InGraphStatus(status int) bool {
	graphStatus := GraphStatus(status)
	return graphStatus >= GraphSuccess && graphStatus <= GraphInterruptFail
}

// BuildGraphProgress 将图谱原始状态码映射为进度百分比 (0-100)，用于前端进度条展示。
// 同一进度值同时代表"该阶段完成"和"在下一阶段失败"，进度条停留在最后成功的位置。
func BuildGraphProgress(status GraphStatus) int {
	switch status {
	case GraphProcessing, GraphSchemaFail: // 开始解析 / schema 阶段失败
		return 5
	case GraphSchemaSuccess, GraphChunkFail: // schema 完成 / chunk 文本阶段失败
		return 25
	case GraphChunkSuccess, GraphExtractFail: // chunk 完成 / 提取阶段失败
		return 45
	case GraphExtractSuccess, GraphStoreFail: // 提取完成 / 持久化阶段失败
		return 65
	case GraphStoreSuccess: // 持久化完成
		return 85
	case GraphSuccess: // 全部完成
		return 100
	default: // 0 待处理 / 119 中断 / 未知码
		return 0
	}
}
