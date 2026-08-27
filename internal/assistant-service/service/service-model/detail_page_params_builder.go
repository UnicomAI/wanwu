package service_model

import assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"

type DetailPageParams struct {
	PageNum        int    `json:"pageNum"`
	PageSize       int    `json:"pageSize"`
	ConversationID string `json:"conversationID"`
	DetailID       string `json:"detailID"`
	UserID         string `json:"userID"`
	OrgID          string `json:"orgID"`
	HasDeleted     bool   `json:"hasDeleted"`   //true : 不过滤逻辑删除数据，false：过滤逻辑删除数据
	OpenMinioUrl   bool   `json:"openMinioUrl"` //true : 替换对外的url
}

// DetailPageParamsBuilder 用于以链式调用方式构造 DetailPageParams。
type DetailPageParamsBuilder struct {
	params DetailPageParams
}

// NewDetailPageParamsBuilder 创建并返回一个新的 DetailPageParamsBuilder。
func NewDetailPageParamsBuilder() *DetailPageParamsBuilder {
	return &DetailPageParamsBuilder{}
}

// WithPageParam 设置。
func (b *DetailPageParamsBuilder) WithPageParam(pageNum int32, pageSize int32) *DetailPageParamsBuilder {
	b.params.PageNum = int(pageNum)
	b.params.PageSize = int(pageSize)
	return b
}

// WithPageNum 设置页码。
func (b *DetailPageParamsBuilder) WithPageNum(pageNum int) *DetailPageParamsBuilder {
	b.params.PageNum = pageNum
	return b
}

// WithPageSize 设置每页大小。
func (b *DetailPageParamsBuilder) WithPageSize(pageSize int) *DetailPageParamsBuilder {
	b.params.PageSize = pageSize
	return b
}

// WithConversationID 设置会话 ID。
func (b *DetailPageParamsBuilder) WithConversationID(conversationID string) *DetailPageParamsBuilder {
	b.params.ConversationID = conversationID
	return b
}

// WithDetailID 设置详情 ID。
func (b *DetailPageParamsBuilder) WithDetailID(detailID string) *DetailPageParamsBuilder {
	b.params.DetailID = detailID
	return b
}

// WithIdentity 设置用户 数据。
func (b *DetailPageParamsBuilder) WithIdentity(identity *assistant_service.Identity) *DetailPageParamsBuilder {
	if identity != nil {
		b.params.UserID = identity.UserId
		b.params.OrgID = identity.OrgId
	}
	return b
}

// WithUserID 设置用户 ID。
func (b *DetailPageParamsBuilder) WithUserID(userID string) *DetailPageParamsBuilder {
	b.params.UserID = userID
	return b
}

// WithOrgID 设置组织 ID。
func (b *DetailPageParamsBuilder) WithOrgID(orgID string) *DetailPageParamsBuilder {
	b.params.OrgID = orgID
	return b
}

func (b *DetailPageParamsBuilder) WithHasDeleted(hasDeleted bool) *DetailPageParamsBuilder {
	b.params.HasDeleted = hasDeleted
	return b
}

func (b *DetailPageParamsBuilder) WithOpenMinioUrl(openMinioUrl bool) *DetailPageParamsBuilder {
	b.params.OpenMinioUrl = openMinioUrl
	return b
}

// Build 返回填充完成的 DetailPageParams 指针。
func (b *DetailPageParamsBuilder) Build() *DetailPageParams {
	return &b.params
}
