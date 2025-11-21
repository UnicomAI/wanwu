package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
)

type CozeWorkflowModelInfo struct {
	ModelInfo
	ModelAbility CozeWorkflowModelInfoAbility `json:"model_ability"`
	ModelParams  []config.WorkflowModelParam  `json:"model_params"`
}

type CozeWorkflowModelInfoAbility struct {
	CotDisplay         bool `json:"cot_display"`
	FunctionCall       bool `json:"function_call"`
	ImageUnderstanding bool `json:"image_understanding"`
	AudioUnderstanding bool `json:"audio_understanding"`
	VideoUnderstanding bool `json:"video_understanding"`
}

type CozeWorkflowListResp struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data *CozeWorkflowListData `json:"data,omitempty"`
}

type CozeWorkflowListData struct {
	Workflows []*CozeWorkflowListDataWorkflow `json:"workflow_list"`
}

type CozeWorkflowListDataWorkflow struct {
	WorkflowId string `json:"workflow_id"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	URL        string `json:"url"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

type CozeWorkflowIDResp struct {
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data *CozeWorkflowIDData `json:"data,omitempty"`
}

type CozeWorkflowIDData struct {
	WorkflowID string `json:"workflow_id"`
}

type CozeWorkflowDeleteResp struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data *CozeWorkflowDeleteData `json:"data,omitempty"`
}

type CozeWorkflowDeleteData struct {
	Status int64 `json:"status"`
}

func (d *CozeWorkflowDeleteData) GetStatus() int64 {
	if d == nil {
		return 0
	}
	return d.Status
}

type CozeWorkflowExportResp struct {
	Code int                     `json:"code"`
	Msg  string                  `json:"msg"`
	Data *CozeWorkflowExportData `json:"data,omitempty"`
}

type CozeWorkflowExportData struct {
	WorkflowName string `json:"name"`
	WorkflowDesc string `json:"desc"`
	Schema       string `json:"schema"`
}

type ToolDetail4Workflow struct {
	Inputs     []interface{} `json:"inputs"`
	Outputs    []interface{} `json:"outputs"`
	ActionName string        `json:"actionName"`
	ActionID   string        `json:"actionId"`
	IconUrl    string        `json:"iconUrl"`
}

// ToolActionParamWithoutTypeList4Workflow type non-list definition
type ToolActionParamWithoutTypeList4Workflow struct {
	Input       struct{}      `json:"input"`
	Description string        `json:"description"`
	Name        string        `json:"name"`
	Type        string        `json:"type"` // non-list
	Required    bool          `json:"required"`
	Children    []interface{} `json:"schema"`
}

// ToolActionParamWithTypeList4Workflow type is the definition of list
type ToolActionParamWithTypeList4Workflow struct {
	Input       struct{}                           `json:"input"`
	Description string                             `json:"description"`
	Name        string                             `json:"name"`
	Type        string                             `json:"type"` // list
	Required    bool                               `json:"required"`
	Schema      ToolActionParamInTypeList4Workflow `json:"schema"`
}

type ToolActionParamInTypeList4Workflow struct {
	Type     string        `json:"type"`
	Children []interface{} `json:"schema"`
}
