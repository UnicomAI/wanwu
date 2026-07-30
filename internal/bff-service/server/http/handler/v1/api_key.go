package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateAPIKey
//
//	@Tags			api.key
//	@Summary		创建API密钥
//	@Description	创建API密钥
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateAPIKeyRequest	true	"创建API密钥参数"
//	@Success		200		{object}	response.Response{data=response.APIKeyDetailResponse}
//	@Router			/api/key [post]
func CreateAPIKey(ctx *gin.Context) {
	var req request.CreateAPIKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateApiKey(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteAPIKey
//
//	@Tags			api.key
//	@Summary		删除API	密钥
//	@Description	删除API密钥
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteAPIKeyRequest	true	"删除API密钥参数"
//	@Success		200		{object}	response.Response
//	@Router			/api/key [delete]
func DeleteAPIKey(ctx *gin.Context) {
	var req request.DeleteAPIKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteApiKey(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// ListAPIKeys
//
//	@Tags			api.key
//	@Summary		获取API密钥列表
//	@Description	获取API密钥列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			pageNo		query		int	true	"页面编号，从1开始"
//	@Param			pageSize	query		int	true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.APIKeyDetailResponse}}
//	@Router			/api/key/list [get]
func ListAPIKeys(ctx *gin.Context) {
	resp, err := service.ListApiKeys(ctx, getUserID(ctx), getOrgID(ctx), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// UpdateAPIKey
//
//	@Tags			api.key
//	@Summary		更新API密钥
//	@Description	更新API密钥
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateAPIKeyRequest	true	"更新API密钥参数"
//	@Success		200		{object}	response.Response
//	@Router			/api/key [put]
func UpdateAPIKey(ctx *gin.Context) {
	var req request.UpdateAPIKeyRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateApiKey(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// UpdateAPIKeyStatus
//
//	@Tags			api.key
//	@Summary		更新API密钥状态
//	@Description	更新API密钥状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateAPIKeyStatusRequest	true	"更新API密钥状态参数"
//	@Success		200		{object}	response.Response
//	@Router			/api/key/status [put]
func UpdateAPIKeyStatus(ctx *gin.Context) {
	var req request.UpdateAPIKeyStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateApiKeyStatus(ctx, req))
}
