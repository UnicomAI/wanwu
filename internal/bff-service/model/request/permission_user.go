package request

type UserCreate struct {
	Username string `json:"username" validate:"required"` // 用户名 [EN] username
	UserInfo
}

func (u *UserCreate) Check() error {
	return nil
}

type UserUpdate struct {
	UserID string `json:"userId" validate:"required"` // 用户ID [EN] User ID
	UserInfo
}

func (u *UserUpdate) Check() error {
	return nil
}

type UserInfo struct {
	Nickname string   `json:"nickname"`                     // 昵称 [EN] Nick name
	Password string   `json:"password" validate:"required"` // 密码 [EN] password
	Phone    string   `json:"phone" validate:"required"`    // 电话 [EN] Telephone
	Remark   string   `json:"remark"`                       // 备注 [EN] Remark
	Gender   string   `json:"gender"`                       // 性别（0-女，1-男，空-未知） [EN] Gender (0-female, 1-male, empty-unknown)
	Company  string   `json:"company"`                      // 公司 [EN] company
	RoleIDs  []string `json:"roleIds" validate:"max=1"`     // 角色列表 [EN] role list
}

type UserID struct {
	UserID string `json:"userId" validate:"required"` // 用户ID [EN] User ID
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
