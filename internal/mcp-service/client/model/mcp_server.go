package model

import "github.com/UnicomAI/wanwu/pkg/db"

type MCPServer struct {
	ID          uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id'"`
	MCPServerID string      `gorm:"uniqueIndex:idx_unique_mcp_server_id;column:mcp_server_id;type:varchar(255);not null;comment:'mcp server id'"`
	Name        string      `gorm:"column:name;index:idx_user_id_name,priority:2;type:varchar(255);comment:'mcp server名称'"`
	Description db.LongText `gorm:"column:description;comment:'mcp server描述'"`
	AvatarPath  string      `gorm:"column:avatar_path;comment:mcp server头像"`
	UserID      string      `gorm:"column:user_id;index:idx_user_id_name,priority:1;type:varchar(64);not null;comment:'用户id'"`
	OrgID       string      `gorm:"column:org_id;type:varchar(64);not null;comment:'组织id'"`
	CreatedAt   int64       `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64       `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
