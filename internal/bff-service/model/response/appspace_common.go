package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type AppBriefInfo struct {
	UniqueId    string         `json:"uniqueId"`    // Random unique id (dynamically generated each time)
	AppId       string         `json:"appId"`       // application id
	AppType     string         `json:"appType"`     // Application type
	Avatar      request.Avatar `json:"avatar"`      // application icon
	Name        string         `json:"name"`        // Application name
	Desc        string         `json:"desc"`        // Application description
	CreatedAt   string         `json:"createdAt"`   // Application creation time
	UpdatedAt   string         `json:"updatedAt"`   // Apply update time (for history sorting)
	PublishType string         `json:"publishType"` // Release type (public: public release, private: private release)
}

type AppUrlInfo struct {
	UrlId               string `json:"urlId"`               // UrlID
	AppId               string `json:"appId"`               // Application ID
	AppType             string `json:"appType"`             // Application type
	Name                string `json:"name"`                // Url name
	CreatedAt           string `json:"createdAt"`           // creation time
	ExpiredAt           string `json:"expiredAt"`           // Expiration time
	Copyright           string `json:"copyright"`           // intellectual property
	CopyrightEnable     bool   `json:"copyrightEnable"`     // Intellectual property switch
	PrivacyPolicy       string `json:"privacyPolicy"`       // privacy policy
	PrivacyPolicyEnable bool   `json:"privacyPolicyEnable"` // privacy policy switch
	Disclaimer          string `json:"disclaimer"`          // Disclaimer
	DisclaimerEnable    bool   `json:"disclaimerEnable"`    // disclaimer switch
	Suffix              string `json:"suffix"`              // Generate Url suffix
	Status              bool   `json:"status"`              // Apply Url switch
	UserId              string `json:"userId"`              // User ID
	OrgId               string `json:"orgId"`               // Organization ID
	Description         string `json:"description"`         // Application description
}

type AppUrlConfig struct {
	Assistant  *Assistant  `json:"assistant"`  // Basic information
	AppUrlInfo *AppUrlInfo `json:"appUrlInfo"` // Application Url information
}

type VisionConfig struct {
	MaxPicNum int32 `json:"maxPicNum"` // Maximum number of pictures
	PicNum    int32 `json:"picNum"`    // Number of visual configuration pictures
}
