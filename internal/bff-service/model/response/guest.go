package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// RSAPublicKey RSA公钥响应
type RSAPublicKey struct {
	KeyID     string `json:"keyId"`     // 密钥ID
	PublicKey string `json:"publicKey"` // PEM格式的公钥
	Challenge string `json:"challenge"` // 服务端下发的Challenge，前端需嵌入cipher加密载荷中与密码一同加密提交
	ExpiresIn int64  `json:"expiresIn"` // 建议缓存时间(秒)
}

type Login struct {
	UID              string             `json:"uid"`
	Username         string             `json:"username"`
	Token            string             `json:"token"`
	ExpiresAt        int64              `json:"expiresAt"`
	ExpireIn         string             `json:"expireIn"`
	Nickname         string             `json:"nickname"`
	OrgPermission    UserOrgPermission  `json:"orgPermission"`    // 用户所在组织权限
	Orgs             []IDNameWithAvatar `json:"orgs"`             // 用户所在组织列表
	Language         Language           `json:"language"`         // 语言
	IsUpdatePassword bool               `json:"isUpdatePassword"` // 是否已更新密码
}

type LoginByEmail struct {
	IsEmailCheck     bool   `json:"isEmailCheck"`
	Token            string `json:"token"`
	IsUpdatePassword bool   `json:"isUpdatePassword"` // 是否已更新密码
}

type Captcha struct {
	Key string `json:"key"` // 客户端key
	B64 string `json:"b64"` // 验证码png图片base64字符串
}

type LogoCustomInfo struct {
	Login              CustomLogin         `json:"login"`              // 登录页标题信息
	Home               CustomHome          `json:"home"`               // 首页标题信息
	Tab                CustomTab           `json:"tab"`                // 标签页信息
	About              CustomAbout         `json:"about"`              // 关于信息
	LinkList           map[string]string   `json:"linkList"`           // 跳转链接列表,key为链接名称,value为URL
	Register           CustomRegister      `json:"register"`           // 注册信息
	ResetPassword      CustomResetPassword `json:"resetPassword"`      // 重置密码信息
	LoginEmail         CustomLoginEmail    `json:"loginEmail"`         // 邮箱登录信息
	DefaultIcon        CustomDefaultIcon   `json:"defaultIcon"`        // 应用默认图片
	UserPhoneRequired  bool                `json:"userPhoneRequired"`  // 是否需要手机号
	GeneralAgent       CustomGeneralAgent  `json:"generalAgent"`       // 通用智能体配置
	PasswordRSAEncrypt bool                `json:"passwordRSAEncrypt"` // 是否启用密码RSA加密
}

type CustomLogin struct {
	Background       request.Avatar `json:"background"`       // 登录页背景图
	Logo             request.Avatar `json:"logo"`             // 登录页图标
	LoginButtonColor string         `json:"loginButtonColor"` // 登录按钮颜色
	WelcomeText      string         `json:"welcomeText"`      // 登录页欢迎标词
	PlatformDesc     string         `json:"platformDesc"`     // 平台描述词
}

type CustomHome struct {
	Logo            request.Avatar `json:"logo"`            // 首页logo
	Title           string         `json:"title"`           // 平台名称
	BackgroundColor string         `json:"backgroundColor"` // 平台背景色
}

type CustomTab struct {
	Logo  request.Avatar `json:"logo"`  // 标签页图标
	Title string         `json:"title"` // 标签页标题
}

type CustomAbout struct {
	LogoPath  string `json:"logoPath"` // 关于图标路径
	Version   string `json:"version"`
	Copyright string `json:"copyright"` // 版权
}

type CustomRegister struct {
	Email CustomEmail `json:"email"` // 注册邮箱
}

type CustomResetPassword struct {
	Email CustomEmail `json:"email"` // 邮箱
}

type CustomLoginEmail struct {
	Email CustomEmail `json:"email"` // 登录邮箱
}

type CustomDefaultIcon struct {
	RagIcon      string `json:"ragIcon"`
	AgentIcon    string `json:"agentIcon"`
	WorkflowIcon string `json:"workflowIcon"`
	PromptIcon   string `json:"promptIcon"`
	ChatflowIcon string `json:"chatflowIcon"`
	ModelIcon    string `json:"modelIcon"`
}

type CustomGeneralAgent struct {
	Logo        request.Avatar `json:"logo"`        // 通用智能体图标
	WelcomeText string         `json:"welcomeText"` // 通用智能体欢迎语
	MenuName    string         `json:"menuName"`    // 通用智能体菜单名称
}

type CustomEmail struct {
	Status bool `json:"status"`
}

type LanguageSelect struct {
	Languages       []Language `json:"languages"`
	DefaultLanguage Language   `json:"defaultLanguage"`
}

type Language struct {
	Code string `json:"code"` // 语言代码
	Name string `json:"name"` // 语言名称
}

type ReleaseNotesResp struct {
	Version      string `json:"version"`
	ReleaseNotes string `json:"releaseNotes"`
}
