package response

// ParseTemplateBind 知识库上某文档类型选定的解析模板
type ParseTemplateBind struct {
	DocType    string `json:"docType"`
	TemplateId string `json:"templateId"`
}

type ParseTemplateListResp struct {
	List []*ParseTemplateInfo `json:"list"`
}

type ParseTemplateInfo struct {
	TemplateId        string           `json:"templateId"`
	Name              string           `json:"name"`
	DocType           string           `json:"docType"`
	DocSegment        *DocSegmentParam `json:"docSegment"`
	DocAnalyzer       []string         `json:"docAnalyzer"`
	ParserModelId     string           `json:"parserModelId"`
	AsrModelId        string           `json:"asrModelId"`
	MultimodalModelId string           `json:"multimodalModelId"`
	DocPreprocess     []string         `json:"docPreprocess"`
	UpdatedAt         string           `json:"updatedAt"`
	CreatedAt         string           `json:"createdAt"`
	BuiltIn           bool             `json:"builtIn"`
}

type DocSegmentParam struct {
	SegmentMethod  string   `json:"segmentMethod"`  // 分段方法 0：通用分段；1：父子分段
	SegmentType    string   `json:"segmentType"`    // 分段方式 0：自动分段；1：自定义分段
	Splitter       []string `json:"splitter"`       // 分隔符（只有自定义分段必填）
	MaxSplitter    int      `json:"maxSplitter"`    // 可分隔最大值（只有自定义分段必填）
	Overlap        float32  `json:"overlap"`        // 可重叠值（只有自定义分段必填）
	SubSplitter    []string `json:"subSplitter"`    // 分隔符（只有父子分段必填）
	SubMaxSplitter int      `json:"subMaxSplitter"` // 可分隔最大值（只有父子分段必填）
}
