package response

type KnowledgeSplitterListResp struct {
	KnowledgeSplitterList []*KnowledgeSplitter `json:"knowledgeSplitterList"`
}

type KnowledgeSplitter struct {
	SplitterId    string `json:"splitterId"`    //knowledge base separator id
	SplitterName  string `json:"splitterName"`  //Knowledge base separator name
	SplitterValue string `json:"splitterValue"` //Knowledge base delimiter value
	Type          string `json:"type"`          //Separator type: preset/custom
}
