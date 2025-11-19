package request

type UserCreate struct {
	Username string `json:"username" validate:"required"` // username
	UserInfo
}

func (u *UserCreate) Check() error {
	return nil
}

type UserUpdate struct {
	UserID string `json:"userId" validate:"required"` // User ID
	UserInfo
}

func (u *UserUpdate) Check() error {
	return nil
}

type UserInfo struct {
	Nickname string   `json:"nickname"`                     // Nick name
	Password string   `json:"password" validate:"required"` // password
	Phone    string   `json:"phone" validate:"required"`    // Telephone
	Remark   string   `json:"remark"`                       // Remark
	Gender   string   `json:"gender"`                       // Gender (0-female, 1-male, empty-unknown)
	Company  string   `json:"company"`                      // company
	RoleIDs  []string `json:"roleIds" validate:"max=1"`     // role list
}

type UserID struct {
	UserID string `json:"userId" validate:"required"` // User ID
}

func (u *UserID) Check() error {
	return nil
}

type UserStatus struct {
	UserID
	Status bool `json:"status"`
}

func (u *UserStatus) Check() error {
	return nil
}

type UserPassword struct {
	UserID
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

func (u *UserPassword) Check() error {
	return nil
}

type UserPasswordByAdmin struct {
	UserID
	Password string `json:"password" validate:"required"`
}

func (u *UserPasswordByAdmin) Check() error {
	return nil
}
