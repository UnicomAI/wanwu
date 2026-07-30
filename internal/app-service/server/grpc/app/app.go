package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetExplorationAppList(ctx context.Context, req *app_service.GetExplorationAppListReq) (*app_service.ExplorationAppList, error) {
	appList, err := s.cli.GetExplorationAppList(ctx, req.UserId, req.OrgId, req.Name, req.AppType, req.SearchType)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	ret := &app_service.ExplorationAppList{
		Total: int64(len(appList)),
	}
	for _, app := range appList {
		ret.Infos = append(ret.Infos, toProtoExpApp(app))
	}
	return ret, nil
}

func (s *Service) ChangeExplorationAppFavorite(ctx context.Context, req *app_service.ChangeExplorationAppFavoriteReq) (*emptypb.Empty, error) {
	err := s.cli.ChangeExplorationAppFavorite(ctx, req.UserId, req.OrgId, req.AppId, req.AppType, req.IsFavorite)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) PublishApp(ctx context.Context, req *app_service.PublishAppReq) (*emptypb.Empty, error) {
	err := s.cli.PublishApp(ctx, req.UserId, req.OrgId, req.AppId, req.AppType, req.PublishType)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UnPublishApp(ctx context.Context, req *app_service.UnPublishAppReq) (*emptypb.Empty, error) {
	err := s.cli.UnPublishApp(ctx, req.AppId, req.AppType, req.UserId)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetAppList(ctx context.Context, req *app_service.GetAppListReq) (*app_service.AppList, error) {
	publishAppList, err := s.cli.GetAppList(ctx, req.OrgIds, req.UserIds, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	ret := &app_service.AppList{
		Total: int64(len(publishAppList)),
	}
	for _, publishApp := range publishAppList {
		ret.Infos = append(ret.Infos, toProtoApp(publishApp))
	}
	return ret, nil
}

func (s *Service) DeleteApp(ctx context.Context, req *app_service.DeleteAppReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteApp(ctx, req.AppId, req.AppType, req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_AppGeneral, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetAppListByIds(ctx context.Context, req *app_service.GetAppListByIdsReq) (*app_service.AppList, error) {
	publishAppList, err := s.cli.GetAppListByIds(ctx, req.AppIdsList, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppExploration, err)
	}
	ret := &app_service.AppList{
		Total: int64(len(publishAppList)),
	}
	for _, publishApp := range publishAppList {
		ret.Infos = append(ret.Infos, toProtoApp(publishApp))
	}
	return ret, nil
}

func (s *Service) GetAppInfo(ctx context.Context, req *app_service.GetAppInfoReq) (*app_service.AppInfo, error) {
	publishApp, err := s.cli.GetAppInfo(ctx, req.AppId, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppGeneral, err)
	}
	return toProtoApp(publishApp), nil
}

func (s *Service) RecordAppHistory(ctx context.Context, req *app_service.RecordAppHistoryReq) (*emptypb.Empty, error) {
	err := s.cli.RecordAppHistory(ctx, req.UserId, req.AppId, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppGeneral, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ConvertAppType(ctx context.Context, req *app_service.ConvertAppTypeReq) (*emptypb.Empty, error) {
	err := s.cli.ConvertAppType(ctx, req.AppId, req.OldAppType, req.NewAppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppGeneral, err)
	}
	return &emptypb.Empty{}, nil
}

// --- internal ---
func toProtoExpApp(record *orm.ExplorationAppInfo) *app_service.ExplorationAppInfo {
	return &app_service.ExplorationAppInfo{
		UserId:      record.UserID,
		OrgId:       record.OrgID,
		AppId:       record.AppId,
		AppType:     record.AppType,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		PublishType: record.PublishType,
		IsFavorite:  record.IsFavorite,
	}
}

func toProtoApp(record *model.App) *app_service.AppInfo {
	return &app_service.AppInfo{
		UserId:      record.UserID,
		OrgId:       record.OrgID,
		AppId:       record.AppID,
		AppType:     record.AppType,
		CreatedAt:   record.CreatedAt,
		PublishType: record.PublishType,
	}
}
