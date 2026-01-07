package model

import "github.com/UnicomAI/wanwu/pkg/db"

type Assistant struct {
	ID                  uint32      `gorm:"primarykey;column:id;comment:智能体Id"`
	UUID                string      `gorm:"column:uuid;type:varchar(255);uniqueIndex:idx_unique_uuid;comment:智能体uuid"`
	AvatarPath          string      `gorm:"column:avatar_path;comment:智能体头像"`
	Name                string      `gorm:"column:name;comment:智能体名称"`
	Desc                string      `gorm:"column:desc;comment:智能体介绍"`
	Instructions        string      `gorm:"column:instructions;comment:系统提示词"`
	Prologue            string      `gorm:"column:prologue;comment:开场白"`
	RecommendQuestion   string      `gorm:"column:recommend_question;comment:推荐问题列表"`
	ModelConfig         db.LongText `gorm:"column:model_config;comment:模型配置"`
	RerankConfig        db.LongText `gorm:"column:rerank_config;comment:rerank模型配置"`
	KnowledgebaseConfig db.LongText `gorm:"column:knowledgebase_config;comment:知识库配置"`
	SafetyConfig        db.LongText `gorm:"column:safety_config;comment:安全配置"`
	VisionConfig        db.LongText `gorm:"column:vision_config;comment:视觉配置"`
	Scope               int         `gorm:"column:scope;comment:智能体可见范围"`
	UserId              string      `gorm:"column:user_id;index:idx_assistant_user_id;comment:用户id"`
	OrgId               string      `gorm:"column:org_id;index:idx_assistant_org_id;comment:组织id"`
	CreatedAt           int64       `gorm:"autoCreateTime:milli;comment:创建时间"`
	UpdatedAt           int64       `gorm:"autoUpdateTime:milli;comment:更新时间"`
}
