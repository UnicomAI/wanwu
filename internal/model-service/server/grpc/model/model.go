package model

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) ImportModel(ctx context.Context, req *model_service.ModelInfo) (*emptypb.Empty, error) {
	if err := s.cli.ImportModel(ctx, &model.ModelImported{
		UUID:           util.NewID(),
		Provider:       req.Provider,
		ModelType:      req.ModelType,
		Model:          req.Model,
		DisplayName:    req.DisplayName,
		ModelIconPath:  req.ModelIconPath,
		IsActive:       req.IsActive,
		PublishDate:    req.PublishDate,
		ProviderConfig: db.LongText(req.ProviderConfig),
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
		ModelDesc: db.LongText(req.ModelDesc),
	}); err != nil {
		return nil, errStatus(errs.Code_ModelImportedModel, err)
	}
	return nil, nil
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
		ProviderConfig: db.LongText(req.ProviderConfig),
		ModelDesc:      db.LongText(req.ModelDesc),
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
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

func (s *Service) GetModelByIds(ctx context.Context, req *model_service.GetModelByIdsReq) (*model_service.ModelInfos, error) {
	var modelIDs []uint32
	for _, modelID := range req.ModelIds {
		modelIDs = append(modelIDs, util.MustU32(modelID))
	}
	modelInfos, err := s.cli.GetModelByIds(ctx, modelIDs)
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

func (s *Service) ListModels(ctx context.Context, req *model_service.ListModelsReq) (*model_service.ModelInfos, error) {
	modelInfos, err := s.cli.ListModels(ctx, &model.ModelImported{
		Provider:    req.Provider,
		ModelType:   req.ModelType,
		IsActive:    req.IsActive,
		DisplayName: req.DisplayName,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_ModelListModels, err)
	}
	return toModelInfos(modelInfos), nil
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
	return &model_service.ModelInfo{
		ModelId:        util.Int2Str(modelInfo.ID),
		Uuid:           modelInfo.UUID,
		Provider:       modelInfo.Provider,
		ModelType:      modelInfo.ModelType,
		Model:          modelInfo.Model,
		DisplayName:    modelInfo.DisplayName,
		ModelIconPath:  modelInfo.ModelIconPath,
		IsActive:       modelInfo.IsActive,
		PublishDate:    modelInfo.PublishDate,
		ProviderConfig: string(modelInfo.ProviderConfig),
		UserId:         modelInfo.UserID,
		OrgId:          modelInfo.OrgID,
		CreatedAt:      modelInfo.CreatedAt,
		UpdatedAt:      modelInfo.UpdatedAt,
		ModelDesc:      string(modelInfo.ModelDesc),
	}
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
