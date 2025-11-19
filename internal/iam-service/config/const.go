package config

var (
	topOrgID    uint32 // 系统内部唯一顶级组织ID [EN] The unique top-level organization ID within the system
	adminRoleID uint32 // 系统顶级组织内部管理员角色ID [EN] System top-level organization internal administrator role ID
	adminUserID uint32 // 系统顶级组织内部管理员用户ID [EN] System top-level organization internal administrator user ID
)

func InitTopOrgID(orgID uint32) {
	if topOrgID != 0 {
		return
	}
	topOrgID = orgID
}

func TopOrgID() uint32 {
	return topOrgID
}

func InitAdminRoleID(roleID uint32) {
	if adminRoleID != 0 {
		return
	}
	adminRoleID = roleID
}

func AdminRoleID() uint32 {
	return adminRoleID
}

func InitAdminUserID(userID uint32) {
	if adminUserID != 0 {
		return
	}
	adminUserID = userID
}

func AdminUserID() uint32 {
	return adminUserID
}
