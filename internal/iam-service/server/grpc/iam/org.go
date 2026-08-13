package iam

import (
	"context"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetOrgSelect(ctx context.Context, req *iam_service.GetOrgSelectReq) (*iam_service.OrgSelectResp, error) {
	orgs, err := s.cli.SelectOrgs(ctx, util.MustU32(req.UserId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.OrgSelectResp{Selects: toProtoIDNameWithAvatars(orgs)}, nil
}

func (s *Service) GetOrgList(ctx context.Context, req *iam_service.GetOrgListReq) (*iam_service.GetOrgListResp, error) {
	orgs, count, err := s.cli.GetOrgs(ctx, util.MustU32(req.ParentId), req.Name, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	resp := &iam_service.GetOrgListResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, org := range orgs {
		resp.Orgs = append(resp.Orgs, toOrgInfo(org))
	}
	return resp, nil
}

func (s *Service) GetOrgInfo(ctx context.Context, req *iam_service.GetOrgInfoReq) (*iam_service.OrgInfo, error) {
	org, err := s.cli.GetOrg(ctx, util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return toOrgInfo(org), nil
}

func (s *Service) GetOrgByOrgIDs(ctx context.Context, req *iam_service.GetOrgByOrgIDsReq) (*iam_service.GetOrgByOrgIDsResp, error) {
	var orgIDs []uint32
	for _, userID := range req.OrgIds {
		orgIDs = append(orgIDs, util.MustU32(userID))
	}
	orgs, err := s.cli.GetOrgByOrgIDs(ctx, orgIDs)
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.GetOrgByOrgIDsResp{
		Orgs: toIDFullNames(orgs),
	}, nil
}

func (s *Service) GetOrgAndSubOrgSelectByUser(ctx context.Context, req *iam_service.GetOrgAndSubOrgSelectByUserReq) (*iam_service.GetOrgAndSubOrgSelectByUserResp, error) {
	orgs, err := s.cli.GetOrgAndSubOrgSelectByUser(ctx, util.MustU32(req.UserId), util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.GetOrgAndSubOrgSelectByUserResp{
		Orgs: toIDNames(orgs),
	}, nil
}

func (s *Service) GetFirstClassOrgAndSubs(ctx context.Context, req *iam_service.GetFirstClassOrgAndSubsReq) (*iam_service.GetFirstClassOrgAndSubsResp, error) {
	orgTree, err := s.cli.GetFirstClassOrgAndSubs(ctx, util.MustU32(req.UserId), util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.GetFirstClassOrgAndSubsResp{
		Orgs: toIDNames(orgTree),
	}, nil
}

func (s *Service) GetAdminOrgSelect(ctx context.Context, req *iam_service.GetAdminOrgSelectReq) (*iam_service.OrgSelectResp, error) {
	orgs, status := s.cli.GetAdminOrgSelect(ctx, util.MustU32(req.UserId))
	if status != nil {
		return nil, errStatus(errs.Code_IAMOrg, status)
	}
	return &iam_service.OrgSelectResp{
		Selects: toProtoIDNameWithAvatars(orgs),
	}, nil
}

func (s *Service) GetAdminOrgSubTree(ctx context.Context, req *iam_service.GetAdminOrgSubTreeReq) (*iam_service.AdminOrgSubTreeResp, error) {
	nodes, status := s.cli.GetAdminOrgSubTree(ctx, util.MustU32(req.UserId))
	if status != nil {
		return nil, errStatus(errs.Code_IAMOrg, status)
	}
	return &iam_service.AdminOrgSubTreeResp{
		Orgs: toAdminOrgTreeNodes(nodes),
	}, nil
}

func (s *Service) GetAdminOrgIDs(ctx context.Context, req *iam_service.GetAdminOrgIDsReq) (*iam_service.GetAdminOrgIDsResp, error) {
	orgIDs, status := s.cli.GetAdminOrgIDs(ctx, util.MustU32(req.UserId))
	if status != nil {
		return nil, errStatus(errs.Code_IAMOrg, status)
	}
	resp := &iam_service.GetAdminOrgIDsResp{}
	for _, id := range orgIDs {
		resp.OrgIds = append(resp.OrgIds, strconv.Itoa(int(id)))
	}
	return resp, nil
}

func (s *Service) CreateOrg(ctx context.Context, req *iam_service.CreateOrgReq) (*iam_service.IDNameWithAvatar, error) {
	orgID, err := s.cli.CreateOrg(ctx, &model.Org{
		Status:     true,
		CreatorID:  util.MustU32(req.CreatorId),
		ParentID:   util.MustU32(req.ParentId),
		Name:       req.Name,
		Remark:     req.Remark,
		AvatarPath: req.AvatarPath,
	})
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.IDNameWithAvatar{Id: strconv.Itoa(int(orgID)), Name: req.Name, AvatarPath: req.AvatarPath}, nil
}

func (s *Service) UpdateOrg(ctx context.Context, req *iam_service.UpdateOrgReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdateOrg(ctx, &model.Org{
		ID:         util.MustU32(req.OrgId),
		Name:       req.Name,
		Remark:     req.Remark,
		AvatarPath: req.AvatarPath,
	}); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteOrg(ctx context.Context, req *iam_service.DeleteOrgReq) (*emptypb.Empty, error) {
	if err := s.cli.DeleteOrg(ctx, util.MustU32(req.OrgId)); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ChangeOrgStatus(ctx context.Context, req *iam_service.ChangeOrgStatusReq) (*emptypb.Empty, error) {
	if err := s.cli.ChangeOrgStatus(ctx, util.MustU32(req.OrgId), req.Status); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) AddOrgUser(ctx context.Context, req *iam_service.AddOrgUserReq) (*emptypb.Empty, error) {
	if err := s.cli.AddOrgUser(ctx, util.MustU32(req.OrgId), util.MustU32(req.UserId), util.MustU32(req.RoleId)); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveOrgUser(ctx context.Context, req *iam_service.RemoveOrgUserReq) (*emptypb.Empty, error) {
	if err := s.cli.RemoveOrgUser(ctx, util.MustU32(req.OrgId), util.MustU32(req.UserId)); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) BatchRemoveOrgUser(ctx context.Context, req *iam_service.BatchRemoveOrgUserReq) (*emptypb.Empty, error) {
	if err := s.cli.BatchRemoveOrgUser(ctx, util.MustU32(req.OrgId), util.MustU32s(req.UserIds)); err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &emptypb.Empty{}, nil
}

// ValidateUserOrgPairs 校验用户-组织二元组有效性（去重 + 过滤禁用账号/禁用成员 + 数量上限）。
// 消息中心名单型受众的统一入口清洗：校验失败/超限只影响消息生成，不向业务主流程抛错，
// 故超限返回 exceeded=true 而非 error。
//
// 注意：此处的 string→uint32 必须用 util.U32 显式校验，禁止用 util.MustU32——
// 后者静默把解析失败返回 0，会让非法 userId 变成 userID=0 参与受众计算、把消息投给错误的人。
func (s *Service) ValidateUserOrgPairs(ctx context.Context, req *iam_service.ValidateUserOrgPairsReq) (*iam_service.ValidateUserOrgPairsResp, error) {
	if req.Limit > 0 && int32(len(req.Pairs)) > req.Limit {
		return &iam_service.ValidateUserOrgPairsResp{Exceeded: true}, nil
	}
	var filtered int32
	seen := make(map[orm.UserOrgPair]struct{}, len(req.Pairs))
	pairs := make([]orm.UserOrgPair, 0, len(req.Pairs))
	for _, p := range req.Pairs {
		userID, uErr := util.U32(p.UserId)
		orgID, oErr := util.U32(p.OrgId)
		if uErr != nil || oErr != nil || userID == 0 || orgID == 0 {
			filtered++
			continue
		}
		pair := orm.UserOrgPair{UserID: userID, OrgID: orgID}
		if _, ok := seen[pair]; ok {
			continue // 去重不计入 filtered
		}
		seen[pair] = struct{}{}
		pairs = append(pairs, pair)
	}
	valid, err := s.cli.FilterValidUserOrgPairs(ctx, pairs)
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	resp := &iam_service.ValidateUserOrgPairsResp{
		FilteredCount: filtered + int32(len(pairs)-len(valid)),
	}
	for _, p := range valid {
		resp.Pairs = append(resp.Pairs, &iam_service.UserOrgPair{
			UserId: util.Int2Str(p.UserID),
			OrgId:  util.Int2Str(p.OrgID),
		})
	}
	return resp, nil
}

// GetUserOrgMembership 查询单用户在指定组织的成员关系（加入时间 + 状态）。
// 供消息中心读侧 «vis» 的 joinedAt 屏蔽（新成员不追溯历史消息）使用。
func (s *Service) GetUserOrgMembership(ctx context.Context, req *iam_service.GetUserOrgMembershipReq) (*iam_service.GetUserOrgMembershipResp, error) {
	userID, uErr := util.U32(req.UserId)
	orgID, oErr := util.U32(req.OrgId)
	if uErr != nil || oErr != nil || userID == 0 || orgID == 0 {
		return &iam_service.GetUserOrgMembershipResp{Exists: false}, nil
	}
	membership, err := s.cli.GetUserOrgMembership(ctx, userID, orgID)
	if err != nil {
		return nil, errStatus(errs.Code_IAMOrg, err)
	}
	return &iam_service.GetUserOrgMembershipResp{
		Exists:   membership.Exists,
		JoinedAt: membership.JoinedAt,
		Active:   membership.Active,
	}, nil
}

// --- internal function ---

func toOrgInfo(org *orm.OrgInfo) *iam_service.OrgInfo {
	return &iam_service.OrgInfo{
		OrgId:      strconv.Itoa(int(org.ID)),
		Name:       org.Name,
		Remark:     org.Remark,
		Status:     org.Status,
		CreatedAt:  org.CreatedAt,
		Creator:    toIDName(org.Creator),
		UserCount:  org.UserCount,
		AvatarPath: org.AvatarPath,
		Admins:     org.Admins,
	}
}

func toAdminOrgTreeNodes(nodes []*orm.AdminOrgTreeNode) []*iam_service.AdminOrgTreeNode {
	if len(nodes) == 0 {
		return nil
	}
	var ret []*iam_service.AdminOrgTreeNode
	for _, node := range nodes {
		ret = append(ret, toAdminOrgTreeNode(node))
	}
	return ret
}

func toAdminOrgTreeNode(node *orm.AdminOrgTreeNode) *iam_service.AdminOrgTreeNode {
	return &iam_service.AdminOrgTreeNode{
		OrgId:      strconv.Itoa(int(node.ID)),
		Name:       node.Name,
		AvatarPath: node.AvatarPath,
		HasPerm:    node.HasPerm,
		Children:   toAdminOrgTreeNodes(node.Children),
	}
}
