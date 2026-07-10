package request

type OrgCreate struct {
	Name   string `json:"name" validate:"required"` // 组织名
	Remark string `json:"remark"`                   // 备注
	OrgID
	Avatar Avatar `json:"avatar"` // 组织头像
}

func (o *OrgCreate) Check() error {
	return nil
}

type OrgUpdate struct {
	OrgCreate
}

func (o *OrgUpdate) Check() error {
	return nil
}

type OrgID struct {
	OrgID string `json:"orgId" validate:"required"` // 组织ID
}

func (o *OrgID) Check() error {
	return nil
}

type OrgStatus struct {
	OrgID
	Status bool `json:"status"`
}

func (o *OrgStatus) Check() error {
	return nil
}

type OrgUserAdd struct {
	UserWithOrgID
	RoleID string `json:"roleId"`
}

func (o *OrgUserAdd) Check() error {
	return nil
}

type UserAvatarUpdate struct {
	Avatar Avatar `json:"avatar"`
}

func (u *UserAvatarUpdate) Check() error {
	return nil
}
