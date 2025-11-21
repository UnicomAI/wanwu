package model

type AppUrl struct {
	ID                  uint32 `gorm:"primarykey;column:id;comment:Application URL ID"`
	AppID               string `gorm:"column:app_id;comment:Associated application ID;index:idx_app_url_app_id"`
	AppType             string `gorm:"column:app_type;comment:Application type;index:idx_app_url_app_type"`
	Name                string `gorm:"column:name;comment:Configuration name;index:idx_app_url_name"`
	CreatedAt           int64  `gorm:"autoCreateTime:milli;comment:Creation time"`
	ExpiredAt           int64  `gorm:"column:expired_at;comment:Configuration expiration timestamp"`
	Copyright           string `gorm:"column:copyright;type:text;comment:Copyright notice content"`
	CopyrightEnable     bool   `gorm:"column:copyright_enable;type:tinyint;comment:Enable copyright notice"`
	PrivacyPolicy       string `gorm:"column:privacy_policy;type:text;comment:Privacy policy content"`
	PrivacyPolicyEnable bool   `gorm:"column:privacy_policy_enable;type:tinyint;comment:Enable privacy policy"`
	Disclaimer          string `gorm:"column:disclaimer;type:text;comment:Disclaimer content"`
	DisclaimerEnable    bool   `gorm:"column:disclaimer_enable;type:tinyint;comment:Enable disclaimer"`
	Suffix              string `gorm:"column:suffix;type:varchar(255);comment:Application URL;index:idx_app_url_suffix"`
	UserId              string `gorm:"column:user_id;index:idx_assistant_url_user_id;comment:User ID;index:idx_app_url_user_id"`
	OrgId               string `gorm:"column:org_id;index:idx_assistant_url_org_id;comment:Organization ID;index:idx_app_url_org_id"`
	Status              bool   `gorm:"column:status;type:tinyint;default:true;comment:Application URL switch;index:idx_app_url_status"`
	Description         string `gorm:"column:description;type:text;comment:App description"`
}
