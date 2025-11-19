package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// MCPServerInfo MCP Server信息 [EN] MCPServerInfo MCP Server information
type MCPServerInfo struct {
	MCPServerID string         `json:"mcpServerId"` // mcpServerId
	Avatar      request.Avatar `json:"avatar"`      // 图标 [EN] icon
	Name        string         `json:"name"`        // 名称 [EN] name
	Desc        string         `json:"desc"`        // 描述 [EN] describe
	ToolNum     int64          `json:"toolNum"`     // 绑定工具数量 [EN] Number of binding tools
}

// MCPServerDetail MCP Server详情 [EN] MCPServerDetail MCP ServerDetails
type MCPServerDetail struct {
	MCPServerID       string              `json:"mcpServerId"`       // mcpServerId
	Avatar            request.Avatar      `json:"avatar"`            // 图标 [EN] icon
	Name              string              `json:"name"`              // 名称 [EN] name
	Desc              string              `json:"desc"`              // 描述 [EN] describe
	SSEURL            string              `json:"sseUrl"`            // sse url
	SSEExample        string              `json:"sseExample"`        // sse连接示例 [EN] sse connection example
	StreamableURL     string              `json:"streamableUrl"`     // streamable http url
	StreamableExample string              `json:"streamableExample"` // streamable http 连接示例 [EN] streamable http connection example
	Tools             []MCPServerToolInfo `json:"tools"`             // 绑定工具列表 [EN] Binding tool list
}

// MCPServerToolInfo MCP Server 绑定工具信息 [EN] MCPServerToolInfo MCP Server binding tool information
type MCPServerToolInfo struct {
	MCPServerToolID string `json:"mcpServerToolId"` // mcpServerToolId
	MethodName      string `json:"methodName"`      // 显示名称 [EN] display name
	Type            string `json:"type"`            // 类型 [EN] type
	Id              string `json:"id"`              // 应用或工具id [EN] application or tool id
	Name            string `json:"name"`            // 应用或工具名称 [EN] Application or tool name
	Desc            string `json:"desc"`            // 描述 [EN] describe
}

// MCPServerCreateResp MCP Server ID
type MCPServerCreateResp struct {
	MCPServerID string `json:"mcpServerId"` // mcpServerId
}

// MCPServerCustomToolSelect MCP Server自定义工具选择列表 [EN] MCPServerCustomToolSelect MCP Server custom tool selection list
type MCPServerCustomToolSelect struct {
	UniqueId     string                   `json:"uniqueId"`     // 统一的id [EN] unified id
	CustomToolId string                   `json:"customToolId"` // 自定义工具id [EN] Custom tool id
	Name         string                   `json:"name"`         // 名称 [EN] name
	Description  string                   `json:"description"`  // 描述 [EN] describe
	Methods      []MCPServerCustomToolApi `json:"methods"`      // 方法 [EN] method
}

type MCPServerCustomToolApi struct {
	MethodName  string `json:"methodName"`  // 方法名称 [EN] method name
	Description string `json:"description"` // 方法描述 [EN] Method description
}
