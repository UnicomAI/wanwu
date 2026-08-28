package params_process

import (
	"encoding/json"
	"fmt"
	"strconv"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
)

const (
	metaTypeNumber          = "number"
	metaTypeTime            = "time"
	externalKnowledge int32 = 1 //外部知识库
)

type KnowledgeParams struct {
	UserId               string                        `json:"userId"`          // 用户id
	KnowledgeIdList      []string                      `json:"knowledgeIdList"` // 知识库id列表
	Question             string                        `json:"question"`
	Threshold            float32                       `json:"threshold"` // Score阈值
	TopK                 int32                         `json:"topK"`
	Stream               bool                          `json:"stream"`
	Chichat              bool                          `json:"chichat"` // 当知识库召回结果为空时是否使用默认话术（兜底），默认为true
	RerankModelId        string                        `json:"rerank_model_id"`
	CustomModelInfo      *CustomModelInfo              `json:"custom_model_info"`
	MaxHistory           int32                         `json:"max_history"`
	RewriteQuery         bool                          `json:"rewrite_query"`   // 是否query改写
	RerankMod            string                        `json:"rerank_mod"`      // rerank_model:重排序模式，weighted_score：权重搜索
	RetrieveMethod       string                        `json:"retrieve_method"` // hybrid_search:混合搜索， semantic_search:向量搜索， full_text_search：文本搜索
	Weight               *config.WeightParams          `json:"weights"`         // 权重搜索下的权重配置
	Temperature          float32                       `json:"temperature,omitempty"`
	TopP                 float32                       `json:"top_p,omitempty"`               // 多样性
	RepetitionPenalty    float32                       `json:"repetition_penalty,omitempty"`  // 重复惩罚/频率惩罚
	ReturnMeta           bool                          `json:"return_meta,omitempty"`         // 是否返回元数据
	AutoCitation         bool                          `json:"auto_citation"`                 // 是否自动角标
	TermWeight           float32                       `json:"term_weight_coefficient"`       // 关键词系数
	MetaFilter           bool                          `json:"metadata_filtering"`            // 元数据过滤开关
	MetaFilterConditions []*config.MetadataFilterParam `json:"metadata_filtering_conditions"` // 元数据过滤条件
	UseGraph             bool                          `json:"use_graph"`                     // 是否启动知识图谱查询
}
type CustomModelInfo struct {
	LlmModelID string `json:"llm_model_id"`
}

type KnowledgeIdParams struct {
	KnowledgeBaseIds []string `json:"knowledgeBaseIds"` // 知识库信息
}

// RAGKnowledgeBaseConfig 知识库配置结构体
type RAGKnowledgeBaseConfig struct {
	KnowledgeBaseIds     []string                `json:"knowledgeBaseIds"`     // 知识库信息
	MaxHistory           int32                   `json:"maxHistory"`           // 最长上下文
	Threshold            float32                 `json:"threshold"`            // 过滤阈值
	TopK                 int32                   `json:"topK"`                 // topK
	MatchType            string                  `json:"matchType"`            // 检索类型：vector（向量检索）、text（文本检索）、mix（混合检索）
	KeywordPriority      float32                 `json:"keywordPriority"`      // 关键词权重
	PriorityMatch        int32                   `json:"priorityMatch"`        // 权重匹配，仅混合检索模式下有效，1 表示启用
	SemanticsPriority    float32                 `json:"semanticsPriority"`    // 语义权重
	TermWeight           float32                 `json:"termWeight"`           // 关键词系数, 默认为1
	TermWeightEnable     bool                    `json:"termWeightEnable"`     // 关键词系数开关
	AppKnowledgeBaseList []*AppKnowledgeBaseInfo `json:"AppKnowledgeBaseList"` // 知识库元数据
	UseGraph             bool                    `json:"useGraph"`             // 知识图谱开关
}

type AppKnowledgeBaseInfo struct {
	KnowledgeBaseId      string                `json:"knowledgeBaseId"`
	KnowledgeBaseName    string                `json:"knowledgeBaseName"`
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
}

type MetaDataFilterParams struct {
	FilterEnable     bool                `json:"filterEnable"`     // 元数据过滤开关
	FilterLogicType  string              `json:"filterLogicType"`  // 元数据逻辑条件：and/or
	MetaFilterParams []*MetaFilterParams `json:"metaFilterParams"` // 元数据过滤参数列表
}

