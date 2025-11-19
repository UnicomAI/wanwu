package request

import (
	"encoding/json"
	"fmt"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
)

type Avatar struct {
	Key  string `json:"key"`  // The front end is transparently transmitted to the back end for saving avatar, for example: custom-upload/avatar/abc/def.png
	Path string `json:"path"` // Front-end request address, for example: /v1/static/avatar/abc/def.png (request is optional)
}

type AppBriefConfig struct {
	Avatar Avatar `json:"avatar"`                   // icon
	Name   string `json:"name" validate:"required"` // name
	Desc   string `json:"desc"`                     // describe
}

func (a AppBriefConfig) Check() error {
	return nil
}

type AppModelConfig struct {
	Provider    string      `json:"provider"`    // model supplier
	Model       string      `json:"model"`       // Model name
	ModelId     string      `json:"modelId"`     // Model ID
	ModelType   string      `json:"modelType"`   // Model type (llm/embedding/rerank)
	DisplayName string      `json:"displayName"` // Model display name (optional for request)
	Config      interface{} `json:"config"`      // Model configuration

	Examples *mp.AppModelParams // Only used for swagger display; the corresponding llm, embedding or rerank structure in the model corresponding supplier is the actual parameter of config
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
	Knowledgebases []AppKnowledgeBase     `json:"knowledgebases"` // Knowledge base id, name
	Config         AppKnowledgebaseParams `json:"config"`         // Knowledge base parameters
}

type AppKnowledgeBase struct {
	ID                   string                `json:"id" validate:"required"` // knowledge base id
	Name                 string                `json:"name"`
	GraphSwitch          int32                 `json:"graphSwitch"` // Knowledge graph switch
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
}

type AppKnowledgebaseParams struct {
	MaxHistory int32   `json:"maxHistory"` // longest context
	Threshold  float32 `json:"threshold"`  // filter threshold
	TopK       int32   `json:"topK"`       // Number of knowledge items

	MatchType         string  `json:"matchType"`         //matchType: vector (vector search), text (text search), mix (mixed search: vector + text)
	PriorityMatch     int32   `json:"priorityMatch"`     // Weight matching. This is only set to 1 after selecting the weight setting in mixed search mode.
	SemanticsPriority float32 `json:"semanticsPriority"` // semantic weight
	KeywordPriority   float32 `json:"keywordPriority"`   // Keyword weight
	TermWeight        float32 `json:"termWeight"`        // Keyword coefficient, default is 1
	TermWeightEnable  bool    `json:"termWeightEnable"`  // Keyword coefficient switch
	UseGraph          bool    `json:"useGraph"`          // Knowledge graph switch
}

type MetaDataFilterParams struct {
	FilterEnable     bool                `json:"filterEnable"`     // Metadata filter switch
	MetaFilterParams []*MetaFilterParams `json:"metaFilterParams"` // Metadata filter parameter list
	FilterLogicType  string              `json:"filterLogicType"`  // Metadata logical conditions: and/or
}

type MetaFilterParams struct {
	Key       string `json:"key"`       // Key
	Type      string `json:"type"`      // Type (Time, String, Number)
	Condition string `json:"condition"` // condition
	Value     string `json:"value"`     // value
}

type AppSafetyConfig struct {
	Enable bool             `json:"enable"` // Safety guardrail (switch)
	Tables []SensitiveTable `json:"tables"`
}

type SensitiveTable struct {
	TableId   string `json:"tableId" validate:"required"` // Sensitive word list id
	TableName string `json:"tableName"`                   // Sensitive word list name (request is not required)
}

type VisionConfig struct {
	PicNum int32 `json:"picNum"` // Number of visual configuration pictures
}
