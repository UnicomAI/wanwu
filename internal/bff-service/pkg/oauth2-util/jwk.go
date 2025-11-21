package oauth2_util

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/google/uuid"
)

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func GetJWK() (JWK, error) {
	if err := checkInit(); err != nil {
		return JWK{}, err
	}
	return _jwk, nil
}

func getJWKandKid(pubKey *rsa.PublicKey) (JWK, string) {
	kid := uuid.New().String()

	// Export public key parameters n and e
	n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}, kid
}
