package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AssistantWorkFlowCreate 添加workFlow
func (s *Service) AssistantWorkFlowCreate(ctx context.Context, req *assistant_service.AssistantWorkFlowCreateReq) (*emptypb.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}
	workflow, err := parseWorkFlowApiInfo(req, assistant.ID)
	if err != nil {
		return nil, err
	}

	// 调用client方法创建WorkFlow（在事务中创建WorkFlow并更新Assistant）
	if status := s.cli.CreateAssistantWorkflow(ctx, workflow); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}

func parseWorkFlowApiInfo(req *assistant_service.AssistantWorkFlowCreateReq, assistantId uint32) (*model.AssistantWorkflow, error) {
	userId, orgId := req.Identity.UserId, req.Identity.OrgId
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

// AssistantWorkFlowDelete 删除workFlow
func (s *Service) AssistantWorkFlowDelete(ctx context.Context, req *assistant_service.AssistantWorkFlowDeleteReq) (*emptypb.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}
	assistantId := assistant.ID

	// 调用client方法删除WorkFlow（在事务中删除WorkFlow并更新Assistant）
	if status := s.cli.DeleteAssistantWorkflow(ctx, assistantId, req.WorkFlowId); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantWorkFlowEnableSwitch WorkFlow开关
func (s *Service) AssistantWorkFlowEnableSwitch(ctx context.Context, req *assistant_service.AssistantWorkFlowEnableSwitchReq) (*emptypb.Empty, error) {
	// 通过 UUID 获取自增 ID（带权限校验）
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}
	assistantId := assistant.ID

	// 先获取现有WorkFlow信息
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
	userId, orgId := req.Identity.UserId, req.Identity.OrgId

	if status := s.cli.DeleteAssistantWorkflowByWorkflowId(ctx, workflowId, userId, orgId); status != nil {
		return nil, errStatus(errs.Code_AssistantWorkflowErr, status)
	}

	return &emptypb.Empty{}, nil
}
