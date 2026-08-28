package rag_manage_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	successCode    = 0
	metaTypeString = "string"
	metaTypeNumber = "number"
	metaTypeTime   = "time"
)

type QAHitParams struct {
	UserId                      string                  `json:"userId"`
	Question                    string                  `json:"question" validate:"required"`
	KnowledgeIdList             []string                `json:"knowledgeIdList" validate:"required"`
	ReturnMeta                  bool                    `json:"returnMeta"`
	Threshold                   float64                 `json:"threshold"`
	TopK                        int64                   `json:"topK"`
	RetrieveMethod              string                  `json:"retrieveMethod"`
	RerankMod                   string                  `json:"rerankMod"`
	RerankModelId               string                  `json:"rerankModelId"`
	MetadataFiltering           bool                    `json:"metadataFiltering"`
	MetadataFilteringConditions []*QAMetadataFilterItem `json:"metadataFilteringConditions"`
	Weight                      *WeightParams           `json:"weights"`
}

type QAMetadataFilterItem struct {
	FilteringQaBaseName string      `json:"filtering_qa_base_name"`
	LogicalOperator     string      `json:"logical_operator"`
	MetaList            []*MetaItem `json:"conditions"` // 元数据过滤列表
}

type RagKnowledgeHitResp struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *HitData `json:"data"`
}

type HitData struct {
	SearchList []*SearchData `json:"searchList"`
	Score      []float64     `json:"score"` //实际是[]float
}

type SearchData struct {
	Title       string      `json:"title"`
	Question    string      `json:"question"`
	Answer      string      `json:"answer"`
	QaPairId    string      `json:"qa_pair_id"`
	QABase      string      `json:"QABase"`
	QAId        string      `json:"QAId"`
	Status      bool        `json:"status"`
	Score       float64     `json:"score"`
	MetaData    interface{} `json:"meta_data"`
	ContentType string      `json:"content_type"`
}

// BuildQaHitParams 构造rag 会话参数
func BuildQaHitParams(req *rag_service.ChatRagReq, rag *model.RagInfo, knowledgeIDToName map[string]string, qaIds []string) (*QAHitParams, error) {
	// 知识库参数
	ragChatParams := &QAHitParams{}
	ragChatParams.UserId = req.Identity.UserId
	ragQAConfigData := rag_service.RagQAKnowledgeBaseConfig{}
	err := json.Unmarshal([]byte(rag.QAKnowledgebaseConfig), &ragQAConfigData)
	if err != nil {
		return nil, errors.New("rag_get_err")
	}
	ragQAConfig := ragQAConfigData.GlobalConfig
	ragChatParams.Threshold = float64(ragQAConfig.Threshold)
	ragChatParams.TopK = int64(ragQAConfig.TopK)
	ragChatParams.RetrieveMethod = buildRetrieveMethod(ragQAConfig.MatchType)
	ragChatParams.RerankMod = buildRerankMod(ragQAConfig.PriorityMatch)
	ragChatParams.Weight = buildQAWeight(ragQAConfig)

	ragChatParams.KnowledgeIdList = qaIds
	ragChatParams.RerankModelId = buildRerankId(ragQAConfig.PriorityMatch, rag.QARerankConfig.ModelId)
	// RAG属性参数
	ragChatParams.Question = req.Question
	ragChatParams.ReturnMeta = true

	metaParams, err := buildMetaDataFilterParams(ragQAConfigData.PerKnowledgeConfigs, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	ragChatParams.MetadataFiltering = len(metaParams) > 0
	ragChatParams.MetadataFilteringConditions = metaParams
	return ragChatParams, nil
}

// RagQASearch rag命中测试
func RagQASearch(ctx context.Context, knowledgeHitParams *QAHitParams) (*RagKnowledgeHitResp, error) {
	paramsByte, err := json.Marshal(knowledgeHitParams)
	if err != nil {
		log.Errorf("ragQASearch params marsh error %v", err)
		return nil, err
	}
	url := fmt.Sprintf("%s%s", config.Cfg().RagServer.ChatEndpoint, config.Cfg().RagServer.QASearchUrl)
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(1) * time.Minute,
		MonitorKey: "rag_qa_hit",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagKnowledgeHitResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf("ragQASearch result Unmarshal error %v", err)
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	return &resp, nil
}

// buildWeight 构造权重信息
func buildQAWeight(qaConfig *rag_service.RagQAGlobalConfig) *WeightParams {
	if qaConfig.PriorityMatch != 1 {
		return nil
	}
	return &WeightParams{
		VectorWeight: qaConfig.SemanticsPriority,
		TextWeight:   qaConfig.KeywordPriority,
	}
}

// buildMetaDataFilterParams 构造元数据过滤参数
func buildMetaDataFilterParams(qaKnowledgeList []*rag_service.RagPerQAKnowledgeConfig, knowledgeIDToName map[string]string) ([]*QAMetadataFilterItem, error) {
	var ragMetaDataFilterParams []*QAMetadataFilterItem
	for _, k := range qaKnowledgeList {
		if k.RagMetaFilter == nil || !k.RagMetaFilter.FilterEnable ||
			len(k.RagMetaFilter.FilterItems) == 0 {
			continue
		}
		item, err := buildMetadataFilterItem(k.RagMetaFilter.FilterItems)
		if err != nil {
			log.Errorf("buildMetaDataFilterParams error %v", err)
			return nil, err
		}
		ragMetaDataFilterParams = append(ragMetaDataFilterParams, &QAMetadataFilterItem{
			FilteringQaBaseName: knowledgeIDToName[k.KnowledgeId],
			LogicalOperator:     k.RagMetaFilter.FilterLogicType,
			MetaList:            item,
		})
	}
	return ragMetaDataFilterParams, nil
}

// buildMetadataFilterItem 构造元数据过滤项
func buildMetadataFilterItem(metaFilterParams []*rag_service.RagMetaFilterItem) ([]*MetaItem, error) {
	var ragMetaDataFilterItem []*MetaItem
	for _, k := range metaFilterParams {
		data, err := buildValueData(k.Type, k.Value, k.Condition)
		if err != nil {
			log.Errorf("buildMetadataFilterItem error %v", err)
			return nil, err
		}
		ragMetaDataFilterItem = append(ragMetaDataFilterItem, &MetaItem{
			ComparisonOperator: k.Condition,
			MetaName:           k.Key,
			MetaType:           k.Type,
			Value:              data,
		})
	}
	return ragMetaDataFilterItem, nil
}

// buildValueData 进行值转换
func buildValueData(valueType string, value string, condition string) (interface{}, error) {
	if condition == "empty" {
		return nil, nil
	}
	switch valueType {
	case metaTypeNumber:
	case metaTypeTime:
		//valueResult, err := parseToTimestamp(value)
		//if err != nil || valueResult == 0 {
		//	return strconv.ParseInt(value, 10, 64)
		//}
		//return valueResult, nil
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}
