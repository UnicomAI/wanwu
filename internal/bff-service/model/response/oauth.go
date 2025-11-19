package response

import (
	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
)

type OAuthTokenResponse struct {
	AccessToken  string   `json:"access_token"`  // 访问令牌 [EN] access token
	ExpiresIn    int64    `json:"expires_in"`    // token过期时间(毫秒时间戳) [EN] token expiration time (millisecond timestamp)
	IDToken      string   `json:"id_token"`      // ID令牌 [EN] ID token
	TokenType    string   `json:"token_type"`    // 令牌类型(bearer) [EN] Token type (bearer)
	RefreshToken string   `json:"refresh_token"` // 刷新令牌 [EN] refresh token
	Scope        []string `json:"scope"`         // 权限范围 [EN] Scope of authority
}

type OAuthRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`  // 访问令牌 [EN] access token
	RefreshToken string `json:"refresh_token"` // 刷新令牌(可选) [EN] Refresh token (optional)
	ExpiresAt    string `json:"expires_at"`    // token过期时间(毫秒时间戳) [EN] token expiration time (millisecond timestamp)
}

type OAuthConfig struct {
	Issuer           string   `json:"issuer"`                                // auth的base url [EN] auth base url
	AuthEndpoint     string   `json:"authorization_endpoint"`                // 获取授权码接口 [EN] Obtain authorization code interface
	TokenEndpoint    string   `json:"token_endpoint"`                        // 获取token接口 [EN] Get token interface
	JwksUri          string   `json:"jwks_uri"`                              // 获取jwt公钥 [EN] Get jwt public key
	UserInfoEndpoint string   `json:"userinfo_endpoint"`                     //获取用户信息接口 [EN] Obtain user information interface
	ResponseTypes    []string `json:"response_types_supported"`              // 授权模式，默认code [EN] Authorization mode, default code
	IDtokenSignAlg   []string `json:"id_token_signing_alg_values_supported"` //jwt签名算法 [EN] jwt signature algorithm
	SubjectTypes     []string `json:"subject_types_supported"`               // 用户标识类型，即 ID Token中的sub是如何生成的。 [EN] The user identification type, that is, how the sub in the ID Token is generated.
}

type OAuthAppInfo struct {
	ClientID     string `json:"clientId"`     // 客户端ID [EN] Client ID
	Name         string `json:"name"`         // 应用名称 [EN] Application name
	Desc         string `json:"desc"`         // 应用描述 [EN] Application description
	ClientSecret string `json:"clientSecret"` // 客户端密钥 [EN] client key
	RedirectURI  string `json:"redirectUri"`  // oauth重定向地址 [EN] oauth redirect address
	Status       bool   `json:"status"`       // oauth应用开关 [EN] oauth application switch
}

type OAuthJWKS struct {
	Keys []oauth2_util.JWK `json:"keys"`
}

type OAuthGetUserInfo struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
	Remark    string `json:"remark"`
	Company   string `json:"company"`
	AvatarUri string `json:"avatar"`
}
