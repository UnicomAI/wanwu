package request

type OrgCreate struct {
	Name   string `json:"name" validate:"required"` // Organization name
	Remark string `json:"remark"`                   // Remark
}

func (o *OrgCreate) Check() error {
	return nil
}

type OrgUpdate struct {
	OrgID
	OrgCreate
}

func (o *OrgUpdate) Check() error {
	return nil
}

type OrgID struct {
	OrgID string `json:"orgId" validate:"required"` // Organization ID
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
	UserID
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
