package request

type ListKeywordsReq struct {
	PageSize int    `json:"pageSize" form:"pageSize"`
	PageNum  int    `json:"pageNo" form:"pageNo"`
	Name     string `json:"name" form:"name"` // 用于搜索【问题中的关键词】/【文档中的词语】 [EN] Used to search for [keywords in the question]/[words in the document]
	CommonCheck
}

type GetKeywordsDetailReq struct {
	Id uint32 `json:"id" form:"id"`
	CommonCheck
}

type CreateKeywordsReq struct {
	Name             string   `json:"name" validate:"required"`
	Alias            string   `json:"alias" validate:"required"`
	KnowledgeBaseIds []string `json:"knowledgeBaseIds" validate:"required"`
	CommonCheck
}

type UpdateKeywordsReq struct {
	Id               uint32   `json:"id"   validate:"required"`
	Name             string   `json:"name" validate:"required"`
	Alias            string   `json:"alias" validate:"required"`
	KnowledgeBaseIds []string `json:"knowledgeBaseIds" validate:"required"`
	CommonCheck
}

type DeleteKeywordsReq struct {
	Id uint32 `json:"id"`
	CommonCheck
}
