package request

type UrlConversationCreateRequest struct {
	Prompt string `json:"prompt"  validate:"required"`
}

func (c *UrlConversationCreateRequest) Check() error { return nil }

type UrlConversationIdRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
	DetailId       string `json:"detailId" form:"detailId"` // 可选，传值则删除单条对话，不传则清空全部对话
}

func (c *UrlConversationIdRequest) Check() error { return nil }

type GetUrlConversationListReq struct {
	PageSearch
	SearchText string `json:"searchText" form:"searchText"` // 标题关键词，模糊匹配，空则不过滤
}

func (c *GetUrlConversationListReq) Check() error { return nil }

type UrlConversionStreamRequest struct {
	ConversationId string                 `json:"conversationId" form:"conversionId"`
	Prompt         string                 `json:"prompt" form:"prompt"  validate:"required"`
	FileInfo       []ConversionStreamFile `json:"fileInfo" form:"fileInfo"`
}

func (c *UrlConversionStreamRequest) Check() error {
	return nil
}

type UrlPendingConversionRequest struct {
	ConversationId string `json:"conversationId" form:"conversationId"  validate:"required"`
	CommonCheck
}

type UrlConversionStreamConnectRequest struct {
	UrlPendingConversionRequest
}

type UrlConversionStreamCancelRequest struct {
	UrlPendingConversionRequest
}

type UrlQuestionRecommendRequest struct {
	ConversationId string `json:"conversationId" form:"conversionId"`
	Query          string `json:"query" form:"query"  validate:"required"`
	CommonCheck
}
