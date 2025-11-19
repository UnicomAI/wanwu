package request

// Login 登录请求的参数 [EN] Login parameters of login request
type Login struct {
	Username string `json:"username" validate:"required"` // 用户名 [EN] username
	Password string `json:"password" validate:"required"` // 密码 [EN] password
	Key      string `json:"key" validate:"required"`      // 客户端key [EN] client key
	Code     string `json:"code" validate:"required"`     // 验证码 [EN] Verification code
}

type RegisterByEmail struct {
	Username string `json:"username" validate:"required"` // 用户名 [EN] username
	Email    string `json:"email" validate:"required"`    // 邮箱 [EN] Mail
	Code     string `json:"code" validate:"required"`     // 邮箱验证码 [EN] Email verification code
}

type RegisterSendEmailCode struct {
	Username string `json:"username" validate:"required"` // 用户名 [EN] username
	Email    string `json:"email" validate:"required"`    // 邮箱 [EN] Mail
}

type ResetPasswordSendEmailCode struct {
	Email string `json:"email" validate:"required"` // 邮箱 [EN] Mail
}

type ResetPasswordByEmail struct {
	Email    string `json:"email" validate:"required"`    // 邮箱 [EN] Mail
	Code     string `json:"code" validate:"required"`     // 邮箱验证码 [EN] Email verification code
	Password string `json:"password" validate:"required"` // 密码 [EN] password
}

type LoginSendEmailCode struct {
	Email string `json:"email" validate:"required"` // 邮箱 [EN] Mail
}

func (l *Login) Check() error {
	return nil
}

func (l *RegisterByEmail) Check() error {
	return nil
}

func (l *RegisterSendEmailCode) Check() error {
	return nil
}

func (r *ResetPasswordSendEmailCode) Check() error {
	return nil
}

func (r *ResetPasswordByEmail) Check() error {
	return nil
}

func (l *LoginSendEmailCode) Check() error {
	return nil
}
