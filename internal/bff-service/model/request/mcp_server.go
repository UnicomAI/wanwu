package request

import "github.com/UnicomAI/wanwu/pkg/util"

type MCPServerCreateReq struct {
	Avatar Avatar `json:"avatar"`                   // icon
	Name   string `json:"name" validate:"required"` // name
	Desc   string `json:"desc" validate:"required"` // describe
}

func (req *MCPServerCreateReq) Check() error { return nil }

type MCPServerUpdateReq struct {
	MCPServerID string `json:"mcpServerId" validate:"required"` // mcp server Id
	Avatar      Avatar `json:"avatar"`                          // icon
	Name        string `json:"name" validate:"required"`        // name
	Desc        string `json:"desc" validate:"required"`        // describe
}

func (req *MCPServerUpdateReq) Check() error { return nil }

type MCPServerIDReq struct {
	MCPServerID string `json:"mcpServerId" validate:"required"`
}

func (req *MCPServerIDReq) Check() error {
	return nil
}

type MCPServerToolCreateReq struct {
	MCPServerID string `json:"mcpServerId" validate:"required"` // mcp server Id
	Id          string `json:"id" validate:"required"`          // application or tool id
	Type        string `json:"type" validate:"required"`        // mcp server tool type
	MethodName  string `json:"methodName" validate:"required"`  // display name
}

func (req *MCPServerToolCreateReq) Check() error { return nil }

type MCPServerToolUpdateReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` // mcp server tool id
	MethodName      string `json:"methodName" validate:"required"`      // display name
	Desc            string `json:"desc" validate:"required"`            // describe
}

func (req *MCPServerToolUpdateReq) Check() error { return nil }

type MCPServerToolIDReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` //mcp server tool id
}

func (req *MCPServerToolIDReq) Check() error { return nil }

type MCPServerOpenAPIToolCreate struct {
	MCPServerID   string                 `json:"mcpServerId" validate:"required"` // mcp server Id
	Name          string                 `json:"name" validate:"required"`        // name
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // privacy policy
	MethodNames   []string               `json:"methodNames" validate:"required"` // API name list
}

func (req *MCPServerOpenAPIToolCreate) Check() error { return nil }
