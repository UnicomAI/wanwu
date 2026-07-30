package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CustomSkillCreate(ctx context.Context, req *mcp_service.CustomSkillCreateReq) (*mcp_service.CustomSkillCreateResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_create", "identity is empty"))
	}

	skillId, err := s.cli.CreateCustomSkill(ctx, &model.CustomSkill{
		Name:            req.Name,
		Avatar:          req.Avatar,
		Author:          req.Author,
		Desc:            req.Desc,
		WgaThreadID:     req.WgaThreadId,
		PreviewThreadID: req.PreviewThreadId,
		UserID:          req.Identity.UserId,
		OrgID:           req.Identity.OrgId,
	})
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}

	return &mcp_service.CustomSkillCreateResp{SkillId: skillId}, nil
}

func (s *Service) CustomSkillDelete(ctx context.Context, req *mcp_service.CustomSkillDeleteReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_delete", "identity is empty"))
	}
	userId := req.Identity.UserId
	orgId := req.Identity.OrgId
	err := s.cli.DeleteCustomSkill(ctx, req.SkillId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) CustomSkillGet(ctx context.Context, req *mcp_service.CustomSkillGetReq) (*mcp_service.PublishCustomSkill, error) {
	customSkill, err := s.cli.GetCustomSkill(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	publish, err := s.cli.GetPublishCustomSkillByLatest(ctx, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return toProtoPublishCustomSkill(customSkill, publish), nil
}

func (s *Service) GetCustomSkillByPreviewID(ctx context.Context, req *mcp_service.GetCustomSkillByPreviewIDReq) (*mcp_service.GetCustomSkillByPreviewIDResp, error) {
	customSkill, st := s.cli.GetCustomSkillByPreviewThreadID(ctx, req.GetPreviewThreadId())
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	if customSkill == nil {
		return &mcp_service.GetCustomSkillByPreviewIDResp{}, nil
	}
	return &mcp_service.GetCustomSkillByPreviewIDResp{Skill: toProtoCustomSkill(customSkill)}, nil
}

func (s *Service) GetCustomSkillByThreadID(ctx context.Context, req *mcp_service.GetCustomSkillByThreadIDReq) (*mcp_service.GetCustomSkillByThreadIDResp, error) {
	customSkill, st := s.cli.GetCustomSkillByWgaThreadID(ctx, req.GetWgaThreadId())
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	if customSkill == nil {
		return &mcp_service.GetCustomSkillByThreadIDResp{}, nil
	}
	return &mcp_service.GetCustomSkillByThreadIDResp{Skill: toProtoCustomSkill(customSkill)}, nil
}

func (s *Service) GetCustomSkillListByThreadIDList(ctx context.Context, req *mcp_service.GetCustomSkillListByThreadIDListReq) (*mcp_service.CustomSkillGetListResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_get_by_wga_thread_list", "identity is empty"))
	}
	userId, orgId := req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId()
	customSkills, st := s.cli.GetCustomSkillListByWgaThreadIDList(ctx, userId, orgId, req.GetWgaThreadIdList())
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	foundByThread := make(map[string]*model.CustomSkill, len(customSkills))
	for _, cs := range customSkills {
		if cs.WgaThreadID != "" {
			foundByThread[cs.WgaThreadID] = cs
		}
	}
	ordered := make([]*model.CustomSkill, 0, len(req.GetWgaThreadIdList()))
	for _, threadID := range req.GetWgaThreadIdList() {
		if threadID == "" {
			continue
		}
		if cs, ok := foundByThread[threadID]; ok {
			ordered = append(ordered, cs)
			continue
		}
		ordered = append(ordered, &model.CustomSkill{WgaThreadID: threadID})
	}
	resp, st := s.buildCustomSkillGetListResp(ctx, ordered)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	return resp, nil
}

func (s *Service) CustomSkillGetList(ctx context.Context, req *mcp_service.CustomSkillGetListReq) (*mcp_service.CustomSkillGetListResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_list", "identity is empty"))
	}
	userId, orgId := req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId()
	customSkills, total, err := s.cli.GetCustomSkillList(ctx, userId, orgId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	resp, st := s.buildCustomSkillGetListResp(ctx, customSkills)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	resp.Total = total
	return resp, nil
}

// AdminCustomSkillPageList 管理员中心 skill 全局分页列表（跨用户，按 userId[]/orgId[]/name 过滤后分页）
func (s *Service) AdminCustomSkillPageList(ctx context.Context, req *mcp_service.AdminCustomSkillPageListReq) (*mcp_service.CustomSkillGetListResp, error) {
	customSkills, total, err := s.cli.GetCustomSkillPageList(ctx, req.UserId, req.OrgId, req.Name, int(req.PageNum), int(req.PageSize))
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	resp, st := s.buildCustomSkillGetListResp(ctx, customSkills)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	resp.Total = total
	return resp, nil
}

// listPublishCustomSkills 按租户拉取草稿列表并组装 PublishCustomSkill（含可选最新发布）。
func (s *Service) listPublishCustomSkills(ctx context.Context, userId, orgId, name string) (*mcp_service.CustomSkillGetListResp, *errs.Status) {
	customSkills, _, st := s.cli.GetCustomSkillList(ctx, userId, orgId, name)
	if st != nil {
		return nil, st
	}
	return s.buildCustomSkillGetListResp(ctx, customSkills)
}

func (s *Service) GetCustomSkillDetailByIdList(ctx context.Context, req *mcp_service.CustomSkillDetailByIdListReq) (*mcp_service.CustomSkillDetailByIdListResp, error) {
	customSkills, err := s.cli.GetCustomSkillBySkillIds(ctx, req.SkillIds)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	skillDetails := make([]*mcp_service.CustomSkill, 0, len(customSkills))
	for _, customSkill := range customSkills {
		skillDetails = append(skillDetails, toProtoCustomSkill(customSkill))
	}
	return &mcp_service.CustomSkillDetailByIdListResp{SkillDetails: skillDetails}, nil
}

func (s *Service) buildCustomSkillGetListResp(ctx context.Context, customSkills []*model.CustomSkill) (*mcp_service.CustomSkillGetListResp, *errs.Status) {
	if len(customSkills) == 0 {
		return &mcp_service.CustomSkillGetListResp{}, nil
	}
	skillIDs := make([]string, 0, len(customSkills))
	skillByID := make(map[string]*model.CustomSkill, len(customSkills))
	for _, cs := range customSkills {
		if cs.ID == 0 {
			continue
		}
		sid := util.Int2Str(cs.ID)
		skillIDs = append(skillIDs, sid)
		skillByID[sid] = cs
	}
	if len(skillIDs) == 0 {
		list := make([]*mcp_service.PublishCustomSkill, 0, len(customSkills))
		for _, cs := range customSkills {
			list = append(list, &mcp_service.PublishCustomSkill{
				Skill: &mcp_service.CustomSkill{WgaThreadId: cs.WgaThreadID},
			})
		}
		return &mcp_service.CustomSkillGetListResp{List: list, Total: int64(len(list))}, nil
	}
	publishes, err := s.cli.GetPublishCustomSkillByIDList(ctx, skillIDs)
	if err != nil {
		return nil, err
	}
	publishBySkill := make(map[string]*model.CustomSkillPublish, len(publishes))
	for _, p := range publishes {
		publishBySkill[p.SkillID] = p
	}
	list := make([]*mcp_service.PublishCustomSkill, 0, len(customSkills))
	for _, cs := range customSkills {
		if cs.ID == 0 {
			list = append(list, &mcp_service.PublishCustomSkill{
				Skill: &mcp_service.CustomSkill{WgaThreadId: cs.WgaThreadID},
			})
			continue
		}
		sid := util.Int2Str(cs.ID)
		list = append(list, toProtoPublishCustomSkill(skillByID[sid], publishBySkill[sid]))
	}
	return &mcp_service.CustomSkillGetListResp{
		List:  list,
		Total: int64(len(list)),
	}, nil
}
