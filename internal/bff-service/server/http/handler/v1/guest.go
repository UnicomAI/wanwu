package v1

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform API
//	@version	v0.0.1
//	@description.markdown
//	@securityDefinitions.apikey	JWT
//	@in							header
//	@name						Authorization

//	@BasePath	/v1

// Login
//
//	@Tags		guest
//	@Summary	用户登录 [EN] @Summary User login
//	@Accept		json
//	@Produce	json
//	@Param		X-Language	header		string			false	"语言" [EN] @Param X-Language header string false "Language"
//	@Param		data		body		request.Login	true	"用户名+密码" [EN] @Param data body request.Login true "Username + Password"
//	@Success	200			{object}	response.Response{data=response.Login}
//	@Router		/base/login [post]
func Login(ctx *gin.Context) {
	var req request.Login
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.Login(ctx, &req, getLanguage(ctx))
	gin_util.Response(ctx, resp, err)
}

// LoginByEmail
//
//	@Tags		guest
//	@Summary	用户邮箱双因子登录 [EN] @Summary User email two-factor login
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.Login	true	"用户名+密码" [EN] @Param data body request.Login true "Username + Password"
//	@Success	200		{object}	response.Response{data=response.LoginByEmail}
//	@Router		/base/login/email [post]
func LoginByEmail(ctx *gin.Context) {
	var req request.Login
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.LoginByEmail(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// GetCaptcha
//
//	@Tags		guest
//	@Summary	获取验证码 [EN] @Summary Get verification code
//	@Accept		json
//	@Produce	json
//	@Param		X-Language	header		string	false	"语言" [EN] @Param X-Language header string false "Language"
//	@Success	200			{object}	response.Response{data=response.Captcha}
//	@Router		/base/captcha [get]
func GetCaptcha(ctx *gin.Context) {
	resp, err := service.GetCaptcha(ctx,
		util.MD5([]byte(ctx.ClientIP()+ctx.GetHeader("User-Agent")+ctx.GetHeader("Date"))))
	gin_util.Response(ctx, resp, err)
}

// GetLogoCustomInfo
//
//	@Tags		guest
//	@Summary	自定义logo和title [EN] @Summary Customized logo and title
//	@Produce	application/json
//	@Param		X-Language	header		string	false	"语言" [EN] @Param X-Language header string false "Language"
//	@Success	200			{object}	response.Response{data=response.LogoCustomInfo}
//	@Router		/base/custom [get]
func GetLogoCustomInfo(ctx *gin.Context) {
	resp, err := service.GetLogoCustomInfo(ctx, config.Cfg().CustomInfo.DefaultMode)
	gin_util.Response(ctx, resp, err)
}

// GetLanguageSelect
//
//	@Tags		guest
//	@Summary	获取语言列表（用于下拉选择） [EN] @Summary Get the language list (for drop-down selection)
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.LanguageSelect}
//	@Router		/base/language/select [get]
func GetLanguageSelect(ctx *gin.Context) {
	resp := service.GetLanguageSelect()
	gin_util.Response(ctx, resp, nil)
}

// RegisterByEmail
//
//	@Tags		guest
//	@Summary	用户邮箱注册 [EN] @Summary User email registration
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RegisterByEmail	true	"邮箱注册信息" [EN] @Param data body request.RegisterByEmail true "Email registration information"
//	@Success	200		{object}	response.Response
//	@Router		/base/register/email [post]
func RegisterByEmail(ctx *gin.Context) {
	var req request.RegisterByEmail
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.RegisterByEmail(ctx, &req))
}

// ResgisterSendEmailCode
//
//	@Tags		guest
//	@Summary	邮箱注册验证码发送 [EN] @Summary Email registration verification code sent
//	@Accept		json
//	@Produce	application/json
//	@Param		data	body		request.RegisterSendEmailCode	true	"邮箱地址" [EN] @Param data body request.RegisterSendEmailCode true "Email address"
//	@Success	200		{object}	response.Response
//	@Router		/base/register/email/code [post]
func ResgisterSendEmailCode(ctx *gin.Context) {
	var req request.RegisterSendEmailCode
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.RegisterSendEmailCode(ctx, req.Username, req.Email)
	gin_util.Response(ctx, nil, err)
}

// ResetPasswordSendEmailCode
//
//	@Tags		guest
//	@Summary	重置密码邮箱验证码发送 [EN] @Summary Reset password email verification code sent
//	@Accept		json
//	@Produce	application/json
//	@Param		data	body		request.ResetPasswordSendEmailCode	true	"邮箱地址" [EN] @Param data body request.ResetPasswordSendEmailCode true "Email address"
//	@Success	200		{object}	response.Response
//	@Router		/base/password/email/code [post]
func ResetPasswordSendEmailCode(ctx *gin.Context) {
	var req request.ResetPasswordSendEmailCode
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ResetPasswordSendEmailCode(ctx, req.Email)
	gin_util.Response(ctx, nil, err)
}

