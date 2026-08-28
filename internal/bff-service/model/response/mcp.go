package response

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type MCPSelect struct {
	UniqueId      string                 `json:"uniqueId"`    // 随机unique id(每次动态生成)
	MCPID         string                 `json:"mcpId"`       // mcpId
	MCPSquareID   string                 `json:"mcpSquareId"` // 广场mcpId(非空表示来源于广场)
	Name          string                 `json:"name"`        // 名称
	Type          string                 `json:"type"`
	ToolId        string                 `json:"toolId"`                                           // 工具id
	ToolName      string                 `json:"toolName"`                                         // 工具名称
	ToolType      string                 `json:"toolType" validate:"required,oneof=mcp mcpserver"` // 工具类型
	Description   string                 `json:"description"`                                      // 描述
	ServerFrom    string                 `json:"serverFrom"`                                       // 来源
	ServerURL     string                 `json:"serverUrl"`                                        // sseUrl
	StreamableURL string                 `json:"streamableUrl"`                                    // streamableUrl
	Transport     string                 `json:"transport"`                                        // 传输协议
	Avatar        request.Avatar         `json:"avatar"`                                           // 图标
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`                      // api身份认证
	Headers       map[string]string      `json:"headers"`                                          // 请求头
}

type MCPToolList struct {
	Tools []*protocol.Tool `json:"tools"`
}

// MCPDetail MCP自定义详情
type MCPDetail struct {
	MCPInfo
	MCPSquareIntro
}

// MCPInfo MCP自定义信息
type MCPInfo struct {
	MCPID         string                 `json:"mcpId"`         // mcpId
	Type          string                 `json:"type"`          // type: mcp
	SSEURL        string                 `json:"sseUrl"`        // SSE URL
	StreamableURL string                 `json:"streamableUrl"` // Streamable HTTP URL
	Transport     string                 `json:"transport"`     // 传输协议: "sse" 或 "streamable"
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth"`       // api身份认证
	Headers       map[string]string      `json:"headers"`       // 请求头
	MCPSquareInfo
}

// MCPSquareDetail MCP广场详情
type MCPSquareDetail struct {
	MCPSquareInfo
	MCPSquareIntro
	MCPActions
}

// MCPSquareInfo MCP广场信息
type MCPSquareInfo struct {
	MCPSquareID string         `json:"mcpSquareId"` // 广场mcpId(非空表示来源于广场)
	Avatar      request.Avatar `json:"avatar"`      // 图标
	Name        string         `json:"name"`        // 名称
	Desc        string         `json:"desc"`        // 描述
	From        string         `json:"from"`        // 来源
	Category    string         `json:"category"`    // 类型(data:数据,create:创作,search:搜索)
}

type MCPSquareIntro struct {
	Summary  string `json:"summary"`  // 使用概述
	Feature  string `json:"feature"`  // 特性说明
	Scenario string `json:"scenario"` // 应用场景
	Manual   string `json:"manual"`   // 使用说明
	Detail   string `json:"detail"`   // 详情
}

type MCPActions struct {
	SSEURL    string           `json:"sseUrl"`    // SSE URL
	Tools     []*protocol.Tool `json:"tools"`     // 工具列表
	HasCustom bool             `json:"hasCustom"` // 是否已经发送到自定义
}

type MCPActionList struct {
	Actions []*protocol.Tool `json:"actions"`
}
