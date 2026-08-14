package rag_manage_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	DefaultTemperature      = 0.14
	DefaultTopP             = 0.85
	DefaultFrequencyPenalty = 1.1
	DefaultThinkingEnable   = true
	DefaultTermWeight       = 1
	InitialBufferSize       = 64 * 1024        // 初始缓冲区大小：64KB
	MaxBufferCapacity       = 10 * 1024 * 1024 // 最大缓冲区容量：10MB
	MetaValueTypeNumber     = "number"
	MetaValueTypeTime       = "time"
	MetaConditionEmpty      = "empty"
	MetaConditionNotEmpty   = "not empty"
)

type RagChatParams struct {
	KnowledgeIdList      []string              `json:"knowledgeIdList"` // 知识库ID列表
	Question             string                `json:"question"`
	Threshold            float32               `json:"threshold"` // Score阈值
	TopK                 int32                 `json:"topK"`
	Stream               bool                  `json:"stream"`
	Chichat              bool                  `json:"chichat"` // 当知识库召回结果为空时是否使用默认话术（兜底），默认为false
	RerankModelId        string                `json:"rerank_model_id"`
	CustomModelInfo      *CustomModelInfo      `json:"custom_model_info"`
	History              []*HistoryItem        `json:"history"`
	MaxHistory           int32                 `json:"max_history"`
	RewriteQuery         bool                  `json:"rewrite_query"`   // 是否query改写
	RerankMod            string                `json:"rerank_mod"`      // rerank_model:重排序模式，weighted_score：权重搜索
	RetrieveMethod       string                `json:"retrieve_method"` // hybrid_search:混合搜索， semantic_search:向量搜索， full_text_search：文本搜索
	Weight               *WeightParams         `json:"weights"`         // 权重搜索下的权重配置
	Temperature          float32               `json:"temperature"`
	TopP                 float32               `json:"top_p"`                         // 多样性
	RepetitionPenalty    float32               `json:"repetition_penalty"`            // 重复惩罚/频率惩罚
	ReturnMeta           bool                  `json:"return_meta"`                   // 是否返回元数据
	AutoCitation         bool                  `json:"auto_citation"`                 // 是否自动角标
	TermWeight           float32               `json:"term_weight_coefficient"`       // 关键词系数
	MetaFilter           bool                  `json:"metadata_filtering"`            // 元数据过滤开关
	MetaFilterConditions []*MetadataFilterItem `json:"metadata_filtering_conditions"` // 元数据过滤条件
	UseGraph             bool                  `json:"use_graph"`                     // 是否启动知识图谱查询
	AttachmentList       []*AttachmentInfo     `json:"attachment_files"`              // 上传的文件
	EnableThinking       bool                  `json:"enable_thinking"`               // 是否启用思考模式
}

type MetadataFilterItem struct {
	FilterKnowledgeName string      `json:"filtering_kb_name"`
	LogicalOperator     string      `json:"logical_operator"`
	Conditions          []*MetaItem `json:"conditions"`
}

type MetaItem struct {
	MetaName           string      `json:"meta_name"`           // 元数据名称
	MetaType           string      `json:"meta_type"`           // 元数据类型
	ComparisonOperator string      `json:"comparison_operator"` // 比较运算符
	Value              interface{} `json:"value,omitempty"`     // 用于过滤的条件值
}

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //语义权重
	TextWeight   float32 `json:"text_weight"`   //关键字权重
}

type CustomModelInfo struct {
	LlmModelID string `json:"llm_model_id"`
}

type HistoryItem struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type AttachmentInfo struct {
	FileType string `json:"file_type"`
	FileUrl  string `json:"file_url"`
}

func RagStreamChat(ctx context.Context, userId string, req *RagChatParams) (<-chan string, error) {
	//1.执行http请求
	sseResp, err := requestRagStreamChat(ctx, userId, req)
	if err != nil {
		return nil, err
	}
	//2.读取结果
	SSEReader := &sse_util.SSEReader[string]{
		BusinessKey:    "rag_stream_chat",
		StreamReceiver: sse_util.NewHttpStreamReceiver(sseResp),
	}
	return SSEReader.ReadStream(ctx)
}

