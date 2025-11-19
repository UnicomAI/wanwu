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
//	@Summary Authorization code method
//	@Description Authorization code method-get Token
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param grant_type formData string true "Grant type"
//	@Param code formData string true "Authorization code"
//	@Param			client_id		formData	string	true	"Client ID"
//	@Param redirect_uri formData string true "callback address"
//	@Param client_secret formData string true "Registration Key"
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
//	@Summary refresh token
//	@Description refresh token
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
//	@Summary Dynamic client discovery configuration
//	@Description automatically obtains OP configuration information
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
//	@Summary Public key acquisition link
//	@Description Automatically obtain OAuthJWKS
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
//	@Summary OAuth gets user information
//	@Description Get user information through access token
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
//	@Summary OAuth login authorization
//	@Description Return to OAuth login page
//	@Accept			json
//	@Produce		json
//	@Param client_id query string true "Registration ID"
//	@Param redirect_uri query string true "Redirect URI"
//	@Param response_type query string true "response type"
//	@Param scope query string false "Permission scope"
//	@Param state query string true "state parameter"
//	@Success 302 {string} string "Redirect to specified URI"
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
