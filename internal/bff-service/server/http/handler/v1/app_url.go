package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// AppUrlCreate
//
//	@Tags			app.url
//	@Summary		创建应用Url [EN] @Summary Create application Url
//	@Description	创建应用Url [EN] @Description Create application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data					body		request.AppUrlCreateRequest	true	"应用Url基本信息" [EN] @Param data body request.AppUrlCreateRequest true "Basic information of application Url"
//	@Success		200						{object}	response.Response
//	@Router			/appspace/app/openurl	[post]
func AppUrlCreate(ctx *gin.Context) {
	var req request.AppUrlCreateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AppUrlCreate(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// AppUrlDelete
//
//	@Tags			app.url
//	@Summary		删除应用Url [EN] @Summary Delete application Url
//	@Description	删除应用Url [EN] @Description Delete application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data					body		request.AppUrlIdRequest	true	"应用UrlId" [EN] @Param data body request.AppUrlIdRequest true "AppUrlId"
//	@Success		200						{object}	response.Response
//	@Router			/appspace/app/openurl	[delete]
func AppUrlDelete(ctx *gin.Context) {
	var req request.AppUrlIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AppUrlDelete(ctx, req)
	gin_util.Response(ctx, nil, err)
}

// AppUrlUpdate
//
//	@Tags			app.url
//	@Summary		编辑应用Url [EN] @Summary Edit application Url
//	@Description	编辑应用Url [EN] @Description Edit application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data					body		request.AppUrlUpdateRequest	true	"应用Url基本信息" [EN] @Param data body request.AppUrlUpdateRequest true "Basic information of application Url"
//	@Success		200						{object}	response.Response
//	@Router			/appspace/app/openurl	[put]
func AppUrlUpdate(ctx *gin.Context) {
	var req request.AppUrlUpdateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AppUrlUpdate(ctx, req)
	gin_util.Response(ctx, nil, err)
}

// GetAppUrlList
//
//	@Tags			app.url
//	@Summary		获取应用Url列表 [EN] @Summary Get the application URL list
//	@Description	获取应用Url列表 [EN] @Description Get the application URL list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data						query		request.AppUrlListRequest	true	"应用id和type" [EN] @Param data query request.AppUrlListRequest true "Application id and type"
//	@Success		200							{object}	response.Response{data=response.ListResult{list=[]response.AppUrlInfo}}
//	@Router			/appspace/app/openurl/list 	[get]
func GetAppUrlList(ctx *gin.Context) {
	var req request.AppUrlListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppUrlList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AppUrlStatusSwitch
//
//	@Tags			app.url
//	@Summary		启用/停用应用Url [EN] @Summary Enable/disable application Url
//	@Description	启用/停用应用Url [EN] @Description Enable/disable application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data							body		request.AppUrlStatusRequest	true	"应用UrlId" [EN] @Param data body request.AppUrlStatusRequest true "AppUrlId"
//	@Success		200								{object}	response.Response
//	@Router			/appspace/app/openurl/status 	[put]
func AppUrlStatusSwitch(ctx *gin.Context) {
	var req request.AppUrlStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AppUrlStatusSwitch(ctx, req)
	gin_util.Response(ctx, nil, err)
}
