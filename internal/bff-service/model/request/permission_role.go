package request

type RoleWithOrgID struct {
	RoleID string `json:"roleId" validate:"required"` // 角色ID
	OrgID  string `json:"orgId" validate:"required"`  // 组织ID
}

func (u *RoleWithOrgID) Check() error {
	return nil
}

type RoleCreate struct {
	Name        string   `json:"name" validate:"required"`  // 角色名
	Remark      string   `json:"remark"`                    // 备注
	IsGlobal    bool     `json:"isGlobal"`                  // 是否全局角色
	OrgID       string   `json:"orgId" validate:"required"` // 组织ID
	Permissions []string // 权限列表
	Avatar      Avatar   `json:"avatar"` // 角色头像
}

func (r *RoleCreate) Check() error {
	return nil
}

type RoleUpdate struct {
	RoleID string `json:"roleId" validate:"required"`
	RoleCreate
}

func (r *RoleUpdate) Check() error {
	return nil
}

type RoleDelete struct {
	RoleWithOrgID
}

func (r *RoleDelete) Check() error {
	return nil
}

type RoleStatus struct {
	RoleWithOrgID
	Status bool `json:"status"`
}

func (r *RoleStatus) Check() error {
	return nil
}

type RoleUserRemove struct {
	RoleWithOrgID
	UserID string `json:"userId" validate:"required"` // 用户ID
}

func (r *RoleUserRemove) Check() error {
	return nil
}
