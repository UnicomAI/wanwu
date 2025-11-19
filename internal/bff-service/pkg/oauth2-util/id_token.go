package oauth2_util

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

const IDTokenTimeout = int64(60 * 30 * 24) // 1day

type idTokenClaims struct {
	UserID   string `json:"userId"`   // 用户ID [EN] User ID
	UserName string `json:"userName"` // 用户名称 [EN] Username
	jwt.StandardClaims
}

func GenerateIDToken(userID, userName, clientID string, timeout int64) (string, error) {
	if err := checkInit(); err != nil {
		return "", err
	}
	nowTime := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &idTokenClaims{
		UserID:   userID,
		UserName: userName,
		StandardClaims: jwt.StandardClaims{
			Issuer:    _issuer, // oidc root path
			Subject:   userID,  // 用途，目前固定user [EN] Purpose, currently fixed user
			Audience:  clientID,
			NotBefore: nowTime,           // 生效时间 [EN] Effective time
			ExpiresAt: nowTime + timeout, // 过期时间 [EN] Expiration time
		},
	})
	token.Header["kid"] = _kid
	tokenString, err := token.SignedString(_rsaPrivateKey)
	if err != nil {
		return "", err
	}
	return tokenString, err
}
