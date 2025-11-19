package request

import "github.com/UnicomAI/wanwu/pkg/util"

type MCPServerCreateReq struct {
	Avatar Avatar `json:"avatar"`                   // 图标 [EN] icon
	Name   string `json:"name" validate:"required"` // 名称 [EN] name
	Desc   string `json:"desc" validate:"required"` // 描述 [EN] describe
}

func (req *MCPServerCreateReq) Check() error { return nil }

type MCPServerUpdateReq struct {
	MCPServerID string `json:"mcpServerId" validate:"required"` // mcp server Id
	Avatar      Avatar `json:"avatar"`                          // 图标 [EN] icon
	Name        string `json:"name" validate:"required"`        // 名称 [EN] name
	Desc        string `json:"desc" validate:"required"`        // 描述 [EN] describe
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
	Id          string `json:"id" validate:"required"`          // 应用或工具id [EN] application or tool id
	Type        string `json:"type" validate:"required"`        // mcp server tool类型 [EN] mcp server tool type
	MethodName  string `json:"methodName" validate:"required"`  // 显示名称 [EN] display name
}

func (req *MCPServerToolCreateReq) Check() error { return nil }

type MCPServerToolUpdateReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` // mcp server tool id
	MethodName      string `json:"methodName" validate:"required"`      // 显示名称 [EN] display name
	Desc            string `json:"desc" validate:"required"`            // 描述 [EN] describe
}

func (req *MCPServerToolUpdateReq) Check() error { return nil }

type MCPServerToolIDReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` //mcp server tool id
}

func (req *MCPServerToolIDReq) Check() error { return nil }

type MCPServerOpenAPIToolCreate struct {
	MCPServerID   string                 `json:"mcpServerId" validate:"required"` // mcp server Id
	Name          string                 `json:"name" validate:"required"`        // 名称 [EN] name
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api身份认证 [EN] api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // 隐私政策 [EN] privacy policy
	MethodNames   []string               `json:"methodNames" validate:"required"` // API名称列表 [EN] API name list
}

func (req *MCPServerOpenAPIToolCreate) Check() error { return nil }
