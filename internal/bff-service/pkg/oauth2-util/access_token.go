package oauth2_util

import (
	"time"

	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/dgrijalva/jwt-go"
)

const (
	SUBJECT_ACCESS     = "access"
	AccessTokenTimeout = int64(60 * 60 * 24) // 1day
)

type AccessTokenClaims struct {
	Scope    []string `json:"scope"`    // access token访问范围 [EN] access token access scope
	UserID   string   `json:"userId"`   // 用户ID [EN] User ID
	ClientID string   `json:"clientId"` // Client ID
	jwt.StandardClaims
}

func GenerateAccessToken(userID, clientID string, scopes []string, timeout int64) (string, error) {
	if err := checkInit(); err != nil {
		return "", err
	}
	nowTime := time.Now().Unix()
	access_token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &AccessTokenClaims{
		UserID:   userID,
		ClientID: clientID,
		Scope:    scopes,
		StandardClaims: jwt.StandardClaims{
			Issuer:    _issuer,
			Subject:   SUBJECT_ACCESS,    // 用途，目前固定access [EN] Purpose, currently fixed access
			NotBefore: nowTime,           // 生效时间 [EN] Effective time
			ExpiresAt: nowTime + timeout, // 过期时间 [EN] Expiration time
		},
	}).SignedString([]byte(_jwtSecret))
	if err != nil {
		return "", err
	}
	return access_token, err
}

func ParseAccessToken(token string) (*AccessTokenClaims, error) {
	if err := checkInit(); err != nil {
		return nil, err
	}
	tokenClaims, err := jwt.ParseWithClaims(token, &AccessTokenClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(_jwtSecret), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, jwt_util.ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, jwt_util.ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, jwt_util.ErrTokenNotValidYet
			} else {
				return nil, jwt_util.ErrTokenInvalid
			}
		}
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*AccessTokenClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
		return nil, jwt_util.ErrTokenInvalid
	} else {
		return nil, jwt_util.ErrTokenInvalid
	}
}
