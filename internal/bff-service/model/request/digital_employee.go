package request

// DigitalEmployeeChatReq 数字员工发布对话请求（SSE，wga 模式）
// 会话需先通过 POST /digital-employee/conversation 创建，对话时 conversationId 必填（强制两步）
type DigitalEmployeeChatReq struct {
	EmployeeId     string                            `json:"employeeId" validate:"required"`     // 数字员工ID
	ConversationId string                            `json:"conversationId" validate:"required"` // 会话ID
	Messages       []GeneralAgentConversationMessage `json:"messages" validate:"required"`       // 消息
}

func (req *DigitalEmployeeChatReq) Check() error {
	return nil
}

// CreateDigitalEmployeeConversationReq 创建数字员工发布会话请求
type CreateDigitalEmployeeConversationReq struct {
	EmployeeId  string          `json:"employeeId" validate:"required"` // 数字员工ID
	Title       string          `json:"title" validate:"required"`      // 会话标题
	ModelConfig *AppModelConfig `json:"modelConfig"`                    // 模型配置（可选）
}

func (req *CreateDigitalEmployeeConversationReq) Check() error {
	return nil
}

// DeleteDigitalEmployeeConversationReq 删除数字员工发布会话请求
type DeleteDigitalEmployeeConversationReq struct {
	ConversationId string `json:"conversationId" validate:"required"` // 会话ID
}

func (req *DeleteDigitalEmployeeConversationReq) Check() error {
	return nil
}

// GetDigitalEmployeeConversationListReq 数字员工发布会话列表请求
type GetDigitalEmployeeConversationListReq struct {
	EmployeeId string `json:"employeeId" form:"employeeId" validate:"required"` // 数字员工ID
	PageNo     int    `json:"pageNo" form:"pageNo" validate:"required"`         // 页码
	PageSize   int    `json:"pageSize" form:"pageSize" validate:"required"`     // 每页数量
	SearchText string `json:"searchText" form:"searchText"`                     // 标题关键词，模糊匹配，空则不过滤
}

func (req *GetDigitalEmployeeConversationListReq) Check() error {
	return nil
}

// GetDigitalEmployeeConversationDetailReq 数字员工发布会话详情请求
type GetDigitalEmployeeConversationDetailReq struct {
	ConversationId string `json:"conversationId" form:"conversationId" validate:"required"` // 会话ID
}

func (req *GetDigitalEmployeeConversationDetailReq) Check() error {
	return nil
}

// DigitalEmployeePublishSyncReq 数字员工发布同步请求（外部系统回调 BFF callback 端口，万悟自拟格式）
// 仅支持发布（upsert app 表）；不支持取消发布
type DigitalEmployeePublishSyncReq struct {
	EmployeeId  string `json:"employeeId" validate:"required"`  // 数字员工ID（即 app 表 appId）
	PublishType string `json:"publishType" validate:"required"` // 发布类型(public/organization/private)
	UserId      string `json:"userId" validate:"required"`      // 发布者用户ID（外部系统提供）
	OrgId       string `json:"orgId" validate:"required"`       // 发布者组织ID（外部系统提供）
}

func (req *DigitalEmployeePublishSyncReq) Check() error {
	return nil
}

// DeleteDigitalEmployeePublishSyncReq 数字员工删除/下架同步请求（外部系统回调 BFF callback 端口，万悟自拟格式）
// 以 employeeId 为准删除 app 表行（幂等，目标不存在也返回成功）
type DeleteDigitalEmployeePublishSyncReq struct {
	EmployeeId string `json:"employeeId" validate:"required"` // 数字员工ID（即 app 表 appId）
	UserId     string `json:"userId" validate:"required"`     // 操作人用户ID（外部系统提供）
}

func (req *DeleteDigitalEmployeePublishSyncReq) Check() error {
	return nil
}

// GetDigitalEmployeeSquareDetailReq 数字员工广场详情请求
type GetDigitalEmployeeSquareDetailReq struct {
	EmployeeId string `json:"employeeId" form:"employeeId" validate:"required"` // 数字员工ID
}

func (req *GetDigitalEmployeeSquareDetailReq) Check() error {
	return nil
}

// GetDigitalEmployeeConversationConfigReq 数字员工发布会话配置查询请求
type GetDigitalEmployeeConversationConfigReq struct {
	ConversationId string `json:"conversationId" form:"conversationId" validate:"required"` // 会话ID
}

func (req *GetDigitalEmployeeConversationConfigReq) Check() error {
	return nil
}

// UpdateDigitalEmployeeConversationConfigReq 更新数字员工发布会话配置请求
type UpdateDigitalEmployeeConversationConfigReq struct {
	ConversationId string          `json:"conversationId" validate:"required"` // 会话ID
	ModelConfig    *AppModelConfig `json:"modelConfig" validate:"required"`    // 模型配置（必填）
}

func (req *UpdateDigitalEmployeeConversationConfigReq) Check() error {
	return nil
}
