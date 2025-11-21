package app

import (
	"context"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AppUrlCreate creates the agent Url
func (s *Service) AppUrlCreate(ctx context.Context, req *app_service.AppUrlCreateReq) (*emptypb.Empty, error) {
	// Assemble model parameters
	appUrl := &model.AppUrl{
		AppID:               req.AppUrlInfo.AppId,
		AppType:             req.AppUrlInfo.AppType,
		Name:                req.AppUrlInfo.Name,
		ExpiredAt:           req.AppUrlInfo.ExpiredAt,
		Copyright:           req.AppUrlInfo.Copyright,
		CopyrightEnable:     req.AppUrlInfo.CopyrightEnable,
		PrivacyPolicy:       req.AppUrlInfo.PrivacyPolicy,
		PrivacyPolicyEnable: req.AppUrlInfo.PrivacyPolicyEnable,
		Disclaimer:          req.AppUrlInfo.Disclaimer,
		DisclaimerEnable:    req.AppUrlInfo.DisclaimerEnable,
		UserId:              req.AppUrlInfo.UserId,
		OrgId:               req.AppUrlInfo.OrgId,
		Suffix:              util.GenUUID(),
		Description:         req.AppUrlInfo.Description,
	}
	// Call the client method to create the agent Url
	if err := s.cli.CreateAppUrl(ctx, appUrl); err != nil {
		return nil, errStatus(errs.Code_AppUrl, err)
	}
	return &emptypb.Empty{}, nil
}

// AppUrlDelete deletes the agent Url
func (s *Service) AppUrlDelete(ctx context.Context, req *app_service.AppUrlDeleteReq) (*emptypb.Empty, error) {
	// Call the client method to delete the agent Url
	if status := s.cli.DeleteAppUrl(ctx, util.MustU32(req.UrlId)); status != nil {
		return nil, errStatus(errs.Code_AppUrl, status)
	}
	return &emptypb.Empty{}, nil
}

// AppUrlUpdate Modifies the agent Url
func (s *Service) AppUrlUpdate(ctx context.Context, req *app_service.AppUrlUpdateReq) (*emptypb.Empty, error) {
	// Call the client method to update the agent Url
	if status := s.cli.UpdateAppUrl(ctx, &model.AppUrl{
		ID:                  util.MustU32(req.AppUrlInfo.UrlId),
		Name:                req.AppUrlInfo.Name,
		ExpiredAt:           req.AppUrlInfo.ExpiredAt,
		Copyright:           req.AppUrlInfo.Copyright,
		CopyrightEnable:     req.AppUrlInfo.CopyrightEnable,
		PrivacyPolicy:       req.AppUrlInfo.PrivacyPolicy,
		PrivacyPolicyEnable: req.AppUrlInfo.PrivacyPolicyEnable,
		Disclaimer:          req.AppUrlInfo.Disclaimer,
		DisclaimerEnable:    req.AppUrlInfo.DisclaimerEnable,
		Description:         req.AppUrlInfo.Description,
	}); status != nil {
		return nil, errStatus(errs.Code_AppUrl, status)
	}
	return &emptypb.Empty{}, nil
}

// GetAppUrlList Agent Url list
func (s *Service) GetAppUrlList(ctx context.Context, req *app_service.GetAppUrlListReq) (*app_service.GetAppUrlListResp, error) {
	appUrls, err := s.cli.GetAppUrlList(ctx, req.AppId, req.AppType)
	if err != nil {
		return nil, errStatus(errs.Code_AppUrl, err)
	}
	// Convert to responsive format
	var existingAppUrls []*app_service.AppUrlInfo
	for _, appUrl := range appUrls {
		existingAppUrls = append(existingAppUrls, appUrl2pb(appUrl))
	}
	return &app_service.GetAppUrlListResp{
		AppUrlInfos: existingAppUrls,
	}, nil
}

func (s *Service) GetAppUrlInfoBySuffix(ctx context.Context, req *app_service.GetAppUrlInfoBySuffixReq) (*app_service.AppUrlInfo, error) {
	// Get existing agent Url information
	appUrl, err := s.cli.GetAppUrlInfoBySuffix(ctx, req.Suffix)
	if err != nil {
		return nil, errStatus(errs.Code_AppUrl, err)
	}
	return appUrl2pb(appUrl), nil
}

// AppUrlStatusSwitch AppUrl switch
func (s *Service) AppUrlStatusSwitch(ctx context.Context, req *app_service.AppUrlStatusSwitchReq) (*emptypb.Empty, error) {
	if err := s.cli.AppUrlStatusSwitch(ctx, util.MustU32(req.UrlId), req.Status); err != nil {
		return nil, errStatus(errs.Code_AppUrl, err)
	}
	return &emptypb.Empty{}, nil
}

// --- internal ---
func appUrl2pb(appUrl *model.AppUrl) *app_service.AppUrlInfo {
	return &app_service.AppUrlInfo{
		UrlId:               util.Int2Str(appUrl.ID),
		AppId:               appUrl.AppID,
		AppType:             appUrl.AppType,
		Name:                appUrl.Name,
		CreatedAt:           appUrl.CreatedAt,
		ExpiredAt:           appUrl.ExpiredAt,
		Copyright:           appUrl.Copyright,
		CopyrightEnable:     appUrl.CopyrightEnable,
		PrivacyPolicy:       appUrl.PrivacyPolicy,
		PrivacyPolicyEnable: appUrl.PrivacyPolicyEnable,
		Disclaimer:          appUrl.Disclaimer,
		DisclaimerEnable:    appUrl.DisclaimerEnable,
		Suffix:              appUrl.Suffix,
		Status:              appUrl.Status,
		OrgId:               appUrl.OrgId,
		UserId:              appUrl.UserId,
		Description:         appUrl.Description,
	}
}
