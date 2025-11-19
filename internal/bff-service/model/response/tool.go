package response

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type CustomToolDetail struct {
	CustomToolInfo
	Schema        string                 `json:"schema"`        // schema
	ApiAuth       util.ApiAuthWebRequest `json:"apiAuth"`       // apiAuth
	ApiList       []CustomToolActionInfo `json:"apiList"`       // action list
	PrivacyPolicy string                 `json:"privacyPolicy"` // privacy policy
}

type CustomToolInfo struct {
	CustomToolId string         `json:"customToolId"` // Custom tool id
	Name         string         `json:"name"`         // name
	Description  string         `json:"description"`  // describe
	Avatar       request.Avatar `json:"avatar"`       // icon
}

type CustomToolActionInfo struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ToolSquareDetail struct {
	ToolSquareInfo
	ToolSquareActions
	Schema string `json:"schema"`
}

type ToolSquareInfo struct {
	ToolSquareID string         `json:"toolSquareId"` // Square mcpId (non-empty means it comes from square)
	Avatar       request.Avatar `json:"avatar"`       // icon
	Name         string         `json:"name"`         // name
	Desc         string         `json:"desc"`         // describe
	Tags         []string       `json:"tags"`         // Label
}

type ToolSquareActions struct {
	NeedApiKeyInput bool                   `json:"needApiKeyInput"` // Whether apiKey input is required
	APIKey          string                 `json:"apiKey"`          // apiKey
	ApiAuth         util.ApiAuthWebRequest `json:"apiAuth"`         // apiAuth
	Tools           []*protocol.Tool       `json:"tools"`           // action list
	Detail          string                 `json:"detail"`          // Detailed description
	ActionSum       int64                  `json:"actionSum"`       // total number of actions
}

type ToolSelect struct {
	UniqueId string `json:"uniqueId"` // unique id
	ToolInfo
}

type ToolInfo struct {
	ToolId          string         `json:"toolId"`                                            // tool id
	ToolName        string         `json:"toolName"`                                          // Tool name
	ToolType        string         `json:"toolType" validate:"required,oneof=custom builtin"` // Tool type
	Desc            string         `json:"desc"`                                              // Tool description
	NeedApiKeyInput bool           `json:"needApiKeyInput"`                                   // Whether apiKey input is required
	APIKey          string         `json:"apiKey"`                                            // apiKey
	Avatar          request.Avatar `json:"avatar"`                                            // icon
}

type ToolActionList struct {
	Actions []*protocol.Tool `json:"actions"` // action list
}

type ToolActionDetail struct {
	NeedApiKeyInput bool           `json:"needApiKeyInput"` // Whether apiKey input is required
	APIKey          string         `json:"apiKey"`          // apiKey
	Action          *protocol.Tool `json:"action"`          // action list
}
