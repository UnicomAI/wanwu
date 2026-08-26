// @Author wangxm 3/10/星期二 15:04:00
package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) AssistantSkillCreate(ctx context.Context, req *assistant_service.AssistantSkillCreateReq) (*empty.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantSkillCreateErr, status)
	}
	assistantId := assistant.ID

	if status := s.cli.CreateAssistantSkill(ctx, assistantId, req.SkillId, req.SkillType, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_AssistantSkillCreateErr, status)
	}

	return &empty.Empty{}, nil
}

func (s *Service) AssistantSkillDelete(ctx context.Context, req *assistant_service.AssistantSkillDeleteReq) (*empty.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantSkillDeleteErr, status)
	}
	assistantId := assistant.ID

	if status := s.cli.DeleteAssistantSkill(ctx, assistantId, req.SkillId, req.SkillType); status != nil {
		return nil, errStatus(errs.Code_AssistantSkillDeleteErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) AssistantSkillEnableSwitch(ctx context.Context, req *assistant_service.AssistantSkillEnableSwitchReq) (*empty.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantSkillEnableSwitchErr, status)
	}
	assistantId := assistant.ID

	if status := s.cli.UpdateAssistantSkillEnable(ctx, assistantId, req.SkillId, req.SkillType, req.Enable); status != nil {
		return nil, errStatus(errs.Code_AssistantSkillEnableSwitchErr, status)
	}
	return &empty.Empty{}, nil
}
