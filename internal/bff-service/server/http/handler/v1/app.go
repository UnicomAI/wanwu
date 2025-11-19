package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// DeleteAppSapceApp
//
//	@Tags			app
//	@Summary		刪除应用 [EN] @Summary Delete app
//	@Description	刪除智能体、工作流、文本问答等应用 [EN] @Description Delete applications such as agents, workflows, and text Q&A
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteAppSpaceAppRequest	true	"删除应用空间App参数" [EN] @Param data body request.DeleteAppSpaceAppRequest true "Delete application space App parameters"
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

// GetAppSpaceAppList
//
//	@Tags			app
//	@Summary		获取应用列表 [EN] @Summary Get the application list
//	@Description	获取智能体、工作流、文本问答等应用 [EN] @Description Get applications such as agents, workflows, and text question and answer
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"应用名(模糊查询)" [EN] @Param name query string false "Application name (fuzzy query)"
//	@Param			appType	query		string	false	"应用类型 Enums(agent,workflow,rag,chatflow)" [EN] @Param appType query string false "Application type Enums(agent,workflow,rag,chatflow)"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.AppBriefInfo}}
//	@Router			/appspace/app/list [get]
func GetAppSpaceAppList(ctx *gin.Context) {
	var req request.GetAppSpaceAppListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAppSpaceAppList(ctx, getUserID(ctx), getOrgID(ctx), req.Name, req.AppType)
	gin_util.Response(ctx, resp, err)
}

// PublishApp
//
//	@Tags			app
//	@Summary		发布应用 [EN] @Summary Publish the app
//	@Description	发布应用 [EN] @Description publish application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.PublishAppRequest	true	"发布应用参数" [EN] @Param data body request.PublishAppRequest true "Publish application parameters"
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
//	@Summary		取消发布应用 [EN] @Summary Unpublish app
//	@Description	取消发布应用 [EN] @Description Unpublish application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UnPublishAppRequest	true	"取消发布应用参数" [EN] @Param data body request.UnPublishAppRequest true "Unpublish application parameters"
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
