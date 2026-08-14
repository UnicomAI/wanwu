package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type OpenAPICreateRagResponse struct {
	UUID string `json:"uuid"`
}

// OpenAPIRagBriefInfo 知识问答列表条目（OpenAPI 专用，以 uuid 作为主键）
type OpenAPIRagBriefInfo struct {
	UUID        string         `json:"uuid"`        // 知识问答唯一标识，供后续接口使用
	Name        string         `json:"name"`        // 名称
	Desc        string         `json:"desc"`        // 描述
	Avatar      request.Avatar `json:"avatar"`      // 头像信息
	PublishType string         `json:"publishType"` // public/organization/private，空字符串表示未发布（草稿）
	Version     string         `json:"version"`     // 已发布版本号，未发布时为空
	CreatedAt   string         `json:"createdAt"`   // 创建时间
	UpdatedAt   string         `json:"updatedAt"`   // 最后更新时间
}

// OpenAPIRagListResponse 知识问答列表响应（非分页接口，全量返回）
type OpenAPIRagListResponse struct {
	List []OpenAPIRagBriefInfo `json:"list"`
}

type OpenAPIRagCreateConversationResponse struct {
	ConversationID string `json:"conversation_id"`
}
