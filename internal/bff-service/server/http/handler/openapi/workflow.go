package openapi

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// WorkflowRun
//
//	@Tags			openapi
//	@Summary		工作流OpenAPI
//	@Description	工作流OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIWorkflowRunReq	true	"工作流运行参数"
//	@Success		200		{object}	response.Response
//	@Router			/workflow/run [post]
func WorkflowRun(ctx *gin.Context) {
	var req request.OpenAPIWorkflowRunReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	// OpenAPI 权限检查：仅允许调用自己创建的已发布应用
	if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeWorkflow, userID, orgID); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	resp, err := service.OpenAPIWorkflowRun(ctx, userID, orgID, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	_, err = ctx.Writer.Write(resp)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, string(resp))
	ctx.Writer.Flush()
}

// CreateChatflowConversation
//
//	@Tags			openapi
//	@Summary		对话流创建对话OpenAPI
//	@Description	对话流创建对话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIChatflowCreateConversationRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPIChatflowCreateConversationResponse}
//	@Router			/chatflow/conversation [post]
func CreateChatflowConversation(ctx *gin.Context) {
	var req request.OpenAPIChatflowCreateConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	resp, err := service.CreateChatflowConversation(ctx, userID, orgID, req.UUID, req.ConversationName)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, resp, nil)
}

// GetConversationMessageList
//
//	@Tags			openapi
//	@Summary		对话流根据conversationId获取历史对话
//	@Description	对话流根据conversationId获取历史对话
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIChatflowGetConversationMessageListRequest	false	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPIChatflowGetConversationMessageListResponse}
//	@Router			/chatflow/conversation/message/list [post]
func GetConversationMessageList(ctx *gin.Context) {
	var req request.OpenAPIChatflowGetConversationMessageListRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	resp, err := service.GetConversationMessageList(ctx, userID, orgID, req.UUID, req.ConversationId, req.Limit)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, resp, nil)
}

// DeleteChatflowConversation
//
//	@Tags			openapi
//	@Summary		对话流删除会话OpenAPI
//	@Description	对话流删除会话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIChatflowDeleteConversationRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/chatflow/conversation [delete]
func DeleteChatflowConversation(ctx *gin.Context) {
	var req request.OpenAPIChatflowDeleteConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	err := service.DeleteChatflowConversationByConversationId(ctx, userID, orgID, req.UUID, req.ConversationId)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, nil)
}

// GetChatflowConversationList
//
//	@Tags			openapi
//	@Summary		对话流获取会话列表OpenAPI
//	@Description	对话流获取会话列表OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIChatflowGetConversationListRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPIChatflowConversationListResponse}
//	@Router			/chatflow/conversation/list [post]
func GetChatflowConversationList(ctx *gin.Context) {
	var req request.OpenAPIChatflowGetConversationListRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	resp, err := service.GetChatflowConversationList(ctx, userID, orgID, req.UUID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, resp, nil)
}

// ChatflowChat
//
//	@Tags			openapi
//	@Summary		对话流OpenAPI
//	@Description	对话流OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIChatflowChatRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/chatflow/chat [post]
func ChatflowChat(ctx *gin.Context) {
	var req request.OpenAPIChatflowChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	// OpenAPI 权限检查：仅允许调用自己创建的已发布应用
	if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeChatflow, userID, orgID); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// 流式处理 - 直接操作响应流
	err := service.ChatflowChat(ctx, userID, orgID, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// WorkflowFileUpload
//
//	@Tags			openapi
//	@Summary		工作流OpenAPI文件上传
//	@Description	工作流OpenAPI文件上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"文件"
//	@Success		200		{object}	string
//	@Success		400		{object}	response.Response
//	@Router			/workflow/file/upload [post]
func WorkflowFileUpload(ctx *gin.Context) {
	resp, err := service.OpenAPIWorkflowFileUpload(ctx)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.String(http.StatusOK, resp)
}

// ChatflowFileUpload
//
//	@Tags			openapi
//	@Summary		对话流OpenAPI文件上传
//	@Description	对话流OpenAPI文件上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"文件"
//	@Success		200		{object}	string
//	@Success		400		{object}	response.Response
//	@Router			/chatflow/file/upload [post]
func ChatflowFileUpload(ctx *gin.Context) {
	resp, err := service.OpenAPIWorkflowFileUpload(ctx)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.String(http.StatusOK, resp)
}
