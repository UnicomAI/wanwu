package jwt_util

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	// jwt subject
	SUBJECT_USER = "user"

	UserTokenTimeout      = int64(60 * 60 * 24) // 1天 [EN] 1 day
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
	UserID     string `json:"userId"` // 用户ID [EN] User ID
	BufferTime int64  `json:"bufferTime"`
	jwt.StandardClaims
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
		BufferTime: nowTime + BufferTime, // 缓冲时间，当nowTime大于等于BufferTime and nowTime小于ExpiresAt是获得新的token [EN] Buffer time, when nowTime is greater than or equal to BufferTime and nowTime is less than ExpiresAt, a new token is obtained
		StandardClaims: jwt.StandardClaims{
			Issuer:    "wanwu",
			Subject:   SUBJECT_USER,      // 用途，目前固定user [EN] Purpose, currently fixed user
			NotBefore: nowTime,           // 生效时间 [EN] Effective time
			ExpiresAt: nowTime + timeout, // 过期时间 [EN] Expiration time
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
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
	}
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
		return nil, ErrTokenInvalid

	} else {
		return nil, ErrTokenInvalid
	}
}
