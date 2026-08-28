package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GenAppKey(ctx context.Context, req *app_service.GenAppKeyReq) (*app_service.AppKeyInfo, error) {
	appKey, err := s.cli.GenAppKey(ctx, req.UserId, req.OrgId, req.AppId, req.AppType, util.GenApiUUID())
	if err != nil {
		return nil, errStatus(errs.Code_Appkey, err)
	}
	return toProtoAppKey(appKey), nil
}

func (s *Service) GetAppKeyList(ctx context.Context, req *app_service.GetAppKeyListReq) (*app_service.AppKeyInfoList, error) {
	appKeyList, err := s.cli.GetAppKeyList(ctx, req.UserId, req.OrgId, req.AppId, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_Appkey, err)
	}
	ret := &app_service.AppKeyInfoList{
		Total: int64(len(appKeyList)),
	}
	for _, appKey := range appKeyList {
		ret.Info = append(ret.Info, toProtoAppKey(appKey))
	}
	return ret, nil
}

func (s *Service) DelAppKey(ctx context.Context, req *app_service.DelAppKeyReq) (*emptypb.Empty, error) {
	err := s.cli.DelAppKey(ctx, util.MustU32(req.AppKeyId), req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_Appkey, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetAppKeyByKey(ctx context.Context, req *app_service.GetAppKeyByKeyReq) (*app_service.AppKeyInfo, error) {
	appKey, err := s.cli.GetAppKeyByKey(ctx, req.AppKey)
	if err != nil {
		return nil, errStatus(errs.Code_Appkey, err)
	}
	return toProtoAppKey(appKey), nil
}

// --- internal ---

func toProtoAppKey(appKey *model.ApiKey) *app_service.AppKeyInfo {
	return &app_service.AppKeyInfo{
		AppKeyId:  util.Int2Str(appKey.ID),
		AppKey:    appKey.ApiKey,
		UserId:    appKey.UserID,
		OrgId:     appKey.OrgID,
		AppId:     appKey.AppID,
		AppType:   appKey.AppType,
		CreatedAt: appKey.CreatedAt,
	}
}
