package response

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type MCPSelect struct {
	UniqueId    string         `json:"uniqueId"`    // 随机unique id(每次动态生成) [EN] Random unique id (dynamically generated each time)
	MCPID       string         `json:"mcpId"`       // mcpId
	MCPSquareID string         `json:"mcpSquareId"` // 广场mcpId(非空表示来源于广场) [EN] Square mcpId (non-empty means it comes from square)
	Name        string         `json:"name"`        // 名称 [EN] name
	Type        string         `json:"type"`
	ToolId      string         `json:"toolId"`                                           // 工具id [EN] tool id
	ToolName    string         `json:"toolName"`                                         // 工具名称 [EN] Tool name
	ToolType    string         `json:"toolType" validate:"required,oneof=mcp mcpserver"` // 工具类型 [EN] Tool type
	Description string         `json:"description"`                                      // 描述 [EN] describe
	ServerFrom  string         `json:"serverFrom"`                                       // 来源 [EN] source
	ServerURL   string         `json:"serverUrl"`                                        // sseUrl
	Avatar      request.Avatar `json:"avatar"`                                           // 图标 [EN] icon
}

type MCPToolList struct {
	Tools []*protocol.Tool `json:"tools"`
}

// MCPDetail MCP自定义详情 [EN] MCPDetail MCP custom details
type MCPDetail struct {
	MCPInfo
	MCPSquareIntro
}

// MCPInfo MCP自定义信息 [EN] MCPInfo MCP custom information
type MCPInfo struct {
	MCPID  string `json:"mcpId"`  // mcpId
	SSEURL string `json:"sseUrl"` // SSE URL
	MCPSquareInfo
}

// MCPSquareDetail MCP广场详情 [EN] MCPSquareDetail MCP square details
type MCPSquareDetail struct {
	MCPSquareInfo
	MCPSquareIntro
	MCPActions
}

// MCPSquareInfo MCP广场信息 [EN] MCPSquareInfo MCP square information
type MCPSquareInfo struct {
	MCPSquareID string         `json:"mcpSquareId"` // 广场mcpId(非空表示来源于广场) [EN] Square mcpId (non-empty means it comes from square)
	Avatar      request.Avatar `json:"avatar"`      // 图标 [EN] icon
	Name        string         `json:"name"`        // 名称 [EN] name
	Desc        string         `json:"desc"`        // 描述 [EN] describe
	From        string         `json:"from"`        // 来源 [EN] source
	Category    string         `json:"category"`    // 类型(data:数据,create:创作,search:搜索) [EN] Type (data: data, create: creation, search: search)
}

type MCPSquareIntro struct {
	Summary  string `json:"summary"`  // 使用概述 [EN] Usage overview
	Feature  string `json:"feature"`  // 特性说明 [EN] Feature description
	Scenario string `json:"scenario"` // 应用场景 [EN] Application scenarios
	Manual   string `json:"manual"`   // 使用说明 [EN] Instructions for use
	Detail   string `json:"detail"`   // 详情 [EN] Details
}

type MCPActions struct {
	SSEURL    string           `json:"sseUrl"`    // SSE URL
	Tools     []*protocol.Tool `json:"tools"`     // 工具列表 [EN] Tool list
	HasCustom bool             `json:"hasCustom"` // 是否已经发送到自定义 [EN] Has it been sent to custom
}

type MCPActionList struct {
	Actions []*protocol.Tool `json:"actions"`
}
