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
//	@Summary Create Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.AppBriefConfig true "Create request parameters for Chatflow"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/chatflow [post]
func CreateChatflow(ctx *gin.Context) {
	var req request.AppBriefConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateChatflow(ctx, getOrgID(ctx), req.Name, req.Desc, req.Avatar.Key)
	gin_util.Response(ctx, resp, err)
}

// CopyChatflow
//
//	@Tags		chatflow
//	@Summary Copy Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.WorkflowIDReq true "Copy Chatflow's request parameters"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/chatflow/copy [post]
func CopyChatflow(ctx *gin.Context) {
	var req request.WorkflowIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyWorkflow(ctx, getOrgID(ctx), req.WorkflowID)
	gin_util.Response(ctx, resp, err)
}

// ImportChatflow
//
//	@Tags			chatflow
//	@Summary Import Chatflow
//	@Description Import workflow via JSON file
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param file formData file true "Workflow JSON file"
//	@Success		200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router			/appspace/chatflow/import [post]
func ImportChatflow(ctx *gin.Context) {
	resp, err := service.ImportWorkflow(ctx, getOrgID(ctx), constant.AppTypeChatflow)
	gin_util.Response(ctx, resp, err)
}

// ExportChatflow
//
//	@Tags			chatflow
//	@Summary Export Chatflow
//	@Description Export the json file of the workflow
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param workflow_id query string true "workflow ID"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/chatflow/export [get]
func ExportChatflow(ctx *gin.Context) {
	fileName := "chatflow_export.json"
	resp, err := service.ExportWorkflow(ctx, getOrgID(ctx), ctx.Query("workflow_id"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// Set response headers
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	// Write byte data directly
	ctx.Data(http.StatusOK, "application/octet-stream", resp)
}
