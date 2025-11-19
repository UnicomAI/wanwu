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
//	@Summary	创建Chatflow [EN] @Summary Create Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.AppBriefConfig	true	"创建Chatflow的请求参数" [EN] @Param data body request.AppBriefConfig true "Create request parameters for Chatflow"
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
//	@Summary	拷贝Chatflow [EN] @Summary Copy Chatflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.WorkflowIDReq	true	"拷贝Chatflow的请求参数" [EN] @Param data body request.WorkflowIDReq true "Copy Chatflow's request parameters"
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
//	@Summary		导入Chatflow [EN] @Summary Import Chatflow
//	@Description	通过JSON文件导入工作流 [EN] @Description Import workflow via JSON file
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"工作流JSON文件" [EN] @Param file formData file true "Workflow JSON file"
//	@Success		200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router			/appspace/chatflow/import [post]
func ImportChatflow(ctx *gin.Context) {
	resp, err := service.ImportWorkflow(ctx, getOrgID(ctx), constant.AppTypeChatflow)
	gin_util.Response(ctx, resp, err)
}

// ExportChatflow
//
//	@Tags			chatflow
//	@Summary		导出Chatflow [EN] @Summary Export Chatflow
//	@Description	导出工作流的json文件 [EN] @Description Export the json file of the workflow
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workflow_id	query		string	true	"工作流ID" [EN] @Param workflow_id query string true "workflow ID"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/chatflow/export [get]
func ExportChatflow(ctx *gin.Context) {
	fileName := "chatflow_export.json"
	resp, err := service.ExportWorkflow(ctx, getOrgID(ctx), ctx.Query("workflow_id"))
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
