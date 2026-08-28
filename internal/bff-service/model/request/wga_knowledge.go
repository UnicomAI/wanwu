package request

// WgaRagSearchKnowledgeBaseReq WGA知识库检索请求
type WgaRagSearchKnowledgeBaseReq struct {
	KnowledgeIdList []string `json:"knowledgeIdList" validate:"required"`
	Question        string   `json:"question" validate:"required"`
}

func (r *WgaRagSearchKnowledgeBaseReq) Check() error {
	return nil
}
