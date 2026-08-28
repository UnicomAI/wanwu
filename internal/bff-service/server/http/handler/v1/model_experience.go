package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// ModelExperienceLLM
//
//	@Tags			model.experience
//	@Summary		模型体验
//	@Description	LLM模型体验
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelExperienceLlmRequest	true	"LLM模型体验"
//	@Success		200		{object}	response.Response
//	@Router			/model/experience/llm [post]
func ModelExperienceLLM(ctx *gin.Context) {
	var req request.ModelExperienceLlmRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	service.ModelExperienceLLM(ctx, getUserID(ctx), getOrgID(ctx), getClientID(ctx), &req)
}

// ModelExperienceLLMConnect
//
//	@Tags			model.experience
//	@Summary		模型体验流式问答断开后重连
//	@Description	模型体验流式问答断开后重连
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelExperienceLlmConnectRequest	true	"模型体验流式重连"
//	@Success		200		{object}	response.Response
//	@Router			/model/experience/llm/connect [post]
func ModelExperienceLLMConnect(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ModelExperienceLlmConnectRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ModelExperienceLLMConnect(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, nil)
	}
}

// ModelExperienceLLMCancel
//
//	@Tags			model.experience
//	@Summary		模型体验流式问答手动停止
//	@Description	模型体验流式问答手动停止
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelExperienceLlmCancelRequest	true	"模型体验流式问答手动停止参数"
//	@Success		200		{object}	response.Response
//	@Router			/model/experience/llm/cancel [post]
func ModelExperienceLLMCancel(ctx *gin.Context) {
	_, _, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ModelExperienceLlmCancelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ModelExperienceLLMCancel(req, clientId); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, nil)
}

// ModelExperienceSaveDialog
//
//	@Tags			model.experience
//	@Summary		新建/保存对话
//	@Description	新建/保存对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ModelExperienceDialogRequest	true	"模型体验对话"
//	@Success		200		{object}	response.Response{}
//	@Router			/model/experience/dialog [post]
func ModelExperienceSaveDialog(ctx *gin.Context) {
	var req request.ModelExperienceDialogRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.SaveModelExperienceDialog(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ModelExperienceListDialogs
//
//	@Tags			model.experience
//	@Summary		获取模型体验对话列表
//	@Description	获取模型体验对话列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.ListResult{list=model_service.ModelExperienceDialog}}
//	@Router			/model/experience/dialogs [get]
func ModelExperienceListDialogs(ctx *gin.Context) {
	resp, err := service.ListModelExperienceDialogs(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// ModelExperienceDeleteDialog
//
//	@Tags			model.experience
//	@Summary		删除模型体验对话
//	@Description	删除模型体验对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data						body		request.ModelExperienceDialogIDRequest	true	"模型体验对话ID"
//	@Success		200							{object}	response.Response
//	@Router			/model/experience/dialog	 [delete]
func ModelExperienceDeleteDialog(ctx *gin.Context) {
	var req request.ModelExperienceDialogIDRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteModelExperienceDialog(ctx, getUserID(ctx), getOrgID(ctx), req.ModelExperienceId)
	gin_util.Response(ctx, nil, err)
}

// ModelExperienceListDialogRecords
//
//	@Tags			model.experience
//	@Summary		获取模型体验对话记录列表
//	@Description	获取模型体验对话记录列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			modelExperienceId	query		uint32	true	"模型体验对话ID"
//	@Success		200					{object}	response.Response{data=response.ListResult{list=response.ModelExperienceDialogRecord}}
//	@Router			/model/experience/dialog/records [get]
func ModelExperienceListDialogRecords(ctx *gin.Context) {
	var req request.ModelExperienceDialogRecordRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListModelExperienceDialogRecords(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}
