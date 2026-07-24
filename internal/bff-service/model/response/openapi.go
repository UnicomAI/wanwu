package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type OpenAPICreateAgentResponse struct {
	UUID string `json:"uuid"`
}

// OpenAPIAgentBriefInfo 智能体列表条目（OpenAPI 专用，以 uuid 作为主键）
type OpenAPIAgentBriefInfo struct {
	UUID        string         `json:"uuid"`        // 智能体唯一标识，供后续接口使用
	Name        string         `json:"name"`        // 名称
	Desc        string         `json:"desc"`        // 描述
	Avatar      request.Avatar `json:"avatar"`      // 头像信息
	Category    int32          `json:"category"`    // 1:单智能体 2:多智能体
	PublishType string         `json:"publishType"` // public/organization/private，空字符串表示未发布（草稿）
	Version     string         `json:"version"`     // 已发布版本号，未发布时为空
	CreatedAt   string         `json:"createdAt"`   // 创建时间
	UpdatedAt   string         `json:"updatedAt"`   // 最后更新时间
}

// OpenAPIAgentListResponse 智能体列表响应（非分页接口，全量返回，调用方按需 len(list) 计数即可）
type OpenAPIAgentListResponse struct {
	List []OpenAPIAgentBriefInfo `json:"list"`
}

type OpenAPIAgentCreateConversationResponse struct {
	ConversationID string `json:"conversation_id"`
}

type OpenAPIAgentChatResponse struct {
	Code               int                    `json:"code"`
	Message            string                 `json:"message"`
	Response           string                 `json:"response"`
	GenFileUrlList     []OpenAPIAgentChatFile `json:"gen_file_url_list"`
	SearchList         []OpenAIChatSearch     `json:"search_list"`
	History            []OpenAIChatHistory    `json:"history"`
	Usage              OpenAPIAgentChatUsage  `json:"usage"`
	Finish             int                    `json:"finish"`
	RecommendQuestions []string               `json:"recommend_questions,omitempty"`
}

type OpenAPIAgentChatFile struct {
	OutputFileUrl string `json:"output_file_url"`
}

type OpenAPIAgentChatUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAPIRagChatResponse struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	MsgID   string              `json:"msg_id"`
	Data    OpenAPIRagChatData  `json:"data"`
	History []OpenAIChatHistory `json:"history"`
	Finish  int                 `json:"finish"`
}

type OpenAPIRagChatData struct {
	Output     string             `json:"output"`
	SearchList []OpenAIChatSearch `json:"searchList"`
}

type OpenAIChatSearch struct {
	KBName  string `json:"kb_name"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	// MetaData 知识库命中片段的元信息，原样透传 rag 返回（与流式一致），如 file_name/download_link/page_num/chunk_total_num 等
	MetaData interface{} `json:"meta_data,omitempty"`
}

type OpenAIChatHistory struct {
	Query    string `json:"query"`
	Response string `json:"response"`
}

type OpenAPIChatflowCreateConversationResponse struct {
	ConversationId string `json:"conversation_id"`
}

type OpenAPIChatflowGetConversationMessageListResponse struct {
	Messages []*OpenMessageApi `json:"data"`
	HasMore  bool              `json:"has_more"`
	FirstID  int64             `json:"first_id"`
	LastID   int64             `json:"last_id"`
}

type OpenMessageApi struct {
	ID               int64             `json:"id,string"`
	BotID            int64             `json:"bot_id,string"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ConversationID   int64             `json:"conversation_id,string"`
	MetaData         map[string]string `json:"meta_data"`
	CreatedAt        int64             `json:"created_at"`
	UpdatedAt        int64             `json:"updated_at"`
	ChatID           int64             `json:"chat_id,string"`
	ContentType      string            `json:"content_type"`
	Type             string            `json:"type"`
	SectionID        string            `json:"section_id"`
	ReasoningContent string            `json:"reasoning_content"`
}
