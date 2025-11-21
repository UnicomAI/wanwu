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
//	@Summary Delete app
//	@Description Delete applications such as agents, workflows, and text Q&A
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteAppSpaceAppRequest true "Delete application space App parameters"
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
//	@Summary Get the application list
//	@Description Get applications such as agents, workflows, and text question and answer
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param name query string false "Application name (fuzzy query)"
//	@Param appType query string false "Application type Enums(agent,workflow,rag,chatflow)"
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
//	@Summary Publish the app
//	@Description publish application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.PublishAppRequest true "Publish application parameters"
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
//	@Summary Unpublish app
//	@Description Unpublish application
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UnPublishAppRequest true "Unpublish application parameters"
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
