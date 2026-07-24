package request

import "github.com/UnicomAI/wanwu/pkg/util"

type MCPIDReq struct {
	MCPID string `json:"mcpId" validate:"required"`
}

func (req *MCPIDReq) Check() error {
	return nil
}

type MCPCreate struct {
	Avatar        Avatar                  `json:"avatar"`                   // 图标
	MCPSquareID   string                  `json:"mcpSquareId"`              // 广场mcpId(非空表示来源于广场)
	Name          string                  `json:"name" validate:"required"` // 名称
	Desc          string                  `json:"desc" validate:"required"` // 描述
	From          string                  `json:"from" validate:"required"` // 来源
	SSEURL        string                  `json:"sseUrl"`                   // SSE URL
	StreamableURL string                  `json:"streamableUrl"`            // Streamable HTTP URL
	Transport     string                  `json:"transport"`                // 传输协议: "sse" 或 "streamable"
	ApiAuth       *util.ApiAuthWebRequest `json:"apiAuth"`                  // api身份认证(可选，为空表示无鉴权)
	Headers       map[string]string       `json:"headers"`                  // 请求头
}

func (req *MCPCreate) Check() error {
	if err := util.ValidateNameAllowSpace(&req.Name, util.SubjectMCP); err != nil {
		return err
	}
	return util.ValidateDesc(&req.Desc, util.SubjectMCP)
}

type MCPUpdate struct {
	Avatar        Avatar                  `json:"avatar"` // 图标
	MCPID         string                  `json:"mcpId" validate:"required"`
	Name          string                  `json:"name" validate:"required"` // 名称
	Desc          string                  `json:"desc" validate:"required"` // 描述
	From          string                  `json:"from" validate:"required"` // 来源
	SSEURL        string                  `json:"sseUrl"`                   // SSE URL
	StreamableURL string                  `json:"streamableUrl"`            // Streamable HTTP URL
	Transport     string                  `json:"transport"`                // 传输协议: "sse" 或 "streamable"
	ApiAuth       *util.ApiAuthWebRequest `json:"apiAuth"`                  // api身份认证(可选，为空表示无鉴权)
	Headers       map[string]string       `json:"headers"`                  // 请求头
}

func (req *MCPUpdate) Check() error {
	return nil
}

type MCPActionListReq struct {
	ToolId   string `form:"toolId" json:"toolId" validate:"required"`
	ToolType string `form:"toolType" json:"toolType" validate:"required,oneof=mcp mcpserver"`
}

func (req *MCPActionListReq) Check() error {
	return nil
}

type MCPActionReq struct {
	ToolId     string `json:"toolId" validate:"required"`
	ToolType   string `json:"toolType" validate:"required,oneof=mcp mcpserver"`
	ActionName string `json:"actionName" validate:"required"`
}

func (req *MCPActionReq) Check() error {
	return nil
}

type MCPToolListReq struct {
	MCPID     string                  `json:"mcpId"`     // mcpId
	Type      string                  `json:"type"`      // mcp/mcpserver
	ServerURL string                  `json:"serverUrl"` // "serverUrl,就是sseUrl/streamable(和mcpId、type 传一个)"
	Transport string                  `json:"transport"` // 传输协议: "sse" 或 "streamable"
	ApiAuth   *util.ApiAuthWebRequest `json:"apiAuth"`   // api身份认证
	Headers   map[string]string       `json:"headers"`   // 请求头
}

func (req *MCPToolListReq) Check() error {
	return nil
}
