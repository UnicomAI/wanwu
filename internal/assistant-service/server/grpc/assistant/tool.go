// @Author wangxm 8/13/星期三 15:20:00 [EN] @Author wangxm 8/13/Wednesday 15:20:00
package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/util"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) AssistantToolDeleteByToolId(ctx context.Context, req *assistant_service.AssistantToolDeleteByToolIdReq) (*empty.Empty, error) {
	if status := s.cli.DeleteAssistantToolByToolId(ctx, req.ToolId, req.ToolType); status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) AssistantToolCreate(ctx context.Context, req *assistant_service.AssistantToolCreateReq) (*empty.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)

	if status := s.cli.CreateAssistantTool(ctx, assistantId, req.ToolId, req.ToolType, req.ActionName, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}

	return &empty.Empty{}, nil
}

func (s *Service) AssistantToolDelete(ctx context.Context, req *assistant_service.AssistantToolDeleteReq) (*empty.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)

	if status := s.cli.DeleteAssistantTool(ctx, assistantId, req.ToolId, req.ToolType, req.ActionName); status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) AssistantToolEnableSwitch(ctx context.Context, req *assistant_service.AssistantToolEnableSwitchReq) (*empty.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)

	existingCustom, status := s.cli.GetAssistantTool(ctx, assistantId, req.ToolId, req.ToolType, req.ActionName)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}

	existingCustom.Enable = req.Enable
	if status := s.cli.UpdateAssistantTool(ctx, existingCustom); status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) AssistantToolConfig(ctx context.Context, req *assistant_service.AssistantToolConfigReq) (*empty.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)

	if status := s.cli.UpdateAssistantToolConfig(ctx, assistantId, req.ToolId, req.ToolConfig); status != nil {
		return nil, errStatus(errs.Code_AssistantToolErr, status)
	}

	return &empty.Empty{}, nil
}
