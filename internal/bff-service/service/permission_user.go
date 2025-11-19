package service

import (
	"encoding/base64"
	"fmt"
	"net/url"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateUser(ctx *gin.Context, creatorID, orgID string, userCreate *request.UserCreate) (*response.UserID, error) {
	password, err := decryptPD(userCreate.Password)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}
	resp, err := iam.CreateUser(ctx.Request.Context(), &iam_service.CreateUserReq{
		CreatorId: creatorID,
		OrgId:     orgID,
		UserName:  userCreate.Username,
		NickName:  userCreate.Nickname,
		Gender:    userCreate.Gender,
		Phone:     userCreate.Phone,
		Company:   userCreate.Company,
		Remark:    userCreate.Remark,
		Password:  password,
		RoleIds:   userCreate.RoleIDs,
	})
	if err != nil {
		return nil, err
	}
	return &response.UserID{UserID: resp.Id}, nil
}

func ChangeUser(ctx *gin.Context, orgID string, userUpdate *request.UserUpdate) error {
	_, err := iam.UpdateUser(ctx.Request.Context(), &iam_service.UpdateUserReq{
		UserId:   userUpdate.UserID,
		OrgId:    orgID,
		NickName: userUpdate.Nickname,
		Gender:   userUpdate.Gender,
		Phone:    userUpdate.Phone,
		Company:  userUpdate.Company,
		Remark:   userUpdate.Remark,
		RoleIds:  userUpdate.RoleIDs,
	})
	return err
}

func DeleteUser(ctx *gin.Context, userID string) error {
	_, err := iam.DeleteUser(ctx.Request.Context(), &iam_service.DeleteUserReq{
		UserId: userID,
	})
	return err
}

func GetUserInfo(ctx *gin.Context, userID, orgID string) (*response.UserInfo, error) {
	resp, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	return toUserInfo(ctx, resp), nil
}

func GetUserList(ctx *gin.Context, orgID, name string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := iam.GetUserList(ctx.Request.Context(), &iam_service.GetUserListReq{
		OrgId:    orgID,
		UserName: name,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	var users []*response.UserInfo
	for _, user := range resp.Users {
		users = append(users, toUserInfo(ctx, user))
	}
	return &response.PageResult{
		List:     users,
		Total:    resp.Total,
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func ChangeUserStatus(ctx *gin.Context, userID, orgID string, status bool) error {
	_, err := iam.ChangeUserStatus(ctx.Request.Context(), &iam_service.ChangeUserStatusReq{
		UserId: userID,
		OrgId:  orgID,
		Status: status,
	})
	return err
}

func ChangeUserPassword(ctx *gin.Context, userID, oldPwd, newPwd string) error {
	oldPassword, err := decryptPD(oldPwd)
	if err != nil {
		return fmt.Errorf("decrypt password err: %v", err)
	}
	newPassword, err := decryptPD(newPwd)
	if err != nil {
		return fmt.Errorf("decrypt password err: %v", err)
	}
	_, err = iam.UpdateUserPassword(ctx.Request.Context(), &iam_service.UpdateUserPasswordReq{
		UserId:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	return err
}

func AdminChangeUserPassword(ctx *gin.Context, userID, pwd string) error {
	password, err := decryptPD(pwd)
	if err != nil {
		return fmt.Errorf("decrypt password err: %v", err)
	}
	_, err = iam.ResetUserPassword(ctx.Request.Context(), &iam_service.ResetUserPasswordReq{
		UserId:   userID,
		Password: password,
	})
	return err
}

func GetOrgUserNotSelect(ctx *gin.Context, orgID, name string) (*response.Select, error) {
	users, err := iam.GetUserSelectNotInOrg(ctx.Request.Context(), &iam_service.GetUserSelectNotInOrgReq{
		OrgId:    orgID,
		UserName: name,
	})
	if err != nil {
		return nil, err
	}
	return &response.Select{Select: toIDNames(users.Selects)}, nil
}

func GetRoleSelect(ctx *gin.Context, orgID string) (*response.Select, error) {
	roles, err := iam.GetRoleSelect(ctx.Request.Context(), &iam_service.GetRoleSelectReq{
		OrgId: orgID,
	})
	if err != nil {
		return nil, err
	}
	return &response.Select{Select: toRoleIDNames(ctx, roles.Roles)}, nil
}

func AddOrgUser(ctx *gin.Context, orgID, userID, roleID string) error {
	_, err := iam.AddOrgUser(ctx.Request.Context(), &iam_service.AddOrgUserReq{
		OrgId:  orgID,
		UserId: userID,
		RoleId: roleID,
	})
	return err
}

func RemoveOrgUser(ctx *gin.Context, orgID, userID string) error {
	_, err := iam.RemoveOrgUser(ctx.Request.Context(), &iam_service.RemoveOrgUserReq{
		OrgId:  orgID,
		UserId: userID,
	})
	return err
}

func UpdateUserAvatar(ctx *gin.Context, userID, key string) error {
	_, err := iam.UpdateUserAvatar(ctx.Request.Context(), &iam_service.UpdateUserAvatarReq{
		UserId:     userID,
		AvatarPath: key,
	})
	return err
}

// --- internal ---

func toUserInfo(ctx *gin.Context, user *iam_service.UserInfo) *response.UserInfo {
	ret := &response.UserInfo{
		UserID:    user.UserId,
		Username:  user.UserName,
		Nickname:  user.NickName,
		Phone:     user.Phone,
		Email:     user.Email,
		Gender:    user.Gender,
		Remark:    user.Remark,
		Company:   user.Company,
		CreatedAt: util.Time2Str(user.CreatedAt),
		Creator:   toIDName(user.Creator),
		Status:    user.Status,
		Language:  getLanguageByCode(user.Language),
		Avatar:    cacheUserAvatar(ctx, user.AvatarPath),
	}
	for _, userOrg := range user.Orgs {
		ret.Orgs = append(ret.Orgs, toOrgRole(ctx, userOrg))
	}
	return ret
}

func toOrgRole(ctx *gin.Context, userOrg *iam_service.UserOrg) response.OrgRole {
	return response.OrgRole{
		Org:   toOrgIDName(ctx, userOrg.Org),
		Roles: toRoleIDNames(ctx, userOrg.Roles),
	}
}

// 解密password [EN] Decrypt password
func decryptPD(encryptStr string) (string, error) {
	var (
		err                      error
		urlUnescape              string
		base64Decode, decryptAes []byte
	)
	if encryptStr == "" {
		return "", nil
	}

	if urlUnescape, err = url.QueryUnescape(encryptStr); nil != err {
		return "", err
	}

	if base64Decode, err = base64.StdEncoding.DecodeString(urlUnescape); nil != err {
		return "", err
	}

	iv := []byte(config.Cfg().Decrypt.IV)
	key := []byte(config.Cfg().Decrypt.Key)
	if decryptAes, err = util.DecryptAES(base64Decode, key, iv); nil != err {
		return "", err
	}

	return string(decryptAes), nil
}
