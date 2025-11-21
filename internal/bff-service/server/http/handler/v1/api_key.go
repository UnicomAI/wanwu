package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetApiBaseUrl
//
//	@Tags			app.key
//	@Summary Get the Api root address
//	@Description Get the Api root address
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetApiBaseUrlRequest true "Get Api root address parameters"
//	@Success		200		{object}	response.Response{data=string}
//	@Router			/appspace/app/url [get]
func GetApiBaseUrl(ctx *gin.Context) {
	var req request.GetApiBaseUrlRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetApiBaseUrl(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// GenApiKey
//
//	@Tags			app.key
//	@Summary Generate ApiKey
//	@Description Generate ApiKey
//	@Accept			json
//	@Produce		json
//	@Param data body request.GenApiKeyRequest true "Generate ApiKey parameters"
//	@Success		200		{object}	response.Response{data=response.ApiResponse}
//	@Router			/appspace/app/key [post]
func GenApiKey(ctx *gin.Context) {
	var req request.GenApiKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GenApiKey(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DelApiKey
//
//	@Tags			app.key
//	@Summary Delete ApiKey
//	@Description Delete ApiKey
//	@Accept			json
//	@Produce		json
//	@Param data body request.DelApiKeyRequest true "Delete Apikey parameters"
//	@Success		200		{object}	response.Response
//	@Router			/appspace/app/key [delete]
func DelApiKey(ctx *gin.Context) {
	var req request.DelApiKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DelApiKey(ctx, req)
	gin_util.Response(ctx, nil, err)
}

// GetApiKeyList
//
//	@Tags			app.key
//	@Summary Get ApiKey
//	@Description Get ApiKey
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetApiKeyListRequest true "Get ApiKey parameters"
//	@Success		200		{object}	response.Response{data=[]response.ApiResponse}
//	@Router			/appspace/app/key/list [get]
func GetApiKeyList(ctx *gin.Context) {
	var req request.GetApiKeyListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetApiKeyList(ctx, getUserID(ctx), req)
	gin_util.Response(ctx, resp, err)
}
