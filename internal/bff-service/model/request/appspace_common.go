package request

import (
	"encoding/json"
	"fmt"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
)

type Avatar struct {
	Key  string `json:"key"`  // 前端透传给后端用于保存avatar，例如：custom-upload/avatar/abc/def.png [EN] The front end is transparently transmitted to the back end for saving avatar, for example: custom-upload/avatar/abc/def.png
	Path string `json:"path"` // 前端请求地址，例如：/v1/static/avatar/abc/def.png (请求非必填) [EN] Front-end request address, for example: /v1/static/avatar/abc/def.png (request is optional)
}

type AppBriefConfig struct {
	Avatar Avatar `json:"avatar"`                   // 图标 [EN] icon
	Name   string `json:"name" validate:"required"` // 名称 [EN] name
	Desc   string `json:"desc"`                     // 描述 [EN] describe
}

func (a AppBriefConfig) Check() error {
	return nil
}

type AppModelConfig struct {
	Provider    string      `json:"provider"`    // 模型供应商 [EN] model supplier
	Model       string      `json:"model"`       // 模型名称 [EN] Model name
	ModelId     string      `json:"modelId"`     // 模型ID [EN] Model ID
	ModelType   string      `json:"modelType"`   // 模型类型(llm/embedding/rerank) [EN] Model type (llm/embedding/rerank)
	DisplayName string      `json:"displayName"` // 模型展示名称(请求非必填) [EN] Model display name (optional for request)
	Config      interface{} `json:"config"`      // 模型配置 [EN] Model configuration

	Examples *mp.AppModelParams // 仅用于swagger展示；模型对应供应商中的对应llm、embedding或rerank结构是config实际的参数 [EN] Only used for swagger display; the corresponding llm, embedding or rerank structure in the model corresponding supplier is the actual parameter of config
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
	Knowledgebases []AppKnowledgeBase     `json:"knowledgebases"` // 知识库id、名字 [EN] Knowledge base id, name
	Config         AppKnowledgebaseParams `json:"config"`         // 知识库参数 [EN] Knowledge base parameters
}

type AppKnowledgeBase struct {
	ID                   string                `json:"id" validate:"required"` // 知识库id [EN] knowledge base id
	Name                 string                `json:"name"`
	GraphSwitch          int32                 `json:"graphSwitch"` // 知识图谱开关 [EN] Knowledge graph switch
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
}

type AppKnowledgebaseParams struct {
	MaxHistory int32   `json:"maxHistory"` // 最长上下文 [EN] longest context
	Threshold  float32 `json:"threshold"`  // 过滤阈值 [EN] filter threshold
	TopK       int32   `json:"topK"`       // 知识条数 [EN] Number of knowledge items

	MatchType         string  `json:"matchType"`         //matchType：vector（向量检索）、text（文本检索）、mix（混合检索：向量+文本） [EN] matchType: vector (vector search), text (text search), mix (mixed search: vector + text)
	PriorityMatch     int32   `json:"priorityMatch"`     // 权重匹配，只有在混合检索模式下，选择权重设置后，这个才设置为1 [EN] Weight matching. This is only set to 1 after selecting the weight setting in mixed search mode.
	SemanticsPriority float32 `json:"semanticsPriority"` // 语义权重 [EN] semantic weight
	KeywordPriority   float32 `json:"keywordPriority"`   // 关键词权重 [EN] Keyword weight
	TermWeight        float32 `json:"termWeight"`        // 关键词系数，默认为1 [EN] Keyword coefficient, default is 1
	TermWeightEnable  bool    `json:"termWeightEnable"`  // 关键词系数开关 [EN] Keyword coefficient switch
	UseGraph          bool    `json:"useGraph"`          // 知识图谱开关 [EN] Knowledge graph switch
}

type MetaDataFilterParams struct {
	FilterEnable     bool                `json:"filterEnable"`     // 元数据过滤开关 [EN] Metadata filter switch
	MetaFilterParams []*MetaFilterParams `json:"metaFilterParams"` // 元数据过滤参数列表 [EN] Metadata filter parameter list
	FilterLogicType  string              `json:"filterLogicType"`  // 元数据逻辑条件：and/or [EN] Metadata logical conditions: and/or
}

type MetaFilterParams struct {
	Key       string `json:"key"`       // Key
	Type      string `json:"type"`      // 类型（Time, String, Number） [EN] Type (Time, String, Number)
	Condition string `json:"condition"` // 条件 [EN] condition
	Value     string `json:"value"`     // value
}

type AppSafetyConfig struct {
	Enable bool             `json:"enable"` // 安全护栏(开关) [EN] Safety guardrail (switch)
	Tables []SensitiveTable `json:"tables"`
}

type SensitiveTable struct {
	TableId   string `json:"tableId" validate:"required"` // 敏感词表id [EN] Sensitive word list id
	TableName string `json:"tableName"`                   // 敏感词表名称(请求非必填) [EN] Sensitive word list name (request is not required)
}

type VisionConfig struct {
	PicNum int32 `json:"picNum"` // 视觉配置图片数量 [EN] Number of visual configuration pictures
}
