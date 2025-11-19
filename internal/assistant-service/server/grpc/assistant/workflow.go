package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AssistantWorkFlowCreate Add workFlow
func (s *Service) AssistantWorkFlowCreate(ctx context.Context, req *assistant_service.AssistantWorkFlowCreateReq) (*emptypb.Empty, error) {
	workflow, err := parseWorkFlowApiInfo(req)
	if err != nil {
		return nil, err
	}

	// Call the client method to create WorkFlow (create WorkFlow and update Assistant in transaction)
	if status := s.cli.CreateAssistantWorkflow(ctx, workflow); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}

func parseWorkFlowApiInfo(req *assistant_service.AssistantWorkFlowCreateReq) (*model.AssistantWorkflow, error) {
	userId, orgId := req.Identity.UserId, req.Identity.OrgId
	assistantId := util.MustU32(req.AssistantId)
	workFlowId := req.WorkFlowId
	workFlow := &model.AssistantWorkflow{
		WorkflowId:  workFlowId,
		AssistantId: assistantId,
		Enable:      true,
		UserId:      userId,
		OrgId:       orgId,
	}

	return workFlow, nil
}

// AssistantWorkFlowDelete delete workFlow
func (s *Service) AssistantWorkFlowDelete(ctx context.Context, req *assistant_service.AssistantWorkFlowDeleteReq) (*emptypb.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)

	// Call the client method to delete the WorkFlow (delete the WorkFlow and update the Assistant in the transaction)
	if status := s.cli.DeleteAssistantWorkflow(ctx, assistantId, req.WorkFlowId); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantWorkFlowEnableSwitch WorkFlow switch
func (s *Service) AssistantWorkFlowEnableSwitch(ctx context.Context, req *assistant_service.AssistantWorkFlowEnableSwitchReq) (*emptypb.Empty, error) {
	// Conversion ID
	assistantId := util.MustU32(req.AssistantId)

	// First obtain existing WorkFlow information
	existingWorkflow, status := s.cli.GetAssistantWorkflow(ctx, assistantId, req.WorkFlowId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	existingWorkflow.Enable = req.Enable
	if status := s.cli.UpdateAssistantWorkflow(ctx, existingWorkflow); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) AssistantWorkFlowDeleteByWorkflowId(ctx context.Context, req *assistant_service.AssistantWorkFlowDeleteByWorkflowIdReq) (*emptypb.Empty, error) {
	workflowId := req.WorkflowId

	if status := s.cli.DeleteAssistantWorkflowByWorkflowId(ctx, workflowId); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}
