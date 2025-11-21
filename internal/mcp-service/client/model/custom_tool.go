package model

const (
	ApiAuthNone = "none"
)

// CustomToolCustom tools
type CustomTool struct {
	ID               uint32 `gorm:"primary_key"`
	ToolSquareId     string `gorm:"index:idx_custom_tool_square_id"`
	Name             string `gorm:"column:name;type:varchar(255);comment:'自定义工具名称'"`
	AvatarPath       string `gorm:"column:avatar_path;comment:'自定义工具头像'"`
	Description      string `gorm:"column:description;type:longtext;comment:'自定义工具描述'"`
	Schema           string `gorm:"column:schema;type:longtext;comment:'schema配置'"`
	PrivacyPolicy    string `gorm:"column:privacy_policy;type:longtext;comment:'隐私政策'"`
	Type             string `gorm:"column:type;type:varchar(255);comment:'apiAuth认证类型(none/apiKey)'"` // DEPRECATED
	APIKey           string `gorm:"column:api_key;type:varchar(255);comment:'api_key，0.2.6作为内置工具专属'"`
	AuthType         string `gorm:"column:auth_type;type:varchar(255);comment:'authType(basic/bearer/custom)'"` // DEPRECATED
	CustomHeaderName string `gorm:"column:custom_header_name;type:varchar(255);comment:'自定义header名称'"`          // DEPRECATED
	AuthJSON         string `gorm:"column:auth_json;type:longtext;comment:'鉴权json'"`
	UserID           string `gorm:"column:user_id;index:idx_custom_tool_user_id;comment:'用户id'"`
	OrgID            string `gorm:"column:org_id;index:idx_custom_tool_org_id;comment:'组织id'"`
	CreatedAt        int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt        int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
