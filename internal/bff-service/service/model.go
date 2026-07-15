package service

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type ModelInfoOptions struct {
	UserId string
	OrgId  string
}

func DefaultModelInfoOptions() *ModelInfoOptions {
	return &ModelInfoOptions{
		UserId: "",
		OrgId:  "",
	}
}

func ImportModel(ctx *gin.Context, userId, orgId string, req *request.ImportOrUpdateModelRequest) error {
	if err := util.ValidateBriefCreate(&req.DisplayName, &req.ModelDesc, util.SubjectModel); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	clientReq, err := parseImportAndUpdateClientReq(userId, orgId, req)
	if err != nil {
		return err
	}
	if err = ValidateModel(ctx, clientReq); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("An error occurred during model import validation: Invalid model: %v, err : %v", clientReq.Model, err))
	}
	_, err = model.ImportModel(ctx.Request.Context(), clientReq)
	if err != nil {
		return err
	}
	return nil
}

func UpdateModel(ctx *gin.Context, userId, orgId string, req *request.ImportOrUpdateModelRequest) error {
	if req.ModelId == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "modelId cannot be empty")
	}
	existingModel, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}
	if err := util.ValidateBriefUpdate(&req.DisplayName, existingModel.DisplayName, &req.ModelDesc, existingModel.ModelDesc, util.SubjectModel); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	clientReq, err := parseImportAndUpdateClientReq(userId, orgId, req)
	if err != nil {
		return err
	}
	if err = ValidateModel(ctx, clientReq); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("An error occurred during model update validation: Invalid model: %v, err : %v", clientReq.Model, err))
	}
	_, err = model.UpdateModel(ctx.Request.Context(), clientReq)
	if err != nil {
		return err
	}
	return nil
}

func DeleteModel(ctx *gin.Context, userId, orgId string, req *request.DeleteModelRequest) error {
	_, err := model.DeleteModel(ctx.Request.Context(), &model_service.DeleteModelReq{
		ModelId: req.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetModel(ctx *gin.Context, userId, orgId string, req *request.GetModelRequest) (*response.ModelInfo, error) {
	resp, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return nil, err
	}
	return toModelInfo(ctx, resp, &ModelInfoOptions{UserId: userId, OrgId: orgId})
}

func GetModelById(ctx *gin.Context, req *request.GetModelRequest) (*response.ModelInfo, error) {
	return GetModel(ctx, "", "", req)
}

func ListModels(ctx *gin.Context, userId, orgId string, req *request.ListModelsRequest) (*response.ListResult, error) {
	resp, err := model.ListModels(ctx.Request.Context(), &model_service.ListModelsReq{
		Provider:    req.Provider,
		ModelType:   req.ModelType,
		DisplayName: req.DisplayName,
		IsActive:    req.IsActive,
		UserId:      userId,
		OrgId:       orgId,
		FilterScope: req.FilterScope,
		ScopeType:   req.ScopeType,
	})
	if err != nil {
		return nil, err
	}
	list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userId, OrgId: orgId})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  list,
		Total: resp.Total,
	}, nil
}

func ChangeModelStatus(ctx *gin.Context, userId, orgId string, req *request.ModelStatusRequest) error {
	_, err := model.ChangeModelStatus(ctx.Request.Context(), &model_service.ModelStatusReq{
		ModelId:  req.ModelId,
		IsActive: req.IsActive,
		UserId:   userId,
		OrgId:    orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func ListTypeModels(ctx *gin.Context, userId, orgId string, req *request.ListTypeModelsRequest) (*response.ListResult, error) {
	resp, err := model.ListTypeModels(ctx.Request.Context(), &model_service.ListTypeModelsReq{
		ModelType: req.ModelType,
		UserId:    userId,
		OrgId:     orgId,
	})
	if err != nil {
		return nil, err
	}
	list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userId, OrgId: orgId})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  list,
		Total: resp.Total,
	}, nil
}

func CheckModelUserPermission(ctx *gin.Context, userId, orgId string, modelIds []string) ([]string, error) {
	resp, err := model.ListModelsByIds(ctx.Request.Context(), &model_service.ListModelsByIdsReq{ModelIds: modelIds})
	if err != nil {
		return nil, err
	}
	// 创建模型ID到模型信息的映射
	modelMap := make(map[string]*model_service.ModelInfo)
	for _, model := range resp.Models {
		modelMap[model.ModelId] = model
	}
	// 校验所有传入的modelIds，收集有权限的模型ID
	var authorizedModelIds []string
	var unauthorizedModelId string
	for _, modelId := range modelIds {
		model, exists := modelMap[modelId]
		if !exists {
			// 模型不存在
			unauthorizedModelId = modelId
			continue
		}
		// 校验模型权限
		var hasPermission bool
		switch model.GetScopeType() {
		case config.ModelScopeTypePrivate: // 私有
			hasPermission = (model.UserId == userId) && (model.OrgId == orgId)
		case config.ModelScopeTypePublic: // 公开
			hasPermission = true // 公开模型，任何人都可以访问
		case config.ModelScopeTypeOrg: // 指定组织可见
			hasPermission = (model.OrgId == orgId)
		default:
			hasPermission = (model.UserId == userId) && (model.OrgId == orgId)
		}
		if hasPermission {
			authorizedModelIds = append(authorizedModelIds, modelId)
		} else {
			unauthorizedModelId = modelId
		}
	}

	if unauthorizedModelId != "" {
		return authorizedModelIds, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_model_perm", unauthorizedModelId)
	}
	return authorizedModelIds, nil
}

