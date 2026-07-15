package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type UserID struct {
	UserID string `json:"userId"`
}

type Users struct {
	Users []IDName `json:"users"` // 去重后的用户列表
}

type UserInfo struct {
	UserID    string         `json:"userId"`
	Username  string         `json:"username"`
	Phone     string         `json:"phone"`
	Email     string         `json:"email"`
	CreatedAt string         `json:"createdAt"`
	Creator   IDName         `json:"creator"` // 创建人
	Status    bool           `json:"status"`
	Language  Language       `json:"language"`
	Orgs      []OrgRole      `json:"orgs"` // 用户的组织角色列表
	Avatar    request.Avatar `json:"avatar"`
}

type OrgRole struct {
	Org   IDName       `json:"org"`   // 组织
	Roles []RoleIDName `json:"roles"` // 角色列表
}

// UserBatchImportResult 批量导入用户结果
type UserBatchImportResult struct {
	Total   int                    `json:"total"`            // 总数
	Success int                    `json:"success"`          // 成功数
	Failed  int                    `json:"failed"`           // 失败数
	Errors  []UserBatchImportError `json:"errors,omitempty"` // 失败详情
}

// UserBatchImportError 导入错误详情
type UserBatchImportError struct {
	Row      int    `json:"row"`      // Excel行号（从2开始，1是表头）
	Username string `json:"username"` // 用户名
	Reason   string `json:"reason"`   // 失败原因
}
