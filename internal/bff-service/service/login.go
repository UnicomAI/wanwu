package service

import (
	"fmt"
	"strconv"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context, login *request.Login, language string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail != 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFSingleLoginDisable)
	}

	// 使用RSA解密密码
	password, err := decryptCipherRSA(ctx.Request.Context(), login.Cipher, login.KeyID, challengeConsume)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}

	resp, err := iam.Login(ctx.Request.Context(), &iam_service.LoginReq{
		UserName: login.Username,
		Password: password,
		Key:      login.Key,
		Code:     login.Code,
		Language: language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

func LoginByEmail(ctx *gin.Context, login *request.Login) (*response.LoginByEmail, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}

	// 使用RSA解密密码
	password, err := decryptCipherRSA(ctx.Request.Context(), login.Cipher, login.KeyID, challengeConsume)
	if err != nil {
		return nil, fmt.Errorf("decrypt password err: %v", err)
	}

	resp, err := iam.LoginByEmail(ctx.Request.Context(), &iam_service.LoginByEmailReq{
		UserName: login.Username,
		Password: password,
		Key:      login.Key,
		Code:     login.Code,
	})
	if err != nil {
		return nil, err
	}
	// jwt token
	claims, token, err := jwt_util.GenerateToken(
		resp.UserId,
		jwt_util.UserLoginTokenTimeout,
	)
	if err != nil {
		return nil, err
	}
	ctx.Set(gin_util.CLAIMS, &claims)
	// resp
	return &response.LoginByEmail{
		IsEmailCheck:     resp.IsEmailCheck,
		Token:            token,
		IsUpdatePassword: resp.LastUpdatePasswordAt != 0,
	}, nil
}

func LoginEmailCheck(ctx *gin.Context, login *request.LoginEmailCheck, language, userId string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	resp, err := iam.LoginEmailCheck(ctx.Request.Context(), &iam_service.LoginEmailCheckReq{
		UserId:   userId,
		Email:    login.Email,
		Code:     login.Code,
		Language: language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

func ChangeUserPasswordByEmail(ctx *gin.Context, login *request.ChangeUserPasswordByEmail, language, userId string) (*response.Login, error) {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	oldPassword, err := decryptCipherRSA(ctx.Request.Context(), login.OldCipher, login.KeyID, challengeValidateOnly)
	if err != nil {
		return nil, fmt.Errorf("decrypt old password err: %v", err)
	}
	newPassword, err := decryptCipherRSA(ctx.Request.Context(), login.NewCipher, login.KeyID, challengeConsume)
	if err != nil {
		return nil, fmt.Errorf("decrypt new password err: %v", err)
	}
	resp, err := iam.ChangeUserPasswordByEmail(ctx.Request.Context(), &iam_service.ChangeUserPasswordByEmailReq{
		UserId:      userId,
		OldPassword: oldPassword,
		NewPassword: newPassword,
		Email:       login.Email,
		Code:        login.Code,
		Language:    language,
	})
	if err != nil {
		return nil, err
	}
	return getLoginResp(ctx, resp)
}

// --- login email code---
func LoginSendEmailCode(ctx *gin.Context, email string) error {
	if config.Cfg().CustomInfo.LoginByEmail == 0 {
		return grpc_util.ErrorStatus(err_code.Code_BFFLoginDisable)
	}
	_, err := iam.LoginSendEmailCode(ctx.Request.Context(), &iam_service.LoginSendEmailCodeReq{
		Email: email,
	})
	return err
}

func getLoginResp(ctx *gin.Context, resp *iam_service.LoginResp) (*response.Login, error) {
	// orgs
	orgs, err := iam.GetOrgSelect(ctx.Request.Context(), &iam_service.GetOrgSelectReq{UserId: resp.User.GetUserId()})
	if err != nil {
		return nil, err
	}
	// jwt token
	claims, token, err := jwt_util.GenerateToken(
		resp.User.GetUserId(),
		jwt_util.UserTokenTimeout,
	)
	if err != nil {
		return nil, err
	}
	ctx.Set(gin_util.CLAIMS, &claims)
	// resp
	return &response.Login{
		UID:              resp.User.GetUserId(),
		Username:         resp.User.GetUserName(),
		Nickname:         resp.User.GetNickName(),
		Token:            token,
		ExpiresAt:        claims.ExpiresAt * 1000, // 超时事件戳毫秒
		ExpireIn:         strconv.FormatInt(jwt_util.UserTokenTimeout, 10),
		Orgs:             toOrgIDNamesWithAvatar(ctx, orgs.Selects, resp.User.GetUserId() == config.SystemAdminUserID),
		OrgPermission:    toOrgPermission(ctx, resp.Permission),
		Language:         getLanguageByCode(resp.User.Language),
		IsUpdatePassword: resp.Permission.LastUpdatePasswordAt != 0,
	}, nil
}
