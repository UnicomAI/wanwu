package v1

import (
	"net/http"
	"net/url"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateChatflow
//
//	@Tags		chatflow
//	@Summary	创建Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.CreateChatflowReq	true	"创建Chatflow的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/chatflow [post]
func CreateChatflow(ctx *gin.Context) {
	var req request.CreateChatflowReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateChatflow(ctx, getOrgID(ctx), req.Name, req.Desc, req.Avatar.Key)
	gin_util.Response(ctx, resp, err)
}

// CopyChatflow
//
//	@Tags		chatflow
//	@Summary	拷贝Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.WorkflowIDReq	true	"拷贝Chatflow的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/chatflow/copy [post]
func CopyChatflow(ctx *gin.Context) {
	var req request.WorkflowIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyWorkflow(ctx, getOrgID(ctx), req.WorkflowID, true)
	gin_util.Response(ctx, resp, err)
}

// CopyChatflowDraft
//
//	@Tags		chatflow
//	@Summary	拷贝Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.WorkflowIDReq	true	"拷贝Chatflow草稿的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/chatflow/copy/draft [post]
func CopyChatflowDraft(ctx *gin.Context) {
	var req request.WorkflowIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyWorkflow(ctx, getOrgID(ctx), req.WorkflowID, false)
	gin_util.Response(ctx, resp, err)
}

// ImportChatflow
//
//	@Tags			chatflow
//	@Summary		导入Chatflow
//	@Description	通过JSON文件导入工作流
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"工作流JSON文件"
//	@Success		200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router			/appspace/chatflow/import [post]
func ImportChatflow(ctx *gin.Context) {
	resp, err := service.ImportWorkflow(ctx, getOrgID(ctx), constant.AppTypeChatflow)
	gin_util.Response(ctx, resp, err)
}

// ExportChatflow
//
//	@Tags			chatflow
//	@Summary		导出Chatflow
//	@Description	导出对话流的json文件
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workflow_id	query		string	true	"工作流ID"
//	@Param			version		query		string	false	"版本"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/chatflow/export [get]
func ExportChatflow(ctx *gin.Context) {
	fileName := "chatflow_export.json"
	workflowID := ctx.Query("workflow_id")
	version := ctx.Query("version")
	resp, err := service.ExportWorkFlow(ctx, getOrgID(ctx), workflowID, version, true)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 设置响应头
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	// 直接写入字节数据
	ctx.Data(http.StatusOK, "application/octet-stream", resp)
}

// ExportChatflowDraft
//
//	@Tags			chatflow
//	@Summary		导出Chatflow草稿
//	@Description	导出对话流草稿的json文件
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workflow_id	query		string	true	"工作流ID"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/chatflow/export/draft [get]
func ExportChatflowDraft(ctx *gin.Context) {
	fileName := "chatflow_export.json"
	workflowID := ctx.Query("workflow_id")
	resp, err := service.ExportWorkFlow(ctx, getOrgID(ctx), workflowID, "", false)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 设置响应头
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	// 直接写入字节数据
	ctx.Data(http.StatusOK, "application/octet-stream", resp)
}

// ChatflowConvert
//
//	@Tags			chatflow
//	@Summary		chatflow转为workflow
//	@Description	chatflow转为workflow
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.WorkflowConvertReq	true	"对话流工作流转换参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/appspace/chatflow/convert [post]
func ChatflowConvert(ctx *gin.Context) {
	var req request.WorkflowConvertReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.WorkflowConvert(ctx, getOrgID(ctx), req.WorkflowID, constant.AppTypeWorkflow))
}

// ChatflowApplicationList
//
//	@Tags			chatflow
//	@Summary		应用广场对话流关联应用
//	@Description	应用广场对话流关联应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ChatflowApplicationListReq	true	"应用广场对话流关联应用参数"
//	@Success		200		{object}	response.Response{data=response.CozeDraftIntelligenceListData}
//	@Router			/chatflow/application/list [post]
func ChatflowApplicationList(ctx *gin.Context) {
	var req request.ChatflowApplicationListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ChatflowApplicationList(ctx, getUserID(ctx), getOrgID(ctx), req.WorkflowID)
	gin_util.Response(ctx, resp, err)
}

// ChatflowApplicationInfo
//
//	@Tags			chatflow
//	@Summary		应用广场对话流关联应用信息
//	@Description	应用广场对话流关联应用信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ChatflowApplicationInfoReq	true	"应用广场对话流关联应用信息参数"
//	@Success		200		{object}	response.Response{data=response.CozeGetDraftIntelligenceInfoData}
//	@Router			/chatflow/application/info [post]
func ChatflowApplicationInfo(ctx *gin.Context) {
	var req request.ChatflowApplicationInfoReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ChatflowApplicationInfo(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteChatflowConversation
//
//	@Tags			chatflow
//	@Summary		删除对话流会话
//	@Description	删除对话流会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ChatflowConversationDeleteReq	true	"删除对话流会话请求参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/chatflow/conversation/delete [delete]
func DeleteChatflowConversation(ctx *gin.Context) {
	var req request.ChatflowConversationDeleteReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteChatflowConversation(ctx, getOrgID(ctx), req.ProjectId, req.UniqueId))
}
