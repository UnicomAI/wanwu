package response

type KnowledgeReportPageResult struct {
	List          []*KnowledgeReportInfo `json:"list"`          // 社区报告内容列表 [EN] Community report content list
	Total         int32                  `json:"total"`         // 社区报告数量：如果为0显示- [EN] Number of community reports: if 0 show -
	PageNo        int                    `json:"pageNo"`        // 当前页码 [EN] Current page number
	PageSize      int                    `json:"pageSize"`      // 每页数量 [EN] Quantity per page
	CreatedAt     string                 `json:"createdAt"`     // 生成时间：unix毫秒级时间戳，若为空串显示- [EN] Generation time: unix millisecond timestamp, if it is an empty string, it will be displayed -
	Status        int32                  `json:"status"`        // 状态：0.未生成(-) 1.生成中 2.已生成 3.生成失败 [EN] Status: 0. Not generated (-) 1. Generating 2. Generated 3. Generation failed
	CanGenerate   bool                   `json:"canGenerate"`   // 是否可生成：true.可生成 false.不可生成 [EN] Whether it can be generated: true. Can be generated. False. Cannot be generated.
	CanAddReport  bool                   `json:"canAddReport"`  // 是否可新增社区报告：true.可新增 false.不可新增 [EN] Whether community reports can be added: true. Can be added. False. Cannot be added.
	GenerateLabel string                 `json:"generateLabel"` // 生成社区报告按钮文案: 生成/重新生成 [EN] Generate community report button copy: Generate/Regenerate
}

type KnowledgeReportInfo struct {
	ContentId string `json:"contentId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}
