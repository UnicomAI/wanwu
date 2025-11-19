package request

import "github.com/UnicomAI/wanwu/pkg/util"

type CustomToolCreate struct {
	Avatar        Avatar                 `json:"avatar"`                          // 图标 [EN] icon
	Name          string                 `json:"name" validate:"required"`        // 名称 [EN] name
	Description   string                 `json:"description" validate:"required"` // 描述 [EN] describe
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api身份认证 [EN] api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // 隐私政策 [EN] privacy policy
}

func (req *CustomToolCreate) Check() error { return nil }

type CustomToolUpdateReq struct {
	Avatar        Avatar                 `json:"avatar"`                           // 图标 [EN] icon
	CustomToolID  string                 `json:"customToolId" validate:"required"` // 自定义工具ID [EN] Custom tool ID
	Name          string                 `json:"name" validate:"required"`         // 名称 [EN] name
	Description   string                 `json:"description" validate:"required"`  // 描述 [EN] describe
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`      // api身份认证 [EN] api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`      // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                    // 隐私政策 [EN] privacy policy
}

func (req *CustomToolUpdateReq) Check() error { return nil }

type CustomToolIDReq struct {
	CustomToolID string `json:"customToolId" validate:"required"` // 自定义工具id [EN] Custom tool id
}

func (req *CustomToolIDReq) Check() error { return nil }

type CustomToolSchemaReq struct {
	Schema string `json:"schema" validate:"required"` // schema
}

func (req *CustomToolSchemaReq) Check() error { return nil }

type ToolSquareAPIKeyReq struct {
	ToolSquareID string `json:"toolSquareId" validate:"required"` // 广场toolId [EN] square toolId
	APIKey       string `json:"apiKey"`                           // apiKey
}

func (req *ToolSquareAPIKeyReq) Check() error { return nil }

type ToolActionListReq struct {
	ToolId   string `form:"toolId" json:"toolId" validate:"required"`                          // 工具id [EN] tool id
	ToolType string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // 工具类型 [EN] Tool type
}

func (req *ToolActionListReq) Check() error { return nil }

type ToolActionReq struct {
	ToolId     string `form:"toolId" json:"toolId" validate:"required"`                          // 工具id [EN] tool id
	ToolType   string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // 工具类型 [EN] Tool type
	ActionName string `form:"actionName" json:"actionName" validate:"required"`                  // action名称 [EN] action name
}

func (req *ToolActionReq) Check() error { return nil }
