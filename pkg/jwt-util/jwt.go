package jwt_util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// jwt subject
	SUBJECT_USER = "user"

	UserTokenTimeout      = int64(60 * 60 * 24) // 1天
	BufferTime            = int64(60 * 60 * 2)
	UserLoginTokenTimeout = int64(60 * 5) // 5 min
)

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

var (
	userSecretKey string
)

type CustomClaims struct {
	UserID     string `json:"userId"` // 用户ID
	BufferTime int64  `json:"bufferTime"`
	jwt.RegisteredClaims
}

func InitUserJWT(key string) error {
	if userSecretKey != "" {
		return errors.New("already init")
	}
	if key == "" {
		return errors.New("secret key empty")
	}
	userSecretKey = key
	return nil
}

func GenerateToken(userID string, timeout int64) (*CustomClaims, string, error) {
	return generateToken(userID, timeout, userSecretKey)
}

func ParseToken(token string) (*CustomClaims, error) {
	return parseToken(token, userSecretKey)
}

func generateToken(id string, timeout int64, secretKey string) (*CustomClaims, string, error) {
	if secretKey == "" {
		return nil, "", errors.New("jwt secret key empty")
	}
	nowTime := time.Now().Unix()
	claims := &CustomClaims{
		UserID:     id,
		BufferTime: nowTime + BufferTime, // 缓冲时间，当nowTime大于等于BufferTime and nowTime小于ExpiresAt是获得新的token
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "wanwu",
			Subject:   SUBJECT_USER, // 用途，目前固定user
			NotBefore: jwt.NewNumericDate(time.Unix(nowTime, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(nowTime+timeout, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return nil, "", err
	}
	return claims, token, err
}

func parseToken(token, secretKey string) (*CustomClaims, error) {
	if secretKey == "" {
		return nil, errors.New("jwt secret key empty")
	}
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 算法白名单断言：仅允许 HMAC 算法，拒绝 alg 混淆攻击
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrTokenInvalid
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
		return nil, ErrTokenInvalid
	}
	return nil, ErrTokenInvalid
}
