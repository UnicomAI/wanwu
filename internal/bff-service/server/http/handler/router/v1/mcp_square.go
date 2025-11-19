package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerMCPSquare(apiV1 *gin.RouterGroup) {
	// MCP广场 [EN] MCP Plaza
	mid.Sub("mcp").Reg(apiV1, "/mcp/square", http.MethodGet, v1.GetMCPSquareDetail, "获取广场MCP详情")
	mid.Sub("mcp").Reg(apiV1, "/mcp/square/list", http.MethodGet, v1.GetMCPSquareList, "获取广场MCP列表")
	mid.Sub("mcp").Reg(apiV1, "/mcp/square/recommend", http.MethodGet, v1.GetMCPSquareRecommends, "获取广场MCP推荐列表")
}
