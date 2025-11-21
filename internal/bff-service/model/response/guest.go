package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type Login struct {
	UID              string            `json:"uid"`
	Username         string            `json:"username"`
	Token            string            `json:"token"`
	ExpiresAt        int64             `json:"expiresAt"`
	ExpireIn         string            `json:"expireIn"`
	Nickname         string            `json:"nickname"`
	OrgPermission    UserOrgPermission `json:"orgPermission"`    // User organization permissions
	Orgs             []IDName          `json:"orgs"`             // List of organizations the user belongs to
	Language         Language          `json:"language"`         // language
	IsUpdatePassword bool              `json:"isUpdatePassword"` // Has the password been updated?
}

type LoginByEmail struct {
	IsEmailCheck     bool   `json:"isEmailCheck"`
	Token            string `json:"token"`
	IsUpdatePassword bool   `json:"isUpdatePassword"` // Has the password been updated?
}

type Captcha struct {
	Key string `json:"key"` // client key
	B64 string `json:"b64"` // Verification code png picture base64 string
}

type LogoCustomInfo struct {
	Login         CustomLogin         `json:"login"`         // Login page title information
	Home          CustomHome          `json:"home"`          // Home page title information
	Tab           CustomTab           `json:"tab"`           // Tab information
	About         CustomAbout         `json:"about"`         // About information
	LinkList      map[string]string   `json:"linkList"`      // Jump link list, key is the link name, value is the URL
	Register      CustomRegister      `json:"register"`      // Registration information
	ResetPassword CustomResetPassword `json:"resetPassword"` // Reset password information
	LoginEmail    CustomLoginEmail    `json:"loginEmail"`    // Email login information
	DefaultIcon   CustomDefaultIcon   `json:"defaultIcon"`   // Apply default image
}

type CustomLogin struct {
	Background       request.Avatar `json:"background"`       // Login page background image
	Logo             request.Avatar `json:"logo"`             // Login page icon
	LoginButtonColor string         `json:"loginButtonColor"` // Login button color
	WelcomeText      string         `json:"welcomeText"`      // Login page welcome tag
	PlatformDesc     string         `json:"platformDesc"`     // platform descriptor
}

type CustomHome struct {
	Logo            request.Avatar `json:"logo"`            // Home logo
	Title           string         `json:"title"`           // Platform name
	BackgroundColor string         `json:"backgroundColor"` // Platform background color
}

type CustomTab struct {
	Logo  request.Avatar `json:"logo"`  // tab icon
	Title string         `json:"title"` // Tab title
}

type CustomAbout struct {
	LogoPath  string `json:"logoPath"` // About icon paths
	Version   string `json:"version"`
	Copyright string `json:"copyright"` // copyright
}

type CustomRegister struct {
	Email CustomEmail `json:"email"` // Register email
}

type CustomResetPassword struct {
	Email CustomEmail `json:"email"` // Mail
}

type CustomLoginEmail struct {
	Email CustomEmail `json:"email"` // Login email
}

type CustomDefaultIcon struct {
	RagIcon      string `json:"ragIcon"`
	AgentIcon    string `json:"agentIcon"`
	WorkflowIcon string `json:"workflowIcon"`
	PromptIcon   string `json:"promptIcon"`
	ChatflowIcon string `json:"chatflowIcon"`
}

type CustomEmail struct {
	Status bool `json:"status"`
}

type LanguageSelect struct {
	Languages       []Language `json:"languages"`
	DefaultLanguage Language   `json:"defaultLanguage"`
}

type Language struct {
	Code string `json:"code"` // language code
	Name string `json:"name"` // Language name
}
