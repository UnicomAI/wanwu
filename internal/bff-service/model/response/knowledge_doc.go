package response

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type DocConfigResult struct {
	DocImportType     int32       `json:"docImportType"`     //文档导入类型，0：文件上传，1：url上传，2.批量url上传
	DocSegment        *DocSegment `json:"docSegment"`        //分段信息配置
	DocAnalyzer       []string    `json:"docAnalyzer"`       //文档解析类型
	ParserModelId     string      `json:"parserModelId"`     //ocr模型id
	AsrModelId        string      `json:"asrModelId"`        //asr模型id
	MultimodalModelId string      `json:"multimodalModelId"` //多模态模型id
	DocPreprocess     []string    `json:"docPreprocess"`     //文本预处理规则
}

type DocSegment struct {
	SegmentMethod  string   `json:"segmentMethod" validate:"required"` // 分段方法 0：通用分段；1：父子分段
	SegmentType    string   `json:"segmentType"`                       // 分段方式，只有通用分段必填 0：自动分段；1：自定义分段
	Splitter       []string `json:"splitter,omitempty"`                // 分隔符（只有自定义分段必填）
	MaxSplitter    *int     `json:"maxSplitter,omitempty"`             // 可分隔最大值（只有自定义分段必填）
	Overlap        *float32 `json:"overlap,omitempty"`                 // 可重叠值（只有自定义分段必填）
	SubSplitter    []string `json:"subSplitter,omitempty"`             // 分隔符（只有父子分段必填）
	SubMaxSplitter *int     `json:"subMaxSplitter,omitempty"`          // 可分隔最大值（只有父子分段必填）
}

type DocKnowledgeInfo struct {
	KnowledgeId     string          `json:"knowledgeId"`
	KnowledgeName   string          `json:"knowledgeName"`
	GraphSwitch     int32           `json:"graphSwitch"`
	ShowGraphReport bool            `json:"showGraphReport"`
	Description     string          `json:"description"`
	Keywords        []*KeywordsInfo `json:"keywords"`
	EmbeddingModel  *ModelInfo      `json:"embeddingModel"`
	LlmModelId      string          `json:"llmModelId"`
	Category        int32           `json:"category"`       // 0: 知识库 1: 问答库 2: 多模态知识库
	Avatar          request.Avatar  `json:"avatar"`         // 头像
	PermissionType  int32           `json:"permissionType"` // 当前用户对该知识库的权限类型 -1:无权限 0:查看 10:编辑 20:授权 30:系统管理授权
	OwnerUserId     string          `json:"ownerUserId"`    // 知识库拥有者的userID，因为知识库可以转让，所以拥有者未必是创建者
	OwnerOrgId      string          `json:"ownerOrgId"`     // 知识库拥有者的orgID，因为知识库可以转让，所以拥有者未必是创建者
	CreatedAt       string          `json:"createdAt"`      // 创建时间
	UpdatedAt       string          `json:"updatedAt"`      // 更新时间
}

type ListDocResp struct {
	DocId         string `json:"docId"`
	DocName       string `json:"docName"`       //文档名称
	DocType       string `json:"docType"`       //文档类型
	KnowledgeId   string `json:"knowledgeId"`   //知识库id
	UploadTime    string `json:"uploadTime"`    //上传时间
	Status        int    `json:"status"`        //处理状态
	ErrorMsg      string `json:"errorMsg"`      //解析错误信息，预留
	FileSize      int64  `json:"fileSize"`      //文件大小，单位字节(Byte)
	FileSizeStr   string `json:"fileSizeStr"`   //文件大小的人眼友好展示
	SegmentMethod string `json:"segmentMethod"` //分段模式 0:通用分段，1：父子分段
	Author        string `json:"author"`        //上传文档 作者
	GraphStatus   int32  `json:"graphStatus"`   //图谱状态 0:待处理，1.解析中，2.解析成功，3.解析失败 -1. 当文档状态为解析失败时，显示 -
	GraphErrMsg   string `json:"graphErrMsg"`   //图谱错误信息
	IsMultimodal  bool   `json:"isMultimodal"`  // 是否为多模态文件
	DocProgress   int32  `json:"docProgress"`   //文档解析进度百分比 (0-100)
	GraphProgress int32  `json:"graphProgress"` //图谱解析进度百分比 (0-100)
}

type DocImportTipResp struct {
	Message       string `json:"msg"`
	UploadStatus  int32  `json:"uploadstatus"`  //上传状态
	KnowledgeId   string `json:"knowledgeId"`   //知识库id
	KnowledgeName string `json:"knowledgeName"` //知识库名称
}

