package model

type AssistantSnapshotExtra struct {
	HideKnowledge int `json:"hideKnowledge"`
}

type AssistantSnapshot struct {
	ID                      uint32 `gorm:"primarykey;column:id;comment:智能体快照Id"`
	AssistantID             uint32 `gorm:"column:assistant_id;type:int;index:idx_assistant_snapshot_assistant_id;comment:智能体Id"`
	Version                 string `gorm:"column:version;type:varchar(64);comment:智能体版本"`
	Category                int    `gorm:"column:category;type:tinyint(4);not null;default:1;comment:'1-单智能体，2-多智能体';" json:"category"`
	SnapshotDesc            string `gorm:"column:desc;type:longtext;comment:智能体介绍"`
	SnapshotExtra           string `gorm:"column:extra;type:longtext;comment:智能体扩展信息"`
	AssistantInfo           string `gorm:"column:assistant_info;type:longtext;comment:智能体基本信息"`
	AssistantToolConfig     string `gorm:"column:assistant_tool_config;type:longtext;comment:智能体工具配置"`
	AssistantMCPConfig      string `gorm:"column:assistant_mcp_config;type:longtext;comment:智能体MCP配置"`
	AssistantWorkflowConfig string `gorm:"column:assistant_workflow_config;type:longtext;comment:智能体工作流配置"`
	AssistantSkillConfig    string `gorm:"column:assistant_skill_config;type:longtext;comment:智能体技能配置"`
	UserId                  string `gorm:"column:user_id;index:idx_assistant_user_id;comment:用户id"`
	OrgId                   string `gorm:"column:org_id;index:idx_assistant_org_id;comment:组织id"`
	CreatedAt               int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt               int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
