package request

type CustomTabConfig struct {
	TabLogo  Avatar `json:"tabLogo"`  // tab icon
	TabTitle string `json:"tabTitle"` // Tab title
	CommonCheck
}

type CustomLoginConfig struct {
	LoginBg          Avatar `json:"loginBg"`          // Login page background image
	LoginLogo        Avatar `json:"loginLogo"`        // Login page icon
	LoginWelcomeText string `json:"loginWelcomeText"` // Login page welcome message
	LoginButtonColor string `json:"loginButtonColor"` // Login button color
	CommonCheck
}

type CustomHomeConfig struct {
	HomeLogo    Avatar `json:"homeLogo"`    // platform icon
	HomeName    string `json:"homeName"`    // Platform name
	HomeBgColor string `json:"homeBgColor"` // Platform background color
	CommonCheck
}
