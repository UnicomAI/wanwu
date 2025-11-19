package response

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
)

type MCPSelect struct {
	UniqueId    string         `json:"uniqueId"`    // Random unique id (dynamically generated each time)
	MCPID       string         `json:"mcpId"`       // mcpId
	MCPSquareID string         `json:"mcpSquareId"` // Square mcpId (non-empty means it comes from square)
	Name        string         `json:"name"`        // name
	Type        string         `json:"type"`
	ToolId      string         `json:"toolId"`                                           // tool id
	ToolName    string         `json:"toolName"`                                         // Tool name
	ToolType    string         `json:"toolType" validate:"required,oneof=mcp mcpserver"` // Tool type
	Description string         `json:"description"`                                      // describe
	ServerFrom  string         `json:"serverFrom"`                                       // source
	ServerURL   string         `json:"serverUrl"`                                        // sseUrl
	Avatar      request.Avatar `json:"avatar"`                                           // icon
}

type MCPToolList struct {
	Tools []*protocol.Tool `json:"tools"`
}

// MCPDetail MCP custom details
type MCPDetail struct {
	MCPInfo
	MCPSquareIntro
}

// MCPInfo MCP custom information
type MCPInfo struct {
	MCPID  string `json:"mcpId"`  // mcpId
	SSEURL string `json:"sseUrl"` // SSE URL
	MCPSquareInfo
}

// MCPSquareDetail MCP square details
type MCPSquareDetail struct {
	MCPSquareInfo
	MCPSquareIntro
	MCPActions
}

// MCPSquareInfo MCP square information
type MCPSquareInfo struct {
	MCPSquareID string         `json:"mcpSquareId"` // Square mcpId (non-empty means it comes from square)
	Avatar      request.Avatar `json:"avatar"`      // icon
	Name        string         `json:"name"`        // name
	Desc        string         `json:"desc"`        // describe
	From        string         `json:"from"`        // source
	Category    string         `json:"category"`    // Type (data: data, create: creation, search: search)
}

type MCPSquareIntro struct {
	Summary  string `json:"summary"`  // Usage overview
	Feature  string `json:"feature"`  // Feature description
	Scenario string `json:"scenario"` // Application scenarios
	Manual   string `json:"manual"`   // Instructions for use
	Detail   string `json:"detail"`   // Details
}

type MCPActions struct {
	SSEURL    string           `json:"sseUrl"`    // SSE URL
	Tools     []*protocol.Tool `json:"tools"`     // Tool list
	HasCustom bool             `json:"hasCustom"` // Has it been sent to custom
}

type MCPActionList struct {
	Actions []*protocol.Tool `json:"actions"`
}
