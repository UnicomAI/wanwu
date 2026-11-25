package config

var (
	topOrgID          string // 系统内部唯一顶级组织ID
	systemAdminUserID string // 系统内置管理员用户ID
)

func InitTopOrgID(orgID string) {
	if topOrgID != "" {
		return
	}
	topOrgID = orgID
}

func TopOrgID() string {
	return topOrgID
}

func InitSystemAdminUserID(userID string) {
	if systemAdminUserID != "" {
		return
	}
	systemAdminUserID = userID
}

func SystemAdminUserID() string {
	return systemAdminUserID
}

const (
	ModelScopeTypePrivate = "1" // 私有模型
	ModelScopeTypePublic  = "2" // 公有模型
	ModelScopeTypeOrg     = "3" // 组织模型
	ModelSourceExternal   = "external"
	ModelSourceBuiltin    = "builtin"
)
