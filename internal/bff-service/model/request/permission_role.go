package request

type RoleCreate struct {
	Name   string `json:"name" validate:"required"` // Character name
	Remark string `json:"remark"`                   // Remark

	Permissions []string // Permission list
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

type RoleID struct {
	RoleID string `json:"roleId" validate:"required"` // Role ID
}

func (r *RoleID) Check() error {
	return nil
}

type RoleStatus struct {
	RoleID
	Status bool `json:"status"`
}

func (r *RoleStatus) Check() error {
	return nil
}
