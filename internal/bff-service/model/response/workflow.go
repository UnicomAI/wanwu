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
	WorkflowId string           `json:"workflow_id"`
	Name       string           `json:"name"`
	Desc       string           `json:"desc"`
	URL        string           `json:"url"`
	IconURI    string           `json:"icon_uri"`
	CreateTime int64            `json:"create_time"`
	UpdateTime int64            `json:"update_time"`
	FlowMode   int64            `json:"flow_mode"`
	SpaceID    string           `json:"space_id"`
	Creator    *WorkflowCreator `json:"creator"`
}

// WorkflowCreator 工作流创建者
type WorkflowCreator struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
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

type CozeWorkflowTestRunResponse struct {
	Data *CozeWorkflowTestRunData `json:"data"`
	Code int64                    `json:"code"`
	Msg  string                   `json:"msg"`
}

type CozeWorkflowTestRunData struct {
	WorkflowID string `json:"workflow_id"`
	ExecuteID  string `json:"execute_id"`
	SessionID  string `json:"session_id"`
}

type CozeGetWorkflowProcessResponse struct {
	Code int64                       `json:"code"`
	Msg  string                      `json:"msg"`
	Data *CozeGetWorkFlowProcessData `json:"data"`
}

type CozeGetWorkFlowProcessData struct {
	WorkFlowId       string            `json:"workFlowId"`
	ExecuteId        string            `json:"executeId"`
	ExecuteStatus    int64             `json:"executeStatus"`
	NodeResults      []*CozeNodeResult `json:"nodeResults"`
	Rate             string            `json:"rate"`
	ExeHistoryStatus int64             `json:"exeHistoryStatus"`
	WorkflowExeCost  string            `json:"workflowExeCost"`
	TokenAndCost     *CozeTokenAndCost `json:"tokenAndCost,omitempty"`
	Reason           *string           `json:"reason,omitempty"`
	LastNodeID       *string           `json:"lastNodeID,omitempty"`
	LogID            string            `json:"logID"`
	NodeEvents       []*CozeNodeEvent  `json:"nodeEvents"`
	ProjectId        string            `json:"projectId"`
}

type CozeTokenAndCost struct {
	InputTokens  *string `json:"inputTokens,omitempty"`
	InputCost    *string `json:"inputCost,omitempty"`
	OutputTokens *string `json:"outputTokens,omitempty"`
	OutputCost   *string `json:"outputCost,omitempty"`
	TotalTokens  *string `json:"totalTokens,omitempty"`
	TotalCost    *string `json:"totalCost,omitempty"`
}

type CozeNodeResult struct {
	NodeId          string            `json:"nodeId"`
	NodeType        string            `json:"NodeType"`
	NodeName        string            `json:"NodeName"`
	NodeStatus      int64             `json:"nodeStatus"`
	ErrorInfo       string            `json:"errorInfo"`
	Input           string            `json:"input"`
	Output          string            `json:"output"`
	NodeExeCost     string            `json:"nodeExeCost"`
	TokenAndCost    *CozeTokenAndCost `json:"tokenAndCost,omitempty"`
	RawOutput       *string           `json:"raw_output,omitempty"`
	ErrorLevel      string            `json:"errorLevel"`
	Index           *int32            `json:"index,omitempty"`
	Items           *string           `json:"items,omitempty"`
	MaxBatchSize    *int32            `json:"maxBatchSize,omitempty"`
	LimitVariable   *string           `json:"limitVariable,omitempty"`
	LoopVariableLen *int32            `json:"loopVariableLen,omitempty"`
	Batch           *string           `json:"batch,omitempty"`
	IsBatch         *bool             `json:"isBatch,omitempty"`
	LogVersion      int32             `json:"logVersion"`
	Extra           string            `json:"extra"`
	ExecuteId       *string           `json:"executeId,omitempty"`
	SubExecuteId    *string           `json:"subExecuteId,omitempty"`
	NeedAsync       *bool             `json:"needAsync,omitempty"`
}

