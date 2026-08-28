package request

import "github.com/UnicomAI/wanwu/pkg/util"

type CustomToolCreate struct {
	Avatar        Avatar                 `json:"avatar"`                          // 图标
	Name          string                 `json:"name" validate:"required"`        // 名称
	Description   string                 `json:"description" validate:"required"` // 描述
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`     // api身份认证
	Schema        string                 `json:"schema"  validate:"required"`     // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                   // 隐私政策
}

func (req *CustomToolCreate) Check() error {
	if err := util.ValidateName(&req.Name, util.SubjectCustomTool); err != nil {
		return err
	}
	return util.ValidateDesc(&req.Description, util.SubjectCustomTool)
}

type CustomToolUpdateReq struct {
	Avatar        Avatar                 `json:"avatar"`                           // 图标
	CustomToolID  string                 `json:"customToolId" validate:"required"` // 自定义工具ID
	Name          string                 `json:"name" validate:"required"`         // 名称
	Description   string                 `json:"description" validate:"required"`  // 描述
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth" validate:"required"`      // api身份认证
	Schema        string                 `json:"schema"  validate:"required"`      // schema
	PrivacyPolicy string                 `json:"privacyPolicy"`                    // 隐私政策
}

func (req *CustomToolUpdateReq) Check() error { return nil }

type CustomToolIDReq struct {
	CustomToolID string `json:"customToolId" validate:"required"` // 自定义工具id
}

func (req *CustomToolIDReq) Check() error { return nil }

type CustomToolSchemaReq struct {
	Schema string `json:"schema" validate:"required"` // schema
}

func (req *CustomToolSchemaReq) Check() error { return nil }

type ToolSquareAPIKeyReq struct {
	ToolSquareID string `json:"toolSquareId" validate:"required"` // 广场toolId
	APIKey       string `json:"apiKey"`                           // apiKey
}

func (req *ToolSquareAPIKeyReq) Check() error { return nil }

type ToolActionListReq struct {
	ToolId   string `form:"toolId" json:"toolId" validate:"required"`                          // 工具id
	ToolType string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // 工具类型
}

func (req *ToolActionListReq) Check() error { return nil }

type ToolActionReq struct {
	ToolId     string `form:"toolId" json:"toolId" validate:"required"`                          // 工具id
	ToolType   string `form:"toolType" json:"toolType" validate:"required,oneof=builtin custom"` // 工具类型
	ActionName string `form:"actionName" json:"actionName" validate:"required"`                  // action名称
}

func (req *ToolActionReq) Check() error { return nil }
