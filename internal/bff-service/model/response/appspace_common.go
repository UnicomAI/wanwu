package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type AppBriefInfo struct {
	UniqueId    string         `json:"uniqueId"`    // 随机unique id(每次动态生成) [EN] Random unique id (dynamically generated each time)
	AppId       string         `json:"appId"`       // 应用id [EN] application id
	AppType     string         `json:"appType"`     // 应用类型 [EN] Application type
	Avatar      request.Avatar `json:"avatar"`      // 应用图标 [EN] application icon
	Name        string         `json:"name"`        // 应用名称 [EN] Application name
	Desc        string         `json:"desc"`        // 应用描述 [EN] Application description
	CreatedAt   string         `json:"createdAt"`   // 应用创建时间 [EN] Application creation time
	UpdatedAt   string         `json:"updatedAt"`   // 应用更新时间(用于历史记录排序) [EN] Apply update time (for history sorting)
	PublishType string         `json:"publishType"` // 发布类型(public:公开发布,private:私密发布) [EN] Release type (public: public release, private: private release)
}

type AppUrlInfo struct {
	UrlId               string `json:"urlId"`               // UrlID
	AppId               string `json:"appId"`               // 应用ID [EN] Application ID
	AppType             string `json:"appType"`             // 应用类型 [EN] Application type
	Name                string `json:"name"`                // Url名称 [EN] Url name
	CreatedAt           string `json:"createdAt"`           // 创建时间 [EN] creation time
	ExpiredAt           string `json:"expiredAt"`           // 过期时间 [EN] Expiration time
	Copyright           string `json:"copyright"`           // 知识产权 [EN] intellectual property
	CopyrightEnable     bool   `json:"copyrightEnable"`     // 知识产权开关 [EN] Intellectual property switch
	PrivacyPolicy       string `json:"privacyPolicy"`       // 隐私政策 [EN] privacy policy
	PrivacyPolicyEnable bool   `json:"privacyPolicyEnable"` // 隐私政策开关 [EN] privacy policy switch
	Disclaimer          string `json:"disclaimer"`          // 免责声明 [EN] Disclaimer
	DisclaimerEnable    bool   `json:"disclaimerEnable"`    // 免责声明开关 [EN] disclaimer switch
	Suffix              string `json:"suffix"`              // 生成Url后缀 [EN] Generate Url suffix
	Status              bool   `json:"status"`              // 应用Url开关 [EN] Apply Url switch
	UserId              string `json:"userId"`              // 用户ID [EN] User ID
	OrgId               string `json:"orgId"`               // 组织ID [EN] Organization ID
	Description         string `json:"description"`         // 应用描述 [EN] Application description
}

type AppUrlConfig struct {
	Assistant  *Assistant  `json:"assistant"`  // 基本信息 [EN] Basic information
	AppUrlInfo *AppUrlInfo `json:"appUrlInfo"` // 应用Url信息 [EN] Application Url information
}

type VisionConfig struct {
	MaxPicNum int32 `json:"maxPicNum"` // 最大图片数量 [EN] Maximum number of pictures
	PicNum    int32 `json:"picNum"`    // 视觉配置图片数量 [EN] Number of visual configuration pictures
}
