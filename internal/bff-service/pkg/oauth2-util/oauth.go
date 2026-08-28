package oauth2_util

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"

	rsa_util "github.com/UnicomAI/wanwu/pkg/rsa-util"
)

var (
	_redis *redis.Client

	_rsaPublicKey  *rsa.PublicKey
	_rsaPrivateKey *rsa.PrivateKey

	_jwk JWK
	_kid string

	_issuer    string
	_jwtSecret string
)

type RSAConfig struct {
	PublicKeyPath  string `json:"public_key_path" mapstructure:"public_key_path"`
	PrivateKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
}

func Init(redisCli *redis.Client, rsa RSAConfig, issuer, jwtSecret string) error {
	if _redis != nil || _rsaPublicKey != nil || _rsaPrivateKey != nil || _kid != "" || _issuer != "" || _jwtSecret != "" {
		return errors.New("already init")
	}
	if redisCli == nil {
		return errors.New("redis nil")
	}
	if rsa.PublicKeyPath == "" || rsa.PrivateKeyPath == "" {
		return errors.New("rsa empty")
	}
	if issuer == "" {
		return errors.New("issuer empty")
	}

	// 自动生成RSA密钥文件（如果不存在）
	if err := rsa_util.EnsureKeyFiles(rsa.PrivateKeyPath, rsa.PublicKeyPath); err != nil {
		return fmt.Errorf("ensure oauth rsa key files failed: %w", err)
	}

	rsaPublicKey, err := os.ReadFile(rsa.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("read rsa public key err: %v", err)
	}
	rsaPrivateKey, err := os.ReadFile(rsa.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("read rsa private key err: %v", err)
	}

	pubKey, err := toRsaPublicKey(rsaPublicKey)
	if err != nil {
		return err
	}
	privKey, err := toRsaPrivateKey(rsaPrivateKey)
	if err != nil {
		return err
	}

	_redis = redisCli
	_rsaPublicKey = pubKey
	_rsaPrivateKey = privKey
	_jwk, _kid = getJWKandKid(pubKey)
	_issuer = issuer
	_jwtSecret = jwtSecret
	return nil
}

func GetIssuer() (string, error) {
	if err := checkInit(); err != nil {
		return "", err
	}
	return _issuer, nil
}

// --- internal ---

func checkInit() error {
	if _redis == nil || _rsaPublicKey == nil || _rsaPrivateKey == nil || _kid == "" || _issuer == "" || _jwtSecret == "" {
		return errors.New("not init")
	}
	return nil
}
