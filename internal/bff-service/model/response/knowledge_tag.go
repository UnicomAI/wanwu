package response

type KnowledgeTagListResp struct {
	KnowledgeTagList []*KnowledgeTag `json:"knowledgeTagList"`
}

type TagBindResp struct {
	BindCount int64 `json:"tagBindCount"`
}

type KnowledgeTag struct {
	TagId    string `json:"tagId"`    //Knowledge base tag id
	TagName  string `json:"tagName"`  //Knowledge base tag name
	Selected bool   `json:"selected"` //Is this table label selected?
}

type CreateKnowledgeTagResp struct {
	KnowledgeId string `json:"knowledgeId"`
}
