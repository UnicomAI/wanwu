package model

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/internal/model-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/model-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ImportModel 导入模型，返回落库后的模型信息（含新生成的 modelId）。
// 返回值供调用方做后续关联操作（如消息中心的「模型已上线」跳转需要 modelId）。
func (s *Service) ImportModel(ctx context.Context, req *model_service.ModelInfo) (*model_service.ModelInfo, error) {
	imported := &model.ModelImported{
		UUID:           util.NewID(),
		Provider:       req.Provider,
		ModelType:      req.ModelType,
		Model:          req.Model,
		DisplayName:    req.DisplayName,
		ModelIconPath:  req.ModelIconPath,
		IsActive:       req.IsActive,
		PublishDate:    req.PublishDate,
		ProviderConfig: req.ProviderConfig,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
		ModelDesc:    req.ModelDesc,
		ScopeType:    util.MustU32(req.ScopeType),
		ImportSource: req.ImportSource,
	}
	// db.Create 会把自增主键回填进 imported，因此这里能直接拿到新模型 ID
	if err := s.cli.ImportModel(ctx, imported); err != nil {
		return nil, errStatus(errs.Code_ModelImportedModel, err)
	}
	return toModelInfo(imported), nil
}

func (s *Service) UpdateModel(ctx context.Context, req *model_service.ModelInfo) (*emptypb.Empty, error) {
	if err := s.cli.UpdateModel(ctx, &model.ModelImported{
		ID:             util.MustU32(req.ModelId),
		Provider:       req.Provider,
		ModelType:      req.ModelType,
		Model:          req.Model,
		DisplayName:    req.DisplayName,
		ModelIconPath:  req.ModelIconPath,
		PublishDate:    req.PublishDate,
		ProviderConfig: req.ProviderConfig,
		ModelDesc:      req.ModelDesc,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
		ScopeType: util.MustU32(req.ScopeType),
	}); err != nil {
		return nil, errStatus(errs.Code_ModelUpdateModel, err)
	}
	return nil, nil
}

func (s *Service) DeleteModel(ctx context.Context, req *model_service.DeleteModelReq) (*emptypb.Empty, error) {
	if err := s.cli.DeleteModel(ctx, &model.ModelImported{
		ID: util.MustU32(req.ModelId),
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	}); err != nil {
		return nil, errStatus(errs.Code_ModelDeleteModel, err)
	}
	return nil, nil
}

func (s *Service) ChangeModelStatus(ctx context.Context, req *model_service.ModelStatusReq) (*emptypb.Empty, error) {
	if err := s.cli.ChangeModelStatus(ctx, &model.ModelImported{
		ID:       util.MustU32(req.ModelId),
		IsActive: req.IsActive,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	}); err != nil {
		return nil, errStatus(errs.Code_ModelChangeModelStatus, err)
	}
	return nil, nil
}

func (s *Service) ListModelsByIds(ctx context.Context, req *model_service.ListModelsByIdsReq) (*model_service.ModelInfos, error) {
	var modelIDs []uint32
	for _, modelID := range req.ModelIds {
		modelIDs = append(modelIDs, util.MustU32(modelID))
	}
	modelInfos, err := s.cli.ListModelsByIds(ctx, modelIDs)
	if err != nil {
		return nil, errStatus(errs.Code_ModelGetModelByIds, err)
	}
	return toModelInfos(modelInfos), nil
}

func (s *Service) GetModel(ctx context.Context, req *model_service.GetModelReq) (*model_service.ModelInfo, error) {
	modelInfo, err := s.cli.GetModel(ctx, &model.ModelImported{
		ID: util.MustU32(req.ModelId),
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_ModelGetModel, err)
	}
	return toModelInfo(modelInfo), nil
}

func (s *Service) GetModelByUuid(ctx context.Context, req *model_service.GetModelByUuidReq) (*model_service.ModelInfo, error) {
	modelInfo, err := s.cli.GetModelByUUID(ctx, req.Uuid)
	if err != nil {
		return nil, errStatus(errs.Code_ModelGetModelByUUID, err)
	}
	return toModelInfo(modelInfo), nil
}

func (s *Service) ListModelsByUuids(ctx context.Context, req *model_service.ListModelsByUuidsReq) (*model_service.ModelInfos, error) {
	modelInfos, err := s.cli.ListModelsByUuids(ctx, req.Uuids)
	if err != nil {
		return nil, errStatus(errs.Code_ModelListModelsByUuids, err)
	}
	return toModelInfos(modelInfos), nil
}

