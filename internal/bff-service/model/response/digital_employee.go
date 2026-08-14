package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// DigitalEmployeeBrief 数字员工简要信息（广场/下拉列表用）
type DigitalEmployeeBrief struct {
	ID     string         `json:"id"`     // 数字员工ID（外部）
	Name   string         `json:"name"`   // 名称
	Desc   string         `json:"desc"`   // 描述
	Avatar request.Avatar `json:"avatar"` // 头像
}

// DigitalEmployeeConversationInfo 数字员工发布会话列表项
type DigitalEmployeeConversationInfo struct {
	ConversationID string `json:"conversationId"` // 会话ID（即 wga threadId）
	EmployeeID     string `json:"employeeId"`     // 数字员工ID
	Title          string `json:"title"`          // 会话标题
	CreatedAt      string `json:"createdAt"`      // 创建时间
	UpdatedAt      string `json:"updatedAt"`      // 更新时间
}

// CreateDigitalEmployeeConversationResp 创建数字员工发布会话响应
type CreateDigitalEmployeeConversationResp struct {
	ConversationID string `json:"conversationId"` // 会话ID（即 wga threadId）
}

// GetDigitalEmployeeConversationConfigResp 数字员工发布会话配置响应
// 对齐通用智能体 GET /general/agent/conversation/config；模型无效/未配置时 ModelConfig 为空
type GetDigitalEmployeeConversationConfigResp struct {
	ConversationID string                 `json:"conversationId"` // 会话ID
	EmployeeID     string                 `json:"employeeId"`     // 数字员工ID
	Title          string                 `json:"title"`          // 会话标题
	ModelConfig    request.AppModelConfig `json:"modelConfig"`    // 模型配置
}

// DigitalEmployeeSquareDetail 数字员工广场详情（实时调外部详情）
// 当前前端仅消费 name/avatar，其余字段（role/task/workflow/skills/knowledge/作者信息）不返回
type DigitalEmployeeSquareDetail struct {
	Name        string         `json:"name"`        // 名称
	Avatar      request.Avatar `json:"avatar"`      // 头像
	Placeholder string         `json:"placeholder"` // 数字员工发布对话输入框占位文案（config 下发，区别于通用智能体 DIP Agent 的 placeholder）
}
