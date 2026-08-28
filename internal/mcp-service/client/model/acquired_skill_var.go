package model

type AcquiredSkillVariable struct {
	ID              uint32 `gorm:"primarykey"`
	AcquiredSkillID string `gorm:"column:acquired_skill_id;uniqueIndex:ux_acquired_skill_var_skill_user_org_name;type:varchar(128);comment:acquired skill id"`
	UserID          string `gorm:"column:user_id;uniqueIndex:ux_acquired_skill_var_skill_user_org_name;type:varchar(64);comment:用户id"`
	OrgID           string `gorm:"column:org_id;uniqueIndex:ux_acquired_skill_var_skill_user_org_name;type:varchar(64);comment:组织id"`
	Name            string `gorm:"column:name;uniqueIndex:ux_acquired_skill_var_skill_user_org_name;type:varchar(255);comment:工具名称"`
	Desc            string `gorm:"column:desc;type:text;comment:描述"`
	VariableKey     string `gorm:"column:variable_key;comment:变量Key"`
	VariableValue   string `gorm:"column:variable_value;comment:变量Value"`
	CreatedAt       int64  `gorm:"column:created_at;autoCreateTime:milli;comment:创建时间"`
	UpdatedAt       int64  `gorm:"column:updated_at;autoUpdateTime:milli;comment:更新时间"`
}
