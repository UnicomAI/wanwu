package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	http_client "github.com/UnicomAI/wanwu/internal/rag-service/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	DefaultTemperature      = 0.14
	DefaultTopP             = 0.85
	DefaultFrequencyPenalty = 1.1
	DefaultTermWeight       = 1
	InitialBufferSize       = 64 * 1024        // Initial buffer size: 64KB
	MaxBufferCapacity       = 10 * 1024 * 1024 // Maximum buffer capacity: 10MB
	MetaValueTypeNumber     = "number"
	MetaValueTypeTime       = "time"
	MetaConditionEmpty      = "empty"
	MetaConditionNotEmpty   = "not empty"
)

type RagChatParams struct {
	KnowledgeBase        []string                        `json:"knowledgeBase"`        // Knowledge base name list
	KnowledgeIdList      []string                        `json:"knowledgeIdList"`      // Knowledge base ID list
	KnowledgeBaseInfo    map[string][]*RagKnowledgeInfo  `json:"knowledge_base_info"`  // Knowledge base info grouped by user ID (required by Python RAG service)
	Question             string                          `json:"question"`
	Threshold            float32                         `json:"threshold"` // Score threshold
	TopK                 int32                           `json:"topK"`
	Stream               bool                            `json:"stream"`
	Chichat              bool                            `json:"chichat"` // Whether to use the default speech technique (covering the bottom) when the knowledge base recall result is empty, the default is true
	RerankModelId        string                          `json:"rerank_model_id"`
	CustomModelInfo      *CustomModelInfo                `json:"custom_model_info"`
	History              []*HistoryItem                  `json:"history"`
	MaxHistory           int32                           `json:"max_history"`
	RewriteQuery         bool                            `json:"rewrite_query"`   // Whether to rewrite query
	RerankMod            string                          `json:"rerank_mod"`      // rerank_model: reranking mode, weighted_score: weighted search
	RetrieveMethod       string                          `json:"retrieve_method"` // hybrid_search: hybrid search, semantic_search: vector search, full_text_search: text search
	Weight               *WeightParams                   `json:"weights"`         // Weight configuration under weight search
	Temperature          float32                         `json:"temperature"`
	TopP                 float32                         `json:"top_p"`                         // Diversity
	RepetitionPenalty    float32                         `json:"repetition_penalty"`            // Repetition Penalty/Frequency Penalty
	ReturnMeta           bool                            `json:"return_meta"`                   // Whether to return metadata
	AutoCitation         bool                            `json:"auto_citation"`                 // Whether to auto-mark
	TermWeight           float32                         `json:"term_weight_coefficient"`       // keyword coefficient
	MetaFilter           bool                            `json:"metadata_filtering"`            // Metadata filter switch
	MetaFilterConditions []*MetadataFilterItem           `json:"metadata_filtering_conditions"` // Metadata filters
	UseGraph             bool                            `json:"use_graph"`                     // Whether to start knowledge graph query
}

