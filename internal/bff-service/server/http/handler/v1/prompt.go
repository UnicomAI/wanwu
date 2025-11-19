package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateCustomPrompt
//
//	@Tags			prompt
//	@Summary		创建自定义Prompt [EN] @Summary Create a custom Prompt
//	@Description	创建自定义Prompt [EN] @Description Create a custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomPromptCreate	true	"自定义Prompt信息" [EN] @Param data body request.CustomPromptCreate true "Custom Prompt information"
//	@Success		200		{object}	response.Response{data=response.CustomPromptIDResp}
//	@Router			/prompt/custom [post]
func CreateCustomPrompt(ctx *gin.Context) {
	var req request.CustomPromptCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateCustomPrompt(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetCustomPrompt
//
//	@Tags			prompt
//	@Summary		获取自定义Prompt详情 [EN] @Summary Get custom Prompt details
//	@Description	获取自定义Prompt详情 [EN] @Description Get custom Prompt details
//	@Accept			json
//	@Produce		json
//	@Param			customPromptId	query		string	true	"customPromptId"
//	@Success		200				{object}	response.Response{data=response.CustomPrompt}
//	@Router			/prompt/custom [get]
func GetCustomPrompt(ctx *gin.Context) {
	resp, err := service.GetCustomPrompt(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("customPromptId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteCustomPrompt
//
//	@Tags			prompt
//	@Summary		删除自定义Prompt [EN] @Summary Delete custom Prompt
//	@Description	删除自定义Prompt [EN] @Description Delete custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomPromptIDReq	true	"自定义PromptID" [EN] @Param data body request.CustomPromptIDReq true "Custom PromptID"
//	@Success		200		{object}	response.Response{}
//	@Router			/prompt/custom [delete]
func DeleteCustomPrompt(ctx *gin.Context) {
	var req request.CustomPromptIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteCustomPrompt(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// UpdateCustomPrompt
//
//	@Tags			prompt
//	@Summary		更新自定义Prompt [EN] @Summary Update custom Prompt
//	@Description	更新自定义Prompt [EN] @Description Update custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateCustomPrompt	true	"自定义Prompt信息" [EN] @Param data body request.UpdateCustomPrompt true "Custom Prompt information"
//	@Success		200		{object}	response.Response{}
//	@Router			/prompt/custom [put]
func UpdateCustomPrompt(ctx *gin.Context) {
	var req request.UpdateCustomPrompt
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateCustomPrompt(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetCustomPromptList
//
//	@Tags			prompt
//	@Summary		获取自定义Prompt列表 [EN] @Summary Get the custom Prompt list
//	@Description	获取自定义Prompt列表 [EN] @Description Get the custom prompt list
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.CustomPrompt}}
//	@Router			/prompt/custom/list [get]
func GetCustomPromptList(ctx *gin.Context) {
	resp, err := service.GetCustomPromptList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// CopyCustomPrompt
//
//	@Tags			prompt
//	@Summary		复制自定义Prompt [EN] @Summary Copy custom Prompt
//	@Description	复制自定义Prompt [EN] @Description Copy custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomPromptIDReq	true	"自定义PromptID" [EN] @Param data body request.CustomPromptIDReq true "Custom PromptID"
//	@Success		200		{object}	response.Response{data=response.CustomPromptIDResp}
//	@Router			/prompt/custom/copy [post]
func CopyCustomPrompt(ctx *gin.Context) {
	var req request.CustomPromptIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyCustomPrompt(ctx, getUserID(ctx), getOrgID(ctx), req.CustomPromptID)
	gin_util.Response(ctx, resp, err)
}

// CreatePromptByTemplate
//
//	@Tags			prompt
//	@Summary		复制提示词模板 [EN] @Summary Copy prompt word template
//	@Description	复制提示词模板 [EN] @Description Copy prompt word template
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreatePromptByTemplateReq	true	"通过模板创建提示词的请求参数" [EN] @Param data body request.CreatePromptByTemplateReq true "Create request parameters for prompt words through templates"
//	@Success		200		{object}	response.Response{data=response.PromptIDData}
//	@Router			/prompt/template [post]
func CreatePromptByTemplate(ctx *gin.Context) {
	var req request.CreatePromptByTemplateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreatePromptByTemplate(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetPromptOptimize
//
//	@Tags			prompt
//	@Summary		获取提示词优化结果 [EN] @Summary Get the prompt word optimization results
//	@Description	获取提示词优化结果 [EN] @Description Get the prompt word optimization results
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.PromptOptimizeReq	true	"提示词优化请求参数" [EN] @Param data body request.PromptOptimizeReq true "Prompt word optimization request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/prompt/optimize [post]
func GetPromptOptimize(ctx *gin.Context) {
	var req request.PromptOptimizeReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	service.GetPromptOptimize(ctx, getUserID(ctx), getOrgID(ctx), req)
}