// ResetPasswordByEmail
//
//	@Tags		guest
//	@Summary	邮箱重置密码 [EN] @Summary Email password reset
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.ResetPasswordByEmail	true	"重置密码信息" [EN] @Param data body request.ResetPasswordByEmail true "Reset password information"
//	@Success	200		{object}	response.Response
//	@Router		/base/password/email [post]
func ResetPasswordByEmail(ctx *gin.Context) {
	var req request.ResetPasswordByEmail
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.ResetPasswordByEmail(ctx, &req))
}

// GetWorkflowTemplateList
//
//	@Tags			guest
//	@Summary		获取工作流模板列表 [EN] @Summary Get the list of workflow templates
//	@Description	获取工作流模板列表 [EN] @Description Get the list of workflow templates
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"模板名称" [EN] @Param name query string false "template name"
//	@Param			category	query		string	false	"模板分类" [EN] @Param category query string false "template category"
//	@Success		200			{object}	response.Response{data=response.GetWorkflowTemplateListResp}
//	@Router			/workflow/template/list [get]
func GetWorkflowTemplateList(ctx *gin.Context) {
	resp, err := service.GetWorkflowTemplateList(ctx, getClientID(ctx), ctx.Query("category"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetWorkflowTemplateDetail
//
//	@Tags			guest
//	@Summary		获取工作流模板详情 [EN] @Summary Get workflow template details
//	@Description	获取工作流模板详情 [EN] @Description Get workflow template details
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			templateId	query		string	true	"模板ID" [EN] @Param templateId query string true "template ID"
//	@Success		200			{object}	response.Response{data=response.WorkflowTemplateDetail}
//	@Router			/workflow/template/detail [get]
func GetWorkflowTemplateDetail(ctx *gin.Context) {
	resp, err := service.GetWorkflowTemplateDetail(ctx, getClientID(ctx), ctx.Query("templateId"))
	gin_util.Response(ctx, resp, err)
}

// GetWorkflowTemplateRecommend
//
//	@Tags			guest
//	@Summary		获取工作流模板推荐 [EN] @Summary Get workflow template recommendations
//	@Description	获取工作流模板推荐 [EN] @Description Get workflow template recommendations
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			templateId	query		string	false	"模板ID" [EN] @Param templateId query string false "template ID"
//	@Success		200			{object}	response.Response{data=response.GetWorkflowTemplateListResp}
//	@Router			/workflow/template/recommend [get]
func GetWorkflowTemplateRecommend(ctx *gin.Context) {
	resp, err := service.GetWorkflowTemplateRecommend(ctx, getClientID(ctx), ctx.Query("templateId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadWorkflowTemplate
//
//	@Tags			guest
//	@Summary		下载工作流模板 [EN] @Summary Download workflow template
//	@Description	下载工作流模板 [EN] @Description Download workflow template
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			templateId	query		string	true	"模板ID" [EN] @Param templateId query string true "template ID"
//	@Success		200			{object}	response.Response
//	@Router			/workflow/template/download [get]
func DownloadWorkflowTemplate(ctx *gin.Context) {
	fileName := fmt.Sprintf("%s.json", ctx.Query("templateId"))
	resp, err := service.DownloadWorkflowTemplate(ctx, getClientID(ctx), ctx.Query("templateId"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 设置响应头 [EN] Set response headers
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	// 直接写入字节数据 [EN] Write byte data directly
	ctx.Data(http.StatusOK, "application/octet-stream", resp)
}

// GetPromptTemplateList
//
//	@Tags			guest
//	@Summary		获取提示词模板列表 [EN] @Summary Get the prompt word template list
//	@Description	获取提示词模板列表 [EN] @Description Get the prompt word template list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"模板名称" [EN] @Param name query string false "template name"
//	@Param			category	query		string	false	"模板分类" [EN] @Param category query string false "template category"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.PromptTemplateDetail}}
//	@Router			/prompt/template/list [get]
func GetPromptTemplateList(ctx *gin.Context) {
	resp, err := service.GetPromptTemplateList(ctx, ctx.Query("category"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetPromptTemplateDetail
//
//	@Tags			guest
//	@Summary		获取提示词模板详情 [EN] @Summary Get prompt word template details
//	@Description	获取提示词模板详情 [EN] @Description Get prompt word template details
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			templateId	query		string	true	"模板ID" [EN] @Param templateId query string true "template ID"
//	@Success		200			{object}	response.Response{data=response.PromptTemplateDetail}
//	@Router			/prompt/template/detail [get]
func GetPromptTemplateDetail(ctx *gin.Context) {
	resp, err := service.GetPromptTemplateDetail(ctx, ctx.Query("templateId"))
	gin_util.Response(ctx, resp, err)
}
