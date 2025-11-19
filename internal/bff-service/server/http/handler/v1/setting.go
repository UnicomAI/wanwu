package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// UploadCustomTab
//
//	@Tags			setting
//	@Summary		标签页自定义配置 [EN] @Summary tag page custom configuration
//	@Description	上传标签页图标、标签页标题 [EN] @Description Upload tab icon, tab title
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomTabConfig	true	"标签页配置请求参数" [EN] @Param data body request.CustomTabConfig true "Tab page configuration request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/custom/tab [post]
func UploadCustomTab(ctx *gin.Context) {
	var req request.CustomTabConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UploadCustomTab(ctx, getUserID(ctx), getOrgID(ctx), config.Cfg().CustomInfo.DefaultMode, &req)
	gin_util.Response(ctx, nil, err)

}

// UploadCustomLogin
//
//	@Tags			setting
//	@Summary		登录页自定义配置 [EN] @Summary Login page custom configuration
//	@Description	上传登录页背景图、登录页欢迎语、登录按钮颜色 [EN] @Description Upload login page background image, login page welcome message, login button color
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomLoginConfig	true	"登录页配置请求参数" [EN] @Param data body request.CustomLoginConfig true "Login page configuration request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/custom/login [post]
func UploadCustomLogin(ctx *gin.Context) {
	var req request.CustomLoginConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UploadCustomLogin(ctx, getUserID(ctx), getOrgID(ctx), config.Cfg().CustomInfo.DefaultMode, &req)
	gin_util.Response(ctx, nil, err)
}

// UploadCustomHome
//
//	@Tags			setting
//	@Summary		平台自定义配置 [EN] @Summary Platform custom configuration
//	@Description	配置平台名称、平台图标、平台背景色 [EN] @Description Configure platform name, platform icon, platform background color
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomHomeConfig	true	"平台配置请求参数" [EN] @Param data body request.CustomHomeConfig true "Platform configuration request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/custom/home [post]
func UploadCustomHome(ctx *gin.Context) {
	var req request.CustomHomeConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UploadCustomHome(ctx, getUserID(ctx), getOrgID(ctx), config.Cfg().CustomInfo.DefaultMode, &req)
	gin_util.Response(ctx, nil, err)
}
