package request

import (
	"errors"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"
)

type KnowledgeSelectReq struct {
	Name      string   `json:"name" form:"name" `
	TagIdList []string `json:"tagId" form:"tagId" `
	CommonCheck
}

type KnowledgeBatchSelectReq struct {
	KnowledgeIdList []string `json:"knowledgeIdList" form:"knowledgeIdList" `
	UserId          string   `json:"userId" form:"userId" `
	CommonCheck
}

type CreateKnowledgeReq struct {
	Name           string          `json:"name"  validate:"required"`
	Description    string          `json:"description"`
	EmbeddingModel *EmbeddingModel `json:"embeddingModelInfo" validate:"required"`
	KnowledgeGraph *KnowledgeGraph `json:"knowledgeGraph" validate:"required"`
}

type UpdateKnowledgeReq struct {
	KnowledgeId string `json:"knowledgeId"   validate:"required"`
	Name        string `json:"name"   validate:"required"`
	Description string `json:"description"`
	CommonCheck
}

type KnowledgeHitReq struct {
	KnowledgeList        []*AppKnowledgeBase   `json:"knowledgeList"`
	Question             string                `json:"question"   validate:"required"`
	KnowledgeMatchParams *KnowledgeMatchParams `json:"knowledgeMatchParams"   validate:"required"`
	CommonCheck
}

type KnowledgeMatchParams struct {
	MatchType         string  `json:"matchType"  validate:"required"` //matchType: vector (vector search), text (text search), mix (mixed search: vector + text)
	RerankModelId     string  `json:"rerankModelId"`                  //rerank model id
	PriorityMatch     int32   `json:"priorityMatch"`                  // Weight matching. This is only set to 1 after selecting the weight setting in mixed search mode.
	SemanticsPriority float32 `json:"semanticsPriority"`              // semantic weight
	KeywordPriority   float32 `json:"keywordPriority"`                // Keyword weight
	TopK              int32   `json:"topK"`                           //topK gets the highest rows
	Threshold         float32 `json:"threshold"`                      //threshold filter score threshold
	TermWeight        float32 `json:"termWeight"`                     // keyword coefficient
	TermWeightEnable  bool    `json:"termWeightEnable"`               // Keyword coefficient switch
	UseGraph          bool    `json:"useGraph"`                       // Whether to use knowledge graph
	CommonCheck
}

type EmbeddingModel struct {
	ModelId string `json:"modelId"  validate:"required"`
}

// KnowledgeGraph knowledge graph information
type KnowledgeGraph struct {
	Switch     bool   `json:"switch"`     //Knowledge graph switch
	LLMModelId string `json:"llmModelId"` //Large model id, the switch is true and required
	SchemaUrl  string `json:"schemaUrl"`  //Model schema file address, can be empty
}

type DeleteKnowledge struct {
	KnowledgeId string `json:"knowledgeId" validate:"required"`
	CommonCheck
}

type GetKnowledgeReq struct {
	KnowledgeId string `json:"knowledgeId" validate:"required"`
	CommonCheck
}

type CallbackUpdateDocStatusReq struct {
	DocId        string              `json:"id" validate:"required"`
	Status       int32               `json:"status" validate:"required"`
	MetaDataList []*CallbackMetaData `json:"metaDataList"`
	CommonCheck
}

type CallbackUpdateKnowledgeStatusReq struct {
	KnowledgeId  string `json:"knowledgeId" validate:"required"`
	ReportStatus int32  `json:"reportStatus" validate:"required"` //This status will not be 0
	CommonCheck
}

type CallbackMetaData struct {
	Key    string `json:"key"`
	MetaId string `json:"metaId" validate:"required"`
	Value  string `json:"value" validate:"required"`
}

type DocMetaData struct {
	MetaId        string `json:"metaId"`        // metadata id
	MetaKey       string `json:"metaKey"`       // key
	MetaValue     string `json:"metaValue"`     // Determine value
	MetaValueType string `json:"metaValueType"` // string，number，time
	MetaRule      string `json:"metaRule"`      // regular expression
	Option        string `json:"option"`        // option: add (new), update (update), delete (delete), metaId cannot be empty when updating and deleting.
}

type SearchKnowledgeInfoReq struct {
	KnowledgeName string `json:"categoryName" form:"categoryName" validate:"required"`
	UserId        string `json:"userId" form:"userId" validate:"required"`
	OrgId         string `json:"orgId"`
	CommonCheck
}

type GetKnowledgeMetaSelectReq struct {
	KnowledgeId string `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	CommonCheck
}

type KnowledgeMetaValueListReq struct {
	KnowledgeId string   `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	DocIdList   []string `json:"docIdList" form:"docIdList" validate:"required" `
	CommonCheck
}

