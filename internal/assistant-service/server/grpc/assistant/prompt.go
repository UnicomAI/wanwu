package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CustomPromptCreate(ctx context.Context, req *assistant_service.CustomPromptCreateReq) (*assistant_service.CustomPromptIDResp, error) {
	customPromptId, err := s.cli.CreateCustomPrompt(ctx, req.AvatarPath, req.Name, req.Desc, req.Prompt, req.Identity.UserId, req.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	return &assistant_service.CustomPromptIDResp{
		CustomPromptId: customPromptId,
	}, nil
}

func (s *Service) CustomPromptDelete(ctx context.Context, req *assistant_service.CustomPromptDeleteReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteCustomPrompt(ctx, util.MustU32(req.CustomPromptId), req.Identity.UserId, req.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CustomPromptUpdate(ctx context.Context, req *assistant_service.CustomPromptUpdateReq) (*emptypb.Empty, error) {
	err := s.cli.UpdateCustomPrompt(ctx, req)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CustomPromptGet(ctx context.Context, req *assistant_service.CustomPromptGetReq) (*assistant_service.CustomPromptInfo, error) {
	customPrompt, err := s.cli.GetCustomPrompt(ctx, util.MustU32(req.CustomPromptId))
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	return toCustomPromptInfo(customPrompt), nil
}

func (s *Service) CustomPromptGetList(ctx context.Context, req *assistant_service.CustomPromptGetListReq) (*assistant_service.CustomPromptList, error) {
	customPrompts, count, err := s.cli.GetCustomPromptList(ctx, req.Identity.UserId, req.Identity.OrgId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	customPromptInfos := make([]*assistant_service.CustomPromptInfo, 0, len(customPrompts))
	for _, customPrompt := range customPrompts {
		customPromptInfos = append(customPromptInfos, toCustomPromptInfo(customPrompt))
	}

	return &assistant_service.CustomPromptList{
		CustomPromptInfos: customPromptInfos,
		Total:             count,
	}, nil
}

// GetAdminCustomPromptPageList 管理员中心跨组织查询自定义提示词列表（SQL分页）
func (s *Service) GetAdminCustomPromptPageList(ctx context.Context, req *assistant_service.GetAdminCustomPromptPageListReq) (*assistant_service.CustomPromptList, error) {
	customPrompts, count, err := s.cli.GetCustomPromptListAdmin(ctx, req.OrgIdList, req.UserIdList, req.Name, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	customPromptInfos := make([]*assistant_service.CustomPromptInfo, 0, len(customPrompts))
	for _, customPrompt := range customPrompts {
		customPromptInfos = append(customPromptInfos, toCustomPromptInfo(customPrompt))
	}

	return &assistant_service.CustomPromptList{
		CustomPromptInfos: customPromptInfos,
		Total:             count,
	}, nil
}

func (s *Service) CustomPromptCopy(ctx context.Context, req *assistant_service.CustomPromptCopyReq) (*assistant_service.CustomPromptIDResp, error) {
	customPromptId, err := s.cli.CopyCustomPrompt(ctx, util.MustU32(req.CustomPromptId), req.Identity.UserId, req.Identity.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantCustomPromptErr, err)
	}

	return &assistant_service.CustomPromptIDResp{
		CustomPromptId: customPromptId,
	}, nil
}

// --- internal ---
func toCustomPromptInfo(customPrompt *model.CustomPrompt) *assistant_service.CustomPromptInfo {
	return &assistant_service.CustomPromptInfo{
		CustomPromptId: util.Int2Str(customPrompt.ID),
		AvatarPath:     customPrompt.AvatarPath,
		Name:           customPrompt.Name,
		Desc:           customPrompt.Desc,
		Prompt:         customPrompt.Prompt,
		CreatedAt:      customPrompt.CreatedAt,
		UpdatedAt:      customPrompt.UpdatedAt,
		Identity: &assistant_service.Identity{
			UserId: customPrompt.UserId,
			OrgId:  customPrompt.OrgId,
		},
	}
}