// RagKnowledgeInfo contains knowledge base id and name for Python RAG service
type RagKnowledgeInfo struct {
	KnowledgeId   string `json:"kb_id"`
	KnowledgeName string `json:"kb_name"`
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

type CustomModelInfo struct {
	LlmModelID string `json:"llm_model_id"`
}

type HistoryItem struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type ModelConfig struct {
	Temperature            float32 `json:"temperature"`
	TemperatureEnable      bool    `json:"temperatureEnable"`
	TopP                   float32 `json:"topP"`
	TopPEnable             bool    `json:"topPEnable"`
	FrequencyPenalty       float32 `json:"frequencyPenalty"`
	FrequencyPenaltyEnable bool    `json:"frequencyPenaltyEnable"`
	PresencePenalty        float32 `json:"presencePenalty"`
	PresencePenaltyEnable  bool    `json:"presencePenaltyEnable"`
}

func RagStreamChat(ctx context.Context, userId string, req *RagChatParams) (<-chan string, error) {
	params, err := buildHttpParams(userId, req)
	if err != nil {
		log.Errorf("build http params fail", err.Error())
		return nil, err
	}
	ret := make(chan string, 1024)
	go func() {
		// Ensure the channel is eventually closed
		defer close(ret)

		// Catch panic and log (do not rethrow to avoid crash)
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("RagStreamChat panic: %v", r)
			}
		}()

		//1. Turn on timeout monitoring
		if params.Timeout == 0 {
			params.Timeout = time.Minute * 10
		}
		ctx, cancel := context.WithTimeout(ctx, params.Timeout)
		defer cancel()

		resp, err := http_client.GetClient().PostJsonOriResp(ctx, params)
		if err != nil {
			errMsg := fmt.Sprintf("error: 调用下游服务异常: %v", err)
			log.Errorf(errMsg)
			ret <- errMsg
			return
		}
		defer resp.Body.Close() // Make sure the response body is closed

		if resp.StatusCode != http.StatusOK {
			errMsg := fmt.Sprintf("error: 调用下游服务异常: %s", resp.Status)
			log.Errorf(errMsg)
			ret <- errMsg
			return
		}
		scan := bufio.NewScanner(resp.Body)

		//Set initial buffer to 64KB, maximum allowed is 10MB
		buf := make([]byte, InitialBufferSize)
		scan.Buffer(buf, MaxBufferCapacity)

		for scan.Scan() {
			ret <- scan.Text()
		}
		if err := scan.Err(); err != nil {
			errMsg := fmt.Sprintf("error: 调用下游服务异常: %v", err)
			log.Errorf(errMsg)
			ret <- errMsg
		}
	}()

	return ret, nil
}

func buildHttpParams(userId string, req *RagChatParams) (*http_client.HttpRequestParams, error) {
	url := fmt.Sprintf("%s%s", config.Cfg().RagServer.ChatEndpoint, config.Cfg().RagServer.ChatUrl)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return &http_client.HttpRequestParams{
		Url:        url,
		Body:       body,
		Headers:    map[string]string{"X-uid": userId},
		Timeout:    time.Minute * 10,
		MonitorKey: "rag_search_service",
		LogLevel:   http_client.LogAll,
	}, nil
}

