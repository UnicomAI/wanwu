package request

type AppUrlIdRequest struct {
	UrlId string `json:"urlId" form:"urlId" validate:"required"` // UrlID
}

func (a *AppUrlIdRequest) Check() error { return nil }

type AppUrlConfig struct {
	Name                string `json:"name" validate:"required"` // 名称 [EN] name
	ExpiredAt           string `json:"expiredAt"`                // 过期时间 [EN] Expiration time
	Copyright           string `json:"copyright"`                // 版权 [EN] copyright
	CopyrightEnable     bool   `json:"copyrightEnable"`          // 版权开关 [EN] Copyright switch
	PrivacyPolicy       string `json:"privacyPolicy"`            // 隐私协议 [EN] privacy agreement
	PrivacyPolicyEnable bool   `json:"privacyPolicyEnable"`      // 隐私协议开关 [EN] Privacy protocol switch
	Disclaimer          string `json:"disclaimer"`               // 免责声明 [EN] Disclaimer
	DisclaimerEnable    bool   `json:"disclaimerEnable"`         // 免责声明开关 [EN] disclaimer switch
	Description         string `json:"description"`
}

func (cfg AppUrlConfig) Check() error {
	return nil
}

type AppUrlCreateRequest struct {
	AppId   string `json:"appId" validate:"required"`   // 应用id [EN] application id
	AppType string `json:"appType" validate:"required"` // 应用类型 [EN] Application type
	AppUrlConfig
}

type AppUrlUpdateRequest struct {
	UrlId string `json:"urlId" validate:"required"` // UrlID
	AppUrlConfig
}

type AppUrlListRequest struct {
	AppId   string `json:"appId" form:"appId" validate:"required"`     // 应用id [EN] application id
	AppType string `json:"appType" form:"appType" validate:"required"` // 应用类型 [EN] Application type
}

func (a *AppUrlListRequest) Check() error { return nil }

type AppUrlStatusRequest struct {
	UrlId  string `json:"urlId" validate:"required"` // UrlID
	Status bool   `json:"status"`                    // 启停状态 [EN] Start and stop status
}

func (a *AppUrlStatusRequest) Check() error { return nil }
