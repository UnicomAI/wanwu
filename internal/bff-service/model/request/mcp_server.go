package request

import "github.com/UnicomAI/wanwu/pkg/util"

type MCPServerCreateReq struct {
	Avatar Avatar `json:"avatar"`                   // 图标
	Name   string `json:"name" validate:"required"` // 名称
	Desc   string `json:"desc" validate:"required"` // 描述
}

func (req *MCPServerCreateReq) Check() error {
	if err := util.ValidateName(&req.Name, util.SubjectMCPServer); err != nil {
		return err
	}
	return util.ValidateDesc(&req.Desc, util.SubjectMCPServer)
}

type MCPServerUpdateReq struct {
	MCPServerID string `json:"mcpServerId" validate:"required"` // mcp server Id
	Avatar      Avatar `json:"avatar"`                          // 图标
	Name        string `json:"name" validate:"required"`        // 名称
	Desc        string `json:"desc" validate:"required"`        // 描述
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
	Id          string `json:"id" validate:"required"`          // 应用或工具id
	Type        string `json:"type" validate:"required"`        // mcp server tool类型
	MethodName  string `json:"methodName" validate:"required"`  // 显示名称
}

func (req *MCPServerToolCreateReq) Check() error { return nil }

type MCPServerToolUpdateReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` // mcp server tool id
	MethodName      string `json:"methodName" validate:"required"`      // 显示名称
	Desc            string `json:"desc" validate:"required"`            // 描述
}

func (req *MCPServerToolUpdateReq) Check() error { return nil }

type MCPServerToolIDReq struct {
	MCPServerToolID string `json:"mcpServerToolId" validate:"required"` //mcp server tool id
}

func (req *MCPServerToolIDReq) Check() error { return nil }

type MCPServerOpenAPIToolCreate struct {
	MCPServerID   string                 `json:"mcpServerId" validate:"required"` // mcp server Id
	Name          string                 `json:"name" validate:"required"`        // 名称
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api身份认证
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // 隐私政策
	MethodNames   []string               `json:"methodNames" validate:"required"` // API名称列表
}

func (req *MCPServerOpenAPIToolCreate) Check() error { return nil }
