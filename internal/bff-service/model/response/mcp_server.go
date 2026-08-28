package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// MCPServerInfo MCP Server信息
type MCPServerInfo struct {
	MCPServerID string         `json:"mcpServerId"` // mcpServerId
	Avatar      request.Avatar `json:"avatar"`      // 图标
	Name        string         `json:"name"`        // 名称
	Desc        string         `json:"desc"`        // 描述
	ToolNum     int64          `json:"toolNum"`     // 绑定工具数量
}

// MCPServerDetail MCP Server详情
type MCPServerDetail struct {
	MCPServerID       string              `json:"mcpServerId"`       // mcpServerId
	Avatar            request.Avatar      `json:"avatar"`            // 图标
	Name              string              `json:"name"`              // 名称
	Desc              string              `json:"desc"`              // 描述
	SSEURL            string              `json:"sseUrl"`            // sse url
	SSEExample        string              `json:"sseExample"`        // sse连接示例
	StreamableURL     string              `json:"streamableUrl"`     // streamable http url
	StreamableExample string              `json:"streamableExample"` // streamable http 连接示例
	Tools             []MCPServerToolInfo `json:"tools"`             // 绑定工具列表
	Transport         string              `json:"transport"`         // sse/streamable
}

// MCPServerToolInfo MCP Server 绑定工具信息
type MCPServerToolInfo struct {
	MCPServerToolID string `json:"mcpServerToolId"` // mcpServerToolId
	MethodName      string `json:"methodName"`      // 显示名称
	Type            string `json:"type"`            // 类型
	Id              string `json:"id"`              // 应用或工具id
	Name            string `json:"name"`            // 应用或工具名称
	Desc            string `json:"desc"`            // 描述
}

// MCPServerCreateResp MCP Server ID
type MCPServerCreateResp struct {
	MCPServerID string `json:"mcpServerId"` // mcpServerId
}

// MCPServerCustomToolSelect MCP Server自定义工具选择列表
type MCPServerCustomToolSelect struct {
	UniqueId     string                   `json:"uniqueId"`     // 统一的id
	CustomToolId string                   `json:"customToolId"` // 自定义工具id
	Name         string                   `json:"name"`         // 名称
	Description  string                   `json:"description"`  // 描述
	Methods      []MCPServerCustomToolApi `json:"methods"`      // 方法
}

type MCPServerCustomToolApi struct {
	MethodName  string `json:"methodName"`  // 方法名称
	Description string `json:"description"` // 方法描述
}
