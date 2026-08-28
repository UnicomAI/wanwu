package request

import (
	"encoding/json"
	"fmt"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
)

type Avatar struct {
	Key  string `json:"key"`  // 前端透传给后端用于保存avatar，例如：custom-upload/avatar/abc/def.png
	Path string `json:"path"` // 前端请求地址，例如：/v1/static/avatar/abc/def.png (请求非必填)
}

type AppBriefConfig struct {
	Avatar Avatar `json:"avatar"`                   // 图标
	Name   string `json:"name" validate:"required"` // 名称
	Desc   string `json:"desc"`                     // 描述
}

func (a AppBriefConfig) Check() error {
	return nil
}

type AppModelConfig struct {
	Provider    string      `json:"provider"`    // 模型供应商
	Model       string      `json:"model"`       // 模型名称
	ModelId     string      `json:"modelId"`     // 模型ID
	ModelType   string      `json:"modelType"`   // 模型类型(llm/embedding/rerank)
	DisplayName string      `json:"displayName"` // 模型展示名称(请求非必填)
	Config      interface{} `json:"config"`      // 模型配置

	Examples *mp.AppModelParams // 仅用于swagger展示；模型对应供应商中的对应llm、embedding或rerank结构是config实际的参数
}

func (cfg *AppModelConfig) Check() error {
	_, err := cfg.ConfigString()
	return err
}

func (cfg *AppModelConfig) ConfigString() (string, error) {
	if cfg.Config == nil {
		return "", nil
	}
	b, err := json.Marshal(cfg.Config)
	if err != nil {
		return "", fmt.Errorf("marshal app model config err: %v", err)
	}
	modelParams, _, err := mp.ToModelParams(cfg.Provider, cfg.ModelType, string(b))
	if err != nil {
		return "", err
	}
	b, err = json.Marshal(modelParams)
	if err != nil {
		return "", fmt.Errorf("marshal app model config err: %v", err)
	}
	return string(b), nil
}

type AppKnowledgebaseConfig struct {
	Knowledgebases []AppKnowledgeBase     `json:"knowledgebases"` // 知识库id、名字
	Config         AppKnowledgebaseParams `json:"config"`         // 知识库参数
}

type AppKnowledgeBase struct {
	ID                   string                `json:"id" validate:"required"` // 知识库id
	Name                 string                `json:"name"`
	GraphSwitch          int32                 `json:"graphSwitch"` // 知识图谱开关
	External             int32                 `json:"external"`    // 外部数据库标识
	Category             int32                 `json:"category"`    // 知识库类别 0-知识库，1-问答库，2-多模态知识库
	OrgName              string                `json:"orgName"`     //知识库所属名称
	Share                bool                  `json:"share"`       //是分享，还是私有
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
	Avatar               Avatar                `json:"avatar"`      // 头像
	Description          string                `json:"description"` // 知识库描述
}

type AppKnowledgebaseParams struct {
	MaxHistory int32   `json:"maxHistory"` // 最长上下文
	Threshold  float32 `json:"threshold"`  // 过滤阈值
	TopK       int32   `json:"topK"`       // 知识条数

	MatchType         string  `json:"matchType"`         //matchType：vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）
	PriorityMatch     int32   `json:"priorityMatch"`     // 权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1
	SemanticsPriority float32 `json:"semanticsPriority"` // 语义权重
	KeywordPriority   float32 `json:"keywordPriority"`   // 关键词权重
	TermWeight        float32 `json:"termWeight"`        // 关键词系数，默认为1
	TermWeightEnable  bool    `json:"termWeightEnable"`  // 关键词系数开关
	UseGraph          bool    `json:"useGraph"`          // 知识图谱开关
}

type AppQAKnowledgebaseConfig struct {
	Knowledgebases []AppQAKnowledgeBase     `json:"knowledgebases"` // 问答库id、名字
	Config         AppQAKnowledgebaseParams `json:"config"`         // 问答库参数
}

type AppQAKnowledgeBase struct {
	ID                   string                `json:"id" validate:"required"` // 知识库id
	Name                 string                `json:"name"`
	Category             int32                 `json:"category"` // 知识库类别 0-知识库，1-问答库，2-多模态知识库
	OrgName              string                `json:"orgName"`  //知识库所属名称
	Share                bool                  `json:"share"`    //是分享，还是私有
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
	Avatar               Avatar                `json:"avatar"`      // 头像
	Description          string                `json:"description"` // 描述
}

type AppQAKnowledgebaseParams struct {
	MaxHistory int32   `json:"maxHistory"` // 最长上下文
	Threshold  float32 `json:"threshold"`  // 过滤阈值
	TopK       int32   `json:"topK"`       // 知识条数

	MatchType         string  `json:"matchType"`         //matchType：vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本）
	PriorityMatch     int32   `json:"priorityMatch"`     // 权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1
	SemanticsPriority float32 `json:"semanticsPriority"` // 语义权重
	KeywordPriority   float32 `json:"keywordPriority"`   // 关键词权重
}

type MetaDataFilterParams struct {
	FilterEnable     bool                `json:"filterEnable"`     // 元数据过滤开关
	MetaFilterParams []*MetaFilterParams `json:"metaFilterParams"` // 元数据过滤参数列表
	FilterLogicType  string              `json:"filterLogicType"`  // 元数据逻辑条件：and/or
}

type MetaFilterParams struct {
	Key       string `json:"key"`       // Key
	Type      string `json:"type"`      // 类型（Time, String, Number）
	Condition string `json:"condition"` // 条件
	Value     string `json:"value"`     // value
}

type AppSafetyConfig struct {
	Enable bool             `json:"enable"` // 安全护栏(开关)
	Tables []SensitiveTable `json:"tables"`
}

type SensitiveTable struct {
	TableId   string `json:"tableId" validate:"required"` // 敏感词表id
	TableName string `json:"tableName"`                   // 敏感词表名称(请求非必填)
}

type VisionConfig struct {
	PicNum int32 `json:"picNum"` // 视觉配置图片数量
}

type RecommendConfig struct {
	RecommendEnable bool           `json:"recommendEnable"` // 追问配置开关
	ModelConfig     AppModelConfig `json:"modelConfig"`     // 模型信息
	PromptEnable    bool           `json:"promptEnable"`    // 提示词开关
	Prompt          string         `json:"prompt"`          // 提示词
	MaxHistory      int32          `json:"maxHistory"`      // 最大历史会话轮次
}

type AppPublishConfig struct {
	PublishType string `json:"publishType"`
}