// requestRagStreamChat 执行rag流式会话
func requestRagStreamChat(ctx context.Context, userId string, req *RagChatParams) (*http.Response, error) {
	params, err := buildHttpParams(userId, req)
	if err != nil {
		log.Errorf("build http params fail %s", err.Error())
		return nil, err
	}
	sseResp, err := http_client.Default().PostJsonOriResp(ctx, params)
	if err != nil {
		log.Errorf("error: 调用下游服务异常: %v", err)
		return nil, fmt.Errorf("error: 调用下游服务异常: %v", err)
	}
	if sseResp.StatusCode != http.StatusOK {
		all, err := io.ReadAll(sseResp.Body)
		var errResp = ""
		if len(all) > 0 {
			errResp = string(all)
		}
		log.Errorf("error: %s 调用下游服务异常: %v", errResp, err)
		return nil, fmt.Errorf("error: 调用下游服务异常: %v", err)
	}
	return sseResp, nil
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

// BuildChatConsultParams 构造rag 会话参数
func BuildChatConsultParams(req *rag_service.ChatRagReq, rag *model.RagInfo, knowledgeIDToName map[string]string, knowledgeIds []string) (*RagChatParams, error) {
	// rag固定属性
	ragChatParams := &RagChatParams{
		Question:     req.Question,
		Stream:       true,
		Chichat:      false,
		RewriteQuery: true,
		ReturnMeta:   true,
		AutoCitation: true,
		History:      make([]*HistoryItem, 0),
	}

	// 知识库参数
	kc := rag.KnowledgeBaseConfig
	ragChatParams.MaxHistory = int32(kc.MaxHistory)
	ragChatParams.Threshold = float32(kc.Threshold)
	ragChatParams.TopK = int32(kc.TopK)
	ragChatParams.RetrieveMethod = buildRetrieveMethod(kc.MatchType)
	ragChatParams.RerankMod = buildRerankMod(kc.PriorityMatch)
	ragChatParams.Weight = buildWeight(kc)
	ragChatParams.KnowledgeIdList = knowledgeIds
	ragChatParams.RerankModelId = buildRerankId(kc.PriorityMatch, rag.RerankConfig.ModelId)
	ragChatParams.UseGraph = kc.UseGraph
	if kc.TermWeightEnable {
		ragChatParams.TermWeight = float32(kc.TermWeight)
	}

	// 模型参数
	ragChatParams.CustomModelInfo = &CustomModelInfo{LlmModelID: rag.ModelConfig.ModelId}
	if err := applyModelParams(ragChatParams, rag.ModelConfig); err != nil {
		return nil, err
	}

	// 元数据过滤
	filterEnable, metaParams, err := buildRagMetaParams(rag, knowledgeIDToName)
	if err != nil {
		return nil, err
	}
	ragChatParams.MetaFilter = filterEnable
	ragChatParams.MetaFilterConditions = metaParams

	ragChatParams.History = buildHistory(req.History)
	ragChatParams.AttachmentList = buildAttachmentList(req.FileInfoList)

	log.Infof("rag params = %+v", http_client.Convert2LogString(ragChatParams))
	return ragChatParams, nil
}

// applyModelParams 将模型配置写入ragChatParams，解析失败时使用默认值
func applyModelParams(ragChatParams *RagChatParams, modelCfg model.AppModelConfig) error {
	_, configMap, err := mp.ToModelParams(modelCfg.Provider, modelCfg.ModelType, modelCfg.Config)
	if err != nil {
		log.Errorf("model config unmarshal fail, config: %s, err: %v", modelCfg.Config, err)
		ragChatParams.Temperature = DefaultTemperature
		ragChatParams.TopP = DefaultTopP
		ragChatParams.RepetitionPenalty = DefaultFrequencyPenalty
		ragChatParams.EnableThinking = DefaultThinkingEnable
		return nil
	}

	ragChatParams.Temperature = getFloat32(configMap, "temperature", DefaultTemperature)
	ragChatParams.TopP = getFloat32(configMap, "top_p", DefaultTopP)
	ragChatParams.RepetitionPenalty = getFloat32(configMap, "frequency_penalty", DefaultFrequencyPenalty)
	ragChatParams.EnableThinking = getBool(configMap, "enable_thinking", DefaultThinkingEnable)
	return nil
}

func getFloat32(m map[string]interface{}, key string, defaultVal float32) float32 {
	v, ok := m[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch val := v.(type) {
	case float32:
		return val
	case float64:
		return float32(val) // 保险起见兜底
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return defaultVal
	}
	if val, ok := v.(bool); ok {
		return val
	}
	return defaultVal
}

func buildAttachmentList(fileInfos []*rag_service.FileInfo) []*AttachmentInfo {
	retList := make([]*AttachmentInfo, 0)
	for _, file := range fileInfos {
		ext := strings.ToLower(filepath.Ext(file.FileUrl))
		switch ext {
		case ".png", ".jpg", ".jpeg":
			retList = append(retList, &AttachmentInfo{
				FileType: "image",
				FileUrl:  file.FileUrl,
			})
		}
	}
	return retList
}

// 构建历史参数
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
	filterEnable := false // 标记是否有启用的元数据过滤
	var metaFilterConditions []*MetadataFilterItem
	for _, k := range perKbConfig {
		// 检查元数据过滤参数是否有效
		filterParams := k.RagMetaFilter
		if !isValidFilterParams(k.RagMetaFilter) {
			continue
		}
		// 校验合法值
		if k.RagMetaFilter.FilterLogicType == "" {
			return false, nil, errors.New("rag meta FilterLogicType is empty")
		}
		// 标记元数据过滤生效
		filterEnable = true
		// 构建元数据过滤条件
		metaItems, err := buildRagMetaItems(k.KnowledgeId, filterParams.FilterItems)
		if err != nil {
			return false, nil, err
		}
		// 添加过滤项到结果
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
		len(params.FilterItems) > 0
}

// 构建元数据项列表
func buildRagMetaItems(knowledgeID string, params []*rag_service.RagMetaFilterItem) ([]*MetaItem, error) {
	var metaItems []*MetaItem
	for _, param := range params {
		// 基础参数校验
		if err := validateMetaFilterParam(knowledgeID, param); err != nil {
			return nil, err
		}
		// 转换参数值
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
	// 根据类型转换value
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

// 校验元数据过滤参数
func validateMetaFilterParam(knowledgeID string, param *rag_service.RagMetaFilterItem) error {
	// 检查关键参数是否为空
	if param.Key == "" || param.Type == "" || param.Condition == "" {
		errMsg := "key/type/condition cannot be empty"
		log.Errorf("kbId: %s, %s", knowledgeID, errMsg)
		return errors.New(errMsg)
	}

	// 检查空条件与值的匹配性
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

// buildRerankId 构造重排序模型id
func buildRerankId(priorityType int32, rerankId string) string {
	if priorityType == 1 {
		return ""
	}
	return rerankId
}

// buildRetrieveMethod 构造检索方式
func buildRetrieveMethod(matchType string) string {
	switch matchType {
	case "vector":
		return "semantic_search" // 向量检索
	case "text":
		return "full_text_search" // 全文检索
	case "mix":
		return "hybrid_search" // 混合检索
	}
	return ""
}

// buildRerankMod 构造重排序模式
func buildRerankMod(priorityType int32) string {
	if priorityType == 1 {
		return "weighted_score"
	}
	return "rerank_model"
}

// buildWeight 构造权重信息
func buildWeight(knowConfig model.KnowledgeBaseConfig) *WeightParams {
	if knowConfig.PriorityMatch != 1 {
		return nil
	}
	return &WeightParams{
		VectorWeight: float32(knowConfig.SemanticsPriority),
		TextWeight:   float32(knowConfig.KeywordPriority),
	}
}
