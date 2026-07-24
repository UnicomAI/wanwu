package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type RoleTemplate struct {
	Routes []Route `json:"routes"` // 一级路由
}

type Route struct {
	Name     string  `json:"name"`     // 路由名
	Perm     string  `json:"perm"`     // 权限
	Children []Route `json:"children"` // 子路由
}

type RoleID struct {
	RoleID string `json:"roleId"`
}

type RoleInfo struct {
	RoleID    string           `json:"roleId"`
	Name      string           `json:"name"`
	Remark    string           `json:"remark"`
	CreatedAt string           `json:"createdAt"`
	Creator   IDNameWithAvatar `json:"creator"`
	Status    bool             `json:"status"`
	IsAdmin   bool             `json:"isAdmin"`   // 是否组织内置管理员角色
	IsGlobal  bool             `json:"isGlobal"`  // 是否全局角色
	OrgName   string           `json:"orgName"`   // 角色所属组织名（全局角色为空，组织角色为所属组织全名）
	UserCount int32            `json:"userCount"` // 被赋予该角色的用户数量
	Avatar    request.Avatar   `json:"avatar"`    // 角色头像

	*RoleTemplate
	Permissions []Permission `json:"permissions"` // 权限列表
}

type RoleUser struct {
	UserID   string             `json:"userId"`
	UserName string             `json:"userName"`
	Phone    string             `json:"phone"`
	Email    string             `json:"email"`
	Orgs     []IDNameWithAvatar `json:"orgs"`
	Avatar   request.Avatar     `json:"avatar"` // 用户头像
}
