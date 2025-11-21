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
	DocName       string `json:"docName"`       //file name
	DocType       string `json:"docType"`       //Document type
	KnowledgeId   string `json:"knowledgeId"`   //knowledge base id
	UploadTime    string `json:"uploadTime"`    //Upload time
	Status        int    `json:"status"`        //Processing status
	ErrorMsg      string `json:"errorMsg"`      //Parse error information, reserved
	FileSize      string `json:"fileSize"`      //File size, reserved
	SegmentMethod string `json:"segmentMethod"` //Segmentation mode 0: universal segmentation, 1: parent-child segmentation
	Author        string `json:"author"`        //Upload document author
	GraphStatus   int32  `json:"graphStatus"`   //Spectrum status 0: pending, 1. parsing, 2. parsing successful, 3. parsing failed
	GraphErrMsg   string `json:"graphErrMsg"`   //Plot error message
}

type DocImportTipResp struct {
	Message       string `json:"msg"`
	UploadStatus  int32  `json:"uploadstatus"`  //Upload status
	KnowledgeId   string `json:"knowledgeId"`   //knowledge base id
	KnowledgeName string `json:"knowledgeName"` //Knowledge base name
}

type DocSegmentResp struct {
	FileName            string            `json:"fileName"`            //name
	PageTotal           int               `json:"pageTotal"`           //Total pages
	SegmentTotalNum     int               `json:"segmentTotalNum"`     //Number of segments
	MaxSegmentSize      int               `json:"maxSegmentSize"`      //Set maximum length
	SegmentType         string            `json:"segmentType"`         //Segmentation mode 0 automatic segmentation 1 custom segmentation
	UploadTime          string            `json:"uploadTime"`          //Upload time
	Splitter            string            `json:"splitter"`            //Delimiter (required only for custom segments)
	MetaDataList        []*DocMetaData    `json:"metaDataList"`        //Document metadata
	SegmentContentList  []*SegmentContent `json:"contentList"`         //content
	SegmentImportStatus string            `json:"segmentImportStatus"` //Staged import status description
	SegmentMethod       string            `json:"segmentMethod"`       //Segmentation method: parent-child segmentation/universal segmentation
}

type DocMetaData struct {
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // Determine value
	MetaValueType string `json:"metaValueType"` // number，time, string
	MetaRule      string `json:"metaRule"`      // regular expression
	MetaId        string `json:"metaId"`        // metadata id
}

type SegmentContent struct {
	Content    string   `json:"content"`
	Available  bool     `json:"available"`
	ContentId  string   `json:"contentId"`
	ContentNum int      `json:"contentNum"`
	Labels     []string `json:"labels"`
	IsParent   bool     `json:"isParent"` // Parent-child segmentation/universal segmentation true is the parent segment, false is the universal segmentation
	ChildNum   int      `json:"childNum"` // Number of subsegments
}

type ChildSegmentInfo struct {
	Content  string `json:"content"`  // content
	ChildId  string `json:"childId"`  // subsegment id
	ChildNum int    `json:"childNum"` // subsegment number
	ParentId string `json:"parentId"` // parent segment id
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
	SegmentContentList []*ChildSegmentInfo `json:"contentList"` //content
}
