package request

// Login parameters of login request
type Login struct {
	Username string `json:"username" validate:"required"` // username
	Password string `json:"password" validate:"required"` // password
	Key      string `json:"key" validate:"required"`      // client key
	Code     string `json:"code" validate:"required"`     // Verification code
}

type RegisterByEmail struct {
	Username string `json:"username" validate:"required"` // username
	Email    string `json:"email" validate:"required"`    // Mail
	Code     string `json:"code" validate:"required"`     // Email verification code
}

type RegisterSendEmailCode struct {
	Username string `json:"username" validate:"required"` // username
	Email    string `json:"email" validate:"required"`    // Mail
}

type ResetPasswordSendEmailCode struct {
	Email string `json:"email" validate:"required"` // Mail
}

type ResetPasswordByEmail struct {
	Email    string `json:"email" validate:"required"`    // Mail
	Code     string `json:"code" validate:"required"`     // Email verification code
	Password string `json:"password" validate:"required"` // password
}

type LoginSendEmailCode struct {
	Email string `json:"email" validate:"required"` // Mail
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
