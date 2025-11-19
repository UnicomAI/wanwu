package request

type AppUrlIdRequest struct {
	UrlId string `json:"urlId" form:"urlId" validate:"required"` // UrlID
}

func (a *AppUrlIdRequest) Check() error { return nil }

type AppUrlConfig struct {
	Name                string `json:"name" validate:"required"` // name
	ExpiredAt           string `json:"expiredAt"`                // Expiration time
	Copyright           string `json:"copyright"`                // copyright
	CopyrightEnable     bool   `json:"copyrightEnable"`          // Copyright switch
	PrivacyPolicy       string `json:"privacyPolicy"`            // privacy agreement
	PrivacyPolicyEnable bool   `json:"privacyPolicyEnable"`      // Privacy protocol switch
	Disclaimer          string `json:"disclaimer"`               // Disclaimer
	DisclaimerEnable    bool   `json:"disclaimerEnable"`         // disclaimer switch
	Description         string `json:"description"`
}

func (cfg AppUrlConfig) Check() error {
	return nil
}

type AppUrlCreateRequest struct {
	AppId   string `json:"appId" validate:"required"`   // application id
	AppType string `json:"appType" validate:"required"` // Application type
	AppUrlConfig
}

type AppUrlUpdateRequest struct {
	UrlId string `json:"urlId" validate:"required"` // UrlID
	AppUrlConfig
}

type AppUrlListRequest struct {
	AppId   string `json:"appId" form:"appId" validate:"required"`     // application id
	AppType string `json:"appType" form:"appType" validate:"required"` // Application type
}

func (a *AppUrlListRequest) Check() error { return nil }

type AppUrlStatusRequest struct {
	UrlId  string `json:"urlId" validate:"required"` // UrlID
	Status bool   `json:"status"`                    // Start and stop status
}

func (a *AppUrlStatusRequest) Check() error { return nil }
