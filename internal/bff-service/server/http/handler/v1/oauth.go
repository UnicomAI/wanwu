package v1

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// OAuthAuthorize
//
//	@Tags			oauth
//	@Summary		授权码方式 [EN] @Summary Authorization code method
//	@Description	授权码方式-获取授权码 [EN] @Description Authorization code method-obtain authorization code
//	@Accept			json
//	@Produce		json
//	@Param			client_id		query		string	true	"备案ID" [EN] @Param client_id query string true "Registration ID"
//	@Param			redirect_uri	query		string	true	"重定向URI" [EN] @Param redirect_uri query string true "Redirect URI"
//	@Param			response_type	query		string	true	"响应类型" [EN] @Param response_type query string true "response type"
//	@Param			scope			query		string	false	"权限范围" [EN] @Param scope query string false "Permission scope"
//	@Param			state			query		string	true	"状态参数" [EN] @Param state query string true "state parameter"
//	@Success		302				{string}	string	"重定向到指定URI" [EN] @Success 302 {string} string "Redirect to specified URI"
//	@Router			/oauth/code/authorize [get]
func OAuthAuthorize(ctx *gin.Context) {
	var req request.OAuthRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	userID := getUserID(ctx)
	callbackUri, err := service.OAuthAuthorize(ctx, &req, userID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	ctx.Redirect(http.StatusFound, callbackUri)
}

// CreateOauthApp
//
//	@Tags			oauth
//	@Summary		创建OAuth应用 [EN] @Summary Creating an OAuth application
//	@Description	创建新的OAuth应用 [EN] @Description Create a new OAuth application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateOauthAppReq	true	"OAuth应用创建请求参数" [EN] @Param data body request.CreateOauthAppReq true "OAuth application creation request parameters"
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
//	@Tags			oauth
//	@Summary		删除OAuth应用 [EN] @Summary Delete OAuth application
//	@Description	删除指定的OAuth应用 [EN] @Description Delete the specified OAuth application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteOauthAppReq	true	"OAuth应用ID" [EN] @Param data body request.DeleteOauthAppReq true "OAuth application ID"
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
//	@Tags			oauth
//	@Summary		更新OAuth应用 [EN] @Summary Update OAuth application
//	@Description	更新OAuth应用信息 [EN] @Description Update OAuth application information
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateOauthAppReq	true	"OAuth应用更新请求参数" [EN] @Param data body request.UpdateOauthAppReq true "OAuth application update request parameters"
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
//	@Tags			oauth
//	@Summary		获取OAuth应用列表 [EN] @Summary Get the OAuth application list
//	@Description	获取OAuth应用分页列表 [EN] @Description Get the OAuth application paging list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"第三方平台名(模糊查询)" [EN] @Param name query string false "Third-party platform name (fuzzy query)"
//	@Param			pageNo		query		int		true	"页面编号，从1开始" [EN] @Param pageNo query int true "Page number, starting from 1"
//	@Param			pageSize	query		int		true	"单页数量，从1开始" [EN] @Param pageSize query int true "Number of single pages, starting from 1"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.OAuthAppInfo}}
//	@Router			/oauth/app/list [get]
func GetOauthAppList(ctx *gin.Context) {
	resp, err := service.GetOauthAppList(ctx, getUserID(ctx), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)

}

// UpdateOauthAppStatus
//
//	@Tags			oauth
//	@Summary		更新OAuth应用状态 [EN] @Summary Update OAuth application status
//	@Description	启用或禁用OAuth应用 [EN] @Description Enable or disable OAuth application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateOauthAppStatusReq	true	"OAuth应用状态更新请求参数" [EN] @Param data body request.UpdateOauthAppStatusReq true "OAuth application status update request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/oauth/app/status [put]
func UpdateOauthAppStatus(ctx *gin.Context) {
	var req request.UpdateOauthAppStatusReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateOauthAppStatus(ctx, &req))
}
