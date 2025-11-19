package request

type CustomTabConfig struct {
	TabLogo  Avatar `json:"tabLogo"`  // 标签页图标 [EN] tab icon
	TabTitle string `json:"tabTitle"` // 标签页标题 [EN] Tab title
	CommonCheck
}

type CustomLoginConfig struct {
	LoginBg          Avatar `json:"loginBg"`          // 登录页背景图 [EN] Login page background image
	LoginLogo        Avatar `json:"loginLogo"`        // 登录页图标 [EN] Login page icon
	LoginWelcomeText string `json:"loginWelcomeText"` // 登录页欢迎语 [EN] Login page welcome message
	LoginButtonColor string `json:"loginButtonColor"` // 登录按钮颜色 [EN] Login button color
	CommonCheck
}

type CustomHomeConfig struct {
	HomeLogo    Avatar `json:"homeLogo"`    // 平台图标 [EN] platform icon
	HomeName    string `json:"homeName"`    // 平台名称 [EN] Platform name
	HomeBgColor string `json:"homeBgColor"` // 平台背景颜色 [EN] Platform background color
	CommonCheck
}
