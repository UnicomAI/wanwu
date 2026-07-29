package oauth2_util

import (
	"errors"
	"fmt"
	"time"

	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SUBJECT_ACCESS     = "access"
	AccessTokenTimeout = int64(60 * 60 * 24) // 1day
)

type AccessTokenClaims struct {
	Scope    []string `json:"scope"`    // access token访问范围
	UserID   string   `json:"userId"`   // 用户ID
	ClientID string   `json:"clientId"` // Client ID
	jwt.RegisteredClaims
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
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    _issuer,
			Subject:   SUBJECT_ACCESS, // 用途，目前固定access
			NotBefore: jwt.NewNumericDate(time.Unix(nowTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(nowTime+timeout, 0)),
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
	tokenClaims, err := jwt.ParseWithClaims(token, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 算法白名单断言：仅允许 HMAC 算法，拒绝 alg 混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(_jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, jwt_util.ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt_util.ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, jwt_util.ErrTokenNotValidYet
		}
		return nil, jwt_util.ErrTokenInvalid
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*AccessTokenClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
		return nil, jwt_util.ErrTokenInvalid
	}
	return nil, jwt_util.ErrTokenInvalid
}
