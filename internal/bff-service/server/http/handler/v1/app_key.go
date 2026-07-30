package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAppBaseUrl
//
//	@Tags			app.key
//	@Summary		获取App根地址
//	@Description	获取App根地址
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetAppBaseUrlRequest	true	"获取App根地址参数"
//	@Success		200		{object}	response.Response{data=string}
//	@Router			/appspace/app/url [get]
func GetAppBaseUrl(ctx *gin.Context) {
	var req request.GetAppBaseUrlRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppBaseUrl(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// GenAppKey
//
//	@Tags			app.key
//	@Summary		生成AppKey
//	@Description	生成AppKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GenAppKeyRequest	true	"生成AppKey参数"
//	@Success		200		{object}	response.Response{data=response.AppKeyInfo}
//	@Router			/appspace/app/key [post]
func GenAppKey(ctx *gin.Context) {
	var req request.GenAppKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GenAppKey(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DelAppKey
//
//	@Tags			app.key
//	@Summary		删除AppKey
//	@Description	删除AppKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DelAppKeyRequest	true	"删除Appkey参数"
//	@Success		200		{object}	response.Response
//	@Router			/appspace/app/key [delete]
func DelAppKey(ctx *gin.Context) {
	var req request.DelAppKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DelAppKey(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetAppKeyList
//
//	@Tags			app.key
//	@Summary		获取AppKey
//	@Description	获取AppKey
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetAppKeyListRequest	true	"获取AppKeyList参数"
//	@Success		200		{object}	response.Response{data=[]response.AppKeyInfo}
//	@Router			/appspace/app/key/list [get]
func GetAppKeyList(ctx *gin.Context) {
	var req request.GetAppKeyListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppKeyList(ctx, getUserID(ctx), req)
	gin_util.Response(ctx, resp, err)
}
