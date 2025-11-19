package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetMCPSquareDetail
//
//	@Tags			mcp.square
//	@Summary		获取广场MCP详情 [EN] @Summary Get the details of the square MCP
//	@Description	获取广场MCP详情 [EN] @Description Get square MCP details
//	@Accept			json
//	@Produce		json
//	@Param			mcpSquareId	query		string	true	"mcpSquareId"
//	@Success		200			{object}	response.Response{data=response.MCPSquareDetail}
//	@Router			/mcp/square [get]
func GetMCPSquareDetail(ctx *gin.Context) {
	resp, err := service.GetMCPSquareDetail(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("mcpSquareId"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPSquareList
//
//	@Tags			mcp.square
//	@Summary		获取广场MCP列表 [EN] @Summary Get the square MCP list
//	@Description	获取广场MCP列表 [EN] @Description Get the square MCP list
//	@Accept			json
//	@Produce		json
//	@Param			category	query		string	false	"mcp类型"	Enums(all,data,create,search) [EN] @Param category query string false "mcp type" Enums(all,data,create,search)
//	@Param			name		query		string	false	"mcp名称" [EN] @Param name query string false "mcp name"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.MCPSquareInfo}}
//	@Router			/mcp/square/list [get]
func GetMCPSquareList(ctx *gin.Context) {
	resp, err := service.GetMCPSquareList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("category"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPSquareRecommends
//
//	@Tags			mcp.square
//	@Summary		获取广场MCP推荐列表 [EN] @Summary Get the square MCP recommendation list
//	@Description	获取广场MCP推荐列表 [EN] @Description Get the square MCP recommendation list
//	@Accept			json
//	@Produce		json
//	@Param			mcpId		query		string	false	"mcpId"
//	@Param			mcpSquareId	query		string	false	"mcpSquareId"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.MCPSquareInfo}}
//	@Router			/mcp/square/recommend [get]
func GetMCPSquareRecommends(ctx *gin.Context) {
	resp, err := service.GetMCPSquareList(ctx, getUserID(ctx), getOrgID(ctx), "", "")
	gin_util.Response(ctx, resp, err)
}
