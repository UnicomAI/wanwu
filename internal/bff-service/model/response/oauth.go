package response

import (
	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
)

type OAuthTokenResponse struct {
	AccessToken  string   `json:"access_token"`  // access token
	ExpiresIn    int64    `json:"expires_in"`    // token expiration time (millisecond timestamp)
	IDToken      string   `json:"id_token"`      // ID token
	TokenType    string   `json:"token_type"`    // Token type (bearer)
	RefreshToken string   `json:"refresh_token"` // refresh token
	Scope        []string `json:"scope"`         // Scope of authority
}

type OAuthRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`  // access token
	RefreshToken string `json:"refresh_token"` // Refresh token (optional)
	ExpiresAt    string `json:"expires_at"`    // token expiration time (millisecond timestamp)
}

type OAuthConfig struct {
	Issuer           string   `json:"issuer"`                                // auth base url
	AuthEndpoint     string   `json:"authorization_endpoint"`                // Obtain authorization code interface
	TokenEndpoint    string   `json:"token_endpoint"`                        // Get token interface
	JwksUri          string   `json:"jwks_uri"`                              // Get jwt public key
	UserInfoEndpoint string   `json:"userinfo_endpoint"`                     //Obtain user information interface
	ResponseTypes    []string `json:"response_types_supported"`              // Authorization mode, default code
	IDtokenSignAlg   []string `json:"id_token_signing_alg_values_supported"` //jwt signature algorithm
	SubjectTypes     []string `json:"subject_types_supported"`               // The user identification type, that is, how the sub in the ID Token is generated.
}

type OAuthAppInfo struct {
	ClientID     string `json:"clientId"`     // Client ID
	Name         string `json:"name"`         // Application name
	Desc         string `json:"desc"`         // Application description
	ClientSecret string `json:"clientSecret"` // client key
	RedirectURI  string `json:"redirectUri"`  // oauth redirect address
	Status       bool   `json:"status"`       // oauth application switch
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