// BuildChatConsultParams constructs rag session parameters
func BuildChatConsultParams(req *rag_service.ChatRagReq, rag *model.RagInfo, knowledgeInfoList *knowledgeBase_service.KnowledgeDetailSelectListResp, knowledgeIds []string) (*RagChatParams, error) {
	// Knowledge base parameters
	ragChatParams := &RagChatParams{}
	knowledgeConfig := rag.KnowledgeBaseConfig
	ragChatParams.MaxHistory = int32(knowledgeConfig.MaxHistory)
	ragChatParams.Threshold = float32(knowledgeConfig.Threshold)
	ragChatParams.TopK = int32(knowledgeConfig.TopK)
	ragChatParams.RetrieveMethod = buildRetrieveMethod(knowledgeConfig.MatchType)
	ragChatParams.RerankMod = buildRerankMod(knowledgeConfig.PriorityMatch)
	ragChatParams.Weight = buildWeight(knowledgeConfig)
	var kbNameList []string
	knowledgeIDToName := make(map[string]string)
	// Build knowledge_base_info map grouped by user ID (required by Python RAG service)
	knowledgeBaseInfo := make(map[string][]*RagKnowledgeInfo)
	for _, v := range knowledgeInfoList.List {
		kbNameList = append(kbNameList, v.Name)
		if _, exists := knowledgeIDToName[v.KnowledgeId]; !exists {
			knowledgeIDToName[v.KnowledgeId] = v.Name
		}
		// Group knowledge base info by user ID (CreateUserId)
		userId := v.CreateUserId
		if userId == "" {
			userId = rag.UserID // fallback to rag owner's user ID
		}
		knowledgeBaseInfo[userId] = append(knowledgeBaseInfo[userId], &RagKnowledgeInfo{
			KnowledgeId:   v.KnowledgeId,
			KnowledgeName: v.RagName, // Use RagName as it's the internal name used for ES/vector store
		})
	}
	ragChatParams.KnowledgeBase = kbNameList
	ragChatParams.KnowledgeIdList = knowledgeIds
	ragChatParams.KnowledgeBaseInfo = knowledgeBaseInfo
	ragChatParams.RerankModelId = buildRerankId(knowledgeConfig.PriorityMatch, rag.RerankConfig.ModelId)
	if rag.KnowledgeBaseConfig.TermWeightEnable {
		ragChatParams.TermWeight = float32(rag.KnowledgeBaseConfig.TermWeight)
	}
	// RAG attribute parameters
	ragChatParams.Question = req.Question
	ragChatParams.Stream = true
	ragChatParams.Chichat = true
	ragChatParams.History = make([]*HistoryItem, 0)
	ragChatParams.RewriteQuery = true
	ragChatParams.ReturnMeta = true
	//Automatic corner mark
	ragChatParams.AutoCitation = true

	// Model parameters
	ragChatParams.CustomModelInfo = &CustomModelInfo{LlmModelID: rag.ModelConfig.ModelId}
	modelConfigStr := rag.ModelConfig.Config
	modelConfig := ModelConfig{}
	err := json.Unmarshal([]byte(modelConfigStr), &modelConfig)
	if err != nil {
		log.Errorf("model config unmarshal fail: %s", modelConfigStr)
		ragChatParams.Temperature = DefaultTemperature
		ragChatParams.TopP = DefaultTopP
		ragChatParams.RepetitionPenalty = DefaultFrequencyPenalty
		return ragChatParams, nil
	}
	if modelConfig.TemperatureEnable {
		ragChatParams.Temperature = modelConfig.Temperature
	} else {
		ragChatParams.Temperature = DefaultTemperature
	}
	if modelConfig.TopPEnable {
		ragChatParams.TopP = modelConfig.TopP
	} else {
		ragChatParams.TopP = DefaultTopP
	}
	if modelConfig.FrequencyPenaltyEnable {
		ragChatParams.RepetitionPenalty = modelConfig.FrequencyPenalty
	} else {
		ragChatParams.RepetitionPenalty = DefaultFrequencyPenalty
	}
	filterEnable, metaParams, err := buildRagMetaParams(rag, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	ragChatParams.MetaFilter = filterEnable
	ragChatParams.MetaFilterConditions = metaParams
	ragChatParams.History = buildHistory(req.History)
	ragChatParams.UseGraph = knowledgeConfig.UseGraph
	log.Infof("ragparams = %+v", http_client.Convert2LogString(ragChatParams))
	return ragChatParams, nil
}

// Build history parameters
func buildHistory(historyList []*rag_service.HistoryItem) []*HistoryItem {
	var retList = make([]*HistoryItem, 0)
	if len(historyList) == 0 {
		return retList
	}
	for _, item := range historyList {
		retList = append(retList, &HistoryItem{
			NeedHistory: item.NeedHistory,
			Query:       item.Query,
			Response:    item.Response,
		})
	}
	return retList
}

func buildRagMetaParams(rag *model.RagInfo, knowledgeIDToName map[string]string) (bool, []*MetadataFilterItem, error) {
	var perKbConfig []*rag_service.RagPerKnowledgeConfig
	if rag.KnowledgeBaseConfig.MetaParams != "" {
		err := json.Unmarshal([]byte(rag.KnowledgeBaseConfig.MetaParams), &perKbConfig)
		if err != nil {
			return false, nil, errors.New("rag meta params unmarshal fail: " + err.Error())
		}
	}
	filterEnable := false // Whether the tag has metadata filtering enabled
	var metaFilterConditions []*MetadataFilterItem
	for _, k := range perKbConfig {
		// Check if metadata filtering parameters are valid
		filterParams := k.RagMetaFilter
		if !isValidFilterParams(k.RagMetaFilter) {
			continue
		}
		// Verify legal value
		if k.RagMetaFilter.FilterLogicType == "" {
			return false, nil, errors.New("rag meta FilterLogicType is empty")
		}
		// Tag metadata filtering takes effect
		filterEnable = true
		// Build metadata filters
		metaItems, err := buildRagMetaItems(k.KnowledgeId, filterParams.FilterItems)
		if err != nil {
			return false, nil, err
		}
		// Add filter items to results
		metaFilterConditions = append(metaFilterConditions, &MetadataFilterItem{
			FilterKnowledgeName: knowledgeIDToName[k.KnowledgeId],
			LogicalOperator:     filterParams.FilterLogicType,
			Conditions:          metaItems,
		})
	}
	return filterEnable, metaFilterConditions, nil
}

func isValidFilterParams(params *rag_service.RagMetaFilter) bool {
	return params != nil &&
		params.FilterEnable &&
		params.FilterItems != nil &&
		len(params.FilterItems) > 0
}

// Build a list of metadata items
func buildRagMetaItems(knowledgeID string, params []*rag_service.RagMetaFilterItem) ([]*MetaItem, error) {
	var metaItems []*MetaItem
	for _, param := range params {
		// Basic parameter verification
		if err := validateMetaFilterParam(knowledgeID, param); err != nil {
			return nil, err
		}
		// Conversion parameter value
		ragValue, err := convertValue(param.Value, param.Type)
		if err != nil {
			log.Errorf("kbId: %s, convert value failed: %v", knowledgeID, err)
			return nil, fmt.Errorf("convert value for key %s failed: %s", param.Key, err.Error())
		}
		metaItems = append(metaItems, &MetaItem{
			MetaName:           param.Key,
			MetaType:           param.Type,
			ComparisonOperator: param.Condition,
			Value:              ragValue,
		})
	}
	return metaItems, nil
}

func convertValue(value, valueType string) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	// Convert value according to type
	if valueType == MetaValueTypeNumber {
		ragValue, err := strconv.Atoi(value)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return ragValue, nil
	}
	if valueType == MetaValueTypeTime {
		parseInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return parseInt, nil
	}
	return value, nil
}

