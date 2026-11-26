package mcpoauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
)

// Codec encrypts persisted grants with an administrator-provided AES-256 key.
// The associated identity prevents moving ciphertext between users or resources.
type Codec struct{ aead cipher.AEAD }

func NewCodec(key []byte) (*Codec, error) {
	if len(key) != 32 {
		return nil, errors.New("OAuth encryption key must contain 32 bytes")
	}
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	a, e := cipher.NewGCM(b)
	if e != nil {
		return nil, e
	}
	return &Codec{a}, nil
}
func binding(o Owner, resource string) []byte {
	b, _ := json.Marshal([]string{o.Organization, o.User, resource})
	return b
}
func (c *Codec) Seal(o Owner, resource string, g *Grant) ([]byte, error) {
	plain, e := json.Marshal(g)
	if e != nil {
		return nil, e
	}
	return c.sealBytes(o, resource, plain)
}
func (c *Codec) sealBytes(o Owner, resource string, plain []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, e := rand.Read(nonce); e != nil {
		return nil, e
	}
	return c.aead.Seal(nonce, nonce, plain, binding(o, resource)), nil
}
func (c *Codec) Open(o Owner, resource string, data []byte) (*Grant, error) {
	plain, e := c.openBytes(o, resource, data)
	if e != nil {
		return nil, e
	}
	var g Grant
	if e = json.Unmarshal(plain, &g); e != nil {
		return nil, ErrUnauthorized
	}
	return &g, nil
}

func (c *Codec) openBytes(o Owner, resource string, data []byte) ([]byte, error) {
	n := c.aead.NonceSize()
	if len(data) < n {
		return nil, ErrUnauthorized
	}
	plain, err := c.aead.Open(nil, data[:n], data[n:], binding(o, resource))
	if err != nil {
		return nil, ErrUnauthorized
	}
	return plain, nil
}
