package app

import (
	"context"
	"strconv"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreateApiKey(ctx context.Context, req *app_service.CreateApiKeyReq) (*app_service.ApiKeyInfo, error) {
	apiKey, err := s.cli.CreateApiKey(ctx, req.UserId, req.OrgId, req.Name, req.Desc, req.ExpiredAt, util.GenApiUUID())
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	return toProtoApiKey(apiKey), nil
}

func (s *Service) DeleteApiKey(ctx context.Context, req *app_service.DeleteApiKeyReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteApiKey(ctx, util.MustU32(req.KeyId), req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateApiKey(ctx context.Context, req *app_service.UpdateApiKeyReq) (*emptypb.Empty, error) {
	err := s.cli.UpdateApiKey(ctx, util.MustU32(req.KeyId), req.UserId, req.OrgId, req.Name, req.Desc, req.ExpiredAt)
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ListApiKeys(ctx context.Context, req *app_service.ListApiKeysReq) (*app_service.ApiKeyInfoList, error) {
	apiKeyList, count, err := s.cli.ListApiKeys(ctx, req.OrgIds, req.UserIds, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	ret := &app_service.ApiKeyInfoList{
		Total: int32(count),
	}
	for _, apiKey := range apiKeyList {
		ret.Items = append(ret.Items, toProtoApiKey(apiKey))
	}
	return ret, nil
}

func (s *Service) UpdateApiKeyStatus(ctx context.Context, req *app_service.UpdateApiKeyStatusReq) (*emptypb.Empty, error) {
	err := s.cli.UpdateApiKeyStatus(ctx, util.MustU32(req.KeyId), req.Status)
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetApiKeyByKey(ctx context.Context, req *app_service.GetApiKeyByKeyReq) (*app_service.ApiKeyInfo, error) {
	apiKey, err := s.cli.GetApiKeyByKey(ctx, req.ApiKey)
	if err != nil {
		return nil, errStatus(errs.Code_ApiKey, err)
	}
	return toProtoApiKey(apiKey), nil
}

// --- internal ---
func toProtoApiKey(apiKey *model.OpenApiKey) *app_service.ApiKeyInfo {
	return &app_service.ApiKeyInfo{
		KeyId:     strconv.Itoa(int(apiKey.ID)),
		Key:       apiKey.Key,
		UserId:    apiKey.UserID,
		Name:      apiKey.Name,
		Desc:      apiKey.Description,
		ExpiredAt: apiKey.ExpiredAt,
		CreatedAt: apiKey.CreatedAt,
		Status:    apiKey.Status,
		OrgId:     apiKey.OrgID,
	}
}
