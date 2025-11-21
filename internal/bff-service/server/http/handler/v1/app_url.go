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
//	@Summary Create application Url
//	@Description Create application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AppUrlCreateRequest true "Basic information of application Url"
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
//	@Summary Delete application Url
//	@Description Delete application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AppUrlIdRequest true "AppUrlId"
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
//	@Summary Edit application Url
//	@Description Edit application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AppUrlUpdateRequest true "Basic information of application Url"
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
//	@Summary Get the application URL list
//	@Description Get the application URL list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.AppUrlListRequest true "Application id and type"
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
//	@Summary Enable/disable application Url
//	@Description Enable/disable application Url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AppUrlStatusRequest true "AppUrlId"
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
