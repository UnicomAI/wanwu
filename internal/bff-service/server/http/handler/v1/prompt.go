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
//	@Summary Create a custom Prompt
//	@Description Create a custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param data body request.CustomPromptCreate true "Custom Prompt information"
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
//	@Summary Get custom Prompt details
//	@Description Get custom Prompt details
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
//	@Summary Delete custom Prompt
//	@Description Delete custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param data body request.CustomPromptIDReq true "Custom PromptID"
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
//	@Summary Update custom Prompt
//	@Description Update custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateCustomPrompt true "Custom Prompt information"
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
//	@Summary Get the custom Prompt list
//	@Description Get the custom prompt list
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
//	@Summary Copy custom Prompt
//	@Description Copy custom Prompt
//	@Accept			json
//	@Produce		json
//	@Param data body request.CustomPromptIDReq true "Custom PromptID"
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
//	@Summary Copy prompt word template
//	@Description Copy prompt word template
//	@Accept			json
//	@Produce		json
//	@Param data body request.CreatePromptByTemplateReq true "Create request parameters for prompt words through templates"
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
//	@Summary Get the prompt word optimization results
//	@Description Get the prompt word optimization results
//	@Accept			json
//	@Produce		json
//	@Param data body request.PromptOptimizeReq true "Prompt word optimization request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/prompt/optimize [post]
func GetPromptOptimize(ctx *gin.Context) {
	var req request.PromptOptimizeReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	service.GetPromptOptimize(ctx, getUserID(ctx), getOrgID(ctx), req)
}
