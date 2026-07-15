package request

type OrgIDsReq struct {
	IsAllOrg  bool     `json:"isAllOrg"`            // 是否查询全部组织（用户有权限的组织）
	OrgIDList []string `json:"orgIdList,omitempty"` // 组织ID列表，isAllOrg=false 时必填
}

func (o *OrgIDsReq) Check() error {
	return nil
}

type UserCreate struct {
	OrgID string `json:"orgId" validate:"required"`
	UserInfo
}

func (u *UserCreate) Check() error {
	return nil
}

type UserUpdate struct {
	UserID string `json:"userId" validate:"required"` // 用户ID
	OrgID  string `json:"orgId" validate:"required"`  // 组织ID
	UserInfo
}

func (u *UserUpdate) Check() error {
	return nil
}

type UserInfo struct {
	UserName string   `json:"userName" validate:"required"` // 用户名
	Cipher   string   `json:"cipher" validate:"required"`   // RSA加密后的Base64字符串，包含{password, challenge}
	Phone    string   `json:"phone"`                        // 电话
	Remark   string   `json:"remark"`                       // 备注
	Email    string   `json:"email"`                        // 邮箱
	RoleIDs  []string `json:"roleIds" validate:"max=1"`     // 角色列表
	KeyID    string   `json:"keyId" validate:"required"`    // RSA公钥ID
}

type UserID struct {
	UserID string `json:"userId" validate:"required"` // 用户ID
}

type UserWithOrgID struct {
	UserID string `json:"userId" validate:"required"` // 用户ID
	OrgID  string `json:"orgId" validate:"required"`  // 组织ID
}

func (u *UserWithOrgID) Check() error {
	return nil
}

type UserDelete struct {
	UserWithOrgID
}

func (u *UserDelete) Check() error {
	return nil
}

type UserStatus struct {
	UserWithOrgID
	Status bool `json:"status"`
}

func (u *UserStatus) Check() error {
	return nil
}

type UserPassword struct {
	UserID
	OldCipher string `json:"oldCipher" validate:"required"` // 旧密码RSA加密后的Base64字符串，包含{password, challenge}
	NewCipher string `json:"newCipher" validate:"required"` // 新密码RSA加密后的Base64字符串，包含{password, challenge}
	KeyID     string `json:"keyId" validate:"required"`     // RSA公钥ID
}

func (u *UserPassword) Check() error {
	return nil
}

type UserPasswordByAdmin struct {
	UserWithOrgID
	Cipher string `json:"cipher" validate:"required"` // RSA加密后的Base64字符串，包含{password, challenge}
	KeyID  string `json:"keyId" validate:"required"`  // RSA公钥ID
}

func (u *UserPasswordByAdmin) Check() error {
	return nil
}
