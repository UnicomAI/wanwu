package oauth2_util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const IDTokenTimeout = int64(60 * 30 * 24) // 1day

type idTokenClaims struct {
	UserID   string `json:"userId"`   // 用户ID
	UserName string `json:"userName"` // 用户名称
	jwt.RegisteredClaims
}

func GenerateIDToken(userID, userName, clientID string, timeout int64) (string, error) {
	if err := checkInit(); err != nil {
		return "", err
	}
	nowTime := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &idTokenClaims{
		UserID:   userID,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    _issuer, // oidc root path
			Subject:   userID,  // 用途，目前固定user
			Audience:  []string{clientID},
			NotBefore: jwt.NewNumericDate(time.Unix(nowTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(nowTime+timeout, 0)),
		},
	})
	token.Header["kid"] = _kid
	tokenString, err := token.SignedString(_rsaPrivateKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
}
