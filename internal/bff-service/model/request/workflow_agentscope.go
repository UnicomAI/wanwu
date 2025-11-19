package request

type DeleteAgentScopeWorkFlowRequest struct {
	AppId string `form:"workflowID" json:"workflowID" validate:"required"` // 应用ID [EN] Application ID
}

func (o *DeleteAgentScopeWorkFlowRequest) Check() error {
	return nil
}

type PublishAgentScopeWorkFlowRequest struct {
	AppId string `form:"workflowID" json:"workflowID" validate:"required"` // 应用ID [EN] Application ID
}

func (p *PublishAgentScopeWorkFlowRequest) Check() error {
	return nil
}

type UnPublishAgentScopeWorkFlowRequest struct {
	AppId string `form:"workflowID" json:"workflowID" validate:"required"` // 应用ID [EN] Application ID
}

func (p *UnPublishAgentScopeWorkFlowRequest) Check() error {
	return nil
}
