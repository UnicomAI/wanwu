package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerTool(apiV1 *gin.RouterGroup) {
	// 自定义工具 [EN] Custom tools
	mid.Sub("tool").Reg(apiV1, "/tool/custom", http.MethodPost, v1.CreateCustomTool, "创建自定义工具")
	mid.Sub("tool").Reg(apiV1, "/tool/custom", http.MethodGet, v1.GetCustomTool, "获取自定义工具详情")
	mid.Sub("tool").Reg(apiV1, "/tool/custom", http.MethodDelete, v1.DeleteCustomTool, "删除自定义工具")
	mid.Sub("tool").Reg(apiV1, "/tool/custom", http.MethodPut, v1.UpdateCustomTool, "修改自定义工具")
	mid.Sub("tool").Reg(apiV1, "/tool/custom/list", http.MethodGet, v1.GetCustomToolList, "获取自定义工具列表")
	mid.Sub("tool").Reg(apiV1, "/tool/custom/schema", http.MethodPost, v1.GetCustomToolActions, "获取可用API列表（根据Schema）")

	// 内置工具 [EN] Built-in tools
	mid.Sub("tool").Reg(apiV1, "/tool/square", http.MethodGet, v1.GetToolSquareDetail, "获取内置工具详情")
	mid.Sub("tool").Reg(apiV1, "/tool/square/list", http.MethodGet, v1.GetToolSquareList, "获取内置工具列表")
	mid.Sub("tool").Reg(apiV1, "/tool/builtin", http.MethodPost, v1.UpdateToolSquareAPIKey, "修改内置工具")

	// 自定义工具与内置工具 [EN] Custom vs. built-in tools
	mid.Sub("tool").Reg(apiV1, "/tool/select", http.MethodGet, v1.GetToolSelect, "智能体工具下拉列表（自定义与内置）")
	mid.Sub("tool").Reg(apiV1, "/tool/action/list", http.MethodGet, v1.GetToolActionList, "智能体工具action下拉列表（自定义与内置）")
	mid.Sub("tool").Reg(apiV1, "/tool/action/detail", http.MethodGet, v1.GetToolActionDetail, "智能体工具action详情（自定义与内置）")

	// MCP
	mid.Sub("tool").Reg(apiV1, "/mcp", http.MethodPost, v1.CreateMCP, "创建自定义MCP")
	mid.Sub("tool").Reg(apiV1, "/mcp", http.MethodPut, v1.UpdateMCP, "修改自定义MCP")
	mid.Sub("tool").Reg(apiV1, "/mcp", http.MethodGet, v1.GetMCP, "获取自定义MCP详情")
	mid.Sub("tool").Reg(apiV1, "/mcp", http.MethodDelete, v1.DeleteMCP, "删除自定义MCP")
	mid.Sub("tool").Reg(apiV1, "/mcp/list", http.MethodGet, v1.GetMCPList, "获取MCP自定义列表")
	mid.Sub("tool").Reg(apiV1, "/mcp/tool/list", http.MethodGet, v1.GetMCPTools, "获取MCP Tool列表")

	// MCP Server
	mid.Sub("tool").Reg(apiV1, "/mcp/server", http.MethodPost, v1.CreateMCPServer, "创建MCP服务")
	mid.Sub("tool").Reg(apiV1, "/mcp/server", http.MethodGet, v1.GetMCPServer, "获取MCP服务详情")
	mid.Sub("tool").Reg(apiV1, "/mcp/server", http.MethodPut, v1.UpdateMCPServer, "更新MCP服务")
	mid.Sub("tool").Reg(apiV1, "/mcp/server", http.MethodDelete, v1.DeleteMCPServer, "删除MCP服务")
	mid.Sub("tool").Reg(apiV1, "/mcp/server/list", http.MethodGet, v1.GetMCPServerList, "获取MCP服务列表")
	mid.Sub("tool").Reg(apiV1, "/mcp/server/tool", http.MethodPost, v1.CreateMCPServerTool, "创建MCP服务工具")
	mid.Sub("tool").Reg(apiV1, "/mcp/server/tool", http.MethodPut, v1.UpdateMCPServerTool, "更新MCP服务工具")
	mid.Sub("tool").Reg(apiV1, "/mcp/server/tool", http.MethodDelete, v1.DeleteMCPServerTool, "删除MCP服务工具")
	mid.Sub("tool").Reg(apiV1, "/mcp/server/tool/openapi", http.MethodPost, v1.CreateMCPServerOpenAPITool, "创建openapi工具")

	// 自定义MCP与MCP Server [EN] Custom MCP and MCP Server
	mid.Sub("tool").Reg(apiV1, "/mcp/select", http.MethodGet, v1.GetMCPSelect, "智能体mcp下拉列表")
	mid.Sub("tool").Reg(apiV1, "/mcp/action/list", http.MethodGet, v1.GetMCPActionList, "获取MCP Tool列表")

	// Custom Prompt
	mid.Sub("tool").Reg(apiV1, "/prompt/custom", http.MethodPost, v1.CreateCustomPrompt, "创建自定义Prompt")
	mid.Sub("tool").Reg(apiV1, "/prompt/custom", http.MethodGet, v1.GetCustomPrompt, "获取自定义Prompt详情")
	mid.Sub("tool").Reg(apiV1, "/prompt/custom", http.MethodDelete, v1.DeleteCustomPrompt, "删除自定义Prompt")
	mid.Sub("tool").Reg(apiV1, "/prompt/custom", http.MethodPut, v1.UpdateCustomPrompt, "修改自定义Prompt")
	mid.Sub("tool").Reg(apiV1, "/prompt/custom/list", http.MethodGet, v1.GetCustomPromptList, "获取自定义Prompt列表")
	mid.Sub("tool").Reg(apiV1, "/prompt/custom/copy", http.MethodPost, v1.CopyCustomPrompt, "复制自定义Prompt")
	mid.Sub("tool").Reg(apiV1, "/prompt/template", http.MethodPost, v1.CreatePromptByTemplate, "复制提示词模板")
	mid.Sub("tool").Reg(apiV1, "/prompt/optimize", http.MethodPost, v1.GetPromptOptimize, "提示词优化")

}
