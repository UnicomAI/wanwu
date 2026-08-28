package model

type RagKnowledgeHitResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *KnowledgeHitData `json:"data"`
}

type KnowledgeHitData struct {
	Prompt       string             `json:"prompt"`
	SearchList   []*ChunkSearchList `json:"searchList"`
	Score        []float64          `json:"score"`
	AutoCitation bool               `json:"autoCitation"` // 是否自动角标
}

type ChunkSearchList struct {
	Title            string          `json:"title"`
	Snippet          string          `json:"snippet"`
	KbName           string          `json:"kb_name"`
	UserKbName       string          `json:"user_kb_name"`
	MetaData         *MetaData       `json:"meta_data"`
	ChildContentList []*ChildContent `json:"child_content_list"`
	ChildScore       []float64       `json:"child_score"`
	Score            float64         `json:"score"`
}

type ChildContent struct {
	ChildSnippet string  `json:"child_snippet"`
	Score        float64 `json:"score"`
}

type MetaData struct {
	PageNum         []interface{} `json:"page_num"`
	ParentTitle     []interface{} `json:"parent_title"`
	FileName        string        `json:"file_name"`
	DownloadLink    string        `json:"download_link"`
	RowNum          int           `json:"row_num"`
	ChunkCurrentNum int           `json:"chunk_current_num"`
	ChunkTotalNum   int           `json:"chunk_total_num"`
	BucketName      string        `json:"bucket_name"`
	ObjectName      string        `json:"object_name"`
	DocMeta         []interface{} `json:"doc_meta"`
}
