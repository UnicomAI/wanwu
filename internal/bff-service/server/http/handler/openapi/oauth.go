package openapi

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// OAuthToken
//
//	@Tags			openapi.OIDC
//	@Summary		授权码方式 [EN] @Summary Authorization code method
//	@Description	授权码方式-获取Token [EN] @Description Authorization code method-get Token
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param			grant_type		formData	string	true	"授权类型" [EN] @Param grant_type formData string true "Grant type"
//	@Param			code			formData	string	true	"授权码" [EN] @Param code formData string true "Authorization code"
//	@Param			client_id		formData	string	true	"Client ID"
//	@Param			redirect_uri	formData	string	true	"回调地址" [EN] @Param redirect_uri formData string true "callback address"
//	@Param			client_secret	formData	string	true	"备案密钥" [EN] @Param client_secret formData string true "Registration Key"
//	@Success		200				{object}	response.OAuthTokenResponse
//	@Router			/oauth/code/token [post]
func OAuthToken(ctx *gin.Context) {
	var req request.OAuthTokenRequest
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.OAuthToken(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"err": err})
	}
	ctx.JSON(http.StatusOK, resp)
}

// OAuthRefresh
//
//	@Tags			openapi.OIDC
//	@Summary		刷新令牌 [EN] @Summary refresh token
//	@Description	刷新令牌 [EN] @Description refresh token
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OAuthRefreshRequest	true	"RefreshToken"
//	@Success		200		{object}	response.OAuthRefreshTokenResponse
//	@Router			/oauth/code/token/refresh [post]
func OAuthRefresh(ctx *gin.Context) {
	var req request.OAuthRefreshRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.OAuthRefresh(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// OAuthConfig
//
//	@Tags			openapi.OIDC
//	@Summary		动态客户端发现配置 [EN] @Summary Dynamic client discovery configuration
//	@Description	自动获取 OP 的配置信息 [EN] @Description automatically obtains OP configuration information
//	@Produce		json
//	@Success		200	{object}	response.OAuthConfig
//	@Router			/.well-known/openid-configuration [get]
func OAuthConfig(ctx *gin.Context) {
	resp, err := service.OAuthConfig(ctx)
	if err != nil {
		gin_util.Response(ctx, resp, err)
	}
	ctx.JSON(http.StatusOK, resp)
}

// OAuthJWKS
//
//	@Tags			openapi.OIDC
//	@Summary		公钥获取链接 [EN] @Summary Public key acquisition link
//	@Description	自动获取OAuthJWKS [EN] @Description Automatically obtain OAuthJWKS
//	@Produce		json
//	@Success		200	{object}	response.OAuthJWKS
//	@Router			/oauth/jwks [get]
func OAuthJWKS(ctx *gin.Context) {
	resp, err := service.OAuthJWKS(ctx)
	if err != nil {
		gin_util.Response(ctx, resp, err)
	}
	ctx.JSON(http.StatusOK, resp)
}

// OAuthGetUserInfo
//
//	@Tags			openapi.OIDC
//	@Summary		OAuth获取用户信息 [EN] @Summary OAuth gets user information
//	@Description	通过access token获取用户信息 [EN] @Description Get user information through access token
//	@Produce		json
//	@Success		200	{object}	response.OAuthGetUserInfo
//	@Router			/oauth/userinfo [get]
func OAuthGetUserInfo(ctx *gin.Context) {
	userID := getUserID(ctx)
	resp, err := service.OAuthGetUserInfo(ctx, userID)
	gin_util.Response(ctx, resp, err)
}

// OAuthLogin
//
//	@Tags			openapi.OIDC
//	@Summary		OAuth登录授权 [EN] @Summary OAuth login authorization
//	@Description	返回OAuth登录页面 [EN] @Description Return to OAuth login page
//	@Accept			json
//	@Produce		json
//	@Param			client_id		query		string	true	"备案ID" [EN] @Param client_id query string true "Registration ID"
//	@Param			redirect_uri	query		string	true	"重定向URI" [EN] @Param redirect_uri query string true "Redirect URI"
//	@Param			response_type	query		string	true	"响应类型" [EN] @Param response_type query string true "response type"
//	@Param			scope			query		string	false	"权限范围" [EN] @Param scope query string false "Permission scope"
//	@Param			state			query		string	true	"状态参数" [EN] @Param state query string true "state parameter"
//	@Success		302				{string}	string	"重定向到指定URI" [EN] @Success 302 {string} string "Redirect to specified URI"
//	@Router			/oauth/login [get]
func OAuthLogin(ctx *gin.Context) {
	var req request.OAuthRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	loginUri, err := service.OAuthLogin(ctx, &req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	ctx.Redirect(http.StatusFound, loginUri)
}
