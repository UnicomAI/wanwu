package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateOauthApp
//
//	@Tags			admin_center.oauth
//	@Summary		创建OAuth应用
//	@Description	创建新的OAuth应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateOauthAppReq	true	"OAuth应用创建请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/oauth/app [post]
func CreateOauthApp(ctx *gin.Context) {
	var req request.CreateOauthAppReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.CreateOauthApp(ctx, getUserID(ctx), &req))
}

// DeleteOauthApp
//
//	@Tags			admin_center.oauth
//	@Summary		删除OAuth应用
//	@Description	删除指定的OAuth应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteOauthAppReq	true	"OAuth应用ID"
//	@Success		200		{object}	response.Response
//	@Router			/oauth/app [delete]
func DeleteOauthApp(ctx *gin.Context) {
	var req request.DeleteOauthAppReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteOauthApp(ctx, &req))
}

// UpdateOauthApp
//
//	@Tags			admin_center.oauth
//	@Summary		更新OAuth应用
//	@Description	更新OAuth应用信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateOauthAppReq	true	"OAuth应用更新请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/oauth/app [put]
func UpdateOauthApp(ctx *gin.Context) {
	var req request.UpdateOauthAppReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateOauthApp(ctx, &req))
}

// GetOauthAppList
//
//	@Tags			admin_center.oauth
//	@Summary		获取OAuth应用列表
//	@Description	获取OAuth应用分页列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"第三方平台名(模糊查询)"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.OAuthAppInfo}}
//	@Router			/oauth/app/list [get]
func GetOauthAppList(ctx *gin.Context) {
	resp, err := service.GetOauthAppList(ctx, getUserID(ctx), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)

}

// UpdateOauthAppStatus
//
//	@Tags			admin_center.oauth
//	@Summary		更新OAuth应用状态
//	@Description	启用或禁用OAuth应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateOauthAppStatusReq	true	"OAuth应用状态更新请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/oauth/app/status [put]
func UpdateOauthAppStatus(ctx *gin.Context) {
	var req request.UpdateOauthAppStatusReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateOauthAppStatus(ctx, &req))
}