func GetModelIdByUuid(ctx *gin.Context, uuid string) (string, error) {
	resp, err := model.GetModelByUuid(ctx.Request.Context(), &model_service.GetModelByUuidReq{Uuid: uuid})
	if err != nil {
		return "", err
	}
	return resp.ModelId, nil
}

func GetModelUuidById(ctx *gin.Context, modelId string) (string, error) {
	resp, err := GetModelById(ctx, &request.GetModelRequest{BaseModelRequest: request.BaseModelRequest{ModelId: modelId}})
	if err != nil {
		return "", err
	}
	return resp.Uuid, nil
}

func ListModelIdsByUuids(ctx *gin.Context, uuids []string) ([]string, error) {
	resp, err := model.ListModelsByUuids(ctx.Request.Context(), &model_service.ListModelsByUuidsReq{Uuids: uuids})
	if err != nil {
		return nil, err
	}
	modelIds := make([]string, 0, len(resp.Models))
	for _, model := range resp.Models {
		modelIds = append(modelIds, model.ModelId)
	}
	return modelIds, nil
}

// --- internal ---

func parseImportAndUpdateClientReq(userId, orgId string, req *request.ImportOrUpdateModelRequest) (*model_service.ModelInfo, error) {
	if req.ScopeType == config.ModelScopeTypePublic {
		if userId != config.SystemAdminUserID || orgId != config.TopOrgID {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "Only system administrators can make the model public")
		}
	}
	importSource := req.ImportSource
	if importSource == "" {
		importSource = "external"
	}
	clientReq := &model_service.ModelInfo{
		Provider:      req.Provider,
		ModelId:       req.ModelId,
		ModelType:     req.ModelType,
		Model:         req.Model,
		DisplayName:   req.DisplayName,
		ModelIconPath: req.Avatar.Key,
		PublishDate:   req.PublishDate,
		UserId:        userId,
		OrgId:         orgId,
		IsActive:      true,
		ModelDesc:     req.ModelDesc,
		ScopeType:     req.ScopeType,
		ImportSource:  importSource,
	}
	configStr, err := req.ConfigString()
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	clientReq.ProviderConfig = configStr
	return clientReq, nil
}

func toModelInfos(ctx *gin.Context, models []*model_service.ModelInfo, opts ...*ModelInfoOptions) ([]*response.ModelInfo, error) {
	var ret []*response.ModelInfo
	for _, m := range models {
		info, err := toModelInfo(ctx, m, opts...)
		if err != nil {
			return nil, err
		}
		ret = append(ret, info)
	}
	return ret, nil
}

func toModelInfo(ctx *gin.Context, modelInfo *model_service.ModelInfo, opts ...*ModelInfoOptions) (*response.ModelInfo, error) {
	modelConfig, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v get model config err: %v", modelInfo.ModelId, err))
	}

	// 获取模型全量标签（独立函数）
	allModelTags, err := getModelAllTags(modelInfo)
	if err != nil {
		return nil, err
	}

	// 判断模型是否支持编辑
	option := DefaultModelInfoOptions()
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].UserId != "" {
			option.UserId = opts[0].UserId
		}
		if opts[0].OrgId != "" {
			option.OrgId = opts[0].OrgId
		}
	}
	allowEdit := checkAllowEdit(modelInfo, option)

	// 不支持编辑，则不展示apiKey
	if !allowEdit && modelConfig != nil {
		cfg := make(map[string]any)
		if b, err := json.Marshal(modelConfig); err == nil {
			if err = json.Unmarshal(b, &cfg); err == nil {
				cfg["apiKey"] = "it-is-not-your-api-key"
				modelConfig = cfg
			}
		}
	}

	res := &response.ModelInfo{
		ModelId:      modelInfo.ModelId,
		Uuid:         modelInfo.Uuid,
		Provider:     modelInfo.Provider,
		Model:        modelInfo.Model,
		ModelType:    modelInfo.ModelType,
		DisplayName:  modelInfo.DisplayName,
		Avatar:       cacheModelAvatar(ctx, modelInfo.ModelIconPath),
		PublishDate:  modelInfo.PublishDate,
		IsActive:     modelInfo.IsActive,
		UserId:       modelInfo.UserId,
		OrgId:        modelInfo.OrgId,
		CreatedAt:    util.Time2Str(modelInfo.CreatedAt),
		UpdatedAt:    util.Time2Str(modelInfo.UpdatedAt),
		ModelDesc:    modelInfo.ModelDesc,
		Config:       modelConfig,
		Tags:         allModelTags,
		ScopeType:    modelInfo.ScopeType,
		AllowEdit:    allowEdit,
		ImportSource: modelInfo.ImportSource,
	}
	return res, nil
}

