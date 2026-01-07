package model

type AppUrl struct {
	ID                  uint32 `gorm:"primarykey;column:id;comment:应用UrlId"`
	AppID               string `gorm:"column:app_id;comment:关联的应用Id;index:idx_app_url_app_id"`
	AppType             string `gorm:"column:app_type;comment:应用类型;index:idx_app_url_app_type"`
	Name                string `gorm:"column:name;comment:配置名称;index:idx_app_url_name"`
	CreatedAt           int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	ExpiredAt           int64  `gorm:"column:expired_at;comment:配置结束时间戳"`
	Copyright           string `gorm:"column:copyright;type:text;comment:版权声明内容"`
	CopyrightEnable     bool   `gorm:"column:copyright_enable;comment:是否启用版权声明"`
	PrivacyPolicy       string `gorm:"column:privacy_policy;type:text;comment:隐私政策内容"`
	PrivacyPolicyEnable bool   `gorm:"column:privacy_policy_enable;comment:是否启用隐私政策"`
	Disclaimer          string `gorm:"column:disclaimer;type:text;comment:免责声明内容"`
	DisclaimerEnable    bool   `gorm:"column:disclaimer_enable;comment:是否启用免责声明"`
	Suffix              string `gorm:"column:suffix;type:varchar(255);comment:应用Url;index:idx_app_url_suffix"`
	UserId              string `gorm:"column:user_id;index:idx_assistant_url_user_id;comment:用户Id;index:idx_app_url_user_id"`
	OrgId               string `gorm:"column:org_id;index:idx_assistant_url_org_id;comment:组织Id;index:idx_app_url_org_id"`
	Status              bool   `gorm:"column:status;default:true;comment:应用Url开关;index:idx_app_url_status"`
	Description         string `gorm:"column:description;type:text;comment:app描述"`
}
