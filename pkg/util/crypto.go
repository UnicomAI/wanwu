package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func MD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(src string) string {
	hash := sha256.Sum256([]byte(src))
	return hex.EncodeToString(hash[:])
}

func DecryptAES(src []byte, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(src, src)
	if src, err = unPadding(src); err != nil {
		return nil, err
	}
	return src, nil
}

func unPadding(src []byte) ([]byte, error) {
	length := len(src)
	unPadNum := int(src[length-1])
	if length <= unPadNum {
		return nil, errors.New("decryption failed")
	}
	return src[:length-unPadNum], nil
}
