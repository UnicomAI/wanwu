package service

import (
	"net/url"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetAppBaseUrl(ctx *gin.Context, req request.GetAppBaseUrlRequest) (string, error) {
	apiBaseUrl, err := url.JoinPath(config.Cfg().Server.ApiBaseUrl, "/openapi/v1")
	if err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return apiBaseUrl, nil
}

func GetAppKeyByKey(ctx *gin.Context, appKey string) (*app_service.AppKeyInfo, error) {
	return app.GetAppKeyByKey(ctx.Request.Context(), &app_service.GetAppKeyByKeyReq{AppKey: appKey})
}

func GenAppKey(ctx *gin.Context, userId, orgId string, req request.GenAppKeyRequest) (*response.AppKeyInfo, error) {
	key, err := app.GenAppKey(ctx.Request.Context(), &app_service.GenAppKeyReq{
		AppId:   req.AppId,
		AppType: req.AppType,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.AppKeyInfo{
		ApiID:     key.AppKeyId,
		ApiKey:    key.AppKey,
		CreatedAt: util.Time2Str(key.CreatedAt),
	}, nil
}

func DelAppKey(ctx *gin.Context, userId, orgId string, req request.DelAppKeyRequest) error {
	_, err := app.DelAppKey(ctx.Request.Context(), &app_service.DelAppKeyReq{
		AppKeyId: req.ApiId,
		UserId:   userId,
		OrgId:    orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAppKeyList(ctx *gin.Context, userId string, req request.GetAppKeyListRequest) ([]*response.AppKeyInfo, error) {
	appKeyList, err := app.GetAppKeyList(ctx.Request.Context(), &app_service.GetAppKeyListReq{
		AppId:   req.AppId,
		AppType: req.AppType,
		UserId:  userId,
	})
	if err != nil {
		return nil, err
	}
	var appKeys []*response.AppKeyInfo
	for _, appKey := range appKeyList.Info {
		appKeys = append(appKeys, toAppKeyInfo(appKey))
	}
	return appKeys, nil
}

func toAppKeyInfo(appKey *app_service.AppKeyInfo) *response.AppKeyInfo {
	return &response.AppKeyInfo{
		ApiID:     appKey.AppKeyId,
		ApiKey:    appKey.AppKey,
		CreatedAt: util.Time2Str(appKey.CreatedAt),
	}
}