type CozeNodeEvent struct {
	ID           string `json:"id"`
	Type         int64  `json:"type"`
	NodeTitle    string `json:"node_title"`
	Data         string `json:"data"`
	NodeIcon     string `json:"node_icon"`
	NodeID       string `json:"node_id"`
	SchemaNodeID string `json:"schema_node_id"`
}

type ToolDetail4Workflow struct {
	Inputs     []interface{} `json:"inputs"`
	Outputs    []interface{} `json:"outputs"`
	ActionName string        `json:"actionName"`
	ActionID   string        `json:"actionId"`
	IconUrl    string        `json:"iconUrl"`
}

// ToolActionParamWithoutTypeList4Workflow type非list的定义
type ToolActionParamWithoutTypeList4Workflow struct {
	Input       struct{}      `json:"input"`
	Description string        `json:"description"`
	Name        string        `json:"name"`
	Type        string        `json:"type"` // 非list
	Required    bool          `json:"required"`
	Children    []interface{} `json:"schema"`
}

// ToolActionParamWithTypeList4Workflow type是list的定义
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

type CozeCreateConversationResponse struct {
	Code             int64                 `thrift:"code,1" form:"code" json:"code" query:"code"`
	Msg              string                `thrift:"msg,2" form:"msg" json:"msg" query:"msg"`
	ConversationData *CozeConversationData `thrift:"ConversationData,3,optional" form:"data" json:"data,omitempty"`
}

type CozeConversationData struct {
	Id            int64             `thrift:"Id,1" form:"id" json:"id,string"`
	CreatedAt     int64             `thrift:"CreatedAt,2" form:"created_at" json:"created_at"`
	MetaData      map[string]string `thrift:"MetaData,3" form:"meta_data" json:"meta_data"`
	CreatorID     *int64            `thrift:"CreatorID,4,optional" form:"creator_d" json:"creator_d,string,omitempty"`
	ConnectorID   *int64            `thrift:"ConnectorID,5,optional" form:"connector_id" json:"connector_id,string,omitempty"`
	LastSectionID *int64            `thrift:"LastSectionID,6,optional" form:"last_section_id" json:"last_section_id,string,omitempty"`
	AccountID     *int64            `thrift:"AccountID,7,optional" form:"account_id" json:"account_id,omitempty"`
}

type CozeListMessageApiResponse struct {
	Messages []*OpenMessageApi `thrift:"messages,1,optional" form:"data" json:"data,omitempty"`
	// Is there still data, true yes, false no
	HasMore *bool `thrift:"has_more,2,optional" form:"has_more" json:"has_more,omitempty"`
	// The ID of the first piece of data
	FirstID *int64 `thrift:"first_id,3,optional" form:"first_id" json:"first_id,string,omitempty"`
	// The id of the last piece of data.
	LastID *int64 `thrift:"last_id,4,optional" form:"last_id" json:"last_id,string,omitempty"`
	Code   int64  `thrift:"code,253" form:"code" json:"code" query:"code"`
	Msg    string `thrift:"msg,254" form:"msg" json:"msg" query:"msg"`
}

type CozeWorkflowVersionListResp struct {
	Code int                          `json:"code"`
	Msg  string                       `json:"msg"`
	Data *CozeWorkflowVersionListData `json:"data"`
}

type CozeWorkflowVersionListData struct {
	WorkflowID  string                 `json:"workflow_id"`
	VersionList []*CozeWorkflowVersion `json:"version_list"`
	Total       int32                  `json:"total"`
}

type CozeWorkflowVersion struct {
	Version   string `json:"version"`
	Desc      string `json:"version_description"`
	CreatedAt int64  `json:"created_at"`
	CommitId  string `json:"commit_id"`
}

type CozeMGetWorkflowLatestVersionResponse struct {
	Data []*CozeWorkflowVersionInfo `json:"data"`
	Code int64                      `json:"code"`
	Msg  string                     `json:"msg"`
}

type CozeWorkflowVersionInfo struct {
	WorkflowID         string `json:"workflow_id"`
	Version            string `json:"version"`
	VersionDescription string `json:"version_description"`
	CreatedAt          int64  `json:"created_at"`
	CommitID           string `json:"commit_id"`
	Type               int64  `json:"type"`
}

type CozeCommonResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
