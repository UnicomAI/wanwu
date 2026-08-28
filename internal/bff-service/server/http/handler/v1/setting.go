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
//	@Tags			admin_center.setting
//	@Summary		标签页自定义配置
//	@Description	上传标签页图标、标签页标题
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomTabConfig	true	"标签页配置请求参数"
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
//	@Tags			admin_center.setting
//	@Summary		登录页自定义配置
//	@Description	上传登录页背景图、登录页欢迎语、登录按钮颜色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomLoginConfig	true	"登录页配置请求参数"
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
//	@Tags			admin_center.setting
//	@Summary		平台自定义配置
//	@Description	配置平台名称、平台图标、平台背景色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomHomeConfig	true	"平台配置请求参数"
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

// UploadCustomGeneralAgent
//
//	@Tags			admin_center.setting
//	@Summary		通用智能体自定义配置
//	@Description	配置通用智能体图标
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomGeneralAgentConfig	true	"通用智能体配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/custom/general-agent [post]
func UploadCustomGeneralAgent(ctx *gin.Context) {
	var req request.CustomGeneralAgentConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UploadCustomGeneralAgent(ctx, getUserID(ctx), getOrgID(ctx), config.Cfg().CustomInfo.DefaultMode, &req)
	gin_util.Response(ctx, nil, err)
}