type DocSegmentResp struct {
	FileName            string                 `json:"fileName"`            //名称
	FileSize            int64                  `json:"fileSize"`            //文件大小，单位字节(Byte)
	FileSizeStr         string                 `json:"fileSizeStr"`         //文件大小的人眼友好展示
	PageTotal           int                    `json:"pageTotal"`           //总页数
	SegmentTotalNum     int                    `json:"segmentTotalNum"`     //分段数量
	MaxSegmentSize      int                    `json:"maxSegmentSize"`      //设置最大长度
	SegmentType         string                 `json:"segmentType"`         //分段方式 0自动分段 1自定义分段
	UploadTime          string                 `json:"uploadTime"`          //上传时间
	Splitter            string                 `json:"splitter"`            //分隔符（只有自定义分段必填）
	MetaDataList        []*DocMetaData         `json:"metaDataList"`        //文档元数据
	SegmentContentList  []*SegmentContent      `json:"contentList"`         //内容
	SegmentImportStatus string                 `json:"segmentImportStatus"` //分段导入状态描述
	SegmentMethod       string                 `json:"segmentMethod"`       //分段方式 父子分段/通用分段
	DocAnalyzerText     []*DocAnalyzerTextInfo `json:"docAnalyzerText"`     //文档解析类型 文字提取 / OCR解析  / 模型解析
	DownloadUrl         string                 `json:"downloadUrl"`         //文档对外下载地址
	CanPreview          bool                   `json:"canPreview"`          //是否可预览
	PreviewFailReason   string                 `json:"previewFailReason"`   //不可预览原因
}

type DocAnalyzerTextInfo struct {
	Text        string          `json:"text"`
	DisplayName string          `json:"displayName,omitempty"`
	Tags        []mp_common.Tag `json:"tags,omitempty"`
}

type DocMetaData struct {
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // 确定值
	MetaValueType string `json:"metaValueType"` // number，time, string
	MetaRule      string `json:"metaRule"`      // 正则表达式
	MetaId        string `json:"metaId"`        // 元数据id
}

type SegmentContent struct {
	Content    string   `json:"content"`
	Available  bool     `json:"available"`
	ContentId  string   `json:"contentId"`
	ContentNum int      `json:"contentNum"`
	Labels     []string `json:"labels"`
	IsParent   bool     `json:"isParent"` // 父子分段/通用分段 true是父分段，false是通用分段
	ChildNum   int      `json:"childNum"` // 子分段数量
}

type ChildSegmentInfo struct {
	Content  string `json:"content"`  // 内容
	ChildId  string `json:"childId"`  // 子分段id
	ChildNum int    `json:"childNum"` // 子分段序号
	ParentId string `json:"parentId"` // 父分段id
}

type AnalysisDocUrlResp struct {
	UrlList []*DocUrl `json:"urlList"`
}

type DocUrl struct {
	Url      string `json:"url"`
	FileName string `json:"fileName"`
	FileSize int    `json:"fileSize"`
}

type DocChildSegmentResp struct {
	SegmentContentList []*ChildSegmentInfo `json:"contentList"` //内容
}

type UpdateDocSegmentResp struct {
	ContentId string `json:"contentId"` // 更新后的分段id，内容变化时与请求传入的contentId不同
}

type DocUploadLimitResp struct {
	UploadLimitList []*DocUploadLimit `json:"uploadLimitList"`
}

type DocUploadLimit struct {
	FileType string   `json:"fileType"` // 文件类型 图片：image 视频：video
	MaxSize  int      `json:"maxSize"`  // 文件大小限制，单位MB
	ExtList  []string `json:"extList"`  // 文件后缀列表
}

const (
	fileSizeKB = 1024
	fileSizeMB = fileSizeKB * 1024
	fileSizeGB = fileSizeMB * 1024
)

// FormatFileSize 将字节大小转换为人眼友好的展示字符串。
// 规则：fileSize <= 0 返回 "--"；其余最小以 K 展示，下限 0.01 K；之后依次为 M、G，保留两位小数。
func FormatFileSize(fileSize int64) string {
	if fileSize <= 0 {
		return "--"
	}
	switch {
	case fileSize < fileSizeMB:
		return fmt.Sprintf("%.2f K", maxFloat(float64(fileSize)/float64(fileSizeKB), 0.01))
	case fileSize < fileSizeGB:
		return fmt.Sprintf("%.2f M", float64(fileSize)/float64(fileSizeMB))
	default:
		return fmt.Sprintf("%.2f G", float64(fileSize)/float64(fileSizeGB))
	}
}

// maxFloat 返回两个浮点数中的较大值，用于约束最小显示为 0.01 K。
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
