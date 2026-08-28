package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateMCP
//
//	@Tags			mcp
//	@Summary		创建自定义MCP
//	@Description	创建自定义MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPCreate	true	"自定义MCP信息"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [post]
func CreateMCP(ctx *gin.Context) {
	var req request.MCPCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.CreateMCP(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// UpdateMCP
//
//	@Tags			mcp
//	@Summary		修改自定义MCP
//	@Description	修改自定义MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPUpdate	true	"自定义MCP信息"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [put]
func UpdateMCP(ctx *gin.Context) {
	var req request.MCPUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateMCP(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetMCP
//
//	@Tags			mcp
//	@Summary		获取自定义MCP详情
//	@Description	获取自定义MCP详情
//	@Accept			json
//	@Produce		json
//	@Param			mcpId	query		string	true	"mcpId"
//	@Success		200		{object}	response.Response{data=response.MCPDetail}
//	@Router			/mcp [get]
func GetMCP(ctx *gin.Context) {
	resp, err := service.GetMCP(ctx, ctx.Query("mcpId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteMCP
//
//	@Tags			mcp
//	@Summary		删除自定义MCP
//	@Description	删除自定义MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPIDReq	true	"mcpId"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [delete]
func DeleteMCP(ctx *gin.Context) {
	var req request.MCPIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteMCP(ctx, getUserID(ctx), getOrgID(ctx), req.MCPID))
}

// GetMCPList
//
//	@Tags			mcp
//	@Summary		获取自定义MCP列表
//	@Description	获取自定义MCP列表
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"mcp名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPInfo}}
//	@Router			/mcp/list [get]
func GetMCPList(ctx *gin.Context) {
	resp, err := service.GetMCPList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPTools
//
//	@Tags			mcp
//	@Summary		获取MCP Tool列表
//	@Description	获取MCP Tool列表
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPToolListReq	true	"mcp工具列表请求参数"
//	@Success		200		{object}	response.Response{data=response.MCPToolList}
//	@Router			/mcp/tool/list [post]
func GetMCPTools(ctx *gin.Context) {
	var req request.MCPToolListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetMCPToolList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// GetMCPSelect
//
//	@Tags			mcp
//	@Summary		获取自定义MCP列表（用于下拉选择）
//	@Description	获取自定义MCP列表（用于下拉选择）
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"mcp名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPSelect}}
//	@Router			/mcp/select [get]
func GetMCPSelect(ctx *gin.Context) {
	resp, err := service.GetMCPSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPActionList
//
//	@Tags			mcp
//	@Summary		获取MCP Action列表
//	@Description	获取MCP Action列表
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.MCPActionListReq	true	"mcp信息"
//	@Success		200		{object}	response.Response{data=response.MCPActionList}
//	@Router			/mcp/action/list [get]
func GetMCPActionList(ctx *gin.Context) {
	var req request.MCPActionListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetMCPActionList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}