// Verify metadata filter parameters
func validateMetaFilterParam(knowledgeID string, param *rag_service.RagMetaFilterItem) error {
	// Check if key parameter is empty
	if param.Key == "" || param.Type == "" || param.Condition == "" {
		errMsg := "key/type/condition cannot be empty"
		log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
		return errors.New(errMsg)
	}

	// Check the null condition for matching with the value
	if param.Condition == MetaConditionEmpty || param.Condition == MetaConditionNotEmpty {
		if param.Value != "" {
			errMsg := "condition is empty/non-empty, value should be empty"
			log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
			return errors.New(errMsg)
		}
	} else {
		if param.Value == "" {
			errMsg := "value is empty"
			log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
}

// buildRerankId constructs the reranking model id
func buildRerankId(priorityType int32, rerankId string) string {
	if priorityType == 1 {
		return ""
	}
	return rerankId
}

// buildRetrieveMethod constructs the retrieval method
func buildRetrieveMethod(matchType string) string {
	switch matchType {
	case "vector":
		return "semantic_search" // vector search
	case "text":
		return "full_text_search" // Full text search
	case "mix":
		return "hybrid_search" // Hybrid search
	}
	return ""
}

// buildRerankMod constructs reranking mode
func buildRerankMod(priorityType int32) string {
	if priorityType == 1 {
		return "weighted_score"
	}
	return "rerank_model"
}

// buildWeight constructs weight information
func buildWeight(knowConfig model.KnowledgeBaseConfig) *WeightParams {
	if knowConfig.PriorityMatch != 1 {
		return nil
	}
	return &WeightParams{
		VectorWeight: float32(knowConfig.SemanticsPriority),
		TextWeight:   float32(knowConfig.KeywordPriority),
	}
}
