package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAppVersionList
//
//	@Tags			common
//	@Summary		获取应用版本列表
//	@Description	根据应用类型和应用ID，查询其所有历史版本
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetAppVersionListRequest	true	"获取应用版本列表参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AppVersionInfo}}
//	@Router			/appspace/app/version/list [get]
func GetAppVersionList(ctx *gin.Context) {
	var req request.GetAppVersionListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}

	resp, err := service.GetAppVersionList(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.AppId)
	gin_util.Response(ctx, resp, err)
}

// UpdateAppVersion
//
//	@Tags			common
//	@Summary		更新应用版本信息
//	@Description	更新应用最新版本的描述信息和公开范围
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateAppVersionRequest	true	"更新应用版本信息请求"
//	@Success		200		{object}	response.Response{}
//	@Router			/appspace/app/version [put]
func UpdateAppVersion(ctx *gin.Context) {
	var req request.UpdateAppVersionRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	err := service.UpdateAppVersion(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.AppId, req.Desc, req.PublishType, req.Extra)
	gin_util.Response(ctx, nil, err)
}

// RollbackAppVersion
//
//	@Tags			common
//	@Summary		回滚应用到指定版本
//	@Description	将应用回滚至指定的历史版本
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RollbackAppVersionRequest	true	"回滚应用版本请求"
//	@Success		200		{object}	response.Response{}
//	@Router			/appspace/app/version/rollback [post]
func RollbackAppVersion(ctx *gin.Context) {
	var req request.RollbackAppVersionRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	err := service.RollbackAppVersion(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.AppID, req.Version)
	gin_util.Response(ctx, nil, err)
}

// GetAppLatestVersion
//
//	@Tags			common
//	@Summary		获取应用最新版本信息
//	@Description	根据应用类型和应用ID，查询其最新版本信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetAppVersionListRequest	true	"获取应用最新版本信息请求参数"
//	@Success		200		{object}	response.Response{data=response.AppVersionInfo}
//	@Router			/appspace/app/version [get]
func GetAppLatestVersion(ctx *gin.Context) {
	var req request.GetAppVersionListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}

	resp, err := service.GetAppLatestVersion(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.AppId)
	gin_util.Response(ctx, resp, err)
}