func (s *Service) ListModels(ctx context.Context, req *model_service.ListModelsReq) (*model_service.ModelInfos, error) {
	modelInfos, err := s.cli.ListModels(ctx, &model.ModelImported{
		Provider:    req.Provider,
		ModelType:   req.ModelType,
		IsActive:    req.IsActive,
		DisplayName: req.DisplayName,
		ScopeType:   util.MustU32(req.ScopeType),
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_ModelListModels, err)
	}
	// 筛选公有模型/我的模型
	modelsInfoFiltered := make([]*model.ModelImported, 0, len(modelInfos))

	switch req.FilterScope {
	case config.ScopeTypeStr_PUBLIC:
		for _, modelInfo := range modelInfos {
			scopeTypeInt := int(modelInfo.ScopeType)
			if scopeTypeInt == sqlopt.ModelScopeTypePublic ||
				(scopeTypeInt == sqlopt.ModelScopeTypeOrg && modelInfo.UserID != req.UserId) {
				modelsInfoFiltered = append(modelsInfoFiltered, modelInfo)
			}
		}
	case config.ScopeTypeStr_PRIVATE:
		for _, modelInfo := range modelInfos {
			scopeTypeInt := int(modelInfo.ScopeType)
			if scopeTypeInt == sqlopt.ModelScopeTypePrivate ||
				(scopeTypeInt == sqlopt.ModelScopeTypeOrg && modelInfo.UserID == req.UserId) {
				modelsInfoFiltered = append(modelsInfoFiltered, modelInfo)
			}
		}
	default:
		modelsInfoFiltered = modelInfos
	}

	return toModelInfos(modelsInfoFiltered), nil
}

func (s *Service) ListModelsInStatisticScope(ctx context.Context, req *model_service.ListModelsInStatisticScopeReq) (*model_service.ModelInfos, error) {
	modelInfos, err := s.cli.ListModelsInStatisticScope(ctx, req.OrgIds, req.UserIds, req.ModelType)
	if err != nil {
		return nil, errStatus(errs.Code_ModelListModels, err)
	}
	return toModelInfos(modelInfos), nil
}

// AdminModelPageList 管理员中心模型全局分页列表（跨用户/组织）
func (s *Service) AdminModelPageList(ctx context.Context, req *model_service.AdminModelPageListReq) (*model_service.AdminModelPageListResp, error) {
	modelInfos, total, err := s.cli.AdminModelPageList(ctx, req.Name, req.Provider, req.ModelType, req.UserId, req.OrgId, req.ScopeType, int(req.PageNum), int(req.PageSize))
	if err != nil {
		return nil, errStatus(errs.Code_ModelListModels, err)
	}
	var models []*model_service.ModelInfo
	for _, modelInfo := range modelInfos {
		models = append(models, toModelInfo(modelInfo))
	}
	return &model_service.AdminModelPageListResp{
		Total:  total,
		Models: models,
	}, nil
}

func (s *Service) ListTypeModels(ctx context.Context, req *model_service.ListTypeModelsReq) (*model_service.ModelInfos, error) {
	modelInfos, err := s.cli.ListTypeModels(ctx, &model.ModelImported{
		ModelType: req.ModelType,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_ModelListTypeModels, err)
	}
	return toModelInfos(modelInfos), nil
}

func toModelInfo(modelInfo *model.ModelImported) *model_service.ModelInfo {
	modelInfoRet := &model_service.ModelInfo{
		ModelId:        util.Int2Str(modelInfo.ID),
		Uuid:           modelInfo.UUID,
		Provider:       modelInfo.Provider,
		ModelType:      modelInfo.ModelType,
		Model:          modelInfo.Model,
		DisplayName:    modelInfo.DisplayName,
		ModelIconPath:  modelInfo.ModelIconPath,
		IsActive:       modelInfo.IsActive,
		PublishDate:    modelInfo.PublishDate,
		ProviderConfig: modelInfo.ProviderConfig,
		UserId:         modelInfo.UserID,
		OrgId:          modelInfo.OrgID,
		CreatedAt:      modelInfo.CreatedAt,
		UpdatedAt:      modelInfo.UpdatedAt,
		ModelDesc:      modelInfo.ModelDesc,
		ScopeType:      util.Int2Str(modelInfo.ScopeType),
		ImportSource:   modelInfo.ImportSource,
	}
	if modelInfo.DisplayName == "" {
		modelInfoRet.DisplayName = modelInfo.Model
	}
	return modelInfoRet
}

func toModelInfos(modelInfos []*model.ModelImported) *model_service.ModelInfos {
	var res []*model_service.ModelInfo
	for _, modelInfo := range modelInfos {
		res = append(res, toModelInfo(modelInfo))
	}
	return &model_service.ModelInfos{
		Models: res,
		Total:  int64(len(modelInfos)),
	}
}
