package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type Login struct {
	UID              string            `json:"uid"`
	Username         string            `json:"username"`
	Token            string            `json:"token"`
	ExpiresAt        int64             `json:"expiresAt"`
	ExpireIn         string            `json:"expireIn"`
	Nickname         string            `json:"nickname"`
	OrgPermission    UserOrgPermission `json:"orgPermission"`    // 用户所在组织权限 [EN] User organization permissions
	Orgs             []IDName          `json:"orgs"`             // 用户所在组织列表 [EN] List of organizations the user belongs to
	Language         Language          `json:"language"`         // 语言 [EN] language
	IsUpdatePassword bool              `json:"isUpdatePassword"` // 是否已更新密码 [EN] Has the password been updated?
}

type LoginByEmail struct {
	IsEmailCheck     bool   `json:"isEmailCheck"`
	Token            string `json:"token"`
	IsUpdatePassword bool   `json:"isUpdatePassword"` // 是否已更新密码 [EN] Has the password been updated?
}

type Captcha struct {
	Key string `json:"key"` // 客户端key [EN] client key
	B64 string `json:"b64"` // 验证码png图片base64字符串 [EN] Verification code png picture base64 string
}

type LogoCustomInfo struct {
	Login         CustomLogin         `json:"login"`         // 登录页标题信息 [EN] Login page title information
	Home          CustomHome          `json:"home"`          // 首页标题信息 [EN] Home page title information
	Tab           CustomTab           `json:"tab"`           // 标签页信息 [EN] Tab information
	About         CustomAbout         `json:"about"`         // 关于信息 [EN] About information
	LinkList      map[string]string   `json:"linkList"`      // 跳转链接列表,key为链接名称,value为URL [EN] Jump link list, key is the link name, value is the URL
	Register      CustomRegister      `json:"register"`      // 注册信息 [EN] Registration information
	ResetPassword CustomResetPassword `json:"resetPassword"` // 重置密码信息 [EN] Reset password information
	LoginEmail    CustomLoginEmail    `json:"loginEmail"`    // 邮箱登录信息 [EN] Email login information
	DefaultIcon   CustomDefaultIcon   `json:"defaultIcon"`   // 应用默认图片 [EN] Apply default image
}

type CustomLogin struct {
	Background       request.Avatar `json:"background"`       // 登录页背景图 [EN] Login page background image
	Logo             request.Avatar `json:"logo"`             // 登录页图标 [EN] Login page icon
	LoginButtonColor string         `json:"loginButtonColor"` // 登录按钮颜色 [EN] Login button color
	WelcomeText      string         `json:"welcomeText"`      // 登录页欢迎标词 [EN] Login page welcome tag
	PlatformDesc     string         `json:"platformDesc"`     // 平台描述词 [EN] platform descriptor
}

type CustomHome struct {
	Logo            request.Avatar `json:"logo"`            // 首页logo [EN] Home logo
	Title           string         `json:"title"`           // 平台名称 [EN] Platform name
	BackgroundColor string         `json:"backgroundColor"` // 平台背景色 [EN] Platform background color
}

type CustomTab struct {
	Logo  request.Avatar `json:"logo"`  // 标签页图标 [EN] tab icon
	Title string         `json:"title"` // 标签页标题 [EN] Tab title
}

type CustomAbout struct {
	LogoPath  string `json:"logoPath"` // 关于图标路径 [EN] About icon paths
	Version   string `json:"version"`
	Copyright string `json:"copyright"` // 版权 [EN] copyright
}

type CustomRegister struct {
	Email CustomEmail `json:"email"` // 注册邮箱 [EN] Register email
}

type CustomResetPassword struct {
	Email CustomEmail `json:"email"` // 邮箱 [EN] Mail
}

type CustomLoginEmail struct {
	Email CustomEmail `json:"email"` // 登录邮箱 [EN] Login email
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
	Code string `json:"code"` // 语言代码 [EN] language code
	Name string `json:"name"` // 语言名称 [EN] Language name
}
