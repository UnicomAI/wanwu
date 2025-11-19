package service

import (
	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

func CreateRag(ctx *gin.Context, userId, orgId string, req request.AppBriefConfig) (*request.RagReq, error) {
	resp, err := rag.CreateRag(ctx.Request.Context(), &rag_service.CreateRagReq{
		AppBrief: appBriefConfigModel2Proto(req),
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &request.RagReq{
		RagID: resp.RagId,
	}, err
}

func UpdateRag(ctx *gin.Context, req request.RagBrief, userId, orgId string) error {
	_, err := rag.UpdateRag(ctx.Request.Context(), &rag_service.UpdateRagReq{
		RagId:    req.RagID,
		AppBrief: appBriefConfigModel2Proto(req.AppBriefConfig),
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func UpdateRagConfig(ctx *gin.Context, req request.RagConfig) error {
	modelConfig, err := appModelConfigModel2Proto(req.ModelConfig)
	if err != nil {
		return err
	}
	rerankConfig, err := appModelConfigModel2Proto(req.RerankConfig)
	if err != nil {
		return err
	}
	_, err = rag.UpdateRagConfig(ctx.Request.Context(), &rag_service.UpdateRagConfigReq{
		RagId:               req.RagID,
		ModelConfig:         modelConfig,
		RerankConfig:        rerankConfig,
		KnowledgeBaseConfig: ragKBConfigToProto(req.KnowledgeBaseConfig),
		SensitiveConfig:     ragSensitiveConfigToProto(req.SafetyConfig),
	})
	return err
}

func ragSensitiveConfigToProto(req request.AppSafetyConfig) *rag_service.RagSensitiveConfig {
	var sensitiveTableIds []string
	for _, v := range req.Tables {
		sensitiveTableIds = append(sensitiveTableIds, v.TableId)
	}
	sensitiveConfig := &rag_service.RagSensitiveConfig{
		Enable:   req.Enable,
		TableIds: sensitiveTableIds,
	}
	return sensitiveConfig
}

func ragKBConfigToProto(knowledgeConfig request.AppKnowledgebaseConfig) *rag_service.RagKnowledgeBaseConfig {
	result := &rag_service.RagKnowledgeBaseConfig{
		PerKnowledgeConfigs: make([]*rag_service.RagPerKnowledgeConfig, 0, len(knowledgeConfig.Knowledgebases)),
	}
	for _, knowledge := range knowledgeConfig.Knowledgebases {
		// Initialize a single knowledge base configuration
		perConfig := &rag_service.RagPerKnowledgeConfig{
			KnowledgeId: knowledge.ID,
			GraphSwitch: knowledge.GraphSwitch,
		}
		// Build metadata filters (if enabled)
		if metaFilter := buildRagMetaFilter(knowledge.MetaDataFilterParams); metaFilter != nil {
			perConfig.RagMetaFilter = metaFilter
		}
		// A single knowledge base configuration is added to the result
		result.PerKnowledgeConfigs = append(result.PerKnowledgeConfigs, perConfig)
	}
	result.GlobalConfig = buildRagGlobalConfig(knowledgeConfig.Config)
	return result
}

// Build metadata filters for a single knowledge base
func buildRagMetaFilter(params *request.MetaDataFilterParams) *rag_service.RagMetaFilter {
	// Check whether the filtering parameters are valid (return nil if they are not enabled or have no specific conditions)
	if params == nil || params.MetaFilterParams == nil || len(params.MetaFilterParams) == 0 {
		return nil
	}
	// Convert filter items
	filterItems := make([]*rag_service.RagMetaFilterItem, 0, len(params.MetaFilterParams))
	for _, metaParam := range params.MetaFilterParams {
		filterItems = append(filterItems, &rag_service.RagMetaFilterItem{
			Key:       metaParam.Key,
			Type:      metaParam.Type,
			Value:     metaParam.Value,
			Condition: metaParam.Condition,
		})
	}
	return &rag_service.RagMetaFilter{
		FilterEnable:    params.FilterEnable,
		FilterLogicType: params.FilterLogicType,
		FilterItems:     filterItems,
	}
}

func buildRagGlobalConfig(kbConfig request.AppKnowledgebaseParams) *rag_service.RagGlobalConfig {
	return &rag_service.RagGlobalConfig{
		MaxHistory:        kbConfig.MaxHistory,
		Threshold:         kbConfig.Threshold,
		TopK:              kbConfig.TopK,
		MatchType:         kbConfig.MatchType,
		KeywordPriority:   kbConfig.KeywordPriority,
		PriorityMatch:     kbConfig.PriorityMatch,
		SemanticsPriority: kbConfig.SemanticsPriority,
		TermWeight:        kbConfig.TermWeight,
		TermWeightEnable:  kbConfig.TermWeightEnable,
		UseGraph:          kbConfig.UseGraph,
	}
}

func DeleteRag(ctx *gin.Context, req request.RagReq) error {
	_, err := rag.DeleteRag(ctx.Request.Context(), &rag_service.RagDeleteReq{
		RagId: req.RagID,
	})
	return err
}

func GetRag(ctx *gin.Context, req request.RagReq) (*response.RagInfo, error) {
	resp, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{RagId: req.RagID})
	if err != nil {
		return nil, err
	}
	modelConfig, rerankConfig, err := appModelRerankProto2Model(ctx, resp)
	if err != nil {
		log.Errorf("ragId: %v gets config fail: %v", req.RagID, err.Error())
	}
	ragInfo := &response.RagInfo{
		RagID:               resp.RagId,
		AppBriefConfig:      appBriefConfigProto2Model(ctx, resp.BriefConfig, constant.AppTypeRag),
		ModelConfig:         modelConfig,
		RerankConfig:        rerankConfig,
		KnowledgeBaseConfig: ragKBConfigProto2Model(ctx, resp.KnowledgeBaseConfig),
		SafetyConfig:        ragSafetyConfigProto2Model(ctx, resp.SensitiveConfig),
	}

	return ragInfo, nil
}

func appModelRerankProto2Model(ctx *gin.Context, resp *rag_service.RagInfo) (request.AppModelConfig, request.AppModelConfig, error) {
	var modelConfig, rerankConfig request.AppModelConfig
	if resp.ModelConfig.ModelId != "" {
		modelInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: resp.ModelConfig.ModelId})
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, err
		}
		modelConfig, err = appModelConfigProto2Model(resp.ModelConfig, modelInfo.DisplayName)
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, err
		}
	}
	if resp.RerankConfig.ModelId != "" {
		rerankInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: resp.RerankConfig.ModelId})
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, err
		}
		rerankConfig, err = appModelConfigProto2Model(resp.RerankConfig, rerankInfo.DisplayName)
		if err != nil {
			return request.AppModelConfig{}, request.AppModelConfig{}, err
		}
	}
	return modelConfig, rerankConfig, nil
}

