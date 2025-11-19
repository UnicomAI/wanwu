package request

import "github.com/UnicomAI/wanwu/pkg/util"

type CustomToolCreate struct {
	Avatar        Avatar                 `json:"avatar"`                          // icon
	Name          string                 `json:"name" validate:"required"`        // name
	Description   string                 `json:"description" validate:"required"` // describe
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // privacy policy
}

func (req *CustomToolCreate) Check() error { return nil }

type CustomToolUpdateReq struct {
	Avatar        Avatar                 `json:"avatar"`                           // icon
	CustomToolID  string                 `json:"customToolId" validate:"required"` // Custom tool ID
	Name          string                 `json:"name" validate:"required"`         // name
	Description   string                 `json:"description" validate:"required"`  // describe
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`      // api identity authentication
	Schema        string                 `json:"schema"  validate:"required"`      // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                    // privacy policy
}

func (req *CustomToolUpdateReq) Check() error { return nil }

type CustomToolIDReq struct {
	CustomToolID string `json:"customToolId" validate:"required"` // Custom tool id
}

func (req *CustomToolIDReq) Check() error { return nil }

type CustomToolSchemaReq struct {
	Schema string `json:"schema" validate:"required"` // schema
}

func (req *CustomToolSchemaReq) Check() error { return nil }

type ToolSquareAPIKeyReq struct {
	ToolSquareID string `json:"toolSquareId" validate:"required"` // square toolId
	APIKey       string `json:"apiKey"`                           // apiKey
}

func (req *ToolSquareAPIKeyReq) Check() error { return nil }

type ToolActionListReq struct {
	ToolId   string `form:"toolId" json:"toolId" validate:"required"`                          // tool id
	ToolType string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // Tool type
}

func (req *ToolActionListReq) Check() error { return nil }

type ToolActionReq struct {
	ToolId     string `form:"toolId" json:"toolId" validate:"required"`                          // tool id
	ToolType   string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // Tool type
	ActionName string `form:"actionName" json:"actionName" validate:"required"`                  // action name
}

func (req *ToolActionReq) Check() error { return nil }
