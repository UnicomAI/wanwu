package response

type RoleTemplate struct {
	Routes []Route `json:"routes"` // 一级路由 [EN] Level 1 routing
}

type Route struct {
	Name     string  `json:"name"`     // 路由名 [EN] Route name
	Perm     string  `json:"perm"`     // 权限 [EN] Permissions
	Children []Route `json:"children"` // 子路由 [EN] Subroutes
}

type RoleID struct {
	RoleID string `json:"roleId"`
}

type RoleInfo struct {
	RoleID    string `json:"roleId"`
	Name      string `json:"name"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"createdAt"`
	Creator   IDName `json:"creator"`
	Status    bool   `json:"status"`
	IsAdmin   bool   `json:"isAdmin"` // 是否组织内置管理员角色 [EN] Whether the organization has built-in administrator roles

	*RoleTemplate
	Permissions []Permission `json:"permissions"` // 权限列表 [EN] Permission list
}
