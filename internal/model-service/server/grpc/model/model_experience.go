package model

import (
	"context"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) SaveModelExperienceDialog(ctx context.Context, req *model_service.SaveModelExperienceDialogReq) (*model_service.ModelExperienceDialog, error) {
	dialog, err := s.cli.SaveModelExperienceDialog(ctx, &model.ModelExperienceDialog{
		SessionId:    req.SessionId,
		ModelId:      req.ModelId,
		Title:        req.Title,
		ModelSetting: req.ModelSetting,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	})
	if err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialog, err)
	}
	return toModelExperienceDialog(dialog), nil
}

func (s *Service) GetModelExperienceDialog(ctx context.Context, req *model_service.ModelExperienceDialogReq) (*model_service.ModelExperienceDialog, error) {
	dialog, err := s.cli.GetModelExperienceDialog(ctx, req.UserId, req.OrgId, util.MustU32(req.ModelExperienceId))
	if err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialog, err)
	}
	return toModelExperienceDialog(dialog), nil
}

func (s *Service) GetModelExperienceDialogs(ctx context.Context, req *model_service.ListModelExperienceDialogReq) (*model_service.ModelExperienceDialogs, error) {
	dialogs, err := s.cli.ListModelExperienceDialogs(ctx, req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialog, err)
	}
	var ret []*model_service.ModelExperienceDialog
	for _, dialog := range dialogs {
		ret = append(ret, toModelExperienceDialog(dialog))

	}
	return &model_service.ModelExperienceDialogs{
		Dialogs: ret,
	}, nil
}

func (s *Service) DeleteModelExperienceDialog(ctx context.Context, req *model_service.ModelExperienceDialogReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteModelExperienceDialog(ctx, req.UserId, req.OrgId, util.MustU32(req.ModelExperienceId))
	if err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialog, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) SaveModelExperienceDialogRecord(ctx context.Context, req *model_service.SaveModelExperienceDialogRecordReq) (*emptypb.Empty, error) {
	if err := s.cli.SaveModelExperienceDialogRecord(ctx, &model.ModelExperienceDialogRecord{
		ModelExperienceID: util.MustU32(req.ModelExperienceId),
		SessionId:         req.SessionId,
		ModelId:           req.ModelId,
		OriginalContent:   req.OriginalContent,
		HandledContent:    req.HandledContent,
		ReasoningContent:  req.ReasoningContent,
		Role:              req.Role,
		FileInfo:          req.FileInfo,
		PublicModel: model.PublicModel{
			OrgID:  req.OrgId,
			UserID: req.UserId,
		},
	}); err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialogRecord, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetModelExperienceDialogRecords(ctx context.Context, req *model_service.GetModelExperienceDialogRecordsReq) (*model_service.ModelExperienceDialogRecords, error) {
	records, err := s.cli.ListModelExperienceDialogRecords(ctx, req.UserId, req.OrgId, util.MustU32(req.ModelExperienceId), req.SessionId)
	if err != nil {
		return nil, errStatus(errs.Code_ModelExperienceDialogRecord, err)
	}
	var ret []*model_service.ModelExperienceDialogRecord
	for _, record := range records {
		ret = append(ret, toModelExperienceDialogRecord(record))
	}
	return &model_service.ModelExperienceDialogRecords{
		Records: ret,
	}, nil
}

func toModelExperienceDialog(dialog *model.ModelExperienceDialog) *model_service.ModelExperienceDialog {
	return &model_service.ModelExperienceDialog{
		ModelExperienceId: strconv.Itoa(int(dialog.ID)),
		ModelId:           dialog.ModelId,
		SessionId:         dialog.SessionId,
		Title:             dialog.Title,
		ModelSetting:      dialog.ModelSetting,
		CreatedAt:         dialog.CreatedAt,
		UpdatedAt:         dialog.UpdatedAt,
	}
}

func toModelExperienceDialogRecord(record *model.ModelExperienceDialogRecord) *model_service.ModelExperienceDialogRecord {
	return &model_service.ModelExperienceDialogRecord{
		ModelExperienceId: util.Int2Str(record.ModelExperienceID),
		ModelId:           record.ModelId,
		SessionId:         record.SessionId,
		OriginalContent:   record.OriginalContent,
		HandledContent:    record.HandledContent,
		ReasoningContent:  record.ReasoningContent,
		Role:              record.Role,
		FileInfo:          record.FileInfo,
	}
}