type MetaFilterParams struct {
	Condition string `json:"condition"`
	Key       string `json:"key"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

func init() {
	AddServiceContainer(&KnowledgeProcess{})
}

type KnowledgeProcess struct {
}

func (k *KnowledgeProcess) ServiceType() ServiceType {
	return KnowledgeType
}

func (k *KnowledgeProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	if len(agent.Assistant.KnowledgebaseConfig) > 0 {
		knowledgeParams := &KnowledgeIdParams{}
		if err := json.Unmarshal([]byte(agent.Assistant.KnowledgebaseConfig), knowledgeParams); err != nil {
			err = fmt.Errorf("Assistant服务解析智能体知识库配置失败，assistantId: %d, error: %v, knowledgebaseConfigRaw: %s", agent.Assistant.ID, err, agent.Assistant.KnowledgebaseConfig)
			return err
		}
		knowledgeInfoList, err := clientInfo.Knowledge.SelectKnowledgeDetailByIdList(userQueryParams.Ctx, &knowledgebase_service.KnowledgeDetailSelectListReq{
			KnowledgeIds: knowledgeParams.KnowledgeBaseIds,
		})
		if err != nil {
			err = fmt.Errorf("Assistant服务获取知识库详情失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
			return err
		}
		prepareParams.KnowledgeList = knowledgeInfoList.List
	}
	return nil
}
func (k *KnowledgeProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	knowledgeBaseConfig := &RAGKnowledgeBaseConfig{}
	if err := json.Unmarshal([]byte(assistant.Assistant.KnowledgebaseConfig), knowledgeBaseConfig); err != nil {
		return fmt.Errorf("Assistant服务解析智能体知识库配置失败，assistantId: %d, error: %v, knowledgebaseConfigRaw: %s", assistant.Assistant.ID, err, assistant.Assistant.KnowledgebaseConfig)
	}
	knowledgeList := prepareParams.KnowledgeList
	if len(knowledgeList) > 0 {
		knowledgeIDToName := make(map[string]string)
		var allExternalKnow = true
		for _, v := range knowledgeList {
			if _, exists := knowledgeIDToName[v.KnowledgeId]; !exists {
				knowledgeIDToName[v.KnowledgeId] = v.RagName
			}
			if v.External != externalKnowledge {
				allExternalKnow = false
			}
		}
		params, err := buildMetaDataFilterParams(knowledgeBaseConfig.AppKnowledgeBaseList, knowledgeIDToName)
		if err != nil {
			log.Errorf("Assistant buildMetaDataFilterParams, err: %v", err)
			return err
		}
		rerankEndpoint, err := buildRerank(knowledgeBaseConfig, assistant.Assistant, allExternalKnow)
		if err != nil {
			return err
		}

		var autoCitation = buildAutoCitation(assistant.AssistantSnapshot)
		knowledgeParams := &KnowledgeParams{
			UserId:               assistant.Assistant.UserId,
			KnowledgeIdList:      knowledgeBaseConfig.KnowledgeBaseIds,
			Threshold:            knowledgeBaseConfig.Threshold,
			TopK:                 knowledgeBaseConfig.TopK,
			Stream:               true,
			RerankModelId:        toString(rerankEndpoint["model_id"]),
			MaxHistory:           knowledgeBaseConfig.MaxHistory,
			RewriteQuery:         true,
			RerankMod:            buildRerankMod(knowledgeBaseConfig.PriorityMatch),
			RetrieveMethod:       buildRetrieveMethod(knowledgeBaseConfig.MatchType),
			Weight:               buildWeight(knowledgeBaseConfig),
			TermWeight:           buildTermWeight(knowledgeBaseConfig),
			MetaFilter:           len(params) > 0,
			MetaFilterConditions: params,
			UseGraph:             knowledgeBaseConfig.UseGraph,
			AutoCitation:         autoCitation,
		}
		marshal, err := json.Marshal(knowledgeParams)
		if err != nil {
			return err
		}
		agentChatParams.KnowledgeParams = string(marshal)
	}
	return nil
}

// buildMetaDataFilterParams 构造元数据过滤参数
func buildMetaDataFilterParams(knowledgeInfos []*AppKnowledgeBaseInfo, knowledgeIDToName map[string]string) ([]*config.MetadataFilterParam, error) {
	if len(knowledgeInfos) == 0 {
		return nil, nil
	}
	var ragMetaDataFilterParams []*config.MetadataFilterParam
	for _, k := range knowledgeInfos {
		if k.MetaDataFilterParams == nil || !k.MetaDataFilterParams.FilterEnable ||
			len(k.MetaDataFilterParams.MetaFilterParams) == 0 {
			continue
		}
		item, err := buildMetadataFilterItem(k.MetaDataFilterParams.MetaFilterParams)
		if err != nil {
			log.Errorf("buildMetaDataFilterParams error %v", err)
			return nil, err
		}
		ragMetaDataFilterParams = append(ragMetaDataFilterParams, &config.MetadataFilterParam{
			FilterKnowledgeName: knowledgeIDToName[k.KnowledgeBaseId],
			LogicalOperator:     k.MetaDataFilterParams.FilterLogicType,
			MetaList:            item,
		})
	}
	return ragMetaDataFilterParams, nil
}

func buildRerank(knowledgeBaseConfig *RAGKnowledgeBaseConfig, assistant *model.Assistant, allExternalKnow bool) (map[string]interface{}, error) {
	var rerankEndpoint map[string]interface{}
	if allExternalKnow { //全是外部知识库则不校验
		return rerankEndpoint, nil
	}
	if knowledgeBaseConfig.PriorityMatch != 1 {
		rerankConfig := &common.AppModelConfig{}
		if assistant.RerankConfig != "" {
			if err := json.Unmarshal([]byte(assistant.RerankConfig), rerankConfig); err != nil {
				log.Errorf("Assistant服务解析智能体rerank配置失败，assistantId: %d, error: %v, rerankConfigRaw: %s", assistant.ID, err, assistant.RerankConfig)
				return nil, err
			}
			if rerankConfig.Model == "" || rerankConfig.ModelId == "" {
				log.Errorf("Assistant服务缺少rerank配置，assistantId: %d", assistant.ID)
				return nil, fmt.Errorf("智能体缺少rerank配置")
			}
		}
		rerankEndpoint = mp.ToModelEndpoint(rerankConfig.ModelId, rerankConfig.Model)
	}
	return rerankEndpoint, nil
}

// buildAutoCitation 构造自动引文
func buildAutoCitation(assistant *model.AssistantSnapshot) bool {
	var autoCitation = true
	if assistant != nil && assistant.SnapshotExtra != "" {
		var snapshotExtra = model.AssistantSnapshotExtra{}
		if err := json.Unmarshal([]byte(assistant.SnapshotExtra), &snapshotExtra); err != nil {
			log.Errorf("Assistant buildMetaDataFilterParams, err: %v", err)
		}
		autoCitation = snapshotExtra.HideKnowledge != 1
	}
	return autoCitation
}
func buildMetadataFilterItem(metaFilterParams []*MetaFilterParams) ([]*config.MetadataFilterItem, error) {
	var ragMetaDataFilterItem []*config.MetadataFilterItem
	for _, k := range metaFilterParams {
		data, err := buildValueData(k.Type, k.Value, k.Condition)
		if err != nil {
			log.Errorf("buildMetadataFilterItem error %v", err)
			return nil, err
		}
		ragMetaDataFilterItem = append(ragMetaDataFilterItem, &config.MetadataFilterItem{
			ComparisonOperator: k.Condition,
			MetaName:           k.Key,
			MetaType:           k.Type,
			Value:              data,
		})
	}
	return ragMetaDataFilterItem, nil
}

func buildValueData(valueType string, value string, condition string) (interface{}, error) {
	if condition == "empty" || condition == "not empty" {
		return nil, nil
	}
	switch valueType {
	case metaTypeNumber:
	case metaTypeTime:
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}

// buildRetrieveMethod 构造检索方式
func buildRetrieveMethod(matchType string) string {
	switch matchType {
	case "vector":
		return "semantic_search"
	case "text":
		return "full_text_search"
	case "mix":
		return "hybrid_search"
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

// buildTermWeight 构造关键词系数
func buildTermWeight(knowConfig *RAGKnowledgeBaseConfig) float32 {
	if knowConfig.TermWeightEnable {
		return knowConfig.TermWeight
	}
	return 0.0
}

// buildWeight 构造权重信息
func buildWeight(knowConfig *RAGKnowledgeBaseConfig) *config.WeightParams {
	if knowConfig.PriorityMatch != 1 {
		return nil
	}
	return &config.WeightParams{
		VectorWeight: knowConfig.SemanticsPriority,
		TextWeight:   knowConfig.KeywordPriority,
	}
}

func toString(data interface{}) string {
	if data != nil {
		return data.(string)
	}
	return ""
}
