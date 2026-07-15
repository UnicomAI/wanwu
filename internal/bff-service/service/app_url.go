package service

import (
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

func AppUrlCreate(ctx *gin.Context, userId, orgId string, req request.AppUrlCreateRequest) error {
	if req.AppType != constant.AppTypeAgent {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "app url create app type mismatch")
	}
	var expiredAt int64
	var err error
	if req.ExpiredAt != "" {
		expiredAt, err = util.Str2Time(req.ExpiredAt)
		if err != nil {
			return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
		}
	}
	_, err = app.AppUrlCreate(ctx.Request.Context(), &app_service.AppUrlCreateReq{
		AppUrlInfo: &app_service.AppUrlInfo{
			AppId:               req.AppId,
			AppType:             req.AppType,
			Name:                req.Name,
			ExpiredAt:           expiredAt,
			Copyright:           req.Copyright,
			CopyrightEnable:     req.CopyrightEnable,
			PrivacyPolicy:       req.PrivacyPolicy,
			PrivacyPolicyEnable: req.PrivacyPolicyEnable,
			Disclaimer:          req.Disclaimer,
			DisclaimerEnable:    req.DisclaimerEnable,
			UserId:              userId,
			OrgId:               orgId,
			Description:         req.Description,
		},
	})
	return err
}

func AppUrlDelete(ctx *gin.Context, req request.AppUrlIdRequest) error {
	_, err := app.AppUrlDelete(ctx.Request.Context(), &app_service.AppUrlDeleteReq{
		UrlId: req.UrlId,
	})
	return err
}

func AppUrlUpdate(ctx *gin.Context, req request.AppUrlUpdateRequest) error {
	var expiredAt int64
	var err error
	if req.ExpiredAt != "" {
		expiredAt, err = util.Str2Time(req.ExpiredAt)
		if err != nil {
			return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
		}
	}
	_, err = app.AppUrlUpdate(ctx.Request.Context(), &app_service.AppUrlUpdateReq{
		AppUrlInfo: &app_service.AppUrlInfo{
			UrlId:               req.UrlId,
			Name:                req.Name,
			ExpiredAt:           expiredAt,
			Copyright:           req.Copyright,
			CopyrightEnable:     req.CopyrightEnable,
			PrivacyPolicy:       req.PrivacyPolicy,
			PrivacyPolicyEnable: req.PrivacyPolicyEnable,
			Disclaimer:          req.Disclaimer,
			DisclaimerEnable:    req.DisclaimerEnable,
			Description:         req.Description,
		},
	})
	return err
}

func GetAppUrlList(ctx *gin.Context, req request.AppUrlListRequest) ([]*response.AppUrlInfo, error) {
	if req.AppType != constant.AppTypeAgent {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "get app url list app type mismatch")
	}
	resp, err := app.GetAppUrlList(ctx.Request.Context(), &app_service.GetAppUrlListReq{
		AppId:   req.AppId,
		AppType: req.AppType,
	})
	if err != nil {
		return nil, err
	}
	return transAppUrlListResp2Model(resp), nil
}

func AppUrlStatusSwitch(ctx *gin.Context, req request.AppUrlStatusRequest) error {
	_, err := app.AppUrlStatusSwitch(ctx.Request.Context(), &app_service.AppUrlStatusSwitchReq{
		UrlId:  req.UrlId,
		Status: req.Status,
	})
	return err
}

func transAppUrlListResp2Model(resp *app_service.GetAppUrlListResp) []*response.AppUrlInfo {
	var results []*response.AppUrlInfo
	for _, appUrl := range resp.AppUrlInfos {
		openUrl, _ := url.JoinPath(config.Cfg().Server.AppOpenUrl, appUrl.Suffix)
		ret := transAppUrlInfo(appUrl)
		ret.Suffix = openUrl
		results = append(results, ret)
	}
	return results
}

func transAppUrlInfo(resp *app_service.AppUrlInfo) *response.AppUrlInfo {
	ret := &response.AppUrlInfo{
		UrlId:               resp.UrlId,
		AppId:               resp.AppId,
		AppType:             resp.AppType,
		Name:                resp.Name,
		CreatedAt:           util.Time2Str(resp.CreatedAt),
		ExpiredAt:           "",
		Copyright:           resp.Copyright,
		CopyrightEnable:     resp.CopyrightEnable,
		PrivacyPolicy:       resp.PrivacyPolicy,
		PrivacyPolicyEnable: resp.PrivacyPolicyEnable,
		Disclaimer:          resp.Disclaimer,
		DisclaimerEnable:    resp.DisclaimerEnable,
		Suffix:              resp.Suffix,
		Status:              resp.Status,
		UserId:              resp.UserId,
		OrgId:               resp.OrgId,
		Description:         resp.Description,
	}
	if resp.ExpiredAt != 0 {
		ret.ExpiredAt = util.Time2Str(resp.ExpiredAt)
	}
	return ret
}
