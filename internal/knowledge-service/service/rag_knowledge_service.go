package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/mq"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/http"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	successCode = 0
)

type RagCreateParams struct {
	UserId               string `json:"userId"`
	Name                 string `json:"knowledgeBase"`
	KnowledgeBaseId      string `json:"kb_id"`
	EmbeddingModelId     string `json:"embedding_model_id"`
	EnableKnowledgeGraph bool   `json:"enable_knowledge_graph"`
}

type RagCommonResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RagDocSegmentResp struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    SegmentResult `json:"data"`
}

type SegmentResult struct {
	SuccessCount int `json:"success_count"` // Number of successful segments imported
}

type RagUpdateParams struct {
	UserId          string `json:"userId"`
	KnowledgeBaseId string `json:"kb_id"`
	OldKbName       string `json:"old_kb_name"`
	NewKbName       string `json:"new_kb_name"`
}

type RagDeleteParams struct {
	UserId            string `json:"userId"`
	KnowledgeBaseName string `json:"knowledgeBase"`
	KnowledgeId       string `json:"kb_id"`
}

type KnowledgeHitParams struct {
	UserId               string                `json:"userId"`
	Question             string                `json:"question" validate:"required"`
	KnowledgeBase        []string              `json:"knowledgeBase" validate:"required"`
	KnowledgeIdList      []string              `json:"knowledgeIdList" validate:"required"`
	Threshold            float64               `json:"threshold"`
	TopK                 int32                 `json:"topK"`
	RerankModelId        string                `json:"rerank_model_id"`               // rerankId
	RerankMod            string                `json:"rerank_mod"`                    // rerank_model: reranking mode, weighted_score: weighted search
	RetrieveMethod       string                `json:"retrieve_method"`               // hybrid_search: hybrid search, semantic_search: vector search, full_text_search: text search
	Weight               *WeightParams         `json:"weights"`                       // Weight configuration under weight search
	TermWeight           float32               `json:"term_weight_coefficient"`       // keyword coefficient
	MetaFilter           bool                  `json:"metadata_filtering"`            // Metadata filter switch
	MetaFilterConditions []*MetadataFilterItem `json:"metadata_filtering_conditions"` // Metadata filters
	UseGraph             bool                  `json:"use_graph"`                     // Whether to use knowledge graph
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

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //semantic weight
	TextWeight   float32 `json:"text_weight"`   //keyword weight
}

type RagKnowledgeHitResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *KnowledgeHitData `json:"data"`
}

type KnowledgeHitData struct {
	Prompt     string             `json:"prompt"`
	SearchList []*ChunkSearchList `json:"searchList"`
	Score      []float64          `json:"score"`
	UseGraph   bool               `json:"use_graph"`
}

type ChunkSearchList struct {
	Title            string          `json:"title"`
	Snippet          string          `json:"snippet"`
	KbName           string          `json:"kb_name"`
	MetaData         interface{}     `json:"meta_data"`
	ChildContentList []*ChildContent `json:"child_content_list"`
	ChildScore       []float64       `json:"child_score"`
	ContentType      string          `json:"content_type"` // graph: knowledge graph (text), text: document segmentation (text), community_report: community report (markdown)
}

type ChildContent struct {
	ChildSnippet string  `json:"child_snippet"`
	Score        float64 `json:"score"`
}

type RagBatchDeleteMetaParams struct {
	UserId        string   `json:"userId"`        // user id
	KnowledgeBase string   `json:"knowledgeBase"` // Knowledge base name
	KnowledgeId   string   `json:"kb_id"`         // knowledge base id
	Keys          []string `json:"keys"`          // List of deleted metadata keys
}

type RagBatchUpdateMetaKeyParams struct {
	UserId        string            `json:"userId"`        // user id
	KnowledgeBase string            `json:"knowledgeBase"` // Knowledge base name
	KnowledgeId   string            `json:"kb_id"`         // knowledge base id
	Mappings      []*RagMetaMapKeys `json:"mappings"`      // Metadata key mapping list
}

type RagMetaMapKeys struct {
	OldKey string `json:"old_key"`
	NewKey string `json:"new_key"`
}

type RagKnowledgeGraphParams struct {
	UserId        string `json:"userId"`        // user id
	KnowledgeBase string `json:"knowledgeBase"` // Knowledge base name
	KnowledgeId   string `json:"kb_id"`         // knowledge base id
}

type RagKnowledgeGraphResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// RagKnowledgeCreate rag creates a knowledge base
func RagKnowledgeCreate(ctx context.Context, ragCreateParams *RagCreateParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.InitKnowledgeUri
	paramsByte, err := json.Marshal(ragCreateParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_knowledge_create",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}

	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagCreateKnowledgeReport creates a knowledge base community report
func RagCreateKnowledgeReport(ctx context.Context, ragImportDocParams *RagImportDocParams) error {
	ragImportDocParams.MessageType = RagCommunityReport
	return mq.SendMessage(&RagOperationParams{
		Operation: "add",
		Type:      "doc",
		Doc:       ragImportDocParams,
	}, config.GetConfig().Kafka.KnowledgeGraphTopic)
}

// RagKnowledgeUpdate rag updates the knowledge base
func RagKnowledgeUpdate(ctx context.Context, ragUpdateParams *RagUpdateParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.UpdateKnowledgeUri
	paramsByte, err := json.Marshal(ragUpdateParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_knowledge_update",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagKnowledgeDelete rag update knowledge base delete
func RagKnowledgeDelete(ctx context.Context, ragDeleteParams *RagDeleteParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.DeleteKnowledgeUri
	paramsByte, err := json.Marshal(ragDeleteParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_knowledge_delete",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		if strings.Contains(resp.Message, "文档不存在") {
			return nil
		}
		return errors.New(resp.Message)
	}
	return nil
}

// RagKnowledgeHit rag hit test
func RagKnowledgeHit(ctx context.Context, knowledgeHitParams *KnowledgeHitParams) (*RagKnowledgeHitResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.ProxyPoint + ragServer.KnowledgeHitUri
	paramsByte, err := json.Marshal(knowledgeHitParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_knowledge_hit",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagKnowledgeHitResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	return &resp, nil
}

func RagBatchDeleteMeta(ctx context.Context, ragDeleteParams *RagBatchDeleteMetaParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.BatchDeleteMetaKeyUri
	paramsByte, err := json.Marshal(ragDeleteParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_delete_meta_key",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

func RagBatchUpdateMeta(ctx context.Context, ragUpdateParams *RagBatchUpdateMetaKeyParams) error {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.BatchRenameMetakeyUri
	paramsByte, err := json.Marshal(ragUpdateParams)
	if err != nil {
		return err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_update_meta_key",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return err
	}
	var resp RagCommonResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return err
	}
	if resp.Code != successCode {
		return errors.New(resp.Message)
	}
	return nil
}

// RagKnowledgeGraph rag knowledge graph
func RagKnowledgeGraph(ctx context.Context, knowledgeGraphParams *RagKnowledgeGraphParams) (*RagKnowledgeGraphResp, error) {
	ragServer := config.GetConfig().RagServer
	url := ragServer.Endpoint + ragServer.KnowledgeGraphUri
	paramsByte, err := json.Marshal(knowledgeGraphParams)
	if err != nil {
		return nil, err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(ragServer.Timeout) * time.Second,
		MonitorKey: "rag_knowledge_graph",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagKnowledgeGraphResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	return &resp, nil
}
