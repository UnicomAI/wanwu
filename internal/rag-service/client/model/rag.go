package model

import "github.com/UnicomAI/wanwu/pkg/db"

const (
	MatchTypeDefault         = "mix"
	KnowledgePriorityDefault = 1
	QAPriorityDefault        = 0
	KeywordPriorityDefault   = 0.8
	ThresholdDefault         = 0.4
	SemanticsPriorityDefault = 0.2
	TopKDefault              = 5
)

type RagInfo struct {
	ID    int64  `json:"id" gorm:"primaryKey;type:bigint auto_increment;not null;"`
	RagID string `json:"ragId" gorm:"uniqueIndex:idx_unique_rag_id;column:rag_id;type:varchar(255);comment:ragId"`

	// 使用嵌入结构体（将字段直接映射到主表）
	BriefConfig           AppBriefConfig      `gorm:"embedded;embeddedPrefix:brief_"`
	ModelConfig           AppModelConfig      `gorm:"embedded;embeddedPrefix:model_"`
	RerankConfig          AppModelConfig      `gorm:"embedded;embeddedPrefix:rerank_"`
	QARerankConfig        AppModelConfig      `gorm:"embedded;embeddedPrefix:qa_rerank_"`
	KnowledgeBaseConfig   KnowledgeBaseConfig `gorm:"embedded;embeddedPrefix:kb_"`
	QAKnowledgebaseConfig db.LongText         `gorm:"column:qa_knowledgebase_config;comment:问答库配置"`
	SensitiveConfig       SensitiveConfig     `gorm:"embedded;embeddedPrefix:sensitive_"`
	PublicModel
}

type AppBriefConfig struct {
	Name       string `json:"name" gorm:"column:name;type:varchar(255);comment:应用名称"`
	Desc       string `json:"desc" gorm:"column:desc;type:varchar(255);comment:应用描述"`
	AvatarPath string `json:"avatarPath" gorm:"column:avatar_path;type:varchar(255);comment:应用图标"`
}

type AppModelConfig struct {
	Provider  string `json:"provider" gorm:"column:provider;type:varchar(255);comment:模型供应商"`
	Model     string `json:"model" gorm:"column:model;type:varchar(255);comment:模型名称"`
	ModelId   string `json:"modelId" gorm:"column:model_id;type:varchar(255);comment:模型ID"`
	ModelType string `json:"modelType" gorm:"column:model_type;type:varchar(255);comment:模型类型"`
	Config    string `json:"config" gorm:"column:config;type:varchar(255);comment:模型配置"`
}

type KnowledgeBaseConfig struct {
	KnowId            string  `json:"knowId" gorm:"column:know_id;type:text;comment:知识库ID"`
	MaxHistory        int64   `json:"maxHistory" gorm:"column:max_history;type:bigint;comment:最大历史记录"`
	Threshold         float64 `json:"threshold" gorm:"column:threshold;type:decimal(10,2);comment:阈值"`
	TopK              int64   `json:"topK" gorm:"column:top_k;type:bigint;comment:TopK"`
	MatchType         string  `json:"matchType" gorm:"column:match_type;type:varchar(32);not null;default:'';comment:matchType：vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）"`
	PriorityMatch     int32   `json:"priorityMatch" gorm:"column:priority_match;not null;default:0;comment:权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1"`
	SemanticsPriority float64 `json:"semanticsPriority" gorm:"column:semantics_priority;type:decimal(10,2);not null;default:0;comment:语义权重"`
	KeywordPriority   float64 `json:"keywordPriority" gorm:"column:keyword_priority;type:decimal(10,2);not null;default:0;comment:关键词权重"`
	TermWeight        float64 `json:"term_weight" gorm:"column:term_weight;type:decimal(10,2);not null;default:1;comment:关键词系数,默认为1"`
	TermWeightEnable  bool    `json:"term_weight_enable" gorm:"column:term_weight_enable;not null;default:false;comment:是否启用关键词系数"`
	MetaParams        string  `json:"metaParams" gorm:"column:meta_params;type:text;comment:元数据参数"`
	UseGraph          bool    `json:"use_graph" gorm:"column:use_graph;not null;default:false;comment:是否使用知识图谱"`
}

type SensitiveConfig struct {
	Enable   bool   `json:"enable" gorm:"column:enable;comment:是否启用安全护栏"`
	TableIds string `json:"tableIds" gorm:"column:table_ids;type:text;comment:敏感词表ID列表"`
}

type PublicModel struct {
	CreatedAt int64  `json:"createdAt" gorm:"autoCreateTime:milli;index:created_at;column:created_at;type:bigint;comment:创建时间"`
	UpdatedAt int64  `json:"updatedAt" gorm:"autoUpdateTime:milli;index:updated_at;column:updated_at;type:bigint;comment:更新时间"`
	OrgID     string `gorm:"index:org_id;column:org_id;type:varchar(255);comment:组织ID" json:"orgId"`
	UserID    string `gorm:"index:user_id;column:user_id;type:varchar(255);comment:用户ID" json:"userId"`
}

func (r RagInfo) TableName() string {
	return "rag_info"
}
