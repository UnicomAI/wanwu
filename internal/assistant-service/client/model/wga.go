package model

type WgaConfig struct {
	ID                    uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:配置Id"`
	ToolList              string `gorm:"column:tool_list;type:json;comment:工具列表JSON"`
	AssistantList         string `gorm:"column:assistant_list;type:json;comment:智能体列表JSON"`
	McpList               string `gorm:"column:mcp_list;type:json;comment:MCP列表JSON"`
	WorkflowList          string `gorm:"column:workflow_list;type:json;comment:工作流列表JSON"`
	SkillList             string `gorm:"column:skill_list;type:json;comment:技能列表JSON"`
	KnowledgeList         string `gorm:"column:knowledge_list;type:json;comment:知识库列表JSON"`
	OntologyKnowledgeList string `gorm:"column:ontology_knowledge_list;type:json;comment:本体知识网络列表JSON"`
	UserID                string `gorm:"column:user_id;index:idx_wga_config_user_id;type:varchar(64);not null;comment:用户id"`
	OrgID                 string `gorm:"column:org_id;index:idx_wga_config_org_id;type:varchar(64);not null;comment:组织id"`
	CreatedAt             int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt             int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}

type WgaConversationConfig struct {
	ID          uint32 `gorm:"column:id;primary_key;type:bigint(20) auto_increment;not null;comment:配置Id"`
	ThreadID    string `gorm:"column:thread_id;uniqueIndex:idx_wga_conversation_config_thread_id;type:varchar(128);not null;comment:对话ID"`
	Title       string `gorm:"column:title;type:text;comment:对话标题"`
	ModelConfig string `gorm:"column:model_config;type:json;comment:模型配置JSON"`
	UserID      string `gorm:"column:user_id;index:idx_wga_conversation_config_user_id;type:varchar(64);comment:用户id"`
	OrgID       string `gorm:"column:org_id;index:idx_wga_conversation_config_org_id;type:varchar(64);comment:组织id"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
