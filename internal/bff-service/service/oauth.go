package service

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OAuthLogin(ctx *gin.Context, req *request.OAuthLoginRequest) (string, error) {
	issuer, err := oauth2_util.GetIssuer()
	if err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	// e.g. "http://localhost:8081/service/api/openapi/v1" + "../../../../aibase/login" => http://localhost:8081/aibase/login
	loginUri, err := url.JoinPath(issuer, "../../../../aibase/login")
	if err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	// 将 []string 转为以空格分隔的字符串
	scopeStr := strings.Join(req.Scopes, " ")

	oauthApp, err := iam.GetOauthApp(ctx.Request.Context(), &iam_service.GetOauthAppReq{
		ClientId: req.ClientID,
	})
	if err != nil {
		return "", err
	}
	err = oauthValidateReqApp(req.ClientID, "", req.RedirectURI, oauthApp)
	if err != nil {
		return "", err
	}
	loginURI := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&scope=%s&client_name=%s&redirect_uri=%s&state=%s",
		loginUri,
		url.QueryEscape(oauthApp.ClientId), //ID
		url.QueryEscape("code"),
		url.QueryEscape(scopeStr),
		url.QueryEscape(oauthApp.Name),
		url.QueryEscape(oauthApp.RedirectUri),
		url.QueryEscape(req.State), // 对state也进行编码
	)

	return loginURI, nil
}

func OAuthAuthorize(ctx *gin.Context, req *request.OAuthRequest) (string, error) {
	userID, err := jwtUserAuth(ctx, req.JwtToken)
	if err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFJWT, err.Error())
	}
	oauthApp, err := iam.GetOauthApp(ctx.Request.Context(), &iam_service.GetOauthAppReq{
		ClientId: req.ClientID,
	})
	if err != nil {
		return "", err
	}
	err = oauthValidateReqApp(req.ClientID, "", req.RedirectURI, oauthApp)
	if err != nil {
		return "", err
	}
	//code save to redis
	code := uuid.NewString()
	if err := oauth2_util.SaveCode(ctx, code, oauth2_util.CodePayload{
		ClientID: req.ClientID,
		UserID:   userID,
	}); err != nil {
		return "", grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	redirectURI := fmt.Sprintf(
		"%s?code=%s&state=%s",
		oauthApp.RedirectUri,
		url.QueryEscape(code),
		url.QueryEscape(req.State), // 对state也进行编码
	)
	return redirectURI, nil
}

func OAuthToken(ctx *gin.Context, req *request.OAuthTokenRequest) (*response.OAuthTokenResponse, error) {
	oauthApp, err := iam.GetOauthApp(ctx.Request.Context(), &iam_service.GetOauthAppReq{
		ClientId: req.ClientID,
	})
	if err != nil {
		return nil, err
	}
	err = oauthValidateReqApp(req.ClientID, req.ClientSecret, req.RedirectURI, oauthApp)
	if err != nil {
		return nil, err
	}
	codePayload, err := oauth2_util.ValidateCode(ctx, req.Code, req.ClientID)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	user, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: codePayload.UserID,
		OrgId:  "",
	})
	if err != nil {
		return nil, err
	}
	//access token
	scopes := []string{} //预留scope处理
	accessToken, err := oauth2_util.GenerateAccessToken(user.UserId, req.ClientID, scopes, oauth2_util.AccessTokenTimeout)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFJWT, err.Error())
	}
	//id token
	idToken, err := oauth2_util.GenerateIDToken(user.UserId, user.UserName, req.ClientID, oauth2_util.IDTokenTimeout)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFJWT, err.Error())
	}
	//refresh token
	refreshToken, err := oauth2_util.GenerateRefreshToken(ctx, user.UserId, req.ClientID, oauth2_util.RefreshTokenExpiration)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	return &response.OAuthTokenResponse{
		AccessToken:  accessToken,
		ExpiresIn:    oauth2_util.AccessTokenTimeout,
		TokenType:    "Bearer",
		IDToken:      idToken,
		RefreshToken: refreshToken,
		Scope:        scopes,
	}, nil
}

func OAuthRefresh(ctx *gin.Context, req *request.OAuthRefreshRequest) (*response.OAuthRefreshTokenResponse, error) {
	oauthApp, err := iam.GetOauthApp(ctx.Request.Context(), &iam_service.GetOauthAppReq{
		ClientId: req.ClientID,
	})
	if err != nil {
		return nil, err
	}
	err = oauthValidateReqApp(req.ClientID, req.ClientSecret, "", oauthApp)
	if err != nil {
		return nil, err
	}
	refreshPayload, err := oauth2_util.ValidateRefreshToken(ctx, req.RefreshToken, req.ClientID)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	scopes := []string{} //scopes处理预留
	//new access token
	accessToken, err := oauth2_util.GenerateAccessToken(refreshPayload.UserID, req.ClientID, scopes, oauth2_util.AccessTokenTimeout)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFJWT, err.Error())
	}
	//new refresh token
	refreshToken, err := oauth2_util.GenerateRefreshToken(ctx, refreshPayload.UserID, refreshPayload.ClientID, oauth2_util.RefreshTokenExpiration)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	return &response.OAuthRefreshTokenResponse{
		AccessToken:  accessToken,
		ExpiresAt:    strconv.Itoa(int(time.Now().Add(time.Duration(jwt_util.UserTokenTimeout) * time.Second).UnixMilli())),
		RefreshToken: refreshToken,
	}, nil
}

