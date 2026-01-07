package model

import "github.com/UnicomAI/wanwu/pkg/db"

type GraphStatus int

const (
	DocWaitingForUpload = -2 //文档待上传
	DocInit             = 0  //文档待处理
	DocSuccess          = 1  //文档处理完成
	DocProcessing       = 3  //文档处理中
	DocFail             = 5  //文档待处理
	DocSuccessNew       = 10 //文档处理完成

	GraphInit          GraphStatus = 0   //图谱未处理
	GraphSuccess       GraphStatus = 100 //图谱生成成功
	GraphChunkFail     GraphStatus = 101 //图谱生成chunk文本失败
	GraphExtractFail   GraphStatus = 102 //图谱生成提取失败
	GraphStoreFail     GraphStatus = 103 //图谱持久化存储失败
	GraphProcessing    GraphStatus = 110 //图谱开始解析
	GraphInterruptFail GraphStatus = 119
)

type KnowledgeDoc struct {
	Id           uint32      `json:"id" gorm:"primary_key;type:bigint auto_increment;not null;comment:'id';"`   // Primary Key
	DocId        string      `gorm:"uniqueIndex:idx_unique_doc_id;column:doc_id;type:varchar(64)" json:"docId"` // Business Primary Key
	ImportTaskId string      `gorm:"column:batch_id;type:varchar(64);not null;default:'';comment:'导入的任务id'" json:"importTaskId"`
	KnowledgeId  string      `gorm:"column:knowledge_id;index:idx_user_id_knowledge_id_name,priority:2;index:idx_user_id_knowledge_id_tag,priority:2;type:varchar(64);not null;default:''" json:"knowledgeId"`
	FilePathMd5  string      `gorm:"column:file_path_md5;type:varchar(64);not null;default:'';comment:'文件的md5值'" json:"filePathMd5"`
	FilePath     string      `gorm:"column:file_path;type:text;not null" json:"filePath"`
	DirFilePath  string      `gorm:"column:dir_file_path;type:text;not null;comment:'文件在文件夹中的相对目录'" json:"dirFilePath"`
	Name         string      `gorm:"column:name;index:idx_user_id_knowledge_id_name,priority:3;type:varchar(256);not null;default:''" json:"name"`
	FileType     string      `gorm:"column:file_type;type:varchar(20);not null;default:''" json:"fileType"`
	FileSize     int64       `gorm:"column:file_size;type:bigint;COMMENT:'文件大小，单位byte'" json:"fileSize"`
	Status       int         `gorm:"column:status;not null;comment:'0-待处理， 1- 处理完成， 2-正在审核中(目前没有)，3-正在解析中，4-审核未通过（目前没有），5-解析失败';" json:"status"`
	GraphStatus  GraphStatus `gorm:"column:graph_status;type:int;not null;comment:'0-待处理， 100- 生成成功， 101-生成图谱获取chunk文本失败，102-提取图谱失败，103-图谱持久化存储失败，预留100~120';" json:"graphStatus"`
	ErrorMsg     db.LongText `gorm:"column:error_msg;not null;comment:'解析的错误信息'" json:"errorMsg"`
	CreatedAt    int64       `gorm:"column:create_at;type:bigint;autoCreateTime:milli;not null;" json:"createAt"` // Create Time
	UpdatedAt    int64       `gorm:"column:update_at;type:bigint;autoUpdateTime:milli;not null;" json:"updateAt"` // Update Time
	UserId       string      `gorm:"column:user_id;index:idx_user_id_knowledge_id_name,priority:1;index:idx_user_id_knowledge_id_tag,priority:1;type:varchar(64);not null;default:'';" json:"userId"`
	OrgId        string      `gorm:"column:org_id;type:varchar(64);not null;default:''" json:"orgId"`
	Deleted      int         `gorm:"column:deleted;not null;default:0;comment:'是否逻辑删除';" json:"deleted"`
}

func (KnowledgeDoc) TableName() string {
	return "knowledge_doc"
}

func SuccessGraphStatus(status int) bool {
	return GraphStatus(status) == GraphSuccess
}

// BuildGraphShowStatus 报告展示状态 0:待处理，1.解析中，2.解析成功，3.解析失败
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

// todo 多语言没有处理
func buildErrorMessage(status GraphStatus) string {
	switch status {
	case GraphChunkFail:
		return "图谱生成chunk文本失败"
	case GraphExtractFail:
		return "图谱生成提取失败"
	case GraphStoreFail:
		return "图谱持久化存储失败"
	case GraphInterruptFail:
		return "图谱中断失败"

	}
	return ""
}

func InGraphStatus(status int) bool {
	graphStatus := GraphStatus(status)
	return graphStatus >= GraphSuccess && graphStatus <= GraphInterruptFail
}
