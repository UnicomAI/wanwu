package response

type KnowledgeTagListResp struct {
	KnowledgeTagList []*KnowledgeTag `json:"knowledgeTagList"`
}

type TagBindResp struct {
	BindCount int64 `json:"tagBindCount"`
}

type KnowledgeTag struct {
	TagId    string `json:"tagId"`    //知识库标签id [EN] Knowledge base tag id
	TagName  string `json:"tagName"`  //知识库标签名称 [EN] Knowledge base tag name
	Selected bool   `json:"selected"` //此表标签是否选中 [EN] Is this table label selected?
}

type CreateKnowledgeTagResp struct {
	KnowledgeId string `json:"knowledgeId"`
}