// checkAllowEdit 检查用户是否有编辑权限
// 编辑权限要求：创建者本人 + 组织匹配（公开模型要求在顶层组织）
func checkAllowEdit(modelInfo *model_service.ModelInfo, option *ModelInfoOptions) bool {
	// 首先检查是否是创建者
	if modelInfo.UserId != option.UserId {
		return false
	}

	// 根据模型范围类型检查组织
	switch modelInfo.ScopeType {
	case config.ModelScopeTypePrivate:
		// 私有模型：创建者 + 组织匹配
		return modelInfo.OrgId == option.OrgId
	case config.ModelScopeTypePublic:
		// 公开模型：创建者 + 必须在顶层组织操作
		return option.OrgId == config.TopOrgID
	case config.ModelScopeTypeOrg:
		// 组织模型：创建者 + 组织匹配
		return modelInfo.OrgId == option.OrgId
	default:
		return false
	}
}

// getModelAllTags 获取模型所有标签
func getModelAllTags(modelInfo *model_service.ModelInfo) ([]mp_common.Tag, error) {
	var allModelTags []mp_common.Tag

	// - 公开范围标签：如私有/公开/内部测试等
	scopeTags := GetTagsByScopeType(modelInfo.ScopeType)
	allModelTags = append(allModelTags, scopeTags...)

	// - 导入来源标签：如外部URL/内置模型等
	sourceTags := GetTagsByImportSource(modelInfo.ImportSource)
	allModelTags = append(allModelTags, sourceTags...)

	// - 基础属性标签：如模型类型(LLM/rerank)、最大token数、推理能力等
	baseTags, err := mp.ToModelTags(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get model base tags, modelId: %v, err: %v", modelInfo.ModelId, err)
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, errMsg)
	}
	allModelTags = append(allModelTags, baseTags...)

	return allModelTags, nil
}

func GetRecommendModels(_ *gin.Context, req *request.RecommendModelsRequest) (*response.ListResult, error) {
	recommendModels := config.Cfg().RecommendModels

	var matchedProvider *config.RecommendModelsByProvider
	for i := range recommendModels {
		if recommendModels[i].Provider == req.Provider {
			matchedProvider = &recommendModels[i]
			break
		}
	}

	if matchedProvider == nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, fmt.Sprintf("provider %s not found in config", req.Provider))
	}

	validModelTypes := []string{
		mp.ModelTypeLLM,
		mp.ModelTypeTextEmbedding,
		mp.ModelTypeMultiEmbedding,
		mp.ModelTypeTextRerank,
		mp.ModelTypeMultiRerank,
		mp.ModelTypeOcr,
		mp.ModelTypeGui,
		mp.ModelTypePdfParser,
		mp.ModelTypeSyncAsr,
	}
	validModelType := false
	for _, t := range validModelTypes {
		if req.ModelType == t {
			validModelType = true
			break
		}
	}
	if !validModelType {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, fmt.Sprintf("model type %s not found in config", req.ModelType))
	}

	var models []response.RecommendModel
	if req.ModelType == mp.ModelTypeLLM {
		models = convertRecommendModelsLLM(matchedProvider.Llm)
	} else {
		models = getModelsByType(*matchedProvider, req.ModelType)
	}

	return &response.ListResult{
		List:  models,
		Total: int64(len(models)),
	}, nil
}

func getModelsByType(p config.RecommendModelsByProvider, modelType string) []response.RecommendModel {
	modelMap := map[string][]config.RecommendModelItem{
		mp.ModelTypeTextEmbedding:  p.Embedding,
		mp.ModelTypeMultiEmbedding: p.MultiEmbedding,
		mp.ModelTypeTextRerank:     p.Rerank,
		mp.ModelTypeMultiRerank:    p.MultiRerank,
		mp.ModelTypeOcr:            p.Ocr,
		mp.ModelTypeGui:            p.Gui,
		mp.ModelTypePdfParser:      p.PdfParser,
		mp.ModelTypeSyncAsr:        p.SyncAsr,
	}
	if items, ok := modelMap[modelType]; ok {
		return convertRecommendModels(items)
	}
	return nil
}