type UpdateMetaValueReq struct {
	KnowledgeId     string         `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	DocIdList       []string       `json:"docIdList"  validate:"required"`
	MetaValueList   []*DocMetaData `json:"metaValueList"`
	ApplyToSelected bool           `json:"applyToSelected"`
}

// RagSearchKnowledgeBaseReq rag knowledge base query request
type RagSearchKnowledgeBaseReq struct {
	UserId               string                         `json:"userId" validate:"required"`
	Question             string                         `json:"question" validate:"required"`
	KnowledgeIdList      []string                       `json:"knowledgeIdList,omitempty" validate:"required"`
	KnowledgeUser        map[string][]*RagKnowledgeInfo `json:"knowledge_base_info"`
	Threshold            float64                        `json:"threshold"`
	TopK                 int32                          `json:"topK"`
	RerankModelId        string                         `json:"rerank_model_id"`               // rerankId
	RerankMod            string                         `json:"rerank_mod"`                    // rerank_model: reranking mode, weighted_score: weighted search
	RetrieveMethod       string                         `json:"retrieve_method"`               // hybrid_search: hybrid search, semantic_search: vector search, full_text_search: text search
	Weight               *WeightParams                  `json:"weights"`                       // Weight configuration under weight search
	TermWeight           float32                        `json:"term_weight_coefficient"`       // keyword coefficient
	MetaFilter           bool                           `json:"metadata_filtering"`            // Metadata filter switch
	MetaFilterConditions []*MetadataFilterItem          `json:"metadata_filtering_conditions"` // Metadata filters
	UseGraph             bool                           `json:"use_graph"`                     // Whether to start knowledge graph query
	CommonCheck
}

type RagKnowledgeChatReq struct {
	UserId               string                         `json:"userId"`
	KnowledgeUser        map[string][]*RagKnowledgeInfo `json:"knowledge_base_info"`
	KnowledgeIdList      []string                       `json:"knowledgeIdList"` // Knowledge base id list
	Question             string                         `json:"question"`
	Threshold            float32                        `json:"threshold"` // Score threshold
	TopK                 int32                          `json:"topK"`
	Stream               bool                           `json:"stream"`
	Chichat              bool                           `json:"chichat"` // Whether to use the default speech technique (covering the bottom) when the knowledge base recall result is empty, the default is true
	RerankModelId        string                         `json:"rerank_model_id"`
	CustomModelInfo      *CustomModelInfo               `json:"custom_model_info"`
	History              []*HistoryItem                 `json:"history"`
	MaxHistory           int32                          `json:"max_history"`
	RewriteQuery         bool                           `json:"rewrite_query"`   // Whether to rewrite query
	RerankMod            string                         `json:"rerank_mod"`      // rerank_model: reranking mode, weighted_score: weighted search
	RetrieveMethod       string                         `json:"retrieve_method"` // hybrid_search: hybrid search, semantic_search: vector search, full_text_search: text search
	Weight               *WeightParams                  `json:"weights"`         // Weight configuration under weight search
	Temperature          float32                        `json:"temperature,omitempty"`
	TopP                 float32                        `json:"top_p,omitempty"`               // Diversity
	RepetitionPenalty    float32                        `json:"repetition_penalty,omitempty"`  // Repetition Penalty/Frequency Penalty
	ReturnMeta           bool                           `json:"return_meta,omitempty"`         // Whether to return metadata
	AutoCitation         bool                           `json:"auto_citation"`                 // Whether to auto-mark
	TermWeight           float32                        `json:"term_weight_coefficient"`       // keyword coefficient
	MetaFilter           bool                           `json:"metadata_filtering"`            // Metadata filter switch
	MetaFilterConditions []*MetadataFilterItem          `json:"metadata_filtering_conditions"` // Metadata filters
	UseGraph             bool                           `json:"use_graph"`                     // Whether to start knowledge graph query
	CommonCheck
}

type KnowledgeGraphReq struct {
	KnowledgeId string `json:"knowledgeId"  form:"knowledgeId" validate:"required"`
	CommonCheck
}

type CustomModelInfo struct {
	LlmModelID string `json:"llm_model_id"`
}

type HistoryItem struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type RagKnowledgeInfo struct {
	KnowledgeId   string `json:"kb_id"`
	KnowledgeName string `json:"kb_name"`
}

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //semantic weight
	TextWeight   float32 `json:"text_weight"`   //keyword weight
}

type MetadataFilterItem struct {
	FilterKnowledgeName string      `json:"filtering_kb_name"`
	LogicalOperator     string      `json:"logical_operator"`
	Conditions          []*MetaItem `json:"conditions"`
}

type MetaItem struct {
	MetaName           string      `json:"meta_name"`           // metadata name
	MetaType           string      `json:"meta_type"`           // metadata type
	ComparisonOperator string      `json:"comparison_operator"` // comparison operator
	Value              interface{} `json:"value,omitempty"`     // condition value for filtering
}

func (c *UpdateMetaValueReq) Check() error {
	for _, v := range c.MetaValueList {
		if v.Option == "" {
			return errors.New("option为空")
		}
	}
	return nil
}
func (c *CreateKnowledgeReq) Check() error {
	if !util.IsAlphanumeric(c.Name) {
		errMsg := fmt.Sprintf("知识库名称只能包含中文、数字、小写英文，符号之只能包含下划线和减号 参数(%v)", c.Name)
		return errors.New(errMsg)
	}
	if c.KnowledgeGraph == nil {
		return errors.New("knowledge graph can not be nil")
	}
	if c.KnowledgeGraph.Switch && c.KnowledgeGraph.LLMModelId == "" {
		return errors.New("knowledge graph llmModelId can not be empty")
	}
	return nil
}
