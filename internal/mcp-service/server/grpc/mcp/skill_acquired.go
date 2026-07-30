package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) AcquiredSkillCreate(ctx context.Context, req *mcp_service.AcquiredSkillCreateReq) (*mcp_service.AcquiredSkillCreateResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_create", "identity is empty"))
	}
	acquiredSkillId, err := s.cli.CreateAcquiredSkill(ctx, &model.AcquiredSkill{
		CustomSkillID: req.CustomSkillId,
		UserID:        req.GetIdentity().GetUserId(),
		OrgID:         req.GetIdentity().GetOrgId(),
	})
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	return &mcp_service.AcquiredSkillCreateResp{AcquiredSkillId: acquiredSkillId}, nil
}

func (s *Service) AcquiredSkillDelete(ctx context.Context, req *mcp_service.AcquiredSkillDeleteReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_delete", "identity is empty"))
	}
	userId := req.Identity.UserId
	orgId := req.Identity.OrgId
	err := s.cli.DeleteAcquiredSkill(ctx, req.AcquiredSkillId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) AcquiredSkillGet(ctx context.Context, req *mcp_service.AcquiredSkillGetReq) (*mcp_service.AcquiredSkill, error) {
	acquiredSkill, err := s.cli.GetAcquiredSkill(ctx, req.AcquiredSkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	list, protoErr := s.toProtoAcquiredSkills(ctx, []*model.AcquiredSkill{acquiredSkill})
	if protoErr != nil {
		return nil, protoErr
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (s *Service) AcquiredSkillGetList(ctx context.Context, req *mcp_service.AcquiredSkillGetListReq) (*mcp_service.AcquiredSkillGetListResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_list", "identity is empty"))
	}
	userId, orgId := req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId()
	acquiredSkills, total, err := s.cli.GetAcquiredSkillList(ctx, userId, orgId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	list, protoErr := s.toProtoAcquiredSkills(ctx, acquiredSkills)
	if protoErr != nil {
		return nil, protoErr
	}
	return &mcp_service.AcquiredSkillGetListResp{List: list, Total: total}, nil
}

func (s *Service) AcquiredSkillGetByIDList(ctx context.Context, req *mcp_service.AcquiredSkillGetByIDListReq) (*mcp_service.AcquiredSkillGetByIDListResp, error) {
	acquiredSkills, err := s.cli.GetAcquiredSkillByIDList(ctx, req.AcquiredSkillIdList)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	list, protoErr := s.toProtoAcquiredSkills(ctx, acquiredSkills)
	if protoErr != nil {
		return nil, protoErr
	}
	return &mcp_service.AcquiredSkillGetByIDListResp{List: list, Total: int64(len(list))}, nil
}

func (s *Service) AcquiredSkillGetHistoryList(ctx context.Context, req *mcp_service.AcquiredSkillGetHistoryListReq) (*mcp_service.AcquiredSkillHistoryListResp, error) {
	acquiredSkill, err := s.cli.GetAcquiredSkill(ctx, req.AcquiredSkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	history, total, err := s.cli.GetPublishCustomSkillHistoryList(ctx, acquiredSkill.CustomSkillID)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	customSkill, csErr := s.cli.GetCustomSkill(ctx, acquiredSkill.CustomSkillID)
	if csErr != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, csErr)
	}
	list := make([]*mcp_service.PublishCustomSkill, 0, len(history))
	for _, item := range history {
		list = append(list, toProtoPublishCustomSkill(customSkill, item))
	}
	return &mcp_service.AcquiredSkillHistoryListResp{HistoryList: list, Total: total}, nil
}

// CheckAcquiredSkill 批量检查指定 skillIds 是否已被当前用户获取。
func (s *Service) CheckAcquiredSkill(ctx context.Context, req *mcp_service.CheckAcquiredSkillReq) (*mcp_service.CheckAcquiredSkillResp, error) {
	acquiredMap := make(map[string]bool)
	if len(req.GetSkillIds()) > 0 {
		acquiredList, _, st := s.cli.GetAcquiredSkillList(ctx, req.Identity.GetUserId(), req.Identity.GetOrgId(), "")
		if st == nil {
			for _, acquired := range acquiredList {
				if acquired.CustomSkillID != "" {
					acquiredMap[acquired.CustomSkillID] = true
				}
			}
		}
	}
	result := make(map[string]bool, len(req.GetSkillIds()))
	for _, skillId := range req.GetSkillIds() {
		result[skillId] = acquiredMap[skillId]
	}
	return &mcp_service.CheckAcquiredSkillResp{AcquiredMap: result}, nil
}

func (s *Service) toProtoAcquiredSkills(ctx context.Context, acquiredSkills []*model.AcquiredSkill) ([]*mcp_service.AcquiredSkill, error) {
	if len(acquiredSkills) == 0 {
		return []*mcp_service.AcquiredSkill{}, nil
	}

	seenCustomSkillID := make(map[string]struct{})
	customSkillIDs := make([]string, 0, len(acquiredSkills))
	for _, as := range acquiredSkills {
		if as == nil || as.CustomSkillID == "" {
			continue
		}
		if _, ok := seenCustomSkillID[as.CustomSkillID]; ok {
			continue
		}
		seenCustomSkillID[as.CustomSkillID] = struct{}{}
		customSkillIDs = append(customSkillIDs, as.CustomSkillID)
	}

	customSkillByID := make(map[string]*model.CustomSkill, len(customSkillIDs))
	publishBySkillID := make(map[string]*model.CustomSkillPublish, len(customSkillIDs))
	if len(customSkillIDs) > 0 {
		customSkills, st := s.cli.GetCustomSkillBySkillIds(ctx, customSkillIDs)
		if st != nil {
			return nil, errStatus(errs.Code_MCPAcquiredSkillErr, st)
		}
		for _, cs := range customSkills {
			customSkillByID[util.Int2Str(cs.ID)] = cs
		}
		publishes, pst := s.cli.GetPublishCustomSkillByIDList(ctx, customSkillIDs)
		if pst != nil {
			return nil, errStatus(errs.Code_MCPAcquiredSkillErr, pst)
		}
		for _, p := range publishes {
			publishBySkillID[p.SkillID] = p
		}
	}

	list := make([]*mcp_service.AcquiredSkill, 0, len(acquiredSkills))
	for _, as := range acquiredSkills {
		if as == nil {
			continue
		}
		customSkill, ok := customSkillByID[as.CustomSkillID]
		if !ok {
			return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_custom_skill_not_found", as.CustomSkillID))
		}
		list = append(list, &mcp_service.AcquiredSkill{
			AcquiredSkillId: util.Int2Str(as.ID),
			Skill:           toProtoPublishCustomSkill(customSkill, publishBySkillID[as.CustomSkillID]),
			CreatedAt:       as.CreatedAt,
			UpdatedAt:       as.UpdatedAt,
			UserId:          as.UserID,
			OrgId:           as.OrgID,
		})
	}
	return list, nil
}
