package v1

import (
	"net/http"
	"net/url"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
)

// ListLlmModelsByWorkflow
//
//	@Tags		workflow
//	@Summary	llm模型列表（用于workflow）
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.CozeWorkflowModelInfo}}
//	@Router		/appspace/workflow/model/select/llm [get]
func ListLlmModelsByWorkflow(ctx *gin.Context) {
	resp, err := service.ListLlmModelsByWorkflow(ctx, getUserID(ctx), getOrgID(ctx), mp.ModelTypeLLM)
	gin_util.Response(ctx, resp, err)
}

// CreateWorkflow
//
//	@Tags			workflow
//	@Summary		创建工作流或对话流
//	@Description	通过 appType 区分创建工作流(workflow)或对话流(chatflow)，默认为 workflow
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateWorkflowReq	true	"创建工作流/对话流的请求参数"
//	@Success		200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router			/appspace/workflow [post]
func CreateWorkflow(ctx *gin.Context) {
	var req request.CreateWorkflowReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateWorkflowOrChatflow(ctx, getOrgID(ctx), req.AppType, req.Name, req.Desc, req.Avatar.Key)
	gin_util.Response(ctx, resp, err)
}

// CopyWorkflow
//
//	@Tags		workflow
//	@Summary	拷贝Workflow
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.WorkflowIDReq	true	"拷贝Workflow的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/workflow/copy [post]
func CopyWorkflow(ctx *gin.Context) {
	var req request.WorkflowIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyWorkflow(ctx, getOrgID(ctx), req.WorkflowID, true)
	gin_util.Response(ctx, resp, err)
}

// CopyWorkflowDraft
//
//	@Tags		workflow
//	@Summary	拷贝Workflow草稿
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.WorkflowIDReq	true	"拷贝Workflow草稿的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/appspace/workflow/copy/draft [post]
func CopyWorkflowDraft(ctx *gin.Context) {
	var req request.WorkflowIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyWorkflow(ctx, getOrgID(ctx), req.WorkflowID, false)
	gin_util.Response(ctx, resp, err)
}

// ExportWorkflow
//
//	@Tags			workflow
//	@Summary		导出Workflow
//	@Description	导出工作流的json文件
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workflow_id	query		string	true	"工作流ID"
//	@Param			version		query		string	false	"版本"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/workflow/export [get]
func ExportWorkflow(ctx *gin.Context) {
	fileName := "workflow_export.json"
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

// ExportWorkflowDraft
//
//	@Tags			workflow
//	@Summary		导出Workflow草稿
//	@Description	导出工作流草稿的json文件
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workflow_id	query		string	true	"工作流ID"
//	@Success		200			{object}	response.Response{}
//	@Router			/appspace/workflow/export/draft [get]
func ExportWorkflowDraft(ctx *gin.Context) {
	fileName := "workflow_export.json"
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

// ImportWorkflow
//
//	@Tags			workflow
//	@Summary		导入Workflow
//	@Description	通过JSON文件导入工作流
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"工作流JSON文件"
//	@Success		200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router			/appspace/workflow/import [post]
func ImportWorkflow(ctx *gin.Context) {
	resp, err := service.ImportWorkflow(ctx, getOrgID(ctx), constant.AppTypeWorkflow)
	gin_util.Response(ctx, resp, err)
}

// GetWorkflowToolSelect
//
//	@Tags		workflow
//	@Summary	工具列表（用于workflow）
//	@Description工具列表（用于workflow）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		toolType	query		string	false	"工具类型，传空或不传时返回全部"	Enums(builtin,custom)
//	@Param		name		query		string	false	"工具名称"
//	@Success	200			{object}	response.Response{data=response.ListResult{list=[]response.ToolSelect4Workflow}}
//	@Router		/workflow/tool/select [get]
func GetWorkflowToolSelect(ctx *gin.Context) {
	tools, err := service.GetWorkflowToolSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("toolType"), ctx.Query("name"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, tools, err)
}

// GetWorkflowToolDetail
//
//	@Tags		workflow
//	@Summary	工具具体action（用于workflow）
//	@Description工具具体action（用于workflow）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		toolType	query		string	true	"工具类型"	Enums(builtin,custom)
//	@Param		actionName	query		string	true	"工具具体action名称"
//	@Param		toolId		query		string	true	"工具ID"
//	@Success	200			{object}	response.Response{data=response.ToolDetail4Workflow}
//	@Router		/workflow/tool/action [get]
func GetWorkflowToolDetail(ctx *gin.Context) {
	data, err := service.GetWorkflowToolDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("toolId"), ctx.Query("toolType"), ctx.Query("actionName"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, data, err)
}

// GetWorkflowToolBoxDetail
//
//	@Tags			workflow
//	@Summary		工具箱明细查询（已解析 schema），对接本体智能体专用接口
//	@Description按	box_id + box_type 查询工具箱内的 action 明细，schema 已自动解析展开为 tools[]；支持 tool_id 精确过滤
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			box_id		query		string	false	"工具箱id（对应 toolSquareId 或 customToolId）"
//	@Param			box_type	query		string	false	"工具箱类型：builtin / custom"
//	@Param			tool_id		query		string	false	"action 的 operationId，传了则精确过滤"
//	@Param			page		query		int		false	"当前页码"
//	@Param			page_size	query		int		false	"每页显示数量"
//	@Param			status		query		string	false	"工具集状态过滤"
//	@Success		200			{object}	response.Response{data=response.ToolBoxDetail}
//	@Router			/workflow/tool/box [get]
func GetWorkflowToolBoxDetail(ctx *gin.Context) {
	var req request.ToolBoxDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetToolBoxDetail(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// GetWorkflowSelect
//
//	@Tags		workflow
//	@Summary	智能体工作流列表
//	@Description智能体工作流列表
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		name	query		string	false	"workflow名称"
//	@Success	200		{object}	response.Response{data=response.ListResult{list=[]response.ExplorationAppInfo}}
//	@Router		/workflow/select [get]
func GetWorkflowSelect(ctx *gin.Context) {
	req := request.GetExplorationAppListRequest{
		Name: ctx.Query("name"),
	}
	resp, err := service.GetWorkflowSelect(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// CreateWorkflowByTemplate
//
//	@Tags		workflow
//	@Summary	复制工作流模板
//	@Description复制工作流模板
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.CreateWorkflowByTemplateReq	true	"通过模板创建Workflow的请求参数"
//	@Success	200		{object}	response.Response{data=response.CozeWorkflowIDData}
//	@Router		/workflow/template [post]
func CreateWorkflowByTemplate(ctx *gin.Context) {
	var req request.CreateWorkflowByTemplateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateWorkflowByTemplate(ctx, getOrgID(ctx), getClientID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// WorkflowConvert
//
//	@Tags			workflow
//	@Summary		workflow转为chatflow
//	@Description	workflow转为chatflow
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.WorkflowConvertReq	true	"对话流工作流转换参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/appspace/workflow/convert [post]
func WorkflowConvert(ctx *gin.Context) {
	var req request.WorkflowConvertReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.WorkflowConvert(ctx, getOrgID(ctx), req.WorkflowID, constant.AppTypeChatflow))
}

// PublishedWorkflowRun
//
//	@Tags			workflow
//	@Summary		已发布工作流运行接口
//	@Description	已发布工作流运行接口
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.WorkflowRunReq	true	"工作流运行参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/workflow/run [post]
func PublishedWorkflowRun(ctx *gin.Context) {
	var req request.WorkflowRunReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.PublishedWorkflowRun(ctx, getUserID(ctx), getOrgID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, resp, nil)
}
