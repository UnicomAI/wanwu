package request

// ParseTemplateBind 知识库上某文档类型选定的解析模板
type ParseTemplateBind struct {
	DocType    string `json:"docType" validate:"required"`
	TemplateId string `json:"templateId" validate:"required"`
}

type ParseTemplateConfig struct {
	DocSegment        *DocSegment `json:"docSegment" validate:"required"` //文档分段配置
	DocAnalyzer       []string    `json:"docAnalyzer"`                    //文档解析类型 text / model
	ParserModelId     string      `json:"parserModelId"`                  //模型解析或ocr模型id
	AsrModelId        string      `json:"asrModelId"`                     //asr模型id
	MultimodalModelId string      `json:"multimodalModelId"`              //多模态模型id
	DocPreprocess     []string    `json:"docPreprocess"`                  //文本预处理规则 replaceSymbols / deleteLinks
}

type ParseTemplateListReq struct {
	DocType string `json:"docType" form:"docType"` //文档类型，为空则返回全部
	CommonCheck
}

type CreateParseTemplateReq struct {
	Name    string `json:"name" validate:"required"`
	DocType string `json:"docType" validate:"required"`
	ParseTemplateConfig
	CommonCheck
}

type UpdateParseTemplateReq struct {
	TemplateId string `json:"templateId" validate:"required"`
	Name       string `json:"name" validate:"required"`
	ParseTemplateConfig
	CommonCheck
}

type DeleteParseTemplateReq struct {
	TemplateId string `json:"templateId" validate:"required"`
	CommonCheck
}