func ragSafetyConfigProto2Model(ctx *gin.Context, sensitiveCfg *rag_service.RagSensitiveConfig) request.AppSafetyConfig {
	var sensitiveTableList []request.SensitiveTable
	if len(sensitiveCfg.GetTableIds()) != 0 {
		sensitiveWordTable, _ := safety.GetSensitiveWordTableListByIDs(ctx, &safety_service.GetSensitiveWordTableListByIDsReq{TableIds: sensitiveCfg.GetTableIds()})
		if sensitiveWordTable != nil {
			for _, table := range sensitiveWordTable.List {
				sensitiveTableList = append(sensitiveTableList, request.SensitiveTable{
					TableId:   table.TableId,
					TableName: table.TableName,
				})
			}
		}
	}
	safetyConfig := request.AppSafetyConfig{
		Enable: sensitiveCfg.Enable,
		Tables: sensitiveTableList,
	}
	return safetyConfig
}

func ragKBConfigProto2Model(ctx *gin.Context, kbConfig *rag_service.RagKnowledgeBaseConfig) request.AppKnowledgebaseConfig {
	if kbConfig == nil {
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
			Config:         request.AppKnowledgebaseParams{},
		}
	}
	knowledgeList := make([]request.AppKnowledgeBase, 0, len(kbConfig.PerKnowledgeConfigs))

	// Convert individual configurations for each knowledge base
	for _, perConfig := range kbConfig.PerKnowledgeConfigs {
		kbInfo, err := knowledgeBase.SelectKnowledgeDetailById(ctx, &knowledgeBase_service.KnowledgeDetailSelectReq{
			KnowledgeId: perConfig.KnowledgeId,
		})
		if err != nil {
			log.Errorf("select knowledge detail error: %v", err)
			return request.AppKnowledgebaseConfig{
				Knowledgebases: make([]request.AppKnowledgeBase, 0),
				Config:         request.AppKnowledgebaseParams{},
			}
		}
		// Basic information mapping
		knowledge := request.AppKnowledgeBase{
			ID:          perConfig.KnowledgeId,
			Name:        kbInfo.Name,
			GraphSwitch: kbInfo.GraphSwitch,
		}
		// Convert metadata filtering configuration
		metaFilter := perConfig.RagMetaFilter
		knowledge.MetaDataFilterParams = convertRagMetaFilterToParams(metaFilter)

		knowledgeList = append(knowledgeList, knowledge)
	}
	globalConfig := kbConfig.GlobalConfig
	if globalConfig == nil {
		globalConfig = &rag_service.RagGlobalConfig{}
	}
	appConfig := request.AppKnowledgebaseParams{
		MaxHistory:        globalConfig.MaxHistory,
		Threshold:         globalConfig.Threshold,
		TopK:              globalConfig.TopK,
		MatchType:         globalConfig.MatchType,
		KeywordPriority:   globalConfig.KeywordPriority,
		PriorityMatch:     globalConfig.PriorityMatch,
		SemanticsPriority: globalConfig.SemanticsPriority,
		TermWeight:        globalConfig.TermWeight,
		TermWeightEnable:  globalConfig.TermWeightEnable,
		UseGraph:          globalConfig.UseGraph,
	}
	return request.AppKnowledgebaseConfig{
		Knowledgebases: knowledgeList,
		Config:         appConfig,
	}
}
func convertRagMetaFilterToParams(metaFilter *rag_service.RagMetaFilter) *request.MetaDataFilterParams {
	if metaFilter == nil {
		return nil
	}
	// Convert filter items
	filterParams := make([]*request.MetaFilterParams, 0, len(metaFilter.FilterItems))
	for _, item := range metaFilter.FilterItems {
		filterParams = append(filterParams, &request.MetaFilterParams{
			Key:       item.Key,
			Type:      item.Type,
			Value:     item.Value,
			Condition: item.Condition,
		})
	}
	return &request.MetaDataFilterParams{
		FilterEnable:     metaFilter.FilterEnable,
		FilterLogicType:  metaFilter.FilterLogicType,
		MetaFilterParams: filterParams, // Map filter list
	}
}

func CopyRag(ctx *gin.Context, userId, orgId string, req request.RagReq) (*request.RagReq, error) {
	resp, err := rag.CopyRag(ctx.Request.Context(), &rag_service.CopyRagReq{
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		RagId: req.RagID,
	})
	if err != nil {
		return nil, err
	}
	return &request.RagReq{
		RagID: resp.RagId,
	}, err
}
