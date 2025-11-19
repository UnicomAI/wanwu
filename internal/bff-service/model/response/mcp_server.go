package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// MCPServerInfo MCP Server information
type MCPServerInfo struct {
	MCPServerID string         `json:"mcpServerId"` // mcpServerId
	Avatar      request.Avatar `json:"avatar"`      // icon
	Name        string         `json:"name"`        // name
	Desc        string         `json:"desc"`        // describe
	ToolNum     int64          `json:"toolNum"`     // Number of binding tools
}

// MCPServerDetail MCP ServerDetails
type MCPServerDetail struct {
	MCPServerID       string              `json:"mcpServerId"`       // mcpServerId
	Avatar            request.Avatar      `json:"avatar"`            // icon
	Name              string              `json:"name"`              // name
	Desc              string              `json:"desc"`              // describe
	SSEURL            string              `json:"sseUrl"`            // sse url
	SSEExample        string              `json:"sseExample"`        // sse connection example
	StreamableURL     string              `json:"streamableUrl"`     // streamable http url
	StreamableExample string              `json:"streamableExample"` // streamable http connection example
	Tools             []MCPServerToolInfo `json:"tools"`             // Binding tool list
}

// MCPServerToolInfo MCP Server binding tool information
type MCPServerToolInfo struct {
	MCPServerToolID string `json:"mcpServerToolId"` // mcpServerToolId
	MethodName      string `json:"methodName"`      // display name
	Type            string `json:"type"`            // type
	Id              string `json:"id"`              // application or tool id
	Name            string `json:"name"`            // Application or tool name
	Desc            string `json:"desc"`            // describe
}

// MCPServerCreateResp MCP Server ID
type MCPServerCreateResp struct {
	MCPServerID string `json:"mcpServerId"` // mcpServerId
}

// MCPServerCustomToolSelect MCP Server custom tool selection list
type MCPServerCustomToolSelect struct {
	UniqueId     string                   `json:"uniqueId"`     // unified id
	CustomToolId string                   `json:"customToolId"` // Custom tool id
	Name         string                   `json:"name"`         // name
	Description  string                   `json:"description"`  // describe
	Methods      []MCPServerCustomToolApi `json:"methods"`      // method
}

type MCPServerCustomToolApi struct {
	MethodName  string `json:"methodName"`  // method name
	Description string `json:"description"` // Method description
}
