package model

import "github.com/UnicomAI/wanwu/pkg/db"

type MCPServerTool struct {
	ID              uint32      `gorm:"column:id;primary_key;type:bigint auto_increment;not null;comment:'id'"`
	MCPServerToolId string      `gorm:"uniqueIndex:idx_unique_mcp_server_tool_id;column:mcp_server_tool_id;type:varchar(255);not null;comment:'mcp server tool id'"`
	McpServerId     string      `gorm:"uniqueIndex:idx_unique_mcp_server_id_name,priority:1;column:mcp_server_id;index:idx_mcp_server_id;type:varchar(255);comment:'mcp server id'"`
	AppToolId       string      `gorm:"column:app_tool_id;type:varchar(255);comment:'应用或工具id'"`
	Type            string      `gorm:"column:type;type:varchar(255);comment:'mcp server tool类型'"`
	AppToolName     string      `gorm:"column:app_tool_name;type:varchar(255);comment:'应用或工具名称'"`
	Name            string      `gorm:"uniqueIndex:idx_unique_mcp_server_id_name,priority:2;column:name;type:varchar(255);comment:'mcp server tool名称'"`
	Description     db.LongText `gorm:"column:description;comment:'mcp server tool描述'"`
	Schema          db.LongText `gorm:"column:schema;comment:'openapi schema'"`
	AuthType        string      `gorm:"column:auth_type;type:varchar(255);comment:'鉴权类型'"`
	AuthIn          string      `gorm:"column:auth_in;type:varchar(255);comment:'鉴权位置'"`
	AuthName        string      `gorm:"column:auth_name;type:varchar(255);comment:'鉴权名称'"`
	AuthValue       string      `gorm:"column:auth_value;type:varchar(255);comment:'鉴权值'"`
	UserID          string      `gorm:"column:user_id;type:varchar(64);not null;comment:'用户id'"`
	OrgID           string      `gorm:"column:org_id;type:varchar(64);not null;comment:'组织id'"`
	CreatedAt       int64       `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt       int64       `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
