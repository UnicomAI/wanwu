package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetMCP
//
//	@Tags			mcp
//	@Summary		获取自定义MCP详情
//	@Description	获取自定义MCP详情
//	@Accept			json
//	@Produce		json
//	@Param			mcpId	query		string	true	"mcpId"
//	@Success		200		{object}	response.Response{data=response.MCPDetail}
//	@Router			/callback/mcp [get]
func GetMCP(ctx *gin.Context) {
	resp, err := service.GetMCPForRuntime(ctx, ctx.Query("mcpId"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPServer
//
//	@Tags			mcp.server
//	@Summary		获取MCP Server详情
//	@Description	获取MCP Server详情
//	@Accept			json
//	@Produce		json
//	@Param			mcpServerId	query		string	true	"mcpServerId"
//	@Success		200			{object}	response.Response{data=response.MCPServerDetail}
//	@Router			/callback/mcp/server [get]
func GetMCPServer(ctx *gin.Context) {
	resp, err := service.GetMCPServerDetail(ctx, ctx.Query("mcpServerId"))
	gin_util.Response(ctx, resp, err)
}
