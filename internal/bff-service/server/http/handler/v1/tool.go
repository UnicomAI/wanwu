package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateCustomTool
//
//	@Tags			tool.custom
//	@Summary		创建自定义工具 [EN] @Summary Create a custom tool
//	@Description	创建自定义工具 [EN] @Description Create a custom tool
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomToolCreate	true	"自定义工具信息" [EN] @Param data body request.CustomToolCreate true "Custom tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/tool/custom [post]
func CreateCustomTool(ctx *gin.Context) {
	var req request.CustomToolCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.CreateCustomTool(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetCustomTool
//
//	@Tags			tool.custom
//	@Summary		获取自定义工具详情 [EN] @Summary Get custom tool details
//	@Description	获取自定义工具详情 [EN] @Description Get custom tool details
//	@Accept			json
//	@Produce		json
//	@Param			customToolId	query		string	true	"customToolId"
//	@Success		200				{object}	response.Response{data=response.CustomToolDetail}
//	@Router			/tool/custom [get]
func GetCustomTool(ctx *gin.Context) {
	resp, err := service.GetCustomTool(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("customToolId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteCustomTool
//
//	@Tags			tool.custom
//	@Summary		删除自定义工具 [EN] @Summary Delete custom tools
//	@Description	删除自定义工具 [EN] @Description Delete custom tool
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomToolIDReq	true	"自定义工具ID" [EN] @Param data body request.CustomToolIDReq true "Custom tool ID"
//	@Success		200		{object}	response.Response{}
//	@Router			/tool/custom [delete]
func DeleteCustomTool(ctx *gin.Context) {
	var req request.CustomToolIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteCustomTool(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// UpdateCustomTool
//
//	@Tags			tool.custom
//	@Summary		修改自定义工具 [EN] @Summary Modify custom tools
//	@Description	修改自定义工具 [EN] @Description Modify custom tools
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomToolUpdateReq	true	"自定义工具信息" [EN] @Param data body request.CustomToolUpdateReq true "Custom tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/tool/custom [put]
func UpdateCustomTool(ctx *gin.Context) {
	var req request.CustomToolUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateCustomTool(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetCustomToolList
//
//	@Tags			tool.custom
//	@Summary		获取自定义工具列表 [EN] @Summary Get the list of custom tools
//	@Description	获取自定义工具列表 [EN] @Description Get the list of custom tools
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"CustomTool名称" [EN] @Param name query string false "CustomTool name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.CustomToolInfo}}
//	@Router			/tool/custom/list [get]
func GetCustomToolList(ctx *gin.Context) {
	resp, err := service.GetCustomToolList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetCustomToolActions
//
//	@Tags			tool.custom
//	@Summary		获取可用API列表（根据Schema） [EN] @Summary Get the list of available APIs (according to Schema)
//	@Description	解析自定义工具的Schema转换为API相关数据 [EN] @Description Parse the Schema of the custom tool and convert it into API related data
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CustomToolSchemaReq	true	"Schema格式数据" [EN] @Param data body request.CustomToolSchemaReq true "Schema format data"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.CustomToolActionInfo}}
//	@Router			/tool/custom/schema [post]
func GetCustomToolActions(ctx *gin.Context) {
	var req request.CustomToolSchemaReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetCustomToolActions(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetToolSquareDetail
//
//	@Tags			tool.square
//	@Summary		获取内置工具详情 [EN] @Summary Get built-in tool details
//	@Description	获取内置工具详情 [EN] @Description Get built-in tool details
//	@Accept			json
//	@Produce		json
//	@Param			toolSquareId	query		string	true	"toolSquareId"
//	@Success		200				{object}	response.Response{data=response.ToolSquareDetail}
//	@Router			/tool/square [get]
func GetToolSquareDetail(ctx *gin.Context) {
	resp, err := service.GetToolSquareDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("toolSquareId"))
	gin_util.Response(ctx, resp, err)
}

// GetToolSquareList
//
//	@Tags			tool.builtin
//	@Summary		获取内置工具列表 [EN] @Summary Get the list of built-in tools
//	@Description	获取内置工具列表 [EN] @Description Get the list of built-in tools
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"tool名称" [EN] @Param name query string false "tool name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ToolSquareInfo}}
//	@Router			/tool/square/list [get]
func GetToolSquareList(ctx *gin.Context) {
	resp, err := service.GetToolSquareList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// UpdateToolSquareAPIKey
//
//	@Tags			tool.builtin
//	@Summary		修改内置工具 [EN] @Summary Modify built-in tools
//	@Description	修改内置工具 [EN] @Description Modify built-in tools
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ToolSquareAPIKeyReq	true	"内置工具信息" [EN] @Param data body request.ToolSquareAPIKeyReq true "Built-in tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/tool/builtin [post]
func UpdateToolSquareAPIKey(ctx *gin.Context) {
	var req request.ToolSquareAPIKeyReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpsertBuiltinToolAPIKey(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetToolSelect
//
//	@Tags			tool
//	@Summary		获取工具列表（用于下拉选择） [EN] @Summary Get the tool list (for drop-down selection)
//	@Description	获取工具列表（用于下拉选择） [EN] @Description Get the tool list (for drop-down selection)
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	true	"工具名" [EN] @Param name query string true "Tool name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ToolSelect}}
//	@Router			/tool/select [get]
func GetToolSelect(ctx *gin.Context) {
	resp, err := service.GetToolSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetToolActionList
//
//	@Tags			tool
//	@Summary		获取工具列表 [EN] @Summary Get the list of tools
//	@Description	获取工具列表 [EN] @Description Get the tool list
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ToolActionListReq	true	"工具信息" [EN] @Param data query request.ToolActionListReq true "Tool information"
//	@Success		200		{object}	response.Response{data=response.ToolActionList}
//	@Router			/tool/action/list [get]
func GetToolActionList(ctx *gin.Context) {
	var req request.ToolActionListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetToolActionList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetToolActionDetail
//
//	@Tags			tool
//	@Summary		获取工具详情 [EN] @Summary Get tool details
//	@Description	获取工具详情 [EN] @Description Get tool details
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ToolActionReq	true	"工具信息" [EN] @Param data query request.ToolActionReq true "Tool information"
//	@Success		200		{object}	response.Response{data=response.ToolActionDetail}
//	@Router			/tool/action/detail [get]
func GetToolActionDetail(ctx *gin.Context) {
	var req request.ToolActionReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetToolActionDetail(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}
