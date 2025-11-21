package response

type RoleTemplate struct {
	Routes []Route `json:"routes"` // Level 1 routing
}

type Route struct {
	Name     string  `json:"name"`     // Route name
	Perm     string  `json:"perm"`     // Permissions
	Children []Route `json:"children"` // Subroutes
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
	IsAdmin   bool   `json:"isAdmin"` // Whether the organization has built-in administrator roles

	*RoleTemplate
	Permissions []Permission `json:"permissions"` // Permission list
}
