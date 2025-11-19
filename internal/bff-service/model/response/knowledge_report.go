package response

type KnowledgeReportPageResult struct {
	List          []*KnowledgeReportInfo `json:"list"`          // Community report content list
	Total         int32                  `json:"total"`         // Number of community reports: if 0 show -
	PageNo        int                    `json:"pageNo"`        // Current page number
	PageSize      int                    `json:"pageSize"`      // Quantity per page
	CreatedAt     string                 `json:"createdAt"`     // Generation time: unix millisecond timestamp, if it is an empty string, it will be displayed -
	Status        int32                  `json:"status"`        // Status: 0. Not generated (-) 1. Generating 2. Generated 3. Generation failed
	CanGenerate   bool                   `json:"canGenerate"`   // Whether it can be generated: true. Can be generated. False. Cannot be generated.
	CanAddReport  bool                   `json:"canAddReport"`  // Whether community reports can be added: true. Can be added. False. Cannot be added.
	GenerateLabel string                 `json:"generateLabel"` // Generate community report button copy: Generate/Regenerate
}

type KnowledgeReportInfo struct {
	ContentId string `json:"contentId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}
