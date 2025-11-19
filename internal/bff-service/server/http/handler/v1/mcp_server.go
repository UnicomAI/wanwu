package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateMCPServer
//
//	@Tags			mcp.server
//	@Summary Create MCP Server
//	@Description Create MCP Server
//	@Accept			json
//	@Produce		json
//	@Param data body request.MCPServerCreateReq true "MCP Server information"
//	@Success		200		{object}	response.Response{data=response.MCPServerCreateResp}
//	@Router			/mcp/server [post]
func CreateMCPServer(ctx *gin.Context) {
	var req request.MCPServerCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateMCPServer(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// UpdateMCPServer
//
//	@Tags			mcp.server
//	@Summary Update MCP Server
//	@Description Update MCP Server
//	@Accept			json
//	@Produce		json
//	@Param data body request.MCPServerUpdateReq true "MCP Server information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server [put]
func UpdateMCPServer(ctx *gin.Context) {
	var req request.MCPServerUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateMCPServer(ctx, req))
}

// GetMCPServer
//
//	@Tags			mcp.server
//	@Summary Get MCP Server details
//	@Description Get MCP Server details
//	@Accept			json
//	@Produce		json
//	@Param			mcpServerId	query		string	true	"mcpServerId"
//	@Success		200			{object}	response.Response{data=response.MCPServerDetail}
//	@Router			/mcp/server [get]
func GetMCPServer(ctx *gin.Context) {
	resp, err := service.GetMCPServerDetail(ctx, ctx.Query("mcpServerId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteMCPServer
//
//	@Tags			mcp.server
//	@Summary Delete MCP Server
//	@Description Delete MCP Server
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPServerIDReq	true	"mcpServerId"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server [delete]
func DeleteMCPServer(ctx *gin.Context) {
	var req request.MCPServerIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteMCPServer(ctx, req.MCPServerID)
	gin_util.Response(ctx, nil, err)
}

// GetMCPServerList
//
//	@Tags			mcp.server
//	@Summary Get the MCP Server list
//	@Description Get MCP Server list
//	@Accept			json
//	@Produce		json
//	@Param name query string false "mcp server name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPServerInfo}}
//	@Router			/mcp/server/list [get]
func GetMCPServerList(ctx *gin.Context) {
	resp, err := service.GetMCPServerList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// CreateMCPServerTool
//
//	@Tags			mcp.server
//	@Summary Create MCP Server tool
//	@Description Create MCP Server tool
//	@Accept			json
//	@Produce		json
//	@Param data body request.MCPServerToolCreateReq true "MCP Server tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server/tool [post]
func CreateMCPServerTool(ctx *gin.Context) {
	var req request.MCPServerToolCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.CreateMCPServerTool(ctx, req))
}

// UpdateMCPServerTool
//
//	@Tags			mcp.server
//	@Summary Update MCP Server tools
//	@Description Update MCP Server tool
//	@Accept			json
//	@Produce		json
//	@Param data body request.MCPServerToolUpdateReq true "MCP Server tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server/tool [put]
func UpdateMCPServerTool(ctx *gin.Context) {
	var req request.MCPServerToolUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateMCPServerTool(ctx, req))
}

// DeleteMCPServerTool
//
//	@Tags			mcp.server
//	@Summary Delete MCP Server tool
//	@Description Delete MCP Server tool
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPServerToolIDReq	true	"mcpServerToolId"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server/tool [delete]
func DeleteMCPServerTool(ctx *gin.Context) {
	var req request.MCPServerToolIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteMCPServerTool(ctx, req.MCPServerToolID)
	gin_util.Response(ctx, nil, err)
}

// CreateMCPServerOpenAPITool
//
//	@Tags			tool
//	@Summary Create openapi tool
//	@Description Create openapi tool
//	@Accept			json
//	@Produce		json
//	@Param data body request.MCPServerOpenAPIToolCreate true "openapi tool information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp/server/tool/openapi [post]
func CreateMCPServerOpenAPITool(ctx *gin.Context) {
	var req request.MCPServerOpenAPIToolCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateMCPServerOpenAPITool(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}
