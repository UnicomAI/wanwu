package iam

import (
	"context"
	"strconv"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gromitlee/access/pkg/perm"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetRoleSelect(ctx context.Context, req *iam_service.GetRoleSelectReq) (*iam_service.GetRoleSelectResp, error) {
	roles, err := s.cli.SelectRoles(ctx, util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	return &iam_service.GetRoleSelectResp{Roles: toRoleIDNames(roles)}, nil
}

func (s *Service) GetGlobalRoleSelect(ctx context.Context, req *iam_service.GetGlobalRoleSelectReq) (*iam_service.GetGlobalRoleSelectResp, error) {
	roles, err := s.cli.SelectGlobalRoles(ctx)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	return &iam_service.GetGlobalRoleSelectResp{Roles: toRoleIDNames(roles)}, nil
}

func (s *Service) GetGlobalRoleList(ctx context.Context, req *iam_service.GetGlobalRoleListReq) (*iam_service.GetGlobalRoleListResp, error) {
	roles, count, err := s.cli.GetGlobalRoles(ctx, util.MustU32(req.OrgId), req.Name, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	resp := &iam_service.GetGlobalRoleListResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, role := range roles {
		resp.Roles = append(resp.Roles, toRoleInfo(role))
	}
	return resp, nil
}

func (s *Service) GetRoleList(ctx context.Context, req *iam_service.GetRoleListReq) (*iam_service.GetRoleListResp, error) {
	roles, count, err := s.cli.GetRoles(ctx, util.MustU32(req.OrgId), req.Name, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	resp := &iam_service.GetRoleListResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, role := range roles {
		resp.Roles = append(resp.Roles, toRoleInfo(role))
	}
	return resp, nil
}

func (s *Service) GetRoleInfo(ctx context.Context, req *iam_service.GetRoleInfoReq) (*iam_service.RoleInfo, error) {
	roleID := util.MustU32(req.RoleId)
	isGlobal, err := s.cli.IsGlobalRole(ctx, roleID)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_get", util.Int2Str(roleID), err.Error()))
	}
	if isGlobal {
		role, status := s.cli.GetGlobalRole(ctx, util.MustU32(req.OrgId), roleID)
		if status != nil {
			return nil, errStatus(errs.Code_IAMRole, status)
		}
		return toRoleInfo(role), nil
	}
	role, status := s.cli.GetRole(ctx, util.MustU32(req.OrgId), roleID)
	if status != nil {
		return nil, errStatus(errs.Code_IAMRole, status)
	}
	return toRoleInfo(role), nil
}

func (s *Service) CreateRole(ctx context.Context, req *iam_service.CreateRoleReq) (*iam_service.RoleIDName, error) {
	var perms []perm.Perm
	for _, p := range req.Perms {
		perms = append(perms, perm.Perm{Obj: perm.Obj(p.Perm)})
	}
	// isGlobal: create global role
	if req.IsGlobal {
		roleID, err := s.cli.CreateGlobalRole(ctx, util.MustU32(req.CreatorId), req.Name, req.Remark, req.AvatarPath, perms)
		if err != nil {
			return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_create", err.Error()))
		}
		return &iam_service.RoleIDName{
			Id:       strconv.Itoa(int(roleID)),
			Name:     req.Name,
			IsAdmin:  false,
			IsSystem: true,
			IsGlobal: true,
		}, nil
	}
	// non-global: create org role
	roleID, err := s.cli.CreateRole(ctx, util.MustU32(req.OrgId), util.MustU32(req.CreatorId), req.Name, req.Remark, req.AvatarPath, perms)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_create", err.Error()))
	}
	return &iam_service.RoleIDName{
		Id:       strconv.Itoa(int(roleID)),
		Name:     req.Name,
		IsAdmin:  false,
		IsSystem: util.MustU32(req.OrgId) == config.TopOrgID(),
		IsGlobal: false,
	}, nil
}

func (s *Service) UpdateRole(ctx context.Context, req *iam_service.UpdateRoleReq) (*emptypb.Empty, error) {
	var perms []perm.Perm
	for _, p := range req.Perms {
		perms = append(perms, perm.Perm{Obj: perm.Obj(p.Perm)})
	}
	roleID := util.MustU32(req.RoleId)
	// auto-detect global role
	isGlobal, err := s.cli.IsGlobalRole(ctx, roleID)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_update", util.Int2Str(roleID), err.Error()))
	}
	if isGlobal {

		if status := s.cli.UpdateGlobalRole(ctx, roleID, req.Name, req.Remark, req.AvatarPath, perms); status != nil {
			return nil, errStatus(errs.Code_IAMRole, status)
		}
		return &emptypb.Empty{}, nil
	}

	if err := s.cli.UpdateRole(ctx, util.MustU32(req.OrgId), roleID, req.Name, req.Remark, req.AvatarPath, perms); err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteRole(ctx context.Context, req *iam_service.DeleteRoleReq) (*emptypb.Empty, error) {
	roleID := util.MustU32(req.RoleId)
	// auto-detect global role
	isGlobal, err := s.cli.IsGlobalRole(ctx, roleID)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_delete", util.Int2Str(roleID), err.Error()))
	}
	if isGlobal {
		if status := s.cli.DeleteGlobalRole(ctx, roleID); status != nil {
			return nil, errStatus(errs.Code_IAMRole, status)
		}
		return &emptypb.Empty{}, nil
	}
	if err := s.cli.DeleteRole(ctx, util.MustU32(req.OrgId), roleID); err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ChangeRoleStatus(ctx context.Context, req *iam_service.ChangeRoleStatusReq) (*emptypb.Empty, error) {
	roleID := util.MustU32(req.RoleId)

	// auto-detect global role
	isGlobal, err := s.cli.IsGlobalRole(ctx, roleID)
	if err != nil {
		return nil, errStatus(errs.Code_IAMRole, toErrStatus("iam_role_change", util.Int2Str(roleID), err.Error()))
	}
	if isGlobal {
		if status := s.cli.ChangeGlobalRoleStatus(ctx, roleID, req.Status); status != nil {
			return nil, errStatus(errs.Code_IAMRole, status)
		}
		return &emptypb.Empty{}, nil
	}
	if err := s.cli.ChangeRoleStatus(ctx, util.MustU32(req.OrgId), roleID, req.Status); err != nil {
		return nil, errStatus(errs.Code_IAMRole, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetRoleUsers(ctx context.Context, req *iam_service.GetRoleUsersReq) (*iam_service.GetRoleUsersResp, error) {
	users, count, status := s.cli.GetRoleUsers(ctx, util.MustU32(req.RoleId), util.MustU32(req.OrgId), req.Name, toOffset(req), req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_IAMRole, status)
	}
	resp := &iam_service.GetRoleUsersResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, u := range users {
		ru := &iam_service.RoleUser{
			UserId:     strconv.Itoa(int(u.UserID)),
			UserName:   u.UserName,
			Phone:      u.Phone,
			Email:      u.Email,
			AvatarPath: u.AvatarPath,
		}
		if len(u.Orgs) > 0 {
			ru.Orgs = toIDNames(u.Orgs)
		}
		resp.Users = append(resp.Users, ru)
	}
	return resp, nil
}

func (s *Service) RemoveRoleUser(ctx context.Context, req *iam_service.RemoveRoleUserReq) (*emptypb.Empty, error) {
	if status := s.cli.RemoveRoleUser(ctx, util.MustU32(req.RoleId), util.MustU32(req.UserId), util.MustU32(req.OrgId)); status != nil {
		return nil, errStatus(errs.Code_IAMRole, status)
	}
	return &emptypb.Empty{}, nil
}

// --- internal function ---

func toRoleInfo(role *orm.RoleInfo) *iam_service.RoleInfo {
	ret := &iam_service.RoleInfo{
		RoleId:     strconv.Itoa(int(role.ID)),
		Name:       role.Name,
		Remark:     role.Remark,
		IsAdmin:    role.IsAdmin,
		IsSystem:   role.IsSystem,
		IsGlobal:   role.IsGlobal,
		Status:     role.Status,
		CreatedAt:  role.CreatedAt,
		Creator:    toIDName(role.Creator),
		OrgName:    role.OrgName,
		UserCount:  role.UserCount,
		AvatarPath: role.AvatarPath,
	}
	for _, perm := range role.Perms {
		ret.Perms = append(ret.Perms, toPerm(perm))
	}
	return ret
}
