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

func (s *Service) GetUserSelectNotInOrg(ctx context.Context, req *iam_service.GetUserSelectNotInOrgReq) (*iam_service.Select, error) {
	users, err := s.cli.SelectUsersNotInOrg(ctx, util.MustU32(req.OrgId), req.UserName)
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &iam_service.Select{Selects: toIDNames(users)}, nil
}

func (s *Service) GetUserSelectByUserIDs(ctx context.Context, req *iam_service.GetUserSelectByUserIDsReq) (*iam_service.Select, error) {
	var userIDs []uint32
	for _, userID := range req.UserIds {
		userIDs = append(userIDs, util.MustU32(userID))
	}
	users, err := s.cli.SelectUsersByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &iam_service.Select{Selects: toIDNames(users)}, nil
}

func (s *Service) GetUserList(ctx context.Context, req *iam_service.GetUserListReq) (*iam_service.GetUserListResp, error) {
	var roleIDs []uint32
	for _, roleID := range req.RoleIds {
		roleIDs = append(roleIDs, util.MustU32(roleID))
	}
	users, count, err := s.cli.GetUsers(ctx, util.MustU32(req.OrgId), req.UserName, req.Email, roleIDs, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	resp := &iam_service.GetUserListResp{
		Total:    count,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}
	for _, user := range users {
		resp.Users = append(resp.Users, toUserInfo(user))
	}
	return resp, nil
}

func (s *Service) GetUserInfo(ctx context.Context, req *iam_service.GetUserInfoReq) (*iam_service.UserInfo, error) {
	user, err := s.cli.GetUser(ctx, util.MustU32(req.UserId), util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return toUserInfo(user), nil
}

func (s *Service) CreateUser(ctx context.Context, req *iam_service.CreateUserReq) (*iam_service.IDName, error) {
	var roleIDs []uint32
	for _, roleID := range req.RoleIds {
		roleIDs = append(roleIDs, util.MustU32(roleID))
	}
	userID, err := s.cli.CreateUser(ctx, &model.User{
		Status:    true,
		CreatorID: util.MustU32(req.CreatorId),
		Name:      req.UserName,
		Nick:      req.NickName,
		Gender:    req.Gender,
		Phone:     req.Phone,
		Email:     req.Email,
		Company:   req.Company,
		Remark:    req.Remark,
		Password:  req.Password,
	}, util.MustU32(req.OrgId), roleIDs)
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &iam_service.IDName{Id: strconv.Itoa(int(userID)), Name: req.UserName}, nil
}

func (s *Service) CreateUsers(ctx context.Context, req *iam_service.CreateUsersReq) (*iam_service.CreateUsersResp, error) {
	result, err := s.cli.CreateUsers(ctx, toUsersInfo(req.Users), util.MustU32(req.CreatorId), util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	resp := &iam_service.CreateUsersResp{
		Total:   int32(result.Total),
		Success: int32(result.Success),
		Failed:  int32(result.Failed),
	}
	for _, e := range result.Errors {
		resp.Errors = append(resp.Errors, &iam_service.CreateUsersError{
			Index:  int32(e.Index),
			Reason: e.Reason,
		})
	}
	return resp, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *iam_service.UpdateUserReq) (*emptypb.Empty, error) {
	var roleIDs []uint32
	for _, roleID := range req.RoleIds {
		roleIDs = append(roleIDs, util.MustU32(roleID))
	}
	if err := s.cli.UpdateUser(ctx, &model.User{
		ID:     util.MustU32(req.UserId),
		Name:   req.UserName,
		Phone:  req.Phone,
		Email:  req.Email,
		Remark: req.Remark,
	}, util.MustU32(req.OrgId), roleIDs); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *iam_service.DeleteUserReq) (*emptypb.Empty, error) {
	if err := s.cli.DeleteUser(ctx, util.MustU32(req.UserId)); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ChangeUserStatus(ctx context.Context, req *iam_service.ChangeUserStatusReq) (*emptypb.Empty, error) {
	if err := s.cli.ChangeUserStatus(ctx, util.MustU32(req.UserId), util.MustU32(req.OrgId), req.Status); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, req *iam_service.UpdateUserPasswordReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdateUserPassword(ctx, util.MustU32(req.UserId), req.OldPassword, req.NewPassword); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) ResetUserPassword(ctx context.Context, req *iam_service.ResetUserPasswordReq) (*emptypb.Empty, error) {
	if err := s.cli.ResetUserPassword(ctx, util.MustU32(req.UserId), req.Password); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetUserPermission(ctx context.Context, req *iam_service.GetUserPermissionReq) (*iam_service.UserPermission, error) {
	permission, err := s.cli.GetUserPermission(ctx, util.MustU32(req.UserId), util.MustU32(req.OrgId))
	if err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	ret := toPermission(permission)
	ret.UserId = req.UserId
	return ret, nil
}

func (s *Service) IsUserOrgAdmin(ctx context.Context, req *iam_service.IsUserOrgAdminReq) (*iam_service.IsUserOrgAdminResp, error) {
	isOrgAdmin, status := s.cli.IsUserOrgAdmin(ctx, util.MustU32(req.UserId))
	if status != nil {
		return nil, errStatus(errs.Code_IAMUser, status)
	}
	return &iam_service.IsUserOrgAdminResp{IsOrgAdmin: isOrgAdmin}, nil
}

func (s *Service) IsAdminInOrgs(ctx context.Context, req *iam_service.IsAdminInOrgsReq) (*iam_service.IsAdminInOrgsResp, error) {
	var orgIDs []uint32
	for _, orgID := range req.OrgIds {
		orgIDs = append(orgIDs, util.MustU32(orgID))
	}
	isAdmin, status := s.cli.IsAdminInOrgs(ctx, util.MustU32(req.UserId), orgIDs)
	if status != nil {
		return nil, errStatus(errs.Code_IAMUser, status)
	}
	return &iam_service.IsAdminInOrgsResp{IsAdmin: isAdmin}, nil
}

func (s *Service) ChangeUserLanguage(ctx context.Context, req *iam_service.ChangeUserLanguageReq) (*emptypb.Empty, error) {
	if err := s.cli.ChangeUserLanguage(ctx, util.MustU32(req.UserId), req.Language); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateUserAvatar(ctx context.Context, req *iam_service.UpdateUserAvatarReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdateUserAvatar(ctx, util.MustU32(req.UserId), req.AvatarPath); err != nil {
		return nil, errStatus(errs.Code_IAMUser, err)
	}
	return &emptypb.Empty{}, nil
}

// --- internal function ---

func toUserInfo(user *orm.UserInfo) *iam_service.UserInfo {
	ret := &iam_service.UserInfo{
		UserId:     strconv.Itoa(int(user.ID)),
		Status:     user.Status,
		UserName:   user.Name,
		NickName:   user.Nick,
		Gender:     user.Gender,
		Phone:      user.Phone,
		Email:      user.Email,
		Company:    user.Company,
		Remark:     user.Remark,
		CreatedAt:  user.CreatedAt,
		Creator:    toIDName(user.Creator),
		Language:   user.Language,
		AvatarPath: user.AvatarPath,
	}
	for _, userOrg := range user.Orgs {
		ret.Orgs = append(ret.Orgs, &iam_service.UserOrg{
			Org:   toIDName(userOrg.Org.IDName),
			Roles: toRoleIDNames(userOrg.Roles),
		})
	}
	return ret
}

func toPermission(permission *orm.Permission) *iam_service.UserPermission {
	ret := &iam_service.UserPermission{
		IsAdmin:              permission.IsAdmin,
		IsSystem:             permission.IsSystem,
		Org:                  toIDName(permission.Org),
		Roles:                toRoleIDNames(permission.Roles),
		LastUpdatePasswordAt: permission.LastUpdatePasswordAt,
	}
	for _, perm := range permission.Perms {
		ret.Perms = append(ret.Perms, toPerm(perm))
	}
	return ret
}

func toPerm(perm orm.Perm) *iam_service.Perm {
	return &iam_service.Perm{
		Perm: perm.Perm,
	}
}

func toUsersInfo(req []*iam_service.CreateUsersInfo) []*orm.UsersInfo {
	ret := make([]*orm.UsersInfo, 0, len(req))
	for _, user := range req {
		ret = append(ret, &orm.UsersInfo{
			UserName: user.UserName,
			Phone:    user.Phone,
			Email:    user.Email,
			Company:  user.Company,
			Remark:   user.Remark,
			Password: user.Password,
			RoleName: user.RoleName,
		})
	}
	return ret
}