func convertRecommendModels(items []config.RecommendModelItem) []response.RecommendModel {
	result := make([]response.RecommendModel, 0, len(items))
	for _, item := range items {
		result = append(result, response.RecommendModel{
			Model:       item.Model,
			DisplayName: item.DisplayName,
		})
	}
	return result
}

func convertRecommendModelsLLM(items []config.RecommendModelItemLLM) []response.RecommendModel {
	result := make([]response.RecommendModel, 0, len(items))
	for _, item := range items {
		tags := make([]response.ModelTagItem, 0, len(item.Tags))
		hasVision := false
		hasToolCall := false
		hasThinking := false

		for _, t := range item.Tags {
			tags = append(tags, response.ModelTagItem{Text: t})
			if t == mp_common.TagVisionSupport {
				hasVision = true
			}
			if t == mp_common.TagToolCall {
				hasToolCall = true
			}
			if t == mp_common.TagThinkingSupport {
				hasThinking = true
			}
		}

		result = append(result, response.RecommendModel{
			Model:           item.Model,
			DisplayName:     item.DisplayName,
			Tags:            tags,
			VisionSupport:   mapSupport(hasVision),
			FunctionCalling: mapToolCall(hasToolCall),
			ThinkingSupport: mapSupport(hasThinking),
		})
	}
	return result
}

func mapSupport(has bool) string {
	if has {
		return "support"
	}
	return "noSupport"
}

func mapToolCall(has bool) string {
	if has {
		return "toolCall"
	}
	return "noSupport"
}

var (
	modelTypeNameMap = map[string]string{
		mp.ModelTypeLLM:            mp.MTNameLLM,
		mp.ModelTypeTextEmbedding:  mp.MTNameTextEmbedding,
		mp.ModelTypeMultiEmbedding: mp.MTNameMultiEmbedding,
		mp.ModelTypeTextRerank:     mp.MTNameTextRerank,
		mp.ModelTypeMultiRerank:    mp.MTNameMultiRerank,
		mp.ModelTypeOcr:            mp.MTNameOcr,
		mp.ModelTypeGui:            mp.MTNameGui,
		mp.ModelTypePdfParser:      mp.MTNamePdfParser,
		mp.ModelTypeSyncAsr:        mp.MTNameSyncAsr,
	}
	providerNameMap = map[string]string{
		mp.ProviderOpenAICompatible: mp.PNameOpenAICompatible,
		mp.ProviderYuanJing:         mp.PNameYuanJing,
		mp.ProviderHuoshan:          mp.PNameHuoshan,
		mp.ProviderOllama:           mp.PNameOllama,
		mp.ProviderQwen:             mp.PNameQwen,
		mp.ProviderInfini:           mp.PNameInfini,
		mp.ProviderQianfan:          mp.PNameQianfan,
		mp.ProviderDeepSeek:         mp.PNameDeepSeek,
		mp.ProviderJina:             mp.PNameJina,
		mp.ProviderZhipu:            mp.PNameZhipu,
	}
)

func ListImportProviders(ctx *gin.Context, req *request.ListImportProvidersRequest) (*response.ListResult, error) {
	recommendModels := config.Cfg().RecommendModels

	var list []*response.ProviderModelTypes
	for _, p := range recommendModels {
		if req.Provider != "" {
			matched := false
			for providerKey, displayName := range providerNameMap {
				if strings.Contains(displayName, req.Provider) && p.Provider == providerKey {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		var modelTypes []string
		if len(p.Llm) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeLLM)
		}
		if len(p.Embedding) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeTextEmbedding)
		}
		if len(p.MultiEmbedding) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeMultiEmbedding)
		}
		if len(p.Rerank) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeTextRerank)
		}
		if len(p.MultiRerank) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeMultiRerank)
		}
		if len(p.Ocr) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeOcr)
		}
		if len(p.Gui) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeGui)
		}
		if len(p.PdfParser) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypePdfParser)
		}
		if len(p.SyncAsr) > 0 {
			modelTypes = append(modelTypes, mp.ModelTypeSyncAsr)
		}

		if len(modelTypes) == 0 {
			continue
		}

		if req.ModelType != "" && !slices.Contains(modelTypes, req.ModelType) {
			continue
		}

		children := make([]*response.ModelTypeItem, 0, len(modelTypes))
		for _, mt := range modelTypes {
			name := modelTypeNameMap[mt]
			if name == "" {
				name = mt
			}
			children = append(children, &response.ModelTypeItem{
				Key:  mt,
				Name: name,
			})
		}

		providerName := providerNameMap[p.Provider]
		if providerName == "" {
			providerName = p.Provider
		}

		list = append(list, &response.ProviderModelTypes{
			Key:      p.Provider,
			Name:     providerName,
			Children: children,
		})
	}

	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}
