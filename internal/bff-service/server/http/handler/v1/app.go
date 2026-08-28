package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// DeleteAppSapceApp
//
//	@Tags			app
//	@Summary		刪除应用
//	@Description	刪除智能体、工作流、文本问答等应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteAppSpaceAppRequest	true	"删除应用空间App参数"
//	@Success		200		{object}	response.Response{}
//	@Router			/appspace/app [delete]
func DeleteAppSapceApp(ctx *gin.Context) {
	var req request.DeleteAppSpaceAppRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteAppSpaceApp(ctx, getUserID(ctx), getOrgID(ctx), req.AppId, req.AppType)
	gin_util.Response(ctx, nil, err)
}

// GetWorkflowAndChatflowList
//
//	@Tags			workflow
//	@Summary		获取工作流和对话流列表
//	@Description	获取工作流和对话流列表，不传appType返回全部，传workflow只返回工作流，传chatflow只返回对话流
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"名称(模糊查询)"
//	@Param			appType	query		string	false	"类型: workflow/chatflow, 不传返回全部"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AppBriefInfo}}
//	@Router			/appspace/workflow/list [get]
func GetWorkflowAndChatflowList(ctx *gin.Context) {
	var req request.GetAppSpaceAppListByTypeRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetWorkflowAndChatflowList(ctx, getUserID(ctx), getOrgID(ctx), req.Name, req.AppType)
	gin_util.Response(ctx, resp, err)
}

// GetRagAppSpaceAppList
//
//	@Tags			rag
//	@Summary		获取文本问答应用列表
//	@Description	获取文本问答应用列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"应用名(模糊查询)"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AppBriefInfo}}
//	@Router			/appspace/rag/list [get]
func GetRagAppSpaceAppList(ctx *gin.Context) {
	var req request.GetAppSpaceAppListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppSpaceAppList(ctx, getUserID(ctx), getOrgID(ctx), req.Name, constant.AppTypeRag)
	gin_util.Response(ctx, resp, err)
}

// GetAgentAppSpaceAppList
//
//	@Tags			assistant
//	@Summary		获取智能体应用列表
//	@Description	获取智能体应用列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"应用名(模糊查询)"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AppBriefInfo}}
//	@Router			/appspace/assistant/list [get]
func GetAgentAppSpaceAppList(ctx *gin.Context) {
	var req request.GetAppSpaceAppListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppSpaceAppList(ctx, getUserID(ctx), getOrgID(ctx), req.Name, constant.AppTypeAgent)
	gin_util.Response(ctx, resp, err)
}

// PublishApp
//
//	@Tags			app
//	@Summary		发布应用
//	@Description	发布应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.PublishAppRequest	true	"发布应用参数"
//	@Success		200		{object}	response.Response
//	@Router			/appspace/app/publish [post]
func PublishApp(ctx *gin.Context) {
	var req request.PublishAppRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.PublishApp(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// UnPublishApp
//
//	@Tags			app
//	@Summary		取消发布应用
//	@Description	取消发布应用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UnPublishAppRequest	true	"取消发布应用参数"
//	@Success		200		{object}	response.Response
//	@Router			/appspace/app/publish [delete]
func UnPublishApp(ctx *gin.Context) {
	var req request.UnPublishAppRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UnPublishApp(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}
