package client

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/model-service/client/model"
)

type IClient interface {
	ImportModel(ctx context.Context, req *model.ModelImported) *errs.Status
	UpdateModel(ctx context.Context, req *model.ModelImported) *errs.Status
	DeleteModel(ctx context.Context, req *model.ModelImported) *errs.Status
	ChangeModelStatus(ctx context.Context, req *model.ModelImported) *errs.Status
	GetModel(ctx context.Context, req *model.ModelImported) (*model.ModelImported, *errs.Status)
	GetModelByUUID(ctx context.Context, uuid string) (*model.ModelImported, *errs.Status)
	ListModelsByUuids(ctx context.Context, uuids []string) ([]*model.ModelImported, *errs.Status)
	ListModelsByIds(ctx context.Context, modelIds []uint32) ([]*model.ModelImported, *errs.Status)
	ListModels(ctx context.Context, req *model.ModelImported) ([]*model.ModelImported, *errs.Status)
	ListModelsInStatisticScope(ctx context.Context, orgIds, userIds []string, modelType string) ([]*model.ModelImported, *errs.Status)
	AdminModelPageList(ctx context.Context, name, provider string, modelTypes, userIds, orgIds []string, scopeTypes []uint32, pageNum, pageSize int) ([]*model.ModelImported, int64, *errs.Status)
	ListTypeModels(ctx context.Context, req *model.ModelImported) ([]*model.ModelImported, *errs.Status)

	// --- model experience ---
	SaveModelExperienceDialog(ctx context.Context, dialog *model.ModelExperienceDialog) (*model.ModelExperienceDialog, *errs.Status)
	GetModelExperienceDialog(ctx context.Context, userId, orgId string, modelExperienceId uint32) (*model.ModelExperienceDialog, *errs.Status)
	ListModelExperienceDialogs(ctx context.Context, userId, orgId string) ([]*model.ModelExperienceDialog, *errs.Status)
	DeleteModelExperienceDialog(ctx context.Context, userId, orgId string, modelExperienceId uint32) *errs.Status

	SaveModelExperienceDialogRecord(ctx context.Context, record *model.ModelExperienceDialogRecord) *errs.Status
	ListModelExperienceDialogRecords(ctx context.Context, userId, orgId string, modelExperienceId uint32, sessionId string) ([]*model.ModelExperienceDialogRecord, *errs.Status)
}