func OAuthConfig(ctx *gin.Context) (*response.OAuthConfig, error) {
	issuer, err := oauth2_util.GetIssuer()
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return &response.OAuthConfig{
		Issuer:           issuer,
		AuthEndpoint:     issuer + "/oauth/login",
		TokenEndpoint:    issuer + "/oauth/code/token",
		JwksUri:          issuer + "/oauth/jwks",
		UserInfoEndpoint: issuer + "/oauth/userinfo",
		ResponseTypes:    []string{"code"},
		IDtokenSignAlg:   []string{"RS256"},
		SubjectTypes:     []string{"public"},
	}, nil
}

func OAuthJWKS(ctx *gin.Context) (*response.OAuthJWKS, error) {
	jwk, err := oauth2_util.GetJWK()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_err", err.Error())
	}
	return &response.OAuthJWKS{Keys: []oauth2_util.JWK{jwk}}, nil
}

func OAuthGetUserInfo(ctx *gin.Context, userID string) (*response.OAuthGetUserInfo, error) {
	user, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
		UserId: userID,
		OrgId:  "",
	})
	if err != nil {
		return nil, err
	}
	issuer, err := oauth2_util.GetIssuer()
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	avatar := cacheUserAvatar(ctx, user.AvatarPath)
	// e.g. "http://localhost:8081/service/api/openapi/v1" + "../.." + "/v1/static/icon/user-default-icon.png" => http://localhost:8081/service/api/v1/static/icon/user-default-icon.png
	avatarUri, err := url.JoinPath(issuer, "../..", avatar.Path)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return &response.OAuthGetUserInfo{
		UserID:    user.UserId,
		Username:  user.UserName,
		Email:     user.Email,
		Nickname:  user.NickName,
		Phone:     user.Phone,
		Gender:    user.Gender,
		AvatarUri: avatarUri,
		Remark:    user.Remark,
		Company:   user.Company,
	}, nil
}

func oauthValidateReqApp(clientID, clientSecret, redirectUri string, appInfo *iam_service.OauthApp) error {
	if !appInfo.Status {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_status", clientID)
	}
	if appInfo.ClientId != clientID {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_client_id", clientID)
	}
	if clientSecret != "" && appInfo.ClientSecret != clientSecret {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_secret")
	}
	if redirectUri != "" {
		if redirectUri != appInfo.RedirectUri {
			return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_oauth_redirect_uri")
		}
	}
	return nil
}

func CreateOauthApp(ctx *gin.Context, userId string, req *request.CreateOauthAppReq) error {
	isURI := isValidURI(req.RedirectURI)
	if !isURI {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "redirect uri invalid")
	}
	_, err := iam.CreateOauthApp(ctx.Request.Context(), &iam_service.CreateOauthAppReq{
		UserId:      userId,
		Name:        req.Name,
		Desc:        req.Desc,
		RedirectUri: req.RedirectURI,
	})
	if err != nil {
		return err
	}
	return nil
}

func DeleteOauthApp(ctx *gin.Context, req *request.DeleteOauthAppReq) error {
	_, err := iam.DeleteOauthApp(ctx.Request.Context(), &iam_service.DeleteOauthAppReq{
		ClientId: req.ClientID,
	})
	if err != nil {
		return err
	}
	return nil
}

func UpdateOauthApp(ctx *gin.Context, req *request.UpdateOauthAppReq) error {
	isURI := isValidURI(req.RedirectURI)
	if !isURI {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "redirect uri invalid")
	}
	_, err := iam.UpdateOauthApp(ctx.Request.Context(), &iam_service.UpdateOauthAppReq{
		ClientId:    req.ClientID,
		Name:        req.Name,
		Desc:        req.Desc,
		RedirectUri: req.RedirectURI,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetOauthAppList(ctx *gin.Context, userId, name string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := iam.GetOauthAppList(ctx.Request.Context(), &iam_service.GetOauthAppListReq{
		UserId:   userId,
		Name:     name,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	var retList []*response.OAuthAppInfo
	for _, app := range resp.Apps {
		retList = append(retList, &response.OAuthAppInfo{
			ClientID:     app.ClientId,
			Name:         app.Name,
			Desc:         app.Desc,
			ClientSecret: app.ClientSecret,
			RedirectURI:  app.RedirectUri,
			Status:       app.Status,
		})
	}
	return &response.PageResult{
		List:     retList,
		Total:    resp.Total,
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func UpdateOauthAppStatus(ctx *gin.Context, req *request.UpdateOauthAppStatusReq) error {
	_, err := iam.UpdateOauthAppStatus(ctx.Request.Context(), &iam_service.UpdateOauthAppStatusReq{
		ClientId: req.ClientID,
		Status:   req.Status,
	})
	if err != nil {
		return err
	}
	return nil
}

func isValidURI(rawURI string) bool {
	u, err := url.ParseRequestURI(rawURI)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func jwtUserAuth(ctx *gin.Context, token string) (string, error) {
	claims, err := jwt_util.ParseToken(token)
	if err != nil {
		return "", err
	}
	if claims.Subject != jwt_util.SUBJECT_USER {
		return "", fmt.Errorf("token subject错误")
	}

	//ctx.Set(gin_util.CLAIMS, claims)
	return claims.UserID, nil
}
