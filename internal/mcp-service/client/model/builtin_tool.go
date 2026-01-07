package model

import "github.com/UnicomAI/wanwu/pkg/db"

// BuiltinTool 自定义工具
type BuiltinTool struct {
	ID           uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id'"`
	ToolSquareId string      `gorm:"column:tool_square_id;index:idx_custom_tool_square_id;not null;comment:'自定义工具id'"`
	AuthJSON     db.LongText `gorm:"column:auth_json;comment:'鉴权json'"`
	UserID       string      `gorm:"column:user_id;index:idx_user_id_name,priority:1;type:varchar(64);not null;comment:'用户id'"`
	OrgID        string      `gorm:"column:org_id;type:varchar(64);not null;comment:'组织id'"`
	CreatedAt    int64       `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt    int64       `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
