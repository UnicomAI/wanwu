package response

type DocPageResult struct {
	List             []*ListDocResp    `json:"list"`
	Total            int64             `json:"total"`
	PageNo           int               `json:"pageNo"`
	PageSize         int               `json:"pageSize"`
	DocKnowledgeInfo *DocKnowledgeInfo `json:"docKnowledgeInfo"`
}

type DocKnowledgeInfo struct {
	KnowledgeId     string `json:"knowledgeId"`
	KnowledgeName   string `json:"knowledgeName"`
	GraphSwitch     int32  `json:"graphSwitch"`
	ShowGraphReport bool   `json:"showGraphReport"`
}

type ListDocResp struct {
	DocId         string `json:"docId"`
	DocName       string `json:"docName"`       //文档名称 [EN] file name
	DocType       string `json:"docType"`       //文档类型 [EN] Document type
	KnowledgeId   string `json:"knowledgeId"`   //知识库id [EN] knowledge base id
	UploadTime    string `json:"uploadTime"`    //上传时间 [EN] Upload time
	Status        int    `json:"status"`        //处理状态 [EN] Processing status
	ErrorMsg      string `json:"errorMsg"`      //解析错误信息，预留 [EN] Parse error information, reserved
	FileSize      string `json:"fileSize"`      //文件大小，预留 [EN] File size, reserved
	SegmentMethod string `json:"segmentMethod"` //分段模式 0:通用分段，1：父子分段 [EN] Segmentation mode 0: universal segmentation, 1: parent-child segmentation
	Author        string `json:"author"`        //上传文档 作者 [EN] Upload document author
	GraphStatus   int32  `json:"graphStatus"`   //图谱状态 0:待处理，1.解析中，2.解析成功，3.解析失败 [EN] Spectrum status 0: pending, 1. parsing, 2. parsing successful, 3. parsing failed
	GraphErrMsg   string `json:"graphErrMsg"`   //图谱错误信息 [EN] Plot error message
}

type DocImportTipResp struct {
	Message       string `json:"msg"`
	UploadStatus  int32  `json:"uploadstatus"`  //上传状态 [EN] Upload status
	KnowledgeId   string `json:"knowledgeId"`   //知识库id [EN] knowledge base id
	KnowledgeName string `json:"knowledgeName"` //知识库名称 [EN] Knowledge base name
}

type DocSegmentResp struct {
	FileName            string            `json:"fileName"`            //名称 [EN] name
	PageTotal           int               `json:"pageTotal"`           //总页数 [EN] Total pages
	SegmentTotalNum     int               `json:"segmentTotalNum"`     //分段数量 [EN] Number of segments
	MaxSegmentSize      int               `json:"maxSegmentSize"`      //设置最大长度 [EN] Set maximum length
	SegmentType         string            `json:"segmentType"`         //分段方式 0自动分段 1自定义分段 [EN] Segmentation mode 0 automatic segmentation 1 custom segmentation
	UploadTime          string            `json:"uploadTime"`          //上传时间 [EN] Upload time
	Splitter            string            `json:"splitter"`            //分隔符（只有自定义分段必填） [EN] Delimiter (required only for custom segments)
	MetaDataList        []*DocMetaData    `json:"metaDataList"`        //文档元数据 [EN] Document metadata
	SegmentContentList  []*SegmentContent `json:"contentList"`         //内容 [EN] content
	SegmentImportStatus string            `json:"segmentImportStatus"` //分段导入状态描述 [EN] Staged import status description
	SegmentMethod       string            `json:"segmentMethod"`       //分段方式 父子分段/通用分段 [EN] Segmentation method: parent-child segmentation/universal segmentation
}

type DocMetaData struct {
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // 确定值 [EN] Determine value
	MetaValueType string `json:"metaValueType"` // number，time, string
	MetaRule      string `json:"metaRule"`      // 正则表达式 [EN] regular expression
	MetaId        string `json:"metaId"`        // 元数据id [EN] metadata id
}

type SegmentContent struct {
	Content    string   `json:"content"`
	Available  bool     `json:"available"`
	ContentId  string   `json:"contentId"`
	ContentNum int      `json:"contentNum"`
	Labels     []string `json:"labels"`
	IsParent   bool     `json:"isParent"` // 父子分段/通用分段 true是父分段，false是通用分段 [EN] Parent-child segmentation/universal segmentation true is the parent segment, false is the universal segmentation
	ChildNum   int      `json:"childNum"` // 子分段数量 [EN] Number of subsegments
}

type ChildSegmentInfo struct {
	Content  string `json:"content"`  // 内容 [EN] content
	ChildId  string `json:"childId"`  // 子分段id [EN] subsegment id
	ChildNum int    `json:"childNum"` // 子分段序号 [EN] subsegment number
	ParentId string `json:"parentId"` // 父分段id [EN] parent segment id
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
	SegmentContentList []*ChildSegmentInfo `json:"contentList"` //内容 [EN] content
}
